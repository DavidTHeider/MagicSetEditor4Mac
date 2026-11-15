package domain

import (
	"fmt"
	"strings"
	"time"
)

// **Ein Preis ist eine Beobachtung über einen Druck, kein Feld an ihm**
// ([0052](../../docs/entscheidungen/0052-die-sammlung-ist-nicht-das-set.md)).
//
// Drei Gründe stehen dort: ein Preis altert, ein Preis hat eine Herkunft, und
// aus Momentaufnahmen wird ein Verlauf.
//
// # Was die Quelle über ihre Preise sagt
//
// Nachgelesen am 2026-08-27 unter
// [scryfall.com/docs/api](https://scryfall.com/docs/api):
//
//   - „An object containing **daily** price information … as **strings**".
//   - „Card prices … represent daily estimates and/or market values provided by
//     our affiliates. **Absolutely no guarantee** is made for any price
//     information."
//
// Beides hat Folgen im Code: gelesen wird aus **Text** (siehe [MarktpreisAus]),
// und wo ein Preis erscheint, erscheint auch, woher er kommt und wann er
// erhoben wurde.

// Preisquelle sagt, wer einen Preis genannt hat.
type Preisquelle string

// PreisquelleScryfall ist die einzige Quelle, die es heute gibt.
//
// **Als Aufzählungswert und nicht als fester Text im Code**: die Tabelle trägt
// eine Quelle je Zeile, damit eine zweite ohne Migration danebenpasst. Gebaut
// wird sie erst, wenn jemand sie braucht — und ihre Bedingungen sind dann zu
// lesen, nicht zu vermuten (§16, Satz 1).
const PreisquelleScryfall Preisquelle = "scryfall"

// Quellentrenner steht zwischen einer Quelle und ihrem Anbieter.
//
// **MTGJSON ist ein Sammler, kein Händler** (A3, #1442): eine Antwort trägt
// die Preise von `cardkingdom`, `cardmarket`, `cardsphere`, `tcgplayer` und
// `cardhoarder`. Jeder nennt einen anderen Preis, und eine Quelle für alle
// wäre eine Reihe für fünf Aussagen — deshalb `mtgjson:cardkingdom`.
const Quellentrenner = ":"

// GehoertZu sagt, ob diese Quelle zur genannten Familie gehört.
//
// # Wofür es das braucht
//
// **Für den Widerruf.** Wenn MTGJSON für einen Druck nichts mehr nennt, ist
// **jede** seiner Reihen weggefallen — auch die von `cardkingdom`. Ein
// Vergleich auf Gleichheit fände `mtgjson` und `mtgjson:cardkingdom` ungleich
// und widerriefe nichts.
//
// **Und der Trenner gehört dazu**: ohne ihn gehörte `mtgjsonx` zur Familie
// `mtgjson`, und das ist eine andere Quelle.
func (q Preisquelle) GehoertZu(familie Preisquelle) bool {
	return q == familie ||
		strings.HasPrefix(string(q), string(familie)+Quellentrenner)
}

// Finish ist die Ausführung eines Drucks.
//
// **Der Name ist englisch, weil er einer von Scryfall ist** (§9): dort heißen
// die Werte `nonfoil`, `foil` und `etched`, und ein Preis hängt an genau
// einem davon (`usd`, `usd_foil`, `usd_etched`).
type Finish string

// Die drei Ausführungen, die Scryfall bepreist.
const (
	FinishNonfoil Finish = "nonfoil"
	FinishFoil    Finish = "foil"
	FinishEtched  Finish = "etched"
)

// alleFinishes ist die Aufzählung.
var alleFinishes = []Finish{FinishNonfoil, FinishFoil, FinishEtched}

// IstFinish sagt, ob eine Angabe zur Aufzählung gehört.
func IstFinish(finish Finish) bool {
	for _, bekannt := range alleFinishes {
		if bekannt == finish {
			return true
		}
	}

	return false
}

