package domain

// **Der Vergleich vor der Rückkehr** — Werkstück O7 von Phase 5.2
// (#1431).
//
// # Zwei Fragen, und sie zeigen in verschiedene Richtungen
//
//  1. **Was seit der Sicherung dazukam, verschwindet.** Das ist der
//     Datenverlust, und die Abnahme verlangt ihn beziffert: „Dialog nennt
//     genau den wiederhergestellten Datenzeitpunkt und noch nicht übertragene
//     Änderungen."
//  2. **Was die Sicherung mehr hat als der Bestand, steht danach wieder da** —
//     und niemand weiß, ob es noch jemandem gehört. Das ist die Epoche.
//
// Eine Fläche, die nur die erste zeigt, verschweigt die teurere: verlorene
// Karten kann man neu eintragen, eine zweimal verkaufte nicht.

import (
	"sort"
	"time"
)

// Die Gegenstände, die gezählt werden — **Marken und keine Sätze**.
const (
	GegenstandSets      = "sets"
	GegenstandKarten    = "karten"
	GegenstandDecks     = "decks"
	GegenstandExemplare = "exemplare"
)

// Bestandsstand ist, was ein Bestand über seine Größe hergibt.
//
// **Zwei Seiten desselben Typs.** Der Vergleich stellt die Sicherung neben den
// laufenden Bestand; wären es zwei Typen, hätte die Frage „wer ist wer?" zwei
// Antworten.
type Bestandsstand struct {
	// AngelegtAm ist der Zeitpunkt, den dieser Stand meint.
	AngelegtAm time.Time

	// **−1 heißt „diese Seite sagt nichts darüber"** — eine sehr alte
	// Sicherung kennt eine Tabelle nicht, und das ist etwas anderes als null.
	Sets      int
	Karten    int
	Decks     int
	Exemplare int

	// Sammlung ist der Besitz je Gegenstand und Profil.
	Sammlung []Sammlungsmenge

	// SammlungGelesen sagt, ob [Bestandsstand.Sammlung] eine Auskunft ist.
	//
	// **Eine leere Sammlung und eine ungelesene sehen gleich aus** — und sie
	// bedeuten das Gegenteil: die erste heißt „nichts unklar", die zweite
	// „unbekannt, ob etwas unklar ist". Eine sehr alte Sicherung kennt die
	// Spalten nicht, die für den Vergleich nötig sind; sie deswegen als
	// unbedenklich zu führen wäre die gefährlichste Richtung.
	SammlungGelesen bool
}

// Ruecksprunglage ist das Ergebnis des Vergleichs.
type Ruecksprunglage struct {
	// Datenstand ist der Zeitpunkt, auf den zurückgegangen würde.
	Datenstand time.Time `json:"datenstand"`

	// Zeilen sind die Zählstände nebeneinander.
	Zeilen []Vergleichszeile `json:"zeilen"`

	// Unklar sind die Mengen, die nach der Rückkehr wieder dastünden, ohne
	// dass jemand weiß, ob sie noch da sein sollten.
	Unklar []Sammlungsmenge `json:"unklar"`

	// HandpruefungNoetig sagt, dass eine der beiden Seiten ihre Sammlung nicht
	// hergibt — dann ist **keine** Aussage über unklare Mengen zu treffen.
	//
	// **Das Issue verlangt genau das**: „Bei fehlenden Nachweisen Unsicherheit
	// sichtbar und manuelle Bestandsprüfung erforderlich." Eine leere Liste
	// hieße hier „nichts unklar" und wäre die eine Auskunft, die niemals
	// falsch sein darf.
	HandpruefungNoetig bool `json:"handpruefungNoetig"`
}

// Ruecksprungvergleich stellt eine Sicherung neben den laufenden Bestand.
func Ruecksprungvergleich(sicherung, jetzt Bestandsstand) Ruecksprunglage {
	return Ruecksprunglage{
		Datenstand: sicherung.AngelegtAm,
		Zeilen: []Vergleichszeile{
			{Gegenstand: GegenstandSets, InSicherung: sicherung.Sets, Jetzt: jetzt.Sets},
			{Gegenstand: GegenstandKarten, InSicherung: sicherung.Karten, Jetzt: jetzt.Karten},
			{Gegenstand: GegenstandDecks, InSicherung: sicherung.Decks, Jetzt: jetzt.Decks},
			{Gegenstand: GegenstandExemplare, InSicherung: sicherung.Exemplare, Jetzt: jetzt.Exemplare},
		},
		Unklar:             unklareMengen(sicherung.Sammlung, jetzt.Sammlung),
		HandpruefungNoetig: !sicherung.SammlungGelesen || !jetzt.SammlungGelesen,
	}
}

// unklareMengen gibt, was die Sicherung **mehr** hat als der Bestand.
func unklareMengen(sicherung, jetzt []Sammlungsmenge) []Sammlungsmenge {
	besitz := make(map[mengenkennung]int, len(jetzt))
	for _, menge := range jetzt {
		besitz[menge.kennung()] += menge.Menge
	}

	unklar := []Sammlungsmenge{}

	for _, menge := range sicherung {
		mehr := menge.Menge - besitz[menge.kennung()]
		if mehr <= 0 {
			continue
		}

		// **Der lesbare Bezug kommt aus der Sicherung**, denn sie ist die
		// Seite, die danach dasteht.
		betroffen := menge
		betroffen.Menge = mehr

		unklar = append(unklar, betroffen)
	}

	// **Eine feste Ordnung**, damit die Oberfläche nicht bei jedem Aufruf eine
	// andere Reihe zeigt.
	sort.Slice(unklar, func(links, rechts int) bool {
		return kleiner(unklar[links], unklar[rechts])
	})

	return unklar
}

// kleiner ordnet zwei Mengen — Profil, Druck, Zustand, Sprache, Ausführung.
func kleiner(links, rechts Sammlungsmenge) bool {
	if links.Profil != rechts.Profil {
		return links.Profil < rechts.Profil
	}

	if links.Schluessel != rechts.Schluessel {
		return links.Schluessel < rechts.Schluessel
	}

	if links.Zustand != rechts.Zustand {
		return links.Zustand < rechts.Zustand
	}

	if links.Sprache != rechts.Sprache {
		return links.Sprache < rechts.Sprache
	}

	// **Nach der Ausführung, und zwar alphabetisch**: `etched`, `foil`,
	// `nonfoil`. Ein Vergleich über `!links && rechts` ging nur, solange es
	// zwei gab.
	//
	// **Die Zusage ist die Bestimmtheit und nicht die Reihenfolge** — hier
	// entscheidet sich nur, in welcher Folge zwei Zeilen stehen, die in allem
	// anderen gleich sind. Dass es **immer dieselbe** ist, ist der Punkt: ohne
	// diesen Vergleich stünde in der Oberfläche mal die eine vorn und mal die
	// andere, je nachdem, wie der Bestand sie hergab.
	return links.Ausfuehrung() < rechts.Ausfuehrung()
}
