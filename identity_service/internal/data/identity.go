package data

import (
	"context"
	"crypto/sha256"
	"errors"
	"log"
	"time"

	"github.com/windevkay/flhoutils/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail   = errors.New("duplicate email")
	ErrValidationFailed = errors.New("validation failed")
)

type Identity struct {
	ID        primitive.ObjectID `bson:"_id" json:"id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
	UUID      string             `bson:"uuid" json:"-"`
	Name      string             `bson:"name" json:"name"`
	Email     string             `bson:"email" json:"email"`
	Password  password           `bson:"password" json:"-"`
	Activated bool               `bson:"activated" json:"activated"`
	Version   int                `bson:"version" json:"-"`
}

type password struct {
	plaintext *string `bson:"plaintext"`
	hash      []byte  `bson:"hash"`
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
	Get(id primitive.ObjectID) (*Identity, error)
}

type IdentityModel struct {
	Collection *mongo.Collection
}

func NewIdentityModel(client *mongo.Client, dbName string) IdentityModel {
	collection := client.Database(dbName).Collection("identities")
	model := IdentityModel{
		Collection: collection,
	}

	// Create unique email index
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	if err != nil {
		log.Fatal(err)
	}

	return model
}

func (i IdentityModel) Insert(identity *Identity) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	identity.ID = primitive.NewObjectID()
	identity.CreatedAt = time.Now()
	identity.UpdatedAt = time.Now()

	_, err := i.Collection.InsertOne(ctx, identity)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrDuplicateEmail
		}
		return err
	}

	return nil
}

func (i IdentityModel) GetByEmail(email string) (*Identity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var identity Identity
	err := i.Collection.FindOne(ctx, bson.M{"email": email}).Decode(&identity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &identity, nil
}

func (i IdentityModel) Update(identity *Identity) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	identity.UpdatedAt = time.Now()

	filter := bson.M{"_id": identity.ID, "version": identity.Version}
	update := bson.M{
		"$set": bson.M{
			"name":       identity.Name,
			"email":      identity.Email,
			"password":   identity.Password,
			"activated":  identity.Activated,
			"updated_at": identity.UpdatedAt,
			"version":    identity.Version,
		},
	}

	result, err := i.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrEditConflict
	}

	return nil
}

func (i IdentityModel) GetIdentityForToken(tokenScope, tokenPlaintext string) (*Identity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	pipeline := []bson.M{
		{
			"$lookup": bson.M{
				"from":         "tokens",
				"localField":   "_id",
				"foreignField": "identity_id",
				"as":           "tokens",
			},
		},
		{
			"$match": bson.M{
				"tokens": bson.M{
					"$elemMatch": bson.M{
						"hash":   tokenHash[:],
						"scope":  tokenScope,
						"expiry": bson.M{"$gt": time.Now()},
					},
				},
			},
		},
	}

	var identity Identity
	cursor, err := i.Collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return nil, ErrRecordNotFound
	}

	err = cursor.Decode(&identity)
	if err != nil {
		return nil, err
	}

	return &identity, nil
}

func (i IdentityModel) Get(id primitive.ObjectID) (*Identity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var identity Identity
	err := i.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&identity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &identity, nil
}
