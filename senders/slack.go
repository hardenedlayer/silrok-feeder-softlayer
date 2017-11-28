package senders

import (
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
		fmt.Println("WARN: no webhook URL provided. no alert will be generated!")
	}
	out := make(chan srfsoftlayer.Message)
	go s.run(in, out)
	return out, nil
}

func (s *SlackSender) run(in, out chan srfsoftlayer.Message) {
	for mess := range in {
		if mess.Type == "ticket" {
			accountID := *mess.Content.(datatypes.Ticket).AccountId
			for _, account := range s.Channels {
				if accountID == account {
					ticket := mess.Content.(datatypes.Ticket)
					if *ticket.FirstUpdate.EditorType != "USER" {
						s.send(message{
							Title:     *ticket.Title,
							TicketID:  *ticket.Id,
							Timestamp: ticket.CreateDate.Unix(),
							AccountID: accountID,
							IssuedBy:  *ticket.FirstUpdate.EditorType,
						})
						srfsoftlayer.Inspect("ticket", mess)
					}
				}
			}
		} else {
			srfsoftlayer.Inspect("message", mess)
		}
		out <- mess
	}
}

type message struct {
	Title     string
	TicketID  int
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
		Username:  "Hyeoncheon Silrok",
		IconEmoji: ":hc:",
		Attachments: []*slack.Attachment{
			{
				Pretext:   "_New Ticket Issued!_",
				Color:     "warning",
				Title:     mess.Title,
				TitleLink: "http://alargo.as-a-svc.com/tickets/" + strconv.Itoa(mess.TicketID),
				Fields: []*slack.AttachmentField{
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
				Footer:     "Hyeoncheon",
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
