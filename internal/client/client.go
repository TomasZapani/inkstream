package client

import (
	"github.com/TomasZapani/inkstream/internal/hub"
	"github.com/gorilla/websocket"
)

// Struct que representa un cliente conectado al hub
// - hub: puntero al hub al que pertenece
// - conn: conexión websocket del cliente
// - send: canal para enviar mensajes al cliente
type Client struct {
	hub  *hub.Hub
	conn *websocket.Conn
	send chan []byte
}

// Fn para crear un nuevo cliente
func NewClient(hub *hub.Hub, conn *websocket.Conn) *Client {
	// Crear un nuevo cliente con el hub y la conexión websocket
	// El canal send se inicializa con un buffer de 256 mensajes
	return &Client{
		conn: conn,
		hub:  hub,
		send: make(chan []byte, 256),
	}
}

// Fn que leer lo que el usuario envía por websocket y lo pasa al hub
func (c *Client) WritePump() {
	// Loop infinito que espera mensajes del canal send
	// Cuando recibe un mensaje, lo escribe al cliente por websocket
	// Si el canal se cierra, envía un mensaje de cierre y retorna
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

// Fn que espéra msjs del hub y los escribe al usuario por websocket
func (c *Client) ReadPump() {
	//Leer el msj que llego por websocket:
	//Si hay error,  avisarle al hub y salir del loop
	//Si no hay error, enviar el msj al hub
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			c.hub.Unregister <- c
			return
		}
		c.hub.Broadcast <- message
	}
}

// Fn que devuelve el canal para enviar mensajes al cliente
func (c *Client) SendMessage() chan []byte {
	return c.send
}
