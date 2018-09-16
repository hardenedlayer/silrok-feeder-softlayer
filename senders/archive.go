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
	conn    net.Conn
	buff    *bufio.Writer
}

// Run implements Sender interface
func (s *ArchiveSender) Run(in chan srfsoftlayer.Message) (chan srfsoftlayer.Message, error) {
	if s.Address == "" {
		fmt.Println("WARN: no archive address provided. skip archiving!")
		return in, nil
	}

	if err := s.dial(); err != nil {
		return nil, err
	}

	out := make(chan srfsoftlayer.Message)
	go s.run(in, out)
	return out, nil
}

func (s *ArchiveSender) run(in, out chan srfsoftlayer.Message) {
	defer s.conn.Close()

	for mess := range in {
		if mess.Type == "ticket" {
			bytes, err := json.Marshal(mess.Content)
			if err != nil {
				fmt.Printf("marshalling error: %v\n", err)
			}
			bytes = append(bytes, byte('\n'))

			for i := 0; i < 3; i++ {
				if err := s.write(bytes); err != nil {
					// TODO improve it much safe.
					fmt.Printf("try to reconnection...\n")
					if err := s.dial(); err != nil {
						fmt.Printf("try %v: reconnection error: %v\n", i+1, err)
					} else {
						fmt.Printf("try %v: connected!\n", i+1)
					}
				} else {
					break
				}
			}
			time.Sleep(5 * time.Millisecond)
		} else {
			srfsoftlayer.Inspect("message", mess)
		}
		out <- mess
	}
}

func (s *ArchiveSender) write(b []byte) error {
	nn, err := s.buff.Write(b)
	if err != nil {
		fmt.Printf("-- writer write error: %d, %v\n", nn, err)
	} else {
		err = s.buff.Flush()
		if err != nil {
			fmt.Printf("-- writer flush error: %v\n", err)
		}
	}
	return err
}

func (s *ArchiveSender) dial() (err error) {
	s.conn, err = net.Dial("udp", s.Address)
	if err != nil {
		fmt.Printf("-- dial error: %v", err)
	}
	s.buff = bufio.NewWriter(s.conn)
	return
}

func (s *ArchiveSender) send(mess message) {
}
