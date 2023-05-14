:: build win64 dll
go env -w GOOS=windows
go env -w GOARCH=amd64
go build -buildmode=c-shared -o requests-go-win64.dll export.go