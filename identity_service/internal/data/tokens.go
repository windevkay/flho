package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"

	"github.com/windevkay/flhoutils/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	ScopeActivation = "activation"
)

type Token struct {
	ID         primitive.ObjectID `bson:"_id"`
	Plaintext  string             `bson:"plaintext"`
	Hash       []byte             `bson:"hash"`
	IdentityID primitive.ObjectID `bson:"identity_id"`
	Expiry     time.Time          `bson:"expiry"`
	Scope      string             `bson:"scope"`
}

func generateToken(identityID primitive.ObjectID, ttl time.Duration, scope string) (*Token, error) {
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
	New(identityID primitive.ObjectID, ttl time.Duration, scope string) (*Token, error)
	DeleteScopeTokensForIdentity(scope string, identityID primitive.ObjectID) error
}

type TokenModel struct {
	Collection *mongo.Collection
}

func NewTokenModel(client *mongo.Client, dbName string) TokenModel {
	collection := client.Database(dbName).Collection("tokens")
	return TokenModel{
		Collection: collection,
	}
}

func (t TokenModel) New(identityID primitive.ObjectID, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(identityID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = t.Insert(token)
	return token, err
}

func (t TokenModel) Insert(token *Token) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := t.Collection.InsertOne(ctx, token)
	if err != nil {
		return err
	}

	return nil
}

func (t TokenModel) DeleteScopeTokensForIdentity(scope string, identityID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	filter := bson.M{"scope": scope, "identity_id": identityID}
	_, err := t.Collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	return nil
}
