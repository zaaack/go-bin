package main

//go:generate go build -o ./go-bin ./cmd/go-bin
//go:generate sh -c "GOOS=windows GOARCH=amd64 go build -o ./go-bin.exe ./cmd/go-bin"
