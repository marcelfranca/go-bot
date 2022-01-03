package handlers

import (
	"fmt"
	"github.com/slack-go/slack"
	"time"
)

// HandleHelloCommand will take /hello submissions
func HandleHelloCommand(command slack.SlashCommand, client *slack.Client) error {
	// The Input is found in the text field
	// Create the attachment and assigned based on the message
	attachment := slack.Attachment{}
	// Add Some default Context like the user who mentioned
	attachment.Fields = []slack.AttachmentField{
		{
			Title: "Date",
			Value: time.Now().String(),
		}, {
			Title: "Initializer",
			Value: command.UserName,
		},
	}

	// Greet the user
	attachment.Text = fmt.Sprintf("Hello %s", command.Text)
	attachment.Color = "#4af030"

	// Send the message to the channel
	// The Channel is available in the command.ChannelID
	_, _, err := client.PostMessage(command.ChannelID, slack.MsgOptionAttachments(attachment))
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}
	return nil
}
