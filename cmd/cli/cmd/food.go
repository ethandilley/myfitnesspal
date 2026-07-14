package cmd

import (
	"context"
	"fmt"

	foodv1 "github.com/ethandilley/myfitnesspal/gen/proto/food/v1"
	"github.com/ethandilley/myfitnesspal/internal/client"

	"github.com/spf13/cobra"
)

var foodCmd = &cobra.Command{
	Use:   "food",
	Short: "Manage foods",
}

// --- add ---

var (
	addName     string
	addCalories float64
	addProtein  float64
	addCarbs    float64
	addFat      float64
)

var foodAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Create a new food",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, close, err := client.NewFoodClient(serverAddr)
		if err != nil {
			return err
		}
		defer close()

		resp, err := c.CreateFood(context.Background(), &foodv1.CreateFoodRequest{
			Name:     addName,
			Calories: addCalories,
			ProteinG: addProtein,
			CarbsG:   addCarbs,
			FatG:     addFat,
		})
		if err != nil {
			return err
		}

		fmt.Printf("created food #%d: %s\n", resp.Food.Id, resp.Food.Name)
		return nil
	},
}

// --- ls ---

var foodLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all foods",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, close, err := client.NewFoodClient(serverAddr)
		if err != nil {
			return err
		}
		defer close()

		resp, err := c.ListFoods(context.Background(), &foodv1.ListFoodsRequest{})
		if err != nil {
			return err
		}

		for _, f := range resp.Foods {
			fmt.Printf("#%d  %-20s  cal:%.0f  protein:%.1fg  carbs:%.1fg  fat:%.1fg\n",
				f.Id, f.Name, f.Calories, f.ProteinG, f.CarbsG, f.FatG)
		}
		return nil
	},
}

// --- rm ---

var foodRmCmd = &cobra.Command{
	Use:   "rm [id]",
	Short: "Delete a food by id",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var id int32
		if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
			return fmt.Errorf("invalid id: %s", args[0])
		}

		c, close, err := client.NewFoodClient(serverAddr)
		if err != nil {
			return err
		}
		defer close()

		_, err = c.DeleteFood(context.Background(), &foodv1.DeleteFoodRequest{Id: id})
		if err != nil {
			return err
		}

		fmt.Printf("deleted food #%d\n", id)
		return nil
	},
}

func init() {
	foodAddCmd.Flags().StringVar(&addName, "name", "", "food name")
	foodAddCmd.Flags().Float64Var(&addCalories, "calories", 0, "calories")
	foodAddCmd.Flags().Float64Var(&addProtein, "protein", 0, "protein (g)")
	foodAddCmd.Flags().Float64Var(&addCarbs, "carbs", 0, "carbs (g)")
	foodAddCmd.Flags().Float64Var(&addFat, "fat", 0, "fat (g)")
	foodAddCmd.MarkFlagRequired("name")

	foodCmd.AddCommand(foodAddCmd, foodLsCmd, foodRmCmd)
}
