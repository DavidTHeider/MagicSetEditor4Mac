package domain

// **Welches Modell der Vorrang bei Ollama fragt** — #2054, aus AI-02
// (#1743).
//
// # Warum es eine Wahl gibt
//
// **Ohne sie nimmt der Vorrang das erste Modell, das in fester Form
// antworten kann** — in der Reihenfolge, in der Ollama sie nennt, und
// `/api/tags` ordnet nach Änderungszeit: das zuletzt gezogene zuerst
// (gemessen an 0.35.1). AI-02 hat gezeigt, warum das zählt: an denselben 32
// Fragen passten bei `gemma3:1b` 9 Vorschläge, bei `qwen3:0.6b` keiner. Wer
// beide hat, bekam das jüngere.
//
// # Was die Wahl nicht tut
//
// **Ein gewähltes Modell, das fehlt, still durch das erste ersetzen.** Wer
// gewählt hat, hat entschieden; ein anderes Modell zu fragen, ohne es zu
// sagen, nähme ihm die Entscheidung. Der Fehler nennt das Modell.

import (
	"fmt"
	"strings"
	"unicode"
)

// KIVorrangmodell ist das Ollama-Modell, das der Vorrang fragt — leer heißt:
// das erste, das in fester Form antworten kann.
const KIVorrangmodell = "ki.vorrangmodell"

// vorrangmodellgrenze ist, wie lang ein Modellname sein darf. Ollama nennt
// Namen wie `hf.co/organisation/modell:Q4_K_M`; zweihundert Zeichen lassen
// Luft und halten Unsinn draußen.
const vorrangmodellgrenze = 200

// PruefeKIVorrangmodell nimmt einen Modellnamen oder nichts — ohne Leerraum
// und ohne Steuerzeichen, denn beides kommt in keinem Namen vor, den Ollama
// führt.
func PruefeKIVorrangmodell(wert string) error {
	if len(wert) > vorrangmodellgrenze {
		return fmt.Errorf("%w: ein Modellname hat höchstens %d Zeichen",
			ErrValidation, vorrangmodellgrenze)
	}

	for _, zeichen := range wert {
		if unicode.IsSpace(zeichen) || unicode.IsControl(zeichen) {
			return fmt.Errorf("%w: %q ist kein Modellname — Leerraum und "+
				"Steuerzeichen kommen in keinem vor", ErrValidation, wert)
		}
	}

	return nil
}

// KIVorrangmodellAus gibt das Modell, das der Vorrang fragt.
//
// **Ohne Wahl** das erste mit fester Form, oder keines — was dann zu sagen
// ist, weiß der Aufrufer, der den Anbieter kennt. **Mit Wahl** das gewählte;
// eine Wahl ohne Marke meint `:latest`, wie bei Ollama selbst.
func KIVorrangmodellAus(modelle []KIModell, gewaehlt string) (string, error) {
	gewaehlt = strings.TrimSpace(gewaehlt)

	if gewaehlt == "" {
		for _, modell := range modelle {
			if modell.Kann(KIStrukturierteGenerierung) {
				return modell.Alias, nil
			}
		}

		return "", nil
	}

	gesucht := gewaehlt
	if !strings.Contains(gesucht, ":") {
		gesucht += ":latest"
	}

	for _, modell := range modelle {
		if modell.Alias != gewaehlt && modell.Alias != gesucht {
			continue
		}

		if !modell.Kann(KIStrukturierteGenerierung) {
			return "", fmt.Errorf("%w: das gewählte Vorrangmodell %q kann nicht "+
				"in fester Form antworten", ErrValidation, gewaehlt)
		}

		return modell.Alias, nil
	}

	return "", fmt.Errorf("%w: das gewählte Vorrangmodell %q nennt die "+
		"Laufzeit nicht — in den Einstellungen unter KI ein anderes wählen "+
		"oder die Wahl leeren", ErrNotFound, gewaehlt)
}
