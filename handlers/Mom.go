package handlers

import "github.com/slack-go/slack"

// HandleMom will trigger a Yes/No question to the initializer
func HandleMom(command slack.SlashCommand, client *slack.Client) (interface{}, error) {
	// Create the attachment and assigned based on the message
	attachment := slack.Attachment{}

	// Create the checkbox element
	checkbox := slack.NewCheckboxGroupsBlockElement("answer",
		slack.NewOptionBlockObject("yes", &slack.TextBlockObject{Text: "Yes", Type: slack.MarkdownType}, &slack.TextBlockObject{Text: "Really?", Type: slack.MarkdownType}),
		slack.NewOptionBlockObject("no", &slack.TextBlockObject{Text: "No", Type: slack.MarkdownType}, &slack.TextBlockObject{Text: "You think?", Type: slack.MarkdownType}),
	)
	// Create the Accessory that will be included in the Block and the checkbox to ir
	accessory := slack.NewAccessory(checkbox)
	// Add Blocks to the attachment
	attachment.Blocks = slack.Blocks{
		BlockSet: []slack.Block{
			// Create a new section block element and add some text to the accessory to it
			slack.NewSectionBlock(
				&slack.TextBlockObject{
					Type: slack.MarkdownType,
					Text: "Does your mom knows you are gay?",
				},
				nil,
				accessory,
			),
		},
	}
	attachment.Text = "Dafuq is this?"
	attachment.Color = "#4af030"
	return attachment, nil
}
