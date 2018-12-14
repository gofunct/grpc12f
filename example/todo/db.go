package todo

import (
	"context"
	"github.com/go-pg/pg"
	"github.com/gogo/protobuf/types"
	"github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Store is the service dealing with storing
// and retrieving todo items from the database.
type Store struct {
	DB *pg.DB
}

// CreateTodo creates a todo given a description
func (s Store) CreateTodo(ctx context.Context, req *CreateTodoRequest) (*CreateTodoResponse, error) {
	req.Item.Id = uuid.NewV4().String()
	err := s.DB.Insert(req.Item)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "Could not insert item into the database: %s", err)
	}
	return &CreateTodoResponse{Id: req.Item.Id}, nil
}

// CreateTodos create todo items from a list of todo descriptions
func (s Store) CreateTodos(ctx context.Context, req *CreateTodosRequest) (*CreateTodosResponse, error) {
	var ids []string
	for _, item := range req.Items {
		item.Id = uuid.NewV4().String()
		ids = append(ids, item.Id)
	}
	err := s.DB.Insert(&req.Items)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "Could not insert items into the database: %s", err)
	}
	return &CreateTodosResponse{Ids: ids}, nil
}

// GetTodo retrieves a todo item from its ID
func (s Store) GetTodo(ctx context.Context, req *GetTodoRequest) (*GetTodoResponse, error) {
	var item Todo
	err := s.DB.Model(&item).Where("id = ?", req.Id).First()
	if err != nil {
		return nil, grpc.Errorf(codes.NotFound, "Could not retrieve item from the database: %s", err)
	}
	return &GetTodoResponse{Item: &item}, nil
}

// ListTodo retrieves a todo item from its ID
func (s Store) ListTodo(ctx context.Context, req *ListTodoRequest) (*ListTodoResponse, error) {
	var items []*Todo
	query := s.DB.Model(&items).Order("created_at ASC")
	if req.Limit > 0 {
		query.Limit(int(req.Limit))
	}
	if req.NotCompleted {
		query.Where("completed = false")
	}
	err := query.Select()
	if err != nil {
		return nil, grpc.Errorf(codes.NotFound, "Could not list items from the database: %s", err)
	}
	return &ListTodoResponse{Items: items}, nil
}

// DeleteTodo deletes a todo given an ID
func (s Store) DeleteTodo(ctx context.Context, req *DeleteTodoRequest) (*DeleteTodoResponse, error) {
	err := s.DB.Delete(&Todo{Id: req.Id})
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "Could not delete item from the database: %s", err)
	}
	return &DeleteTodoResponse{}, nil
}

// UpdateTodo updates a todo item
func (s Store) UpdateTodo(ctx context.Context, req *UpdateTodoRequest) (*UpdateTodoResponse, error) {
	req.Item.UpdatedAt = types.TimestampNow()
	res, err := s.DB.Model(req.Item).Column("title", "description", "completed", "updated_at").Update()
	if res.RowsAffected() == 0 {
		return nil, grpc.Errorf(codes.NotFound, "Could not update item: not found")
	}
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "Could not update item from the database: %s", err)
	}
	return &UpdateTodoResponse{}, nil
}

// UpdateTodos updates todo items given their respective title and description.
func (s Store) UpdateTodos(ctx context.Context, req *UpdateTodosRequest) (*UpdateTodosResponse, error) {
	time := types.TimestampNow()
	for _, item := range req.Items {
		item.UpdatedAt = time
	}
	res, err := s.DB.Model(&req.Items).Column("title", "description", "completed", "updated_at").Update()
	if res.RowsAffected() == 0 {
		return nil, grpc.Errorf(codes.NotFound, "Could not update items: not found")
	}
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "Could not update items from the database: %s", err)
	}
	return &UpdateTodosResponse{}, nil
}

