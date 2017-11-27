package srfsoftlayer

// Poller is an interface for event pollers
type Poller interface {
	Run(chan Message) (chan Message, error)
}
