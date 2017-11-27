package senders

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/hardenedlayer/silrok-feeder-softlayer"
)

// ArchiveSender is...
type ArchiveSender struct {
	Address string
}

// Run implements Sender interface
func (s *ArchiveSender) Run(in chan srfsoftlayer.Message) (chan srfsoftlayer.Message, error) {
	conn, err := net.Dial("udp", s.Address)
	if err != nil {
		return nil, err
	}

	out := make(chan srfsoftlayer.Message)
	go s.run(in, out, conn)
	return out, nil
}

func (s *ArchiveSender) run(in, out chan srfsoftlayer.Message, conn net.Conn) {
	defer conn.Close()

	buffer := bufio.NewWriter(conn)
	for mess := range in {
		if mess.Type == "ticket" {
			bytes, err := json.Marshal(mess.Content)
			if err != nil {
				fmt.Printf("marshalling error: %v\n", err)
			}
			bytes = append(bytes, byte('\n'))
			buffer.Write(bytes)
			buffer.Flush()
			time.Sleep(5 * time.Millisecond)
		} else {
			srfsoftlayer.Inspect("message", mess)
		}
		out <- mess
	}
}

func (s *ArchiveSender) send(mess message) {
}
