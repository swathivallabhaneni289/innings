.PHONY: api app test vet reset-db

# Sign-up codes show on screen until .env holds SMTP settings (README → Real emails); then they're emailed.
api:
	INNINGS_ENV_FILE=.env INNINGS_DEV_CODES=1 go run ./cmd/api

app:
	npx expo start --port 8082

test:
	go test ./...
	npx jest

vet:
	go vet ./...

reset-db:
	rm -f innings.db*
