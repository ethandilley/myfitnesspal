package service

import (
	"context"
	"log"
	"strconv"

	foodv1 "github.com/ethandilley/myfitnesspal/gen/proto/food/v1"
	"github.com/ethandilley/myfitnesspal/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

type FoodService struct {
	foodv1.UnimplementedFoodServiceServer

	q *db.Queries
}

func NewFoodService(q *db.Queries) *FoodService {
	log.Printf("Creating New Food Service")
	return &FoodService{q: q}
}

func (s *FoodService) CreateFood(ctx context.Context, req *foodv1.CreateFoodRequest) (*foodv1.CreateFoodResponse, error) {
	log.Printf("Creating new food %v with %v calories", req.Name, req.Calories)
	row, err := s.q.CreateFood(ctx, db.CreateFoodParams{
		Name:     req.Name,
		Calories: floatToNumeric(req.Calories),
		ProteinG: floatToNumeric(req.ProteinG),
		CarbsG:   floatToNumeric(req.CarbsG),
		FatG:     floatToNumeric(req.FatG),
	})
	if err != nil {
		return nil, err
	}
	return &foodv1.CreateFoodResponse{Food: toProtoFood(row)}, nil
}

func (s *FoodService) ListFoods(ctx context.Context, req *foodv1.ListFoodsRequest) (*foodv1.ListFoodsResponse, error) {
	log.Printf("Listing all foods")
	rows, err := s.q.ListFoods(ctx)
	if err != nil {
		return nil, err
	}
	foods := make([]*foodv1.Food, len(rows))
	for i, row := range rows {
		foods[i] = toProtoFood(row)
	}
	return &foodv1.ListFoodsResponse{Foods: foods}, nil
}

func (s *FoodService) GetFood(ctx context.Context, req *foodv1.GetFoodRequest) (*foodv1.GetFoodResponse, error) {
	log.Printf("Getting food with id %v", req.Id)
	row, err := s.q.GetFood(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &foodv1.GetFoodResponse{Food: toProtoFood(row)}, nil
}

func (s *FoodService) DeleteFood(ctx context.Context, req *foodv1.DeleteFoodRequest) (*foodv1.DeleteFoodResponse, error) {
	log.Printf("Deleting food with id %v", req.Id)
	err := s.q.DeleteFood(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &foodv1.DeleteFoodResponse{}, nil
}

func toProtoFood(row db.Food) *foodv1.Food {
	return &foodv1.Food{
		Id:       row.ID,
		Name:     row.Name,
		Calories: numericToFloat(row.Calories),
		ProteinG: numericToFloat(row.ProteinG),
		CarbsG:   numericToFloat(row.CarbsG),
		FatG:     numericToFloat(row.FatG),
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
