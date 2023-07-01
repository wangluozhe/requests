# build arm64 so
GOOS=linux GOARCH=arm64 go build -buildmode=c-shared -o requests-go-arm64.so export.go