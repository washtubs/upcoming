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
		var addresses string
		var sources string
		fs.StringVar(&addresses, "addresses", "", "Comma separated list of server addresses")
		fs.StringVar(&sources, "sources", "", "Comma separated list of sources")
		err := fs.Parse(flag.Args()[1:])
		if err != nil {
			log.Fatal(err)
		}

		all := make([]upcoming.Upcoming, 0)
		for _, addr := range strings.Split(addresses, ",") {
			client := upcoming.NewClient(addr)
			list, err := client.List(upcoming.ListOpts{
				Sources: strings.Split(sources, ","),
			})
			all = append(all, list...)
			if err != nil {
				log.Fatal(err)
			}
		}
		upcoming.SortByDuration(all)
		for _, v := range all {
			fmt.Println(upcoming.Format(v))
		}
	} else if action == "rm" {
	} else if action == "debug" {

		d, err := time.ParseDuration("200h")
		if err != nil {
			panic(err)
		}

		fmt.Println(upcoming.HumanizeDuration(d))

		client := upcoming.NewClient("")
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
