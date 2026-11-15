// **Was dieses Programm durchsetzt, über den Dienst** — Werkstück G4 von
// Phase 5.5
// (#1461).
//
// **Rein lesend.** Sie rechnet über ein Regelprofil und eine gewünschte
// Strenge; im Bestand ändert sich nichts.
//
// **Und sie ist die vierte Fläche derselben Sperre** — das Issue verlangt
// sie: „Turniersperre lässt sich weder über UI noch REST/CLI/MCP umgehen."

package rest

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

// Regelfreigabe ist, was der eine Endpunkt dieser Datei braucht.
type Regelfreigabe interface {
	Regelfreigabe(ctx context.Context, spiel, strenge string, sitze int,
		modus string, abgewaehlt []string) (
		Regelfreigabeangaben, error)
}

// Regelbereichsangaben sind die Aussage über einen Regelbereich.
type Regelbereichsangaben struct {
	Bereich string   `json:"bereich"`
	Grad    string   `json:"grad"`
	Quelle  string   `json:"quelle"`
	Grenzen []string `json:"grenzen"`
}

// Regelfreigabeangaben sind die Antwort auf „darf es so losgehen?".
type Regelfreigabeangaben struct {
	Erlaubt   bool   `json:"erlaubt"`
	Regelwerk string `json:"regelwerk"`
	Fassung   int    `json:"fassung"`
	Wirksam   string `json:"wirksam"`
	Hoechste  string `json:"hoechste"`

	Matrix      []Regelbereichsangaben `json:"matrix"`
	Luecken     []string               `json:"luecken"`
	Hindernisse []string               `json:"hindernisse"`

	// Modus ist der Bedienmodus, über den gerechnet wurde — PLAY-02.AK-01
	// (#1710).
	//
	// **Erweiternd** im Sinne von
	// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md):
	// wer ihn nicht kennt, übergeht ihn.
	Modus string `json:"modus"`

	// Baum ordnet die Bereiche nach Ablauf und Mechanik (AK-02).
	Baum []Regelgruppenangaben `json:"baum"`

	// Widersprueche sind die Abwahlen, an denen noch etwas hängt — sie
	// **erklären** und verhindern nichts.
	Widersprueche []Regelwiderspruchsangaben `json:"widersprueche"`
}

// Regelgruppenangaben sind ein Ast des Regelbaums.
type Regelgruppenangaben struct {
	Kennung  string   `json:"kennung"`
	Bereiche []string `json:"bereiche"`
}

// Regelwiderspruchsangaben sind eine Abwahl, an der noch etwas hängt.
type Regelwiderspruchsangaben struct {
	Bereich string `json:"bereich"`
	Haengt  string `json:"haengt"`
	Quelle  string `json:"quelle"`
}

// regelfreigabeSchauen sagt, was durchgesetzt wird und ob es so losgehen darf.
//
// **Die Sitze kommen als Abfrageparameter** und nicht im Pfad: sie sind eine
// Angabe zur Frage und kein Gegenstand, den man adressiert.
func (d *Dienst) regelfreigabeSchauen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	spiel := anfrage.URL.Query().Get("spiel")
	strenge := anfrage.URL.Query().Get("strenge")

	sitze, fehler := strconv.Atoi(anfrage.URL.Query().Get("sitze"))
	if fehler != nil {
		nein(schreiben, http.StatusBadRequest, "validierung",
			"»sitze« fehlt oder ist keine Zahl.")

		return
	}

	// **Modus und Abwahl sind freiwillig** (PLAY-02, #1710): wer sie
	// weglässt, bekommt den Anfängermodus und keine Abwahl. Ein Weg, der
	// ohne sie eine andere Antwort gäbe, wäre eine brechende Änderung
	// ([0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md)).
	//
	// **Kommagetrennt und nicht mehrfach**: `?ohne=zonen,combat` liest sich
	// in einer Adresszeile, und ein leeres Stück fällt weg — „zonen," ist
	// ein Bereich und nicht zwei.
	angaben, klage := d.naehte.Regelfreigabe.Regelfreigabe(
		anfrage.Context(), spiel, strenge, sitze,
		anfrage.URL.Query().Get("modus"),
		geteilteliste(anfrage.URL.Query().Get("ohne")))
	if klage != nil {
		fehlerhaft(schreiben, klage)

		return
	}

	ja(schreiben, http.StatusOK, angaben)
}

// geteilteliste zerlegt eine kommagetrennte Angabe.
//
// **Leere Stücke fallen weg**: „zonen," ist ein Bereich und nicht zwei, und
// ein leerer wiese die Prüfung mit einer Meldung über etwas ab, das niemand
// getippt hat (§8).
func geteilteliste(roh string) []string {
	if strings.TrimSpace(roh) == "" {
		return nil
	}

	stuecke := []string{}

	for _, stueck := range strings.Split(roh, ",") {
		if geglaettet := strings.TrimSpace(stueck); geglaettet != "" {
			stuecke = append(stuecke, geglaettet)
		}
	}

	return stuecke
}