// Preisstand ist ein Preis zu einem Zeitpunkt.
type Preisstand struct {
	// ID ist die Kennung im Bestand — 0 für einen, den es dort noch nicht gibt.
	ID int64

	// Printing ist die Scryfall-Kennung des Drucks.
	//
	// **Hier bleibt sie Pflicht**: Preise kommen von Scryfall, und ein Druck
	// ohne Alias hat dort keine (A2, #1441).
	Printing string

	// DruckID ist die lokale Zeile des Drucks (0086) — gesetzt von der Ablage.
	DruckID int64

	Quelle Preisquelle

	// Markt sagt, wo der Preis gilt — Papier oder Magic Online (A3, #1442).
	//
	// **Pflicht, und ohne Vorgabe**: „irgendein Markt" gibt es nicht, und ein
	// Papierpreis, der still als Onlinepreis durchginge, wäre eine Zahl in
	// einer Rechnung, in die sie nicht gehört.
	Markt Preismarkt

	// Art sagt, in welche Richtung der Preis gilt — was ein Anbieter verlangt
	// oder was er zahlt (A3, #1442).
	Art Preisart

	// Waehrung ist der Code, unter dem die Quelle den Preis führt.
	//
	// **Nicht auf ISO 4217 eingeschränkt**: Scryfall führt neben `usd` und
	// `eur` auch `tix` — Karten für Magic Online, keine Währung eines Landes.
	// Eine Prüfung gegen ISO wiese sie ab und verlöre eine echte Angabe (§8).
	Waehrung string

	Finish Finish

	// Zustand ist die Erhaltung, für die der Preis gilt — leer heißt unbekannt.
	//
	// **Heute füllt ihn keine Quelle** (A3, #1442): Scryfall und MTGJSON
	// nennen keinen. Er steht trotzdem **in der Reihe**, weil A4 (#1443)
	// Angebote von CardTrader bringt, die sich **genau darin** unterscheiden —
	// ohne ihn liefen zwei Angebote derselben Karte in einer Reihe zusammen,
	// und die Reihe wäre eine Behauptung statt einer Auskunft.
	Zustand string

	// Sprache ist die Sprachfassung, für die der Preis gilt — leer heißt
	// unbekannt. **Derselbe Grund wie bei [Preisstand.Zustand].**
	Sprache string

	// Lage sagt, was über diesen Preis feststeht (A3, #1442).
	//
	// **Sie hat [Preisstand.Weggefallen] abgelöst.** Ein `bool` beantwortete
	// zwei Fragen zugleich: „die Quelle nennt hier nichts" und „diese Antwort
	// deckte die Reihe gar nicht ab" sahen gleich aus. Was die drei Werte
	// bedeuten, steht bei [Preislage].
	Lage Preislage

	// Beobachtet ist der Tag, den die **Quelle** nennt — null heißt: keiner.
	//
	// # Nicht dasselbe wie [Preisstand.Erhoben]
	//
	// MTGJSON führt seine Preise unter einem ISO-8601-**Datum** je Anbieter
	// (`PricePoints`, nachgelesen am 2026-09-11); ein Abruf von heute kann
	// einen Preis von vorgestern bringen, und der Verlauf gehört an **den**
	// Tag. Scryfall nennt keinen — dort bleibt dieses Feld null, und das ist
	// die ehrliche Angabe (§16, Satz 6).
	Beobachtet time.Time

	// Wert ist der Preis in der **kleinsten Einheit** der Währung — bei EUR
	// Cent, beim Yen ganze Yen.
	//
	// Warum ganzzahlig, steht bei [MarktpreisAus].
	Wert int64

	// Erhoben ist der Zeitpunkt, zu dem dieser Stand geholt wurde.
	//
	// **Nicht der Zeitpunkt, den die Quelle nennt** — sie nennt keinen. Was
	// hier steht, ist die Aussage „so stand es, als wir gefragt haben", und
	// mehr ist über einen täglichen Schätzwert auch nicht zu sagen.
	Erhoben time.Time
}

