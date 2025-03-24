package main

import (
	"errors"
	"net/http"

	"github.com/windevkay/flhoutils/helpers"

	"github.com/windevkay/flho/internal/data"
	"github.com/windevkay/flho/internal/services"
	errs "github.com/windevkay/flhoutils/errors"
)

func (app *application) registerIdentityHandler(w http.ResponseWriter, r *http.Request) {
	var input services.RegisterIdentityInput

	err := helpers.ReadJSON(w, r, &input)
	if err != nil {
		errs.BadRequestResponse(w, r, err)
		return
	}

	is := services.IdentityService{ServiceConfig: &serviceConfig}
	identity, err := is.RegisterIdentity(input)
	if err != nil {
		switch {
		case errors.Is(err.(*services.ValidationErr).Err, data.ErrValidationFailed):
			errs.FailedValidationResponse(w, r, err.(*services.ValidationErr).Fields)
		case errors.Is(err.(*services.ValidationErr).Err, data.ErrDuplicateEmail):
			errs.FailedValidationResponse(w, r, err.(*services.ValidationErr).Fields)
		default:
			errs.ServerErrorResponse(w, r, err)
		}
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

	is := services.IdentityService{ServiceConfig: &serviceConfig}
	identity, err := is.ActivateIdentity(input)
	if err != nil {
		switch {
		case errors.Is(err.(*services.ValidationErr).Err, data.ErrValidationFailed):
			errs.FailedValidationResponse(w, r, err.(*services.ValidationErr).Fields)
		case errors.Is(err.(*services.ValidationErr).Err, data.ErrRecordNotFound):
			errs.FailedValidationResponse(w, r, err.(*services.ValidationErr).Fields)
		case errors.Is(err, data.ErrEditConflict):
			errs.EditConflictResponse(w, r)
		default:
			errs.ServerErrorResponse(w, r, err)
		}
		return
	}

	helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"identity": identity}, nil)
}
