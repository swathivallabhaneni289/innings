.PHONY: api app test vet reset-db

api:
	go run ./cmd/api

app:
	npx expo start --port 8082

test:
	go test ./...
	npx jest

vet:
	go vet ./...

reset-db:
	rm -f innings.db*
