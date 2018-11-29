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
	messageRestarted = `@channel *attention please!* silrok feeder for softlayer just (re)started. If this is not planned by you or announced by provider, possibly there was an issue on alerting system. You may need to check your services manually.

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

	for _, channel := range s.ChannelMap {
		s.send(message{
			Channel: channel,
			Level:   "danger",
			Title:   "Red two standing by...",
			Content: messageRestarted,
		})
	}
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
				if len(ticket.AttachedVirtualGuests) > 0 {
					content += "*Affected VSIs*\n\n"
					for _, vm := range ticket.AttachedVirtualGuests {
						ip := "none"
						if vm.PrimaryIpAddress != nil {
							ip = *vm.PrimaryIpAddress
						}
						content += fmt.Sprintf("* %v (%v, %v)\n", *vm.Hostname, *vm.Id, ip)
					}
				}
				if len(ticket.AttachedHardware) > 0 {
					content += "*Affected BMs*\n\n"
					for _, bm := range ticket.AttachedHardware {
						ip := "none"
						if bm.PrimaryIpAddress != nil {
							ip = *bm.PrimaryIpAddress
						}
						content += fmt.Sprintf("* %v (%v, %v)\n", *bm.Hostname, *bm.Id, ip)
					}
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
					fmt.Printf("could not send message for ticket #%v\n", *ticket.Id)
				} else {
					fmt.Printf("slack message for ticket #%v was sent successfully.\n", *ticket.Id)
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
		Channel:  mess.Channel,
		Username: "Hyeoncheon Silrok",
		IconUrl:  "http://hyeoncheon.github.io/images/hyeoncheon-icon.png",
	}

	if mess.Timestamp == 0 {
		mess.Timestamp = time.Now().Unix()
	}

	if mess.TicketID != 0 {
		payload.Attachments = []*slack.Attachment{
			{
				Pretext:   fmt.Sprintf("_New Ticket Issued! %v_", mess.TicketID),
				Color:     mess.Level,
				Title:     mess.Title,
				TitleLink: "https://control.softlayer.com/support/tickets/" + strconv.Itoa(mess.TicketID),
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
				TitleLink: "https://www.cloudz.co.kr/",
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
