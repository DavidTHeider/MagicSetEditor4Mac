package rest

// **Die Kartensuche über die Rückschleife** — Werkstück 1 von Phase 3.6
// (#1167).

import (
	"context"
	"net/http"
)

// Fremdkartensuche ist, was der Dienst von der Suche braucht.
type Fremdkartensuche interface {
	Kartensuche(ctx context.Context, frage string) (Suchergebnis, error)
}

// Suchergebnis ist, was die Suche zurückgibt.
type Suchergebnis struct {
	Treffer      []Suchtreffersicht `json:"treffer"`
	Weitere      bool               `json:"weitere"`
	AlterStunden int                `json:"alterStunden"`
	Befunde      []string           `json:"befunde"`
}

// Suchtreffersicht ist eine gefundene Karte, wie diese Fläche sie sieht.
type Suchtreffersicht struct {
	Kennung    string `json:"kennung"`
	Name       string `json:"name"`
	Setcode    string `json:"setcode"`
	Setname    string `json:"setname"`
	Nummer     string `json:"nummer"`
	Seltenheit string `json:"seltenheit"`
	Manakosten string `json:"manakosten"`
	Typzeile   string `json:"typzeile"`
	Bild       string `json:"bild"`
}

// scryfallSuche sucht Karten bei Scryfall.
//
// # Warum GET und die Frage in der Abfrage
//
// **Weil nichts geändert wird** — dieselbe Überlegung wie bei den
// Regelhinweisen. Damit ist der Weg zwischenspeicherbar und wiederholbar, was
// ein POST nicht zusagt.
func (d *Dienst) scryfallSuche(schreiben http.ResponseWriter, anfrage *http.Request) {
	frage := anfrage.URL.Query().Get("frage")
	if frage == "" {
		nein(schreiben, http.StatusBadRequest, "validierung",
			"die Suchfrage fehlt — sie gehört als ?frage= an die Adresse")

		return
	}

	ergebnis, fehler := d.naehte.Suche.Kartensuche(anfrage.Context(), frage)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{
		"treffer":      ohneNilTreffer(ergebnis.Treffer),
		"weitere":      ergebnis.Weitere,
		"alterStunden": ergebnis.AlterStunden,
		"befunde":      alsListe(ergebnis.Befunde),
	})
}

// ohneNilTreffer macht aus einer fehlenden Liste eine leere.
func ohneNilTreffer(treffer []Suchtreffersicht) []Suchtreffersicht {
	if treffer == nil {
		return []Suchtreffersicht{}
	}

	return treffer
}
