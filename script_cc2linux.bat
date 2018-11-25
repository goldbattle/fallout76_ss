

:: set that we are building for linux
:: https://github.com/golang/go/wiki/GccgoCrossCompilation
SET GOARCH=amd64
SET GOOS=linux
SET CGO_ENABLED=1
SET CC=g++
set GOGCCFLAGS=-Wno-error -m64 -pthread -fmessage-length=0 -gno-record-gcc-switches


:: actually do the build
go build

