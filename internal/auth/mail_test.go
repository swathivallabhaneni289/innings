package auth

import (
	"bufio"
	"context"
	"encoding/base64"
	"io"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeSMTP is just enough of an SMTP server to take one message: EHLO, AUTH PLAIN, MAIL, RCPT, DATA, QUIT. It never
// offers STARTTLS, so only a mailer talking to this machine may sign in to it.
type fakeSMTP struct {
	addr     string
	user, pw string

	mu       sync.Mutex
	authed   string // "user:password" as received
	from, to string
	data     string
}

func startFakeSMTP(t *testing.T, user, pw string) *fakeSMTP {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	f := &fakeSMTP{addr: ln.Addr().String(), user: user, pw: pw}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go f.serve(conn)
		}
	}()
	return f
}

func (f *fakeSMTP) serve(conn net.Conn) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	r := bufio.NewReader(conn)
	say := func(s string) { io.WriteString(conn, s+"\r\n") }
	say("220 fake ESMTP")
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		verb := strings.ToUpper(strings.SplitN(line, " ", 2)[0])
		switch verb {
		case "EHLO", "HELO":
			say("250-fake")
			say("250 AUTH PLAIN")
		case "AUTH":
			parts := strings.Fields(line)
			raw, _ := base64.StdEncoding.DecodeString(parts[len(parts)-1])
			fields := strings.Split(string(raw), "\x00") // authzid, user, password
			f.mu.Lock()
			f.authed = fields[1] + ":" + fields[2]
			f.mu.Unlock()
			if fields[1] == f.user && fields[2] == f.pw {
				say("235 2.7.0 Accepted")
			} else {
				say("535 5.7.8 Username and Password not accepted")
			}
		case "MAIL":
			f.mu.Lock()
			f.from = line
			f.mu.Unlock()
			say("250 OK")
		case "RCPT":
			f.mu.Lock()
			f.to = line
			f.mu.Unlock()
			say("250 OK")
		case "DATA":
			say("354 Go ahead")
			var b strings.Builder
			for {
				l, err := r.ReadString('\n')
				if err != nil {
					return
				}
				if l == ".\r\n" {
					break
				}
				b.WriteString(l)
			}
			f.mu.Lock()
			f.data = b.String()
			f.mu.Unlock()
			say("250 OK queued")
		case "QUIT":
			say("221 Bye")
			return
		default:
			say("502 Not implemented")
		}
	}
}

func (f *fakeSMTP) mailer(host string) *SMTPMailer {
	from, _ := mail.ParseAddress("Innings <diary@example.com>")
	_, port, _ := net.SplitHostPort(f.addr)
	return &SMTPMailer{Host: host, Port: port, Username: "diary@example.com", Password: "app-pass", From: from, Timeout: 5 * time.Second,
		dial: func(ctx context.Context, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "tcp", f.addr)
		}}
}

func TestSMTPMailerSends(t *testing.T) {
	f := startFakeSMTP(t, "diary@example.com", "app-pass")
	msg := codeEmail("fan@example.com", PurposeVerify, "042917")
	msg.Text += ".a line that starts with a dot\nand a long one: " + strings.Repeat("é", 60) + "\n"
	if err := f.mailer("127.0.0.1").Send(context.Background(), msg); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.authed != "diary@example.com:app-pass" || f.from != "MAIL FROM:<diary@example.com>" || f.to != "RCPT TO:<fan@example.com>" {
		t.Fatalf("envelope: auth %q, %q, %q", f.authed, f.from, f.to)
	}
	head, body, ok := strings.Cut(f.data, "\r\n\r\n")
	if !ok {
		t.Fatalf("no header/body break: %q", f.data)
	}
	for _, h := range []string{`From: "Innings" <diary@example.com>`, "To: fan@example.com", "Subject: 042917 is your Innings code",
		"MIME-Version: 1.0", "Content-Type: text/plain; charset=utf-8", "Content-Transfer-Encoding: quoted-printable", "Message-ID: <", "Date: "} {
		if !strings.Contains(head, h) {
			t.Errorf("header %q missing from:\n%s", h, head)
		}
	}
	// SMTP dot-stuffing on the wire, then quoted-printable: undo both and the text comes back exactly.
	unstuffed := strings.ReplaceAll(body, "\r\n..", "\r\n.")
	decoded, err := io.ReadAll(quotedprintable.NewReader(strings.NewReader(unstuffed)))
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.ReplaceAll(string(decoded), "\r\n", "\n"); got != msg.Text {
		t.Fatalf("body round trip:\n got %q\nwant %q", got, msg.Text)
	}
	for _, l := range strings.Split(body, "\r\n") {
		if len(l) > 78 {
			t.Fatalf("line longer than 78 characters: %q", l)
		}
	}
}

func TestSMTPMailerRefusesPlainTextAndBadLogins(t *testing.T) {
	f := startFakeSMTP(t, "diary@example.com", "app-pass")
	msg := codeEmail("fan@example.com", PurposeReset, "123456")

	// A server that isn't this machine and offers no STARTTLS never sees the password.
	err := f.mailer("smtp.example.com").Send(context.Background(), msg)
	if err == nil || !strings.Contains(err.Error(), "STARTTLS") {
		t.Fatalf("plain-text server: %v", err)
	}
	f.mu.Lock()
	if f.authed != "" {
		t.Fatalf("password sent in the clear: %q", f.authed)
	}
	f.mu.Unlock()

	wrong := f.mailer("127.0.0.1")
	wrong.Password = "nope"
	if err := wrong.Send(context.Background(), msg); err == nil || !strings.Contains(err.Error(), "sign in") {
		t.Fatalf("wrong password: %v", err)
	}
	if err := f.mailer("127.0.0.1").Send(context.Background(), Message{To: "fan@example.com\r\nBcc: x@example.com", Subject: "x", Text: "x"}); err == nil {
		t.Fatal("a line break in a header must be refused")
	}
	if !isLocalhost("localhost") || !isLocalhost("::1") || isLocalhost("smtp.gmail.com") || isLocalhost("192.168.1.10") {
		t.Fatal("isLocalhost")
	}
}
