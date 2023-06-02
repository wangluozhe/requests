# apt-get install gcc-10-aarch64-linux-gnu
# build arm64 so
GOOS=linux CGO_ENABLED=1 GOARCH=arm64 CC="aarch64-linux-gnu-gcc-10" go build -buildmode=c-shared -o requests-go-arm64.so export.go