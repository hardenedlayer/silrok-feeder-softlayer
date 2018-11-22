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

const (
	messageRestarted = `@channel *attention please!* silrok feeder for softlayer just (re)started. If this is not planned by you or announced by provider, possibly there was an issue on alerting system. You need to check your services manually.

_May the force be with you..._`
)

// SlackSender is...
type SlackSender struct {
	HookURL      string
	ChannelMap   map[int]string
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
	s.send(message{
		Channel: "#general",
		Level:   "danger",
		Title:   "Red two standing by...",
		Content: messageRestarted,
	})
	return out, nil
}

func (s *SlackSender) run(in, out chan srfsoftlayer.Message) {
	for mess := range in {
		if mess.Type == "ticket" {
			accountID := *mess.Content.(datatypes.Ticket).AccountId
			for id, channel := range s.ChannelMap {
				if accountID != id {
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

				err := s.send(message{
					Channel:   fmt.Sprintf("#%v", channel),
					Level:     "warning",
					Title:     *ticket.Title,
					TicketID:  *ticket.Id,
					Timestamp: ticket.CreateDate.Unix(),
					AccountID: accountID,
					IssuedBy:  *ticket.FirstUpdate.EditorType,
					Content:   content,
				})
				if err != nil {
					fmt.Printf("could not send message for ticket #%v", *ticket.Id)
				} else {
					fmt.Printf("slack message for ticket #%v was sent successfully.", *ticket.Id)
				}
			}
		} else {
			srfsoftlayer.Inspect("message", mess)
		}
		out <- mess
	}
}

type message struct {
	Channel   string
	Level     string
	Title     string
	Timestamp int64

	TicketID  int
	Account   string
	AccountID int
	IssuedBy  string
	Content   string
}

func (s *SlackSender) send(mess message) error {
	hook := slack.NewWebHook(s.HookURL)
	payload := slack.WebHookPostPayload{
		Channel:   mess.Channel,
		Username:  "Hyeoncheon Silrok",
		IconEmoji: ":hc:",
	}

	if mess.Timestamp == 0 {
		mess.Timestamp = time.Now().Unix()
	}

	if mess.TicketID != 0 {
		payload.Attachments = []*slack.Attachment{
			{
				Pretext:   "_New Ticket Issued!_",
				Color:     mess.Level,
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
		}
	} else {
		payload.Attachments = []*slack.Attachment{
			{
				Color:     mess.Level,
				Title:     mess.Title,
				TitleLink: "http://alargo.as-a-svc.com/",
				Fields: []*slack.AttachmentField{
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
		}
	}

	if err := hook.PostMessage(&payload); err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}
	return nil
}
