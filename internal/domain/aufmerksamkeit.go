package domain

// **Aufmerksamkeit aus echten Befunden** — Werkstück U11 von Phase 5.1
// (#1422).
//
// Ein **Hinweis** ist ein Befund an einem Arbeitsobjekt, den jemand
// ansehen sollte: ein Set mit Vorbehalten, ein Deck, dem Karten fehlen, ein
// Paket mit einem Skriptbefund. **Kein Zähler ohne Quelle**: jeder Hinweis
// nennt sein Objekt, seine Zahl und den Grund, und die Übersicht sagt dazu,
// wie viele Objekte sie durchgesehen hat.

import "sort"

// Hinweisart ist, was für ein Befund ein Hinweis ist.
type Hinweisart string

// Die drei Arten — in der Reihenfolge, in der die Übersicht sie zeigt.
const (
	HinweisSetVorbehalte Hinweisart = "setVorbehalte"
	HinweisDeckFehlend   Hinweisart = "deckFehlend"
	HinweisPaketBefund   Hinweisart = "paketBefund"
)

var hinweisreihe = map[Hinweisart]int{
	HinweisSetVorbehalte: 0, HinweisDeckFehlend: 1, HinweisPaketBefund: 2,
}

// Hinweis ist ein Befund an einem Objekt.
type Hinweis struct {
	Art    Hinweisart
	Objekt Arbeitsobjekt
	// Anzahl ist, wie viele — Vorbehalte, fehlende Karten, Befunde.
	Anzahl int
	// Grund ist der Satz dazu — der erste Befund, die erste fehlende Karte.
	//
	// **Deutsch, und der Rückfall** (§9): das Fenster setzt seinen eigenen
	// Satz ein, wenn [Hinweis.Schluessel] einen nennt, den es kennt.
	Grund string

	// Ort ist der Bezeichner, vor dem der Satz steht — die Datei eines
	// Skriptbefunds. Er wird nicht übersetzt; leer heißt: keiner.
	Ort string

	// Schluessel nennt den Satz des Grunds im Katalog des Fensters
	// (#1963) —
	// [GrundVorbehalt] oder der Schlüssel eines Skriptbefunds. Leer heißt:
	// der Grund ist kein Satz, sondern ein Name, etwa der einer Karte.
	Schluessel string

	// Werte füllen die Platzhalter `{0}`, `{1}` … des Satzes.
	Werte []string
}

// GrundVorbehalt ist der Schlüssel des Grunds an einem Set: Datei und Zeile
// des ersten Vorbehalts als `{0}` und `{1}`. **Der Satz steht im Katalog des
// Fensters** unter `seiten.start.aufmerksamkeit.gruende`, und
// `internal/werkzeuge/aufmerksamkeitsgruende_test.go` hält beide Seiten
// zusammen.
const GrundVorbehalt = "vorbehalt"

// Durchgesehen sagt, worüber die Übersicht geschaut hat.
type Durchgesehen struct {
	Sets, Decks, Pakete int
}

// Aufmerksamkeit ist die Übersicht der Hinweise.
type Aufmerksamkeit struct {
	Hinweise     []Hinweis
	Durchgesehen Durchgesehen
}

// OrdneHinweise sortiert: nach Art, dann die größte Zahl zuerst, dann der Name.
//
// **Und leert nichts heimlich**: ein Hinweis ohne Anzahl bleibt draußen — ein
// Set mit null Vorbehalten ist kein Hinweis.
func OrdneHinweise(hinweise []Hinweis) []Hinweis {
	geordnet := make([]Hinweis, 0, len(hinweise))

	for _, eines := range hinweise {
		if eines.Anzahl > 0 {
			geordnet = append(geordnet, eines)
		}
	}

	sort.SliceStable(geordnet, func(i, j int) bool {
		a, b := geordnet[i], geordnet[j]
		if hinweisreihe[a.Art] != hinweisreihe[b.Art] {
			return hinweisreihe[a.Art] < hinweisreihe[b.Art]
		}

		if a.Anzahl != b.Anzahl {
			return a.Anzahl > b.Anzahl
		}

		return a.Objekt.Name < b.Objekt.Name
	})

	return geordnet
}
