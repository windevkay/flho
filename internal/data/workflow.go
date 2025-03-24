package data

import (
	"context"
	"time"

	"github.com/windevkay/flhoutils/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type State struct {
	ID         primitive.ObjectID `bson:"_id" json:"-"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
	Name       string             `bson:"name" json:"name"`
	Step       int                `bson:"step" json:"step"`
	Retry      bool               `bson:"retry" json:"retry"`
	RetryUrl   string             `bson:"retryUrl" json:"retryUrl"`
	RetryAfter Timeout            `bson:"retryAfter" json:"retryAfter"`
}

type Workflow struct {
	ID         primitive.ObjectID `bson:"_id" json:"id"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
	IdentityId primitive.ObjectID `bson:"identity_id" json:"identity_id"`
	UniqueID   string             `bson:"uniqueId" json:"uniqueId"`
	Name       string             `bson:"name" json:"name"`
	States     []State            `bson:"states" json:"states"`
	Active     bool               `bson:"active" json:"active"`
	Version    int32              `bson:"version" json:"version"`
}

func (w *Workflow) GetStateByStep(step int) (State, bool) {
	for _, state := range w.States {
		if state.Step == step {
			return state, true
		}
	}
	return State{}, false
}

func (w *Workflow) HasValidRetries() bool {
	for _, s := range w.States {
		if s.Retry && (s.RetryUrl == "" || s.RetryAfter == 0) {
			return false
		}
	}
	return true
}

func ValidateWorkflow(v *validator.Validator, w *Workflow) {
	v.Check(w.Name != "", "name", "must be provided")
	v.Check(len(w.States) >= 2, "states", "must have at least 2 values")
	v.Check(w.HasValidRetries(), "retry", "when true must have a valid RetryUrl and RetryAfter value")
}

type WorkflowModelInterface interface {
	Insert(workflow *Workflow) error
	Get(uniqueId string) (*Workflow, error)
	GetAll(identityId primitive.ObjectID, filters Filters) ([]*Workflow, Metadata, error)
	Update(workflow *Workflow) error
	Delete(uniqueId string) error
}

type WorkflowModel struct {
	Collection *mongo.Collection
}

func NewWorkflowModel(client *mongo.Client, dbName string) WorkflowModel {
	collection := client.Database(dbName).Collection("workflows")
	return WorkflowModel{
		Collection: collection,
	}
}

func (w WorkflowModel) Insert(workflow *Workflow) error {
	workflow.ID = primitive.NewObjectID()
	workflow.CreatedAt = time.Now()
	workflow.UpdatedAt = time.Now()

	for i := range workflow.States {
		workflow.States[i].ID = primitive.NewObjectID()
		workflow.States[i].CreatedAt = time.Now()
		workflow.States[i].UpdatedAt = time.Now()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := w.Collection.InsertOne(ctx, workflow)

	return err
}

func (w WorkflowModel) Get(uniqueId string) (*Workflow, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var workflow Workflow

	err := w.Collection.FindOne(ctx, bson.M{"uniqueId": uniqueId}).Decode(&workflow)
	if err == mongo.ErrNoDocuments {
		return nil, ErrRecordNotFound
	}

	return &workflow, err
}

func (w WorkflowModel) GetAll(identityId primitive.ObjectID, filters Filters) ([]*Workflow, Metadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	filter := bson.M{"identity_id": identityId}
	opts := options.Find().
		SetSort(bson.D{{Key: filters.sortField(), Value: filters.sortDirection()}}).
		SetLimit(int64(filters.limit())).
		SetSkip(int64(filters.offset()))

	cursor, err := w.Collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer cursor.Close(ctx)

	var workflows []*Workflow

	for cursor.Next(ctx) {
		var workflow Workflow

		if err := cursor.Decode(&workflow); err != nil {
			return nil, Metadata{}, err
		}

		workflows = append(workflows, &workflow)
	}

	if err := cursor.Err(); err != nil {
		return nil, Metadata{}, err
	}

	totalRecords, err := w.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(int(totalRecords), filters.Page, filters.PageSize)

	return workflows, metadata, nil
}

func (w WorkflowModel) Update(workflow *Workflow) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	workflow.UpdatedAt = time.Now()
	filter := bson.M{"_id": workflow.ID, "version": workflow.Version}
	update := bson.M{
		"$set": workflow,
	}

	result, err := w.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrEditConflict
	}
	workflow.Version++

	return nil
}

func (w WorkflowModel) Delete(uniqueId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := w.Collection.DeleteOne(ctx, bson.M{"uniqueId": uniqueId})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrRecordNotFound
	}

	return nil
}
