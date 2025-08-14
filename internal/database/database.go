// Package database provides database functionality for the Discord bot framework.
package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/sawyer/go-discord-bots/internal/errors"
)

// DB represents a database connection with music bot functionality.
type DB struct {
	conn *sql.DB
}

// Playlist represents a music playlist.
type Playlist struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	OwnerID string `json:"owner_id"`
	GuildID string `json:"guild_id"`
	Songs   []Song `json:"songs"`
}

// Song represents a song in a playlist.
type Song struct {
	Title      string `json:"title"`
	URL        string `json:"url"`
	WebpageURL string `json:"webpage_url"`
	Duration   *int   `json:"duration,omitempty"`
	AddedAt    string `json:"added_at"`
}

// NewDB creates a new database connection.
func NewDB(databaseURL string) (*DB, error) {
	if databaseURL == "" {
		databaseURL = "bot.db" // Default SQLite database
	}

	conn, err := sql.Open("sqlite3", databaseURL)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to open database", err)
	}

	db := &DB{conn: conn}

	if err := db.migrate(); err != nil {
		return nil, errors.NewDatabaseError("failed to migrate database", err)
	}

	return db, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// migrate creates the necessary database tables.
func (db *DB) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS playlists (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		owner_id TEXT NOT NULL,
		guild_id TEXT NOT NULL,
		songs TEXT DEFAULT '[]',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_playlists_owner_guild ON playlists(owner_id, guild_id);
	`

	_, err := db.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}

// CreatePlaylist creates a new playlist.
func (db *DB) CreatePlaylist(ctx context.Context, name, ownerID, guildID string) (int, error) {
	query := `INSERT INTO playlists (name, owner_id, guild_id) VALUES (?, ?, ?)`

	result, err := db.conn.ExecContext(ctx, query, name, ownerID, guildID)
	if err != nil {
		return 0, errors.NewDatabaseError("failed to create playlist", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.NewDatabaseError("failed to get playlist ID", err)
	}

	return int(id), nil
}

// GetPlaylist retrieves a playlist by ID.
func (db *DB) GetPlaylist(ctx context.Context, playlistID int) (*Playlist, error) {
	query := `SELECT id, name, owner_id, guild_id, songs FROM playlists WHERE id = ?`

	var playlist Playlist
	var songsJSON string

	err := db.conn.QueryRowContext(ctx, query, playlistID).Scan(
		&playlist.ID,
		&playlist.Name,
		&playlist.OwnerID,
		&playlist.GuildID,
		&songsJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.NewDatabaseError("failed to get playlist", err)
	}

	// Parse songs JSON
	if err := json.Unmarshal([]byte(songsJSON), &playlist.Songs); err != nil {
		return nil, errors.NewDatabaseError("failed to parse songs JSON", err)
	}

	return &playlist, nil
}

// GetUserPlaylists retrieves all playlists for a user in a guild.
func (db *DB) GetUserPlaylists(ctx context.Context, ownerID, guildID string) ([]*Playlist, error) {
	query := `SELECT id, name, owner_id, guild_id, songs FROM playlists WHERE owner_id = ? AND guild_id = ? ORDER BY created_at DESC`

	rows, err := db.conn.QueryContext(ctx, query, ownerID, guildID)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to get user playlists", err)
	}
	defer func() { _ = rows.Close() }()

	var playlists []*Playlist

	for rows.Next() {
		var playlist Playlist
		var songsJSON string

		err := rows.Scan(
			&playlist.ID,
			&playlist.Name,
			&playlist.OwnerID,
			&playlist.GuildID,
			&songsJSON,
		)

		if err != nil {
			return nil, errors.NewDatabaseError("failed to scan playlist row", err)
		}

		// Parse songs JSON
		if err := json.Unmarshal([]byte(songsJSON), &playlist.Songs); err != nil {
			return nil, errors.NewDatabaseError("failed to parse songs JSON", err)
		}

		playlists = append(playlists, &playlist)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseError("error iterating playlist rows", err)
	}

	return playlists, nil
}

// AddSongToPlaylist adds a song to a playlist.
func (db *DB) AddSongToPlaylist(ctx context.Context, playlistID int, song Song) error {
	// Get current playlist
	playlist, err := db.GetPlaylist(ctx, playlistID)
	if err != nil {
		return err
	}

	if playlist == nil {
		return errors.NewNotFoundError("playlist not found")
	}

	// Add timestamp
	song.AddedAt = time.Now().Format(time.RFC3339)

	// Add song to list
	playlist.Songs = append(playlist.Songs, song)

	// Convert back to JSON
	songsJSON, err := json.Marshal(playlist.Songs)
	if err != nil {
		return errors.NewDatabaseError("failed to marshal songs JSON", err)
	}

	// Update database
	query := `UPDATE playlists SET songs = ? WHERE id = ?`
	_, err = db.conn.ExecContext(ctx, query, string(songsJSON), playlistID)
	if err != nil {
		return errors.NewDatabaseError("failed to update playlist", err)
	}

	return nil
}

// RemoveSongFromPlaylist removes a song from a playlist by index.
func (db *DB) RemoveSongFromPlaylist(ctx context.Context, playlistID, songIndex int) error {
	// Get current playlist
	playlist, err := db.GetPlaylist(ctx, playlistID)
	if err != nil {
		return err
	}

	if playlist == nil {
		return errors.NewNotFoundError("playlist not found")
	}

	// Check index bounds
	if songIndex < 0 || songIndex >= len(playlist.Songs) {
		return errors.NewValidationError("invalid song index")
	}

	// Remove song from list
	playlist.Songs = append(playlist.Songs[:songIndex], playlist.Songs[songIndex+1:]...)

	// Convert back to JSON
	songsJSON, err := json.Marshal(playlist.Songs)
	if err != nil {
		return errors.NewDatabaseError("failed to marshal songs JSON", err)
	}

	// Update database
	query := `UPDATE playlists SET songs = ? WHERE id = ?`
	_, err = db.conn.ExecContext(ctx, query, string(songsJSON), playlistID)
	if err != nil {
		return errors.NewDatabaseError("failed to update playlist", err)
	}

	return nil
}

// DeletePlaylist deletes a playlist.
func (db *DB) DeletePlaylist(ctx context.Context, playlistID int, ownerID string) error {
	query := `DELETE FROM playlists WHERE id = ? AND owner_id = ?`

	result, err := db.conn.ExecContext(ctx, query, playlistID, ownerID)
	if err != nil {
		return errors.NewDatabaseError("failed to delete playlist", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("playlist not found or not owned by user")
	}

	return nil
}
