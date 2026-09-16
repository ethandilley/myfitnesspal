package service

import (
	"context"
	"log"
	"strconv"

	foodv1 "github.com/ethandilley/myfitnesspal/gen/proto/food/v1"
	"github.com/ethandilley/myfitnesspal/internal/db"
	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgtype"
)

type FoodService struct {
	q *db.Queries
}

func NewFoodService(q *db.Queries) *FoodService {
	log.Printf("Creating New Food Service")
	return &FoodService{q: q}
}

func (s *FoodService) CreateFood(ctx context.Context, req *connect.Request[foodv1.CreateFoodRequest]) (*connect.Response[foodv1.CreateFoodResponse], error) {
	msg := req.Msg
	log.Printf("Creating new food %v with %v calories", msg.Name, msg.Calories)
	row, err := s.q.CreateFood(ctx, db.CreateFoodParams{
		Name:       msg.Name,
		Calories:   floatToNumeric(msg.Calories),
		ProteinG:   floatToNumeric(msg.ProteinG),
		CarbsG:     floatToNumeric(msg.CarbsG),
		FatG:       floatToNumeric(msg.FatG),
		IsFrequent: msg.IsFrequent,
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&foodv1.CreateFoodResponse{Food: toProtoFood(row)}), nil
}

func (s *FoodService) ListFoods(ctx context.Context, req *connect.Request[foodv1.ListFoodsRequest]) (*connect.Response[foodv1.ListFoodsResponse], error) {
	// pgtype.Bool zero value (Valid: false) means "no filter" in the generated query.
	var filter pgtype.Bool
	if req.Msg.IsFrequent != nil {
		filter = pgtype.Bool{Bool: *req.Msg.IsFrequent, Valid: true}
		log.Printf("Listing foods with is_frequent=%v", *req.Msg.IsFrequent)
	} else {
		log.Printf("Listing all foods")
	}
	rows, err := s.q.ListFoods(ctx, filter)
	if err != nil {
		return nil, err
	}
	foods := make([]*foodv1.Food, len(rows))
	for i, row := range rows {
		foods[i] = toProtoFood(row)
	}
	return connect.NewResponse(&foodv1.ListFoodsResponse{Foods: foods}), nil
}

func (s *FoodService) GetFood(ctx context.Context, req *connect.Request[foodv1.GetFoodRequest]) (*connect.Response[foodv1.GetFoodResponse], error) {
	msg := req.Msg
	log.Printf("Getting food with id %v", msg.Id)
	row, err := s.q.GetFood(ctx, msg.Id)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&foodv1.GetFoodResponse{Food: toProtoFood(row)}), nil
}

func (s *FoodService) SetFoodFrequent(ctx context.Context, req *connect.Request[foodv1.SetFoodFrequentRequest]) (*connect.Response[foodv1.SetFoodFrequentResponse], error) {
	msg := req.Msg
	log.Printf("Setting food with id %v is_frequent=%v", msg.Id, msg.IsFrequent)
	row, err := s.q.SetFoodFrequent(ctx, db.SetFoodFrequentParams{
		ID:         msg.Id,
		IsFrequent: msg.IsFrequent,
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&foodv1.SetFoodFrequentResponse{Food: toProtoFood(row)}), nil
}

func (s *FoodService) DeleteFood(ctx context.Context, req *connect.Request[foodv1.DeleteFoodRequest]) (*connect.Response[foodv1.DeleteFoodResponse], error) {
	msg := req.Msg
	log.Printf("Deleting food with id %v", msg.Id)
	err := s.q.DeleteFood(ctx, msg.Id)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&foodv1.DeleteFoodResponse{}), nil
}

func toProtoFood(row db.Food) *foodv1.Food {
	return &foodv1.Food{
		Id:         row.ID,
		Name:       row.Name,
		Calories:   numericToFloat(row.Calories),
		ProteinG:   numericToFloat(row.ProteinG),
		CarbsG:     numericToFloat(row.CarbsG),
		FatG:       numericToFloat(row.FatG),
		IsFrequent: row.IsFrequent,
	}
}

func floatToNumeric(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	if err := n.Scan(strconv.FormatFloat(f, 'f', -1, 64)); err != nil {
		log.Printf("failed to convert %v to numeric: %v", f, err)
	}
	return n
}
func numericToFloat(n pgtype.Numeric) float64 {
	f, _ := n.Float64Value()
	return f.Float64
}
