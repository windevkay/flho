package data

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Run struct {
	ID        primitive.ObjectID `bson:"_id" json:"-"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
	UniqueID  string             `bson:"uniqueId" json:"uniqueId"`
	Step      int                `bson:"step" json:"step"`
}

type RunModelInterface interface {
	Insert(run *Run) error
	Get(uniqueId string) (*Run, error)
	Update(run *Run) error
}

type RunModel struct {
	Collection *mongo.Collection
}

func NewRunModel(client *mongo.Client, dbName string) RunModel {
	collection := client.Database(dbName).Collection("runs")
	return RunModel{
		Collection: collection,
	}
}

func (r RunModel) Insert(run *Run) error {
	run.ID = primitive.NewObjectID()
	run.CreatedAt = time.Now()
	run.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := r.Collection.InsertOne(ctx, run)

	return err
}

func (r RunModel) Get(uniqueId string) (*Run, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var run Run

	err := r.Collection.FindOne(ctx, bson.M{"uniqueId": uniqueId}).Decode(&run)
	if err == mongo.ErrNoDocuments {
		return nil, ErrRecordNotFound
	}

	return &run, nil
}

func (r RunModel) Update(run *Run) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	run.UpdatedAt = time.Now()
	filter := bson.M{"_id": run.ID}
	update := bson.M{
		"$set": run,
	}

	_, err := r.Collection.UpdateOne(ctx, filter, update)

	return err
}
