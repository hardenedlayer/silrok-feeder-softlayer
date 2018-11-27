package srfsoftlayer

// Gatherer is an interface for event gatherers
type Gatherer interface {
	Run(chan Message) (chan Message, error)
}
