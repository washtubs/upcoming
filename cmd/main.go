package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/washtubs/upcoming"
)

var formatStringReference string = `
type Upcoming struct {
	Source       string
	SourceId     string
	Title        string
	InvokeManual string
	When         time.Time
}
`

func main() {
	flag.Parse()
	action := flag.Arg(0)
	if action == "" {
		log.Fatal("Must provide an action: ls | rm ")
	}
	if action == "ls" {
		fs := flag.NewFlagSet("ls", flag.ExitOnError)
		var (
			addresses string
			sources   string
			format    string
		)
		fs.StringVar(&addresses, "addresses", "", "Comma separated list of server addresses")
		fs.StringVar(&sources, "sources", "", "Comma separated list of sources")
		fs.StringVar(&format, "format", "", fmt.Sprintf("Format string. Reference: %s", formatStringReference))
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

		var tmp *template.Template
		if format != "" {
			tmp = template.New("format")
			tmp, err = tmp.Parse(format)
			if err != nil {
				fs.Usage()
				log.Fatal(err)
			}
		}

		out := os.Stdout
		for _, v := range all {
			if format == "" {
				fmt.Fprintln(out, upcoming.Format(v))
			} else {
				tmp.Execute(out, v)
				fmt.Fprintln(out)
			}
		}
	} else if action == "rm" {
		var (
			address  string
			source   string
			sourceId string
		)
		fs := flag.NewFlagSet("rm", flag.ExitOnError)
		fs.StringVar(&source, "source", "", "Remove all for a given source")
		fs.StringVar(&sourceId, "id", "", "Remove the given source ID")
		fs.StringVar(&address, "address", "", "Server address")
		err := fs.Parse(flag.Args()[1:])
		if err != nil {
			log.Fatal(err)
		}
		if source == "" {
			fs.Usage()
			log.Fatal("Please provide a -source")
		}

		client := upcoming.NewClient(address)
		if sourceId != "" {
			client.Remove(source, sourceId)
		} else {
			client.RemoveAll(source)
		}

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
