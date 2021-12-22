package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"log"
	"os"
	"strings"
	"time"
)

func main() {

	// Load Env variables from .dot file
	err := godotenv.Load(".env.yml")
	if err != nil {
		fmt.Println(err)
	}

	token := os.Getenv("SLACK_AUTH_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")

	// Create a new client to slack by giving token
	// Set debug to true while developing
	client := slack.New(token, slack.OptionDebug(true), slack.OptionAppLevelToken(appToken))

	// go-slack comes with SocketMode package that we need to use that accepts a Slack client and outputs a Socket mode client instead
	socketClient := socketmode.New(
		client,
		socketmode.OptionDebug(true),
		// Option to set a custom logger
		socketmode.OptionLog(log.New(os.Stdout, "socketmode", log.Lshortfile|log.LstdFlags)),
	)

	// Create a context that can be used to cancel goroutine
	ctx, cancel := context.WithCancel(context.Background())
	// Make this cancel called properly in a real program, graceful shutdown
	defer cancel()

	go func(ctx context.Context, client *slack.Client, socketClient *socketmode.Client) {
		// Create a for loop that selects either the context cancellation or the events
		for {
			select {
			// ins-case context cancel is called exit the goroutine
			case <-ctx.Done():
				log.Println("Shutting down socket listener")
				return
			case event := <-socketClient.Events:
				// We have a new Events, let's type the switch event
				// Add more use cases here if you want to listen to other events
				switch event.Type {
				// Handle EventAPI events
				case socketmode.EventTypeEventsAPI:
					// The event sent on the channel is not the same as the EventAPI events, so we need to type cast it
					eventAPIEvent, ok := event.Data.(slackevents.EventsAPIEvent)
					if !ok {
						log.Printf("Could not type cast the event to the EventsAPIEvent: %v\n", event)
						continue
					}
					// New we have an Events API event, but this event type can in turn be many types,
					// so we actually need another type switch
					err := handleEventMessage(eventAPIEvent, client)
					if err != nil {
						// Replace with actual handling
						log.Fatal(err)
					}
					// We need to send an Acknowledgment to the Slack Server
					socketClient.Ack(*event.Request)

				// Handle Slash Commands
				case socketmode.EventTypeSlashCommand:
					// Just like before, type cast to the correct event type
					command, ok := event.Data.(slack.SlashCommand)
					if !ok {
						log.Printf("Could not type cast the message to a SlashCommand: %v\n", command)
						continue
					}
					// handleSlashCommand will take care of the command
					payload, err := handleSlashCommand(command, client)
					if err != nil {
						log.Fatal(err)
					}
					// Acknowledge the request
					socketClient.Ack(*event.Request, payload)

				case socketmode.EventTypeInteractive:
					interaction, ok := event.Data.(slack.InteractionCallback)
					if !ok {
						log.Printf("Could not type cast the message to a Interaction callback: %v\n", interaction)
						continue
					}

					err := handleInteractionEvent(interaction, client)
					if err != nil {
						log.Fatal(err)
					}
					socketClient.Ack(*event.Request)
				}
			}
		}
	}(ctx, client, socketClient)

	err = socketClient.Run()
	if err != nil {
		fmt.Println(err)
	}
}

// handleEventMessage will take an event and handle it properly based on the type of event
func handleEventMessage(event slackevents.EventsAPIEvent, client *slack.Client) error {
	switch event.Type {
	// First we check if this is a CallbackEvent
	case slackevents.CallbackEvent:
		innerEvent := event.InnerEvent
		// Yet Another Type switch on the actual Data to see if it is an AppMentionEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			// The applications have been mentioned since this Event is a Mention event
			err := handleAppMentionEvent(ev, client)
			if err != nil {
				return err
			}

		}
	default:
		return errors.New("unsupported event type")
	}
	return nil
}

// handleAppMentionEvent is used to take care of the AppMentionEvent when the bot is mentioned
func handleAppMentionEvent(event *slackevents.AppMentionEvent, client *slack.Client) error {
	// Grab the name based on the ID of the one who mentioned the bot
	user, err := client.GetUserInfo(event.User)
	if err != nil {
		return err
	}
	// Check if the user said Hello to the bot
	text := strings.ToLower(event.Text)

	// Create the attachment and assigned based on the message
	attachment := slack.Attachment{}
	// Add some default context like user who mentioned the bot
	attachment.Fields = []slack.AttachmentField{
		{
			Title: "Data",
			Value: time.Now().String(),
		}, {
			Title: "Initializer",
			Value: user.Name,
		},
	}
	if strings.Contains(text, "hello") {
		// Greet the user
		attachment.Text = fmt.Sprintf("How can I help you %s?", user.Name)
		attachment.Pretext = "Greetings!"
		attachment.Color = "#4ad030"
	} else {
		// Send a message to the user
		attachment.Text = fmt.Sprintf("How can I help you %s?", user.Name)
		attachment.Pretext = "How can I bee of service?"
		attachment.Color = "#3d3d3d"
	}
	// Send the message to the channel
	// The Channel is available in the event message
	_, _, err = client.PostMessage(event.Channel, slack.MsgOptionAttachments(attachment))
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}
	return nil
}

// handleSlashCommand will take a slash command and route to the appropriate function
func handleSlashCommand(command slack.SlashCommand, client *slack.Client) (interface{}, error) {
	// Switch depending on the command given
	switch command.Command {
	case "/hello":
		// Hello command, pass it to the proper function
		return nil, handleHelloCommand(command, client)
	case "/mom-gay":
		return handleMom(command, client)
	}
	return nil, nil
}

// handleHelloCommand will take /hello submissions
func handleHelloCommand(command slack.SlashCommand, client *slack.Client) error {
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

// handleMom will trigger a Yes/No question to the initializer
func handleMom(command slack.SlashCommand, client *slack.Client) (interface{}, error) {
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

// handleInteractionEvent will print info about the interaction
func handleInteractionEvent(interaction slack.InteractionCallback, client *slack.Client) error {
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
