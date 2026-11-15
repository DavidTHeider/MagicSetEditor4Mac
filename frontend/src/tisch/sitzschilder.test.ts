// **Was an einem Sitz steht** — PLAY-03
// (#1711),
// gezeichnet als `tsitz-kopf` im Artboard `Tisch` des Entwurfs.
import { describe, expect, it } from 'vitest'

import { Sitzschildgroesse, sitzschilder } from './sitzschilder'
import { Kopfhoehe, Randbreite } from './masse'
import { einePartie } from '../seiten/partieprobe'
import { Schildgroesse } from './zonenschilder'
import { tischplan } from './tischplan'
import { zweiSitze } from '../test/planprobe'

const plan = tischplan(einePartie({ ...zweiSitze, objekte: [] }))
const wort = (sitz: string) => `${sitz} — 20 Leben`

describe('sitzschilder', () => {
  // **Je Sitz eines, und keines mehr.** Zwei Schilder an einem Platz wären
  // zwei Auskünfte über dieselbe Zahl.
  it('gibt je Sitz genau ein Schild', () => {
    const schilder = sitzschilder(plan, wort)

    expect(schilder).toHaveLength(plan.sitze.length)
    expect(new Set(schilder.map((eines) => eines.kennung)).size).toBe(
      schilder.length,
    )
  })

  it('schreibt hin, was das Wort sagt', () => {
    expect(sitzschilder(plan, wort).map((eines) => eines.text)).toContain(
      'ich — 20 Leben',
    )
  })

  // **Das Schild sitzt an der Außenkante des Sitzes** — so zeichnet es der
  // Entwurf: der Kopf des Gegenübers steht **über** seinen Zonen, der eigene
  // **unter** den eigenen. Zur Mitte hin stünde er zwischen zwei
  // Schlachtfeldern, und niemand wüsste, zu welchem er gehört. **Seit dem
  // 2026-09-24 in einem eigenen Band** (`masse.ts`, [Kopfhoehe]) — vorher
  // hing er in den Rand und lag auf dem Namen der ersten Zone.
  it('setzt das Schild in das Band an der Außenkante', () => {
    const schilder = sitzschilder(plan, wort)

    for (const flaeche of plan.sitze) {
      const schild = schilder.find((eines) =>
        eines.kennung.startsWith(`${flaeche.sitz}:`),
      )

      expect(schild?.y).toBe(
        flaeche.zurMitte === -1
          ? flaeche.y + flaeche.hoehe - Randbreite - Kopfhoehe / 2
          : flaeche.y + Randbreite + Kopfhoehe / 2,
      )
    }
  })

  // **Und es steht mittig in seinem Band** — oben wie unten dasselbe, seit
  // das Band nicht mehr der Rand ist.
  it('hängt das Schild mittig in sein Band', () => {
    const schilder = sitzschilder(plan, wort)

    for (const flaeche of plan.sitze) {
      const schild = schilder.find((eines) =>
        eines.kennung.startsWith(`${flaeche.sitz}:`),
      )

      expect(schild?.ankerY).toBe(0.5)
    }
  })

  // **Ein Sitzschild ist größer als ein Zonenschild.** Es nennt den Spieler
  // und seinen Stand; ein Zonenname sagt, was in einem Feld liegt. Gleich
  // groß wäre beides gleich wichtig.
  it('schreibt größer als ein Zonenschild', () => {
    expect(Sitzschildgroesse).toBeGreaterThan(Schildgroesse)

    for (const schild of sitzschilder(plan, wort)) {
      expect(schild.groesse).toBe(Sitzschildgroesse)
    }
  })
})
