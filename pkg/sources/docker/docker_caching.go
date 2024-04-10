package docker

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// connectToLayersDB opens the SQLite database and returns a connection.
func connectToLayersDB(dbName string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}
	return conn, nil
}

// initializeDB initializes the SQLite database with the Digest table.
// It returns an error if encountered
func InitializeLayersDB(dbName string) error {
	conn, err := connectToLayersDB(dbName)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Execute SQL to create the Digest table if it does not exist
	_, err = conn.Exec("CREATE TABLE IF NOT EXISTS digest (digest TEXT UNIQUE, verified BOOLEAN, unverified_with_error BOOLEAN, completed BOOLEAN)")
	if err != nil {
		return err
	}
	return nil
}

// AddToLayersDB inserts a digest into the database. It ignores if the digest already exists.
// Sets the secret and completed fields to false.
// It returns an error if encountered
func AddDigestToLayersDB(dbName string, digest string) error {
	conn, err := connectToLayersDB(dbName)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Exec("INSERT OR REPLACE INTO digest (digest, verified, unverified_with_error, completed) VALUES (?, ?, ?, ?)", digest, false, false, false)
	if err != nil {
		return err
	}
	return nil
}

// UpdateStatusInLayersDB updates the completed field of a digest in the database.
// It returns an error if encountered
func UpdateStatusInLayersDB(dbName string, digest string, completed bool) error {
	conn, err := connectToLayersDB(dbName)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Prepare the SQL statement for update
	query := "UPDATE digest SET completed = ? WHERE digest = ?"
	_, err = conn.Exec(query, completed, digest)
	if err != nil {
		return err
	}
	return nil
}

// SkipDockerLayer will return True iff the layer has been scanned before and no secrets were found.
// This function factors in previously unverified secrets that had errors, since they could still be valid.
// It returns an error if encountered
func SkipDockerLayer(dbName string, digest string) (bool, error) {
	conn, err := connectToLayersDB(dbName)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	// Query the database for the digest
	rows, err := conn.Query("SELECT verified, unverified_with_error FROM digest WHERE digest = ? and completed = true", digest)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	if rows.Next() {
		var verified, unverified_with_error bool
		err = rows.Scan(&verified, &unverified_with_error)
		if err != nil {
			return false, err
		}
		if !verified && !unverified_with_error {
			return true, nil
		}
	}
	return false, nil
}

// SetVerified sets the verified field of a digest in the database to true.
// It returns an error if encountered
func SetVerified(dbName string, digest string) error {
	conn, err := connectToLayersDB(dbName)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Prepare the SQL statement for update
	query := "UPDATE digest SET verified = true WHERE digest = ?"
	_, err = conn.Exec(query, digest)
	if err != nil {
		return err
	}
	return nil
}

// SetUnverifiedWithError sets the unverified_with_error field of a digest in the database to true.
// It returns an error if encountered
func SetUnverifiedWithError(dbName string, digest string) error {
	conn, err := connectToLayersDB(dbName)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Prepare the SQL statement for update
	query := "UPDATE digest SET unverified_with_error = true WHERE digest = ?"
	_, err = conn.Exec(query, digest)
	if err != nil {
		return err
	}
	return nil
}
