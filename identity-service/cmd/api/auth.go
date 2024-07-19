package main

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			errs.InvalidCredentialsResponse(w, r)
		default:
			errs.ServerErrorResponse(w, r, err)
		}
		return
	}

	if !user.Activated {
		errs.InactiveAccountResponse(w, r)
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		errs.ServerErrorResponse(w, r, err)
		return
	}

	if !match {
		errs.InvalidCredentialsResponse(w, r)
		return
	}

	// fetch private key needed for token creation
	privKeyPath := filepath.Join(".", "keys", "ec_private_key.pem")
	privPem, err := os.ReadFile(privKeyPath)
	if err != nil {
		errs.ServerErrorResponse(w, r, err)
		return
	}
	block, _ := pem.Decode(privPem)
	priv, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		errs.ServerErrorResponse(w, r, err)
		return
	}

	var claims jwt.Claims
	claims.Subject = strconv.FormatInt(user.ID, 10)
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(time.Now().Add(24 * time.Hour))
	// swap to env variables
	claims.Issuer = "github.com/windevkay/flho/identity-service"
	claims.Audiences = []string{"github.com/windevkay/flho"}

	// swap out claims function for one that accepts assymetric algo - private key
	//jwtBytes, err := claims.HMACSign(jwt.HS256, []byte(app.config.jwt.secret))
	jwtBytes, err := claims.ECDSASign(jwt.ES256, priv)
	if err != nil {
		errs.ServerErrorResponse(w, r, err)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"authentication_token": string(jwtBytes)}, nil)
}
