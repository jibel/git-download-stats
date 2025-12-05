package internal

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	defaultDBPath = "github-stats.db"
	statsTable    = `
	CREATE TABLE IF NOT EXISTS stats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		owner TEXT NOT NULL,
		repo TEXT NOT NULL,
		tag TEXT NOT NULL,
		release_name TEXT NOT NULL,
		total_downloads INTEGER NOT NULL,
		fetched_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_owner_repo_fetched 
		ON stats(owner, repo, fetched_at DESC);
	`
	assetsTable = `
	CREATE TABLE IF NOT EXISTS assets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		stat_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		download_count INTEGER NOT NULL,
		size INTEGER NOT NULL,
		content_type TEXT,
		FOREIGN KEY (stat_id) REFERENCES stats(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_stat_id ON assets(stat_id);
	`
)

type Database struct {
	db   *sql.DB
	path string
}

// NewDatabase creates or opens an SQLite database.
func NewDatabase(dbPath string) (*Database, error) {
	if dbPath == "" {
		dbPath = defaultDBPath
	}

	// Ensure directory exists
	if dir := filepath.Dir(dbPath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	d := &Database{db: db, path: dbPath}

	// Create tables
	if err := d.createTables(); err != nil {
		db.Close()
		return nil, err
	}

	return d, nil
}

func (d *Database) createTables() error {
	if _, err := d.db.Exec(statsTable); err != nil {
		return fmt.Errorf("failed to create stats table: %w", err)
	}

	if _, err := d.db.Exec(assetsTable); err != nil {
		return fmt.Errorf("failed to create assets table: %w", err)
	}

	return nil
}

// StoreStats stores release statistics in the database.
func (d *Database) StoreStats(stats *ReleaseStats) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, rel := range stats.Releases {
		var statID int64
		err := tx.QueryRow(
			`INSERT INTO stats (owner, repo, tag, release_name, total_downloads, fetched_at, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?)
			 RETURNING id`,
			stats.Owner, stats.Repo, rel.Tag, rel.Name, rel.TotalDownloads, stats.FetchedAt, rel.CreatedAt,
		).Scan(&statID)
		if err != nil {
			return fmt.Errorf("failed to insert stat: %w", err)
		}

		// Insert assets
		for _, asset := range rel.Assets {
			_, err := tx.Exec(
				`INSERT INTO assets (stat_id, name, download_count, size, content_type)
				 VALUES (?, ?, ?, ?, ?)`,
				statID, asset.Name, asset.DownloadCount, asset.Size, asset.ContentType,
			)
			if err != nil {
				return fmt.Errorf("failed to insert asset: %w", err)
			}
		}
	}

	return tx.Commit()
}

// GetLatestStats retrieves the most recent statistics for a given owner/repo.
func (d *Database) GetLatestStats(owner, repo string) (*ReleaseStats, error) {
	stats := &ReleaseStats{
		Owner:    owner,
		Repo:     repo,
		Releases: make([]Release, 0),
	}

	rows, err := d.db.Query(
		`SELECT id, tag, release_name, total_downloads, fetched_at, created_at
		 FROM stats
		 WHERE owner = ? AND repo = ?
		 ORDER BY fetched_at DESC
		 LIMIT 1`,
		owner, repo,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query latest stats: %w", err)
	}
	defer rows.Close()

	// Get the first result to determine FetchedAt
	var statID int64
	if rows.Next() {
		var rel Release
		err := rows.Scan(&statID, &rel.Tag, &rel.Name, &rel.TotalDownloads, &stats.FetchedAt, &rel.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stats row: %w", err)
		}

		// Get assets for this release
		assets, err := d.getAssets(statID)
		if err != nil {
			return nil, err
		}
		rel.Assets = assets
		stats.Releases = append(stats.Releases, rel)
		stats.TotalDownloads += rel.TotalDownloads
	}

	return stats, nil
}

