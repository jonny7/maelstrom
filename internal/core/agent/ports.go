package agent

// Member is a node in the Maelstrom cluster.
type Member struct {
	Name   string
	Addr   string
	Port   int
	Tags   map[string]string
	Status string
}

// Membership is the driven port for cluster membership.
type Membership interface {
	Members() []Member
	Leave() error
}
