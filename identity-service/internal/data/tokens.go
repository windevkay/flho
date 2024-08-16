package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"github.com/windevkay/flhoutils/validator"
)

const (
	ScopeActivation = "activation"
)

type Token struct {
	Plaintext  string
	Hash       []byte
	IdentityID int64
	Expiry     time.Time
	Scope      string
}

func generateToken(identityID int64, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		IdentityID: identityID,
		Expiry:     time.Now().Add(ttl),
		Scope:      scope,
	}

	randomBytes := make([]byte, 16)

	// read random bytes into the byte slice (using the OS' CSPRNG)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}

func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

type TokenModelInterface interface {
	Insert(token *Token) error
	New(identityID int64, ttl time.Duration, scope string) (*Token, error)
	DeleteScopeTokensForIdentity(scope string, identityID int64) error
}

type TokenModel struct {
	DB *sql.DB
}

func (t TokenModel) New(identityID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(identityID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = t.Insert(token)
	return token, err
}

func (t TokenModel) Insert(token *Token) error {
	query := `INSERT INTO tokens (hash, identity_id, expiry, scope)
				VALUES ($1, $2, $3, $4)`

	args := []any{token.Hash, token.IdentityID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := t.DB.ExecContext(ctx, query, args...)
	return err
}

func (t TokenModel) DeleteScopeTokensForIdentity(scope string, identityID int64) error {
	query := `DELETE FROM tokens WHERE scope = $1 AND identity_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := t.DB.ExecContext(ctx, query, scope, identityID)
	return err
}
