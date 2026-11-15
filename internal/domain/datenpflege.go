package domain

// **Ein Befund ist kein Auftrag** — Werkstück O8 von Phase 5.2
// (#1432).
//
// # Warum es Klassen braucht und nicht nur Zahlen
//
// `storage.Waisen` zählt seit #1220 drei Stellen und sagt eine Zahl. Wer sie
// größer als null sieht, hat einen Befund und **keinen Weg** — und der
// naheliegende Weg wäre der falsche: löschen, bis die Zahl null ist.
//
// **Das verdeckt echte Korruption.** Eine Zeile, die auf eine fehlende Karte
// zeigt, ist ein Befund über den Bestand; sie wegzuräumen macht die Zahl hübsch
// und den Schaden unsichtbar. Das Issue sagt es ausdrücklich.
//
// Deshalb drei Klassen, und die Grenze zwischen ihnen ist die Arbeit dieses
// Stücks:
//
//   - **`sicher`** — die Zeile zeigt auf nichts und hält nichts fest.
//   - **`entscheidung`** — es sieht nach einem Fehler aus und **kann** einer
//     sein. Zwei Sammlungszeilen mit demselben Druck sind das Beispiel: Erwerb,
//     Zustand, Ort und Ausführung unterscheiden sie, und **gleiche
//     Printingdaten sind nicht automatisch Duplikate**.
//   - **`nurInformieren`** — auffällig, und es wird **nicht** angefasst.
//
// # Und warum der Plan eine Revision trägt
//
// Zwischen „zeig mir, was los ist" und „mach es" liegt eine Entscheidung eines
// Menschen — und in dieser Zeit kann sich der Bestand ändern. Eine Vorschau,
// deren Ergebnis man aufhebt und später ausführt, beschreibt die Vergangenheit.
// Die Revision ist die Prüfsumme über die betroffenen Zeilen; stimmt sie nicht
// mehr, **stoppt die betroffene Maßnahme — und nur sie**.

import (
	"fmt"
	"time"
)

// Die dreizehn Arten, nach denen gesehen wird — **Marken und keine Sätze**.
//
// **Sie stehen hier und nicht nur in `storage`**, obwohl die Abfragen dort
// liegen: `Pflegeabgleich` muss eine erfundene Art abweisen können, und dafür
// braucht er die Liste. Ein Wächter hält beide Seiten zusammen — was hier
// steht, hat dort eine Abfrage, und umgekehrt.
const (
	PflegeKartenwerte = "kartenwerteOhneKarte"
	PflegeSetangaben  = "setangabenOhneSet"
	PflegeSuchzeilen  = "suchzeilenOhneKarte"
	PflegeDeckplaetze = "deckplaetzeOhneDeck"
	PflegeZuordnungen = "zuordnungenOhneExemplar"
	PflegeVerlauf     = "verlaufOhneSet"
	PflegeNachweise   = "abgelaufeneNachweise"
	PflegeZugaenge    = "abgelaufeneToken"
	PflegeDubletten   = "moeglicheDubletten"
	PflegeRueckkehren = "offeneRueckkehren"

	// PflegeIntegritaet und PflegeFremdschluessel sind, was **SQLite selbst**
	// über seine Datei sagt — Befund 2 des Zweitblicks über Phase 5.2.
	//
	// **Beide nur informierend**, und das steht so im Issue: „echte
	// Korruption nicht durch kosmetisches Löschen verdecken". Was hier
	// auffällt, ist die Aufforderung, eine Sicherung einzuspielen.
	PflegeIntegritaet     = "integritaet"
	PflegeFremdschluessel = "fremdschluessel"

	// PflegeProtokolle sind **beendete** Partien jenseits der
	// Aufbewahrungsfrist (G5,
	// #1462).
	//
	// **Eine Entscheidung und kein sicherer Fall.** Ein beendetes Protokoll
	// ist kein Rest: es ist der Beleg darüber, wie eine Partie ausging. Wer
	// es von selbst wegräumte, nähme jemandem die Antwort auf eine Frage,
	// die er erst später stellt.
	PflegeProtokolle = "alteProtokolle"
)

// Pflegearten gibt die Reihe.
func Pflegearten() []string {
	return []string{
		PflegeKartenwerte, PflegeSetangaben, PflegeSuchzeilen, PflegeDeckplaetze,
		PflegeZuordnungen, PflegeVerlauf, PflegeNachweise, PflegeZugaenge,
		PflegeDubletten, PflegeRueckkehren, PflegeProtokolle,
		PflegeIntegritaet, PflegeFremdschluessel,
	}
}

