package main

import (
	"log"
	"net/http"

	"github.com/TomasZapani/inkstream/internal/client"
	"github.com/TomasZapani/inkstream/internal/hub"
	"github.com/gorilla/websocket"
)

// Upgrader convierte una conexión HTTP normal a WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Handler que atiende cada nueva conexión WebSocket
func serveWs(h *hub.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("error al hacer upgrade:", err)
		return
	}

	// Crear el cliente y registrarlo en el Hub
	c := client.NewClient(h, conn)
	h.Register <- c

	// Arrancar las dos goroutines del cliente
	go c.WritePump()
	go c.ReadPump()
}

func main() {
	// Crear el Hub y ponerlo a correr en segundo plano
	h := hub.NewHub()
	go h.Run()

	// Servir los archivos del frontend
	http.Handle("/", http.FileServer(http.Dir("./web")))

	// Endpoint WebSocket
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(h, w, r)
	})

	log.Println("Servidor corriendo en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
