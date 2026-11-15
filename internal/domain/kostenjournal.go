package domain

// **Was ein Auftrag draußen kosten darf** — Werkstück O12 von Phase 5.2
// (#1439).
//
// # Warum das nicht dasselbe ist wie die Guthaben aus O5
//
// [Haushaltsgrenzen] zählen **Rufe und Plätze** je Prozess: sie halten eine
// Maschine davon ab, sich selbst lahmzulegen. Hier geht es um **Geld**, und
// Geld überlebt den Prozess: zwei gleichzeitig gestartete Fenster, ein
// Befehl und ein Dienst müssen zusammen unter derselben Grenze bleiben. Das
// kann nur der Bestand entscheiden.
//
// # Vier Stände, und der vierte ist der Grund für das Ganze
//
// Ein Auftrag, dessen Frist **nach** der möglichen Annahme durch den Anbieter
// abläuft, ist nicht „nichts passiert". Er ist **unklar**, und bis ihn jemand
// klärt, zählt er wie verbraucht. Das ist die konservative Seite, und sie
// steht so in der Abnahme: „unbekannte Kosten nicht als kostenlos behandeln".

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Betrag ist Geld in der **kleinsten Einheit** seiner Währung.
//
// **Ganzzahlig und ohne Rundung** — dieselbe Entscheidung wie bei
// [Kostengrenze]: 1,20 € sind 120, und `float64` hätte für einen Betrag, der
// gegen eine Grenze geht, nichts zu suchen.
type Betrag struct {
	Wert     int64  `json:"wert"`
	Waehrung string `json:"waehrung"`
}

// Pruefe weist einen Betrag ab, mit dem sich nicht rechnen lässt.
func (b Betrag) Pruefe() error {
	switch {
	case b.Wert < 0:
		return NewFieldError("betrag", SchluesselValidierung,
			"Ein Betrag ist nicht negativ.")
	// **Erst die Schreibweise, dann die Gestalt** — und die Gestalt in
	// Zeichen, nicht in Bytes (UX-06, #1678): `€` hat in UTF-8 drei davon.
	case b.Waehrung != strings.ToUpper(b.Waehrung) && istWaehrungscode(strings.ToUpper(b.Waehrung)):
		return NewFieldError("waehrung", SchluesselValidierung,
			"Eine Währung steht in Großbuchstaben — nicht %q.", b.Waehrung)
	case !istWaehrungscode(b.Waehrung):
		return NewFieldError("waehrung", SchluesselValidierung,
			"%q ist kein Währungscode — der hat drei Buchstaben, wie EUR.", b.Waehrung)
	}

	return nil
}

// Leer sagt, ob dieser Betrag nichts ist.
func (b Betrag) Leer() bool { return b.Wert == 0 }

// Plus addiert zwei Beträge **derselben** Währung.
//
// # Keine Umrechnung, und das ist die Entscheidung
//
// Das Issue verlangt es ausdrücklich: „Keine Wechselkursumrechnung ohne eigene
// Quelle." Diese Anwendung hat keine — ein Kurs von irgendwoher wäre eine
// Zahl, für die niemand einsteht, und aus ihr entstünde eine Grenze, die
// jemand für geprüft hält.
//
// **Ein Betrag von null passt zu jeder Währung.** Sonst brauchte jede Summe
// einen Anfangswert, der schon weiß, worin gerechnet wird.
func (b Betrag) Plus(dazu Betrag) (Betrag, error) {
	if b.Leer() && b.Waehrung == "" {
		return dazu, nil
	}

	if dazu.Leer() && dazu.Waehrung == "" {
		return b, nil
	}

	if b.Waehrung != dazu.Waehrung {
		return Betrag{}, fmt.Errorf(
			"%w: %s und %s lassen sich nicht addieren — diese Anwendung "+
				"rechnet keine Währungen um", ErrValidation, b.Waehrung, dazu.Waehrung)
	}

	return Betrag{Wert: b.Wert + dazu.Wert, Waehrung: b.Waehrung}, nil
}

// Minus zieht ab — nie unter null.
func (b Betrag) Minus(ab Betrag) (Betrag, error) {
	if ab.Leer() && ab.Waehrung == "" {
		return b, nil
	}

	if b.Waehrung != ab.Waehrung {
		return Betrag{}, fmt.Errorf(
			"%w: %s und %s lassen sich nicht abziehen — diese Anwendung "+
				"rechnet keine Währungen um", ErrValidation, b.Waehrung, ab.Waehrung)
	}

	if ab.Wert >= b.Wert {
		return Betrag{Wert: 0, Waehrung: b.Waehrung}, nil
	}

	return Betrag{Wert: b.Wert - ab.Wert, Waehrung: b.Waehrung}, nil
}

