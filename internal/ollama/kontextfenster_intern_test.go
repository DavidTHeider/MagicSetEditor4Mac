package ollama

import (
	"encoding/json"
	"testing"
)

// **Was fehlt oder nicht passt, heißt: nicht angegeben** — und nicht null
// Tokens.
func TestEinUnlesbaresKontextfensterIstKeines(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name      string
		parameter string
		angaben   map[string]json.RawMessage
	}{
		{"ohne alles", "", nil},
		{"ohne Architektur", "", map[string]json.RawMessage{"llama.context_length": json.RawMessage(`8192`)}},
		{"Architektur ohne Länge", "", map[string]json.RawMessage{"general.architecture": json.RawMessage(`"llama"`)}},
		{"Länge als Text", "", map[string]json.RawMessage{
			"general.architecture": json.RawMessage(`"llama"`), "llama.context_length": json.RawMessage(`"viel"`),
		}},
		{"num_ctx ohne Zahl", "num_ctx viel\ntemperature 0.7", nil},
		{"num_ctx null", "num_ctx 0", nil},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if fenster, beleg := kontextAus(fall.parameter, fall.angaben); fenster != 0 || beleg != "" {
				t.Errorf("%d, %q", fenster, beleg)
			}
		})
	}
}

// **`num_ctx` ohne trainierte Länge** gilt trotzdem — der Beleg sagt nur
// weniger.
func TestNumCtxGiltAuchAllein(t *testing.T) {
	t.Parallel()

	if fenster, beleg := kontextAus("num_ctx 4096", nil); fenster != 4096 || beleg == "" {
		t.Errorf("%d, %q", fenster, beleg)
	}
}
