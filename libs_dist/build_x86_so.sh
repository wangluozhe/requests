# build x86 so
go env -w GOOS=linux
go env -w GOARCH=386
go build -buildmode=c-shared -o requests-go-x86.so export.go