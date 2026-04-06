package hub

import (
	"encoding/json"

	"github.com/TomasZapani/inkstream/internal/canvas"
)

type Hub struct {
	//Lista de usuarios conectados
	clients map[Client]bool

	//Canal para cuando alguien se conecta
	Register chan Client

	//Canal para cuando alguien se desconecta
	Unregister chan Client

	//Canal para cuando llega un mensaje a repartir
	Broadcast chan []byte

	//Manejo del  canvas
	canvas *canvas.Canvas
}

type Client interface {
	SendMessage() chan []byte
}

// Fn para crear y dev un hub listo para usar
func NewHub() *Hub {
	//Esta fn retorna un puntero a un hub recien creado, retornará:
	// - clients: mapa vacío de clientes
	// - register: canal para registrar clientes
	// - unregister: canal para desconectar clientes
	// - broadcast: canal para enviar mensajes a todos los clientes
	// - canvas: canvas vacío
	return &Hub{
		clients:    make(map[Client]bool),
		Register:   make(chan Client),
		Unregister: make(chan Client),
		Broadcast:  make(chan []byte),
		canvas:     &canvas.Canvas{},
	}
}

// Fn para: resgitrar, desconectar y difundir mensajes
func (h *Hub) Run() {
	for {
		select {
		//Alguien se conectó
		case client := <-h.Register:
			//agrega el cliente al mapa
			h.clients[client] = true
			//envía el estado actual del canvas al cliente en formato JSON
			payload, _ := json.Marshal(map[string]interface{}{
				"type":    "sync",
				"strokes": h.canvas.GetStrokes(),
			})
			client.SendMessage() <- payload

		//Alguien se desconectó
		case client := <-h.Unregister:
			//Verif que el  cliente está en el mapa(por si llega duplicado) y si está, sacarlo del maoa y cerrar el canal de envío
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.SendMessage())
			}

		//Alguien envió un mensaje
		case message := <-h.Broadcast:
			//Leer el tipo de mensaje
			var msg struct {
				Type string `json:"type"`
			}
			json.Unmarshal(message, &msg)
			//Actualizar el canvas segun el tipo
			switch msg.Type {
			case "draw":
				var stroke canvas.Stroke
				json.Unmarshal(message, &stroke)
				h.canvas.AddStroke(stroke)
			case "clear":
				h.canvas.Clear()
			}
			//Repartir el mensaje a todos los clientes
			for cliente := range h.clients {
				select {
				case cliente.SendMessage() <- message:
				default:
					close(cliente.SendMessage())
					delete(h.clients, cliente)
				}
			}

		}
	}
}
