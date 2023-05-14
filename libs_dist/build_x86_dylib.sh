# build x86 dylib
go env -w GOOS=darwin
go env -w GOARCH=386
go build -buildmode=c-shared -o requests-go-x86.dylib export.go