// **Was ein Profil vom Menschen erbt** — Werkstück 9 von Phase 4.12
// (#1351),
// Punkt 9 der Kommentierung vom 2026-09-05.
//
// # Die Richtung
//
// **Der Mensch hält die Vorgabe, das Profil sticht sie.** „Meine Sprache" ist
// eine Aussage über eine Person; sie in jedem Profil noch einmal zu setzen
// hieße, dieselbe Antwort so oft zu geben, wie es Arbeitsumgebungen gibt.
// **Wer für ein Turnier auf Englisch arbeiten will, sagt es aber genau dort**
// — und deshalb erbt ein Profil, statt festzuliegen.
//
// # Warum nur vier
//
// **Weil nur diese vier beschreiben, wie die Anwendung aussieht und
// spricht.** Das ist eine Eigenschaft des Menschen. `aktualisierung` betrifft
// den **Rechner** (geht etwas ins Netz?), `bildkonflikt` einen **Ausgabelauf**,
// `ersteschritte` ein **Fenster**, das einmal aufgeht. Keine davon wird
// dadurch besser, dass sie zwei Ebenen hat — sie bekämen nur eine Frage mehr.
//
// # Und warum `erben` die Vorgabe ist und nicht ein Wert
//
// **Weil ein frisches Profil nichts über den Menschen behaupten soll.** Stünde
// dort `system`, hätte es eine Wahl getroffen, die niemand getroffen hat — und
// die Vorgabe des Menschen käme nie zum Tragen.
//
// **Ein bestehender Bestand merkt nichts davon**: seine Profile tragen
// **gesetzte** Werte, und gesetzt sticht. Die Vorgabe gilt nur, wo nie jemand
// gewählt hat, und dort galt vorher dieselbe Vorgabe — nur eine Ebene tiefer.

package domain

import "slices"

// Erben heißt: nimm, was beim Menschen steht.
//
// **Nur auf der Profilebene.** Der Mensch ist die unterste; wer dort „erben"
// wählen könnte, zeigte auf sich selbst.
const Erben = "erben"

// uebersteuerbare sind die Schlüssel, die ein Profil vom Menschen erbt.
//
// **Eine Aufzählung und keine Regel über Namen.** Ob eine Einstellung
// übersteuerbar ist, ist eine **fachliche** Entscheidung; sie an einem Präfix
// festzumachen hieße, sie beim nächsten Namen zu vergessen.
var uebersteuerbare = []string{
	EinstellungSprache,
	EinstellungErscheinungsbild,
	EinstellungSchriftgroesse,
	EinstellungSchriftart,
}

// Uebersteuerbare gibt die Schlüssel, die ein Profil vom Menschen erbt.
//
// **Eine Kopie**: ein Aufrufer, der die Liste änderte, änderte die Regel.
func Uebersteuerbare() []string { return slices.Clone(uebersteuerbare) }

// Verschmolzen sagt, welcher Wert für ein Profil gilt.
//
// **Drei Stufen, und die Reihenfolge ist die Aussage**: was das Profil
// gewählt hat, sonst was der Mensch gewählt hat, sonst die Vorgabe.
//
// **Ein Schlüssel, den niemand erbt, gibt seinen Profilwert zurück** — ohne
// Sonderfall beim Aufrufer: `erben` steht bei ihm gar nicht zur Wahl, also
// trifft die erste Stufe immer.
func Verschmolzen(schluessel, profilwert, nutzerwert string) string {
	if profilwert != Erben {
		return profilwert
	}

	if nutzerwert != "" {
		return nutzerwert
	}

	return NutzerVorgabe(schluessel)
}
