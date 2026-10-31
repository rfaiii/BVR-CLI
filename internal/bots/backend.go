package bots

import "context"

// Reply is one streamed piece of a demo response. Emote and Boop are set
// on the first chunk only. Boop asks the UI to play the bot's notify cue.
type Reply struct {
	Text  string
	Emote string
	Boop  bool
}

// Backend lists bots and streams a reply. Send must not dial the network.
// The channel is closed when the reply finishes or ctx is canceled.
type Backend interface {
	List() []Bot
	Send(ctx context.Context, botID, msg string) (<-chan Reply, error)
}
