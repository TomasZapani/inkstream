package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/TomasZapani/inkstream/internal/ai"
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

// serveGuess recibe la imagen del canvas en base64, le pide a Claude que adivine
// qué es, y broadcastea la respuesta a todos los clientes conectados.
func serveGuess(h *hub.Hub, aic *ai.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if aic == nil {
		http.Error(w, "IA no configurada: falta ANTHROPIC_API_KEY", http.StatusServiceUnavailable)
		return
	}

	// 8 MB es de sobra para un PNG del canvas a tamaño de pantalla.
	r.Body = http.MaxBytesReader(w, r.Body, 8<<20)

	var req struct {
		Image string `json:"image"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if req.Image == "" {
		http.Error(w, "campo image vacío", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	guess, err := aic.Guess(ctx, req.Image)
	if err != nil {
		log.Println("error llamando a Claude:", err)
		http.Error(w, "error consultando a la IA", http.StatusBadGateway)
		return
	}

	payload, _ := json.Marshal(map[string]string{
		"type": "guess",
		"text": guess,
	})
	h.Broadcast <- payload

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	// Crear el Hub y ponerlo a correr en segundo plano
	h := hub.NewHub()
	go h.Run()

	// Cliente de IA (opcional: si no hay API key, /guess responde 503)
	aic, err := ai.New()
	if err != nil {
		log.Println("aviso:", err, "— el endpoint /guess quedará deshabilitado")
		aic = nil
	}

	// Servir los archivos del frontend
	http.Handle("/", http.FileServer(http.Dir("./web")))

	// Endpoint WebSocket
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(h, w, r)
	})

	// Endpoint de adivinanza
	http.HandleFunc("/guess", func(w http.ResponseWriter, r *http.Request) {
		serveGuess(h, aic, w, r)
	})

	log.Println("Servidor corriendo en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
