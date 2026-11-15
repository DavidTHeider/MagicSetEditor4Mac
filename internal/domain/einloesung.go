package domain

// **Eine Einladung genügt** — NET-01b
// (#1954),
// entschieden in
// [0156](../../docs/entscheidungen/0156-eine-einladung-genuegt.md).
//
// # Was hier über die Leitung geht
//
// **Zwei Nachrichten über eine Paarungssitzung**: wer eine Einladung
// angenommen hat, schickt ihre Nonce an den Einladenden — die [Einloesung] —,
// und der antwortet mit einem [Einloeseergebnis]. Wer spricht, hat der
// Handschlag davor geprüft; die Kennung steht deshalb **nicht** in der
// Nachricht, und was drüben aufgenommen wird, ist die geprüfte.
//
// # Warum die Paarung und kein neuer Zweck
//
// **Weil die Paarung das Kennenlernen ist** ([Peerkontakt.Darf]): jeder, der
// nicht gesperrt ist, darf sie öffnen. Ein eigener Zweck hätte eine neue
// Zeile im Handschlag gebraucht, und eine ältere Gegenstelle wiese ihn mit
// „kein Zweck dieses Programms" ab, bevor sie zuhört. So weist sie die
// Paarung ab — und das tut sie heute auch.

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Einloesefassung ist die Fassung von [Einloesung] und [Einloeseantwort].
//
// **Eine eigene Zahl neben der des Handschlags**: die Einlösung kann sich
// ändern, ohne dass sich der Handschlag ändert, und umgekehrt.
const Einloesefassung = 1

// einloesegrenze ist die größte Nachricht, die gelesen wird.
//
// **Geprüft, bevor gelesen wird**: die Leitung lässt größere zu, und eine
// Einlösung hat nichts, was 4 KiB bräuchte.
const einloesegrenze = 4 << 10

// Einloesung ist, was der Annehmende dem Einladenden schickt.
type Einloesung struct {
	Fassung int    `json:"fassung"`
	Nonce   string `json:"nonce"`

	// Adresse ist, wo der Annehmende erreichbar ist — leer, wenn er es
	// nicht weiß. **Kein Nachweis**, wie die Adresse einer Einladung.
	Adresse string `json:"adresse,omitempty"`

	// Name ist, wie der Annehmende sich nennt — **kein Nachweis**, und
	// gewählt von ihm: ein Name, der hinausgeht, wird nicht vorausgefüllt.
	Name string `json:"name,omitempty"`
}

// Einloeseergebnis ist die Antwort des Einladenden.
type Einloeseergebnis string

// Die Ergebnisse, die ein Einladender geben kann.
const (
	EinloesungAufgenommen Einloeseergebnis = "aufgenommen"
	EinloesungFremd       Einloeseergebnis = "fremd"
	EinloesungAbgelaufen  Einloeseergebnis = "abgelaufen"
	EinloesungVerbraucht  Einloeseergebnis = "verbraucht"
	EinloesungGesperrt    Einloeseergebnis = "gesperrt"
	EinloesungGedrosselt  Einloeseergebnis = "gedrosselt"
	EinloesungZuNeu       Einloeseergebnis = "fassung"
	EinloesungGescheitert Einloeseergebnis = "gescheitert"

	// EinloesungUnbeantwortet geht nie über die Leitung — es ist, was der
	// Annehmende sieht, wenn keine Antwort kam.
	EinloesungUnbeantwortet Einloeseergebnis = "unbeantwortet"
)

// Einloeseantwort ist die zweite Nachricht.
type Einloeseantwort struct {
	Fassung  int              `json:"fassung"`
	Ergebnis Einloeseergebnis `json:"ergebnis"`
}

