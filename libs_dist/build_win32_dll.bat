:: build win32 dll
go env -w GOOS=windows
go env -w GOARCH=386
go build -buildmode=c-shared -o requests-go-win32.dll export.go