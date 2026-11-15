// **Was jemand hinlegt, bleibt liegen** — PLAY-04.AK-03
// (#1712).
//
// # Der Befund
//
// **Heute liegt jede Karte auf einem gerechneten Platz.** `zonenlage.ts`
// legt Reihen, `schlachtfeldlagen` gruppiert nach CR 403.1 — und eine von
// Hand gewählte Lage gibt es **nirgends**. Das ist für die meisten das
// Richtige und für manche das Gegenteil: wer seine Kreaturen nach
// Angriffsreihen sortiert, will, dass es so bleibt.
//
// Das Kriterium verlangt: „Autoanordnung ist abschaltbar und berücksichtigt
// Anhänge und Gruppen; manuelle Positionen bleiben bei zulässigen Updates
// erhalten."
//
// # Was „aus" heißt und was nicht
//
// **Nicht „ohne Ordnung", sondern „ohne Nachordnung".** Eine neu gezogene
// Karte hat keine eigene Lage, und die Rechnung ist die einzige, die ihr
// einen Platz geben kann. „Aus" heißt: wo eine eigene Lage steht, gilt sie.
import { describe, expect, it } from 'vitest'

import { eigeneLagen } from './eigenelagen'

import type { Tischlage } from './tischplan'

function eineLage(teile: Partial<Tischlage>): Tischlage {
  return {
    kennung: 'k1',
    x: 100,
    y: 200,
    breite: 63,
    hoehe: 88,
    drehung: 0,
    hoehenrang: 1,
    verdeckt: false,
    zone: 'tisch',
    sitz: 'sitz-1',
    seite: 'vorne',
    karteID: '7',
    name: 'Wald',
    geheim: false,
    zaehler: {},
    haengtAn: '',
    haufen: '',
    untergelegt: false,
    art: 'karte',
    haufenzahl: 1,
    ...teile,
  }
}

describe('eigeneLagen', () => {
  // **Ohne eigene Lage ändert sich nichts** — der Steuerlauf (§10): sonst
  // sagte kein Lauf darüber, ob die Rechnung überhaupt noch gilt.
  it('lässt die Rechnung stehen, wo nichts abgelegt ist', () => {
    const gerechnet = [eineLage({})]

    expect(eigeneLagen(gerechnet, {})).toEqual(gerechnet)
  })

  it('setzt eine Karte auf ihre eigene Lage', () => {
    const [gelegt] = eigeneLagen([eineLage({})], { k1: { x: 10, y: 20 } })

    expect(gelegt?.x).toBe(10)
    expect(gelegt?.y).toBe(20)
  })

  // **Ein Haufen zieht mit** — CR 400.5 kennt keinen Haufen, aber wer ihn
  // verschiebt, verschiebt ihn ganz. Blieben die unteren Blätter liegen,
  // wäre der Haufen auseinandergerissen.
  it('nimmt den ganzen Haufen mit', () => {
    const gelegte = eigeneLagen(
      [
        eineLage({ kennung: 'a', haufen: 'a', x: 100, y: 200 }),
        eineLage({ kennung: 'b', haufen: 'a', x: 104, y: 204 }),
      ],
      { a: { x: 10, y: 20 } },
    )

    // **Der Versatz bleibt** — ein Haufen liegt versetzt, und das ist es,
    // woran man ihn als Haufen erkennt.
    expect(gelegte[0]?.x).toBe(10)
    expect(gelegte[1]?.x).toBe(14)
    expect(gelegte[1]?.y).toBe(24)
  })

  // **Und was daran hängt, zieht ebenso mit** — CR 303.4: eine Aura liegt
  // bei dem, was sie verzaubert. Bliebe sie stehen, sähe man nicht mehr,
  // woran sie hängt.
  it('nimmt mit, was an der Karte hängt', () => {
    const gelegte = eigeneLagen(
      [
        eineLage({ kennung: 'a', x: 100, y: 200 }),
        eineLage({ kennung: 'aura', haengtAn: 'a', x: 110, y: 210 }),
      ],
      { a: { x: 10, y: 20 } },
    )

    expect(gelegte[1]?.x).toBe(20)
    expect(gelegte[1]?.y).toBe(30)
  })

  // **Eine eigene Lage an einem Anhängsel schlägt die seines Wirts** — wer
  // eine Aura selbst hinlegt, meint es so.
  it('lässt eine eigene Lage am Anhängsel gelten', () => {
    const gelegte = eigeneLagen(
      [
        eineLage({ kennung: 'a', x: 100, y: 200 }),
        eineLage({ kennung: 'aura', haengtAn: 'a', x: 110, y: 210 }),
      ],
      { a: { x: 10, y: 20 }, aura: { x: 500, y: 600 } },
    )

    expect(gelegte[1]?.x).toBe(500)
    expect(gelegte[1]?.y).toBe(600)
  })

  // **Eine Lage für etwas, das nicht da ist, tut nichts.** Ein Objekt kann
  // die Zone gewechselt haben oder ganz weg sein; eine gespeicherte Lage
  // überlebt das und darf dann nichts anrichten (§8).
  it('übergeht eine Lage ohne Karte', () => {
    const gerechnet = [eineLage({})]

    expect(eigeneLagen(gerechnet, { gibtsnicht: { x: 1, y: 2 } })).toEqual(
      gerechnet,
    )
  })
})