var (
	// ErrEinladungFremd heißt: diese Nonce hat hier niemand ausgestellt.
	ErrEinladungFremd = fmt.Errorf(
		"%w: diese Einladung stammt nicht von hier", ErrNotFound)

	// ErrEinladungAbgelaufen heißt: sie galt, aber nicht mehr.
	ErrEinladungAbgelaufen = fmt.Errorf(
		"%w: die Einladung ist abgelaufen — bitte um eine neue", ErrValidation)

	// ErrEinladungVerbraucht heißt: sie ist schon eingelöst.
	ErrEinladungVerbraucht = fmt.Errorf(
		"%w: diese Einladung ist schon eingelöst — für eine neue Verbindung "+
			"bitte um eine neue", ErrConflict)

	// ErrEinloesungGesperrt heißt: wer einlöst, ist beim Einladenden
	// gesperrt.
	ErrEinloesungGesperrt = fmt.Errorf(
		"%w: dieses Gegenüber ist beim Einladenden gesperrt", ErrConflict)

	// ErrEinloesenGedrosselt ist die Ratengrenze, benannt (AK-02).
	ErrEinloesenGedrosselt = fmt.Errorf(
		"%w: zu viele Einlöseversuche — bitte später noch einmal",
		ErrGedrosselt)

	// ErrEinloesefassung heißt: die andere Seite spricht eine neuere
	// Fassung der Einlösung.
	ErrEinloesefassung = fmt.Errorf(
		"%w: die Gegenstelle spricht eine andere Fassung der Einlösung — "+
			"tauscht zwei Einladungen aus", ErrValidation)

	// ErrEinloesungGescheitert heißt: drüben ging etwas schief, das nicht
	// an der Einladung lag.
	ErrEinloesungGescheitert = errors.New(
		"der Einladende konnte die Einlösung nicht vermerken")

	// ErrEinloesungUnbeantwortet heißt: der Einladende war nicht zu
	// erreichen oder hat nicht geantwortet — **der Rückfall ist der Weg mit
	// zwei Einladungen**, und der Satz sagt ihn.
	ErrEinloesungUnbeantwortet = errors.New(
		"dein Gegenüber hat die Einlösung nicht beantwortet — schick ihm " +
			"deine eigene Einladung")

	// ErrEinladungNichtEinloesbar heißt: die Einladung sagt nicht, dass sie
	// eingelöst werden kann — sie stammt aus einer älteren Fassung oder nennt
	// keine Adresse.
	ErrEinladungNichtEinloesbar = errors.New(
		"diese Einladung lässt sich nicht einlösen — schick deinem " +
			"Gegenüber deine eigene")
)

// ergebnisse ordnet jedem benannten Fehler sein Ergebnis zu — **eine**
// Tafel für beide Richtungen.
var ergebnisse = []struct {
	fehler   error
	ergebnis Einloeseergebnis
}{
	{ErrEinladungFremd, EinloesungFremd},
	{ErrEinladungAbgelaufen, EinloesungAbgelaufen},
	{ErrEinladungVerbraucht, EinloesungVerbraucht},
	{ErrEinloesungGesperrt, EinloesungGesperrt},
	{ErrEinloesenGedrosselt, EinloesungGedrosselt},
	{ErrEinloesefassung, EinloesungZuNeu},
	{ErrEinloesungGescheitert, EinloesungGescheitert},
	{ErrEinloesungUnbeantwortet, EinloesungUnbeantwortet},
}

// Kodiere schreibt die Einlösung in der Fassung dieses Programms.
func (e Einloesung) Kodiere() ([]byte, error) {
	e.Fassung = Einloesefassung
	if fehler := e.pruefe(); fehler != nil {
		return nil, fehler
	}

	return json.Marshal(e) //nolint:wrapcheck // zwei Zeichenketten und eine Zahl
}

// ZerlegeEinloesung liest eine Einlösung — und prüft, was sie sagt.
func ZerlegeEinloesung(roh []byte) (Einloesung, error) {
	var einloesung Einloesung
	if fehler := lesen(roh, &einloesung, "eine Einlösung"); fehler != nil {
		return Einloesung{}, fehler
	}

	if fehler := einloesung.pruefe(); fehler != nil {
		return Einloesung{}, fehler
	}

	return einloesung, nil
}

