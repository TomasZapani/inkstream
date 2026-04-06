## Inkstream

  Pizarrón colaborativo en tiempo real. Varios usuarios se
  conectan y dibujan en el mismo canvas — cada trazo aparece
  instantáneamente en la pantalla de todos.

## Stack

  - Backend: Go + gorilla/websocket
  - Frontend: HTML5 Canvas + Vanilla JS
  - Estado: En memoria (sin base de datos)

## Conceptos aplicados en el backend

### WebSockets

  Conexión bidireccional persistente entre cliente y servidor. 
  A diferencia de HTTP (pedido → respuesta), WebSocket mantiene
  la conexión abierta para que el servidor pueda enviar mensajes
  en cualquier momento sin que el cliente lo pida.

### Patrón fan-out

  Cuando un usuario dibuja algo, el servidor recibe el mensaje y
  lo distribuye a todos los clientes conectados. Un mensaje
  entra, N mensajes salen.

### Goroutines y channels

  Cada usuario conectado corre dos goroutines. ReadPump lee
  mensajes del WebSocket y los manda al Hub. WritePump espera
  mensajes del Hub y los escribe al WebSocket. Las goroutines se
  comunican por channels en lugar de memoria compartida,
  evitando condiciones de carrera sin locks explícitos.

### Patrón Hub

  Una única goroutine Hub es la dueña de todo el estado
  compartido — clientes conectados y trazos del canvas. Cada
  mutación fluye por su event loop a través de channels, un
  evento a la vez, sin mutexes en el camino crítico.

### Sincronización al conectarse

  Cuando un usuario nuevo se conecta, el Hub le manda
  inmediatamente la lista completa de trazos existentes (mensaje
  sync), para que vea el estado actual del pizarrón sin que
  nadie tenga que redibujar.

### Expulsión de clientes lentos

  Si el buffer de salida de un cliente se llena, el Hub lo
  expulsa en lugar de bloquear el loop de broadcast. El cliente
  puede reconectarse y recibir un sync actualizado.

### Protocolo de mensajes
  <img width="813" height="395" alt="imagen" src="https://github.com/user-attachments/assets/bb262abd-d0cd-4b99-aeb0-5b35e2c1bd4b" />
## Diagrama de Secuencia:
```mermaid
sequenceDiagram
      participant A as Cliente A (ReadPump/WritePump)
      participant H as Hub (event loop)
      participant C as Canvas (estado)
      participant B as Cliente B (ReadPump/WritePump)

      B->>H: register <- clientB
      H->>C: GetStrokes()
      C-->>H: []Stroke
      H->>B: send <- {type: sync, strokes: [...]}

      A->>H: broadcast <- {type: draw, x, y, color, width}
      H->>C: AddStroke(stroke)
      H->>A: send <- {type: draw, ...}
      H->>B: send <- {type: draw, ...}

      A->>H: broadcast <- {type: clear}
      H->>C: Clear()
      H->>A: send <- {type: clear}
      H->>B: send <- {type: clear}

      B->>H: unregister <- clientB
      H->>H: delete(clients, clientB)
      H->>B: close(send)
```
## Muestra:
<img width="1914" height="1015" alt="imagen" src="https://github.com/user-attachments/assets/7b6c34ef-6bf7-42cf-be22-7a2d28537eac" />

## Correr localmente: 
* go run ./cmd/server/main.go

