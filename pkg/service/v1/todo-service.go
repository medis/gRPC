package v1

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	v1 "github.com/medis/grpc/pkg/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	apiVersion = "v1"
)

type toDoServiceServer struct {
	db *sql.DB
}

func NewToDoServiceServer(db *sql.DB) *toDoServiceServer {
	return &toDoServiceServer{db: db}
}

func (s *toDoServiceServer) checkAPI(api string) error {
	if len(api) > 0 {
		if apiVersion != api {
			return status.Errorf(codes.Unimplemented, "unsupported API version: service implements API version '%s', bus aksed '%s'", apiVersion, api)
		}
	}
	return nil
}

func (s *toDoServiceServer) connect(ctx context.Context) (*sql.Conn, error) {
	c, err := s.db.Conn(ctx)
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to connect to database.")
	}
	return c, nil
}

func (s *toDoServiceServer) Create(ctx context.Context, req *v1.CreateRequest) (*v1.CreateResponse, error) {
	// Check if the API version requested by client is supported by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	// Get SQL connection from pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	reminder, err := ptypes.Timestamp(req.ToDo.Reminder)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "reminder field has invalid format-> "+err.Error())
	}

	// Insert new Task.
	res, err := c.ExecContext(ctx, "INSERT INTO ToDo (Title, Description, Reminder) VALUES (?, ?, ?)", req.ToDo.Title, req.ToDo.Description, reminder)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to insert into ToDo-> "+err.Error())
	}

	// Get id of created task.
	id, err := res.LastInsertId()
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve id for created ToDo-> "+err.Error())
	}

	return &v1.CreateResponse{
		Api: apiVersion,
		Id:  id,
	}, nil
}

func (s *toDoServiceServer) Read(ctx context.Context, req *v1.ReadRequest) (*v1.ReadResponse, error) {
	// Check if the API version requested by client is supported by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	// Get SQL connection from pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	rows, err := c.QueryContext(ctx, "SELECT ID, Title, Description, Reminder FROM ToDo WHERE ID = ?", req.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to select from ToDo-> "+err.Error())
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, status.Error(codes.Unknown, "failed to retrieve data from ToDo-> "+err.Error())
		}
		return nil, status.Error(codes.NotFound, fmt.Sprintf("ToDo with ID='%d' is not found", req.Id))
	}

	// Get ToDo data.
	var td v1.ToDo
	var reminder time.Time

	if err := rows.Scan(&td.Id, &td.Title, &td.Description, &reminder); err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve field values from ToDo row-> "+err.Error())
	}

	td.Reminder, err = ptypes.TimestampProto(reminder)
	if err != nil {
		return nil, status.Error(codes.Unknown, "reminder field has invalid format-> "+err.Error())
	}

	return &v1.ReadResponse{
		Api:  apiVersion,
		ToDo: &td,
	}, nil
}

func (s *toDoServiceServer) Update(ctx context.Context, req *v1.UpdateRequest) (*v1.UpdateResponse, error) {
	// Check if the API version requested by client is supported by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	// Get SQL connection from pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	reminder, err := ptypes.Timestamp(req.ToDo.Reminder)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "reminder field has invalid format-> "+err.Error())
	}

	// Update ToDo
	res, err := c.ExecContext(ctx, "UPDATE ToDo SET Title=?, Description=?, Reminder=? WHERE ID=?", req.ToDo.Title, req.ToDo.Description, reminder, req.ToDo.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to update ToDo-> "+err.Error())
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve rows affected value-> "+err.Error())
	}

	if rows == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("ToDo with ID='%d' is not found",
			req.ToDo.Id))
	}

	return &v1.UpdateResponse{
		Api:     apiVersion,
		Updated: rows,
	}, nil
}

// Delete todo task.
func (s *toDoServiceServer) Delete(ctx context.Context, req *v1.DeleteRequest) (*v1.DeleteResponse, error) {
	// check if the API version requested by client is supported by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	// get SQL connection from pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	// delete ToDo
	res, err := c.ExecContext(ctx, "DELETE FROM ToDo WHERE ID=?", req.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to delete ToDo-> "+err.Error())
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve rows affected value-> "+err.Error())
	}

	if rows == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("ToDo with ID='%d' is not found",
			req.Id))
	}

	return &v1.DeleteResponse{
		Api:     apiVersion,
		Deleted: rows,
	}, nil
}

// Read all todo tasks.
func (s *toDoServiceServer) ReadAll(ctx context.Context, req *v1.ReadAllRequest) (*v1.ReadAllResponse, error) {
	// check if the API version requested by client is supported by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	// get SQL connection from pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	// get ToDo list
	rows, err := c.QueryContext(ctx, "SELECT ID, Title, Description, Reminder FROM ToDo")
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to select from ToDo-> "+err.Error())
	}
	defer rows.Close()

	var reminder time.Time
	list := []*v1.ToDo{}
	for rows.Next() {
		td := new(v1.ToDo)
		if err := rows.Scan(&td.Id, &td.Title, &td.Description, &reminder); err != nil {
			return nil, status.Error(codes.Unknown, "failed to retrieve field values from ToDo row-> "+err.Error())
		}
		td.Reminder, err = ptypes.TimestampProto(reminder)
		if err != nil {
			return nil, status.Error(codes.Unknown, "reminder field has invalid format-> "+err.Error())
		}
		list = append(list, td)
	}

	if err := rows.Err(); err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve data from ToDo-> "+err.Error())
	}

	return &v1.ReadAllResponse{
		Api:   apiVersion,
		ToDos: list,
	}, nil
}
