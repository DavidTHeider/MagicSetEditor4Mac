package domain

// **Was eine Preisreihe ausmacht** — Werkstück A3 von Phase 5.3
// (#1442).
//
// # Warum es diese drei Angaben gibt
//
// Bis hierher war Scryfall die einzige Quelle, und eine Reihe war „Quelle,
// Währung, Ausführung". Das reichte, weil Scryfall je Karte **einen** Preis je
// Währung nennt und sonst nichts.
//
// MTGJSON führt dieselbe Karte bei **vier** Anbietern (`cardkingdom`,
// `cardmarket`, `cardsphere`, `tcgplayer`), auf **zwei** Märkten (`paper` und
// `mtgo`) und in **zwei** Richtungen: was ein Händler verlangt (`retail`) und
// was er zahlt (`buylist`). Nachgelesen am 2026-09-11 unter
// <https://mtgjson.com/data-models/price/price-list/> und
// <https://mtgjson.com/data-models/price/price-points/>.
//
// **Wer das in einer Reihe zusammenlegte, verglich einen Einkaufspreis mit
// einem Verkaufspreis** — und niemand sähe es, weil beide Zahlen aussehen wie
// Preise.

// Preismarkt sagt, auf welchem Markt ein Preis gilt.
//
// **Die Werte sind MTGJSONs** (§9): `paper` und `mtgo`. Eine deutsche
// Entsprechung in der Tafel machte jede Zeile, die von dort kommt, unlesbar für
// den, der beide Seiten vergleicht.
type Preismarkt string

const (
	// MarktPapier ist die gedruckte Karte.
	MarktPapier Preismarkt = "paper"

	// MarktOnline ist Magic Online — eigene Ware, eigene Währung.
	//
	// **Und deshalb nie mit Papier zu verrechnen**: Scryfalls `tix` sind
	// Tickets, keine Währung eines Landes ([Preisstand.Waehrung]).
	MarktOnline Preismarkt = "mtgo"
)

// Preisart sagt, in welche Richtung ein Preis gilt.
type Preisart string

const (
	// PreisartVerkauf ist, was der Anbieter verlangt — MTGJSONs `retail`.
	PreisartVerkauf Preisart = "retail"

	// PreisartAnkauf ist, was der Anbieter zahlt — MTGJSONs `buylist`.
	//
	// **Immer niedriger als der Verkaufspreis**, und zwar deutlich. Eine
	// Sammlungsbewertung, die beides mischt, ist keine.
	PreisartAnkauf Preisart = "buylist"
)

// Preislage sagt, was über einen Preis feststeht.
//
// # Drei Zustände, nicht zwei
//
// Bis hierher gab es `Weggefallen bool`, und das war eine Antwort auf zwei
// Fragen zugleich. Eine **Teilantwort** — eine Quelle, die den Papiermarkt
// nennt und über Magic Online schweigt — ist etwas anderes als ein Widerruf:
// über die ungenannte Reihe steht **nichts** fest, und sie als weggefallen zu
// führen wäre eine Aussage, die niemand gemacht hat (#1442).
type Preislage string

const (
	// PreisGenannt — die Quelle nennt einen Wert, er steht in [Preisstand.Wert].
	PreisGenannt Preislage = "genannt"

	// PreisWeggefallen — die Quelle hat vollständig geantwortet und nennt
	// hier **nichts**. Das ist eine Angabe und keine Lücke.
	PreisWeggefallen Preislage = "weggefallen"

	// PreisUngeprueft — die Antwort deckte diese Reihe nicht ab.
	//
	// **Der Unterschied zu [PreisWeggefallen]**: dort hat jemand nachgesehen
	// und nichts gefunden, hier hat niemand nachgesehen. Der letzte bekannte
	// Wert bleibt damit gültig, und er sagt, wie alt er ist (§16, Satz 5).
	PreisUngeprueft Preislage = "ungeprueft"
)

var (
	preismaerkte = []Preismarkt{MarktPapier, MarktOnline}
	preisarten   = []Preisart{PreisartVerkauf, PreisartAnkauf}
	preislagen   = []Preislage{PreisGenannt, PreisWeggefallen, PreisUngeprueft}
)

// IstPreismarkt sagt, ob eine Angabe zur Aufzählung gehört.
func IstPreismarkt(markt Preismarkt) bool { return kennt(preismaerkte, markt) }

// IstPreisart sagt, ob eine Angabe zur Aufzählung gehört.
func IstPreisart(art Preisart) bool { return kennt(preisarten, art) }

// IstPreislage sagt, ob eine Angabe zur Aufzählung gehört.
func IstPreislage(lage Preislage) bool { return kennt(preislagen, lage) }
