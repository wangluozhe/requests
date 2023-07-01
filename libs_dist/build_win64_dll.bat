:: build win64 dll
GOOS=windows GOARCH=arm64 go build -buildmode=c-shared -o requests-go-win64.dll export.go