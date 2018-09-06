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
	//! in SL API, all date values should be represented as US/Central. @.@
	loc, _ := time.LoadLocation("US/Central")

	// initial start date for between filter: 10s before starting.
	dateNow := time.Now().Add(-10 * time.Second)
	if p.FetchAll {
		dateNow = time.Now().AddDate(-3, 0, 0)
	}
	dateLast := dateNow
	dateStartStr := dateNow.In(loc).Format("01/02/2006 15:04:05")

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
					// shorten period. sometimes api call fails when too many tickets.
					dateNow = dateNow.AddDate(0, 0, -2)
					retry = false
				}

				if dateNow.After(time.Now()) {
					dateNow = time.Now()
					p.FetchAll = false
				}
			} else { // end date for between filter
				dateNow = time.Now()
			}

			dateStartStr = dateLast.Add(1 * time.Second).In(loc).Format("01/02/2006 15:04:05")
			dateEndStr := dateNow.In(loc).Format("01/02/2006 15:04:05")
			//! for DateBetween, entiries exactly matched with start and end date also included
			data, err := service.
				Mask("attachedVirtualGuests.id;attachedVirtualGuests.hostname;attachedVirtualGuests.domain;attachedVirtualGuests.typeId;attachedVirtualGuests.location.pathString;attachedVirtualGuests.tagReferences;attachedHardware.id;attachedHardware.hostname;attachedHardware.domain;attachedHardware.location.pathString;attachedHardware.tagReferences;id;accountId;assignedUserId;groupId;createDate;lastEditDate;lastEditType;lastResponseDate;locationId;modifyDate;priority;responsibleBrandId;statusId;subjectId;title;firstUpdate.editorType;firstUpdate.editorId;status").
				Filter(filter.Build(
					filter.Path("tickets.createDate").DateBetween(dateStartStr, dateEndStr),
				)).
				GetTickets()

			if err != nil {
				fmt.Printf("ERROR on API call! sleep 5s and continue.")
				srfsoftlayer.Inspect("ERROR", err)
				time.Sleep(5 * time.Second)
				retry = true
				continue
			}

			fmt.Printf("total open tickets: %v from %v to %v\n", len(data), dateStartStr, dateEndStr)
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
					fmt.Printf("- ticket #%v create: %v (last: %v", *ticket.Id, ticket.CreateDate.Time, dateLast)
					if ticket.CreateDate.Time.After(dateLast) {
						dateLast = ticket.CreateDate.Time
					}
					fmt.Printf(" -> %v\n", dateLast)
				}
			}
			time.Sleep(5 * time.Second)
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
