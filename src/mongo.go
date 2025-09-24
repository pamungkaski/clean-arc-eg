package src

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Mongo DAO shape (matches collection)
type mongoBudget struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Amount      float64            `bson:"amount"`
	Currency    string             `bson:"currency"`
	LastUpdated time.Time          `bson:"last_updated"`
}

func (m mongoBudget) toUsecaseBudget() Budget {
	return Budget{
		ID:          m.ID.Hex(), // if usecase.Budget expects string IDs
		Name:        m.Name,
		Amount:      m.Amount,
		Currency:    m.Currency,
		LastUpdated: m.LastUpdated,
	}
}

type BudgetMongo struct {
	coll *mongo.Collection
}

func NewBudgetMongo(db *mongo.Database, collectionName string) *BudgetMongo {
	return &BudgetMongo{coll: db.Collection(collectionName)}
}

func (b *BudgetMongo) GetAllBudget() ([]Budget, error) {
	if b.coll == nil {
		return nil, mongo.ErrClientDisconnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cur, err := b.coll.Find(ctx, bson.D{}) // empty filter = all docs
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var raws []mongoBudget
	if err := cur.All(ctx, &raws); err != nil {
		return nil, err
	}

	out := make([]Budget, 0, len(raws))
	for _, r := range raws {
		out = append(out, r.toUsecaseBudget())
	}
	return out, nil
}
