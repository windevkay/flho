package main

import (
	"errors"
	"net/http"

	// "time"

	"github.com/windevkay/flhoutils/helpers"

	"github.com/windevkay/flho/identity_service/internal/data"
	errs "github.com/windevkay/flhoutils/errors"
	"github.com/windevkay/flhoutils/validator"
)

func (app *application) registerIdentityHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := helpers.ReadJSON(w, r, &input)
	if err != nil {
		errs.BadRequestResponse(w, r, err)
		return
	}

	identity := &data.Identity{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = identity.Password.Set(input.Password)
	if err != nil {
		errs.ServerErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateIdentity(v, identity); !v.Valid() {
		errs.FailedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Identities.Insert(identity)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "address already in use")
			errs.FailedValidationResponse(w, r, v.Errors)
		default:
			errs.ServerErrorResponse(w, r, err)
		}
		return
	}

	// generate user activation token
	//token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	// if err != nil {
	// 	errs.ServerErrorResponse(w, r, err)
	// 	return
	// }

	// helpers.RunInBackground(func() {
	// 	data := map[string]any{
	// 		"activationToken": token.Plaintext,
	// 		"name":            user.Name,
	// 	}
	// 	// publish to rabbitMQ for mailer service to consume
	// 	err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
	// 	if err != nil {
	// 		app.logger.Error(err.Error())
	// 	}
	// }, &app.wg)

	helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"identity": identity}, nil)
}

func (app *application) activateIdentityHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := helpers.ReadJSON(w, r, &input)
	if err != nil {
		errs.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		errs.FailedValidationResponse(w, r, v.Errors)
		return
	}

	identity, err := app.models.Identities.GetIdentityForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			errs.FailedValidationResponse(w, r, v.Errors)
		default:
			errs.ServerErrorResponse(w, r, err)
		}
		return
	}

	identity.Activated = true

	err = app.models.Identities.Update(identity)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			errs.EditConflictResponse(w, r)
		default:
			errs.ServerErrorResponse(w, r, err)
		}
		return
	}

	err = app.models.Tokens.DeleteScopeTokensForIdentity(data.ScopeActivation, identity.ID)
	if err != nil {
		errs.ServerErrorResponse(w, r, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"identity": identity}, nil)
}
