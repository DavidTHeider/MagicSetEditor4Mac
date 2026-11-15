// **Die Rechnungen des Symboleditors, benannt statt aufgezählt**
// (#1391).
//
// # Warum eine Namenstabelle und nicht zwölf Nähte
//
// `internal/mcp` und `internal/rest` sind **Blätter** (§2): sie kennen
// `internal/domain` und sonst nichts aus diesem Projekt. Jede Rechnung
// einzeln durchzureichen hieße, ihre zwölf Argumentgestalten dort ein zweites
// Mal zu beschreiben — und die zweite Beschreibung liefe an dem Tag
// auseinander, an dem eine Rechnung ein Feld bekommt.
//
// Hier steht **eine** Naht: ein Name und ein JSON-Rumpf hinein, ein Ergebnis
// heraus. Was ein Aufrufer schicken darf, sagt das Schema an seiner Fläche;
// was daraus wird, steht hier.
//
// # Was das kostet, und es gehört ausgesprochen
//
// **Der Übersetzer prüft die Argumente nicht mehr.** Ein Tippfehler im
// Feldnamen fällt erst beim Lauf auf, und ein unbekannter Name gibt einen
// Fehler statt eines Baufehlers. Dagegen steht [Symbolrechnungsnamen] und ein
// Wächter, der die Namen jeder Fläche dagegen hält.

package app

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// symbolargumente ist die Vereinigung dessen, was die Rechnungen brauchen.
//
// **Ein Verbund und nicht neun.** Die Felder überschneiden sich stark —
// `teile` in acht, `stelle` in vier —, und neun Verbunde nebeneinander wären
// neunmal dieselbe Frage, welcher gerade gilt.
type symbolargumente struct {
	Teile   []SymbolteilDTO `json:"teile"`
	Teil    int             `json:"teil"`
	Punkt   int             `json:"punkt"`
	Stelle  int             `json:"stelle"`
	Stellen []int           `json:"stellen"`
	Von     int             `json:"von"`
	Nach    int             `json:"nach"`
	Weise   string          `json:"weise"`
	Davor   bool            `json:"davor"`
	X       float64         `json:"x"`
	Y       float64         `json:"y"`
	Sperre  string          `json:"sperre"`

	// Die Grundform.
	Art   string `json:"art"`
	Ecken int    `json:"ecken"`
	Eng   bool   `json:"eng"`

	// Die Symmetrie.
	Spiegelart string  `json:"spiegelart"`
	Kopien     int     `json:"kopien"`
	MitteX     float64 `json:"mitteX"`
	MitteY     float64 `json:"mitteY"`
	GriffX     float64 `json:"griffX"`
	GriffY     float64 `json:"griffY"`

	// Der Zuschnitt — **in Bildpunkten und nicht in Anteilen**: `AuswahlDTO`
	// rechnet in ganzen Zahlen, weil ein Ausschnitt aus einem Bild keine
	// halben Punkte hat.
	QuelleBreite int    `json:"quelleBreite"`
	QuelleHoehe  int    `json:"quelleHoehe"`
	ZielBreite   int    `json:"zielBreite"`
	ZielHoehe    int    `json:"zielHoehe"`
	AuswahlX     int    `json:"auswahlX"`
	AuswahlY     int    `json:"auswahlY"`
	Breite       int    `json:"breite"`
	Hoehe        int    `json:"hoehe"`
	Frei         bool   `json:"frei"`
	Bevorzugt    string `json:"bevorzugt"`
}

// Symbolrechnungsnamen gibt die Namen, die [SymbolRechnen] kennt.
//
// **Damit jede Fläche dagegen zu halten ist.** Ein Werkzeug mit einem Namen,
// den dieser Rechner nicht kennt, wäre eine Fläche ohne Wirkung — und der
// Fehler käme erst, wenn ihn jemand ruft.
func Symbolrechnungsnamen() []string {
	return []string{
		"lesen.symbol.grundform",
		"lesen.symbol.griff",
		"lesen.symbol.punkt",
		"lesen.symbol.gruppe",
		"lesen.symbol.reihenfolge",
		"lesen.symbol.symmetrie",
		"lesen.symbol.symmetriekopien",
		"lesen.symbol.zuschnitt",
	}
}

// SymbolRechnen führt die benannte Rechnung aus.
//
// **Der Fehler kommt aus der Antwort, nicht aus dem Rechnen.** Die Rechnungen
// geben eine `SymbolAntwort` mit einem `*Fehler` darin — die Gestalt, die das
// Fenster braucht. Kopflos ist ein `error` die Gestalt, die zählt: die
// Kommandozeile setzt daraus ihren Rückgabewert, MCP seinen `isError`, und
// der Dienst seinen Status.
func SymbolRechnen(was string, roh json.RawMessage) (any, error) {
	var wahl symbolargumente

	if len(roh) > 0 {
		if err := json.Unmarshal(roh, &wahl); err != nil {
			return nil, domain.MitGrund(domain.ErrValidation, err,
				"die Argumente sind kein JSON")
		}
	}

	antwort, fehler := symbolrechnung(was, wahl)
	if fehler != nil {
		return nil, fehler
	}

	return antwort, nil
}

