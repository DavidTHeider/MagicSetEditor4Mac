package domain

// **Was ein Recherchetreffer sagt** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Warum so viel am Treffer steht
//
// **Weil er ohne das eine Behauptung ist.** Das Issue zählt auf: „Treffer
// enthalten Konzept-/Druck-/Eigenkartenbezug, relevante Faces/Chunks,
// Quelle/Stand/Sprache, Snippet, Index-/Modell-/Suchkonfiguration, Modus und
// begründete Einschränkungen." Jede dieser Angaben beantwortet die Frage
// „woher weißt du das?" — und ohne sie ist ein Treffer nur ein Text, den
// jemand gezeigt hat.
//
// # Und warum die Konfiguration am Ergebnis und nicht am Treffer steht
//
// **Weil sie für alle gilt.** Modell, Index und Suchverfahren sind Angaben
// über den **Lauf**; sie an jeden Treffer zu hängen wäre dieselbe Zeile
// fünfzigmal.

import (
	"slices"
	"strings"
)

// Recherchetreffer ist ein Fund samt seiner Identität.
type Recherchetreffer struct {
	// Rang zählt von eins — **keine Wahrscheinlichkeit** (Issue).
	Rang int

	// Kennung benennt den Abschnitt im Index.
	Kennung string

	Quelle Retrievalquelle

	// Kennungssystem, Fachkennung und Semantik sind die Identität nach A2.
	Kennungssystem Kennungssystem
	Fachkennung    string
	Semantik       Semantik

	// EigeneKarteID ist gesetzt, wo der Treffer eine Karte aus diesem
	// Bestand ist — dann führt der Weg dorthin und nicht nach außen.
	EigeneKarteID int64

	Titel string

	// Ausschnitt ist die Textstelle, Text der ganze Abschnitt.
	Ausschnitt string
	Text       string

	Sprache     string
	Regelnummer string
	Urheber     Rulingurheber

	// Abschnitt, Von und Bis sagen, **welcher Teil** gemeint ist.
	Abschnitt int
	Von       int
	Bis       int

	Bezugsadresse string

	// Quellrevision, Quelldatum und Abgerufen sind die Herkunft — drei
	// Angaben, weil das Abrufdatum kein Gültigkeitsdatum ist.
	Quellrevision string
	Quelldatum    string
	Abgerufen     string

	// Weitere nennt die übrigen Abschnitte desselben Konzepts, die
	// **gebündelt** wurden.
	//
	// **Gebündelt und nicht weggelassen** (Issue: „Oracle-Duplikate
	// nachvollziehbar bündeln, Druckfilter dabei erhalten"): zwei Seiten
	// derselben Karte sind ein Treffer mit zwei Abschnitten und nicht zwei
	// Treffer.
	Weitere []string
}

// Rechercheergebnis ist die Antwort samt der Auskunft, wie sie zustande kam.
type Rechercheergebnis struct {
	Treffer []Recherchetreffer

	// Modus ist, womit gesucht wurde — `hybrid`, `lexikalisch` oder
	// `ohne-index`.
	Modus string

	// Leere sagt, **warum** nichts kam — [LeereKeine], wo etwas kam.
	Leere Rechercheleere

	// Grund ist der Satz dazu, für Menschen.
	Grund string

	// Modell, Indexstand und Suchkonfiguration sagen, **wogegen** gesucht
	// wurde.
	Modell            string
	Modellrevision    string
	Suchkonfiguration string

	// Indexfingerabdruck und Raumabdruck sind die zwei Abdrücke, und sie
	// beantworten **zwei verschiedene Fragen**.
	//
	// **Der Raum sagt, ob zwei Vektoren vergleichbar sind**
	// ([Vektorraum.Abdruck]); der Index sagt, ob zwei Antworten aus
	// derselben Liste stammen — er zählt Verfahren, Reranker und Fassungen
	// mit. Zwei Indizes im selben Raum sind nicht derselbe Index.
	//
	// **Eine Modellrevision ist weder das eine noch das andere**: als
	// Raumabdruck verschwiege sie vier andere Angaben (Zerteiler, Pooling,
	// Normalisierung, Abstandsmaß).
	Indexfingerabdruck string
	Raumabdruck        string

	// Ausgangsstand ist der Stand der Karte, zu der Ähnliches gesucht wurde
	// — leer bei jedem anderen Fall.
	//
	// **Er steht in der Antwort, weil ihn keine Fläche vorher kennt**: er
	// wird aus dem Bestand gerechnet. Wer die nächste Seite holt, nennt ihn
	// und bekommt eine Absage, wenn sich die Karte inzwischen geändert hat.
	Ausgangsstand string

	// Korpusstaende nennt je Quelle ihren Stand — einzeln, weil es keinen
	// gemeinsamen gibt.
	Korpusstaende []Korpusstand

	// Einschraenkungen nennt, was **nicht** getan wurde — eine Sprache, die
	// gar nicht im Index steht, ein Filter ohne Wirkung.
	Einschraenkungen []string

	// Fortsetzung ist der Zeiger auf die nächste Seite — leer heißt: das war
	// alles.
	Fortsetzung string
}

