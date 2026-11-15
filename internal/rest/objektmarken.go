package rest

// **Marken an jedem Fachobjekt** — UX-12
// (#1684),
// Vermerk 0135.
//
// # Die Wege
//
// `/objektmarken` zählt, `/objektmarken/{art}` nennt je Objekt,
// `/objektmarken/{art}/filter` filtert, `/objektmarken/{art}/{kennung}` liest
// und schreibt eine Marke an einem Objekt. **Die Pflege einer Marke als
// Ganzes** — umbenennen, zusammenführen, überall entfernen — steht unter
// `/marken`, weil sie an keinem Objekt hängt.
//
// **Ein Paket heißt `art/name`** — im Weg mit `%2F` geschrieben; der
// Wegweiser trennt daran nicht.

import (
	"context"
	"net/http"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Objektmarken sind die Marken jeder Art, wie der Dienst sie ruft.
type Objektmarken interface {
	ObjektmarkeSetzen(ctx context.Context, art, kennung, marke string) error
	ObjektmarkeEntfernen(ctx context.Context, art, kennung, marke string) error
	ObjektmarkenAn(ctx context.Context, art, kennung string) ([]string, error)
	MarkenJeObjekt(ctx context.Context, art string) (map[string][]string, error)
	MarkeVergeben(ctx context.Context, art string, kennungen []string, marke string) error
	MarkenFiltern(ctx context.Context, art string, marken []string, verknuepfung string) ([]string, error)
	Markenkatalog(ctx context.Context) ([]domain.Markenposten, error)
	MarkeUmbenennen(ctx context.Context, alt, neu string, zusammenfuehren bool) (int, error)
	MarkeUeberallEntfernen(ctx context.Context, marke string) (int, error)
}

// antworte schreibt die Antwort — oder den Fehler.
func antworte(schreiben http.ResponseWriter, antwort any, fehler error) {
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, antwort)
}

func (d *Dienst) markenkatalogZeigen(schreiben http.ResponseWriter, anfrage *http.Request) {
	katalog, fehler := d.naehte.Objektmarken.Markenkatalog(anfrage.Context())
	antworte(schreiben, map[string]any{"marken": katalog}, fehler)
}

func (d *Dienst) markenJeObjekt(schreiben http.ResponseWriter, anfrage *http.Request) {
	je, fehler := d.naehte.Objektmarken.MarkenJeObjekt(anfrage.Context(), anfrage.PathValue("art"))
	antworte(schreiben, map[string]any{"marken": je}, fehler)
}

func (d *Dienst) markenFiltern(schreiben http.ResponseWriter, anfrage *http.Request) {
	frage := anfrage.URL.Query()

	kennungen, fehler := d.naehte.Objektmarken.MarkenFiltern(anfrage.Context(),
		anfrage.PathValue("art"), frage["marke"], frage.Get("verknuepfung"))
	antworte(schreiben, map[string]any{"kennungen": markenliste(kennungen)}, fehler)
}

func (d *Dienst) objektmarkenZeigen(schreiben http.ResponseWriter, anfrage *http.Request) {
	marken, fehler := d.naehte.Objektmarken.ObjektmarkenAn(anfrage.Context(),
		anfrage.PathValue("art"), anfrage.PathValue("kennung"))
	antworte(schreiben, map[string]any{"marken": markenliste(marken)}, fehler)
}

func (d *Dienst) objektmarkeSetzen(schreiben http.ResponseWriter, anfrage *http.Request) {
	marke, gut := markeImWeg(schreiben, anfrage)
	if !gut {
		return
	}

	fehler := d.naehte.Objektmarken.ObjektmarkeSetzen(anfrage.Context(),
		anfrage.PathValue("art"), anfrage.PathValue("kennung"), marke)
	antworte(schreiben, map[string]any{"marke": domain.Markennormalform(marke), "gesetzt": true}, fehler)
}

func (d *Dienst) objektmarkeEntfernen(schreiben http.ResponseWriter, anfrage *http.Request) {
	marke, gut := markeImWeg(schreiben, anfrage)
	if !gut {
		return
	}

	fehler := d.naehte.Objektmarken.ObjektmarkeEntfernen(anfrage.Context(),
		anfrage.PathValue("art"), anfrage.PathValue("kennung"), marke)
	antworte(schreiben, map[string]any{"marke": domain.Markennormalform(marke), "entfernt": true}, fehler)
}

func (d *Dienst) markeVergeben(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Kennungen []string `json:"kennungen"`
		Marke     string   `json:"marke"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	fehler := d.naehte.Objektmarken.MarkeVergeben(anfrage.Context(), anfrage.PathValue("art"), rumpf.Kennungen, rumpf.Marke)
	antworte(schreiben, map[string]any{"marke": domain.Markennormalform(rumpf.Marke), "vergeben": len(rumpf.Kennungen)}, fehler)
}

func (d *Dienst) markeUmbenennen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Alt             string `json:"alt"`
		Neu             string `json:"neu"`
		Zusammenfuehren bool   `json:"zusammenfuehren"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	zahl, fehler := d.naehte.Objektmarken.MarkeUmbenennen(anfrage.Context(), rumpf.Alt, rumpf.Neu, rumpf.Zusammenfuehren)
	antworte(schreiben, map[string]any{"marke": domain.Markennormalform(rumpf.Neu), "anzahl": zahl}, fehler)
}

func (d *Dienst) markeUeberallEntfernen(schreiben http.ResponseWriter, anfrage *http.Request) {
	marke, gut := markeImWeg(schreiben, anfrage)
	if !gut {
		return
	}

	zahl, fehler := d.naehte.Objektmarken.MarkeUeberallEntfernen(anfrage.Context(), marke)
	antworte(schreiben, map[string]any{"marke": domain.Markennormalform(marke), "anzahl": zahl}, fehler)
}

// markeImWeg liest die Marke aus dem Weg — leer ist keine.
func markeImWeg(schreiben http.ResponseWriter, anfrage *http.Request) (string, bool) {
	marke := anfrage.PathValue("marke")
	if strings.TrimSpace(marke) == "" {
		nein(schreiben, http.StatusBadRequest, "validierung", "die Marke fehlt — sie gehört in den Weg")

		return "", false
	}

	return marke, true
}
