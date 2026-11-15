package domain

// **Was neben dem Spiel läuft** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Zwei Dinge, die nur eines gemeinsam haben
//
// **Sie ändern den Stand nicht.** Ein Satz im Gespräch und ein Mauszeiger auf
// einer Karte gehen ans Fenster und nicht in den Strom; das Issue verlangt die
// Trennung ausdrücklich: „bestätigte Events, Kontrollnachrichten, Chat und
// flüchtige Präsenz getrennt".
//
// **Sonst sind sie Gegensätze.** Das Gespräch wird aufbewahrt und hat
// Sichtrechte; die Präsenz ist flüchtig — keine Folge, kein Kettenglied, keine
// Aufnahme in ein Replay.
//
// # Warum beide sagen, wer spricht
//
// **Weil der Umschlag etwas anderes sagt.** [Partienachricht.Von] ist der
// Absender **dieses Sprungs**: reicht ein Wirt Annas Satz an Bert weiter,
// steht dort der Wirt — er hat ihn geschickt, und das stimmt. Wer **geredet**
// hat, steht in der Nutzlast.
//
// **Hineingeschrieben wird sie vom Wirt**, aus der Sitzung, über die die
// Nachricht kam. Ein Gast schreibt sie auch, und sie gilt nicht — dieselbe
// Regel wie bei [Spielbefehl.Absender], und derselbe Schaden, wenn sie fehlt:
// ein Gast redete im Namen eines anderen.
//
// # Und warum der Empfänger einen Stand führt
//
// **Weil Präsenz verwerfbar zugestellt wird.** Sie hat keine Folge — hätte sie
// eine, blockierte jede verlorene Mausbewegung den Ereignisstrom. Damit kann
// eine ältere nach einer neueren ankommen, und ohne Stand spränge der Zeiger
// zurück. Das Issue nennt es beim Namen: „monotonen Empfängerstand und
// Ablauf".
//
// **Der Ablauf steht nicht hier**, und das ist eine Aussage und kein
// Versehen (§10). Er ist eine Frage an die Wanduhr und an das, was jemand
// **sieht** — er gehört dorthin, wo er sichtbar wird, und das ist das
// Fenster. Hier steht die Zusage, die das Fenster **nicht** halten kann: eine
// Präsenz, die das andere Ende überholt hat, kommt gar nicht erst an.

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

// Chatlaenge ist die Grenze, ab der ein Satz kein Satz mehr ist.
//
// **In Zeichen und nicht in Bytes.** Ein Satz mit Umlauten ist nicht kürzer,
// weil er mehr Bytes braucht — wer in Bytes zählte, schnitte deutschen Text
// früher ab als englischen.
const Chatlaenge = 500

// Chat ist, was jemand am Tisch gesagt hat.
type Chat struct {
	// Von ist der Sitz, der geredet hat — **nicht** der, der die Nachricht
	// geschickt hat. Der Grund steht im Kopf dieser Datei.
	Von string `json:"von"`

	Text string `json:"text"`
}

// Pruefe hält einen Satz gegen seine Grenzen.
func (c Chat) Pruefe() error {
	switch {
	case strings.TrimSpace(c.Von) == "":
		return fmt.Errorf("%w: ein Satz ohne Sprecher lässt sich nicht "+
			"zuordnen", ErrValidation)

	case strings.TrimSpace(c.Text) == "":
		return fmt.Errorf("%w: ein leerer Satz sagt nichts", ErrValidation)

	case utf8.RuneCountInString(c.Text) > Chatlaenge:
		return fmt.Errorf("%w: ein Satz fasst %d Zeichen, hier sind es %d",
			ErrValidation, Chatlaenge, utf8.RuneCountInString(c.Text))

	// **Kein Steuerzeichen und keine Richtungsumkehr** — QA-06 (#1737),
	// gefunden von `FuzzChat`. Ein Satz kommt von einem Fremden und steht in
	// jedem Fenster am Tisch; ein U+202E dreht, was die anderen lesen. Das
	// Fenster schickt nur eine Zeile, also verliert niemand einen Umbruch.
	// **Formatzeichen sonst bleiben**: U+200D fügt Emoji zusammen.
	case strings.IndexFunc(c.Text, steuerzeichenOderUmkehr) >= 0:
		return fmt.Errorf("%w: ein Satz enthält ein Steuerzeichen oder eine "+
			"Richtungsumkehr", ErrValidation)
	}

	return nil
}

