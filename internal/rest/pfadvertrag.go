package rest

// **Der Pfadvertrag des Dienstes** — OPS-02.AK-04
// (#1719):
// „Ein offener Server darf über Pfadangaben keine beliebigen fremden
// Hostdateien offenlegen oder überschreiben."
//
// # Vor dem Weg und nicht in ihm
//
// **Achtzehn Felder in elf Dateien tragen einen Pfad**, und ein neuer
// Endpunkt bringt das neunzehnte. Eine Prüfung je Weg wäre achtzehnmal zu
// vergessen; hier sieht sie jeder Rumpf und jede Adresse, bevor ein Weg sie
// bekommt, an den Namen aus `domain.IstPfadfeld`. Dass jedes Pfadfeld des
// Dienstes einen dieser Namen trägt, hält
// `internal/werkzeuge/pfadfelder_test.go`.
//
// # In jeder Schreibweise, und unter den Namen eines Wegs
//
// **Ein Name gilt so, wie der Dekodierer ihn zuordnet** (#2073): `"Ziel"`
// landet in `json:"ziel"`, und `domain.PfadeIn` vergleicht deshalb ohne
// Rücksicht auf die Schreibweise. **Und fünf Wege nehmen einen Pfad unter
// einem Namen, der nach keinem klingt** — `quelle` und `ort` der Paketwege,
// `quelle` des Indexaufbaus.
// Sie stehen in [wegpfadfelder]; für alle Wege ginge es nicht, denn `quelle`
// eines Preises ist ein Anbieter. Dass jedes solche Feld entschieden ist,
// hält `internal/werkzeuge/pfadstrukturen_test.go`.
//
// # Hinter der Marke
//
// **Erst wer ein Token hat, erfährt, dass ein Pfad nicht geht** — sonst
// fragte ein Fremder ohne Marke ab, welche Orte es gibt.

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Pfadvertrag urteilt über einen Pfad, den der Dienst bekommt.
type Pfadvertrag func(ctx context.Context, pfad string) error

// MitPfadvertrag setzt den Vertrag — ohne ihn urteilt niemand, und das ist
// der Stand der Läufe, die einen Dienst aus Doppeln bauen.
// `cmd/mse-rest` setzt ihn; `internal/werkzeuge/pfadfelder_test.go` hält,
// dass es dabei bleibt.
func (d *Dienst) MitPfadvertrag(vertrag Pfadvertrag) *Dienst {
	d.pfadvertrag = vertrag

	return d
}

// wegpfadfelder sind Pfadfelder, deren Name nach keinem Pfad klingt — je
// Weg, weil derselbe Name anderswo keinen trägt.
var wegpfadfelder = map[string][]string{
	"PUT /pakete/eingebunden":  {"quelle"},
	"PUT /pakete/ersetzt":      {"quelle"},
	"GET /pakete/plan":         {"quelle"},
	"POST /pakete/uebernommen": {"ort"},

	// **Die Sammeldatei eines Indexaufbaus** geht bis `os.Open` — gefunden,
	// als der Wächter jedes `quelle` entschieden haben wollte.
	"POST /index/aufbau": {"quelle"},
}

// mitPfadvertrag prüft jeden Pfad in Adresse und Rumpf.
func (d *Dienst) mitPfadvertrag(weg *http.ServeMux, weiter http.Handler) http.Handler {
	return http.HandlerFunc(func(schreiben http.ResponseWriter, anfrage *http.Request) {
		_, muster := weg.Handler(anfrage)

		if fehler := d.pfadeVerwehrt(anfrage, muster); fehler != nil {
			fehlerhaft(schreiben, fehler)

			return
		}

		weiter.ServeHTTP(schreiben, anfrage)
	})
}

// pfadeVerwehrt urteilt über die Pfade einer Anfrage — **auch über die einer
// Wiederholung**, die nicht mehr durch die Wachen davor kommt
// (`auftraege.go`): der Vertrag kann sich seit dem ersten Lauf geändert
// haben.
//
// **Die Adresse liest der Weg unter dem genauen Namen** (`Query().Get`), und
// nur unter ihm wird sie deshalb geprüft.
func (d *Dienst) pfadeVerwehrt(anfrage *http.Request, muster string) error {
	if d.pfadvertrag == nil {
		return nil
	}

	weitere := wegpfadfelder[muster]

	for name, werte := range anfrage.URL.Query() {
		if !domain.IstPfadfeld(name) && !slices.Contains(weitere, name) {
			continue
		}

		for _, wert := range werte {
			if fehler := d.pfadvertrag(anfrage.Context(), wert); fehler != nil {
				return fehler
			}
		}
	}

	return d.rumpfPruefen(anfrage, weitere)
}

// rumpfPruefen liest den Rumpf, prüft seine Pfadfelder und legt ihn zurück.
//
// **Gelesen wird höchstens, was `gelesen` annimmt, und eins mehr**: ein
// größerer Rumpf ist dort ohnehin eine Abweisung, und hier wird er
// unverändert weitergereicht — der Rest des Stroms hängt hinten an.
func (d *Dienst) rumpfPruefen(anfrage *http.Request, weitere []string) error {
	// **Ein Rumpf ist immer da** — `mitMarke` davor hat ihn in eine Grenze
	// gefasst, auch einen leeren. Ein leerer hat keine Pfade.
	roh, fehler := io.ReadAll(io.LimitReader(anfrage.Body, rumpfgrenze+1))
	anfrage.Body = struct {
		io.Reader
		io.Closer
	}{io.MultiReader(bytes.NewReader(roh), anfrage.Body), anfrage.Body}

	if fehler != nil {
		return domain.MitGrund(domain.ErrValidation, fehler, "den Rumpf lesen")
	}

	for _, pfad := range domain.PfadeIn(roh, weitere...) {
		if fehler := d.pfadvertrag(anfrage.Context(), pfad); fehler != nil {
			return fehler
		}
	}

	return nil
}
