package service

import (
	"context"
	"log"

	foodv1 "github.com/ethandilley/myfitnesspal/gen/proto/food/v1"
)

type FoodService struct {
	foodv1.UnimplementedFoodServiceServer

	foods  map[int32]*foodv1.Food
	nextID int32
}

func NewFoodService() *FoodService {
	log.Printf("Creating New Food Service")
	return &FoodService{
		foods:  make(map[int32]*foodv1.Food),
		nextID: 1,
	}
}

func (s *FoodService) CreateFood(ctx context.Context, req *foodv1.CreateFoodRequest) (*foodv1.CreateFoodResponse, error) {
	log.Printf("Creating new food %v with id %v and %v calories", req.Name, s.nextID, req.Calories)
	f := &foodv1.Food{
		Id:       s.nextID,
		Name:     req.Name,
		Calories: req.Calories,
		ProteinG: req.ProteinG,
		CarbsG:   req.CarbsG,
		FatG:     req.FatG,
	}
	s.foods[s.nextID] = f
	s.nextID++

	return &foodv1.CreateFoodResponse{Food: f}, nil
}

func (s *FoodService) ListFoods(ctx context.Context, req *foodv1.ListFoodsRequest) (*foodv1.ListFoodsResponse, error) {
	log.Printf("Listing all foods")
	foods := make([]*foodv1.Food, 0, len(s.foods))
	for _, f := range s.foods {
		foods = append(foods, f)
	}
	return &foodv1.ListFoodsResponse{Foods: foods}, nil
}

func (s *FoodService) GetFood(ctx context.Context, req *foodv1.GetFoodRequest) (*foodv1.GetFoodResponse, error) {
	log.Printf("Getting food with id %v", req.Id)
	f := s.foods[req.Id]
	return &foodv1.GetFoodResponse{Food: f}, nil
}

func (s *FoodService) DeleteFood(ctx context.Context, req *foodv1.DeleteFoodRequest) (*foodv1.DeleteFoodResponse, error) {
	log.Printf("Deleting food with id %v", req.Id)
	delete(s.foods, req.Id)
	return &foodv1.DeleteFoodResponse{}, nil
}
