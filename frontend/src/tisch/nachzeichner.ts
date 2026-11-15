// **Ein Bild, das ankommt, muss gezeichnet werden** — und zwar einmal, auch
// wenn zwanzig im selben Augenblick ankommen.
//
// # Warum es ihn gibt
//
// **Der Tisch zeichnet auf Zuruf und nicht im Takt** (`pixinaht.ts`,
// „autoStart"): sechzig Bilder je Sekunde für einen stillstehenden Tisch
// wären der Preis für nichts. Eine Textur kommt aber **nach** dem Zeichnen
// an, das sie angefordert hat. Bis zum 2026-09-24 setzte der Zeichner sie
// und rief niemanden — die zuletzt angekommene blieb unsichtbar, bis etwas
// anderes neu zeichnen ließ (gemessen in WebKit, Set „Green": drei Karten mit
// geladenem Bild, drei Platzhalter auf dem Tisch).

// Nachzeichner nimmt Bitten an und zeichnet im nächsten Bild einmal.
export interface Nachzeichner {
  bitte: () => void

  // halte beendet ihn — nach dem Aufräumen zeichnet niemand mehr.
  halte: () => void
}

export function neuerNachzeichner(
  zeichne: () => void,
  planen: (lauf: () => void) => void,
): Nachzeichner {
  let vorgemerkt = false
  let gehalten = false

  return {
    bitte() {
      if (vorgemerkt || gehalten) return

      vorgemerkt = true
      planen(() => {
        vorgemerkt = false
        if (!gehalten) zeichne()
      })
    },

    halte() {
      gehalten = true
    },
  }
}
