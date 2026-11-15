package domain

// **Was eine Recherche ist** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Vier Fälle und nicht einer
//
// **Weil sie vier verschiedene Fragen sind.** Das Issue zählt sie auf:
// Kartenrecherche, ähnliche Karten, Regelrecherche, Rulingsrecherche. Sie
// teilen den Index aus A15 und unterscheiden sich darin, **wonach** gefragt
// wird und **was** als Antwort zählt — eine Regelfrage will Regelstellen, und
// eine Ähnlichkeitsfrage will die Ausgangskarte **nicht**.
//
// # Der Rang ist keine Wahrscheinlichkeit
//
// **Das Issue sagt es wörtlich**, und es ist die Folge aus
// [0103](../../docs/entscheidungen/0103-wie-zwei-trefferlisten-zusammenkommen.md):
// eine Rangfusion ordnet, sie bewertet nicht. Wer eine Zahl zwischen null und
// eins zeigte, behauptete eine Genauigkeit, die es nicht gibt — und jemand
// würde sie vergleichen.

import (
	"fmt"
	"slices"
	"strings"
)

// Recherchefall sagt, welche Frage gestellt wird.
type Recherchefall string

// Die vier Fälle.
const (
	// FallKarten sucht Karten über ihren Text.
	FallKarten Recherchefall = "karten"

	// FallAehnlich sucht Karten, die einer anderen ähneln.
	FallAehnlich Recherchefall = "aehnlich"

	// FallRegeln sucht Stellen im Regelwerk.
	FallRegeln Recherchefall = "regeln"

	// FallRulings sucht Einzelfallauskünfte.
	FallRulings Recherchefall = "rulings"

	// FallRegelhinweis sucht **alles**, worauf sich eine Regelaussage
	// stützen darf — Werkstück A13.2a von Phase 5.3
	// (#1452).
	//
	// # Warum ein fünfter Fall und nicht drei Rufe
	//
	// **Weil drei Rufe drei Trefferlisten wären und niemand sie fusionieren
	// dürfte.** Das Issue verbietet eine zweite Retrievalpipeline
	// ausdrücklich; die Rangfusion aus
	// [0103](../../docs/entscheidungen/0103-wie-zwei-trefferlisten-zusammenkommen.md)
	// arbeitet **innerhalb** eines Rufs über die Quellen dieses Falls.
	//
	// # Und warum er nicht `FallRegeln` ersetzt
	//
	// **Weil eine Regelfrage etwas anderes ist als eine Regelaussage über
	// eine Karte.** Wer wissen will, was 509.1 sagt, will keine Rulings und
	// keine Kartentexte dazwischen; wer wissen will, wie **diese** Karte
	// wirkt, braucht alle drei.
	FallRegelhinweis Recherchefall = "regelhinweis"
)

var recherchefaelle = []Recherchefall{
	FallKarten, FallAehnlich, FallRegeln, FallRulings, FallRegelhinweis,
}

// Recherchefaelle gibt sie in ihrer Reihenfolge.
func Recherchefaelle() []Recherchefall { return slices.Clone(recherchefaelle) }

// Quellen sagt, in welchen Korpora dieser Fall sucht.
//
// **Sie stehen am Fall und nicht beim Aufrufer.** Eine Regelfrage, die in
// Kartentexten sucht, findet Karten, die das Wort enthalten — und das ist
// eine andere Frage als die gestellte.
func (f Recherchefall) Quellen() []Retrievalquelle {
	switch f {
	case FallRegeln:
		return []Retrievalquelle{QuelleRegelwerk}

	case FallRulings:
		return []Retrievalquelle{QuelleRuling}

	// **Alle drei, und die Reihenfolge ist keine Rangfolge**: welcher Beleg
	// oben steht, entscheidet die Fusion und nicht diese Zeile.
	case FallRegelhinweis:
		return []Retrievalquelle{QuelleRegelwerk, QuelleRuling, QuelleOracle}

	case FallKarten, FallAehnlich:
		return []Retrievalquelle{QuelleOracle, QuelleEigeneKarte}

	default:
		return nil
	}
}

// Rechercheleere sagt, **warum** nichts gefunden wurde.
//
// # Warum das ein eigener Typ ist
//
// **Weil „nichts gefunden" fünf verschiedene Dinge heißen kann**, und das
// Issue verlangt, sie zu unterscheiden: „Leere Treffer unterscheiden von
// Index fehlt, Modell fehlt, unzureichender Quelle, gesperrtem Zugriff und
// Budgetgrenze." Vier davon sagen jemandem, was zu tun ist; eine sagt, dass
// es die Antwort nicht gibt.
type Rechercheleere string

// Die sechs Gründe.
const (
	// LeereKeine heißt: es wurde etwas gefunden.
	LeereKeine Rechercheleere = ""

	// LeereNichtsGefunden heißt: gesucht wurde vollständig, es gibt nichts.
	LeereNichtsGefunden Rechercheleere = "nichts-gefunden"

	// LeereOhneIndex heißt: es ist kein Suchindex aufgebaut.
	LeereOhneIndex Rechercheleere = "ohne-index"

	// LeereOhneModell heißt: der Index steht, aber sein Modell fehlt — dann
	// wurde nur lexikalisch gesucht.
	LeereOhneModell Rechercheleere = "ohne-modell"

	// LeereOhneQuelle heißt: der gefragte Korpus ist nicht im Index.
	LeereOhneQuelle Rechercheleere = "ohne-quelle"

	// LeereGesperrt heißt: es gäbe etwas, aber nicht für diesen Aufrufer.
	LeereGesperrt Rechercheleere = "gesperrt"

	// LeereBudget heißt: die Grenze war vorher erreicht.
	LeereBudget Rechercheleere = "budget"
)

