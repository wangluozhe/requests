# build x86 dylib
GOOS=darwin GOARCH=386 go build -buildmode=c-shared -o requests-go-x86.dylib export.go