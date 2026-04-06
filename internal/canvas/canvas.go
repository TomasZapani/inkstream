package canvas

//Definicion del struct para los parametros de una linea -> punto de inicio; punto final; color; grueso
type Stroke struct {
	//Punto de inicio:
	StartX float64 `json:"prevX"`
	StartY float64 `json:"prevY"`

	//Puntos final:
	EndX float64 `json:"x"`
	EndY float64 `json:"y"`

	//Color:
	Color string `json:"color"`

	//Grosor:
	Thickness int `json:"width"`
}

//Definicion del struct para el pizarron:
type Canvas struct {
	//Lista de trazos que componen el pizarron:
	Strokes []Stroke `json:"strokes"`
}

//Funcion para agregar trazos al pizarron:
func (c *Canvas) AddStroke(stroke Stroke) {
	//Lógica para añadir un trazon al pizarron
	c.Strokes = append(c.Strokes, stroke)
}

//Funcion para limpiar el pizarron:
func (c *Canvas) Clear() {
	//Lógica para limpiar el pizarron:
	c.Strokes = []Stroke{}
}

//Funcion para obtener los trazos del pizarron:
func (c *Canvas) GetStrokes() []Stroke {
	//Lógica para devolver una lista de trazon:
	return append([]Stroke{}, c.Strokes...)
}
