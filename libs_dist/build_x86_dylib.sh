# build x86 dylib
GOOS=darwin GOARCH=amd64 go build -buildmode=c-shared -o requests-go-x86.dylib export.go