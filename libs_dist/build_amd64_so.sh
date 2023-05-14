# build amd64 so
go env -w GOOS=linux
go env -w GOARCH=amd64
go build -buildmode=c-shared -o requests-go-amd64.so export.go