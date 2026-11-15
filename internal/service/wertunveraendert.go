package service

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"

// wertfelder ist die Zahl der Felder von [msescript.Wert], die [wertBleibt]
// vergleicht — ein Wächter hält sie gegen den Typ.
const wertfelder = 11

// wertBleibt sagt, ob ein Wertskript seinen Wert gelassen hat, wie er war —
// PERF-03 (#2170).
//
// **Feld für Feld, nicht nach MSEs Gleichheit**: `==` in MSE-Script sieht
// über die Herkunft eines Werts hinweg, `card` trägt sie aber mit. Und ein
// zusammengesetzter Wert gilt immer als geändert — ihn zu vergleichen kostete,
// was das Überspringen sparen soll, und ein falsches Ja ließe `card` alt.
func wertBleibt(alt, neu msescript.Wert) bool {
	return alt.Art == neu.Art && alt.Zahl == neu.Zahl && alt.Text == neu.Text &&
		alt.Wahrheit == neu.Wahrheit && alt.Feld == neu.Feld && alt.Verbund == neu.Verbund &&
		alt.Kennung == neu.Kennung && alt.Marke == neu.Marke &&
		alt.Funktion == nil && neu.Funktion == nil &&
		len(alt.Eintraege) == 0 && len(neu.Eintraege) == 0 &&
		len(alt.Glieder) == 0 && len(neu.Glieder) == 0
}
