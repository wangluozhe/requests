:: build win32 dll
GOOS=windows GOARCH=386 go build -buildmode=c-shared -o requests-go-win32.dll export.go