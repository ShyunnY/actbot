package actors

// Constant definitions related to GitHub labels
const (
	// HelpWantedLabel The value of the help wanted label has been defined
	HelpWantedLabel = "help wanted"
)

// Constant definitions related to GitHub comment reaction
const (
	// CommendReaction The value of the "+1" reaction has been defined
	CommendReaction = "+1"
)

type Actor interface {
	Handler() error

	Capture(event GenericEvent) bool

	Name() string
}

type GenericEvent struct {
	// This represents the actual GitHub events
	Event any
}
