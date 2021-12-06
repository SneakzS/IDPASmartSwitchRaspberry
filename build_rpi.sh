#!/bin/bash
export GOOS=linux
export GOARCH=arm
export GOARM=7
export CGO_ENABLED=1
export CC=arm-linux-gnueabihf-gcc
#aarch64-none-elf-gcc
#export CROSS_COMPILE=aarch64-none-gnu-

go build -o idpa_rpi ./cmd/idpa 