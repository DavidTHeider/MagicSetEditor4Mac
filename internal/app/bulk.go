package app

import (
	"fmt"
	"os"
	"path/filepath"
)

// bulkart ist die Sammeldatei, auf die dieses Programm zurückfällt.
//
// # Warum `oracle_cards` und nicht `all_cards`
//
// **24 MB gegen 392 MB**, gemessen am 2026-08-24 am Verzeichnis von Scryfall.
// `oracle_cards` führt je Orakel-Kennung **eine** Karte — genau das, wonach
// ein Feldabgleich sucht. `all_cards` führt jeden Druck jeder Sprache; das ist
// das Sechzehnfache an Bytes für eine Antwort, die dieselbe ist.
const bulkart = "oracle_cards"

// Bulkverzeichnis gibt den Ordner für die Sammeldateien.
//
// **Neben dem Antwortspeicher und den Bildern**, aus demselben Grund: eine
// Datei von 24 MB hat eine andere Frist als eine Kartenantwort, und ein
// Aufräumen des einen soll das andere nicht treffen.
func Bulkverzeichnis() (string, error) {
	ort, fehler := Scryfallverzeichnis()
	if fehler != nil {
		return "", fehler
	}

	ordner := filepath.Join(ort, "sammeldaten")
	if err := os.MkdirAll(ordner, 0o700); err != nil {
		return "", fmt.Errorf("den Ordner für die Sammeldaten anlegen: %w", err)
	}

	return ordner, nil
}

// Bulkpfad gibt den Ort der abgelegten Sammeldatei.
func Bulkpfad() (string, error) {
	ordner, fehler := Bulkverzeichnis()
	if fehler != nil {
		return "", fehler
	}

	return filepath.Join(ordner, bulkart+".jsonl.gz"), nil
}

// Sammeldatenstand sagt, was auf der Platte liegt und was angeboten wird.
type Sammeldatenstand struct {
	// Art ist die Sammeldatei, auf die dieses Programm zurückfällt.
	Art string `json:"art"`

	// Pfad ist ihr Ort auf der Platte.
	Pfad string `json:"pfad"`

	// Vorhanden sagt, ob sie schon geholt wurde.
	Vorhanden bool `json:"vorhanden"`

	// Bytes ist ihre Größe auf der Platte — 0, wenn sie fehlt.
	Bytes int64 `json:"bytes"`

	// Geholt ist, wann die Datei auf der Platte geschrieben wurde, in RFC 3339
	// und UTC — leer, wenn sie fehlt. **Die Abrufzeit und nicht der Stand der
	// Quelle** (F04, #2246): den nennt [Sammeldatenstand.Stand], und nur, wenn
	// Scryfall antwortet. Ohne Netz ist dies die einzige Auskunft, wie alt die
	// Datei ist.
	Geholt string `json:"geholt"`

	// Angeboten ist die Größe, die Scryfall nennt — 0, wenn nicht gefragt wurde.
	Angeboten int64 `json:"angeboten"`

	// Stand ist der Zeitpunkt, den Scryfall für die angebotene Datei nennt.
	Stand string `json:"stand"`
}
