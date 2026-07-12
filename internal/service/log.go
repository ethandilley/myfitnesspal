package service

import (
	"context"
	"log"
	"time"

	logv1 "github.com/ethandilley/myfitnesspal/gen/proto/log/v1"
	"github.com/ethandilley/myfitnesspal/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

const dateLayout = "2006-01-02"

type LogService struct {
	logv1.UnimplementedLogServiceServer

	q *db.Queries
}

func NewLogService(q *db.Queries) *LogService {
	return &LogService{q: q}
}

func (s *LogService) CreateLogEntry(ctx context.Context, req *logv1.CreateLogEntryRequest) (*logv1.CreateLogEntryResponse, error) {
	log.Printf("CreateLogEntry called: %+v", req)

	loggedAt, err := stringToDate(req.LoggedAt)
	if err != nil {
		return nil, err
	}

	row, err := s.q.CreateLogEntry(ctx, db.CreateLogEntryParams{
		FoodID:     req.GetFoodId(),
		Multiplier: floatToNumeric(req.Multiplier),
		LoggedAt:   loggedAt,
	})
	if err != nil {
		return nil, err
	}
	return &logv1.CreateLogEntryResponse{LogEntry: toProtoLogEntry(row)}, nil
}

func (s *LogService) DeleteLogEntry(ctx context.Context, req *logv1.DeleteLogEntryRequest) (*logv1.DeleteLogEntryResponse, error) {
	log.Printf("DeleteLogEntry called: %+v", req)
	err := s.q.DeleteLogEntry(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &logv1.DeleteLogEntryResponse{}, nil
}

func (s *LogService) ListLogEntries(ctx context.Context, req *logv1.ListLogEntriesRequest) (*logv1.ListLogEntriesResponse, error) {
	log.Printf("ListLogEntries called: %+v", req)
	rows, err := s.q.ListLogEntries(ctx)
	if err != nil {
		return nil, err
	}
	entries := make([]*logv1.LogEntry, len(rows))
	for i, row := range rows {
		entries[i] = toProtoLogEntry(row)
	}
	return &logv1.ListLogEntriesResponse{LogEntries: entries}, nil
}

func (s *LogService) ListLogEntriesByDate(ctx context.Context, req *logv1.ListLogEntriesByDateRequest) (*logv1.ListLogEntriesByDateResponse, error) {
	log.Printf("ListLogEntriesByDate called: %+v", req)
	date, err := resolveDate(req.Date)
	if err != nil {
		return nil, err
	}

	rows, err := s.q.ListLogEntriesByDate(ctx, date)
	if err != nil {
		return nil, err
	}
	entries := make([]*logv1.LogEntry, len(rows))
	for i, row := range rows {
		entries[i] = toProtoLogEntry(row)
	}
	totals, err := s.q.GetMacroTotalsByDate(ctx, date)
	if err != nil {
		return nil, err
	}
	return &logv1.ListLogEntriesByDateResponse{LogEntries: entries,
		Totals: &logv1.MacroTotals{
			Calories: numericToFloat(totals.Calories),
			ProteinG: numericToFloat(totals.ProteinG),
			CarbsG: numericToFloat(totals.CarbsG),
			FatG: numericToFloat(totals.FatG),
		}}, nil
}

func toProtoLogEntry(row db.LogEntry) *logv1.LogEntry {
	return &logv1.LogEntry{
		Id:         row.ID,
		FoodId:     row.FoodID,
		Multiplier: numericToFloat(row.Multiplier),
		LoggedAt:   dateToString(row.LoggedAt),
	}
}

// stringToDate converts a "YYYY-MM-DD" string into a pgtype.Date.
// An empty string returns an invalid (NULL) Date, which lets the
// CreateLogEntry query's COALESCE(..., CURRENT_DATE) fall back to today.
func stringToDate(s string) (pgtype.Date, error) {
	if s == "" {
		return pgtype.Date{}, nil
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return pgtype.Date{}, err
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}

// resolveDate is like stringToDate, but an empty string resolves to
// today's date instead of NULL, since ListLogEntriesByDate needs a
// concrete date to filter on rather than something the DB can default.
func resolveDate(s string) (pgtype.Date, error) {
	if s == "" {
		now := time.Now().UTC()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		return pgtype.Date{Time: today, Valid: true}, nil
	}
	return stringToDate(s)
}

func dateToString(d pgtype.Date) string {
	if !d.Valid {
		return ""
	}
	return d.Time.Format(dateLayout)
}
