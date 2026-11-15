package domain

// **Was eine Rückkehr kostet, und was sie unklar lässt** — Werkstück O7 von
// Phase 5.2 (#1431).
//
// # Die Frage, für die es diese Datei gibt
//
// > Ein historisches Backup kann bereits abgegebene Karten enthalten.
//
// Wer eine Sicherung von vor drei Wochen einspielt, bekommt den Bestand von
// damals — samt der Exemplare, die seither aus der Sammlung hinausgegangen
// sind. Sie stehen wieder da, und **niemand weiß, ob sie noch jemandem
// gehören**.
//
// # Drei Antworten, und die dritte ist keine Notlösung
//
// „Ist da" und „ist weg" reichen nicht; die richtige Antwort ist **unklar** —
// dieselbe Gestalt wie `ungeklaert` in O6. Daraus folgen zwei Dinge, und
// beide stehen im Issue:
//
//  1. **Unklar heißt nicht handelbar.** Bis jemand nachgesehen hat, ist die
//     betroffene Menge aus dem Verkehr.
//  2. **Unklar heißt ausdrücklich auch nicht ausgebucht.** Eine Anwendung, die
//     nach einer Rückkehr selbsttätig abzöge, machte aus einer Unsicherheit
//     eine Behauptung — und zwar eine, die niemand mehr rückgängig macht.
//
// # Warum das hier steht, obwohl es noch keinen Handel gibt
//
// **Weil der Vertrag vor dem Verbraucher da sein muss.** T2 und T6 (Phase
// 5.6) bauen die Verbuchung; wer die Epoche erst dann erfände, hätte einen
// Bestand voller Rückkehren, über die nichts aufgeschrieben ist.

import "time"

// Vergleichszeile ist eine Zeile des Vergleichs zwischen Sicherung und
// laufendem Bestand.
//
// **−1 heißt „diese Seite sagt nichts darüber"** — dieselbe Bedeutung wie in
// [Sicherungsbefund]: eine alte Sicherung kennt die Tabelle nicht, und das ist
// etwas anderes als null.
type Vergleichszeile struct {
	Gegenstand  string `json:"gegenstand"`
	InSicherung int    `json:"inSicherung"`
	Jetzt       int    `json:"jetzt"`
}

// Unbekannt sagt, ob eine der beiden Seiten schweigt.
func (z Vergleichszeile) Unbekannt() bool { return z.InSicherung < 0 || z.Jetzt < 0 }

// Verlust ist, was seit der Sicherung **dazugekommen** ist und beim
// Zurückspielen verschwindet.
//
// **Nur nach oben.** Was seither weniger geworden ist, kommt zurück — das ist
// kein Verlust, sondern der Zweck der Übung.
func (z Vergleichszeile) Verlust() int {
	if z.Unbekannt() || z.Jetzt <= z.InSicherung {
		return 0
	}

	return z.Jetzt - z.InSicherung
}

// Klaerungsstand sagt, ob jemand nachgesehen hat.
type Klaerungsstand string

const (
	KlaerungOffen     Klaerungsstand = "offen"
	KlaerungQuittiert Klaerungsstand = "quittiert"
)

// Sammlungsmenge ist eine Stückzahl eines Gegenstands der Sammlung.
//
// **Ein Typ für beide Seiten des Vergleichs und für die Epoche.** Er beschreibt
// „so viel von diesem Ding": in einem [Bestandsstand] den Besitz, in einer
// [Wiederherstellungsepoche] das, worüber nichts feststeht. Zwei Typen mit
// denselben acht Feldern liefen früher oder später auseinander.
//
// **Der Schlüssel und die drei Merkmale, nicht die Zeilenkennung.** Nach dem
// Tausch sind die Zeilen die der Sicherung; eine Kennung aus dem Bestand von
// vorher zeigte auf nichts. Was bleibt, ist, **welchen Gegenstand** die Menge
// meint — und das ist derselbe Schlüssel, den die Sammlung für Gleichheit
// benutzt (Druck, Zustand, Sprache, Ausführung).
type Sammlungsmenge struct {
	// Profil ist der Eigentümer.
	//
	// **Ohne ihn sperrte eine Rückkehr in einem Profil die Sammlung des
	// anderen mit** — und das wäre eine Aussage über Bestände, die niemand
	// angefasst hat.
	Profil int64 `json:"profil"`

	Schluessel string `json:"schluessel"`
	Zustand    string `json:"zustand"`
	Sprache    string `json:"sprache"`

	// Finish ist die Ausführung — **leer heißt `nonfoil`** (T2, #1467).
	//
	// **Ein Wahrheitswert könnte `etched` nicht sagen**, und eine Rückkehr,
	// die eine geätzte Karte als foil zurückschriebe, machte aus einer
	// richtigen Sammlung eine falsche.
	Finish Finish `json:"finish"`
	Menge  int    `json:"menge"`

	// Name, Setcode und Nummer sind der lesbare Bezug — **damit die
	// Oberfläche sagen kann, worum es geht**, ohne den Druck nachzuschlagen.
	// Dieselbe Entscheidung wie bei [Exemplar].
	Name    string `json:"name"`
	Setcode string `json:"setcode"`
	Nummer  string `json:"nummer"`
}

