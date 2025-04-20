package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

const (
	csvFilePath = "applications.csv" // Path to your input CSV file
	dbFilePath  = "applications.db"  // Path to the SQLite database file
	// Go's reference time: Mon Jan 2 15:04:05 -0700 MST 2006
	// Match the input format: 03/12/2025 19:41:17
	dateTimeLayout = "01/02/2006 15:04:05"
)

func main() {
	// --- 1. Open CSV File ---
	file, err := os.Open(csvFilePath)
	if err != nil {
		log.Fatalf("Error opening CSV file %s: %v", csvFilePath, err)
	}
	defer file.Close()

	// --- 2. Setup CSV Reader ---
	reader := csv.NewReader(file)
	// You might need to adjust this if your CSV uses a different delimiter
	// reader.Comma = '\t' // Example for tab-separated

	// --- 3. Open/Create SQLite Database ---
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		log.Fatalf("Error opening database %s: %v", dbFilePath, err)
	}
	defer db.Close()

	// Ping to ensure the connection is valid
	if err := db.Ping(); err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	// --- 4. Create Table (if it doesn't exist) ---
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS applications (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		discord_username TEXT NOT NULL,
		minecraft_username TEXT NOT NULL UNIQUE,
		age INTEGER NOT NULL,
		favorite_about_minecraft TEXT NOT NULL,
		understanding_of_smp TEXT NOT NULL,
		joined_discord BOOLEAN NOT NULL,
		submission_date DATETIME NOT NULL,
		is_reviewed BOOLEAN NOT NULL DEFAULT FALSE,
		reviewed_at DATETIME,
		review_notes TEXT,
		acceptance_status TEXT NOT NULL DEFAULT 'pending'
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

	// --- 5. Prepare Insert Statement ---
	insertSQL := `
	INSERT INTO applications (
		discord_username, minecraft_username, age, favorite_about_minecraft,
		understanding_of_smp, joined_discord, submission_date, is_reviewed,
		reviewed_at, review_notes, acceptance_status
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	stmt, err := db.Prepare(insertSQL)
	if err != nil {
		log.Fatalf("Error preparing insert statement: %v", err)
	}
	defer stmt.Close()

	// --- 6. Read CSV and Insert Data ---
	// Read header row (and discard it)
	_, err = reader.Read()
	if err != nil {
		if err == io.EOF {
			log.Println("CSV file is empty or only contains a header.")
			return
		}
		log.Fatalf("Error reading CSV header: %v", err)
	}

	rowCount := 0
	successCount := 0
	errorCount := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break // End of file
		}
		if err != nil {
			log.Printf("Error reading CSV row: %v", err)
			errorCount++
			continue
		}
		rowCount++

		if len(record) != 12 {
			log.Printf("Skipping row %d: expected 12 columns, got %d", rowCount+1, len(record))
			errorCount++
			continue
		}

		// --- Parse and Convert Data ---
		// Skip ID (column 0), it's autoincrement
		discordUsername := record[1]
		minecraftUsername := record[2]

		age, err := strconv.Atoi(strings.TrimSpace(record[3]))
		if err != nil {
			// Handle potential non-integer age like "99"
			if strings.TrimSpace(record[3]) == "99" { // Specific case from data
				age = 99 // Or handle as invalid/log differently
				log.Printf("Warning: Row %d: Found age '99', using as integer.", rowCount+1)
			} else {
				log.Printf("Skipping row %d: invalid age '%s': %v", rowCount+1, record[3], err)
				errorCount++
				continue
			}
		}

		favoriteAboutMinecraft := record[4]
		understandingOfSmp := record[5]

		joinedDiscord, err := strconv.ParseBool(strings.TrimSpace(record[6]))
		if err != nil {
			log.Printf("Skipping row %d: invalid boolean for joined_discord '%s': %v", rowCount+1, record[6], err)
			errorCount++
			continue
		}

		submissionDate, err := time.Parse(dateTimeLayout, strings.TrimSpace(record[7]))
		if err != nil {
			log.Printf("Skipping row %d: invalid submission_date format '%s': %v", rowCount+1, record[7], err)
			errorCount++
			continue
		}

		isReviewed, err := strconv.ParseBool(strings.TrimSpace(record[8]))
		if err != nil {
			log.Printf("Skipping row %d: invalid boolean for is_reviewed '%s': %v", rowCount+1, record[8], err)
			errorCount++
			continue
		}

		// Handle nullable ReviewedAt (column 9)
		var reviewedAt sql.NullTime
		reviewedAtStr := strings.TrimSpace(record[9])
		if reviewedAtStr != "" {
			t, err := time.Parse(dateTimeLayout, reviewedAtStr)
			if err != nil {
				log.Printf("Skipping row %d: invalid reviewed_at format '%s': %v", rowCount+1, reviewedAtStr, err)
				errorCount++
				continue
			}
			reviewedAt = sql.NullTime{Time: t, Valid: true}
		} else {
			reviewedAt = sql.NullTime{Valid: false}
		}

		// Handle nullable ReviewNotes (column 10)
		var reviewNotes sql.NullString
		reviewNotesStr := strings.TrimSpace(record[10])
		if reviewNotesStr != "" {
			reviewNotes = sql.NullString{String: reviewNotesStr, Valid: true}
		} else {
			reviewNotes = sql.NullString{Valid: false}
		}

		// Handle AcceptanceStatus (column 11), default to 'pending' if empty
		acceptanceStatus := strings.TrimSpace(record[11])
		if acceptanceStatus == "" {
			acceptanceStatus = "pending" // Use DB default if CSV is empty
		}

		// --- Execute Insert ---
		_, err = stmt.Exec(
			discordUsername,
			minecraftUsername,
			age,
			favoriteAboutMinecraft,
			understandingOfSmp,
			joinedDiscord,
			submissionDate,
			isReviewed,
			reviewedAt,
			reviewNotes,
			acceptanceStatus,
		)
		if err != nil {
			// Check for unique constraint violation specifically
			if strings.Contains(err.Error(), "UNIQUE constraint failed: applications.minecraft_username") {
				log.Printf("Skipping row %d: Duplicate Minecraft username '%s'", rowCount+1, minecraftUsername)
			} else {
				log.Printf("Error inserting row %d: %v", rowCount+1, err)
			}
			errorCount++
		} else {
			successCount++
		}
	}

	// --- 7. Report Summary ---
	fmt.Println("------------------------------------")
	fmt.Printf("CSV Processing Complete.\n")
	fmt.Printf("Total rows processed: %d\n", rowCount)
	fmt.Printf("Successfully inserted: %d\n", successCount)
	fmt.Printf("Rows skipped/errored: %d\n", errorCount)
	fmt.Println("------------------------------------")
}