// Alter sagt, wie lange dieser Stand schon dasteht.
//
// **Nie negativ.** Eine falsch gehende Uhr ergäbe sonst „vor −3 Stunden", und
// das hilft niemandem; dann ist der Stand eben frisch.
func (p Preisstand) Alter(jetzt time.Time) time.Duration {
	alter := jetzt.Sub(p.Erhoben)
	if alter < 0 {
		return 0
	}

	return alter
}

// Validate gibt einen Befund je Feld, das nicht stimmt.
func (p Preisstand) Validate() []*FieldError {
	var befunde []*FieldError

	if strings.TrimSpace(p.Printing) == "" {
		befunde = append(befunde, NewFieldError("printing", SchluesselPflichtfeld,
			"Ein Preis gehört zu einem Druck."))
	}

	if strings.TrimSpace(string(p.Quelle)) == "" {
		befunde = append(befunde, NewFieldError("quelle", SchluesselPflichtfeld,
			"Ein Preis ohne Quelle ist keine Auskunft."))
	}

	if strings.TrimSpace(p.Waehrung) == "" {
		befunde = append(befunde, NewFieldError("waehrung", SchluesselPflichtfeld,
			"Ein Preis ohne Währung ist eine Zahl."))
	}

	if !IstFinish(p.Finish) {
		befunde = append(befunde, NewFieldError("finish", SchluesselValidierung,
			"Die Ausführung %q gibt es nicht.", p.Finish))
	}

	befunde = append(befunde, p.reihenbefunde()...)

	if p.Wert < 0 {
		befunde = append(befunde, NewFieldError("wert", SchluesselValidierung,
			"Ein Preis ist nicht negativ."))
	}

	// **Was nicht genannt ist, hat keinen Wert** (A3, #1442). Eine Zahl neben
	// „weggefallen" sähe aus wie ein Preis und wäre keiner.
	//
	// **Nur bei einer Lage, die es gibt**: bei einer unbekannten steht der
	// Befund schon oben, und ein zweiter daneben zeigte auf ein Feld, das
	// nichts dafür kann.
	if IstPreislage(p.Lage) && p.Lage != PreisGenannt && p.Wert != 0 {
		befunde = append(befunde, NewFieldError("wert", SchluesselValidierung,
			"Ein Stand, den niemand genannt hat, trägt keinen Wert."))
	}

	if p.Erhoben.IsZero() {
		befunde = append(befunde, NewFieldError("erhoben", SchluesselPflichtfeld,
			"Ein Preis ohne Zeitpunkt sagt nicht, wie alt er ist."))
	}

	// **Ein Beobachtungsdatum nach dem Abruf ist keins** — die Quelle kann
	// nichts über morgen gesagt haben.
	//
	// **Mit einem Tag Luft**: MTGJSON nennt einen Tag ohne Zone, und ein Abruf
	// um 01:00 Uhr in Berlin sieht den Tag, der in UTC gerade angefangen hat.
	if !p.Beobachtet.IsZero() && p.Beobachtet.After(p.Erhoben.Add(24*time.Hour)) {
		befunde = append(befunde, NewFieldError("beobachtet", SchluesselValidierung,
			"Die Quelle kann nichts über einen Tag gesagt haben, der noch kommt."))
	}

	return befunde
}

// reihenbefunde prüft die drei Aufzählungen, die eine Reihe ausmachen.
//
// **Als eigener Weg, damit [Preisstand.Validate] unter der Zeilengrenze
// bleibt** (§2) — und weil die drei zusammengehören: sie sagen, **welche**
// Reihe gemeint ist.
func (p Preisstand) reihenbefunde() []*FieldError {
	var befunde []*FieldError

	if !IstPreismarkt(p.Markt) {
		befunde = append(befunde, NewFieldError("markt", SchluesselValidierung,
			"Den Markt %q gibt es nicht.", string(p.Markt)))
	}

	if !IstPreisart(p.Art) {
		befunde = append(befunde, NewFieldError("art", SchluesselValidierung,
			"Die Preisart %q gibt es nicht.", string(p.Art)))
	}

	if !IstPreislage(p.Lage) {
		befunde = append(befunde, NewFieldError("lage", SchluesselValidierung,
			"Die Lage %q gibt es nicht.", string(p.Lage)))
	}

	return befunde
}

