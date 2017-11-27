package main

import (
	"fmt"
	"os"
	"strconv"

	getopt "github.com/pborman/getopt/v2"
)

// Options is a structure for running configuration
type Options struct {
	SLUser          string
	SLAPIKey        string
	FetchAll        bool
	ArchiveAddress  string
	SlackWebhookURL string
	Clients         []int
}

func main() {
	opts := getOptions()
	if opts == nil {
		os.Exit(1)
	}

	run(opts)
}

func getOptions() *Options {
	opts := &Options{}
	getopt.FlagLong(&opts.ArchiveAddress, "addr", 'a', "Address of Archiver")
	getopt.FlagLong(&opts.SlackWebhookURL, "hook", 'h', "Webhook URL for Slack")
	getopt.FlagLong(&opts.SLUser, "user", 'u', "API Username of SoftLayer Brand Account")
	getopt.FlagLong(&opts.SLAPIKey, "apikey", 'k', "API Key of SoftLayer Brand Account")
	getopt.FlagLong(&opts.FetchAll, "fetchall", 's', "Sync All Tickets")

	getopt.Parse()
	for _, a := range getopt.Args() {
		if c, err := strconv.Atoi(a); err == nil {
			opts.Clients = append(opts.Clients, c)
		}
	}

	fmt.Println("clients:", opts.Clients)

	return opts
}
