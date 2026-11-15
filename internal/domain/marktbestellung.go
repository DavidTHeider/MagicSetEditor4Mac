package domain

// **Eine Bestellung bei einem Marktplatz** — Werkstück `T11` von Phase 5.6
// (#1476).
//
// # Warum die Zustände hier stehen und nicht im Adapter
//
// **Weil an ihnen eine fachliche Regel hängt**, und zwar die teuerste dieses
// Werkstücks: wann darf eine fremde Bestellung den **eigenen** Bestand
// bewegen. Ein Adapter, der sie selbst beantwortete, beantwortete sie je
// Anbieter neu.
//
// # Und warum sie ihre fremden Namen behalten
//
// **Weil sie verglichen und nicht angezeigt werden** (§9). Was von außen
// kommt, kommt in seinen Worten; eine deutsche Entsprechung machte aus einem
// Vergleich eine Zuordnung, die beim nächsten Feldnamen auseinanderläuft.
// Übersetzt wird erst am Fenster.

// Bestellzustand ist der Stand einer Bestellung beim Anbieter.
type Bestellzustand string

// Die acht — **nachgelesen am 2026-09-19** unter „Order Lifecycle"
// ([0122](../../docs/entscheidungen/0122-ein-fremdes-konto-wird-gelesen.md)).
const (
	// BestellungBezahlt heißt: jemand hat gekauft, und es ist zu schicken.
	BestellungBezahlt Bestellzustand = "paid"

	// BestellungVersandt, BestellungAngekommen und BestellungErledigt sind
	// die drei Schritte danach.
	BestellungVersandt   Bestellzustand = "sent"
	BestellungAngekommen Bestellzustand = "arrived"
	BestellungErledigt   Bestellzustand = "done"

	// BestellungHubOffen gibt es nur über **CardTrader Zero**: die Karte ist
	// verkauft, aber noch nicht zu schicken — der Anbieter bündelt sie
	// einmal je Woche zu einer Bestellung im Zustand [BestellungBezahlt].
	//
	// **Hier liegt die Falle dieses Werkstücks.** Wer beide für Verkäufe
	// hielte, nähme dasselbe Stück zweimal aus dem Bestand.
	BestellungHubOffen Bestellzustand = "hub_pending"

	// BestellungWiderrufsantrag, BestellungWiderrufen und
	// BestellungVerschollen sind die drei Sonderfälle.
	BestellungWiderrufsantrag Bestellzustand = "request_for_cancel"
	BestellungWiderrufen      Bestellzustand = "canceled"
	BestellungVerschollen     Bestellzustand = "lost"
)

// Bestellzustaende gibt alle acht, in der Reihenfolge des Lebenslaufs.
//
// **Als Funktion und nicht als Variable**: eine Aufzählung, die sich ändern
// lässt, ist keine.
func Bestellzustaende() []Bestellzustand {
	return []Bestellzustand{
		BestellungBezahlt, BestellungVersandt, BestellungAngekommen,
		BestellungErledigt, BestellungHubOffen,
		BestellungWiderrufsantrag, BestellungWiderrufen,
		BestellungVerschollen,
	}
}

// IstBestellzustand sagt, ob ein Wert einer der acht ist.
//
// **Ein unbekannter ist ein Befund und kein stilles Ja** (§8, §16): der
// Anbieter darf einen neunten einführen, und dann soll dieses Programm es
// sagen statt zu raten.
func IstBestellzustand(wert Bestellzustand) bool {
	for _, bekannt := range Bestellzustaende() {
		if wert == bekannt {
			return true
		}
	}

	return false
}

// BewegtDenBestand sagt, ob diese Bestellung ein Stück aus dem eigenen
// Bestand genommen hat.
//
// # Die Tabelle, so wie sie beim Anbieter steht
//
// **Nachgelesen am 2026-09-19**, und sie hat genau drei Zeilen:
//
//	Zustand       über CardTrader Zero   Bestand bewegen?
//	hub_pending   ja                     ja
//	paid          ja                     nein
//	paid          nein                   ja
//
// **Alles andere bewegt nichts** — `sent`, `arrived` und `done` sind
// Nachrichten über eine Bestellung, die schon gebucht ist; `canceled` und
// `lost` sind Nachrichten darüber, dass sie es nicht mehr ist.
//
// # Was diese Funktion **nicht** tut
//
// **Sie bucht nicht.** Sie beantwortet eine Frage; ob und wann jemand daraus
// eine Bewegung macht, entscheidet ein Mensch (T10,
// #1475) —
// dieses Programm bucht aus einer fremden Bestellung **gar nicht**
// automatisch, und der Grund steht in
// [0122](../../docs/entscheidungen/0122-ein-fremdes-konto-wird-gelesen.md).
//
// **Und sie kennt den Vorverkauf nicht.** Der Anbieter nennt dafür eine
// vierte Bedingung — die Bestellkennung muss der `hub_pending_order_id` der
// Position gleichen —, und die liegt an der **Position** und nicht am
// Zustand. Wer Vorverkauf betreibt, prüft sie zusätzlich; hier steht sie
// nicht, weil diese Funktion sie nicht sehen kann (§10).
func BewegtDenBestand(zustand Bestellzustand, ueberNullkonto bool) bool {
	if zustand == BestellungHubOffen {
		return ueberNullkonto
	}

	return zustand == BestellungBezahlt && !ueberNullkonto
}
