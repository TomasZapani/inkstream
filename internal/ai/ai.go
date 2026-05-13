package ai

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Client envuelve al cliente del SDK de Anthropic para usar visión.
type Client struct {
	c anthropic.Client
}

// New crea un cliente leyendo ANTHROPIC_API_KEY del entorno.
func New() (*Client, error) {
	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		return nil, errors.New("ANTHROPIC_API_KEY no está seteada")
	}
	c := anthropic.NewClient(option.WithAPIKey(key))
	return &Client{c: c}, nil
}

const promptAdivinar = "Estás mirando un dibujo a mano alzada hecho en un pizarrón colaborativo. " +
	"Adiviná en una sola frase corta y divertida qué creés que dibujaron. " +
	"Si el dibujo está vacío o es muy ambiguo, decilo directamente. " +
	"Respondé en español, sin preámbulos."

// Guess manda la imagen PNG (base64 sin el prefijo data:) a Claude y devuelve la adivinanza.
func (a *Client) Guess(ctx context.Context, pngBase64 string) (string, error) {
	// Tolerar el prefijo "data:image/png;base64," si llega así desde el frontend.
	if i := strings.Index(pngBase64, ","); strings.HasPrefix(pngBase64, "data:") && i >= 0 {
		pngBase64 = pngBase64[i+1:]
	}

	resp, err := a.c.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeOpus4_7,
		MaxTokens: 256,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewImageBlockBase64("image/png", pngBase64),
				anthropic.NewTextBlock(promptAdivinar),
			),
		},
	})
	if err != nil {
		return "", err
	}

	for _, block := range resp.Content {
		if t, ok := block.AsAny().(anthropic.TextBlock); ok {
			return strings.TrimSpace(t.Text), nil
		}
	}
	return "", fmt.Errorf("la respuesta del modelo no incluyó texto")
}
