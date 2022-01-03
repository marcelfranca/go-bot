package handlers

import "github.com/slack-go/slack"

// HandleSlashCommand will take a slash command and route to the appropriate function
func HandleSlashCommand(command slack.SlashCommand, client *slack.Client) (interface{}, error) {
	// Switch depending on the command given
	switch command.Command {
	case "/hello":
		// Hello command, pass it to the proper function
		return nil, HandleHelloCommand(command, client)
	case "/mom-gay":
		return HandleMom(command, client)
	}
	return nil, nil
}
