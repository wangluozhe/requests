# build x86 so
GOOS=linux GOARCH=386 go build -buildmode=c-shared -o requests-go-x86.so export.go