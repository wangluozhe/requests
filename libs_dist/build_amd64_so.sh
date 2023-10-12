# build amd64 so
GOOS=linux GOARCH=amd64 go build -buildmode=c-shared -o requests-go-amd64.so export.go