CGO_ENABLED=0
GOAMD64=v4

clean:
	rm -rf bin/
	go clean -cache -testcache

format:
	go fmt ./...
	go run mvdan.cc/gofumpt@latest -w .
	go mod tidy

verify:
	go mod verify
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

check:
	go vet ./...
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks all ./...
	go run go.uber.org/nilaway/cmd/nilaway@latest ./...
	go run golang.org/x/tools/cmd/deadcode@latest -test ./...

test:
	go test -vet off -count 1 -cover ./...

bench:
	go test -vet off -run ^$$ -bench . -benchtime 30s -benchmem ./...

cover:
	go test -vet off -cover -coverprofile cover.out ./...
	go tool cover -html cover.out
	rm -f cover.out

docs:
	go run golang.org/x/tools/cmd/godoc@latest

build:
	go build -o bin/ ./...
