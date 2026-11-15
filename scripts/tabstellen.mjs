// **Was mit der Tabulatortaste erreichbar ist** — der eine Ausdruck, den jedes
// Messgerät dieses Projekts benutzt.
//
// # Warum eine eigene Datei
//
// Er stand in `serienmass.mjs`, und das ist ein Lauf: wer ihn importiert,
// fährt die ganze Messung. Seit `arbeitswegmass.mjs` (Werkstück U1 von Phase
// 5.1, #1412)
// denselben Ausdruck braucht, liegt er hier — **einmal**, damit zwei Messungen
// nicht mit zwei Begriffen von „erreichbar" nebeneinanderliegen (§10: wer
// zwei Zahlen vergleicht, prüft zuerst, ob dasselbe Gerät sie erhoben hat).
//
// # Der Ausdruck für „tabulatorerreichbar" ist der Punkt
//
// **`button:not([disabled])` genügt nicht.** Ein Knopf mit `tabindex="-1"`
// erfüllt ihn und ist trotzdem nicht erreichbar. Solange nichts im Fenster ein
// `tabindex="-1"` trug, war das dasselbe; seit die Kartenliste ein Gitter ist
// (#1323), ist es das nicht mehr — der erste Lauf danach gab weiter 86 statt
// 20 und sah aus wie ein wirkungsloser Eingriff.
//
// **`:not([tabindex="-1"])` an jeder Stelle**, und `[tabindex="0"]` dazu für
// alles, was von sich aus nicht bedienbar wäre — eine Zeile eines Gitters etwa.
export const tabulatorerreichbar = [
  'a[href]:not([tabindex="-1"])',
  'button:not([disabled]):not([tabindex="-1"])',
  'input:not([disabled]):not([tabindex="-1"])',
  'select:not([disabled]):not([tabindex="-1"])',
  'textarea:not([disabled]):not([tabindex="-1"])',
  '[tabindex="0"]',
].join(", ");
