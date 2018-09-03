package senders

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bluele/slack"
	"github.com/softlayer/softlayer-go/datatypes"

	"github.com/hardenedlayer/silrok-feeder-softlayer"
)

// SlackSender is...
type SlackSender struct {
	HookURL      string
	Channels     []int
	TitleIgnored []string
	RefreshedAt  time.Time
}

// Run implements Sender interface
func (s *SlackSender) Run(in chan srfsoftlayer.Message) (chan srfsoftlayer.Message, error) {
	if s.HookURL == "" {
		fmt.Println("WARN: no webhook URL provided. no alert will be generated!")
		return in, nil
	}

	// fill up table for ignored titles
	s.TitleIgnored = append(s.TitleIgnored, "Customer Upgrade Request")
	s.TitleIgnored = append(s.TitleIgnored, "Virtual Server Cancellation -")
	s.TitleIgnored = append(s.TitleIgnored, "Service Cancellation -")
	s.TitleIgnored = append(s.TitleIgnored, "Server Cancellation -")
	s.TitleIgnored = append(s.TitleIgnored, "[ABNS] ")
	s.RefreshedAt = time.Now()

	out := make(chan srfsoftlayer.Message)
	go s.run(in, out)
	return out, nil
}

func (s *SlackSender) run(in, out chan srfsoftlayer.Message) {
	for mess := range in {
		if mess.Type == "ticket" {
			accountID := *mess.Content.(datatypes.Ticket).AccountId
			for _, account := range s.Channels {
				if accountID != account {
					continue
				}

				ticket := mess.Content.(datatypes.Ticket)
				if *ticket.FirstUpdate.EditorType == "USER" {
					continue
				}

				ignorable := false
				for _, ign := range s.TitleIgnored {
					if strings.HasPrefix(*ticket.Title, ign) {
						fmt.Printf("ignore title: %v\n", *ticket.Title)
						ignorable = true
						break
					}
				}
				if ignorable {
					continue
				}

				content := ""
				if len(ticket.AttachedVirtualGuests) == 1 {
					content += "Hostname: " + *ticket.AttachedVirtualGuests[0].Hostname
				}
				if len(ticket.AttachedHardware) == 1 {
					content += "Hostname: " + *ticket.AttachedHardware[0].Hostname
				}

				s.send(message{
					Title:     *ticket.Title,
					TicketID:  *ticket.Id,
					Timestamp: ticket.CreateDate.Unix(),
					AccountID: accountID,
					IssuedBy:  *ticket.FirstUpdate.EditorType,
					Content:   content,
				})
				// remove later
				srfsoftlayer.Inspect("ticket", mess)
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
					{
						Value: mess.Content,
						Short: false,
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
