package pollers

import (
	"errors"
	"fmt"
	"time"

	"github.com/softlayer/softlayer-go/filter"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"

	"github.com/hardenedlayer/silrok-feeder-softlayer"
)

// constants
const (
	APIEndpoint = "https://api.softlayer.com/rest/v3.1"
)

// TicketPoller is ...
type TicketPoller struct {
	User     string
	APIKey   string
	FetchAll bool
}

// Run is...
func (p *TicketPoller) Run(in chan srfsoftlayer.Message) (chan srfsoftlayer.Message, error) {
	bra, err := p.brandService()
	if err != nil {
		return nil, err
	}

	out := make(chan srfsoftlayer.Message)
	go p.run(bra, in, out)
	return out, nil
}

func (p *TicketPoller) run(service services.Brand, in, out chan srfsoftlayer.Message) {
	loc, _ := time.LoadLocation("US/Central")
	dateNow := time.Now().AddDate(0, 0, 0).In(loc)
	if p.FetchAll {
		dateNow = time.Now().AddDate(-3, -1, 0).In(loc)
	}
	dateStartStr := dateNow.Format("01/02/2006 15:04:05")
	retry := false
	for {
		select {
		case _, ok := <-in:
			if !ok {
				return

			}
		default:
			if p.FetchAll {
				if !retry { // for fetching all mode, add 5 days per iteration.
					dateNow = dateNow.AddDate(0, 0, 5)
				} else {
					dateNow = dateNow.AddDate(0, 0, -2)
				}
				if dateNow.After(time.Now().In(loc)) {
					dateNow = time.Now().In(loc)
					p.FetchAll = false
				}
			} else {
				dateNow = time.Now().In(loc)
			}
			dateEndStr := dateNow.Format("01/02/2006 15:04:05")
			//! for DateBetween, entiries exactly matched with start and end date also included
			data, err := service.
				Mask("attachedVirtualGuests.id;attachedVirtualGuests.hostname;attachedVirtualGuests.domain;attachedVirtualGuests.typeId;attachedVirtualGuests.location.pathString;attachedVirtualGuests.tagReferences;attachedHardware.id;attachedHardware.hostname;attachedHardware.domain;attachedHardware.location.pathString;attachedHardware.tagReferences;id;accountId;assignedUserId;groupId;createDate;lastEditDate;lastEditType;lastResponseDate;locationId;modifyDate;priority;responsibleBrandId;statusId;subjectId;title;firstUpdate.editorType;firstUpdate.editorId;status").
				Filter(filter.Build(
					filter.Path("tickets.createDate").DateBetween(dateStartStr, dateEndStr),
				)).
				GetTickets()

			if err != nil {
				fmt.Printf("ERROR on API call!")
				srfsoftlayer.Inspect("ERROR", err)
				time.Sleep(5 * time.Second)
				retry = true
				continue
			} else {
				for _, ticket := range data {
					if ticket.Title == nil {
						fmt.Printf("OOPS! title is nil\n")
						srfsoftlayer.Inspect("ERROR", ticket)
					} else {
						out <- srfsoftlayer.Message{
							Type:    "ticket",
							Title:   *ticket.Title,
							Content: ticket,
						}
					}
				}
				fmt.Printf("total open tickets: %v from %v to %v\n", len(data), dateStartStr, dateEndStr)
				dateStartStr = dateNow.Add(time.Second).Format("01/02/2006 15:04:05")
				time.Sleep(5 * time.Second)
				retry = false
			}
		}
	}
}

func (p *TicketPoller) brandService() (services.Brand, error) {
	sess := session.New(p.User, p.APIKey)
	sess.Endpoint = "https://api.softlayer.com/rest/v3.1"
	account := services.GetAccountService(sess)
	brands, err := account.GetOwnedBrands()
	if err != nil || len(brands) != 1 {
		return services.Brand{}, errors.New("cannot determined brand")
	}

	service := services.GetBrandService(sess).Id(*brands[0].Id)
	return service, nil
}
