package app

import "github.com/luk4z7/notificationhub/app/command"

type Application struct {
	Commands Commands
}

type Commands struct {
	Print *command.PrintHandler
}
