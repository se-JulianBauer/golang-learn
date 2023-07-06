package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
)

type PageData struct {
	Query    string
	Response string
}

type DatabaseResponse struct {
	err  bool
	text string
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	run(ctx)
}

func renderTemplate(w http.ResponseWriter, data PageData) {
	tmpl := template.Must(template.ParseFiles("template.html"))
	err := tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type Handler struct {
	queries chan<- string
	resps   <-chan DatabaseResponse
}

func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		renderTemplate(w, PageData{})
	} else if r.Method == "POST" {
		query := r.FormValue("query")
		handler.queries <- query
		resp := <-handler.resps
		data := PageData{
			Query:    query,
			Response: resp.text,
		}
		renderTemplate(w, data)
	}

}

func httpConnection(ctx context.Context, queries chan<- string, resps <-chan DatabaseResponse) {
	var err error
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{Addr: ":8080", Handler: Handler{queries, resps}}

	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func envOrDefault(key, def string) string {
	val, success := os.LookupEnv(key)
	if !success {
		return def
	}
	return val
}

func databaseConnection(ctx context.Context, queries <-chan string, resps chan<- DatabaseResponse) {
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
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}
		break
	}
	defer func() {
		fmt.Println("Closing database")
		conn.Close(ctx)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case query := <-queries:
			text, err := getQueryResult(ctx, conn, query)
			if err != nil {
				resps <- DatabaseResponse{err: true, text: err.Error()}
			} else {
				resps <- DatabaseResponse{err: false, text: text}
			}
		}

	}
}

func getQueryResult(ctx context.Context, conn *pgx.Conn, query string) (string, error) {
	// execute the query
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	ret := ""
	// print the query result
	for rows.Next() {
		result, err := rows.Values()
		if err != nil {
			return "", err
		}
		for _, v := range result {
			ret += fmt.Sprint(v, "\t")
		}
		ret += "\n"
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	return ret, nil
}

func run(ctx context.Context) {
	queries := make(chan string, 1)
	resps := make(chan DatabaseResponse, 1)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		httpConnection(ctx, queries, resps)
	}()
	go func() {
		defer wg.Done()
		databaseConnection(ctx, queries, resps)
	}()
	fmt.Println("Server started on port 8080")
	wg.Wait()
}
