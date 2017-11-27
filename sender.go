package srfsoftlayer

// Sender is an interface for event senders
type Sender interface {
	Run(chan Message) (chan Message, error)
}
