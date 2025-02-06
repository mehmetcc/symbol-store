package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/mehmetcc/price-store/internal/admin"
	"github.com/mehmetcc/price-store/internal/config"
	"github.com/mehmetcc/price-store/internal/db"
	"github.com/mehmetcc/price-store/internal/routes"
	"github.com/mehmetcc/price-store/internal/websocket"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	log.Printf("config: %+v", cfg)

	db.Connect(cfg)

	ws, err := websocket.NewClient(cfg)
	if err != nil {
		log.Fatalf("error connecting to websocket: %v", err)
	}
	defer ws.Close()
	go ws.Listen()

	client := admin.NewAdminClient(*cfg, *ws)

	mux := http.NewServeMux()
	routes.SetupRoutes(mux, client)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting HTTP server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down")

	if err := server.Close(); err != nil {
		log.Fatalf("Server Close: %v", err)
	}
}
