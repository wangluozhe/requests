# build arm64 dylib
GOOS=darwin GOARCH=arm64 go build -buildmode=c-shared -o requests-go-arm64.dylib export.go