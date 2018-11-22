package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	getopt "github.com/pborman/getopt/v2"
)

// Options is a structure for running configuration
type Options struct {
	SLUser          string
	SLAPIKey        string
	FetchAll        bool
	From            string
	ArchiveAddress  string
	SlackWebhookURL string
	Clients         map[int]string
}

func main() {
	opts := getOptions()
	if opts == nil {
		os.Exit(1)
	}

	run(opts)
}

func getOptions() *Options {
	help := false
	opts := &Options{}
	getopt.FlagLong(&help, "options", 'o', "Show options")
	getopt.FlagLong(&opts.ArchiveAddress, "addr", 'a', "Address of Archiver")
	getopt.FlagLong(&opts.SlackWebhookURL, "hook", 'h', "Webhook URL for Slack")
	getopt.FlagLong(&opts.SLUser, "user", 'u', "API Username of SoftLayer Brand Account")
	getopt.FlagLong(&opts.SLAPIKey, "apikey", 'k', "API Key of SoftLayer Brand Account")
	getopt.FlagLong(&opts.FetchAll, "fetchall", 's', "Sync All Tickets")
	getopt.FlagLong(&opts.From, "from", 'f', "Start date of fetching (YYYY-MM-DD)")

	getopt.Parse()

	if help {
		getopt.Usage()
		return nil
	}
	opts.Clients = map[int]string{}
	for _, a := range getopt.Args() {
		aa := strings.Split(a, ":")
		if id, err := strconv.Atoi(aa[0]); err == nil {
			if len(aa) != 2 {
				aa = append(aa, "ant"+strconv.Itoa(id))
				fmt.Printf("oops! account argument '%s' does not match with 'account_id:channel_name'.\n", a)
				fmt.Printf("- modified value: %v:%v will be used.\n", id, aa[1])
			}
			opts.Clients[id] = aa[1] // channel name
		} else {
			fmt.Printf("oops! 'account_id' part of the '%s' is not a number.\n", a)
			fmt.Printf("use form 'account_id:channel_name'. ignore!\n")
		}
	}

	if jc, err := json.Marshal(opts.Clients); err == nil {
		fmt.Printf("accounts to alert: %v\n", string(jc))
	}

	return opts
}
