.PHONY: api demo app test vet reset-db

# Sign-up codes show on screen until .env holds SMTP settings (README → Real emails); then they're emailed.
api:
	INNINGS_ENV_FILE=.env INNINGS_DEV_CODES=1 go run ./cmd/api

# A demo on a fresh demo.db, in place of `make api`: today's real matches, codes on screen and never emailed, and a
# demo account with stubs, points, friends and a Top 5. It prints the account's new password.
# `INNINGS_MATCHES=sim make demo` runs it on the simulated world instead: offline, and one match is always live.
demo:
	rm -f demo.db*
	env -u INNINGS_ENV_FILE -u SMTP_HOST -u SMTP_USERNAME -u SMTP_PASSWORD DB_PATH=demo.db INNINGS_DEV_CODES=1 INNINGS_DEMO=1 go run ./cmd/api

app:
	npx expo start --port 8082

test:
	go test ./...
	npx jest

vet:
	go vet ./...

reset-db:
	rm -f innings.db*
