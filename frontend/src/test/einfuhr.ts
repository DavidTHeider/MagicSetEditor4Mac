// **Eine Einfuhr ohne Dublette** — DOM-03.AK-01
// (#1698).
//
// Die Vorschau, in der keine Zeile eine vorhandene trifft: dann geht die
// Einfuhr ohne Zwischenschritt durch, wie vor der Vorschau. **Ausdrücklich
// belegt und nicht als Vorgabe von `belegeApi`** — was ein Test nicht belegt,
// soll weiter mit Nennung der Methode scheitern (§3).
export function ohneDublette() {
  return Promise.resolve({
    vorschau: { posten: [], befunde: [] },
    fehler: null,
  })
}
