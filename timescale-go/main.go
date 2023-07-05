package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	run(ctx)
}

func envOrDefault(key, def string) string {
	val, success := os.LookupEnv(key)
	if !success {
		return def
	}
	return val
}

// connect to database using a single connection
func run(ctx context.Context) {
	// lookup environment variables: POSTGRES_USER, POSTGRES_PASSWORD, DB_HOST, DB_PORT, DB_NAME
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		envOrDefault("POSTGRES_USER", "user"),
		envOrDefault("POSTGRES_PASSWORD", "password"),
		envOrDefault("DB_HOST", "localhost"),
		envOrDefault("DB_PORT", "5432"),
		envOrDefault("DB_NAME", "template1"),
	)

	var conn *pgx.Conn
	for {
		var err error
		conn, err = pgx.Connect(ctx, connStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Waiting for database: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}
		defer conn.Close(ctx)
		break
	}

	// allow users to run queries
	scanner := bufio.NewScanner(os.Stdin)
	scan := make(chan string, 1)
	go func() {
		for {
			scanner.Scan()
			scan <- scanner.Text()
		}
	}()

	for {
		fmt.Print("Enter query > ")
		select {
		case <-ctx.Done():
			return
		case query := <-scan:
			// execute the query
			rows, err := conn.Query(ctx, query)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to execute the query: %v\n", err)
				continue
			}
			defer rows.Close()

			// print the query result
			for rows.Next() {
				result, err := rows.Values()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Unable to get values: %v\n", err)
					continue
				}
				for _, v := range result {
					fmt.Print(v, "\t")
				}
				fmt.Println()
			}
			if rows.Err() != nil {
				fmt.Fprintf(os.Stderr, "An error occurred while iterating over the query result: %v\n", rows.Err())
			}
		}
	}
}
