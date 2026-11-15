package domain

// **Die Sicht auf eine Sammlung** — Werkstück #1500 von Phase 5.2
// (#1500),
// Befund aus U8.
//
// # Warum das hierher gewandert ist
//
// **Es stand im Fenster** (`frontend/src/seiten/sammlungssicht.ts`), und der
// Kommentar dort sagte, wann es zu ändern wäre: „Bei zehntausend Exemplaren
// würde die ganze Liste weiterhin geholt und hier gefiltert … wer die Grenze
// verschiebt, verschiebt sie in den Dienst — und dann gehört auch die
// Sortierung dorthin."
//
// **Die Grenze ist verschoben.** Gemessen am 2026-09-09: bei 10.000 Exemplaren
// steht das Gitter nach 209 bis 371 ms, und der Grund ist nicht das Zeichnen —
// es sind rund 3 MB JSON über die Wails-Naht.
//
// # Und warum die Ordnung in Go bleibt und nicht in SQL
//
// **Weil SQLite keine Sprache kennt.** Sein `COLLATE NOCASE` ist ASCII;
// „Ätherfluss" stünde damit hinter „Zauber". Das Fenster ordnete mit
// `localeCompare`, und wer die Ordnung in den Bestand verschiebt, ohne das zu
// prüfen, ändert sie für jeden Namen mit Umlaut.