// kennung ist die Gleichheit einer Sammlungsmenge — Profil, Druck, Zustand,
// Sprache, Ausführung.
func (u Sammlungsmenge) kennung() mengenkennung {
	return mengenkennung{u.Profil, u.Schluessel, u.Zustand, u.Sprache, u.Ausfuehrung()}
}

type mengenkennung struct {
	profil     int64
	schluessel string
	zustand    string
	sprache    string
	finish     Finish
}

// Handelssperre ist die Antwort auf „wie viel davon darf raus?".
//
// **Ein Typ und kein `if` an dreißig Stellen** — dieselbe Überlegung wie bei
// [Exemplar.Besitzt]: die Regel gehört an eine Stelle, sonst hat sie irgendwann
// jemand anders formuliert.
type Handelssperre struct {
	unklar map[mengenkennung]int
}

// NeueHandelssperre baut sie aus den offenen Mengen — `nil` ist der Normalfall.
func NeueHandelssperre(mengen []Sammlungsmenge) Handelssperre {
	sperre := Handelssperre{unklar: make(map[mengenkennung]int, len(mengen))}
	for _, menge := range mengen {
		sperre.unklar[menge.kennung()] += menge.Menge
	}

	return sperre
}

// Offen sagt, ob überhaupt etwas gesperrt ist.
func (h Handelssperre) Offen() bool { return len(h.unklar) > 0 }

// Menge ist die Summe über alles Gesperrte — **eine Zahl für die Oberfläche**:
// „irgendetwas ist unklar" ist keine Auskunft.
func (h Handelssperre) Menge() int {
	var summe int
	for _, menge := range h.unklar {
		summe += menge
	}

	return summe
}

// Freigegeben sagt, wie viel von einem Besitz herausgehen darf.
//
// **Nie unter null.** Steht in der Sperre mehr, als jemand besitzt, ist eben
// alles gesperrt — eine negative Zahl wäre eine Buchung, und gebucht wird hier
// nichts.
func (h Handelssperre) Freigegeben(gegenstand Sammlungsmenge, besitz int) int {
	frei := besitz - h.unklar[gegenstand.kennung()]
	if frei < 0 {
		return 0
	}

	return frei
}

// Wiederherstellungsepoche ist eine Rückkehr, über die etwas aufgeschrieben ist.
//
// **Sie verjährt nicht.** Die Unsicherheit verschwindet nicht mit der Zeit und
// nicht beim nächsten Start; sie endet, wenn jemand quittiert — und dann steht
// sie weiter da, mit dem Zeitpunkt, zu dem er es tat.
type Wiederherstellungsepoche struct {
	ID int64 `json:"id"`

	// Zurueckgespielt ist der Zeitpunkt der Rückkehr.
	Zurueckgespielt time.Time `json:"zurueckgespielt"`

	// Datenstand ist der Zeitpunkt, auf den zurückgegangen wurde — **die Zahl,
	// die die Abnahme verlangt**: „Dialog nennt genau den wiederhergestellten
	// Datenzeitpunkt."
	Datenstand time.Time `json:"datenstand"`

	// Quelle ist der Name der Sicherung — für den Menschen, nicht für die
	// Arbeit.
	Quelle string `json:"quelle"`

	Klaerung   Klaerungsstand `json:"klaerung"`
	GeklaertAm time.Time      `json:"geklaertAm"`

	// Mengen sind die betroffenen Bestände.
	Mengen []Sammlungsmenge `json:"mengen"`
}

// Sperrt sagt, ob diese Epoche noch etwas zurückhält.
func (w Wiederherstellungsepoche) Sperrt() bool { return w.Klaerung != KlaerungQuittiert }

// Ausfuehrung gibt die Ausführung — **leer heißt `nonfoil`**.
//
// **Dieselbe Regel wie bei [Exemplar.Ausfuehrung]**, und an einer Stelle:
// eine Sicherung aus der Zeit vor `T2` trägt keine Angabe, und sie meint
// dasselbe wie ein `foil = 0` von damals.
func (u Sammlungsmenge) Ausfuehrung() Finish {
	if u.Finish == "" {
		return FinishNonfoil
	}

	return u.Finish
}
