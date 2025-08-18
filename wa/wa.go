package wa

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"

	_ "github.com/mattn/go-sqlite3"
)

type Manager struct {
	dbLog waLog.Logger
	waLog waLog.Logger
	C     *sqlstore.Container
}

func (m *Manager) SessionCounts() (int, error) {
	d, err := m.C.GetAllDevices(context.Background())
	return len(d), err
}

func NewManager(logger zerolog.Logger) *Manager {
	dbLog := waLog.Zerolog(logger.With().Str("log", "db").Logger())
	waLog := waLog.Zerolog(logger.With().Str("log", "wa").Logger())

	ctx := context.Background()
	container, err := sqlstore.New(ctx, "sqlite3", "file:watui.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}

	m := &Manager{
		dbLog: dbLog,
		waLog: waLog,
		C:     container,
	}
	return m
}

func (m *Manager) WaLog() waLog.Logger {
	return m.waLog
}

func CreateFileLogger(path string) (zerolog.Logger, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return zerolog.Nop(), err
	}
	z := zerolog.New(file)
	return z, nil
}
