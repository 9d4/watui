package chatstore

import (
	"context"
	"database/sql"
	"time"

	"github.com/9d4/watui/roomlist"
	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

type SyncState struct {
	Progress   int
	ChunkOrder int
	SyncType   string
	InProgress bool
	UpdatedAt  time.Time
}

func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		db.Close()
		return nil, err
	}

	s := &Store{db: db}
	if err := s.ensureSchema(); err != nil {
		db.Close()
		return nil, err
	}

	return s, nil
}

func (s *Store) ensureSchema() error {
	const chats = `
CREATE TABLE IF NOT EXISTS chat_rooms (
	jid TEXT PRIMARY KEY,
	title TEXT,
	last_message TEXT,
	last_ts INTEGER,
	unread_count INTEGER,
	updated_at INTEGER
);`

	const syncState = `
CREATE TABLE IF NOT EXISTS sync_state (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	progress INTEGER,
	chunk_order INTEGER,
	sync_type TEXT,
	in_progress INTEGER,
	updated_at INTEGER
);`

	_, err := s.db.Exec(chats + syncState)
	return err
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) LoadAll(ctx context.Context) ([]roomlist.Room, SyncState, error) {
	if s == nil {
		return nil, SyncState{}, nil
	}

	rows, err := s.db.QueryContext(ctx, `SELECT jid, title, last_message, last_ts, unread_count FROM chat_rooms ORDER BY last_ts DESC, jid ASC`)
	if err != nil {
		return nil, SyncState{}, err
	}
	defer rows.Close()

	var rooms []roomlist.Room
	for rows.Next() {
		var (
			jid, title, lastMessage sql.NullString
			lastTS                  sql.NullInt64
			unread                  sql.NullInt64
		)
		if err := rows.Scan(&jid, &title, &lastMessage, &lastTS, &unread); err != nil {
			return nil, SyncState{}, err
		}

		var ts time.Time
		if lastTS.Valid {
			ts = time.Unix(lastTS.Int64, 0)
		}

		room := roomlist.Room{
			ID:          jid.String,
			Title:       title.String,
			LastMessage: lastMessage.String,
			Time:        ts,
			UnreadCount: int(unread.Int64),
		}
		rooms = append(rooms, room)
	}

	var state SyncState
	var inProgressInt sql.NullInt64
	var updatedUnix sql.NullInt64
	err = s.db.QueryRowContext(ctx, `SELECT progress, chunk_order, sync_type, in_progress, updated_at FROM sync_state WHERE id = 1`).Scan(
		&state.Progress,
		&state.ChunkOrder,
		&state.SyncType,
		&inProgressInt,
		&updatedUnix,
	)
	if err == sql.ErrNoRows {
		err = nil
	} else if err == nil {
		state.InProgress = inProgressInt.Int64 != 0
		if updatedUnix.Valid {
			state.UpdatedAt = time.Unix(updatedUnix.Int64, 0)
		}
	}

	return rooms, state, err
}

func (s *Store) PersistHistory(ctx context.Context, rooms []roomlist.Room, state SyncState) error {
	if s == nil {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if len(rooms) > 0 {
		stmt, errPrepare := tx.PrepareContext(ctx, `
INSERT INTO chat_rooms (jid, title, last_message, last_ts, unread_count, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(jid) DO UPDATE SET
	title=excluded.title,
	last_message=excluded.last_message,
	last_ts=excluded.last_ts,
	unread_count=excluded.unread_count,
	updated_at=excluded.updated_at`)
		if errPrepare != nil {
			err = errPrepare
			return err
		}
		defer stmt.Close()

		for _, room := range rooms {
			_, err = stmt.ExecContext(ctx,
				room.ID,
				room.Title,
				room.LastMessage,
				room.Time.Unix(),
				room.UnreadCount,
				time.Now().Unix(),
			)
			if err != nil {
				return err
			}
		}
	}

	state.InProgress = state.Progress < 100
	state.UpdatedAt = time.Now()

	_, err = tx.ExecContext(ctx, `
INSERT INTO sync_state (id, progress, chunk_order, sync_type, in_progress, updated_at)
VALUES (1, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
	progress=excluded.progress,
	chunk_order=excluded.chunk_order,
	sync_type=excluded.sync_type,
	in_progress=excluded.in_progress,
	updated_at=excluded.updated_at
`, state.Progress, state.ChunkOrder, state.SyncType, boolToInt(state.InProgress), state.UpdatedAt.Unix())
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) UpsertRoom(ctx context.Context, room roomlist.Room) error {
	if s == nil {
		return nil
	}

	_, err := s.db.ExecContext(ctx, `
INSERT INTO chat_rooms (jid, title, last_message, last_ts, unread_count, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(jid) DO UPDATE SET
	title=excluded.title,
	last_message=excluded.last_message,
	last_ts=excluded.last_ts,
	unread_count=excluded.unread_count,
	updated_at=excluded.updated_at
`, room.ID, room.Title, room.LastMessage, room.Time.Unix(), room.UnreadCount, time.Now().Unix())
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
