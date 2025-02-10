build:
	@go build -o bin/dagger ./main.go

run: build
	@./bin/dagger start $(ARGS)

test:
	@go test ./... -v --race

# Database migrations
migrate-up: build
	@./bin/dagger migrate up $(ARGS)

migrate-down: build
	@./bin/dagger migrate down $(ARGS)
