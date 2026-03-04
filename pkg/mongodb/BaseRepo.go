package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type PaginationQuery struct {
	Page  int64 // Current page number (1-indexed)
	Limit int64 // Number of items per page
}

type PaginatedResult[T any] struct {
	Data       []T   // The paginated data
	Total      int64 // Total number of documents matching the filter
	Page       int64 // Current page number
	Limit      int64 // Items per page
	TotalPages int64 // Total number of pages
}

type filterDefine struct {
	column string
	val    any
}

// filter
var (
	In = func(col string, val any) filterDefine {
		return filterDefine{
			column: col,
			val:    bson.M{"$in": val},
		}
	}

	Eq = func(col string, val any) filterDefine {
		return filterDefine{column: col, val: val}
	}

	Query = func(val ...filterDefine) map[string]any {
		var result = make(map[string]any)
		for _, v := range val {
			result[v.column] = v.val
		}

		return result
	}
)

type orderDefine struct {
	column string
	val    int
}

var (
	ASC = func(col string) orderDefine {
		return orderDefine{column: col, val: 1}
	}

	DESC = func(col string) orderDefine {
		return orderDefine{column: col, val: -1}
	}

	Order = func(order ...orderDefine) *options.FindOptionsBuilder {
		d := make([]bson.E, len(order))
		for i, v := range order {
			d[i] = bson.E{Key: v.column, Value: v.val}
		}
		return options.Find().
			SetSort(d)
	}
)

type BaseRepo[T BaseEntity] struct {
	collection *mongo.Collection
	entity     T
}

func NewBaseRepo[T BaseEntity](conn *Conn, entity T) *BaseRepo[T] {
	collection := conn.GetClient().Database(conn.GetDatabaseName()).Collection(entity.GetCollectionName())
	return &BaseRepo[T]{
		entity:     entity,
		collection: collection,
	}
}

func (r *BaseRepo[T]) Store(ctx context.Context, entity BaseEntity) (*mongo.InsertOneResult, error) {
	result, err := r.collection.InsertOne(ctx, entity)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *BaseRepo[T]) BulkUpdate(ctx context.Context, e []T) error {
	if len(e) == 0 {
		return nil
	}

	models := make([]mongo.WriteModel, 0, len(e))
	for _, entity := range e {
		model := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": entity.GetID()}).
			SetUpdate(bson.M{"$set": entity})
		models = append(models, model)
	}

	_, err := r.collection.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
	if err != nil {
		return err
	}

	return nil
}

func (r *BaseRepo[T]) Update(ctx context.Context, e BaseEntity) (*mongo.UpdateResult, error) {
	result, err := r.collection.UpdateByID(ctx, e.GetID(), e)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *BaseRepo[T]) UpdateByCondition(
	ctx context.Context,
	condition map[string]any,
	e BaseEntity,
) (*mongo.UpdateResult, error) {
	result, err := r.collection.UpdateOne(ctx, condition, e)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *BaseRepo[T]) Delete(ctx context.Context, key string, val string) (*mongo.DeleteResult, error) {
	filter := bson.D{{key, val}}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *BaseRepo[T]) DeleteByID(ctx context.Context, id any) (*mongo.DeleteResult, error) {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *BaseRepo[T]) DeleteByFilter(ctx context.Context, filter map[string]any) (*mongo.DeleteResult, error) {
	bsonFilter := bson.M(filter)
	result, err := r.collection.DeleteOne(ctx, bsonFilter)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *BaseRepo[T]) DeleteMany(ctx context.Context, filter map[string]any) (*mongo.DeleteResult, error) {
	bsonFilter := bson.M(filter)
	result, err := r.collection.DeleteMany(ctx, bsonFilter)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *BaseRepo[T]) BulkInsert(ctx context.Context, data []T) (*mongo.InsertManyResult, error) {
	many, err := r.collection.InsertMany(ctx, data)
	if err != nil {
		return nil, err
	}

	return many, nil
}

func (r BaseRepo[T]) Count(ctx context.Context, key string, val string) (int64, error) {
	filter := bson.D{{key, val}}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *BaseRepo[T]) CountByFilter(ctx context.Context, filter map[string]any) (int64, error) {
	bsonFilter := bson.M(filter)
	count, err := r.collection.CountDocuments(ctx, bsonFilter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *BaseRepo[T]) Exists(ctx context.Context, key string, val string) (bool, error) {
	count, err := r.Count(ctx, key, val)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *BaseRepo[T]) ExistsContains(ctx context.Context, key string, val string) (bool, error) {
	filter := bson.D{{key, bson.D{{"$regex", val}}}}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *BaseRepo[T]) ExistsByFilter(ctx context.Context, filter map[string]any) (bool, error) {
	count, err := r.CountByFilter(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *BaseRepo[T]) FindOne(ctx context.Context, key string, val string) (T, error) {
	var result T
	filter := bson.D{{key, val}}
	err := r.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (r *BaseRepo[T]) FindOneByFilter(ctx context.Context, filter map[string]any) (T, error) {
	var result T
	err := r.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (r *BaseRepo[T]) FindAll(ctx context.Context, key string, val string) ([]T, error) {
	var cachedResult = make([]T, 0)
	filter := bson.D{{key, val}}
	return r.findAll(ctx, filter, cachedResult)
}

func (r *BaseRepo[T]) FindAllByFilter(
	ctx context.Context,
	filter map[string]any,
	ops ...options.Lister[options.FindOptions],
) ([]T, error) {
	var cachedResult = make([]T, 0)
	bsonFilter := bson.M(filter)
	return r.findAll(ctx, bsonFilter, cachedResult, ops...)
}

func (r *BaseRepo[T]) FindAllByFilterPaged(ctx context.Context, filter map[string]any, pagination PaginationQuery) (*PaginatedResult[T], error) {
	bsonFilter := bson.M(filter)

	// Get total count
	total, err := r.collection.CountDocuments(ctx, bsonFilter)
	if err != nil {
		return nil, err
	}

	// Calculate skip value (0-indexed)
	skip := (pagination.Page - 1) * pagination.Limit

	// Set default limit if not provided
	if pagination.Limit <= 0 {
		pagination.Limit = 10 // Default page size
	}

	// Set default page if not provided
	if pagination.Page <= 0 {
		pagination.Page = 1
	}

	// Calculate total pages
	totalPages := (total + pagination.Limit - 1) / pagination.Limit

	// Create find options with skip and limit
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(pagination.Limit)

	// Execute query
	cursor, err := r.collection.Find(ctx, bsonFilter, findOptions)
	if err != nil {
		return nil, err
	}

	var results = make([]T, 0)
	err = cursor.All(ctx, &results)
	if err != nil {
		return nil, err
	}

	return &PaginatedResult[T]{
		Data:       results,
		Total:      total,
		Page:       pagination.Page,
		Limit:      pagination.Limit,
		TotalPages: totalPages,
	}, nil
}

func (r *BaseRepo[T]) FindAllContain(ctx context.Context, key string, val string) ([]T, error) {
	var cachedResult = make([]T, 0)
	filter := bson.M{
		key: bson.M{
			"$regex": val,
		},
	}
	return r.findAll(ctx, filter, cachedResult)
}

func (r *BaseRepo[T]) findAll(
	ctx context.Context,
	filter interface{},
	cachedResult []T,
	ops ...options.Lister[options.FindOptions],
) ([]T, error) {
	cursor, err := r.collection.Find(ctx, filter, ops...)
	if err != nil {
		return nil, err
	}

	err = cursor.All(ctx, &cachedResult)
	if err != nil {
		return nil, err
	}

	return cachedResult, nil
}
