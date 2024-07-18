package main

import (
	"net/http"

	"github.com/windevkay/flho/mailer_service/internal/mailer"
	errs "github.com/windevkay/flhoutils/errors"
	"github.com/windevkay/flhoutils/helpers"
)

func (app *application) sendEmail(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Recipient string `json:"recipient"`
		File      string `json:"file"`
		Data      any    `json:"data"`
	}

	err := helpers.ReadJSON(w, r, &input)
	if err != nil {
		errs.BadRequestResponse(w, r, err)
		return
	}

	email := &mailer.Email{
		Recipient: input.Recipient,
		File:      input.File,
		Data:      input.Data,
	}

	err = app.mailer.Send(email.Recipient, email.File, email.Data)
	if err != nil {
		errs.ServerErrorResponse(w, r, err)
		return
	}
}