// Reihe ist der Schlüssel, unter dem zwei Stände dasselbe meinen.
//
// # Was dazugehört und was nicht
//
// **Alles, was die Frage benennt** — Quelle, Markt, Preisart, Währung,
// Ausführung, Zustand, Sprache. **Nichts, was die Antwort ist**: Wert,
// Zeitpunkt und Lage ändern die Reihe nicht, sonst wäre jeder Abruf eine neue,
// und ein Verlauf gäbe es nicht.
//
// **Der Druck steht nicht darin**: eine Reihe wird immer innerhalb eines
// Drucks verglichen, und ihn mitzuführen machte den Schlüssel länger, ohne
// eine Frage zu beantworten.
func (p Preisstand) Reihe() string {
	return strings.Join([]string{
		string(p.Quelle), string(p.Markt), string(p.Art),
		p.Waehrung, string(p.Finish), p.Zustand, p.Sprache,
	}, "\x1f")
}

// MarktpreisAus liest den Preis einer Quelle — LOC-03.AK-02
// (#1689).
//
// # Warum mit Währung
//
// **Die Währung kommt mit, und sie sagt die Stellen.** Zwei feste
// Nachkommastellen erfänden für einen Yenpreis eine Genauigkeit, die es nicht
// gibt, und wiesen bei einem Dinarpreis die dritte Stelle als Drift ab,
// obwohl sie dort die letzte ist. Gerechnet wird über [GeldAus] — ein Vertrag
// für jeden Betrag dieses Programms ([0127](../../docs/entscheidungen/0127-ein-betrag-steht-in-der-kleinsten-einheit.md)).
//
// # Warum ein leerer Text hier ein Fehler ist und bei [GeldAus] nicht
//
// **Weil es zwei verschiedene Fragen sind.** In einer Sammlung heißt „kein
// Preis" **kein Preis erfasst** — das ist die häufigste Zeile. Bei einer
// **Quelle** heißt es: sie hat einen Preis angekündigt und nichts
// geschickt. Eine Null daraus wäre ein Preis von null, und der sähe aus wie
// „geschenkt" (A3, #1442).
//
// # Warum ganzzahlig und nicht als Gleitkommazahl
//
// 4,17 sind hier 417. Ein `float64` daraus ergibt beim Summieren einer
// Sammlung Abweichungen im Cent-Bereich, die niemand erklären kann — und eine
// Portfoliosumme, die zweimal verschieden ausfällt, ist keine.
func MarktpreisAus(text, waehrung string) (int64, error) {
	if strings.TrimSpace(text) == "" {
		return 0, fmt.Errorf("%w: ein leerer Preis ist keiner", ErrValidation)
	}

	betrag, fehler := GeldAus(text, waehrung)
	if fehler != nil {
		return 0, fehler
	}

	// **Ein Marktpreis ist nicht negativ.** [GeldAus] liest ein Minus, weil
	// die kanonische Form eines hat; ob ein Betrag negativ sein darf,
	// entscheidet, wer ihn führt — ein **Erwerbs**preis darf es nicht
	// (`Exemplar.Preisbefunde`). Hier stand bis zum Zweitblick über Phase 6.3
	// das Gegenteil (G19). Was eine Quelle unter einem negativen Marktpreis
	// verstünde, weiß niemand, und Raten ist hier Drift (§16).
	if betrag < 0 {
		return 0, fmt.Errorf("%w: %q ist ein negativer Preis",
			ErrValidation, text)
	}

	return betrag, nil
}
