package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/windevkay/flhoutils/validator"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

type Identity struct {
	ID        int64      `json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	UUID      string     `json:"-"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Password  password   `json:"-"`
	Activated bool       `json:"activated"`
	Version   int        `json:"-"`
}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateIdentity(v *validator.Validator, identity *Identity) {
	v.Check(identity.Name != "", "name", "must be provided")
	v.Check(len(identity.Name) <= 500, "name", "must not be more than 500 bytes long")

	ValidateEmail(v, identity.Email)

	if identity.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *identity.Password.plaintext)
	}

	if identity.Password.hash == nil {
		panic("missing password hash for user")
	}
}

type IdentityModelInterface interface {
	Insert(identity *Identity) error
	GetByEmail(email string) (*Identity, error)
	Update(user *Identity) error
	GetIdentityForToken(tokenScope, tokenPlaintext string) (*Identity, error)
	Get(id int64) (*Identity, error)
}

type IdentityModel struct {
	DB *sql.DB
}

func (i IdentityModel) Insert(identity *Identity) error {
	query := `INSERT INTO identities (uuid, name, email, password_hash, activated)
				VALUES ($1, $2, $3, $4, $5)
				RETURNING id, created_at, updated_at, version`

	args := []any{identity.UUID, identity.Name, identity.Email, identity.Password.hash, identity.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := i.DB.QueryRowContext(ctx, query, args...).Scan(&identity.ID, &identity.CreatedAt, &identity.UpdatedAt, &identity.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "identities_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (i IdentityModel) GetByEmail(email string) (*Identity, error) {
	query := `SELECT id, created_at, updated_at, name, email, password_hash, activated, version
				FROM identities
				WHERE email = $1`

	var identity Identity

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := i.DB.QueryRowContext(ctx, query, email).Scan(
		&identity.ID,
		&identity.CreatedAt,
		&identity.UpdatedAt,
		&identity.Name,
		&identity.Email,
		&identity.Password.hash,
		&identity.Activated,
		&identity.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &identity, nil
}

func (i IdentityModel) Update(identity *Identity) error {
	query := `UPDATE identities
				SET updated_at = NOW(), name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
				WHERE id = $5 AND version = $6
				RETURNING version`

	args := []any{
		identity.Name,
		identity.Email,
		identity.Password.hash,
		identity.Activated,
		identity.ID,
		identity.Version,
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()

	err := i.DB.QueryRowContext(ctx, query, args...).Scan(&identity.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "identities_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (i IdentityModel) GetIdentityForToken(tokenScope, tokenPlaintext string) (*Identity, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := `SELECT i.id, i.created_at, i.updated_at, i.name, i.email, i.password_hash, i.activated, i.version
				FROM identities i
				INNER JOIN tokens t
				ON i.id = t.user_id
				WHERE t.hash = $1
				AND t.scope = $2
				AND t.expiry > $3`

	args := []any{tokenHash[:], tokenScope, time.Now()}

	var identity Identity

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := i.DB.QueryRowContext(ctx, query, args...).Scan(
		&identity.ID, &identity.CreatedAt, &identity.UpdatedAt, &identity.Name, &identity.Email, &identity.Password.hash, &identity.Activated, &identity.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &identity, nil
}

func (i IdentityModel) Get(id int64) (*Identity, error) {
	query := `SELECT id, created_at, name, email, password_hash, activated, version
				FROM identities
				WHERE id = $1`

	var identity Identity

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := i.DB.QueryRowContext(ctx, query, id).Scan(
		&identity.ID,
		&identity.CreatedAt,
		&identity.Name,
		&identity.Email,
		&identity.Password.hash,
		&identity.Activated,
		&identity.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &identity, nil
}
