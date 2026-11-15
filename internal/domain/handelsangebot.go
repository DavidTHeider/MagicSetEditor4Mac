package domain

// **Der Handelsvertrag** — Werkstück `T1` von Phase 5.6
// (#1466).
//
// # Vier Verträge und nicht einer
//
// **Weil vier verschiedene Dinge „Handel" heißen** und nur einer davon
// Exemplare bewegt:
//
//   - **Teilen** gibt eine Sammlungssicht frei. Es wechselt nichts.
//   - **Deckleihe** reicht einen eingefrorenen Deckstand weiter (`G6`,
//     #1463).
//     Sie verändert **null** physische Bestände — das Deck wechselt nicht den
//     Besitzer.
//   - **Tausch** bewegt Exemplare zwischen zwei Beständen, und zwar von Hand:
//     zwei SQLite-Dateien sind keine gemeinsame Transaktion.
//   - **Markt** ist eine Handlung bei einem fremden Anbieter.
//
// **Ein Vertrag für alle vier hätte Zustände, die für drei davon nichts
// bedeuten** — und ein Feld, das meistens leer ist, wird irgendwann gefüllt.
//
// # Warum eine Revision und kein Gespräch
//
// **Weil eine Zusage sich auf einen Stand bezieht.** Ein Angebot ist
// unveränderlich; wer etwas ändert, erzeugt die **nächste** Revision und
// nennt den Abdruck der vorigen. Damit ist „dem habe ich zugestimmt" eine
// nachrechenbare Aussage und keine Erinnerung.
//
// **Und deshalb werden beide Zustimmungen mit jeder Änderung ungültig**: sie
// galten der Revision davor, und die gibt es noch — nur ist sie nicht mehr
// die letzte.

import (
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"strconv"
	"strings"
)

// Handelsart sagt, worum es geht — der Grund für vier steht im Kopf.
type Handelsart string

const (
	// HandelTeilen gibt eine Sammlungssicht frei (T3).
	HandelTeilen Handelsart = "teilen"

	// HandelDeckleihe reicht einen eingefrorenen Deckstand weiter (G6).
	HandelDeckleihe Handelsart = "deckleihe"

	// HandelTausch bewegt Exemplare zwischen zwei Beständen (T2, T6).
	HandelTausch Handelsart = "tausch"

	// HandelMarkt ist eine Handlung bei einem fremden Anbieter (T11).
	HandelMarkt Handelsart = "markt"
)

// BewegtBestand sagt, ob ein Handel dieser Art Exemplare zwischen Beständen
// bewegt — **nur ein Tausch** (der Grund steht im Kopf). Durchgesetzt seit
// TRADE-03 (#1717):
// bis dahin verbuchte eine Deckleihe, die bis zur Übergabe kam, wie ein Tausch.
func (a Handelsart) BewegtBestand() bool { return a == HandelTausch }

// Handelsarten gibt die vier.
func Handelsarten() []Handelsart {
	return []Handelsart{
		HandelTeilen, HandelDeckleihe, HandelTausch, HandelMarkt,
	}
}

// Handelszustand ist ein Feld der Zustandsmatrix.
//
// # Warum dreizehn und nicht fünf
//
// **Weil der Weg über eine Grenze geht.** „Lokal verbucht" und „beidseitig
// verbucht" sind **nicht** dasselbe: zwei Bestände sind keine gemeinsame
// Transaktion, und ein Abschluss auf dieser Seite ist erst dann einer für
// beide, wenn das Gegenüber ihn bestätigt hat. Wer die zwei zusammenlegte,
// zeigte einen Handel als fertig, den nur einer kennt.
type Handelszustand string

const (
	// HandelEntwurf ist, was nur der Initiator sieht.
	HandelEntwurf Handelszustand = "entwurf"

	// HandelAngeboten ist hinausgegangen und wartet.
	HandelAngeboten Handelszustand = "angeboten"

	// HandelVerhandlung heißt: die Gegenseite hat eine Revision geschickt.
	HandelVerhandlung Handelszustand = "verhandlung"

	// HandelAngenommen heißt: beide haben **derselben** Revision zugestimmt.
	HandelAngenommen Handelszustand = "angenommen"

	// HandelVorbereitet heißt: die Exemplare sind reserviert (T2).
	HandelVorbereitet Handelszustand = "vorbereitet"

	// HandelUebergabeOffen ist der Zustand am Tisch — von Hand zu Hand.
	HandelUebergabeOffen Handelszustand = "uebergabeOffen"

	// HandelLokalVerbucht heißt: **dieser** Bestand ist fortgeschrieben.
	HandelLokalVerbucht Handelszustand = "lokalVerbucht"

	// HandelAbgleichOffen heißt: die Gegenbestätigung fehlt noch.
	HandelAbgleichOffen Handelszustand = "abgleichOffen"

	// HandelBeidseitigVerbucht ist der einzige Zustand, der „fertig" heißt.
	HandelBeidseitigVerbucht Handelszustand = "beidseitigVerbucht"

	// HandelAbgelehnt hat die Gegenseite beendet.
	HandelAbgelehnt Handelszustand = "abgelehnt"

	// HandelAbgebrochen hat der Initiator beendet.
	HandelAbgebrochen Handelszustand = "abgebrochen"

	// HandelAbgelaufen ist die Frist.
	HandelAbgelaufen Handelszustand = "abgelaufen"

	// HandelKlaerung ist der Fall, den niemand automatisch löst.
	HandelKlaerung Handelszustand = "klaerung"
)

