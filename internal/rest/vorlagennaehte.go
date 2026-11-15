// **Die Naht zu den Exportvorlagen** —
// #1291.
//
// **Eine eigene Datei**, seit `naehte.go` mit der Restparität über die
// Zeilengrenze aus §2 lief. Der Schnitt folgt der Frage: dort stehen die
// Nähte, die es seit dem ersten Tag gibt, hier die eine, die zuletzt dazukam.

package rest

import "context"

// Vorlagenausgabe führt die Exportvorlagen aus (#1291).
type Vorlagenausgabe interface {
	Vorlagen(ctx context.Context, setID int64) ([]Vorlagenzeile, []string, error)
	Ausfuehren(ctx context.Context, setID int64, paket, ziel string,
		wahl map[string]string) (befunde []string, fehler error)
	// VorlagenoptionSetzen behält den Wert einer Vorlagenoption.
	//
	// **Mit Vorsilbe**: `OptionSetzen` gibt es hier schon — für die
	// **Stiloptionen** eines Sets. Die Doubles der Tests erfüllen jede Naht
	// dieses Pakets, und zwei gleiche Namen mit anderer Signatur ließen sich
	// nicht erfüllen.
	VorlagenoptionSetzen(ctx context.Context, vorlage, feld, wert string) error
}

// Vorlagenzeile ist eine Exportvorlage.
type Vorlagenzeile struct {
	Paket    string           `json:"paket"`
	Name     string           `json:"name"`
	Dateityp string           `json:"dateityp"`
	Optionen []Vorlagenoption `json:"optionen"`
}

// Vorlagenoption ist ein einstellbares Feld einer Exportvorlage.
type Vorlagenoption struct {
	Name string `json:"name"`
	Art  string `json:"art"`
	Wert string `json:"wert"`
}
