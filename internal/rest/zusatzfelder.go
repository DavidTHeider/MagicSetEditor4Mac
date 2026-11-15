package rest

import (
	"context"
	"net/http"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Zusatzwahl liest und setzt die Zusatzfelder einer Karte — die Felder, die
// ihr Stil mitbringt und die sich von Hand setzen lassen
// (#2298).
type Zusatzwahl interface {
	Zusatzfelder(ctx context.Context, setID, karteID int64) ([]Zusatzfeld, error)
	ZusatzwertSetzen(ctx context.Context, setID, karteID int64, feld, wert string) error
}

// Zusatzfeld ist ein Zusatzfeld samt dem Wert, den die Karte dafür trägt.
type Zusatzfeld struct {
	Name         string   `json:"name"`
	Art          string   `json:"art"`
	Beschreibung string   `json:"beschreibung"`
	Wahlen       []string `json:"wahlen"`
	Wert         string   `json:"wert"`
	Gespeichert  bool     `json:"gespeichert"`
}

// zusatzfelder gibt die Zusatzfelder einer Karte.
func (d *Dienst) zusatzfelder(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gutesSet := kennung(schreiben, anfrage, "setID")
	karteID, guteKarte := kennung(schreiben, anfrage, "karteID")

	if !gutesSet || !guteKarte {
		return
	}

	felder, fehler := d.naehte.Zusatz.Zusatzfelder(anfrage.Context(), setID, karteID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	for i := range felder {
		felder[i].Wahlen = ohneNil(felder[i].Wahlen)
	}

	ja(schreiben, http.StatusOK, map[string]any{"felder": ohneNil(felder)})
}

// zusatzwertSetzen setzt den Wert eines Zusatzfelds.
func (d *Dienst) zusatzwertSetzen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gutesSet := kennung(schreiben, anfrage, "setID")
	karteID, guteKarte := kennung(schreiben, anfrage, "karteID")

	if !gutesSet || !guteKarte {
		return
	}

	// **`Wert` als Zeiger**: „fehlt" und „leer" sind zwei Aussagen, wie bei
	// MCP — dieselbe unvollständige Eingabe gibt an jeder Fläche dieselbe
	// Antwort (Befund 5 des Nachlaufs über E/F).
	var rumpf struct {
		Wert *string `json:"wert"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	if rumpf.Wert == nil {
		fehlerhaft(schreiben, domain.NewFieldError("wert", domain.SchluesselPflichtfeld,
			"Es fehlt »wert«. Wer das Feld leeren will, gibt ihn als leere Zeichenkette an."))

		return
	}

	if fehler := d.naehte.Zusatz.ZusatzwertSetzen(
		anfrage.Context(), setID, karteID, anfrage.PathValue("feld"), *rumpf.Wert); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{})
}
