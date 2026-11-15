package domain

// **Was ein Tausch im Journal hinterlässt** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// # Warum ein empfangenes Stück ohne Preis kommt
//
// **Weil das, was jemand anderes bezahlt hat, nicht ist, was es mich gekostet
// hat.** Eine Zahl, die aus dem Erwerbspreis des Senders entstünde, sähe
// später aus wie eine erfasste — und niemand könnte sie mehr von einer
// unterscheiden, die wirklich jemand bezahlt hat. **Ein Marktwert wäre
// dasselbe in Grün**: er ist eine Auskunft über heute und keine über den
// Erwerb.
//
// **Leer heißt hier „nicht erfasst" und nicht „wertlos"** — dieselbe
// Unterscheidung wie am Exemplar selbst.
//
// # Warum ein Handelsbeleg nichts Inneres trägt
//
// **Weil er hinausgeht.** Eine Bestandskennung ist eine Nummer dieses
// Bestands, ein Lagerort sagt, wo im Wohnzimmer etwas liegt, und eine Notiz
// gehört dem, der sie geschrieben hat. Was in einem Handelsbeleg steht, muss ein
// Fremder lesen dürfen — derselbe Satz wie bei der freigegebenen Sicht (T3).
//
// # Und warum die Vergangenheit nicht gelöscht wird
//
// **Weil ein Journal, aus dem man etwas herausnehmen kann, keins ist.** Nach
// einem Vollzug gibt es nur die **Gegenbewegung**: sie dreht die Richtung um,
// nennt ihren Bezug und ihren Grund, und beide Zeilen bleiben stehen.

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Journalart sagt, in welche Richtung eine Zeile geht.
type Journalart string

// Die zwei Richtungen.
const (
	// JournalZugang ist, was hereinkommt.
	JournalZugang Journalart = "zugang"

	// JournalAbgang ist, was hinausgeht.
	JournalAbgang Journalart = "abgang"
)

// Journalanlass sagt, **warum** gebucht wurde.
type Journalanlass string

// Die sechs Anlässe — der ganze Wortschatz des Journals.
const (
	// AnlassTausch ist ein Handel mit einem Gegenüber.
	AnlassTausch Journalanlass = "tausch"

	// AnlassKauf und AnlassVerkauf sind Geld gegen Papier.
	AnlassKauf    Journalanlass = "kauf"
	AnlassVerkauf Journalanlass = "verkauf"

	// AnlassGeschenk geht in beide Richtungen ohne Gegenleistung.
	AnlassGeschenk Journalanlass = "geschenk"

	// AnlassEinfuhr ist, was aus einer Liste oder einer Sicherung kommt —
	// **auch der Ausgangssaldo** beim ersten Mal.
	AnlassEinfuhr Journalanlass = "einfuhr"

	// AnlassKorrektur ist die Gegenbewegung zu einer Buchung, die falsch war.
	AnlassKorrektur Journalanlass = "korrektur"
)

// alleAnlaesse ist die Aufzählung in fester Reihenfolge.
var alleAnlaesse = []Journalanlass{
	AnlassTausch, AnlassKauf, AnlassVerkauf,
	AnlassGeschenk, AnlassEinfuhr, AnlassKorrektur,
}

// Journalanlaesse gibt die sechs — als Kopie.
func Journalanlaesse() []Journalanlass {
	kopie := make([]Journalanlass, len(alleAnlaesse))
	copy(kopie, alleAnlaesse)

	return kopie
}

// IstJournalanlass sagt, ob es den Anlass gibt.
func IstJournalanlass(anlass Journalanlass) bool {
	for _, kandidat := range alleAnlaesse {
		if kandidat == anlass {
			return true
		}
	}

	return false
}

// Journalzeile ist eine Bewegung im Bestand.
type Journalzeile struct {
	// ID ist die Kennung dieser Zeile im Journal — null, solange sie nicht
	// geschrieben ist.
	//
	// **Sie steht hier, weil eine Korrektur sie braucht.** Der Bezug einer
	// Gegenbewegung ist die **Zeile**, die sie aufhebt, und nicht das
	// Exemplar: das kann längst fort sein, und dann zeigte der Bezug auf
	// nichts.
	ID int64

	Art    Journalart
	Anlass Journalanlass

	// Vorgang ist der Handel, zu dem sie gehört — leer bei allem anderen.
	Vorgang string

	// Bezug ist die Zeile, auf die sich eine Korrektur bezieht.
	Bezug int64

	// Grund steht an einer Korrektur und sonst nirgends.
	Grund string

	Printing string
	Menge    int
	Zustand  Zustand
	Sprache  string
	Finish   Finish

	// ExemplarID, Lagerort und Notiz sind **lokal** — sie stehen im Journal
	// und **nicht** im Handelsbeleg.
	ExemplarID int64
	Lagerort   string
	Notiz      string

	Gebucht time.Time
}

// AlsExemplar macht aus einem Zugang die Zeile, die in den Bestand geht.
//
// **Ohne Erwerbspreis und ohne Währung** — der Grund steht im Kopf dieser
// Datei. Die Herkunft geht mit: wer später fragt, woher das Stück kam,
// bekommt eine Antwort statt einer Vermutung.
func (z Journalzeile) AlsExemplar() Exemplar {
	return Exemplar{
		Printing: z.Printing,
		Menge:    z.Menge,
		Zustand:  z.Zustand,
		Sprache:  z.Sprache,
		Finish:   z.Finish,
		Lagerort: z.Lagerort,
		Herkunft: Herkunft{Art: HerkunftEinfuhr, Quelle: z.Vorgang},
	}
}

