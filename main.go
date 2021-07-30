package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/reegnz/sleeper/pkg"
)

func main() {
	at := flag.String("at", "", "When to run the command")
	cmd := flag.String("cmd", "", "The command to run after a certain time")
	flag.Parse()

	if len(*at) == 0 || len(*cmd) == 0 {
		fmt.Println("missing arguments")
		flag.PrintDefaults()
		os.Exit(1)
	}
	dur, err := time.Parse(time.RFC3339, *at)
	if err != nil {
		fmt.Println("invalid argument: at")
		flag.PrintDefaults()
		os.Exit(1)
	}

	cmds := strings.Split(*cmd, " ")
	objects := make(chan *pkg.ScheduledJobs)
	go func() {
		objects <- &pkg.ScheduledJobs{
			At:      dur,
			Command: cmds,
		}
		close(objects)
	}()
	strings := pkg.Sleep(objects)
	pkg.Printer(strings)
}