// Handelszustaende gibt die dreizehn in der Reihenfolge des Weges.
func Handelszustaende() []Handelszustand {
	return []Handelszustand{
		HandelEntwurf, HandelAngeboten, HandelVerhandlung, HandelAngenommen,
		HandelVorbereitet, HandelUebergabeOffen, HandelLokalVerbucht,
		HandelAbgleichOffen, HandelBeidseitigVerbucht,
		HandelAbgelehnt, HandelAbgebrochen, HandelAbgelaufen, HandelKlaerung,
	}
}

// Handelsposition ist eine Zeile eines Angebots.
//
// **Sie nennt eine Seite und nicht eine Richtung**: „von anna" sagt, wessen
// Bestand sie verlässt, und das ist auch dann eindeutig, wenn drei Parteien
// am Vertrag hängen.
type Handelsposition struct {
	// Von ist die Partei, deren Bestand die Position verlässt.
	Von string

	// Printing ist die Scryfall-Kennung des Drucks — englisch, weil sie
	// eine fremde ist (§9).
	Printing string

	// DruckID ist die lokale Zeile des Drucks (A2) — **neben** [Printing]
	// und nicht an seiner Stelle, wie bei [Exemplar].
	DruckID int64

	// Menge ist die Stückzahl — mindestens eine.
	Menge int

	// Guete ist der Zustand des Exemplars — `NM`, `LP`, …
	//
	// **Nicht „Zustand"**: den Namen trägt schon [Handelszustand], und zwei
	// Bedeutungen unter einem Wort sind die Verwechslung, die man später
	// nicht mehr findet.
	Guete string

	// Sprache ist die Sprache des Drucks.
	Sprache string

	// Finish ist `nonfoil`, `foil`, `etched` — englisch, weil es bei
	// Scryfall so heißt.
	Finish string

	// Abbild ist der Abdruck des Exemplarstands beim Anbieten.
	//
	// **Er bindet, was sonst niemand sieht**: wer zwei Wochen später
	// verbucht, verbucht gegen denselben Stand, gegen den unterschrieben
	// wurde — oder er bekommt einen Konflikt.
	Abbild string

	// Bewertung ist der eingefrorene Wert in der kleinsten Währungseinheit.
	//
	// **Sie geht nur ein, wenn sie geteilt ist**
	// ([Handelsangebot.BewertungGeteilt]) — der Grund steht dort.
	Bewertung int64
}

// Handelsvertragsfassung ist die Fassung des Vertrags, die dieses Programm
// schreibt und liest.
//
// # Warum eine Pflichtangabe und keine Vermutung
//
// **Weil ein Angebot über eine Leitung geht.** Wer eine Fassung nicht kennt,
// weiß nicht, was er nicht weiß — und unterschriebe womöglich etwas anderes,
// als er gelesen hat. Eine unbekannte Fassung wird deshalb **abgewiesen**
// statt so gut wie möglich gedeutet ([0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md)).
//
// **Sie geht in die Unterschrift ein**: zwei Fassungen desselben Protokolls
// hätten sonst dieselbe, und die eine Seite unterschriebe eine Lesart, die
// die andere nicht hat.
const Handelsvertragsfassung = 1

// Handelsangebot ist **eine** unveränderliche Revision.
type Handelsangebot struct {
	// Vertragsfassung ist die Protokollfassung dieses Angebots.
	Vertragsfassung int

	// Vorgang ist die Kennung des Handels über alle Revisionen.
	Vorgang string

	// Revision zählt ab eins.
	Revision int

	// Eltern ist der [Handelsangebot.Abdruck] der vorigen Revision.
	//
	// **Leer genau in der ersten.** Eine zweite ohne Eltern hinge an nichts,
	// und eine erste mit Eltern behauptete eine Vorgeschichte, die es nicht
	// gibt.
	Eltern string

	Art       Handelsart
	Initiator string

	// Parteien sind alle Beteiligten, den Initiator eingeschlossen.
	Parteien []string

	Zustand    Handelszustand
	Positionen []Handelsposition

	// BewertungGeteilt sagt, ob der Wert **Vertragsinhalt** ist.
	//
	// # Warum das eine Entscheidung ist und kein Anzeigeschalter
	//
	// **Weil eine Marktbasis eine Anzeige ist.** Wer eine andere Ansicht
	// wählt, ändert keine unterschriebene Einigung — solange die Bewertung
	// **nicht** geteilt ist. Ist sie es, ist ihre Änderung eine Änderung des
	// Angebots und braucht eine neue Revision. Beides ist vertretbar; was
	// nicht geht, ist es offenzulassen.
	BewertungGeteilt bool

	// Zustimmungen sind die Parteien, die **dieser** Revision zugestimmt
	// haben.
	//
	// **Sie gehen nicht in den Abdruck ein** — eine Aussage über eine
	// Revision kann nicht Teil der Revision sein, der sie gilt.
	Zustimmungen []string
}

