package main

import (
	"flag"
	"log"
	"strings"

	"github.com/admpub/ddns/store/engine/redis"
	"github.com/webx-top/echo/defaults"
	"github.com/webx-top/echo/engine/standard"
)

const (
	CmdBackend string = "backend"
	CmdWeb     string = "web"
)

var (
	DdnsMode            string
	DdnsDomain          string
	DdnsWebListenSocket string
	DdnsRedisHost       string
	DdnsSoaFqdn         string
	Verbose             bool
)

func init() {
	flag.StringVar(&DdnsMode, "mode", "", "Run Mode")
	flag.StringVar(&DdnsDomain, "domain", "", "The subdomain which should be handled by DDNS")
	flag.StringVar(&DdnsWebListenSocket, "listen", ":8080", "Which socket should the web service use to bind itself")
	flag.StringVar(&DdnsRedisHost, "redis", ":6379", "The Redis socket that should be used")
	flag.StringVar(&DdnsSoaFqdn, "fqdn", "", "The FQDN of the DNS server which is returned as a SOA record")
	flag.BoolVar(&Verbose, "verbose", false, "Be more verbose")
	flag.Parse()
}

func ValidateCommandArgs() {
	if len(DdnsDomain) == 0 {
		log.Fatal("You have to supply the domain via --domain=DOMAIN")
	} else if !strings.HasPrefix(DdnsDomain, ".") {
		// get the domain in the right format
		DdnsDomain = "." + DdnsDomain
	}

	if DdnsMode == CmdBackend {
		if len(DdnsSoaFqdn) == 0 {
			log.Fatal("You have to supply the server FQDN via --fqdn=FQDN")
		}
	}
}

func main() {
	ValidateCommandArgs()

	stor := redis.OpenConnection(DdnsRedisHost)
	defer stor.Close()

	switch DdnsMode {
	case CmdBackend:
		log.Printf("Starting DDNS Backend\n")
		RunBackend(stor)
	case CmdWeb:
		log.Printf("Starting Web Service\n")
		RunWebService(stor, defaults.Default)
		defaults.Default.Run(standard.New(DdnsWebListenSocket))
	default:
		usage()
	}
}

func usage() {
	log.Fatal("Usage: ./ddns --mode=[backend|web]")
}