// steuerzeichenOderUmkehr erkennt ein Steuerzeichen (Cc) oder eines der
// Zeichen, die die Leserichtung umstellen — U+202A bis U+202E und U+2066 bis
// U+2069.
func steuerzeichenOderUmkehr(zeichen rune) bool {
	return unicode.Is(unicode.Cc, zeichen) ||
		(zeichen >= 0x202A && zeichen <= 0x202E) || (zeichen >= 0x2066 && zeichen <= 0x2069)
}

// Praesenzart sagt, was jemand gerade tut.
type Praesenzart string

const (
	// PraesenzZeiger ist der Mauszeiger auf dem Tisch.
	PraesenzZeiger Praesenzart = "zeiger"

	// PraesenzGreift ist eine Karte, die gerade gezogen wird.
	PraesenzGreift Praesenzart = "greift"

	// PraesenzSchwebt ist ein Verweilen über einer Karte.
	PraesenzSchwebt Praesenzart = "schwebt"

	// PraesenzPing ist „hierher schauen" — die einzige, die ausdrücklich
	// gemeint ist statt nebenher zu entstehen.
	PraesenzPing Praesenzart = "ping"
)

// Praesenzarten gibt alle vier, in fester Reihenfolge.
func Praesenzarten() []Praesenzart {
	return []Praesenzart{
		PraesenzZeiger, PraesenzGreift, PraesenzSchwebt, PraesenzPing,
	}
}

// Pruefe sagt, ob es die Art gibt.
func (a Praesenzart) Pruefe() error {
	if slices.Contains(Praesenzarten(), a) {
		return nil
	}

	return fmt.Errorf("%w: %q ist keine Präsenzart — es gibt %v",
		ErrValidation, string(a), Praesenzarten())
}

