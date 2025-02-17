package main

import (
	"errors"
	"net/http"

	"github.com/windevkay/flhoutils/helpers"

	"github.com/windevkay/flho/identity_service/internal/data"
	"github.com/windevkay/flho/identity_service/internal/queue"
	"github.com/windevkay/flho/identity_service/internal/services"
	errs "github.com/windevkay/flhoutils/errors"
)

// @Summary Register new identity
// @Description Register a new identity in the system
// @Tags identity
// @Accept json
// @Produce json
// @Param request body services.RegisterIdentityInput true "Identity registration details"
// @Success 201 {object} helpers.Envelope{identity=data.Identity}
// @Failure 400 {object} helpers.Envelope{error=string}
// @Failure 422 {object} helpers.Envelope{error=map[string]string}
// @Failure 500 {object} helpers.Envelope{error=string}
// @Router /v1/users [post]
func (app *application) registerIdentityHandler(w http.ResponseWriter, r *http.Request) {
	var input services.RegisterIdentityInput

	err := helpers.ReadJSON(w, r, &input)
	if err != nil {
		errs.BadRequestResponse(w, r, err)
		return
	}

	is := services.NewIdentityService(&serviceConfig, queue.SendMessage, helpers.RunInBackground)
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

// @Summary Activate identity
// @Description Activate a registered identity
// @Tags identity
// @Accept json
// @Produce json
// @Param request body services.ActivateIdentityInput true "Identity activation details"
// @Success 200 {object} helpers.Envelope{identity=data.Identity}
// @Failure 400 {object} helpers.Envelope{error=string}
// @Failure 409 {object} helpers.Envelope{error=string}
// @Failure 422 {object} helpers.Envelope{error=map[string]string}
// @Failure 500 {object} helpers.Envelope{error=string}
// @Router /v1/users/activate [put]
func (app *application) activateIdentityHandler(w http.ResponseWriter, r *http.Request) {
	var input services.ActivateIdentityInput

	err := helpers.ReadJSON(w, r, &input)
	if err != nil {
		errs.BadRequestResponse(w, r, err)
		return
	}

	is := services.NewIdentityService(&serviceConfig, queue.SendMessage, helpers.RunInBackground)
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
