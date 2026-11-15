// **Was an einem Punkt liegt** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Wogegen hier geprüft wird
//
// **Dass ein Zeiger dasselbe trifft, was ein Auge dort sieht.** Der Plan
// weiß, was wo liegt; diese Rechnung dreht die Frage um. Sie ist die
// Voraussetzung für jedes Ziehen — und sie steht hier und nicht in der
// Grafiknaht, weil sie geprüft werden muss.
import { describe, expect, it } from 'vitest'

import { einObjekt, einePartie, einSitz } from '../seiten/partieprobe'
import { objektAn, zoneAn } from './treffer'
import { tischplan } from './tischplan'

const plan = tischplan(
  einePartie({
    fuer: 'a',
    sitze: [
      einSitz({ kennung: 'a', name: 'Ada', deckabdruck: 'x', karten: 9 }),
      einSitz({ kennung: 'b', name: 'Bo', deckabdruck: 'y', karten: 9 }),
    ],
    objekte: [
      einObjekt({
        kennung: 'bär',
        zone: 'tisch',
        eigner: 'a',
        beherrscher: 'a',
        geheim: false,
        name: 'Bär',
      }),
      einObjekt({
        kennung: 'wolf',
        zone: 'tisch',
        eigner: 'a',
        beherrscher: 'a',
        geheim: false,
        name: 'Wolf',
      }),
    ],
  }),
)

const lageVon = (kennung: string) => {
  const lage = plan.lagen.find((eine) => eine.kennung === kennung)
  if (!lage) throw new Error(`keine Lage für ${kennung}`)

  return lage
}

const mitteVon = (kennung: string) => {
  const lage = lageVon(kennung)

  return { x: lage.x + lage.breite / 2, y: lage.y + lage.hoehe / 2 }
}

describe('objektAn', () => {
  it('findet die Karte unter dem Zeiger', () => {
    expect(objektAn(plan, mitteVon('bär'))).toBe('bär')
    expect(objektAn(plan, mitteVon('wolf'))).toBe('wolf')
  })

  it('findet nichts auf leerem Tisch', () => {
    expect(objektAn(plan, { x: -50, y: -50 })).toBeUndefined()
  })

  // **Was oben liegt, wird getroffen.** Bei einem Stapel überlappen die
  // Karten; wer die unterste träfe, zöge etwas heraus, worauf er gar nicht
  // gezeigt hat.
  // **Die Reihenfolge im Plan ist nicht die Höhe.** `tischplan` gibt sie
  // heute aufsteigend, aber `objektAn` soll für **jeden** Plan stimmen —
  // sonst hinge seine Richtigkeit an einer Eigenschaft, die niemand zugesagt
  // hat.
  it('nimmt das oberste, auch wenn es zuerst in der Liste steht', () => {
    const verdreht = {
      ...plan,
      lagen: [
        { ...lageVon('bär'), hoehenrang: 9, x: 0, y: 0 },
        { ...lageVon('wolf'), hoehenrang: 1, x: 0, y: 0 },
      ],
    }

    expect(objektAn(verdreht, { x: 10, y: 10 })).toBe('bär')
  })

  it('nimmt das oberste von mehreren übereinander', () => {
    const stapel = tischplan(
      einePartie({
        fuer: 'a',
        sitze: [
          einSitz({ kennung: 'a', name: 'Ada', deckabdruck: 'x', karten: 3 }),
        ],
        objekte: [0, 1, 2].map((stelle) =>
          einObjekt({
            kennung: `k${String(stelle)}`,
            zone: 'friedhof',
            eigner: 'a',
            beherrscher: 'a',
            ordnung: stelle,
            geheim: false,
          }),
        ),
      }),
    )
    const oben = stapel.lagen[stapel.lagen.length - 1]
    if (!oben) throw new Error('kein Stapel')

    expect(objektAn(stapel, { x: oben.x + 5, y: oben.y + 5 })).toBe(
      oben.kennung,
    )
  })
})

describe('zoneAn', () => {
  it('nennt Zone und Sitz unter dem Zeiger', () => {
    const feld = plan.felder.find(
      (eines) => eines.zone === 'hand' && eines.sitz === 'a',
    )
    if (!feld) throw new Error('kein Handfeld')

    expect(zoneAn(plan, { x: feld.x + 10, y: feld.y + 10 })).toEqual({
      zone: 'hand',
      sitz: 'a',
    })
  })

  it('nennt nichts außerhalb jedes Feldes', () => {
    expect(zoneAn(plan, { x: -100, y: -100 })).toBeUndefined()
  })

  it('findet den Stapel, der keinem Sitz gehört', () => {
    const feld = plan.felder.find((eines) => eines.zone === 'stapel')
    if (!feld) throw new Error('kein Stapelfeld')

    expect(zoneAn(plan, { x: feld.x + 5, y: feld.y + 5 })).toEqual({
      zone: 'stapel',
      sitz: '',
    })
  })
})
