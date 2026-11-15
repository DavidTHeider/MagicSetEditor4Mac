// **Wie weit ein Zug ist** — das Artboard *Tisch* zeigt es als Balken unter
// „Zug 7 · Kampf · Blocker", bei 58 %.
//
// # Wogegen hier geprüft wird
//
// **Gegen das Regelwerk**: zwölf Abschnitte in fester Reihenfolge — CR 500.1
// nennt die fünf Phasen, CR 501.1 die drei Schritte des Anfangs, CR 506.1 die
// fünf des Kampfs und CR 512.1 die zwei des Endes (Fassung 2026-09-25). Der
// Blocker ist der siebte, und sieben von zwölf sind die 58 % des Entwurfs.
import { describe, expect, it } from 'vitest'

import { zugfortschritt } from './zugfortschritt'

describe('zugfortschritt', () => {
  it('zählt den ersten Schritt als ersten von zwölf', () => {
    expect(zugfortschritt('anfang', 'enttappen')).toEqual({
      stelle: 1,
      von: 12,
    })
  })

  it('stellt den Blocker an die siebte Stelle — die 58 % des Entwurfs', () => {
    const stand = zugfortschritt('kampf', 'blocker')

    expect(stand).toEqual({ stelle: 7, von: 12 })
    expect(Math.round((7 / 12) * 100)).toBe(58)
  })

  it('kennt die Hauptphasen, die keinen Schritt haben', () => {
    expect(zugfortschritt('hauptVorKampf', '')).toEqual({ stelle: 4, von: 12 })
    expect(zugfortschritt('hauptNachKampf', '')).toEqual({
      stelle: 10,
      von: 12,
    })
  })

  it('endet mit dem Aufräumen', () => {
    expect(zugfortschritt('ende', 'aufraeumen')).toEqual({
      stelle: 12,
      von: 12,
    })
  })

  // **Ein Abschnitt, den der Kern später einführt, ist kein Fehler** (§8) —
  // aber auch kein Stand, den man zeichnen könnte.
  it('schweigt über einen Abschnitt, den es nicht kennt', () => {
    expect(zugfortschritt('kampf', 'gibtesnicht')).toBeUndefined()
    expect(zugfortschritt('', '')).toBeUndefined()
  })
})
