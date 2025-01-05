package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/pascaldekloe/jwt"
	"github.com/windevkay/flho/identity_service/internal/data"
	errs "github.com/windevkay/flhoutils/errors"
	"github.com/windevkay/flhoutils/helpers"
	"github.com/windevkay/flhoutils/validator"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := helpers.ReadJSON(w, r, &input)
	if err != nil {
		errs.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		errs.FailedValidationResponse(w, r, v.Errors)
		return
	}

	identity, err := app.models.Identities.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			errs.InvalidCredentialsResponse(w, r)
		default:
			errs.ServerErrorResponse(w, r, err)
		}
		return
	}

	if !identity.Activated {
		errs.InactiveAccountResponse(w, r)
		return
	}

	match, err := identity.Password.Matches(input.Password)
	if err != nil {
		errs.ServerErrorResponse(w, r, err)
		return
	}

	if !match {
		errs.InvalidCredentialsResponse(w, r)
		return
	}

	// fetch jwt secret from config
	jwtSecret := []byte(app.config.jwt.secret)

	var claims jwt.Claims
	claims.Subject = identity.UUID
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(time.Now().Add(24 * time.Hour))
	claims.Issuer = "github.com/windevkay/flho/identity-service"
	claims.Audiences = []string{"github.com/windevkay/flho"}

	// sign the token using HMAC algorithm and jwt secret
	jwtBytes, err := claims.HMACSign(jwt.HS256, jwtSecret)
	if err != nil {
		errs.ServerErrorResponse(w, r, err)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"authentication_token": string(jwtBytes)}, nil)
}
