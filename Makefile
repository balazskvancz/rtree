test:
	go test ./...

lint:
	gofmt -w .

bench:
	go test -bench=. -benchmem -run=^#
