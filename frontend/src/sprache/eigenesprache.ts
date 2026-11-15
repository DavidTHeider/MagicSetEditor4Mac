// **Eine eigene Übersetzung hat sich geändert** — LOC-02
// (#1688).
//
// Der Editor schreibt, das Fenster spricht: beide liegen weit auseinander,
// und das Fenster soll einen übernommenen Satz sehen, ohne dass jemand die
// Sprache neu wählt. **Ein Ereignis und kein Zustand** — wer es nicht hört,
// verliert nichts; er lädt beim nächsten Wechsel ohnehin.
const ziel = new EventTarget()

// eigeneSpracheGeaendert meldet, dass sich eine Übersetzung geändert hat.
export function eigeneSpracheGeaendert(kennung: string): void {
  ziel.dispatchEvent(new CustomEvent<string>('geaendert', { detail: kennung }))
}

// aufAenderung hört darauf und gibt zurück, wie man aufhört.
export function aufAenderung(nimm: (kennung: string) => void): () => void {
  const hoerer = (ereignis: Event) => {
    nimm((ereignis as CustomEvent<string>).detail)
  }

  ziel.addEventListener('geaendert', hoerer)

  return () => {
    ziel.removeEventListener('geaendert', hoerer)
  }
}
