package app_test

import (
	"testing"

	"github.com/luk4z7/notificationhub/app"
	"github.com/luk4z7/notificationhub/app/command"
	"github.com/stretchr/testify/assert"
)

func TestApplicationInitialization(t *testing.T) {
	// Create a PrintHandler instance to be used in the Application
	printHandler := command.NewPrintHandler("testPrintHandler")
	assert.NotNil(t, printHandler) // Using assert for consistency with imports

	// Initialize the Application struct using a struct literal
	application := app.Application{
		Commands: app.Commands{
			Print: printHandler,
		},
	}

	// Assert that the Application and its nested Commands and PrintHandler are not nil
	assert.NotNil(t, application)
	assert.NotNil(t, application.Commands)
	assert.NotNil(t, application.Commands.Print, "PrintHandler should be initialized in Commands")
	assert.Equal(t, printHandler, application.Commands.Print, "The PrintHandler in Application should be the one we initialized")
}
