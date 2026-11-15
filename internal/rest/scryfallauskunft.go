package rest

import (
	"context"
	"net/http"
)

// Fremdauskuenfte ist, was die lesenden Auskunftsendpunkte brauchen.
type Fremdauskuenfte interface {
	Regelhinweise(ctx context.Context, name string) (Regelauskunft, error)
	Manakosten(ctx context.Context, kosten string) (unbekannt []string, fehler error)
}

// Regelauskunft sind die Hinweise zu einer Karte.
type Regelauskunft struct {
	Kartenname   string
	Hinweise     []Regelhinweis
	Adresse      string
	Befunde      []string
	AlterStunden int
}

// Regelhinweis ist ein einzelner Hinweis.
type Regelhinweis struct {
	Offiziell bool
	Quelle    string
	Datum     string
	Text      string
}

// scryfallRegeln gibt die Regelhinweise zu einer Karte.
//
// # Warum GET und nicht POST
//
// **Weil er nichts ändert.** Der Kartenname steht als `?name=` an der Adresse,
// wie beim Vorschlag darüber — und damit ist der Weg zwischenspeicherbar und
// wiederholbar, was ein POST nicht zusagt.
func (d *Dienst) scryfallRegeln(schreiben http.ResponseWriter, anfrage *http.Request) {
	name := anfrage.URL.Query().Get("name")
	if name == "" {
		nein(schreiben, http.StatusBadRequest, "validierung",
			"der Kartenname fehlt — er gehört als ?name= an die Adresse")

		return
	}

	geholt, fehler := d.naehte.Auskunft.Regelhinweise(anfrage.Context(), name)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{
		"kartenname":   geholt.Kartenname,
		"hinweise":     alsHinweiszeilen(geholt.Hinweise),
		"adresse":      geholt.Adresse,
		"befunde":      alsListe(geholt.Befunde),
		"alterStunden": geholt.AlterStunden,
	})
}

// scryfallManakosten prüft einen Manakosten gegen die Symbologie.
func (d *Dienst) scryfallManakosten(schreiben http.ResponseWriter, anfrage *http.Request) {
	kosten := anfrage.URL.Query().Get("kosten")
	if kosten == "" {
		nein(schreiben, http.StatusBadRequest, "validierung",
			"der Manakosten fehlt — er gehört als ?kosten= an die Adresse")

		return
	}

	unbekannt, fehler := d.naehte.Auskunft.Manakosten(anfrage.Context(), kosten)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **Eine leere Liste heißt: in Ordnung** — und sie bleibt eine leere
	// Liste, kein `null` (§3).
	ja(schreiben, http.StatusOK, map[string]any{"unbekannt": alsListe(unbekannt)})
}

// hinweiszeile ist ein Hinweis, wie dieser Dienst ihn herausgibt.
type hinweiszeile struct {
	Offiziell bool   `json:"offiziell"`
	Quelle    string `json:"quelle"`
	Datum     string `json:"datum"`
	Text      string `json:"text"`
}

// alsHinweiszeilen übersetzt für die Antwort.
func alsHinweiszeilen(hinweise []Regelhinweis) []hinweiszeile {
	zeilen := make([]hinweiszeile, 0, len(hinweise))
	for _, einer := range hinweise {
		zeilen = append(zeilen, hinweiszeile(einer))
	}

	return zeilen
}
