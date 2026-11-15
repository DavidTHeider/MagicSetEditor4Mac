package domain

// **Welche Wege an welchem Schalter hängen** — OPS-10
// (#1908).
//
// # Warum eine erzeugte Tafel
//
// **Weil die Zuordnung schon einmal dasteht**: `docs/FAEHIGKEITEN.md` nennt
// je Fähigkeit ihre Wege an allen vier Flächen, und `docs/SCHALTER.md` je
// Fähigkeit ihren Schalter. Eine dritte Liste von Hand liefe auseinander;
// deshalb schreibt `internal/werkzeuge/schalterwege_test.go` die Tafel aus
// beiden und wird rot, sobald sie nicht mehr passt.
//
// # Wann ein Weg abweist
//
// **Erst, wenn jede Fähigkeit, die ihn nennt, abgeschaltet ist.** Ein Weg,
// den auch eine Fähigkeit des Kerns braucht, steht gar nicht in der Tafel —
// sonst schaltete „Spielen aus“ das Öffnen eines Sets ab, nur weil ein Deck
// denselben Weg nimmt.

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ErrAbgeschaltet ist die Abweisung eines Wegs, dessen Bereich oder Funktion
// abgeschaltet ist — an allen vier Flächen dieselbe.
var ErrAbgeschaltet = errors.New("dieser Bereich ist abgeschaltet")

// SchluesselBereichAbgeschaltet ist der Schlüssel dieser Abweisung — mit dem
// Schalter als Wert `schalter`.
const SchluesselBereichAbgeschaltet Schluessel = "bereichAbgeschaltet"

// Abschaltung nennt den Schalter, an dem ein Weg abgewiesen wurde.
type Abschaltung struct {
	Schalter string
}

// Error gibt den Satz, der ohne Katalog gelesen wird.
func (a *Abschaltung) Error() string {
	return fmt.Sprintf("%s: %q — er lässt sich in den Einstellungen unter "+
		"„Bereiche und Funktionen“ wieder einschalten", ErrAbgeschaltet, a.Schalter)
}

// Unwrap ordnet sie [ErrAbgeschaltet] zu.
func (a *Abschaltung) Unwrap() error { return ErrAbgeschaltet }

// Flaeche ist eine der vier Flächen, an denen ein Weg ankommt.
type Flaeche string

// Die vier Flächen, mit den Spaltennamen aus `docs/FAEHIGKEITEN.md`.
const (
	FlaecheFenster Flaeche = "fenster"
	FlaecheMCP     Flaeche = "mcp"
	FlaecheCLI     Flaeche = "cli"
	FlaecheREST    Flaeche = "rest"
)

// Wegetafel ordnet je Fläche einem Weg die Schalter seiner Fähigkeiten zu:
// eine Liste je Fähigkeit, die ihn nennt.
type Wegetafel map[Flaeche]map[string][][]string

// Pruefe sagt, ob ein Weg bei diesem Stand abgewiesen wird.
//
// **Eine Fähigkeit ist aus, sobald einer ihrer Schalter aus ist** — der
// Marktplatz hängt am Handeln und am Markt, und jeder der beiden schaltet
// ihn ab. **Der Weg ist aus, wenn jede seiner Fähigkeiten es ist.** Genannt
// wird der erste abgeschaltete Schalter der ersten Fähigkeit, damit dieselbe
// Lage überall denselben Namen trägt.
func (t Wegetafel) Pruefe(flaeche Flaeche, weg string, stand Schalterstand) error {
	if stand.AllesAn() {
		return nil
	}

	faehigkeiten := t[flaeche][weg]
	if len(faehigkeiten) == 0 {
		return nil
	}

	erster := ""

	for _, schalter := range faehigkeiten {
		aus := ersterAus(schalter, stand)
		if aus == "" {
			return nil
		}

		if erster == "" {
			erster = aus
		}
	}

	return &Abschaltung{Schalter: erster}
}

// ersterAus gibt den ersten abgeschalteten Schalter einer Fähigkeit.
func ersterAus(schalter []string, stand Schalterstand) string {
	for _, kennung := range schalter {
		if stand.Aus(kennung) {
			return kennung
		}
	}

	return ""
}

//go:embed schalterwege.json
var eingebetteteWegetafel []byte

// EingebauteWegetafel gibt die Tafel, die mit dieser Fassung ausgeliefert
// wird.
func EingebauteWegetafel() Wegetafel { return wegetafelAus(eingebetteteWegetafel) }

// wegetafelAus liest eine Tafel.
//
// **Eine unlesbare Tafel hält an**, statt leer weiterzumachen: eine leere
// Tafel schaltete stillschweigend jeden Schalter wirkungslos, und das sähe
// niemand, bis er einen bräuchte.
func wegetafelAus(daten []byte) Wegetafel {
	var tafel Wegetafel
	if fehler := json.Unmarshal(daten, &tafel); fehler != nil {
		panic(fmt.Sprintf("die eingebaute Wegetafel ist unlesbar: %v", fehler))
	}

	return tafel
}

// RESTWeg gibt ein Wegmuster ohne die Namen seiner Platzhalter — so steht es
// in der Tafel, und so fragt der Dienst nach.
//
// **Weil die Namen auseinanderlaufen dürfen**: `docs/FAEHIGKEITEN.md` schreibt
// `/sets/{id}`, `net/http` kennt `/sets/{setID}`, und beide sind derselbe Weg
// (`internal/werkzeuge/restspalte_test.go` vergleicht genauso).
func RESTWeg(muster string) string {
	teile := strings.Split(muster, "/")

	for stelle, teil := range teile {
		if strings.HasPrefix(teil, "{") {
			teile[stelle] = "{}"
		}
	}

	return strings.Join(teile, "/")
}
