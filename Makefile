.PHONY: build run test lint clean

build:
	go build -o aws-eks-calculator .

run: build
	./aws-eks-calculator

test:
	go test -v -race -coverprofile=coverage.out ./...

lint:
	go vet ./...

clean:
	rm -f aws-eks-calculator
