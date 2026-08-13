BINARY := word-it-out
MODULES := . app game game/types game/service repository

.PHONY: build run test vet tidy clean

build:
	go build -o $(BINARY) .

run:
	go run .

test:
	go test ./...

vet:
	go vet ./...

tidy:
	for module in $(MODULES); do \
		echo "==> go mod tidy ($$module)"; \
		(cd $$module && go mod tidy) || exit 1; \
	done

clean:
	rm -f $(BINARY)