// GetStatsHistory retrieves all statistics for a given owner/repo, ordered by fetch date.
func (d *Database) GetStatsHistory(owner, repo string, limit int) ([]ReleaseStats, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := d.db.Query(
		`SELECT DISTINCT fetched_at FROM stats
		 WHERE owner = ? AND repo = ?
		 ORDER BY fetched_at DESC
		 LIMIT ?`,
		owner, repo, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query history: %w", err)
	}
	defer rows.Close()

	var fetchDates []time.Time
	for rows.Next() {
		var fetchedAt time.Time
		if err := rows.Scan(&fetchedAt); err != nil {
			return nil, fmt.Errorf("failed to scan fetch date: %w", err)
		}
		fetchDates = append(fetchDates, fetchedAt)
	}

	result := make([]ReleaseStats, 0, len(fetchDates))
	for _, fetchedAt := range fetchDates {
		stats := &ReleaseStats{
			Owner:    owner,
			Repo:     repo,
			Releases: make([]Release, 0),
			FetchedAt: fetchedAt,
		}

		statRows, err := d.db.Query(
			`SELECT id, tag, release_name, total_downloads, created_at
			 FROM stats
			 WHERE owner = ? AND repo = ? AND fetched_at = ?
			 ORDER BY total_downloads DESC`,
			owner, repo, fetchedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to query stats for date: %w", err)
		}

		for statRows.Next() {
			var statID int64
			var rel Release
			if err := statRows.Scan(&statID, &rel.Tag, &rel.Name, &rel.TotalDownloads, &rel.CreatedAt); err != nil {
				statRows.Close()
				return nil, fmt.Errorf("failed to scan stat row: %w", err)
			}

			assets, err := d.getAssets(statID)
			if err != nil {
				statRows.Close()
				return nil, err
			}
			rel.Assets = assets
			stats.Releases = append(stats.Releases, rel)
			stats.TotalDownloads += rel.TotalDownloads
		}
		statRows.Close()

		result = append(result, *stats)
	}

	return result, nil
}

// GetStatsBetween retrieves statistics collected between two dates.
func (d *Database) GetStatsBetween(owner, repo string, start, end time.Time) ([]ReleaseStats, error) {
	rows, err := d.db.Query(
		`SELECT DISTINCT fetched_at FROM stats
		 WHERE owner = ? AND repo = ? AND fetched_at BETWEEN ? AND ?
		 ORDER BY fetched_at DESC`,
		owner, repo, start, end,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query stats between dates: %w", err)
	}
	defer rows.Close()

	var fetchDates []time.Time
	for rows.Next() {
		var fetchedAt time.Time
		if err := rows.Scan(&fetchedAt); err != nil {
			return nil, fmt.Errorf("failed to scan fetch date: %w", err)
		}
		fetchDates = append(fetchDates, fetchedAt)
	}

	result := make([]ReleaseStats, 0, len(fetchDates))
	for _, fetchedAt := range fetchDates {
		stats := &ReleaseStats{
			Owner:    owner,
			Repo:     repo,
			Releases: make([]Release, 0),
			FetchedAt: fetchedAt,
		}

		statRows, err := d.db.Query(
			`SELECT id, tag, release_name, total_downloads, created_at
			 FROM stats
			 WHERE owner = ? AND repo = ? AND fetched_at = ?
			 ORDER BY total_downloads DESC`,
			owner, repo, fetchedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to query stats for date: %w", err)
		}

		for statRows.Next() {
			var statID int64
			var rel Release
			if err := statRows.Scan(&statID, &rel.Tag, &rel.Name, &rel.TotalDownloads, &rel.CreatedAt); err != nil {
				statRows.Close()
				return nil, fmt.Errorf("failed to scan stat row: %w", err)
			}

			assets, err := d.getAssets(statID)
			if err != nil {
				statRows.Close()
				return nil, err
			}
			rel.Assets = assets
			stats.Releases = append(stats.Releases, rel)
			stats.TotalDownloads += rel.TotalDownloads
		}
		statRows.Close()

		result = append(result, *stats)
	}

	return result, nil
}

func (d *Database) getAssets(statID int64) ([]Asset, error) {
	rows, err := d.db.Query(
		`SELECT name, download_count, size, content_type FROM assets WHERE stat_id = ?`,
		statID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query assets: %w", err)
	}
	defer rows.Close()

	assets := make([]Asset, 0)
	for rows.Next() {
		var asset Asset
		if err := rows.Scan(&asset.Name, &asset.DownloadCount, &asset.Size, &asset.ContentType); err != nil {
			return nil, fmt.Errorf("failed to scan asset: %w", err)
		}
		assets = append(assets, asset)
	}

	return assets, nil
}

// Close closes the database connection.
func (d *Database) Close() error {
	return d.db.Close()
}
