// **jsdom bekommt eine Breite** — Werkstück U2 von Phase 5.1
// (#1413).
//
// `listenmass.ts` gibt jsdom eine **Höhe**, damit das Sichtfenster der
// Kartenliste etwas zu rechnen hat. Der Arbeitsplatz braucht das Gegenstück:
// er klappt seinen Navigator unter 1.072 Punkten eigener Breite ein, und jsdom sagt
// für jedes Element `clientWidth === 0` — damit wäre jeder Arbeitsplatz im
// Test eng, und der Fall „breit genug" ließe sich nie betreten.
//
// **Dieselbe Form wie dort**: die Eigenschaft wird am Prototyp beschrieben
// und beim Aufräumen wieder in den Zustand davor versetzt.
export function gibJsdomEineBreite(breite: number): () => void {
  const vorher = Object.getOwnPropertyDescriptor(
    Element.prototype,
    'clientWidth',
  )

  Object.defineProperty(Element.prototype, 'clientWidth', {
    configurable: true,
    get() {
      return breite
    },
  })

  return () => {
    if (vorher === undefined) {
      delete (Element.prototype as unknown as Record<string, unknown>)
        .clientWidth
    } else {
      Object.defineProperty(Element.prototype, 'clientWidth', vorher)
    }
  }
}
