package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/adf-code/beta-book-api/internal/entity"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type BookCoverMongoRepository interface {
	Store(ctx context.Context, cover *entity.BookCover) error
	FetchByBookID(ctx context.Context, bookID uuid.UUID) ([]entity.BookCover, error)
}

type bookCoverMongoRepo struct {
	collection *mongo.Collection
}

func NewBookCoverMongoRepo(db *mongo.Database) BookCoverMongoRepository {
	return &bookCoverMongoRepo{
		collection: db.Collection("book_covers"),
	}
}

func (r *bookCoverMongoRepo) Store(ctx context.Context, cover *entity.BookCover) error {
	now := time.Now()
	cover.ID = uuid.New()
	cover.CreatedAt = &now
	cover.UpdatedAt = &now

	doc := bookCoverDocument{
		ID:        cover.ID.String(),
		BookID:    cover.BookID.String(),
		FileName:  cover.FileName,
		FileURL:   cover.FileURL,
		CreatedAt: cover.CreatedAt,
		UpdatedAt: cover.UpdatedAt,
	}

	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

func (r *bookCoverMongoRepo) FetchByBookID(ctx context.Context, bookID uuid.UUID) ([]entity.BookCover, error) {
	opts := options.Find()
	opts.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{"book_id": bookID.String()}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch book covers: %w", err)
	}
	defer cursor.Close(ctx)

	var covers []entity.BookCover
	for cursor.Next(ctx) {
		var doc bookCoverDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode book cover: %w", err)
		}
		covers = append(covers, doc.toEntity())
	}

	return covers, nil
}

// bookCoverDocument is the MongoDB document representation
type bookCoverDocument struct {
	ID        string     `bson:"_id"`
	BookID    string     `bson:"book_id"`
	FileName  string     `bson:"file_name"`
	FileURL   string     `bson:"file_url"`
	CreatedAt *time.Time `bson:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at"`
}

func (d bookCoverDocument) toEntity() entity.BookCover {
	id, _ := uuid.Parse(d.ID)
	bookID, _ := uuid.Parse(d.BookID)
	return entity.BookCover{
		ID:        id,
		BookID:    bookID,
		FileName:  d.FileName,
		FileURL:   d.FileURL,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}
