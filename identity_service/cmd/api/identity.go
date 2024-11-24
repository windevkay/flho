package main

import (
	"net/http"

	"github.com/windevkay/flhoutils/helpers"

	"github.com/windevkay/flho/identity_service/internal/services"
	errs "github.com/windevkay/flhoutils/errors"
)

func (app *application) registerIdentityHandler(w http.ResponseWriter, r *http.Request) {
	var input services.RegisterIdentityInput

	err := helpers.ReadJSON(w, r, &input)
	if err != nil {
		errs.BadRequestResponse(w, r, err)
		return
	}

	is := services.NewIdentityService(identityServiceConfig)
	identity, err := is.RegisterIdentity(input, w, r)
	if err != nil {
		app.logger.Error(err.Error())
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"identity": identity}, nil)
}

func (app *application) activateIdentityHandler(w http.ResponseWriter, r *http.Request) {
	var input services.ActivateIdentityInput

	err := helpers.ReadJSON(w, r, &input)
	if err != nil {
		errs.BadRequestResponse(w, r, err)
		return
	}

	is := services.NewIdentityService(identityServiceConfig)
	identity, err := is.ActivateIdentity(input, w, r)
	if err != nil {
		app.logger.Error(err.Error())
		return
	}

	helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"identity": identity}, nil)
}
