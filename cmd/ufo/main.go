package main

import (
	"log"
	"os"

	"github.com/btwiuse/multicall"
	"github.com/webteleport/caddy-webteleport/apps/caddy"

	_ "github.com/webteleport/utils/hack/quic-go-disable-receive-buffer-warning"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	err := Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}

var cmdRun multicall.RunnerFuncMap = map[string]multicall.RunnerFunc{
	"caddy":        caddy.Run,
}

func Run(args []string) error {
	return cmdRun.Run(os.Args[1:])
}