// IstPflegeart sagt, ob eine Marke eine der dreizehn ist.
func IstPflegeart(art string) bool {
	for _, bekannt := range Pflegearten() {
		if bekannt == art {
			return true
		}
	}

	return false
}

// Pflegeklasse sagt, wie mit einem Befund umzugehen ist.
type Pflegeklasse string

const (
	// PflegeSicher: die Zeile zeigt auf nichts und hält nichts fest.
	PflegeSicher Pflegeklasse = "sicher"

	// PflegeEntscheidung: es sieht nach einem Fehler aus und kann einer sein.
	PflegeEntscheidung Pflegeklasse = "entscheidung"

	// PflegeNurInformieren: auffällig, und es wird nicht angefasst.
	PflegeNurInformieren Pflegeklasse = "nurInformieren"
)

// Pflegefrist ist, wie lange eine Vorschau etwas beschreibt.
//
// **Eine Viertelstunde, und die Zahl ist eine Wahl.** Sie ist lang genug, um
// nachzudenken, und kurz genug, dass niemand einen Plan von gestern ausführt.
// Die Revision fängt den Rest: läuft der Bestand vorher auseinander, meldet sie
// es, gleich wie frisch der Plan ist.
const Pflegefrist = 15 * time.Minute

// Pflegebefund ist eine Auffälligkeit im Bestand.
type Pflegebefund struct {
	// Art ist die Marke — vier Flächen bauen daraus ihren eigenen Satz.
	Art string `json:"art"`

	Klasse Pflegeklasse `json:"klasse"`

	// Anzahl ist, wie viele Zeilen es betrifft.
	Anzahl int `json:"anzahl"`

	// Revision ist die Prüfsumme über die betroffenen Zeilen.
	//
	// **Über die Zeilen und nicht über die Zahl.** Zwei Zeilen weg und zwei
	// dazu ergäben dieselbe Zahl und einen anderen Bestand.
	Revision string `json:"revision"`

	// Beispiele nennen ein paar der betroffenen Zeilen — **damit jemand
	// entscheiden kann**, nicht damit etwas geprüft wäre.
	Beispiele []string `json:"beispiele"`
}

// OhneRueckfrage sagt, ob dieser Befund ohne Entscheidung laufen darf.
func (b Pflegebefund) OhneRueckfrage() bool { return b.Klasse == PflegeSicher }

// Massnahmenplan ist die Vorschau samt ihrer Frist.
type Massnahmenplan struct {
	Erstellt   time.Time      `json:"erstellt"`
	Gueltigbis time.Time      `json:"gueltigbis"`
	Befunde    []Pflegebefund `json:"befunde"`
}

// Pflegewahl ist eine gewählte Maßnahme samt der Revision, die der Wählende
// **gesehen** hat.
//
// # Warum die Revision mitkommt und nicht der ganze Plan
//
// **Weil sie die Bindung ist.** Der Plan ist eine Vorschau; was ihn an einen
// Bestand bindet, ist die Prüfsumme über die betroffenen Zeilen. Wer sie
// mitschickt, sagt: „ich habe **diese** Zeilen gesehen."
//
// **Und weil vier Flächen ihn sonst herumtragen müssten.** Eine
// Kommandozeile, die einen ganzen Plan als JSON zurückgibt, hat keinen Weg,
// den ein Mensch geht.
type Pflegewahl struct {
	Art      string `json:"art"`
	Revision string `json:"revision"`
}

// Pflegeausgang ist, was aus einer Maßnahme wurde.
type Pflegeausgang string

const (
	// PflegeBereit: die Maßnahme darf laufen.
	PflegeBereit Pflegeausgang = "bereit"

	// PflegeKonflikt: die betroffenen Zeilen haben sich seit der Vorschau
	// geändert.
	PflegeKonflikt Pflegeausgang = "konflikt"

	// PflegeSchonWeg: es gibt den Befund nicht mehr.
	PflegeSchonWeg Pflegeausgang = "schonWeg"

	// PflegeUebersprungen: die Klasse lässt es nicht zu.
	PflegeUebersprungen Pflegeausgang = "uebersprungen"

	// PflegeErledigt: die Maßnahme ist gelaufen.
	PflegeErledigt Pflegeausgang = "erledigt"

	// PflegeGescheitert: sie ist gelaufen und hat nicht geklappt.
	PflegeGescheitert Pflegeausgang = "gescheitert"
)

