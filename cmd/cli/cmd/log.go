package cmd

import (
	"context"
	"fmt"

	logv1 "github.com/ethandilley/myfitnesspal/gen/proto/log/v1"
	"github.com/ethandilley/myfitnesspal/internal/client"

	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Log foods and view daily macros",
}

// --- add ---

var (
	logFoodID     int32
	logMultiplier float64
	logDate       string
)

var logAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Log a food",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, close, err := client.NewLogClient(serverAddr)
		if err != nil {
			return err
		}
		defer close()

		resp, err := c.CreateLogEntry(context.Background(), &logv1.CreateLogEntryRequest{
			FoodId:     logFoodID,
			Multiplier: logMultiplier,
			LoggedAt:   logDate,
		})
		if err != nil {
			return err
		}

		fmt.Printf("logged entry #%d\n", resp.LogEntry.Id)
		return nil
	},
}

// --- rm ---

var logRmCmd = &cobra.Command{
	Use:   "rm [id]",
	Short: "Delete a log entry by id",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var id int32
		if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
			return fmt.Errorf("invalid id: %s", args[0])
		}

		c, close, err := client.NewLogClient(serverAddr)
		if err != nil {
			return err
		}
		defer close()

		_, err = c.DeleteLogEntry(context.Background(), &logv1.DeleteLogEntryRequest{Id: id})
		if err != nil {
			return err
		}

		fmt.Printf("deleted log entry #%d\n", id)
		return nil
	},
}

// --- ls ---

var logLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all log entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, close, err := client.NewLogClient(serverAddr)
		if err != nil {
			return err
		}
		defer close()

		resp, err := c.ListLogEntries(context.Background(), &logv1.ListLogEntriesRequest{})
		if err != nil {
			return err
		}

		for _, e := range resp.LogEntries {
			fmt.Printf("#%d  food:%d  x%.1f  %s\n", e.Id, e.FoodId, e.Multiplier, e.LoggedAt)
		}
		return nil
	},
}

// --- today ---

var todayDate string

var logTodayCmd = &cobra.Command{
	Use:   "today",
	Short: "Show a day's log entries and macro totals (defaults to today)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, close, err := client.NewLogClient(serverAddr)
		if err != nil {
			return err
		}
		defer close()

		resp, err := c.ListLogEntriesByDate(context.Background(), &logv1.ListLogEntriesByDateRequest{
			Date: todayDate,
		})
		if err != nil {
			return err
		}

		for _, e := range resp.LogEntries {
			fmt.Printf("#%d  food:%d  x%.1f\n", e.Id, e.FoodId, e.Multiplier)
		}

		if resp.Totals != nil {
			t := resp.Totals
			fmt.Printf("\ntotals: cal:%.0f  protein:%.1fg  carbs:%.1fg  fat:%.1fg",
				t.Calories, t.ProteinG, t.CarbsG, t.FatG)
		}
		return nil
	},
}

func init() {
	logAddCmd.Flags().Int32Var(&logFoodID, "food-id", 0, "food id")
	logAddCmd.Flags().Float64Var(&logMultiplier, "multiplier", 1, "multiplier")
	logAddCmd.Flags().StringVar(&logDate, "date", "", "date (YYYY-MM-DD), defaults to today")
	logAddCmd.MarkFlagRequired("food-id")

	logTodayCmd.Flags().StringVar(&todayDate, "date", "", "date (YYYY-MM-DD), defaults to today")

	logCmd.AddCommand(logAddCmd, logRmCmd, logLsCmd, logTodayCmd)
}
