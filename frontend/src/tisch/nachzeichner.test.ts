// **Ein Bild, das ankommt, muss gezeichnet werden** — der Tisch zeichnet auf
// Zuruf und nicht im Takt (`pixinaht.ts`, „autoStart").
//
// # Wogegen hier geprüft wird
//
// **Gegen den Tisch vom 2026-09-24**, gemessen mit einer Mitschrift im
// Zeichner: drei Karten in Adas Hand hatten ihr Bild — geladen, gesetzt, das
// Sprite lebendig —, und auf dem Tisch standen drei Platzhalter. Nach dem
// Setzen rief niemand das Zeichnen; die zuletzt angekommene Textur blieb
// unsichtbar, bis irgendetwas anderes neu zeichnen ließ.
import { describe, expect, it, vi } from 'vitest'

import { neuerNachzeichner } from './nachzeichner'

// geplant sammelt, was der Planer für das nächste Bild vorgemerkt hat.
function planer() {
  const vorgemerkt: (() => void)[] = []

  return {
    vorgemerkt,
    planen: (lauf: () => void) => {
      vorgemerkt.push(lauf)
    },
    naechstesBild: () => {
      for (const lauf of vorgemerkt.splice(0)) lauf()
    },
  }
}

describe('der Nachzeichner', () => {
  it('zeichnet nach einer Bitte im nächsten Bild', () => {
    const zeichne = vi.fn()
    const takt = planer()
    const nach = neuerNachzeichner(zeichne, takt.planen)

    nach.bitte()

    expect(zeichne).not.toHaveBeenCalled()

    takt.naechstesBild()

    expect(zeichne).toHaveBeenCalledOnce()
  })

  // **Zwanzig Bilder, die im selben Augenblick ankommen, sind ein Zeichnen.**
  it('fasst Bitten bis zum nächsten Bild zusammen', () => {
    const zeichne = vi.fn()
    const takt = planer()
    const nach = neuerNachzeichner(zeichne, takt.planen)

    nach.bitte()
    nach.bitte()
    nach.bitte()

    expect(takt.vorgemerkt).toHaveLength(1)

    takt.naechstesBild()

    expect(zeichne).toHaveBeenCalledOnce()
  })

  it('nimmt nach dem Zeichnen wieder Bitten an', () => {
    const zeichne = vi.fn()
    const takt = planer()
    const nach = neuerNachzeichner(zeichne, takt.planen)

    nach.bitte()
    takt.naechstesBild()
    nach.bitte()
    takt.naechstesBild()

    expect(zeichne).toHaveBeenCalledTimes(2)
  })

  // **Nach dem Aufräumen zeichnet niemand mehr** — ein vorgemerktes Bild auf
  // einer zerstörten Leinwand wäre ein Aufruf ins Leere.
  it('zeichnet nach dem Anhalten nicht mehr', () => {
    const zeichne = vi.fn()
    const takt = planer()
    const nach = neuerNachzeichner(zeichne, takt.planen)

    nach.bitte()
    nach.halte()
    takt.naechstesBild()
    nach.bitte()

    expect(zeichne).not.toHaveBeenCalled()
    expect(takt.vorgemerkt).toHaveLength(0)
  })
})
