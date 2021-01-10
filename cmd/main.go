package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/washtubs/upcoming"
)

func main() {
	flag.Parse()
	action := flag.Arg(0)
	if action == "" {
		log.Fatal("Must provide an action: ls | rm ")
	}
	if action == "ls" {
		fs := flag.NewFlagSet("ls", flag.ExitOnError)
		var sources string
		fs.StringVar(&sources, "sources", "", "Comma separated list of sources")
		err := fs.Parse(flag.Args()[1:])
		if err != nil {
			log.Fatal(err)
		}

		client := upcoming.DefaultClient()
		list, err := client.List(upcoming.ListOpts{
			Sources: strings.Split(sources, ","),
		})
		if err != nil {
			log.Fatal(err)
		}

		for _, v := range list {
			fmt.Println(upcoming.Format(v))
		}
	} else if action == "rm" {
	} else if action == "debug" {

		d, err := time.ParseDuration("200h")
		if err != nil {
			panic(err)
		}

		fmt.Println(upcoming.HumanizeDuration(d))

		client := upcoming.DefaultClient()
		err = client.Put(upcoming.Upcoming{
			Source:   "test",
			SourceId: "123",
			Title:    "Heyyyyy",
			When:     time.Now().Add(time.Minute * 2),
		})
		if err != nil {
			log.Fatal(err)
		}

		err = client.Put(upcoming.Upcoming{
			Source:   "test",
			SourceId: "124",
			Title:    "Heyyyyy!!!",
			When:     time.Now().Add(time.Minute * 2),
		})
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("Must provide an action: do | add | ls | dismiss ")
	}

}
