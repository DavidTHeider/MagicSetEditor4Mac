// **Das Schlachtfeld liegt in Gruppen** — PLAY-03.AK-02
// (#1711).
//
// Aus `tischplan.test.ts` herausgelöst, als die Datei über 400 Zeilen wuchs
// (§2). Der Schnitt folgt der Frage: dort steht, **was ein Plan ist** — Sitze,
// Felder, Lagen, Zahlen, Heimatlose; hier, **wie das Schlachtfeld darin
// geordnet** ist.
//
// **Der Lauf gehört zum Plan und nicht zu `schlachtfeld.test.ts`**: dort
// steht, dass die Rechnung stimmt, hier, dass der Plan sie **benutzt**. Eine
// Rechnung ohne Aufrufer ist genau der Befund, den §2 bei Paketen beschreibt
// — und der Tisch hatte ihn schon einmal (`Zonenzahl`, gerechnet und nie
// gezeichnet).
import { describe, expect, it } from 'vitest'

import { einePartie } from '../seiten/partieprobe'
import {
  lage,
  meines,
  mitTypen,
  sitzflaeche,
  zweiSitze,
} from '../test/planprobe'
import { tischplan } from './tischplan'

describe('das Schlachtfeld im Plan', () => {
  it('trennt Kreaturen und Länder in zwei Reihen', () => {
    const plan = tischplan(
      einePartie({
        ...zweiSitze,
        objekte: [mitTypen('k', ['creature']), mitTypen('l', ['land'])],
      }),
    )

    expect(lage(plan, 'k').y).not.toBe(lage(plan, 'l').y)
  })

  // **Und die Kreatur liegt näher an der Mitte.** Der eigene Sitz liegt unten,
  // der Stapel darüber — die Kreaturen gehören also nach oben.
  it('stellt die Kreaturen zur Tischmitte', () => {
    const plan = tischplan(
      einePartie({
        ...zweiSitze,
        objekte: [mitTypen('k', ['creature']), mitTypen('l', ['land'])],
      }),
    )

    const meineFlaeche = sitzflaeche(plan, 'ich')

    expect(lage(plan, 'k').y).toBeLessThan(lage(plan, 'l').y)
    expect(meineFlaeche.zurMitte).toBe(-1)
  })

  // **Die übrigen Zonen bleiben, wie sie waren.** Eine Hand ist keine
  // Permanentenreihe, und eine Gruppierung dort wäre eine Erfindung.
  it('lässt die Hand ungruppiert', () => {
    const plan = tischplan(
      einePartie({
        ...zweiSitze,
        objekte: [
          meines({ kennung: 'h1', zone: 'hand', ordnung: 0 }),
          meines({ kennung: 'h2', zone: 'hand', ordnung: 1 }),
        ],
      }),
    )

    expect(lage(plan, 'h1').y).toBe(lage(plan, 'h2').y)
  })
})
