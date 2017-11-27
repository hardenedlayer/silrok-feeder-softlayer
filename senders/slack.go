package senders

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/bluele/slack"
	"github.com/softlayer/softlayer-go/datatypes"

	"github.com/hardenedlayer/silrok-feeder-softlayer"
)

// SlackSender is...
type SlackSender struct {
	HookURL  string
	Channels []int
}

// Run implements Sender interface
func (s *SlackSender) Run(in chan srfsoftlayer.Message) (chan srfsoftlayer.Message, error) {
	if s.HookURL == "" {
		return nil, errors.New("webhook URL could not be null")
	}
	out := make(chan srfsoftlayer.Message)
	go s.run(in, out)
	return out, nil
}

func (s *SlackSender) run(in, out chan srfsoftlayer.Message) {
	for mess := range in {
		for _, account := range s.Channels {
			if mess.Type == "ticket" {
				accountID := *mess.Content.(datatypes.Ticket).AccountId
				if accountID == account {
					ticket := mess.Content.(datatypes.Ticket)
					s.send(message{
						Title:     *ticket.Title,
						Timestamp: ticket.CreateDate.Unix(),
						AccountID: accountID,
						IssuedBy:  *ticket.FirstUpdate.EditorType,
					})
					srfsoftlayer.Inspect("ticket", mess)
				}
			} else {
				srfsoftlayer.Inspect("message", mess)
			}
		}
		out <- mess
	}
}

type message struct {
	Title     string
	Timestamp int64
	Account   string
	AccountID int
	IssuedBy  string
	Content   string
}

func (s *SlackSender) send(mess message) {
	hook := slack.NewWebHook(s.HookURL)
	err := hook.PostMessage(&slack.WebHookPostPayload{
		Channel:   fmt.Sprintf("#ant%v", mess.AccountID),
		Username:  "HardenedLayer Alargo",
		IconEmoji: ":hc:",
		Attachments: []*slack.Attachment{
			{
				Pretext:   "*Attention!* _Automated Ticket Issued!_",
				Fallback:  "Attention! Automated Ticket Issued!",
				Color:     "warning",
				Title:     "Click to see Alargo Alert Details",
				TitleLink: "http://alargo.as-a-svc.com",
				Fields: []*slack.AttachmentField{
					{
						Title: "Title",
						Value: mess.Title,
					},
					{
						Title: "Account",
						Value: strconv.Itoa(mess.AccountID),
						Short: true,
					},
					{
						Title: "Issued by",
						Value: mess.IssuedBy,
						Short: true,
					},
				},
				Footer:     "Alargo",
				FooterIcon: "http://hyeoncheon.github.io/images/hyeoncheon-icon.png",
				TimeStamp:  mess.Timestamp,
				MarkdownIn: []string{"text", "pretext", "fields"},
			},
		},
	})
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
}
