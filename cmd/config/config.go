package config

import (
	"flag"
	"fmt"
)

type Flags struct {
	RunAddr       string
	BaseShortAddr string
}

func (f *Flags) String() string {
	return fmt.Sprintf("RunAddr: %s, BaseShortAddr: %s", f.RunAddr, f.BaseShortAddr)
}

func Init() *Flags {
	addr := flag.String("a", "localhost:8080", "Address and port to run server")
	base := flag.String("b", "http://localhost:8080", "Base shorten url")
	flag.Parse()
	return &Flags{
		RunAddr:       *addr,
		BaseShortAddr: *base,
	}
}
