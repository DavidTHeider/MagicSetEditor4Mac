package domain

// **Eine eigene Übersetzung** — LOC-02
// (#1688).
//
// # Was sie ist
//
// Eine Sprache, die jemand selbst mitbringt: eine Kennung nach BCP 47, eine
// **Basis** unter den Fenstersprachen und die Sätze, die er übersetzt hat —
// jeder mit dem deutschen Satz, gegen den er übersetzt wurde. **Was fehlt,
// kommt aus der Basis, dann aus dem Deutschen**; eine leere Beschriftung gibt
// es damit nicht (AK-03).
//
// # Was sie nicht ist
//
// **Code.** Ein Sprachpaket ist Text, und die Prüfung in
// [Katalogvorlage.PruefeUebersetzung] weist ab, was wie Code aussieht
// (AK-02). Es gibt keinen Ausdruck, keine Formel, keinen Verweis, der beim
// Zeichnen ausgewertet würde.
//
// **new** — MSE2 kennt eigene Übersetzungen nur als `.mse-locale`-Paket
// **seiner** Oberfläche (`src/data/locale.cpp`, Fassung f9d9356 vom
// 2021-07-06); diese Oberfläche ist eine andere, und ihre Kataloge sind es
// auch.

import (
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Sprachpaketformat ist die Kennung einer ausgegebenen Datei samt Fassung.
const Sprachpaketformat = "mse4mac/sprachpaket/1"

// Sprachstand sagt, wie weit eine Übersetzung ist — LOC-01.AK-03.
type Sprachstand string

// Die vier Stände. **Nicht dasselbe**: ein Entwurf ist unfertig, maschinell
// heißt ohne menschliches Lesen, geprüft heißt gelesen, abgenommen heißt
// benutzt.
const (
	StandEntwurf    Sprachstand = "entwurf"
	StandMaschinell Sprachstand = "maschinell"
	StandGeprueft   Sprachstand = "geprueft"
	StandAbgenommen Sprachstand = "abgenommen"
)

// Gueltig sagt, ob ein Stand einer der vier ist.
func (s Sprachstand) Gueltig() bool {
	return slices.Contains(
		[]Sprachstand{StandEntwurf, StandMaschinell, StandGeprueft, StandAbgenommen}, s)
}

// Uebersetzungstext ist ein übersetzter Satz und seine Quelle.
//
// **Die Quelle ist der deutsche Satz zur Zeit der Übersetzung.** Ändert er
// sich mit einer neuen Fassung dieses Programms, ist die Übersetzung
// **veraltet** — markiert, nicht verworfen (AK-03).
//
// **Mit Marken**, weil dieselbe Gestalt im Bestand und in der ausgegebenen
// Datei steht — zwei Schreibweisen für einen Satz liefen auseinander.
type Uebersetzungstext struct {
	Text   string `json:"text"`
	Quelle string `json:"quelle"`
}

// Sprachpaket ist eine eigene Übersetzung.
type Sprachpaket struct {
	Kennung  string
	Name     string
	Basis    string
	Stand    Sprachstand
	Lizenz   string
	Herkunft string
	Texte    map[string]Uebersetzungstext

	Angelegt  time.Time
	Geaendert time.Time
}

// sprachkennung ist BCP 47 in der kanonischen Schreibweise: Sprache klein,
// Schrift mit großem Anfang, Gegend groß oder drei Ziffern, dazu ein
// privater Teil. **Nur diese Schreibweise**, damit `pt-br` und `pt-BR` nicht
// zwei Pakete werden.
var sprachkennung = regexp.MustCompile(
	`^[a-z]{2,3}(-[A-Z][a-z]{3})?(-([A-Z]{2}|[0-9]{3}))?(-x(-[a-z0-9]{1,8})+)?$`)

// hoechstensKennung ist die längste Kennung — eine Obergrenze für den
// Dateinamen der Ausgabe, nicht für BCP 47.
const hoechstensKennung = 35

// PruefeSprachkennung urteilt über die Kennung einer eigenen Übersetzung.
//
// **Eine eingebaute Sprache ist keine eigene** — auch nicht unter ihrer
// BCP-47-Kennung (`zh-Hans`). Wer eine verbessern will, nimmt einen privaten
// Teil: `ja-x-meine` (höchstens acht Zeichen je Teil, wie BCP 47 es verlangt).
func PruefeSprachkennung(kennung string) error {
	if len(kennung) > hoechstensKennung || !sprachkennung.MatchString(kennung) {
		return NewFieldError("kennung", SchluesselSprachkennungUngueltig,
			"%q ist keine Sprachkennung nach BCP 47 — etwa pt-BR oder zh-Hant.", kennung).
			MitWerten(map[string]string{"kennung": kennung})
	}

	if slices.Contains(fenstersprachen, kennung) || kennung == "zh-Hans" {
		return NewFieldError("kennung", SchluesselSprachkennungEingebaut,
			"%q ist eine eingebaute Sprache — nimm eine eigene Kennung wie %s-x-meine.",
			kennung, kennung).
			MitWerten(map[string]string{"kennung": kennung})
	}

	return nil
}

// Pruefe urteilt über die Angaben eines Pakets — nicht über seine Sätze.
func (p Sprachpaket) Pruefe() error {
	if fehler := PruefeSprachkennung(p.Kennung); fehler != nil {
		return fehler
	}

	if name := strings.TrimSpace(p.Name); name == "" ||
		utf8.RuneCountInString(name) > 60 || !lesbar(name) {
		return NewFieldError("name", SchluesselValidierung,
			"Ein Sprachpaket braucht einen Namen mit höchstens 60 Zeichen in einer Zeile.")
	}

	if !slices.Contains(fenstersprachen, p.Basis) {
		return NewFieldError("basis", SchluesselValidierung,
			"%q ist keine Sprache des Fensters.", p.Basis)
	}

	if !p.Stand.Gueltig() {
		return NewFieldError("stand", SchluesselValidierung,
			"%q ist kein Stand einer Übersetzung.", p.Stand)
	}

	// **Eine Reihenfolge und keine Map**: zwei falsche Angaben geben
	// zweimal denselben Befund (§10, #1289).
	for _, angabe := range [][2]string{{"lizenz", p.Lizenz}, {"herkunft", p.Herkunft}} {
		if utf8.RuneCountInString(angabe[1]) > 200 || !lesbar(angabe[1]) {
			return NewFieldError(angabe[0], SchluesselValidierung,
				"Diese Angabe hat höchstens 200 Zeichen in einer Zeile.")
		}
	}

	return nil
}

// lesbar sagt, ob ein Wert ohne Steuerzeichen auskommt.
func lesbar(wert string) bool {
	return !strings.ContainsFunc(wert, unicode.IsControl)
}
