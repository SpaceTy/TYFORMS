import csv
import sqlite3
from datetime import datetime

def create_applications_table(cursor):
    cursor.execute('''
    CREATE TABLE IF NOT EXISTS Applications (
        Id INTEGER PRIMARY KEY AUTOINCREMENT,
        DiscordUsername TEXT NOT NULL,
        MinecraftUsername TEXT NOT NULL UNIQUE,
        Age INTEGER NOT NULL,
        FavoriteAboutMinecraft TEXT NOT NULL,
        UnderstandingOfSMP TEXT NOT NULL,
        JoinedDiscord INTEGER NOT NULL,
        SubmissionDate TEXT NOT NULL,
        IsReviewed INTEGER NOT NULL,
        ReviewedAt TEXT,
        ReviewNotes TEXT,
        AcceptanceStatus TEXT
    )
    ''')

def import_applications():
    # Connect to the SQLite database
    conn = sqlite3.connect('application.db')
    cursor = conn.cursor()

    # Create the table if it doesn't exist
    create_applications_table(cursor)

    # Keep track of statistics
    total_rows = 0
    imported_rows = 0
    skipped_rows = 0

    # Read the CSV file
    with open('data.csv', 'r', encoding='utf-8') as file:
        csv_reader = csv.DictReader(file)
        
        for row in csv_reader:
            total_rows += 1
            try:
                # Convert string values to appropriate types
                is_reviewed = row['IsReviewed'] == '1'
                joined_discord = row['JoinedDiscord'] == '1'
                age = int(row['Age'])
                
                # Convert datetime strings to datetime objects
                submission_date = datetime.strptime(row['SubmissionDate'].split('.')[0], '%Y-%m-%d %H:%M:%S')
                reviewed_at = datetime.strptime(row['ReviewedAt'].split('.')[0], '%Y-%m-%d %H:%M:%S') if row['ReviewedAt'] else None
                
                # Insert the data into the Applications table
                cursor.execute('''
                    INSERT INTO Applications (
                        DiscordUsername, MinecraftUsername, Age, 
                        FavoriteAboutMinecraft, UnderstandingOfSMP, 
                        JoinedDiscord, SubmissionDate, IsReviewed, 
                        ReviewedAt, ReviewNotes, AcceptanceStatus
                    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                ''', (
                    row['DiscordUsername'],
                    row['MinecraftUsername'],
                    age,
                    row['FavoriteAboutMinecraft'],
                    row['UnderstandingOfSMP'],
                    joined_discord,
                    submission_date,
                    is_reviewed,
                    reviewed_at,
                    row['ReviewNotes'],
                    row['AcceptanceStatus']
                ))
                imported_rows += 1
            except sqlite3.IntegrityError:
                # Skip duplicate MinecraftUsername entries
                skipped_rows += 1
                continue
            except Exception as e:
                print(f"Error processing row: {row}")
                raise e

    # Commit the changes and close the connection
    conn.commit()
    conn.close()

    return total_rows, imported_rows, skipped_rows

if __name__ == '__main__':
    try:
        total, imported, skipped = import_applications()
        print(f"Import completed successfully:")
        print(f"Total rows processed: {total}")
        print(f"Rows imported: {imported}")
        print(f"Rows skipped (duplicates): {skipped}")
    except Exception as e:
        print(f"An error occurred: {str(e)}") 