// ZumVersand gibt das Angebot, wie es hinausgeht — **eine ungeteilte
// Bewertung bleibt hier** (TRADE-03.AK-03, #1717). Sie ist eine Anzeige und
// steht dann nicht im Abdruck; was man zu zahlen bereit wäre, erführe das
// Gegenüber sonst nebenbei.
func (a Handelsangebot) ZumVersand() Handelsangebot {
	if a.BewertungGeteilt {
		return a
	}

	a.Positionen = slices.Clone(a.Positionen)
	for stelle := range a.Positionen {
		a.Positionen[stelle].Bewertung = 0
	}

	return a
}

// Unterschriftsform ist, was unterschrieben wird.
//
// # Warum eine gefügte Zeichenkette und kein JSON
//
// **Dieselbe Überlegung wie bei [Deckleihe.Unterschriftsform]**: JSON kennt
// zwei Fassungen derselben Angabe — Einrückung, Feldreihenfolge, die
// Schreibweise einer Zahl. Wer darüber unterschriebe, müsste eine kanonische
// Form verabreden, und **wie** sie aussieht, wäre eine zweite Verabredung,
// die niemand liest.
//
// **Das Nullbyte als Trenner, weil keines der Felder eines enthalten kann**:
// es sind Kennungen, Aufzählungswerte und Zahlen. Und **das Domänenwort
// zuerst**, damit dieselbe Unterschrift nicht für ein anderes Dokument mit
// denselben Feldern gilt.
func (a Handelsangebot) Unterschriftsform() []byte {
	// **Zehn Kopffelder und neun je Position** — die Zahl steht hier, damit
	// die Scheibe einmal wächst und nicht zehnmal.
	teile := make([]string, 0, 10+9*len(a.Positionen))
	teile = append(teile,
		"handelsangebot",
		strconv.Itoa(a.Vertragsfassung),
		a.Vorgang,
		strconv.Itoa(a.Revision),
		a.Eltern,
		string(a.Art),
		a.Initiator,
		strings.Join(a.Parteien, ","),
		strconv.FormatBool(a.BewertungGeteilt),
		strconv.Itoa(len(a.Positionen)),
	)

	for _, position := range a.Positionen {
		teile = append(teile, position.form(a.BewertungGeteilt)...)
	}

	return []byte(strings.Join(teile, "\x00"))
}

// form gibt die Felder einer Position in fester Reihenfolge.
//
// **Die Bewertung nur, wenn sie geteilt ist** — sonst bände die Unterschrift
// eine Zahl, über die niemand gesprochen hat, und jede neue Marktansicht
// machte eine gültige Einigung ungültig.
func (p Handelsposition) form(mitBewertung bool) []string {
	teile := []string{
		p.Von,
		p.Printing,
		strconv.FormatInt(p.DruckID, 10),
		strconv.Itoa(p.Menge),
		p.Guete,
		p.Sprache,
		p.Finish,
		p.Abbild,
	}

	if mitBewertung {
		teile = append(teile, strconv.FormatInt(p.Bewertung, 10))
	}

	return teile
}

// Abdruck ist der Hash der [Handelsangebot.Unterschriftsform].
//
// **Er ist die Kennung einer Revision** — die nächste nennt ihn als Eltern,
// und eine Zustimmung nennt ihn, statt „ja" zu sagen.
func (a Handelsangebot) Abdruck() string {
	summe := sha256.Sum256(a.Unterschriftsform())

	return hex.EncodeToString(summe[:])
}

// Fortgeschrieben gibt die nächste Revision — ohne Zustimmungen.
//
// **Nach einer Änderung sind beide Zustimmungen ungültig**, und das ist keine
// Höflichkeit: sie galten einem Stand, der nicht mehr der letzte ist. Wer sie
// mitnähme, hätte eine Einigung über etwas, das niemand gelesen hat.
//
// # Warum sie den Zustand auf „Verhandlung" setzt
//
// **Weil eine Änderung genau das ist.** Ein Angebot, das nach einer Änderung
// weiter „angenommen" hieße, sagte, dass sich zwei über etwas einig sind,
// das einer von ihnen gerade umgeschrieben hat.
//
// # Und warum sie nicht sequenziert
//
// **Weil das der Initiator tut.** Zwei Vorschläge gegen denselben Stand
// ergeben zwei Angebote mit **denselben** Eltern und derselben Nummer — und
// genau daran erkennt man den Konflikt. Wer hier durchnummerierte, machte aus
// zwei Vorschlägen zwei gültige Stände (T5,
// #1470).
func (a Handelsangebot) Fortgeschrieben(
	positionen []Handelsposition,
) Handelsangebot {
	naechste := a
	naechste.Eltern = a.Abdruck()
	naechste.Revision = a.Revision + 1
	naechste.Positionen = slices.Clone(positionen)
	naechste.Zustimmungen = nil
	naechste.Zustand = HandelVerhandlung

	return naechste
}
