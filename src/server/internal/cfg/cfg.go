package cfg

import (
	"flag"
)

type Flag struct {
	ProtoType string
	IP        string
	Port      string
}

func NewFlagsFromArgs() *Flag {
	f := &Flag{}

	flag.StringVar(&f.IP, "ip", "127.0.0.1", "ip address")
	flag.StringVar(&f.Port, "port", "4545", "port")
	flag.StringVar(&f.ProtoType, "p", "", "protocol type")
	flag.Parse()

	return f
}
