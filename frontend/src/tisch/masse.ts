// **Der virtuelle Tisch ist ein echter Tisch** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Warum Millimeter und keine Punkte
//
// **Weil das Maß einer Karte physisch ist.** 63 × 88 mm steht in diesem
// Bestand schon einmal — `internal/service/bogen.go` legt danach den
// Druckbogen aus. Ein zweites Maß daneben hieße, dass derselbe Gegenstand
// zwei Größen hat, je nachdem, wer ihn ansieht.
//
// Die **Kamera** rechnet Millimeter in Punkte um. Damit hängt keine Zahl hier
// an einer Auflösung, und ein Retina-Bildschirm ändert am Plan nichts.
//
// # Woher die übrigen Zahlen kommen
//
// **Aus der Regel, wo sie eine nennt, und sonst aus dem Tisch.** CR 403.1
// (Fassung 2026-08-19) sagt, dass das Schlachtfeld die Fläche *zwischen* den
// Spielern ist und die Permanenten eines Spielers *vor ihm* liegen — daraus
// folgt die Anordnung, nicht die Zahl. Die Zahlen darunter sind Maße eines
// Spieltisches und als solche benannt; eine Regelnummer daneben wäre erfunden
// (§11).

// Kartenbreite und Kartenhoehe sind das Maß einer Karte in Millimetern.
//
// **Dieselben 63 × 88 wie auf dem Druckbogen** (`internal/service/bogen.go`).
export const Kartenbreite = 63
export const Kartenhoehe = 88

// Kartenabstand ist die Lücke zwischen zwei nebeneinanderliegenden Karten.
//
// **Knapp ein Zehntel der Breite** — enger, und die Ränder berühren sich;
// weiter, und ein Schlachtfeld mit zwanzig Permanenten passt nicht mehr auf
// eine Sitzfläche.
export const Kartenabstand = 6

// Sitzbreite ist die Breite, die ein Sitz einnimmt.
//
// **1200 mm ist die Breite eines Spieltisches.** Die Höhe steht unten bei
// [Sitzhoehe], weil sie aus den Bändern zusammengesetzt ist, die davor
// stehen müssen.
export const Sitzbreite = 1200

// Sitzabstand ist die Lücke zwischen zwei Sitzflächen.
//
// **Senkrecht steht dort der Stapel** — die Zone, die niemandem gehört
// (CR 405.1, Fassung 2026-08-19) und deshalb keiner Sitzfläche zufällt. Eine
// Karte ist 88 hoch; 140 lässt ihr oben und unten Luft.
//
// Seitlich trennt dieselbe Zahl zwei nebeneinanderliegende Sitze. Dort steht
// nichts — sie ist die Lücke und sonst nichts.
export const Sitzabstand = 140

// Ablagebreite ist die Spalte rechts, in der Bibliothek, Friedhof, Exil und
// Kommando als Stapel liegen.
//
// **Vier Stapel nebeneinander**: 4 × 63 mit 3 × 6 dazwischen = 270.
export const Ablagebreite = 4 * Kartenbreite + 3 * Kartenabstand

// Stapelbreite ist die Breite des MagicStack — Platz für sechs Karten:
// 6 × 63 mit 5 × 6 dazwischen = 408.
//
// **Hier und nicht in `tischplan.ts`**, wo sie bis ENG-03 stand
// (#1728):
// `tischgeometrie.ts` las sie von dort, und `tischplan.ts` liest
// `tischgeometrie.ts` — ein Einfuhrzyklus, den erst die Teilung vom
// 2026-09-22 geschaffen hatte. Ein Maß gehört zu den Maßen.
export const Stapelbreite = 6 * Kartenbreite + 5 * Kartenabstand

// Reihenhoehe ist die Höhe des Schlachtfeldteils eines Sitzes.
//
// **Zwei Reihen Permanente** — mehr passt auf 1200 mm Tischbreite ohnehin
// selten, und wer mehr hat, bekommt sie enger gesetzt statt einer dritten
// Reihe, die die Hand verdrängte.
export const Reihenhoehe = 2 * Kartenhoehe + Kartenabstand

// Randbreite ist der Abstand vom Rand einer Sitzfläche zu ihrem Inhalt.
export const Randbreite = 12

// Kopfhoehe ist das Band an der Außenkante eines Sitzes, in dem sein Kopf
// steht — Name, Leben, Gift, ob er am Zug ist (Artboard *Tisch*,
// `tsitz-kopf`).
//
// **Ein eigenes Band und kein Rand**, seit dem 2026-09-24: der Kopf hing
// bis dahin mit 16 Einheiten Schrift in einen Rand von 12 und lag auf dem
// Namen der ersten Zone. **Zweiunddreißig**, weil der Kopf in fester
// Bildschirmgröße geschrieben wird (`schildmass.ts`) und bei der
// Anfangskamera eines Fensters von 1.440 Punkten — rund ein halber Punkt je
// Einheit — sechzehn Punkte Band braucht.
export const Kopfhoehe = 32

// Namenshoehe ist die Leiste über den Karten einer Zone, in der ihr Name und
// ihre Anzahl stehen (Artboard *Tisch*, `tzone-name` und `tzone-zahl`).
//
// **Über den Karten und nicht auf ihnen**: eine Karte beginnt an der
// Oberkante ihres Feldes, und ein Name dort stünde auf der ersten Karte.
// Vierundzwanzig Einheiten tragen bei der Anfangskamera eine Schrift von zehn
// Punkten.
export const Namenshoehe = 24

// Sitzhoehe ist die Höhe, die ein Sitz einnimmt.
//
// **Die Bänder, von außen nach innen**: Rand, Kopf, Leiste und Hand, Lücke,
// Leiste und Schlachtfeld (zwei Reihen à 88 mit 6 dazwischen), Rand —
// 12 + 32 + 24 + 88 + 6 + 24 + 182 + 12 = 380. Bis zum 2026-09-24 waren es
// 300, ohne Kopfband und ohne Leisten.
export const Sitzhoehe =
  2 * Randbreite +
  Kopfhoehe +
  2 * Namenshoehe +
  Kartenhoehe +
  Kartenabstand +
  Reihenhoehe
