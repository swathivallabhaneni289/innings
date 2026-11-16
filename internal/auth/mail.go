package auth

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

// Message is one plain-text email.
type Message struct {
	To      string
	Subject string
	Text    string
}

// Mailer sends sign-in emails. There are two:
//   - DevMailer (INNINGS_DEV_CODES=1): writes each email to the log instead of sending it, and the API hands the code
//     back to the app as devCode, so sign-up can be tried with no mail account.
//   - SMTPMailer: sends through a real mail server, such as Gmail with an app password.
type Mailer interface {
	Send(ctx context.Context, m Message) error
}

// devCodes is implemented only by DevMailer. Its method is unexported, so nothing outside this package can opt in:
// a code that was really emailed never reaches a response.
type devCodes interface{ revealsCodes() }

// DevMailer is for local development with no mail account: it logs each email instead of sending it.
type DevMailer struct{ Log *slog.Logger }

func (DevMailer) revealsCodes() {}

func (d DevMailer) Send(_ context.Context, m Message) error {
	log := d.Log
	if log == nil {
		log = slog.Default()
	}
	log.Info("dev email (not sent)", "to", m.To, "subject", m.Subject)
	return nil
}

// SMTPMailer sends through an SMTP server with a username and password. On port 465 the connection is TLS from the
// start; on any other port (587 is the usual one) the server must offer STARTTLS, so the password never crosses the
// network in the clear. The one exception is a server on this machine, which tests use.
type SMTPMailer struct {
	Host     string
	Port     string
	Username string
	Password string
	From     *mail.Address
	Timeout  time.Duration // for the whole conversation; 0 means 20 seconds

	dial func(ctx context.Context, addr string) (net.Conn, error) // tests connect to a fake server here
}

func (m *SMTPMailer) Send(ctx context.Context, msg Message) error {
	if strings.ContainsAny(msg.To+msg.Subject, "\r\n") {
		return errors.New("smtp: line break in a header")
	}
	timeout := m.Timeout
	if timeout == 0 {
		timeout = 20 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	addr := net.JoinHostPort(m.Host, m.Port)
	dialer := &net.Dialer{}
	var conn net.Conn
	var err error
	switch {
	case m.dial != nil:
		conn, err = m.dial(ctx, addr)
	case m.Port == "465":
		conn, err = (&tls.Dialer{NetDialer: dialer, Config: m.tlsConfig()}).DialContext(ctx, "tcp", addr)
	default:
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("smtp: connect: %w", err)
	}
	deadline, _ := ctx.Deadline()
	conn.SetDeadline(deadline)
	c, err := smtp.NewClient(conn, m.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("smtp: greeting: %w", err)
	}
	defer c.Close()
	if err := c.Hello("innings.local"); err != nil {
		return fmt.Errorf("smtp: hello: %w", err)
	}
	if _, isTLS := conn.(*tls.Conn); !isTLS {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(m.tlsConfig()); err != nil {
				return fmt.Errorf("smtp: starttls: %w", err)
			}
		} else if !isLocalhost(m.Host) {
			return errors.New("smtp: the server doesn't offer STARTTLS, so the password would be sent in the clear")
		}
	}
	if err := c.Auth(smtp.PlainAuth("", m.Username, m.Password, m.Host)); err != nil {
		return fmt.Errorf("smtp: sign in: %w", err)
	}
	if err := c.Mail(m.From.Address); err != nil {
		return fmt.Errorf("smtp: sender: %w", err)
	}
	if err := c.Rcpt(msg.To); err != nil {
		return fmt.Errorf("smtp: recipient: %w", err)
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp: data: %w", err)
	}
	if _, err := w.Write(m.compose(msg, time.Now())); err != nil {
		return fmt.Errorf("smtp: write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp: send: %w", err)
	}
	return c.Quit()
}

func (m *SMTPMailer) tlsConfig() *tls.Config {
	return &tls.Config{ServerName: m.Host, MinVersion: tls.VersionTLS12}
}

func isLocalhost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// compose renders msg as a UTF-8, quoted-printable plain-text email with CRLF line endings.
func (m *SMTPMailer) compose(msg Message, now time.Time) []byte {
	var b bytes.Buffer
	_, domain, _ := strings.Cut(m.From.Address, "@")
	for _, h := range [][2]string{
		{"From", m.From.String()},
		{"To", msg.To},
		{"Subject", mime.QEncoding.Encode("utf-8", msg.Subject)},
		{"Date", now.Format(time.RFC1123Z)},
		{"Message-ID", "<" + newToken()[:32] + "@" + domain + ">"},
		{"MIME-Version", "1.0"},
		{"Content-Type", "text/plain; charset=utf-8"},
		{"Content-Transfer-Encoding", "quoted-printable"},
	} {
		b.WriteString(h[0] + ": " + h[1] + "\r\n")
	}
	b.WriteString("\r\n")
	qp := quotedprintable.NewWriter(&b)
	qp.Write([]byte(strings.ReplaceAll(msg.Text, "\n", "\r\n")))
	qp.Close()
	return b.Bytes()
}

// MailerFromEnv picks how sign-in emails go out and names the choice for the startup log:
//   - "smtp": SMTPMailer, when SMTP_HOST, SMTP_USERNAME and SMTP_PASSWORD are all set. SMTP_PORT defaults to 587,
//     and MAIL_FROM to "Innings <SMTP_USERNAME>";
//   - "dev": DevMailer, when INNINGS_DEV_CODES=1 and SMTP isn't configured;
//   - "off": nil, so every call that would email a code answers 503 email-unavailable.
//
// Setting only some of the SMTP variables is an error rather than a quiet fall back to dev mode, which would hand
// sign-in codes to whoever asks for them.
func MailerFromEnv(getenv func(string) string, logger *slog.Logger) (Mailer, string, error) {
	host, user, pass := getenv("SMTP_HOST"), getenv("SMTP_USERNAME"), getenv("SMTP_PASSWORD")
	switch set := host != "" || user != "" || pass != ""; {
	case host != "" && user != "" && pass != "":
		from := getenv("MAIL_FROM")
		if from == "" {
			from = "Innings <" + user + ">"
		}
		addr, err := mail.ParseAddress(from)
		if err != nil {
			return nil, "", errors.New(`MAIL_FROM must be an address, like "Innings <you@gmail.com>"`)
		}
		port := getenv("SMTP_PORT")
		if port == "" {
			port = "587"
		}
		return &SMTPMailer{Host: host, Port: port, Username: user, Password: pass, From: addr}, "smtp", nil
	case set:
		return nil, "", errors.New("set all of SMTP_HOST, SMTP_USERNAME and SMTP_PASSWORD, or none")
	case getenv("INNINGS_DEV_CODES") == "1":
		return DevMailer{Log: logger}, "dev", nil
	default:
		return nil, "off", nil
	}
}

// codeEmail is the email carrying a code. The code leads the subject so it shows in the notification, and the body
// says it is a code, which is what lets iPhones offer it above the keyboard.
func codeEmail(to string, purpose Purpose, code string) Message {
	what, ignore := "confirm this email address for your Innings account",
		"If you didn't sign up for Innings, you can ignore this email. Nothing happens without the code."
	if purpose == PurposeReset {
		what, ignore = "choose a new password for your Innings account",
			"If you didn't ask to reset your password, you can ignore this email. Your password stays the same."
	}
	return Message{
		To:      to,
		Subject: code + " is your Innings code",
		Text: fmt.Sprintf("Your Innings code is %s\n\nEnter it in the app to %s. It works for %d minutes.\n\n%s\n",
			code, what, int(CodeTTL/time.Minute), ignore),
	}
}