// Praesenz ist ein Zeiger, ein Griff, ein Schweben oder ein Ping.
type Praesenz struct {
	// Von ist der Sitz, der zeigt — siehe den Kopf dieser Datei.
	Von string `json:"von"`

	Art Praesenzart `json:"art"`

	// Stand zählt beim **Absender** hoch und macht die Reihenfolge lesbar.
	//
	// **Er gehört in die Nutzlast und nicht in den Umschlag**: dort steht
	// die Folge, und die gehört dem Ereignisstrom. Eine Präsenz, die eine
	// Folge verbrauchte, ließe bei jeder verworfenen Mausbewegung eine Lücke
	// zurück — und eine Lücke hält den Strom an.
	Stand uint64 `json:"stand"`

	// Objekt ist die Karte, auf die gezeigt wird — leer heißt: keine.
	Objekt string `json:"objekt"`

	// X und Y sind die Stelle auf dem Tisch — in **Millimetern der Tischwelt**.
	//
	// **Nicht in Bildpunkten und nicht als Anteil.** Zwei Fenster sind
	// verschieden groß und stehen an verschiedenen Stellen der Kamera; ein
	// Zeiger bei 800 **Punkten** zeigte beim anderen ins Leere. Die Tischwelt
	// dagegen rechnen beide Enden aus derselben Sicht — eine Stelle darin
	// bedeutet überall dasselbe.
	//
	// **Wie groß der Tisch ist, weiß diese Schicht nicht** und soll es nicht
	// wissen: das ist eine Frage des Aufbaus und nicht der Fachlichkeit. Was
	// hier geprüft wird, ist deshalb nur, dass es überhaupt Zahlen sind —
	// siehe [Tischweite].
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Pruefe hält eine Präsenz gegen ihre Gestalt.
func (p Praesenz) Pruefe() error {
	if strings.TrimSpace(p.Von) == "" {
		return fmt.Errorf("%w: eine Präsenz ohne Sitz zeigt für niemanden",
			ErrValidation)
	}

	if fehler := p.Art.Pruefe(); fehler != nil {
		return fehler
	}

	if p.Stand == 0 {
		return fmt.Errorf("%w: eine Präsenz ohne Stand lässt sich nicht "+
			"einordnen", ErrValidation)
	}

	return p.pruefeStelle()
}

// Tischweite ist die Grenze, hinter der eine Zahl keine Stelle mehr ist.
//
// **Hunderttausend Millimeter, also hundert Meter** — großzügig mit Absicht.
// Diese Schicht kennt den Aufbau des Tisches nicht und soll ihn nicht kennen;
// sie weist deshalb **Unsinn** ab und nicht „außerhalb". Wer hier enger
// prüfte, prüfte eine Annahme über die Oberfläche.
const Tischweite = 100_000

// pruefeStelle hält die Stelle in einer Zahl, die man zeichnen kann.
func (p Praesenz) pruefeStelle() error {
	if fehler := stelleImTisch("x", p.X); fehler != nil {
		return fehler
	}

	return stelleImTisch("y", p.Y)
}

// stelleImTisch weist alles ab, was keine zeichenbare Zahl ist.
//
// **`NaN` und `Inf` ausdrücklich** — beide sind weder kleiner noch größer als
// etwas, und ein Vergleich allein ließe sie durch. Am anderen Ende stünde ein
// Zeiger, den niemand zeichnen kann.
func stelleImTisch(name string, zahl float64) error {
	if math.IsNaN(zahl) || math.IsInf(zahl, 0) ||
		zahl < -Tischweite || zahl > Tischweite {
		return fmt.Errorf(
			"%w: %s liegt bei %v und ist keine Stelle auf einem Tisch",
			ErrValidation, name, zahl)
	}

	return nil
}

// Praesenzkreis hält je Sitz den zuletzt angenommenen Stand.
//
// **Er urteilt und zeigt nicht.** Was auf dem Tisch zu sehen ist, weiß das
// Fenster; hier wird entschieden, ob eine Nachricht überhaupt neu ist.
type Praesenzkreis struct {
	schloss sync.Mutex
	staende map[string]uint64
}

// NeuerPraesenzkreis fängt ohne einen einzigen Sitz an.
func NeuerPraesenzkreis() *Praesenzkreis {
	return &Praesenzkreis{staende: map[string]uint64{}}
}

// Nimm sagt, ob diese Präsenz neu ist — und merkt sie sich, wenn sie es ist.
//
// **Gleich gilt als alt.** Eine zweite Zustellung derselben Nachricht ist der
// Normalfall einer Leitung, und sie soll nichts bewirken — dieselbe
// Überlegung wie bei [Empfaengerstand].
func (k *Praesenzkreis) Nimm(praesenz Praesenz) bool {
	k.schloss.Lock()
	defer k.schloss.Unlock()

	if bisher, kennt := k.staende[praesenz.Von]; kennt &&
		praesenz.Stand <= bisher {
		return false
	}

	k.staende[praesenz.Von] = praesenz.Stand

	return true
}

// Geht vergisst einen Sitz.
//
// **Ohne diesen Weg käme niemand zurück.** Ein Gast, der neu verbindet, fängt
// wieder bei eins an; ein Kreis, der seinen alten Stand behielte, wiese alles
// ab, was danach kommt — und der Zeiger bliebe für immer weg.
func (k *Praesenzkreis) Geht(sitz string) {
	k.schloss.Lock()
	defer k.schloss.Unlock()

	delete(k.staende, sitz)
}