// Uebersteigt sagt, ob dieser Betrag über einer Grenze liegt.
//
// **Eine Grenze von null ist keine Grenze** — dieselbe Lesart wie bei
// [Kostengrenze]: wer keine setzt, hat keine.
func (b Betrag) Uebersteigt(grenze Betrag) (bool, error) {
	if grenze.Leer() {
		return false, nil
	}

	if b.Leer() && b.Waehrung == "" {
		return false, nil
	}

	if b.Waehrung != grenze.Waehrung {
		return false, fmt.Errorf(
			"%w: die Grenze gilt in %s, dieser Auftrag rechnet in %s — "+
				"diese Anwendung rechnet keine Währungen um",
			ErrValidation, grenze.Waehrung, b.Waehrung)
	}

	return b.Wert > grenze.Wert, nil
}

// Kostenstand ist, was über einen Auftrag feststeht.
type Kostenstand string

const (
	// KostenReserviert: gebucht **vor** dem Ruf, Ausgang offen.
	KostenReserviert Kostenstand = "reserviert"

	// KostenAbgerechnet: der Anbieter hat geantwortet, der Betrag steht.
	KostenAbgerechnet Kostenstand = "abgerechnet"

	// KostenFreigegeben: es hat nichts gekostet — der Ruf scheiterte vor
	// jeder möglichen Annahme.
	KostenFreigegeben Kostenstand = "freigegeben"

	// KostenUnklar: die Frist lief **nach** der möglichen Annahme ab, oder
	// die Abrechnung ging verloren.
	//
	// **Er zählt wie verbraucht**, bis ihn jemand klärt. Ein Auftrag, über
	// den nichts feststeht, als kostenlos zu führen, ist die Buchung, die
	// eine Rechnung überrascht.
	KostenUnklar Kostenstand = "unklar"
)

// Kostenstaende gibt alle vier.
func Kostenstaende() []Kostenstand {
	return []Kostenstand{
		KostenReserviert, KostenAbgerechnet, KostenFreigegeben, KostenUnklar,
	}
}

// Offen sagt, ob über diesen Posten noch nichts feststeht.
func (s Kostenstand) Offen() bool {
	return s == KostenReserviert || s == KostenUnklar
}

// Bindet sagt, ob dieser Stand Budget bindet.
//
// **Reserviert, abgerechnet und unklar binden; freigegeben nicht.** Der dritte
// ist der, über den man streiten könnte — und genau deshalb steht er hier.
func (s Kostenstand) Bindet() bool { return s != KostenFreigegeben }

// Kostenposten ist ein Auftrag im Journal.
//
// **Ohne Prompt und ohne Zugangsdaten.** Was hier steht, ist die **Aufgabe**
// — eine Marke wie `karte.text` —, nicht das, was hinausging: ein Journal
// liest, wer den Bestand hat, und eine Sicherung nimmt man mit.
type Kostenposten struct {
	// Kennung ist die **Auftragsidentität** und kommt vom Aufrufer.
	//
	// **Sie ist der Schutz gegen den zweiten kostenpflichtigen Auftrag**: wer
	// mit derselben Kennung noch einmal reserviert, bekommt den vorhandenen
	// Posten zurück und keinen neuen.
	Kennung string `json:"kennung"`

	ProfilID int64 `json:"profilID"`

	// Konto ist das Providerkonto — der Name des Zugangs, nicht sein Schlüssel.
	Konto string `json:"konto"`

	// Aufgabe sagt, wofür — eine Marke, kein Prompt.
	Aufgabe string `json:"aufgabe"`

	// Hoechstbetrag ist, was reserviert wurde. **Er ist Pflicht**: ein
	// Auftrag mit unbekannten Kosten wird nicht als kostenlos geführt.
	Hoechstbetrag Betrag `json:"hoechstbetrag"`

	// Betrag ist, was er wirklich gekostet hat — null, solange nichts
	// feststeht.
	Betrag Betrag `json:"betrag"`

	Stand Kostenstand `json:"stand"`

	// Grund sagt, warum ein Posten unklar ist — für Menschen.
	Grund string `json:"grund"`

	Begonnen time.Time `json:"begonnen"`
	Beendet  time.Time `json:"beendet"`
}

// Gebunden ist, was dieser Posten am Budget hält.
//
// **Der abgerechnete Betrag, solange er feststeht, sonst der reservierte.**
// Ein Auftrag, der weniger gekostet hat als reserviert, gibt den Rest frei;
// einer, über den nichts feststeht, hält den ganzen Höchstbetrag.
func (p Kostenposten) Gebunden() Betrag {
	if !p.Stand.Bindet() {
		return Betrag{Waehrung: p.Hoechstbetrag.Waehrung}
	}

	if p.Stand == KostenAbgerechnet {
		return p.Betrag
	}

	return p.Hoechstbetrag
}

// betragstext macht einen Betrag vergleichbar — Zahl und Währung.
func betragstext(betrag Betrag) string {
	return strconv.FormatInt(betrag.Wert, 10) + " " + betrag.Waehrung
}

