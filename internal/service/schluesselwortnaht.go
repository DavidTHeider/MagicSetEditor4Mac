// **Die Schlüsselwortnaht steht, bevor das Spielskript läuft** — Befund 5 des
// Zweitblicks über Phase 4.11.
//
// # Der Befund
//
// `magic.mse-game` baut `text_filter` beim Initialisieren als **verkettete**
// Funktion:
//
//	text_filter := … + expand_keywords@(…) + …
//
// Die Verkettung hält den Funktions**wert** fest, den `expand_keywords` in
// diesem Augenblick hatte — und das war der Stummel aus `internal/msescript`,
// weil die echte Funktion erst je Karte gesetzt wurde. Ein späteres
// Überschreiben ändert die gebaute Kette nicht mehr.
//
// **Die Folge**: über den produktiven Weg traf **kein** Schlüsselwort. Kein
// Erinnerungstext auf einer Karte, keine Verwendung im Vermerk, und die
// Auszählungsdimension `keywords` blieb leer
// (#1363).
//
// # Warum MSE dieses Problem nicht hat
//
// Dort ist `expand_keywords` eine **globale** eingebaute Funktion
// (`init_script_functions`, `src/script/functions/basic.cpp` Zeile 792,
// Fassung `f9d9356` vom 2021-07-06). Sie wird nie ausgetauscht; sie liest die
// Wortliste zur Laufzeit aus dem Set (`set->keyword_db`, `basic.cpp` Zeilen
// 660–681). Eine eingefrorene Kette hält damit immer die richtige Funktion.
//
// # Was hier daraus folgt
//
// **Die echte Funktion steht da, bevor das Spielskript sie einfriert** — auf
// der Umgebung des **Spiels**, nicht erst auf der der Karte. Genau wie bei
// MSE gibt es sie damit nur einmal.
//
// **Und der Vermerk gehört trotzdem zur Karte.** Er ist deshalb ein Zeiger,
// den die Funktion behält, und wird vor jeder Karte geleert — Karten werden
// nacheinander gerechnet, nicht nebeneinander.

package service

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"

// belegeSchluesselwortnaht setzt `expand_keywords` und `keyword_usage`.
//
// **Beide zusammen**, weil die zweite ohne die erste nie etwas zu lesen hat.
func belegeSchluesselwortnaht(
	umgebung *msescript.Umgebung, liste Wortliste, vermerk *Schluesselwortverwendung,
) {
	umgebung.Setze("expand_keywords", schluesselworteinsatz(liste, vermerk))
	umgebung.Setze("keyword_usage", schluesselwortverwendungsfrage(vermerk))
}

// Leere vergisst, was bisher vermerkt wurde.
//
// **Vor jeder Karte.** Der Vermerk hängt an der Umgebung des Spiels und lebt
// damit über alle Karten; ohne dieses Leeren trüge die zweite Karte die
// Schlüsselwörter der ersten mit.
func (v *Schluesselwortverwendung) Leere() {
	v.eintraege = nil
	v.laufend = ""
}

// Stand gibt den heutigen Inhalt heraus und leert den Vermerk.
//
// **Beides in einem Zug**, weil es zusammengehört: wer den Stand einer Karte
// nimmt, ist mit ihr fertig.
//
// **Wozu es das braucht**: die Auszählung baut **erst alle** Kartenumgebungen
// und wertet **danach** aus — nur so trägt `set.cards` dieselben Abbildungen
// wie `card` (#1356). Der Vermerk hängt aber am Spiel und hielte am Ende nur
// die **letzte** Karte. Gemessen am 2026-08-31: über `White.mse-set` gaben
// deshalb 0 von 3.761 Karten etwas, obwohl die erste Karte einzeln „Equip"
// ergab.
func (v *Schluesselwortverwendung) Stand() []schluesselwortvermerk {
	genommen := v.eintraege
	v.eintraege = nil
	v.laufend = ""

	return genommen
}

// Setze legt einen vorher genommenen Stand zurück.
func (v *Schluesselwortverwendung) Setze(stand []schluesselwortvermerk) {
	v.eintraege = stand
	v.laufend = ""
}
