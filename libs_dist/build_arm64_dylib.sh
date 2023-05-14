# build arm64 dylib
go env -w GOOS=darwin
go env -w GOARCH=arm64
go build -buildmode=c-shared -o requests-go-arm64.dylib export.go