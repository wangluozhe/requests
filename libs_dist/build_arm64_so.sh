# build arm64 so
go env -w GOOS=linux
go env -w GOARCH=arm64
go build -buildmode=c-shared -o requests-go-arm64.so export.go