package domain

// **Ein Angebot ist die Aussage eines fremden Verkäufers** — Werkstück A4 von
// Phase 5.3 (#1443).
//
// # Warum es kein Exemplar ist
//
// Ein [Exemplar] ist Besitz: es liegt in einer Schachtel, und die Sammlung
// beantwortet „was habe ich". Ein Angebot ist eine **Behauptung über fremden
// Besitz** — jemand sagt, er habe etwas und gebe es zu einem Preis ab. Wer
// beides in eine Tabelle legte, könnte „meine Karten" nicht mehr beantworten
// ([0095](../../docs/entscheidungen/0095-was-cardtrader-sagt-und-was-davon-gilt.md)).
//
// # Und warum es kein Preisstand ist
//
// Ein [Preisstand] ist eine **Reihe über die Zeit**: dieselbe Frage, immer
// wieder gestellt. Ein Angebot ist ein einzelner Posten, der morgen weg sein
// kann, mit einer Menge, einem Verkäufer und einem Land. Aus vielen Angeboten
// **wird** ein Preisstand — das ist [Preisstand], und der Weg dahin steht in
// `service`.

import "strings"

// Marktangebot ist ein Posten auf einem fremden Marktplatz.
type Marktangebot struct {
	// Quelle ist der Marktplatz.
	Quelle Preisquelle

	// Fremdkennung ist die Nummer **dieses Postens** beim Anbieter — nicht
	// die der Kartenfassung.
	Fremdkennung string

	// Druck ist der Scryfall-Alias des Drucks, zu dem der Posten gehört.
	Druck string

	// Preis ist der Betrag in der kleinsten Einheit der Währung.
	Preis    int64
	Waehrung string

	// Menge ist, wie viele der Verkäufer hat.
	Menge int

	// Buendel ist die Zahl der Karten, die zusammen abgegeben werden — 1 heißt
	// einzeln, und 0 heißt: der Anbieter hat nichts gesagt.
	Buendel int

	Finish Finish

	// Zustand ist die Erhaltung **in der Ausdrucksweise des Anbieters**.
	//
	// **Nicht [Zustand]**, und das ist kein Versehen: CardTrader kennt sechs
	// Stufen, dieses Programm sieben, und die Skalen sind nicht ineinander
	// überführbar (0095). Eine erfundene Gleichsetzung wäre nicht mehr zu
	// entfernen, sobald ein Verlauf darauf steht.
	Zustand string

	// Sprache ist die Sprachfassung, so wie der Anbieter sie nennt.
	Sprache string

	Signiert   bool
	Veraendert bool
	Gradiert   bool

	// ImUrlaub sagt, dass der Verkäufer gerade nicht versendet.
	//
	// **Es wird gezeigt und nicht weggefiltert**: ein Posten, der da ist und
	// nicht kommt, ist eine andere Auskunft als keiner.
	ImUrlaub bool

	Verkaeufer string
	Land       string
}

// Angebotswunsch engt ein, wonach gefragt wird.
type Angebotswunsch struct {
	// Sprache ist der Code, den der Anbieter erwartet — leer heißt: alle.
	Sprache string

	// Finish engt auf eine Ausführung ein — leer heißt: alle.
	Finish Finish

	// Zustaende sind die Erhaltungen, die durchgelassen werden — leer heißt:
	// alle.
	//
	// **Hier wird nach dem Abruf gefiltert und nicht davor**: CardTrader
	// nimmt keinen Zustandsfilter in der Abfrage (0095), und ein Filter, der
	// so täte, verspräche eine Ersparnis, die es nicht gibt.
	Zustaende []string
}

// Passt sagt, ob ein Angebot dem Wunsch entspricht.
//
// **Sprache und Ausführung stehen hier trotzdem**, obwohl sie schon in der
// Abfrage stehen: ein Anbieter, der einen Filter übergeht, schickt sonst
// Posten, nach denen niemand gefragt hat — und niemand sähe es.
func (w Angebotswunsch) Passt(angebot Marktangebot) bool {
	if w.Sprache != "" && !strings.EqualFold(w.Sprache, angebot.Sprache) {
		return false
	}

	if w.Finish != "" && w.Finish != angebot.Finish {
		return false
	}

	if len(w.Zustaende) == 0 {
		return true
	}

	for _, zustand := range w.Zustaende {
		if strings.EqualFold(zustand, angebot.Zustand) {
			return true
		}
	}

	return false
}

// Fremdfassung ist eine Kartenfassung, wie ein fremder Marktplatz sie führt.
//
// **Sie ist die Brücke, nicht das Ziel**: an ihr hängen die Kennungen, über
// die sich ein Druck dieses Bestands wiederfindet
// ([0086](../../docs/entscheidungen/0086-sieben-begriffe-fuer-eine-karte.md)).
type Fremdfassung struct {
	// Kennung ist die Nummer der Fassung beim Anbieter — als Text, weil sie
	// über die Naht geht (0086: keine int64 ins JavaScript).
	Kennung string

	Name string

	// Kennungen sind die Fremdkennungen dieser Fassung.
	Kennungen []Druckkennung
}

// Nennt sagt, ob diese Fassung die genannte Kennung führt.
func (f Fremdfassung) Nennt(gesucht Druckkennung) bool {
	gesucht.Normalisieren()

	for _, kennung := range f.Kennungen {
		kennung.Normalisieren()

		if kennung == gesucht {
			return true
		}
	}

	return false
}
