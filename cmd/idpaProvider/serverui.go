package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
)

var (
	//go:embed templates/*
	templateFiles embed.FS
	//go:embed static/*
	static embed.FS
)

type dummyfs struct{}

var _ fs.FS = dummyfs{}

func (dummyfs) Open(name string) (fs.File, error) {
	fmt.Println(name)
	return nil, os.ErrNotExist
}
