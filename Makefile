.PHONY: build run test lint clean

build:
	go build -o aws-eks-calculator .

run: build
	./aws-eks-calculator

test:
	go test -v -race -coverprofile=coverage.out ./...

clean:
	rm -f aws-eks-calculator dist/ coverage.out coverage.html

fmt:
	go fmt ./...

vet:
	go vet ./...

go-lint:
	golangci-lint run --timeout 5m

yaml-lint:
	uvx yamllint -c .yamllint.yml .

markdown-lint:
	uvx pymarkdownlnt fix .

lint: fmt vet go-lint yaml-lint markdown-lint
