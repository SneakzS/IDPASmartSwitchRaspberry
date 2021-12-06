#!/bin/bash
export GOOS=windows
export GOARCH=386
export CGO_ENABLED=1
export CC=i686-w64-mingw32-gcc
#aarch64-none-elf-gcc
#export CROSS_COMPILE=aarch64-none-gnu-

go build -o idpa_win.exe ./cmd/idpa 