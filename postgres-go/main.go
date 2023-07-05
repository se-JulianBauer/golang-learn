package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	run(ctx)
}

// connect to database using a single connection
func run(ctx context.Context) {
	ret, success := os.LookupEnv("DB_URL")
	var connStr = "postgres://postgres:password@timescaledb:5432/template1"
	if success {
		connStr = ret
	}

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	// allow users to run queries
	reader := bufio.NewReader(os.Stdin)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Print("Enter query > ")

			query, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to read query: %v\n", err)
				continue
			}
			// remove newline
			query = strings.TrimSuffix(query, "\n")

			// execute the query
			rows, err := conn.Query(ctx, query)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to execute the query: %v\n", err)
				continue
			}
			defer rows.Close()

			// print the query result
			for rows.Next() {
				// print the results, whatever they are
				result, err := rows.Values()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Unable to get values: %v\n", err)
					continue
				}
				// pretty print the result
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
