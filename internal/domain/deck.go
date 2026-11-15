package domain

// **Das Deck** — Werkstück 2 von Phase 4.6
// (#1293),
// nach Vermerk [0061](../../docs/entscheidungen/0061-was-ein-deck-hier-ist.md).
//
// # Warum `new` und nicht `ported`
//
// **Gesucht am 2026-09-01** in allen drei Quellen aus §11: MSE2 (`f9d9356`)
// kennt weder `deck` noch `sideboard` noch `legal`; die `.mse-game`- und
// `.mse-style`-Pakete des Packs führen die Wörter nur in Regeltexten und eine
// **Packart** namens „tournament pack"; `magic-set-generator` kennt nichts
// davon. Die Belege und ihre Reichweite stehen im Vermerk.
//
// # Warum ein Deck kein Set ist
//
// Ein Set ist ein **Entwurf**, ein Deck eine **Zusammenstellung**. Ein Deck
// zieht aus mehreren Sets zugleich und enthält dieselbe Karte mehrfach; ein
// Set kennt jede Karte einmal. Wäre das Deck ein Set, müsste es Karten
// **kopieren** — und eine berichtigte Karte im Ursprungsset erreichte es nie.

import (
	"regexp"
	"strings"
)

// Deckbereich sagt, wo ein Platz im Deck liegt.
//
// **Zwei und nicht mehr.** Was ein drittes wäre, weiß niemand; ein leerer
// Bereich wäre eine stille Zuordnung.
type Deckbereich string

const (
	// BereichHauptdeck ist das Deck, mit dem gespielt wird.
	BereichHauptdeck Deckbereich = "hauptdeck"

	// BereichSideboard ist, was zwischen zwei Partien dazukommt.
	//
	// **Warum es überhaupt eines gibt**: es ist der einzige Teil eines Decks,
	// der außerhalb einer Partie existiert, und jedes Format hat eigene Zahlen
	// dafür. Es wegzulassen hieße, die Formatprüfung später zu zerreißen.
	BereichSideboard Deckbereich = "sideboard"
)

// IstDeckbereich sagt, ob ein Wert einer der beiden ist.
func IstDeckbereich(bereich Deckbereich) bool {
	return bereich == BereichHauptdeck || bereich == BereichSideboard
}