// symbolrechnung wählt die Rechnung und übersetzt ihren Befund.
func symbolrechnung(was string, wahl symbolargumente) (any, error) {
	switch was {
	case "lesen.symbol.grundform":
		return alsErgebnis(SymbolGrundform(SymbolgrundformDTO{
			Art: wahl.Art, Ecken: wahl.Ecken, Eng: wahl.Eng,
		}))

	case "lesen.symbol.griff":
		return alsErgebnis(SymbolGriffZiehen(wahl.Teile, SymbolgriffDTO{
			Teil: wahl.Teil, Punkt: wahl.Punkt, Davor: wahl.Davor,
			X: wahl.X, Y: wahl.Y,
		}))

	case "lesen.symbol.punkt":
		return punktrechnung(wahl)

	case "lesen.symbol.sperre":
		if !slices.Contains(sperren, wahl.Sperre) {
			return nil, unbekannteWeise(wahl.Sperre, sperren...)
		}

		return alsErgebnis(SymbolSperreSetzen(wahl.Teile, SymbolsperreDTO{
			Teil: wahl.Teil, Punkt: wahl.Punkt, Sperre: wahl.Sperre,
		}))

	case "lesen.symbol.gruppe":
		return gruppenrechnung(wahl)

	case "lesen.symbol.reihenfolge":
		return alsErgebnis(SymbolTeilVerschieben(wahl.Teile, wahl.Von, wahl.Nach))

	case "lesen.symbol.symmetrie":
		return symmetrierechnung(wahl)

	case "lesen.symbol.symmetriekopien":
		kopien := SymbolSymmetriekopien(wahl.Teile, wahl.Stelle)
		if kopien.Fehler != nil {
			return nil, kopien.Fehler.AlsFehler()
		}

		return kopien.Kopien, nil

	case "lesen.symbol.zuschnitt":
		return SymbolZuschnittgrenze(
			wahl.QuelleBreite, wahl.QuelleHoehe, wahl.ZielBreite, wahl.ZielHoehe,
			AuswahlDTO{
				X: wahl.AuswahlX, Y: wahl.AuswahlY,
				Breite: wahl.Breite, Hoehe: wahl.Hoehe,
			},
			wahl.Frei, wahl.Bevorzugt), nil
	}

	return nil, fmt.Errorf("%w: die Rechnung %q gibt es nicht", domain.ErrValidation, was)
}

// punktrechnung fügt einen Punkt ein oder löscht einen.
func punktrechnung(wahl symbolargumente) (any, error) {
	weise := SymbolpunktweiseDTO{Teil: wahl.Teil, Punkt: wahl.Punkt}

	switch wahl.Weise {
	case "einfuegen":
		return alsErgebnis(SymbolPunktEinfuegen(wahl.Teile, weise))

	case "loeschen":
		return alsErgebnis(SymbolPunktLoeschen(wahl.Teile, weise))
	}

	return nil, unbekannteWeise(wahl.Weise, "einfuegen", "loeschen")
}

// gruppenrechnung fasst zusammen oder löst auf.
func gruppenrechnung(wahl symbolargumente) (any, error) {
	switch wahl.Weise {
	case "gruppieren":
		return alsErgebnis(SymbolTeileGruppieren(wahl.Teile, wahl.Stellen))

	case "aufloesen":
		return alsErgebnis(SymbolGruppeAufloesen(wahl.Teile, wahl.Stellen))
	}

	return nil, unbekannteWeise(wahl.Weise, "gruppieren", "aufloesen")
}

// symmetrierechnung legt an, setzt oder löst auf.
func symmetrierechnung(wahl symbolargumente) (any, error) {
	switch wahl.Weise {
	case "anlegen":
		return alsErgebnis(SymbolSymmetrieAnlegen(wahl.Teile, wahl.Stelle))

	case "aufloesen":
		return alsErgebnis(SymbolSymmetrieAufloesen(wahl.Teile, wahl.Stelle))

	case "setzen":
		return alsErgebnis(SymbolSymmetrieSetzen(wahl.Teile, SymmetrieweiseDTO{
			Teil: wahl.Stelle, Spiegelart: wahl.Spiegelart, Kopien: wahl.Kopien,
			MitteX: wahl.MitteX, MitteY: wahl.MitteY,
			GriffX: wahl.GriffX, GriffY: wahl.GriffY,
		}))
	}

	return nil, unbekannteWeise(wahl.Weise, "anlegen", "setzen", "aufloesen")
}

// alsErgebnis macht aus einer Symbolantwort ein Ergebnis oder einen Fehler.
//
// **Die Teile roh und nicht in einem Umschlag.** Jede Fläche benennt das
// Ergebnis nach ihrer eigenen Übereinkunft — MCP mit großem Anfangsbuchstaben,
// der Dienst mit kleinem —, und an der Kommandozeile ist die rohe Liste das,
// was sich in den nächsten Aufruf schieben lässt. Ein Umschlag hier zwänge
// jeden Aufrufer, ihn erst wieder abzunehmen.
func alsErgebnis(antwort SymbolAntwort) (any, error) {
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort.Teile, nil
}

// unbekannteWeise sagt, welche Weisen es gibt.
//
// **Sie zählt sie auf.** Eine Meldung „unbekannte Weise" schickte jemanden
// suchen; die Aufzählung ist die Antwort auf die nächste Frage (§8).
func unbekannteWeise(hatte string, erlaubt ...string) error {
	return fmt.Errorf("%w: die Weise %q gibt es nicht — erlaubt sind %v",
		domain.ErrValidation, hatte, erlaubt)
}

// SymbolteileUnlesbar meldet eine Teilliste, die kein JSON ist.
//
// **Hier und nicht beim Aufrufer**: die Meldung ist deutsch (§9) und gehört
// zu dieser Gestalt; drei Flächen schrieben sie sonst dreimal — und beim
// dritten Mal anders.
func SymbolteileUnlesbar(ursache error) error {
	return domain.MitGrund(domain.ErrValidation, ursache, "die Teile sind kein JSON")
}
