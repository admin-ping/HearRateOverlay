package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"

	"HearRateOverlay/internal/stats"
)

// Database wraps SQLite operations for session and heart rate data.
type Database struct {
	mu  sync.Mutex
	db  *sql.DB
	dir string
}

// New opens or creates the SQLite database at the given path.
func New(dbPath string) (*Database, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// SQLite in WAL mode with single connection
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	d := &Database{db: db, dir: dir}
	if err := d.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return d, nil
}

func (d *Database) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			avg_hr REAL,
			max_hr INTEGER,
			min_hr INTEGER,
			total_beats INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS heart_rate_samples (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			bpm INTEGER NOT NULL,
			FOREIGN KEY (session_id) REFERENCES sessions(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_samples_session ON heart_rate_samples(session_id, timestamp)`,
	}

	for _, q := range queries {
		if _, err := d.db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

// SaveSession inserts or updates a session record.
func (d *Database) SaveSession(s *stats.Session) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	summary := s.Summary()
	_, err := d.db.Exec(`
		INSERT OR REPLACE INTO sessions (id, start_time, end_time, avg_hr, max_hr, min_hr, total_beats)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		summary.ID, summary.StartTime, s.EndTime, summary.AvgHR, summary.MaxHR, summary.MinHR, s.TotalBeats,
	)
	return err
}

// SaveHeartRateSamples batch-inserts heart rate readings.
func (d *Database) SaveHeartRateSamples(sessionID string, samples []struct {
	Timestamp time.Time
	BPM       uint8
}) error {
	if len(samples) == 0 {
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO heart_rate_samples (session_id, timestamp, bpm)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range samples {
		if _, err := stmt.Exec(sessionID, s.Timestamp, s.BPM); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetSessions returns all session summaries ordered by start time descending.
func (d *Database) GetSessions(limit int) ([]stats.SessionSummary, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	rows, err := d.db.Query(`
		SELECT id, start_time, end_time, avg_hr, max_hr, min_hr
		FROM sessions
		ORDER BY start_time DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []stats.SessionSummary
	for rows.Next() {
		var s stats.SessionSummary
		var endTime sql.NullTime
		if err := rows.Scan(&s.ID, &s.StartTime, &endTime, &s.AvgHR, &s.MaxHR, &s.MinHR); err != nil {
			return nil, err
		}
		if endTime.Valid {
			s.EndTime = &endTime.Time
			s.Duration = endTime.Time.Sub(s.StartTime).Round(time.Second).String()
		} else {
			s.Duration = time.Since(s.StartTime).Round(time.Second).String()
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// GetHeartRateSamples returns HR samples for a session.
func (d *Database) GetHeartRateSamples(sessionID string) ([]struct {
	Timestamp time.Time
	BPM       uint8
}, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	rows, err := d.db.Query(`
		SELECT timestamp, bpm FROM heart_rate_samples
		WHERE session_id = ?
		ORDER BY timestamp ASC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var samples []struct {
		Timestamp time.Time
		BPM       uint8
	}
	for rows.Next() {
		var s struct {
			Timestamp time.Time
			BPM       uint8
		}
		if err := rows.Scan(&s.Timestamp, &s.BPM); err != nil {
			return nil, err
		}
		samples = append(samples, s)
	}
	return samples, rows.Err()
}

// Close closes the database connection.
func (d *Database) Close() error {
	return d.db.Close()
}
