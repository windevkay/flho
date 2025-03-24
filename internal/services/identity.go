package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/windevkay/flho/internal/data"
	"github.com/windevkay/flhoutils/validator"
)

type IdentityService struct {
	*ServiceConfig
}

type RegisterIdentityInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ActivateIdentityInput struct {
	TokenPlaintext string `json:"token"`
}

func (i *IdentityService) RegisterIdentity(input RegisterIdentityInput) (*data.Identity, error) {
	identity := &data.Identity{
		UUID:      uuid.NewString(),
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err := identity.Password.Set(input.Password)
	if err != nil {
		i.Logger.Error(err.Error())
		return nil, err
	}

	v := validator.New()

	if data.ValidateIdentity(v, identity); !v.Valid() {
		i.Logger.Error(fmt.Sprintf("validation failed: register identity - %v", v.Errors))
		return nil, &ValidationErr{Err: data.ErrValidationFailed, Fields: v.Errors}
	}

	err = i.Models.Identities.Insert(identity)
	if err != nil {
		i.Logger.Error(err.Error())
		if errors.Is(err, data.ErrDuplicateEmail) {
			v.AddError("email", "address already in use")
			return nil, &ValidationErr{Err: data.ErrDuplicateEmail, Fields: v.Errors}
		}
		return nil, err
	}

	// generate user activation token
	token, err := i.Models.Tokens.New(identity.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		i.Logger.Error(err.Error())
		return nil, err
	}

	// send welcome email
	emailData := struct {
		name            string
		activationToken string
	}{
		name:            identity.Name,
		activationToken: token.Plaintext,
	}
	// TODO: Convert this to async
	err = i.Mailer.Send(identity.Email, "user_welcome.tmpl", emailData)

	return identity, err
}

func (i *IdentityService) ActivateIdentity(input ActivateIdentityInput) (*data.Identity, error) {
	v := validator.New()

	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		i.Logger.Error(fmt.Sprintf("validation failed: activate identity - %v", v.Errors))
		return nil, &ValidationErr{Err: data.ErrValidationFailed, Fields: v.Errors}
	}

	identity, err := i.Models.Identities.GetIdentityForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		i.Logger.Error(err.Error())
		if errors.Is(err, data.ErrRecordNotFound) {
			v.AddError("token", "invalid or expired activation token")
			return nil, &ValidationErr{Err: data.ErrRecordNotFound, Fields: v.Errors}
		}
		return nil, err
	}

	identity.Activated = true

	err = i.Models.Identities.Update(identity)
	if err != nil {
		return nil, err
	}

	err = i.Models.Tokens.DeleteScopeTokensForIdentity(data.ScopeActivation, identity.ID)
	if err != nil {
		i.Logger.Error(err.Error())
		return nil, err
	}

	return identity, nil
}