func (e Einloesung) pruefe() error {
	if fehler := pruefeEinloesefassung(e.Fassung); fehler != nil {
		return fehler
	}

	if fehler := pruefeNonce(e.Nonce); fehler != nil {
		return fehler
	}

	return PruefeAdresse(e.Adresse)
}

// AntwortAuf macht aus dem Befund des Einladenden die Antwort.
//
// **Was keinen Namen hat, wird „gescheitert"** — der Grund bleibt im
// Protokoll dieses Rechners und geht nicht über die Leitung.
func AntwortAuf(fehler error) Einloeseantwort {
	antwort := Einloeseantwort{Fassung: Einloesefassung, Ergebnis: EinloesungAufgenommen}
	if fehler == nil {
		return antwort
	}

	antwort.Ergebnis = EinloesungGescheitert

	for _, eintrag := range ergebnisse {
		if errors.Is(fehler, eintrag.fehler) {
			antwort.Ergebnis = eintrag.ergebnis

			break
		}
	}

	return antwort
}

// Kodiere schreibt die Antwort — **in der Fassung dieses Programms**, wie
// [Einloesung.Kodiere]: eine von Hand gebaute Antwort ohne Fassung wiese die
// Gegenseite sonst ab (D05, #2236).
//
// **Ohne Fehler**: eine Zahl und ein Wort lassen sich immer schreiben, und
// ein Zweig für den Fall, dass nicht, wäre einer, den kein Lauf betritt.
func (a Einloeseantwort) Kodiere() []byte {
	a.Fassung = Einloesefassung
	roh, _ := json.Marshal(a) //nolint:errcheck,errchkjson // siehe darüber

	return roh
}

// ZerlegeEinloeseantwort liest eine Antwort.
func ZerlegeEinloeseantwort(roh []byte) (Einloeseantwort, error) {
	var antwort Einloeseantwort
	if fehler := lesen(roh, &antwort, "eine Antwort auf die Einlösung"); fehler != nil {
		return Einloeseantwort{}, fehler
	}

	if fehler := pruefeEinloesefassung(antwort.Fassung); fehler != nil {
		return Einloeseantwort{}, fehler
	}

	if antwort.Ergebnis != EinloesungAufgenommen && antwort.fehlerklasse() == nil {
		return Einloeseantwort{}, fmt.Errorf(
			"%w: %q ist kein Ergebnis einer Einlösung", ErrValidation,
			antwort.Ergebnis)
	}

	return antwort, nil
}

// Fehler ist das Ergebnis als benannter Fehler — `nil` heißt aufgenommen.
func (a Einloeseantwort) Fehler() error {
	return a.fehlerklasse()
}

func (a Einloeseantwort) fehlerklasse() error {
	for _, eintrag := range ergebnisse {
		if eintrag.ergebnis == a.Ergebnis {
			return eintrag.fehler
		}
	}

	return nil
}

// lesen hält die Größe, bevor JSON gelesen wird.
func lesen(roh []byte, ziel any, was string) error {
	if len(roh) > einloesegrenze {
		return fmt.Errorf("%w: %s wiegt %d Byte, erlaubt sind %d",
			ErrValidation, was, len(roh), einloesegrenze)
	}

	if fehler := json.Unmarshal(roh, ziel); fehler != nil {
		return MitGrund(ErrValidation, fehler, "%s lesen", was)
	}

	return nil
}

// pruefeEinloesefassung trennt „krumm" von „zu neu".
func pruefeEinloesefassung(fassung int) error {
	switch {
	case fassung < 1:
		return fmt.Errorf("%w: die Einlösung nennt keine Fassung", ErrValidation)
	case fassung > Einloesefassung:
		return fmt.Errorf("%w (%d gegen %d)", ErrEinloesefassung, fassung,
			Einloesefassung)
	}

	return nil
}

// BrauchbarerName gibt einen Namen aus dem Handschlag zurück, wenn ein Mensch
// ihn lesen kann — sonst nichts. Dieselbe Regel wie für den Namen einer
// Einladung.
func BrauchbarerName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" || pruefeEinladungsname(name) != nil {
		return ""
	}

	return name
}
