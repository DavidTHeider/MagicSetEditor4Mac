package mcp

// Objekthandlung ist die Handlungsauskunft zu einem Objekt — P3
// (#2371):
// welche Befehlsarten der Sitz daran gerade stellen kann, und welche davon
// ein Klick (`primaer`) und ein Doppelklick (`zweit`) am Fenster auslösen.
// **Erweiternd** im Sinne von
// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md).
type Objekthandlung struct {
	Objekt     string   `json:"objekt"`
	Handlungen []string `json:"handlungen"`
	Primaer    string   `json:"primaer"`
	Zweit      string   `json:"zweit"`
}
