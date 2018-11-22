package main

import (
	"fmt"

	"github.com/hardenedlayer/silrok-feeder-softlayer"
	"github.com/hardenedlayer/silrok-feeder-softlayer/pollers"
	"github.com/hardenedlayer/silrok-feeder-softlayer/senders"
)

func run(opts *Options) error {
	in := make(chan srfsoftlayer.Message)

	var ticketPoller srfsoftlayer.Poller
	ticketPoller = &pollers.TicketPoller{User: opts.SLUser, APIKey: opts.SLAPIKey, FetchAll: opts.FetchAll, From: opts.From}
	mid, err := ticketPoller.Run(in)
	if err != nil {
		fmt.Printf("could not start ticket poller: %v\n", err)
		return err
	}

	var archiveSender srfsoftlayer.Sender
	archiveSender = &senders.ArchiveSender{Address: opts.ArchiveAddress}
	mid, err = archiveSender.Run(mid)
	if err != nil {
		fmt.Printf("could not start archiver: %v\n", err)
		return err
	}

	var slackSender srfsoftlayer.Sender
	slackSender = &senders.SlackSender{HookURL: opts.SlackWebhookURL, ChannelMap: opts.Clients}
	out, err := slackSender.Run(mid)
	if err != nil {
		fmt.Printf("could not start slack sender: %v\n", err)
		return err
	}

	for mess := range out {
		if mess.Type == "control" {
			break
		}
	}
	return nil
}