// MeintDasselbeWie sagt, ob zwei Posten derselbe Auftrag sind — `nil`, wenn ja.
//
// **Die Identität ist mehr als die Kennung** — Befund 8 des Zweitblicks über
// Phase 5.2. Sie kommt vom Aufrufer (siehe [Kostenposten.Kennung]), und ein
// Aufrufer kann sich vertun: dieselbe Kennung für einen anderen Auftrag, ein
// anderes Konto, einen anderen Höchstbetrag. Käme dann der alte Posten als
// Erfolg zurück, wäre eine **fremde** Reservierung die Freigabe des neuen
// Rufs — und der Höchstbetrag, den jemand gesehen hat, wäre nicht der, der
// gilt.
//
// **Verglichen wird, was den Auftrag ausmacht**: Profil, Konto, Aufgabe und
// der Höchstbetrag samt Währung. **Nicht** verglichen wird, was sich im Lauf
// eines Auftrags ändert — Stand, tatsächlicher Betrag, Grund und Zeiten.
func (p Kostenposten) MeintDasselbeWie(anderer Kostenposten) error {
	for _, feld := range []struct {
		name           string
		einer, anderer string
	}{
		{"profil", strconv.FormatInt(p.ProfilID, 10), strconv.FormatInt(anderer.ProfilID, 10)},
		{"konto", p.Konto, anderer.Konto},
		{"aufgabe", p.Aufgabe, anderer.Aufgabe},
		{"hoechstbetrag", betragstext(p.Hoechstbetrag), betragstext(anderer.Hoechstbetrag)},
	} {
		if feld.einer != feld.anderer {
			return fmt.Errorf(
				"%w: die Kennung %q gehört zu einem anderen Auftrag — %s ist dort %q, hier %q",
				ErrConflict, p.Kennung, feld.name, feld.einer, feld.anderer)
		}
	}

	return nil
}

// Pruefe weist einen Auftrag ab, der so nicht ins Journal darf.
func (p Kostenposten) Pruefe() error {
	switch {
	case strings.TrimSpace(p.Kennung) == "":
		return NewFieldError("kennung", SchluesselValidierung,
			"Ein Auftrag braucht eine Identität — sonst gäbe es kein "+
				"Wiedererkennen und damit keinen Schutz vor dem zweiten Ruf.")
	case p.ProfilID <= 0:
		return NewFieldError("profilID", SchluesselValidierung,
			"Ein Auftrag gehört zu einem Profil.")
	case strings.TrimSpace(p.Konto) == "":
		return NewFieldError("konto", SchluesselValidierung,
			"Ein Auftrag nennt das Konto, das ihn bezahlt.")
	case strings.TrimSpace(p.Aufgabe) == "":
		return NewFieldError("aufgabe", SchluesselValidierung,
			"Ein Auftrag nennt, wofür er da ist.")
	}

	if fehler := p.Hoechstbetrag.Pruefe(); fehler != nil {
		return fehler
	}

	if p.Hoechstbetrag.Leer() {
		return NewFieldError("hoechstbetrag", SchluesselValidierung,
			"Ein Auftrag nach draußen nennt seinen Höchstbetrag. Unbekannte "+
				"Kosten sind nicht dasselbe wie keine.")
	}

	return nil
}

// Haushaltsstand ist, was an einem Konto steht — die vier Zahlen, die jede
// Fläche zeigt.
type Haushaltsstand struct {
	Konto string `json:"konto"`

	// Zeitraum ist das Fenster, über das gezählt wurde.
	//
	// **Die Antwort sagt es selbst.** Wer nichts nennt, bekommt den Tag; eine
	// Auskunft, die das verschwiege, ließe den Aufrufer glauben, er habe die
	// Monatsgrenze vor sich.
	Zeitraum Zeitraum `json:"zeitraum"`

	// Grenze ist das freigegebene Budget — null heißt: keines eingerichtet,
	// und dann geht **kein** kostenpflichtiger Auftrag.
	Grenze Betrag `json:"grenze"`

	Reserviert Betrag `json:"reserviert"`
	Verbraucht Betrag `json:"verbraucht"`
	Unklar     Betrag `json:"unklar"`

	Verfuegbar Betrag `json:"verfuegbar"`
}

// Zeitraum ist, worüber ein Budget gilt.
type Zeitraum string

const (
	ZeitraumTag   Zeitraum = "tag"
	ZeitraumMonat Zeitraum = "monat"
)

// Zeitraeume gibt beide.
func Zeitraeume() []Zeitraum { return []Zeitraum{ZeitraumTag, ZeitraumMonat} }

// IstZeitraum sagt, ob eine Marke einer der beiden ist.
func IstZeitraum(marke string) bool {
	for _, bekannt := range Zeitraeume() {
		if string(bekannt) == marke {
			return true
		}
	}

	return false
}

// Fenster gibt den Anfang des Zeitraums, in dem dieser Augenblick liegt.
//
// **In UTC**, damit zwei Prozesse in verschiedenen Zonen dasselbe Fenster
// meinen — der Bestand ist gemeinsam, und ein Tag, der zweimal anfängt, wäre
// zwei Budgets.
func (z Zeitraum) Fenster(jetzt time.Time) time.Time {
	inUTC := jetzt.UTC()

	if z == ZeitraumMonat {
		return time.Date(inUTC.Year(), inUTC.Month(), 1, 0, 0, 0, 0, time.UTC)
	}

	return time.Date(inUTC.Year(), inUTC.Month(), inUTC.Day(), 0, 0, 0, 0, time.UTC)
}