// druckkennung trifft eine Scryfall-UUID.
//
// **Dieselbe Gestalt wie in der Sammlung und im Kartenverweis** — es ist
// dieselbe Kennung, und daran hängt später „habe ich das?".
var druckkennung = regexp.MustCompile(
	`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Deckplatz ist ein Eintrag im Deck.
//
// **Er verweist, er kopiert nicht** — dieselbe Überlegung wie bei den
// Autoersetzungen ([0060](../../docs/entscheidungen/0060-eigene-autoersetzungen-liegen-im-bestand.md)).
type Deckplatz struct {
	// ID ist die Kennung im Bestand — 0 für einen, den es dort noch nicht gibt.
	ID int64

	// KarteID nennt eine **eigene** Karte, Printing einen **echten** Druck.
	//
	// **Genau eines von beiden.** Ein Platz ohne Bezug ist kein Platz; einer
	// mit beidem ließe offen, was gilt.
	KarteID  int64
	Printing string

	// DruckID ist die lokale Zeile des Drucks (A2, #1441, 0086) — ein
	// echter Druck ohne Alias hat nur sie.
	DruckID int64

	// Menge ist die Stückzahl — mindestens eine.
	Menge int

	// Bereich ist Hauptdeck oder Sideboard.
	Bereich Deckbereich

	// Bezeichnung ist der **Name** der Karte —
	// #1376.
	//
	// **Sie steht nicht im Bestand.** Ein Platz trägt dort eine Kennung; den
	// Namen holt der Dienst beim Lesen — bei einer eigenen Karte aus dem
	// Bestand, bei einem Druck von Scryfall.
	//
	// **Leer ist erlaubt und heißt: nicht zu haben.** Ohne Netz bleibt der
	// Name eines Drucks aus (§16, Regel 5), und eine Anwendung, die dann keine
	// Deckliste mehr zeigt, ist falsch gebaut. Wer sie liest, prüft auf leer
	// und zeigt sonst die Kennung.
	Bezeichnung string

	// Manakosten sind die Kosten der Karte — für die Kurve eines Decks.
	//
	// **Auf demselben Weg wie die Bezeichnung** und aus demselben Grund: ein
	// Fenster, das sie selbst holte, bräuchte einen Ruf je Karte und ließe die
	// übrigen Flächen ohne (§16).
	//
	// **Leer heißt: nicht ausgefüllt.** `Vanguard` und `Archenemy` führen das
	// Feld und lassen es leer; eine Karte, die wirklich nichts kostet, trägt
	// eine `0`. Wer daraus eine Kurve rechnet, lässt die leeren weg — sonst
	// steht ein Balken bei null über zwölf Karten, der rechnerisch stimmt und
	// nichts sagt.
	Manakosten string

	// Typzeile ist die Zeile mit Ober- und Untertypen — Werkstück U10 von
	// Phase 5.1 (#1421).
	//
	// **Auf demselben Weg wie die Bezeichnung**: bei einer eigenen Karte aus
	// dem Feld `type`, bei einem Druck aus derselben Scryfall-Antwort, die
	// schon den Namen bringt. Ohne sie zählt die Auswertung keine Typen und
	// erkennt kein Land.
	//
	// **Leer heißt: nicht zu haben.** MSE rechnet `type` aus `super type` und
	// `sub type` und schreibt es mit `save value: false` (`docs/FORMATE.md`);
	// steht dort nichts, bleibt die Zeile leer — und das ist die Antwort, kein
	// Fehler.
	Typzeile string

	// Manawert ist die Zahl, die Scryfall als `cmc` führt (U10, #1421).
	//
	// **Nachgelesen und nicht gerechnet** (§15). Eine zweiseitige Karte hat
	// bei Scryfall **leere** Manakosten und trotzdem einen Manawert; wer ihn
	// aus den Kosten rechnete, bekäme dort eine Null und eine Kurve, die lügt.
	// Für eine eigene Karte gibt es keinen — dort rechnet das Fenster aus den
	// Kosten, und der Grund steht bei [Deckplatz.Manakosten].
	Manawert int
}

// Validate prüft einen Platz.
func (p Deckplatz) Validate() []*FieldError {
	var befunde []*FieldError

	eigene := p.KarteID != 0
	alias := strings.TrimSpace(p.Printing) != ""
	echter := alias || p.DruckID != 0

	switch {
	case eigene && echter:
		befunde = append(befunde, NewFieldError("printing", SchluesselValidierung,
			"Ein Platz nennt eine eigene Karte **oder** einen Druck, nicht beides."))

	case !eigene && !echter:
		befunde = append(befunde, NewFieldError("karteID", SchluesselPflichtfeld,
			"Ein Platz nennt eine eigene Karte oder einen Druck."))

	case alias && !druckkennung.MatchString(strings.TrimSpace(p.Printing)):
		befunde = append(befunde, NewFieldError("printing", SchluesselValidierung,
			"%q ist keine Druckkennung.", p.Printing))
	}

	if p.Menge < 1 {
		befunde = append(befunde, NewFieldError("menge", SchluesselValidierung,
			"Die Menge ist %d — mindestens eines muss es sein.", p.Menge))
	}

	if !IstDeckbereich(p.Bereich) {
		befunde = append(befunde, NewFieldError("bereich", SchluesselValidierung,
			"Den Bereich %q gibt es nicht.", string(p.Bereich)))
	}

	return befunde
}

// Deck ist eine Zusammenstellung von Karten.
type Deck struct {
	// ID ist die Kennung im Bestand — 0 für eines, das es dort noch nicht gibt.
	ID int64

	// Fassung ist die Bedingung einer Änderung **am Deckdatensatz** (§7,
	// Regel 2), wie am Exemplar: Name, Notiz und das gewählte Format.
	//
	// # Was sie nicht deckt, und warum das richtig ist
	//
	// **Die Plätze.** Sie liegen in einer eigenen Tabelle mit eigener
	// Eindeutigkeit, und `PlatzSetzen` schreibt sie in **einem Zug**: derselbe
	// Bezug im selben Bereich wird eine höhere Menge, kein Konflikt. Wer eine
	// Karte hineinlegt, meint nicht „ich habe den Stand von vorhin gelesen",
	// sondern „vier Wald sollen drin sein".
	//
	// **Die Wanderung 0028 sagt „steigt mit jedem Schreiben", und das ist zu
	// breit** — angestrichen von Befund 4 des Zweitblicks über Phase 4.6. Sie
	// ist angewandt und damit unveränderlich; die genaue Zusage steht hier.
	Fassung int64

	Name string

	// Spiel ist `magic`, `space`, `vs` — der Name eines Spielpakets.
	//
	// **Ohne Spiel gäbe es kein Format**: ein Format ist immer das eines
	// Spiels, und ein Deck aus Karten zweier Spiele wäre keines.
	Spiel string

	// Notiz ist, was jemand sich dazu aufschreibt.
	Notiz string

	// FormatID ist das gewählte Format — 0 heißt: keines gewählt.
	//
	// **Ohne Format ist ein Deck nicht falsch, sondern ungeprüft.** Das ist
	// dieselbe Haltung wie bei einer eigenen Karte, die Scryfall nicht kennt
	// ([0061](../../docs/entscheidungen/0061-was-ein-deck-hier-ist.md)).
	FormatID int64

	// Plaetze sind die Einträge.
	//
	// **Sie stehen hier und nicht daneben**: ein Deck ohne seine Plätze ist
	// eine Überschrift.
	Plaetze []Deckplatz

	// Herkunft sagt, woher dieses Deck kam — Werkstück 12 von Phase 4.12
	// (#1351).
	//
	// **Die leere heißt: hier entstanden.** Sie reist mit dem Deck und wird
	// nicht einzeln nachgefragt — sonst zeigte jede Fläche, die es vergisst,
	// ein übernommenes Deck als eigenes.
	//
	// **Nur beim einzelnen Deck und nicht in der Liste**: die Übersicht
	// beantwortet „welche Decks gibt es", nicht „woher kommt dieses".
	Herkunft Herkunft

	// Hauptdeck und Sideboard sind die Zahl der **Karten** je Bereich.
	//
	// # Warum sie neben den Plätzen stehen und nicht aus ihnen folgen
	//
	// **Weil die Übersicht keine Plätze lädt.** Für zwanzig Decks wären das
	// zwanzig Abfragen für etwas, das niemand liest — die **Zahlen** liest
	// aber jeder. Der Speicher zählt sie in derselben Abfrage mit.
	//
	// **Gesehen bei der Sichtprüfung am 2026-09-01**: ein Deck mit vier
	// Karten stand als „0 im Hauptdeck" da, weil [Deck.Zaehle] über eine
	// leere Liste rechnete. Kein Test hatte die Übersicht **nach** dem Legen
	// gelesen.
	//
	// **Wer die Plätze hat, rechnet sie aus** — [Deck.Zaehlen] tut es und
	// hält beide Wege an einer Stelle zusammen.
	Hauptdeck int
	Sideboard int
}

// Normalisieren glättet, was aus einer Eingabe kommt.
//
// **Getrennt von [Deck.Validate]**, wie am Exemplar: eine Prüfung, die
// nebenbei ändert, lässt den Aufrufer im Unklaren, was gespeichert wurde.
func (d *Deck) Normalisieren() {
	d.Name = strings.TrimSpace(d.Name)
	d.Spiel = strings.TrimSpace(d.Spiel)
	d.Notiz = strings.TrimSpace(d.Notiz)

	for i := range d.Plaetze {
		d.Plaetze[i].Printing = strings.TrimSpace(d.Plaetze[i].Printing)
	}
}

// Validate prüft das Deck selbst — **nicht** seine Plätze.
//
// **Getrennt, weil die Antworten getrennt gebraucht werden**: wer einen Platz
// hinzufügt, will wissen, ob **dieser** Platz taugt, und nicht, ob das ganze
// Deck in Ordnung ist.
func (d Deck) Validate() []*FieldError {
	var befunde []*FieldError

	if strings.TrimSpace(d.Name) == "" {
		befunde = append(befunde, NewFieldError("name", SchluesselPflichtfeld,
			"Ein Deck hat einen Namen."))
	}

	if strings.TrimSpace(d.Spiel) == "" {
		befunde = append(befunde, NewFieldError("spiel", SchluesselPflichtfeld,
			"Ein Deck gehört zu einem Spiel."))
	}

	return befunde
}

// Zaehlen füllt [Deck.Hauptdeck] und [Deck.Sideboard] aus den Plätzen.
//
// **Für den Weg, der die Plätze ohnehin hat.** Der Speicher zählt sie in der
// Übersicht in SQL; wer ein einzelnes Deck liest, rechnet sie hier — und beide
// meinen dieselbe Zahl.
func (d *Deck) Zaehlen() {
	d.Hauptdeck = d.Zaehle(BereichHauptdeck)
	d.Sideboard = d.Zaehle(BereichSideboard)
}

// Zaehle gibt die Zahl der Karten in einem Bereich — **mit** der Menge.
//
// **Nicht die Zahl der Plätze.** Vier Wald sind vier Karten und ein Platz, und
// jedes Format zählt die Karten.
func (d Deck) Zaehle(bereich Deckbereich) int {
	gezaehlt := 0

	for _, platz := range d.Plaetze {
		if platz.Bereich == bereich {
			gezaehlt += platz.Menge
		}
	}

	return gezaehlt
}

// Deckbezug nennt ein Deck, in dem eine Karte liegt.
//
// **Name und Kennung, sonst nichts** — er steht in einer Rückfrage vor dem
// Löschen (#1369),
// und dort will niemand ein ganzes Deck lesen.
type Deckbezug struct {
	ID   int64
	Name string
}

// Vorlaeufig sagt, ob der Platz einen Druck ohne Alias nennt (0086).
//
// **Eine eigene Karte ist nie vorläufig**: sie hat keinen Druck, und der
// fehlt ihr nicht.
func (p Deckplatz) Vorlaeufig() bool {
	return p.KarteID == 0 && druckVorlaeufig(p.DruckID, p.Printing)
}