// Rechercheleeren gibt alle außer [LeereKeine].
func Rechercheleeren() []Rechercheleere {
	return []Rechercheleere{
		LeereNichtsGefunden, LeereOhneIndex, LeereOhneModell,
		LeereOhneQuelle, LeereGesperrt, LeereBudget,
	}
}

// Recherchehoechstens ist die größte Trefferzahl, die eine Anfrage verlangen
// darf.
//
// **Fünfzig**, und die Zahl ist gewählt: darüber liest niemand mehr, und ein
// Kontextpaket aus fünfzig Abschnitten sprengt jedes Tokenbudget. Wer mehr
// will, blättert.
const Recherchehoechstens = 50

// Rechercheanfrage ist eine Frage an den Wissensbestand.
type Rechercheanfrage struct {
	Fall Recherchefall

	// Frage ist der Text — leer bei [FallAehnlich], dort zählt die
	// Ausgangskarte.
	Frage string

	// Ausgangskarte ist die Karte, zu der Ähnliches gesucht wird.
	//
	// **Mit ihrer Revision**, und das ist die Zusage aus dem Issue: „klar
	// autorisierte Ausgangskarte **samt Revision**". Wer sie ohne schickte,
	// suchte zu einem Stand, den es vielleicht nicht mehr gibt.
	AusgangskarteID int64
	Ausgangsstand   string

	// Sprache schränkt auf eine ein — leer heißt: jede.
	Sprache string

	// Grenze ist die gewünschte Trefferzahl.
	Grenze int

	// Korpusstand verlangt einen bestimmten Stand — leer heißt: der aktive.
	Korpusstand string

	// Freigegeben sind die **privaten** Quellen, die ein Mensch ausdrücklich
	// ausgewählt hat — Werkstück A13.4 von Phase 5.3
	// (#1452).
	//
	// **Leer heißt: keine.** Der eigene Bestand bleibt dann außen, und die
	// Antwort sagt es ([Rechercheanfrage.AusgelasseneQuellen]).
	Freigegeben []Retrievalquelle
}

// Pruefe sagt, ob die Anfrage beantwortbar ist.
func (a Rechercheanfrage) Pruefe() error {
	if !slices.Contains(recherchefaelle, a.Fall) {
		return fmt.Errorf("%w: %q ist kein Recherchefall", ErrValidation, a.Fall)
	}

	if fehler := a.pruefeFreigabe(); fehler != nil {
		return fehler
	}

	if a.Fall == FallAehnlich {
		return a.pruefeAusgang()
	}

	if strings.TrimSpace(a.Frage) == "" {
		return fmt.Errorf(
			"%w: eine Recherche nach %q braucht eine Frage", ErrValidation, a.Fall)
	}

	return a.pruefeGrenze()
}

// pruefeAusgang hält, was eine Ähnlichkeitsfrage braucht.
func (a Rechercheanfrage) pruefeAusgang() error {
	if a.AusgangskarteID == 0 {
		return fmt.Errorf(
			"%w: eine Ähnlichkeitsfrage braucht ihre Ausgangskarte",
			ErrValidation)
	}

	// **Ein leerer Stand ist kein fehlender.**
	//
	// # Warum hier bis zum 2026-09-12 eine Pflicht stand
	//
	// **Weil „eine Karte, die sich geändert hat, ist eine andere Frage"
	// richtig ist** — nur konnte ihn **keine einzige Fläche nennen**: den
	// gerechneten Stand einer Karte ([CardRow.Stand]) rechnet der Bestand,
	// und weder Fenster noch Kommandozeile noch MCP bekommen ihn zu sehen.
	// Die Pflicht hätte die Ähnlichkeitsfrage an allen vier Flächen
	// unbenutzbar gemacht.
	//
	// **Deshalb heißt leer jetzt „der, auf dem sie **jetzt** steht"** — und
	// [Rechercheergebnis.Ausgangsstand] nennt ihn. Damit bindet auch der
	// Fortsetzungszeiger einen wirklichen Stand: wer die zweite Seite holt,
	// nennt ihn, und eine Karte, die sich dazwischen geändert hat, gibt eine
	// Absage statt einer still gemischten Seite.
	//
	// **Wer ihn nennt, wird geprüft.** Ein Stand, den die Karte nicht mehr
	// hat, ist ein Konflikt — das ist die Zusage, um die es ging.
	return a.pruefeGrenze()
}

// pruefeGrenze hält die Trefferzahl.
func (a Rechercheanfrage) pruefeGrenze() error {
	if a.Grenze <= 0 || a.Grenze > Recherchehoechstens {
		return fmt.Errorf(
			"%w: eine Recherche gibt zwischen einem und %d Treffern, "+
				"verlangt sind %d", ErrValidation, Recherchehoechstens, a.Grenze)
	}

	return nil
}
