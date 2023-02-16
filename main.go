package main

import (
	"errors"
	"launcher/launch"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/tidwall/gjson"
)

func init() {
	log.SetPrefix("launcher: ")
}

func main() {
	log.Println("Finding .vscode/launch.json")
	if len(os.Args) != 2 {
		Must(errors.New("expected one positional argument"))
	}
	basePath := os.Args[1]

	launchConf, err := os.ReadFile(filepath.Join(basePath, ".vscode/launch.json"))
	Must(err)

	launchConfJson := gjson.ParseBytes(launchConf)
	allConfs := launchConfJson.Get("configurations").Array()

	log.Println("checking for golang configuration")
	var finalLaunchConf *gjson.Result
	for _, conf := range allConfs {
		if conf.Get("name").String() == "Launch Package" &&
			conf.Get("type").String() == "go" &&
			conf.Get("request").String() == "launch" {
			finalLaunchConf = &conf
		}
	}
	if finalLaunchConf == nil {
		Must(errors.New("golang configuration not found"))
	}

	lc, err := launch.LaunchConfigurationFromString(finalLaunchConf.String())
	Must(err)
	proc, err := launch.Start(lc)
	Must(err)
	log.Println("Started process: ", proc.Pid)

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	incoming := <-osSignals
	log.Println("forwarding process: ", proc.Pid)
	proc.Signal(incoming)
}

func Must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