// GebuendeltNach fasst Treffer desselben Konzepts zusammen.
//
// # Warum nach der Fachkennung und nicht nach dem Titel
//
// **Weil zwei Karten denselben Namen tragen können** und dieselbe Karte in
// zwei Sprachen zwei Titel hat. Das Issue verbietet das andere ausdrücklich
// in A15: „keine Deduplizierung nach Namen".
//
// **Ohne Fachkennung wird nicht gebündelt.** Eine Regelstelle hat keine, und
// zwei Regelstellen sind zwei Treffer.
func GebuendeltNach(treffer []Recherchetreffer) []Recherchetreffer {
	gesehen := map[string]int{}
	gebuendelt := make([]Recherchetreffer, 0, len(treffer))

	for _, einzelner := range treffer {
		schluessel := string(einzelner.Kennungssystem) + ":" + einzelner.Fachkennung

		if einzelner.Fachkennung == "" {
			gebuendelt = append(gebuendelt, einzelner)

			continue
		}

		stelle, schonDa := gesehen[schluessel]
		if !schonDa {
			gesehen[schluessel] = len(gebuendelt)
			gebuendelt = append(gebuendelt, einzelner)

			continue
		}

		// **Der erste bleibt vorn.** Er stand höher im Rang, und der
		// gebündelte Treffer ist derselbe Fund mit einem Abschnitt mehr.
		gebuendelt[stelle].Weitere = append(
			gebuendelt[stelle].Weitere, einzelner.Kennung)
	}

	// **Die Ränge zählen neu durch** — sie sind die Ausgabe und nicht die
	// Eingabe.
	for stelle := range gebuendelt {
		gebuendelt[stelle].Rang = stelle + 1
	}

	return gebuendelt
}

// OhneAusgangskarte wirft die Karte heraus, zu der gesucht wurde.
//
// **Das Issue sagt es als eigene Zeile**: „Eigene Ausgangskarte aus
// Ähnlichkeitstreffern ausschließen." Sie ist sich selbst am ähnlichsten, und
// ein erster Platz für die Frage selbst ist keine Antwort.
func OhneAusgangskarte(
	treffer []Recherchetreffer, kennungen []string,
) []Recherchetreffer {
	if len(kennungen) == 0 {
		return treffer
	}

	geblieben := make([]Recherchetreffer, 0, len(treffer))

	for _, einzelner := range treffer {
		if slices.Contains(kennungen, einzelner.Kennung) ||
			gehoertZu(einzelner, kennungen) {
			continue
		}

		einzelner.Rang = len(geblieben) + 1
		geblieben = append(geblieben, einzelner)
	}

	return geblieben
}

// gehoertZu sagt, ob ein Treffer zu einer der Kennungen gehört.
//
// **Über den Kennungsstamm** — `probe:o1:0` und `probe:o1:1` sind zwei
// Abschnitte derselben Karte, und wer nur den einen ausschlösse, ließe die
// Ausgangskarte über ihre zweite Seite wieder herein.
func gehoertZu(treffer Recherchetreffer, kennungen []string) bool {
	for _, kennung := range kennungen {
		stamm := kennung
		if stelle := strings.LastIndex(kennung, ":"); stelle > 0 {
			stamm = kennung[:stelle]
		}

		if strings.HasPrefix(treffer.Kennung, stamm+":") {
			return true
		}
	}

	return false
}
