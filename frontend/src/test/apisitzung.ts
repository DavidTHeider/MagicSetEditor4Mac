// **Die Vorbelegung der Sitzung** — ID-02 (#1701): angemeldet. Jedes
// Zeichnen von `App` fragt sie über das Tor; wer die Sperre meint, belegt sie
// selbst.
export function angemeldeteSitzung() {
  return Promise.resolve({
    zustand: 'angemeldet' as const,
    nutzerID: '1',
    kennwort: false,
    sensor: false,
    fehler: null,
  })
}
