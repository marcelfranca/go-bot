package handlers

import (
	"errors"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// HandleEventMessage will take an event and handle it properly based on the type of event
func HandleEventMessage(event slackevents.EventsAPIEvent, client *slack.Client) error {
	switch event.Type {
	// First we check if this is a CallbackEvent
	case slackevents.CallbackEvent:
		innerEvent := event.InnerEvent
		// Yet Another Type switch on the actual Data to see if it is an AppMentionEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			// The applications have been mentioned since this Event is a Mention event
			err := HandleAppMentionEvent(ev, client)
			if err != nil {
				return err
			}

		}
	default:
		return errors.New("unsupported event type")
	}
	return nil
}
