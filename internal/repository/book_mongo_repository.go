package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/adf-code/beta-book-api/internal/delivery/request"
	"github.com/adf-code/beta-book-api/internal/entity"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type BookMongoRepository interface {
	FetchWithQueryParams(ctx context.Context, params request.BookListQueryParams) ([]entity.Book, error)
	FetchByID(ctx context.Context, id uuid.UUID) (*entity.Book, error)
	Store(ctx context.Context, book *entity.Book) error
	Remove(ctx context.Context, id uuid.UUID) error
}

type bookMongoRepo struct {
	collection *mongo.Collection
}

func NewBookMongoRepo(db *mongo.Database) BookMongoRepository {
	return &bookMongoRepo{
		collection: db.Collection("books"),
	}
}

func (r *bookMongoRepo) FetchWithQueryParams(ctx context.Context, params request.BookListQueryParams) ([]entity.Book, error) {
	filter := bson.M{}

	// Search
	if params.SearchField != "" && params.SearchValue != "" {
		filter[params.SearchField] = bson.M{
			"$regex":   params.SearchValue,
			"$options": "i",
		}
	}

	// Filters
	for _, f := range params.Filter {
		if len(f.Value) > 0 {
			filter[f.Field] = bson.M{"$in": f.Value}
		}
	}

	// Range
	for _, rng := range params.Range {
		rangeFilter := bson.M{}
		if rng.From != nil {
			rangeFilter["$gte"] = *rng.From
		}
		if rng.To != nil {
			rangeFilter["$lte"] = *rng.To
		}
		if len(rangeFilter) > 0 {
			filter[rng.Field] = rangeFilter
		}
	}

	opts := options.Find()

	// Sort
	if params.SortField != "" && (params.SortDir == "ASC" || params.SortDir == "DESC") {
		direction := 1
		if params.SortDir == "DESC" {
			direction = -1
		}
		opts.SetSort(bson.D{{Key: params.SortField, Value: direction}})
	}

	// Pagination
	if params.Page > 0 && params.PerPage > 0 {
		offset := int64((params.Page - 1) * params.PerPage)
		limit := int64(params.PerPage)
		opts.SetSkip(offset)
		opts.SetLimit(limit)
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch books: %w", err)
	}
	defer cursor.Close(ctx)

	var books []entity.Book
	for cursor.Next(ctx) {
		var doc bookDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode book: %w", err)
		}
		books = append(books, doc.toEntity())
	}

	return books, nil
}

func (r *bookMongoRepo) FetchByID(ctx context.Context, id uuid.UUID) (*entity.Book, error) {
	var doc bookDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	book := doc.toEntity()
	return &book, nil
}

func (r *bookMongoRepo) Store(ctx context.Context, book *entity.Book) error {
	now := time.Now()
	book.ID = uuid.New()
	book.CreatedAt = &now
	book.UpdatedAt = &now

	doc := bookDocument{
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

func (r *bookMongoRepo) Remove(ctx context.Context, id uuid.UUID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.String()})
	return err
}

// bookDocument is the MongoDB document representation
type bookDocument struct {
	ID        string     `bson:"_id"`
	Title     string     `bson:"title"`
	Author    string     `bson:"author"`
	Year      int        `bson:"year"`
	CreatedAt *time.Time `bson:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at"`
}

func (d bookDocument) toEntity() entity.Book {
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