// Die Gründe, aus denen eine Maßnahme nicht läuft — **Marken und keine Sätze**.
const (
	PflegegrundNurInformieren = "nurInformieren"
	PflegegrundGeaendert      = "geaendert"
)

// Pflegezeile ist eine Maßnahme mit ihrem Ausgang.
type Pflegezeile struct {
	Art     string        `json:"art"`
	Ausgang Pflegeausgang `json:"ausgang"`

	// Anzahl ist, wie viele Zeilen sie betrifft — nach der Ausführung, wie
	// viele es wirklich waren.
	Anzahl int `json:"anzahl"`

	// Grund steht bei den Ausgängen, die einen brauchen.
	Grund string `json:"grund"`
}

// Pflegeabgleich stellt eine Wahl neben eine frische Analyse.
//
// **Er ändert nichts.** Was er gibt, ist die Liste dessen, was laufen dürfte —
// und je Maßnahme der Grund, warum nicht.
//
// **Und er entscheidet je Maßnahme, nicht über die Wahl als Ganzes.** Ein
// Konflikt an einer Stelle nimmt den anderen nichts; das ist die Abnahme aus
// dem Issue.
//
// **Die Klasse kommt aus der frischen Analyse und nicht aus der Wahl.** Wer
// sie mitschickte, könnte einen `nurInformieren`-Befund als `sicher`
// ausgeben — und die Zusage hinge daran, dass der Aufrufer die Wahrheit sagt.
func Pflegeabgleich(
	erstellt time.Time, frisch []Pflegebefund, wahl []Pflegewahl, jetzt time.Time,
) ([]Pflegezeile, error) {
	// **Die Frist gilt für die ganze Vorschau** — sie sagt, ob diese
	// überhaupt noch etwas beschreibt.
	if jetzt.After(erstellt.Add(Pflegefrist)) {
		return nil, fmt.Errorf("%w: diese Vorschau ist abgelaufen — sieh noch einmal nach",
			ErrConflict)
	}

	heute := make(map[string]Pflegebefund, len(frisch))
	for _, befund := range frisch {
		heute[befund.Art] = befund
	}

	zeilen := make([]Pflegezeile, 0, len(wahl))

	for _, gewaehlt := range wahl {
		// **Eine erfundene Art wird abgewiesen und nicht als „schon weg"
		// gemeldet.** Beides löscht nichts — aber „es gibt diesen Befund
		// nicht mehr" und „es gibt ihn gar nicht" sind zwei Auskünfte, und
		// wer die erste bekommt, sucht an der falschen Stelle.
		if !IstPflegeart(gewaehlt.Art) {
			return nil, fmt.Errorf("%w: %q ist keine Maßnahme, die es gibt",
				ErrValidation, gewaehlt.Art)
		}

		zeilen = append(zeilen, abgleichen(gewaehlt, heute[gewaehlt.Art]))
	}

	return zeilen, nil
}

// abgleichen entscheidet über eine einzelne Maßnahme.
func abgleichen(gewaehlt Pflegewahl, jetzt Pflegebefund) Pflegezeile {
	zeile := Pflegezeile{Art: gewaehlt.Art, Anzahl: jetzt.Anzahl}

	switch {
	case jetzt.Art == "":
		// **Leer heißt: die frische Analyse kennt ihn nicht mehr.** Das ist
		// erledigt und kein Konflikt — wer ihn als Konflikt meldete, schickte
		// jemanden nachsehen, wo nichts mehr ist.
		zeile.Ausgang, zeile.Anzahl = PflegeSchonWeg, 0

	// **Nur die Klasse und keine zweite Bedingung daneben.** Der erste Anlauf
	// prüfte zusätzlich `!OhneRueckfrage()` — das ist bei dieser Klasse immer
	// wahr, und ein Konjunkt, das nie entscheidet, sieht wie eine Zusage aus
	// und ist keine (§10). **`entscheidung` fällt bewusst durch**: die Klasse
	// heißt „ein Mensch entscheidet", und wer sie wählt, hat entschieden.
	case jetzt.Klasse == PflegeNurInformieren:
		zeile.Ausgang, zeile.Grund = PflegeUebersprungen, PflegegrundNurInformieren
	case jetzt.Revision != gewaehlt.Revision:
		zeile.Ausgang, zeile.Grund = PflegeKonflikt, PflegegrundGeaendert
	default:
		zeile.Ausgang = PflegeBereit
	}

	return zeile
}
