package main

import (
  "database/sql"
  "encoding/hex"
	"flag"
	"fmt"
	"log"
  "time"
  "strconv"

  _ "github.com/lib/pq"
  "github.com/lib/pq"
)

type Database struct {
	conn          *sql.DB
	lastSeenID    int
	pollInterval  time.Duration
}

type Item struct {
  data int
	types  int
	quantity int
}

func NewDatabase(dbURL string, pollInterval int) (*Database, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
  }

  fmt.Println("Connected to database")

	// Get last inserted ID
	lastID, err := getLastInsertedID(db)
	if err != nil {
		return nil, err
	}

	return &Database{
		conn:         db,
		lastSeenID:   lastID,
		pollInterval: time.Duration(pollInterval) * time.Second,
	}, nil
}

func decodeByteaToItems(byteData []byte) ([]Item, error) {
  hexData := hex.EncodeToString(byteData) // Convert BYTEA to hex string
  fmt.Println("Hex = %v", hexData)
	if len(hexData) < 4 {
		return nil, fmt.Errorf("invalid data length")
	}

	// Read item count (first 4 hex digits)
  itemCount, err := strconv.ParseInt(hexData[:4], 16, 32)
  fmt.Printf("Count = %d",itemCount)
	if err != nil {
		return nil, err
	}

	items := []Item{}
	index := 4 // Start after length field

  for i := 0; i < int(itemCount); i++ {
		types, _ := strconv.ParseInt(hexData[index:index+2], 16, 8) // 2 hex digits
		index += 2 + 4                                              // Skip padding "0000"

		code, _ := strconv.ParseInt(hexData[index:index+4], 16, 32) // 4-byte integer (8 hex digits)
		index += 4 + 4                                              // Skip padding "0000"

		count, _ := strconv.ParseInt(hexData[index:index+4], 16, 16) // 4 hex digits
		index += 4 + 8                                               // Skip padding "00000000"

		items = append(items, Item{
			types : int(types),
			data:  int(code),
      quantity: int(count),
		})
	}

	return items, nil
}

// getLastInsertedID fetches the latest inserted ID
func getLastInsertedID(db *sql.DB) (int, error) {
	var lastID int
	err := db.QueryRow("SELECT COALESCE(MAX(id), 0) FROM distribution").Scan(&lastID)
	if err != nil {
		return 0, err
  }
  fmt.Println("Last ID: ", lastID)
	return lastID, nil
}

func bulkInsertItems(db *sql.DB, items []Item, dist_id int) error {
	if len(items) == 0 {
		return nil // Nothing to insert
	}

	// Prepare separate slices for each column
	types := make([]int, len(items))
	codes := make([]int, len(items))
	counts := make([]int, len(items))
	ids := make([]int, len(items))

	for i, item := range items {
		types[i] = item.types
		codes[i] = item.data
    counts[i] = item.quantity
    ids[i] = dist_id
	}

	// Use UNNEST to bulk insert
	query := `
		INSERT INTO distribution_items (distribution_id,item_type, item_id, quantity)
  SELECT * FROM UNNEST($1::int[], $2::int[], $3::int[], $4::int[])
	`

	_, err := db.Exec(query, pq.Array(ids) , pq.Array(types), pq.Array(codes), pq.Array(counts))
	return err
}

func (db *Database) PollNewRows() {
	for {
    rows, err := db.conn.Query("SELECT id, data FROM distribution WHERE id > $1 AND bot IS true ORDER BY id ASC", db.lastSeenID)
		if err != nil {
			log.Fatal(err)
		}

		for rows.Next() {
			var id int
      var data []byte
			if err := rows.Scan(&id, &data); err != nil {
				log.Fatal(err)
      }
      items,err := decodeByteaToItems(data)
      if err != nil {
        fmt.Printf("error decoding: %v", err)
        return
      }

      fmt.Println(items)

      werr := bulkInsertItems(db.conn,items,id)
      if werr != nil {
        fmt.Printf("error inserting: %v", err)
        return
      }

			// Update lastSeenID
			if id > db.lastSeenID {
				db.lastSeenID = id
			}
		}
		rows.Close()

		time.Sleep(db.pollInterval) // Wait before next poll
	}
}
func main() {
	// Command-line flags
	dbURL := flag.String("db_url", "", "PostgreSQL database URL (required)")
	interval := flag.Int("interval", 5, "Polling interval in seconds")
	flag.Parse()

	if *dbURL == "" {
		log.Fatal("Database URL is required. Usage: ./myprogram -db_url=<DB_URL>")
	}

	// Connect to DB
	db, err := NewDatabase(*dbURL, *interval)
	if err != nil {
		log.Fatal(err)
	}

	// Fetch and decode data
	db.PollNewRows()
}
