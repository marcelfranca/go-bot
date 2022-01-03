package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"log"
	"os"
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
					err := HandleEventMessage(eventAPIEvent, client)
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
					// HandleSlashCommand will take care of the command
					payload, err := HandleSlashCommand(command, client)
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

					err := HandleInteractionEvent(interaction, client)
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
