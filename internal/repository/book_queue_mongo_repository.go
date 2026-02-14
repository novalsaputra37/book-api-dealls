package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/adf-code/beta-book-api/internal/entity"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// BookQueueMongoRepository manages the book_queue collection in MongoDB
// for storing non-fibonacci books from the Kafka consumer.
type BookQueueMongoRepository interface {
	Store(ctx context.Context, book *entity.Book) error
	FetchAll(ctx context.Context) ([]entity.Book, error)
	RemoveAll(ctx context.Context) error
}

type bookQueueMongoRepo struct {
	collection *mongo.Collection
}

func NewBookQueueMongoRepo(db *mongo.Database) BookQueueMongoRepository {
	return &bookQueueMongoRepo{
		collection: db.Collection("book_queue"),
	}
}

func (r *bookQueueMongoRepo) Store(ctx context.Context, book *entity.Book) error {
	now := time.Now()
	if book.ID == uuid.Nil {
		book.ID = uuid.New()
	}
	if book.CreatedAt == nil {
		book.CreatedAt = &now
	}
	if book.UpdatedAt == nil {
		book.UpdatedAt = &now
	}

	doc := bookQueueDocument{
		ID:        book.ID.String(),
		Title:     book.Title,
		Author:    book.Author,
		Year:      book.Year,
		CreatedAt: book.CreatedAt,
		UpdatedAt: book.UpdatedAt,
	}

	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

func (r *bookQueueMongoRepo) FetchAll(ctx context.Context) ([]entity.Book, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch book queue: %w", err)
	}
	defer cursor.Close(ctx)

	var books []entity.Book
	for cursor.Next(ctx) {
		var doc bookQueueDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode book queue item: %w", err)
		}
		books = append(books, doc.toEntity())
	}

	return books, nil
}

func (r *bookQueueMongoRepo) RemoveAll(ctx context.Context) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{})
	return err
}

// bookQueueDocument is the MongoDB document representation for queued books.
type bookQueueDocument struct {
	ID        string     `bson:"_id"`
	Title     string     `bson:"title"`
	Author    string     `bson:"author"`
	Year      int        `bson:"year"`
	CreatedAt *time.Time `bson:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at"`
}

func (d bookQueueDocument) toEntity() entity.Book {
	id, _ := uuid.Parse(d.ID)
	return entity.Book{
		ID:        id,
		Title:     d.Title,
		Author:    d.Author,
		Year:      d.Year,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}
