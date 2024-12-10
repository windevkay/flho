package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"github.com/windevkay/flho/identity_service/internal/data"
	"github.com/windevkay/flho/identity_service/internal/queue"
	"github.com/windevkay/flho/identity_service/internal/rpc"
	pb "github.com/windevkay/flho/mailer_service/proto"
	"github.com/windevkay/flhoutils/helpers"
	"github.com/windevkay/flhoutils/validator"
)

type IdentityService struct {
	config *IdentityServiceConfig
}

type IdentityServiceConfig struct {
	Models    data.Models
	Rpclients rpc.Clients
	Channel   *amqp091.Channel
	Wg        *sync.WaitGroup
	Logger    *slog.Logger
}

type RegisterIdentityInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ActivateIdentityInput struct {
	TokenPlaintext string `json:"token"`
}

type ValidationErr struct {
	Err    error
	Fields map[string]string
}

func (c *ValidationErr) Error() string { return "validation error" }

func NewIdentityService(config *IdentityServiceConfig) *IdentityService {
	return &IdentityService{
		config: config,
	}
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
		i.config.Logger.Error(err.Error())
		return nil, err
	}

	v := validator.New()

	if data.ValidateIdentity(v, identity); !v.Valid() {
		i.config.Logger.Error(fmt.Sprintf("validation failed: register identity - %v", v.Errors))
		return nil, &ValidationErr{Err: data.ErrValidationFailed, Fields: v.Errors}
	}

	err = i.config.Models.Identities.Insert(identity)
	if err != nil {
		i.config.Logger.Error(err.Error())
		if errors.Is(err, data.ErrDuplicateEmail) {
			v.AddError("email", "address already in use")
			return nil, &ValidationErr{Err: data.ErrDuplicateEmail, Fields: v.Errors}
		}
		return nil, err
	}

	// publish event
	err = queue.SendMessage(i.config.Channel, identity, "identity", "create")
	if err != nil {
		i.config.Logger.Error(err.Error())
	}

	// generate user activation token
	token, err := i.config.Models.Tokens.New(identity.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		i.config.Logger.Error(err.Error())
		return nil, err
	}

	// send welcome email
	helpers.RunInBackground(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		_, err = i.config.Rpclients.MailerClient.SendWelcomeEmail(ctx, &pb.WelcomeEmailRequest{
			Recipient: identity.Email,
			File:      "user_welcome.tmpl",
			Data:      &pb.Data{Name: identity.Name, ActivationToken: token.Plaintext}})

		if err != nil {
			i.config.Logger.Error(err.Error())
		}
	}, i.config.Wg)

	return identity, err
}

func (i *IdentityService) ActivateIdentity(input ActivateIdentityInput) (*data.Identity, error) {
	v := validator.New()

	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		i.config.Logger.Error(fmt.Sprintf("validation failed: activate identity - %v", v.Errors))
		return nil, &ValidationErr{Err: data.ErrValidationFailed, Fields: v.Errors}
	}

	identity, err := i.config.Models.Identities.GetIdentityForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		i.config.Logger.Error(err.Error())
		if errors.Is(err, data.ErrRecordNotFound) {
			v.AddError("token", "invalid or expired activation token")
			return nil, &ValidationErr{Err: data.ErrRecordNotFound, Fields: v.Errors}
		}
		return nil, err
	}

	identity.Activated = true

	err = i.config.Models.Identities.Update(identity)
	if err != nil {
		return nil, err
	}

	err = i.config.Models.Tokens.DeleteScopeTokensForIdentity(data.ScopeActivation, identity.ID)
	if err != nil {
		i.config.Logger.Error(err.Error())
		return nil, err
	}

	return identity, nil
}
