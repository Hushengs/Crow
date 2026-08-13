package systemlog

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"runtime"
	"strings"

	"crow/internal/conf"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type handler struct {
	next slog.Handler
	db   *sql.DB
}

// NewHandler creates a slog handler that mirrors logs into system_log.
func NewHandler(next slog.Handler, data *conf.Data) (slog.Handler, func(), error) {
	db, err := sql.Open(data.GetDatabase().GetDriver(), data.GetDatabase().GetSource())
	if err != nil {
		return nil, nil, err
	}
	if err := db.PingContext(context.Background()); err != nil {
		_ = db.Close()
		return nil, nil, err
	}

	cleanup := func() {
		_ = db.Close()
	}
	return &handler{next: next, db: db}, cleanup, nil
}

func (h *handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *handler) Handle(ctx context.Context, record slog.Record) error {
	if err := h.next.Handle(ctx, record); err != nil {
		return err
	}

	message := strings.TrimSpace(record.Message)
	if message == "" || h.db == nil {
		return nil
	}

	attrs := map[string]any{}
	record.Attrs(func(attr slog.Attr) bool {
		attrs[attr.Key] = attr.Value.Any()
		return true
	})
	if len(attrs) > 0 {
		if raw, err := json.Marshal(attrs); err == nil {
			message += " | " + string(raw)
		}
	}

	filePath := ""
	lineNumber := 0
	if record.PC != 0 {
		frames := runtime.CallersFrames([]uintptr{record.PC})
		frame, _ := frames.Next()
		filePath = frame.File
		lineNumber = frame.Line
	}

	_, _ = h.db.ExecContext(ctx, `
INSERT INTO system_log (log_uid, log_level, message, file_path, line_number)
VALUES (?, ?, ?, ?, ?)
`, uuid.NewString(), normalizeLevel(record.Level), message, nullablePath(filePath), nullableLine(lineNumber))
	return nil
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &handler{
		next: h.next.WithAttrs(attrs),
		db:   h.db,
	}
}

func (h *handler) WithGroup(name string) slog.Handler {
	return &handler{
		next: h.next.WithGroup(name),
		db:   h.db,
	}
}

func normalizeLevel(level slog.Level) string {
	switch {
	case level <= slog.LevelDebug:
		return "DEBUG"
	case level < slog.LevelWarn:
		return "INFO"
	case level < slog.LevelError:
		return "WARN"
	default:
		return "ERROR"
	}
}

func nullableLine(value int) any {
	if value <= 0 {
		return nil
	}
	return value
}

func nullablePath(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
