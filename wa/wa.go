package wa

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type Manager struct {
	dbLog     waLog.Logger
	container *sqlstore.Container
}

func (m *Manager) SessionCounts() (int, error) {
	d, err := m.container.GetAllDevices(context.Background())
	return len(d), err
}

func NewManager(logger zerolog.Logger) *Manager {
	dbLog := waLog.Zerolog(logger.With().Str("log", "db").Logger())

	ctx := context.Background()
	container, err := sqlstore.New(ctx, "sqlite3", "file:watui.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}

	m := &Manager{
		dbLog:     dbLog,
		container: container,
	}
	return m
}

func CreateFileLogger(path string) (zerolog.Logger, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return zerolog.Nop(), err
	}
	z := zerolog.New(file)
	return z, nil
}