// Gegenbewegung gibt die Zeile, die eine falsche Buchung aufhebt.
//
// **Ohne Grund gibt es keine.** Eine Buchung, die niemand erklärt, ist im
// Journal schlimmer als keine: sie steht da, sie stimmt nicht mit dem
// Bestand überein, und niemand weiß mehr, warum.
func (z Journalzeile) Gegenbewegung(grund string) (Journalzeile, error) {
	if strings.TrimSpace(grund) == "" {
		return Journalzeile{}, fmt.Errorf(
			"%w: eine Korrektur nennt ihren Grund", ErrValidation)
	}

	// **Und ohne geschriebene Zeile auch nicht.** Eine Korrektur, die auf
	// nichts zeigt, ist keine Korrektur, sondern eine zweite Buchung.
	if z.ID == 0 {
		return Journalzeile{}, fmt.Errorf(
			"%w: eine Korrektur nennt die Zeile, die sie aufhebt", ErrValidation)
	}

	gegen := z
	gegen.ID = 0
	gegen.Anlass, gegen.Grund, gegen.Bezug = AnlassKorrektur, grund, z.ID
	gegen.ExemplarID = 0

	if z.Art == JournalZugang {
		gegen.Art = JournalAbgang
	} else {
		gegen.Art = JournalZugang
	}

	return gegen, nil
}

// Pruefe sagt, ob diese Zeile ins Journal darf.
//
// **Geprüft wird vor dem Schreiben und nicht danach**, weil danach nichts mehr
// zu berichtigen ist: aus dem Journal wird nichts herausgenommen, eine falsche
// Zeile bleibt für immer stehen und bekommt allenfalls eine Gegenbewegung
// daneben. Die Tabelle prüft dasselbe noch einmal — sie ist die letzte
// Schranke, nicht die erste.
func (z Journalzeile) Pruefe() error {
	switch {
	case z.Art != JournalZugang && z.Art != JournalAbgang:
		return fmt.Errorf("%w: %q ist keine Richtung", ErrValidation, z.Art)

	case !IstJournalanlass(z.Anlass):
		return fmt.Errorf("%w: %q ist kein Anlass", ErrValidation, z.Anlass)

	case strings.TrimSpace(z.Printing) == "":
		return fmt.Errorf("%w: eine Journalzeile nennt ihren Druck",
			ErrValidation)

	case z.Menge < 1:
		return fmt.Errorf("%w: eine Journalzeile bewegt mindestens ein Stück",
			ErrValidation)

	// **Eine Korrektur nennt beides** — worauf sie sich bezieht und warum.
	// Mit nur einem von beidem sieht sie aus wie eine Buchung.
	case z.Anlass == AnlassKorrektur &&
		(z.Bezug == 0 || strings.TrimSpace(z.Grund) == ""):
		return fmt.Errorf(
			"%w: eine Korrektur nennt ihren Bezug und ihren Grund",
			ErrValidation)

	// **Und ein Tausch nennt seinen Vorgang.** Ohne ihn stünde die Bewegung
	// da, ohne dass jemand den Handel dazu fände.
	case z.Anlass == AnlassTausch && strings.TrimSpace(z.Vorgang) == "":
		return fmt.Errorf("%w: eine Tauschzeile nennt ihren Vorgang",
			ErrValidation)
	}

	return nil
}

// Handelsbeleg ist, was über einen Abschluss hinausgeht.
//
// **Nicht `Beleg`** — den Namen trägt seit Phase 5.5 der Nachweis einer
// Regelstelle (`belegpruefung.go`). Zwei Belege unter einem Namen wären die
// Verwechslung, die man später nicht mehr findet.
type Handelsbeleg struct {
	Vorgang string

	// Abdruck ist der Abdruck der Revision, die abgeschlossen wurde.
	Abdruck string

	Parteien []string
	Zeilen   []Journalzeile

	// Beidseitig sagt, ob beide Seiten unterschrieben haben.
	//
	// **Zwei Unterschriften vorzutäuschen wäre das Schlimmste**: der lokale
	// Abschluss ist echt, der beidseitige noch nicht — und wer beides gleich
	// aussehen ließe, machte aus einer offenen Sache eine erledigte.
	Beidseitig bool

	Gebucht time.Time
}

// Unterschriftsform gibt die Bytes, über die unterschrieben wird.
//
// **Nur das Öffentliche.** Bestandskennung, Lagerort und Notiz stehen nicht
// darin — sie gehen einen Fremden nichts an, und ein Handelsbeleg geht hinaus.
func (b Handelsbeleg) Unterschriftsform() []byte {
	teile := make([]string, 0, 6+5*len(b.Zeilen))
	teile = append(teile,
		"handelsbeleg",
		b.Vorgang,
		b.Abdruck,
		strings.Join(b.Parteien, ","),
		strconv.FormatBool(b.Beidseitig),
		strconv.Itoa(len(b.Zeilen)),
	)

	for _, zeile := range b.Zeilen {
		teile = append(teile,
			string(zeile.Art),
			zeile.Printing,
			strconv.Itoa(zeile.Menge),
			string(zustandOder(zeile.Zustand))+"/"+zeile.Sprache,
			string(ausfuehrungOder(zeile.Finish)),
		)
	}

	return []byte(strings.Join(teile, "\x00"))
}
