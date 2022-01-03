package handlers

import (
	"github.com/slack-go/slack"
	"log"
)

// HandleInteractionEvent will print info about the interaction
func HandleInteractionEvent(interaction slack.InteractionCallback, client *slack.Client) error {
	// handle the interaction
	// Switch depending on the type
	log.Printf("the action called is: %s\n", interaction.ActionID)
	log.Printf("the response was of type: %s\n", interaction.Type)

	switch interaction.Type {
	case slack.InteractionTypeBlockActions:
		// this is a block action, so we need to handle it

		for _, action := range interaction.ActionCallback.BlockActions {
			log.Printf("%+v", action)
			log.Println("Selected option: ", action.SelectedOptions)
		}
	default:

	}
	return nil
}
