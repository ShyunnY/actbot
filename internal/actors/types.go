package actors

// Constant definitions related to GitHub labels
const (
	// HelpWantedLabel The value of the help wanted label has been defined
	HelpWantedLabel = "help wanted"
)

// Constant definitions related to GitHub comment reaction
const (
	// CommendReaction The value of the "+1 👍" reaction has been defined
	CommendReaction = "+1"

	// RocketReaction The value of the "rocket 🚀" reaction has been defined
	RocketReaction = "rocket"
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

type IMClient interface {
	SendMessage(issueNumber int64, content string) error
}

// Options Extended parameter configuration
type Options struct {
	IMClient IMClient
}