import (
	"bytes"
	"sort"
	"strings"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// Ausfuehrung ist der Filter über die Ausführung eines Drucks.
//
// **Vier und nicht drei** (T2,
// #1467):
// `etched` kam dazu, als aus `Foil bool` ein [Finish] wurde. Ein Filter, der
// es nicht kennte, zeigte geätzte Karten unter „nonfoil" — weil sie nicht
// foil sind.
type Ausfuehrung string

const (
	AusfuehrungAlle    Ausfuehrung = "alle"
	AusfuehrungFoil    Ausfuehrung = "foil"
	AusfuehrungNonfoil Ausfuehrung = "nonfoil"
	AusfuehrungEtched  Ausfuehrung = "etched"
)

// Sortierung ist die Ordnung der Liste.
//
// **Drei und nicht neun.** Nach Zustand oder Währung zu sortieren beantwortet
// keine Frage, die jemand an seine Sammlung stellt; nach Karte, Menge und
// Erwerbsdatum sehr wohl.
type Sortierung string

const (
	SortierungBezeichnung  Sortierung = "bezeichnung"
	SortierungMenge        Sortierung = "menge"
	SortierungErwerbsdatum Sortierung = "erwerbsdatum"
)

// Sammlungssicht ist, was jemand von seiner Sammlung sehen will.
type Sammlungssicht struct {
	// Suche geht über Bezeichnung, Druck, Setcode und die eigene Notiz.
	Suche string `json:"suche"`

	Ausfuehrung Ausfuehrung `json:"ausfuehrung"`
	Sortierung  Sortierung  `json:"sortierung"`

	// Sprache ist die des Fensters — LOC-03.AK-01
	// (#1689).
	//
	// **Sie ordnet die Bezeichnungen.** Wer eine Sammlung in einer Sprache
	// liest, liest sie in deren Ordnung; eine fest eingebaute wäre für
	// jeden falsch, der sie nicht spricht.
	//
	// **Leer heißt Deutsch** — das ist, was bis heute galt, und damit die
	// Vorgabe der Flächen ohne Fenster. Ein unbekanntes Kürzel heißt
	// dasselbe: geraten wird nicht.
	Sprache string `json:"sprache"`
}

// Pruefe weist eine Sicht ab, die es so nicht gibt.
//
// **Leer ist erlaubt und heißt „alles"** — ein Aufruf ohne Angaben ist der
// häufige, und er soll nicht an einer Vorgabe scheitern, die er nicht kennt.
func (s Sammlungssicht) Pruefe() error {
	switch s.Ausfuehrung {
	case "", AusfuehrungAlle, AusfuehrungFoil, AusfuehrungNonfoil,
		AusfuehrungEtched:
	default:
		return NewFieldError("ausfuehrung", SchluesselValidierung,
			"%q ist keine Ausführung — es gibt »alle«, »nonfoil«, »foil« "+
				"und »etched«.",
			string(s.Ausfuehrung))
	}

	switch s.Sortierung {
	case "", SortierungBezeichnung, SortierungMenge, SortierungErwerbsdatum:
	default:
		return NewFieldError("sortierung", SchluesselValidierung,
			"%q ist keine Ordnung — es gibt »bezeichnung«, »menge« und "+
				"»erwerbsdatum«.", string(s.Sortierung))
	}

	return nil
}

// Gesehen gibt die Exemplare, die zu dieser Sicht gehören — geordnet.
//
// **Eine Kopie und keine Ordnung an Ort und Stelle**: `sort` änderte sonst das
// Feld, das es bekommt, und das ist hier die gelesene Liste (§7).
func (s Sammlungssicht) Gesehen(exemplare []*Exemplar) []*Exemplar {
	gesucht := strings.ToLower(strings.TrimSpace(s.Suche))

	gefiltert := make([]*Exemplar, 0, len(exemplare))

	for _, eines := range exemplare {
		if s.passtZurSuche(eines, gesucht) && s.passtZurAusfuehrung(eines) {
			gefiltert = append(gefiltert, eines)
		}
	}

	s.ordne(gefiltert)

	return gefiltert
}

// passtZurSuche prüft den Text gegen das, was ein Mensch sucht.
//
// **Über vier Felder und nicht über eines.** Wer „bolt" tippt, meint den
// Namen; wer „m11" tippt, das Set; wer eine Kennung einfügt, den Druck; wer
// „getauscht" tippt, seine eigene Notiz.
func (s Sammlungssicht) passtZurSuche(exemplar *Exemplar, gesucht string) bool {
	if gesucht == "" {
		return true
	}

	for _, feld := range []string{
		exemplar.Bezeichnung(), exemplar.Printing, exemplar.Setcode, exemplar.Notiz,
	} {
		if strings.Contains(strings.ToLower(feld), gesucht) {
			return true
		}
	}

	return false
}

// passtZurAusfuehrung hält die Ausführung gegen den Filter.
//
// **Über [Exemplar.Ausfuehrung] und nicht über das Feld**: ein Exemplar ohne
// Angabe ist nonfoil, und ein Filter, der das Feld läse, fände es unter
// keinem der drei Namen.
func (s Sammlungssicht) passtZurAusfuehrung(exemplar *Exemplar) bool {
	if s.Ausfuehrung == "" || s.Ausfuehrung == AusfuehrungAlle {
		return true
	}

	return string(exemplar.Ausfuehrung()) == string(s.Ausfuehrung)
}

// ordne bringt die gefilterten Exemplare in die gewählte Ordnung.
//
// **Die Karte ist immer der zweite Schlüssel.** Zwei Zeilen mit derselben
// Menge stünden sonst in einer Reihenfolge, die sich bei jedem Nachladen
// ändern kann — und eine Liste, die springt, liest niemand gern.
func (s Sammlungssicht) ordne(exemplare []*Exemplar) {
	// **Ein Ordner je Lauf und kein gemeinsamer.**
	//
	// `collate.Collator` trägt einen Zwischenspeicher und ist **nicht**
	// nebenläufig benutzbar. Ein Paketwert davon gab in zwei gleichzeitig
	// laufenden Tests die Eingabereihenfolge zurück — jeder Vergleich sagte
	// „nicht kleiner", und die Liste blieb, wie sie war. Gesehen am
	// 2026-09-10; im Betrieb wären es zwei Fenster, die zugleich lesen.
	//
	// **Die Sprache kommt aus der Sicht** (LOC-03.AK-01, #1689).
	//
	// **Für Deutsch und Englisch ist die Ordnung dieselbe**, und das ist
	// keine Ausrede, sondern eine Messung: beide legen »ä« neben »a«. Der
	// Unterschied fängt bei der dritten Sprache an — im Schwedischen steht
	// »ä« hinter »z« —, und LOC-01 bringt sie.
	ordner := collate.New(ordnungssprache(s.Sprache))

	// **Ein Schlüssel je Zeile statt einer Zerlegung je Vergleich.**
	//
	// `CompareString` zerlegt bei **jedem** Aufruf beide Zeichenketten neu,
	// und ein Sortierlauf über zehntausend Zeilen ruft es rund 120.000 mal.
	// `KeyFromString` zerlegt einmal je Zeile; verglichen werden danach
	// Bytes. Gemessen am 2026-09-10 an zehntausend Exemplaren (Apple M5,
	// `-benchtime=10x`): **62,4 ms → 10,2 ms**, und die Ordnung ist dieselbe
	// — dafür gibt es den Lauf über die Umlaute.
	//
	// **Der Puffer wird nicht zurückgesetzt.** Er hält den Speicher, in dem
	// die Schlüssel liegen; ein `Reset` machte die schon gebauten ungültig.
	var puffer collate.Buffer

	// **Sortiert werden Paare aus Zeile und Schlüssel**, nicht die Zeilen
	// selbst: eine Zuordnung daneben — ein `map` etwa — kostete bei jedem
	// Vergleich eine Streuwertsuche und wäre der halbe Gewinn wieder weg.
	reihen := make([]geordnete, len(exemplare))
	for stelle, eines := range exemplare {
		reihen[stelle] = geordnete{
			exemplar:   eines,
			schluessel: ordner.KeyFromString(&puffer, eines.Bezeichnung()),
		}
	}

	sort.SliceStable(reihen, func(links, rechts int) bool {
		einer, anderer := reihen[links], reihen[rechts]

		switch s.Sortierung {
		case SortierungMenge:
			// **Absteigend**: wer nach Menge sortiert, sucht die Stapel.
			if einer.exemplar.Menge != anderer.exemplar.Menge {
				return einer.exemplar.Menge > anderer.exemplar.Menge
			}
		case SortierungErwerbsdatum:
			// **Das Jüngste zuerst** — und Zeilen ohne Datum ans Ende, statt
			// sie wegen der leeren Zeichenkette nach vorn zu stellen.
			if einer.exemplar.Erwerbsdatum != anderer.exemplar.Erwerbsdatum {
				return datumsrang(einer.exemplar.Erwerbsdatum) >
					datumsrang(anderer.exemplar.Erwerbsdatum)
			}
		case "", SortierungBezeichnung:
		}

		return bytes.Compare(einer.schluessel, anderer.schluessel) < 0
	})

	for stelle, reihe := range reihen {
		exemplare[stelle] = reihe.exemplar
	}
}

// geordnete ist eine Zeile samt ihrem Sortierschlüssel.
type geordnete struct {
	exemplar   *Exemplar
	schluessel []byte
}

// datumsrang macht aus `JJJJ-MM-TT` eine vergleichbare Zahl — leer ganz nach
// hinten.
//
// **Ohne `time.Parse`**: was hier steht, ist ein Kalendertag ohne Zone, und
// ein Zeitpunkt daraus trüge eine Mitternacht, die niemand gemeint hat. Die
// Form ist sortierbar, wie sie ist.
func datumsrang(datum string) string {
	if datum == "" {
		return ""
	}

	return datum
}

// Ausschnitt gibt das Fenster [ab, ab+anzahl) einer Liste — samt der Gesamtzahl.
//
// **Die Gesamtzahl gehört dazu.** Ohne sie wüsste das Fenster nicht, wie hoch
// die Liste ist, die es zeichnet — und `aria-rowcount` nennt sie einer
// Vorlesehilfe.
//
// **Eine Anzahl von null heißt „alles"**: die Gruppenansicht braucht die ganze
// Liste, und ein zweiter Weg dafür wären zwei Wahrheiten über dieselbe Sicht.
func Ausschnitt(exemplare []*Exemplar, ab, anzahl int) (fenster []*Exemplar, gesamt int) {
	gesamt = len(exemplare)

	if ab < 0 {
		ab = 0
	}

	if ab >= gesamt {
		return []*Exemplar{}, gesamt
	}

	if anzahl <= 0 {
		return exemplare[ab:], gesamt
	}

	bis := min(ab+anzahl, gesamt)

	return exemplare[ab:bis], gesamt
}

// ordnungssprache deutet das Kürzel der Sicht.
//
// **Leer und Unbekanntes heißen Deutsch** — das ist, was bis zum 2026-09-22
// galt. Wer hier riete, ordnete eine Sammlung nach einer Sprache, die
// niemand gewählt hat (§10).
func ordnungssprache(kuerzel string) language.Tag {
	if strings.TrimSpace(kuerzel) == "" {
		return language.German
	}

	gedeutet, fehler := language.Parse(kuerzel)
	if fehler != nil {
		return language.German
	}

	return gedeutet
}
