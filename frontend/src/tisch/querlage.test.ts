// **Eine getappte Karte liegt quer** — CR 701.26a (Fassung 2026-09-25):
// tappen heißt, sie seitwärts zu drehen.
//
// # Wogegen hier geprüft wird
//
// **Gegen den Tisch vom 2026-09-24**, gesehen in WebKit mit dem Set „Green":
// zwei getappte Karten auf Adas Schlachtfeld überdeckten einander und ragten
// links über den Rand ihres Feldes. Die Reihe gab jeder Karte den Platz
// einer **aufrechten**, gedreht wurde um die Mitte — quer ist eine Karte aber
// so breit, wie sie aufrecht hoch ist.
//
// **Und dieselbe Lage las jeder, der wissen wollte, wo die Karte ist**: der
// Zeiger traf über und unter ihr ins Leere und an ihren Enden daneben, der
// Auswahlrahmen stand hochkant um eine liegende Karte, und die Marken
// schwebten unter ihr.
import { describe, expect, it } from 'vitest'

import { Kartenabstand, Kartenbreite, Kartenhoehe } from './masse'
import { an, erstes } from '../test/erstes'
import { einObjekt, einePartie, einSitz } from '../seiten/partieprobe'
import { kartenecken, markenpunkte, rahmenUm } from './weltbild'
import { haufenschilder } from './haufenschilder'
import { objektAn } from './treffer'
import { Schildabstand } from './zonenschilder'
import { sitzflaechen } from './sitzflaeche'
import { tischplan } from './tischplan'
import { Tappdrehung, umriss, zonenlagen } from './zonenlage'
import { zonenfelder } from './zonenfeld'

import type { Gelegtes } from './zonenlage'

const schlachtfeld = () => {
  const eines = zonenfelder(erstes(sitzflaechen(['ich'], 'ich'))).find(
    (jedes) => jedes.zone === 'tisch',
  )
  if (!eines) throw new Error('kein Schlachtfeld')

  return eines
}

const gelegt = (kennung: string, getappt: boolean): Gelegtes => ({
  kennung,
  getappt,
  verdeckt: false,
  haufen: '',
})

// **Ein Tausendstel Millimeter**: die Drehung rechnet über Sinus und
// Kosinus, und 90 Grad sind im Bogenmaß keine glatte Zahl.
const genau = 3

describe('der Umriss einer Lage', () => {
  it('ist quer so breit, wie die Karte aufrecht hoch ist — um ihre Mitte', () => {
    const quer = umriss({
      x: 10,
      y: 20,
      breite: Kartenbreite,
      hoehe: Kartenhoehe,
      drehung: Tappdrehung,
    })

    expect(quer.breite).toBeCloseTo(Kartenhoehe, genau)
    expect(quer.hoehe).toBeCloseTo(Kartenbreite, genau)
    expect(quer.x + quer.breite / 2).toBeCloseTo(10 + Kartenbreite / 2, genau)
    expect(quer.y + quer.hoehe / 2).toBeCloseTo(20 + Kartenhoehe / 2, genau)
  })

  it('ist aufrecht die Lage selbst', () => {
    expect(
      umriss({
        x: 1,
        y: 2,
        breite: Kartenbreite,
        hoehe: Kartenhoehe,
        drehung: 0,
      }),
    ).toEqual({ x: 1, y: 2, breite: Kartenbreite, hoehe: Kartenhoehe })
  })
})

describe('eine Reihe mit getappten Karten', () => {
  it('lässt jede im Feld', () => {
    const feld = schlachtfeld()
    const { lagen } = zonenlagen(feld, [gelegt('a', true), gelegt('b', true)])

    for (const lage of lagen) {
      const sichtbar = umriss(lage)

      expect(sichtbar.x).toBeGreaterThanOrEqual(feld.x - 1e-6)
      expect(sichtbar.x + sichtbar.breite).toBeLessThanOrEqual(
        feld.x + feld.breite + 1e-6,
      )
    }
  })

  it('und legt Nachbarinnen mit vollem Abstand, solange Platz ist', () => {
    const { lagen } = zonenlagen(schlachtfeld(), [
      gelegt('a', false),
      gelegt('b', true),
      gelegt('c', true),
      gelegt('d', false),
    ])

    for (const stelle of [0, 1, 2]) {
      const links = umriss(an(lagen, stelle))
      const rechts = umriss(an(lagen, stelle + 1))

      expect(rechts.x - (links.x + links.breite)).toBeCloseTo(
        Kartenabstand,
        genau,
      )
    }
  })

  // **Eine einzelne quere Karte in einem Feld, das schmaler ist als sie**
  // — eine Gruppe des Schlachtfelds behält mindestens eine Kartenbreite
  // (`schlachtfeld.ts`), und quer ist eine Karte breiter. Enger rücken kann
  // eine einzelne nicht; sie beginnt am linken Rand.
  it('legt eine einzelne quere Karte an den linken Rand eines zu schmalen Feldes', () => {
    const feld = { ...schlachtfeld(), breite: Kartenbreite }
    const { lagen } = zonenlagen(feld, [gelegt('a', true)])

    expect(umriss(an(lagen, 0)).x).toBeCloseTo(feld.x, genau)
  })

  it('rückt enger, wo sie nicht nebeneinander passen — und bleibt im Feld', () => {
    const feld = schlachtfeld()
    const { lagen } = zonenlagen(
      feld,
      Array.from({ length: 40 }, (_, stelle) =>
        gelegt(`k${String(stelle)}`, true),
      ),
    )

    for (const lage of lagen) {
      const sichtbar = umriss(lage)

      expect(sichtbar.x).toBeGreaterThanOrEqual(feld.x - 1e-6)
      expect(sichtbar.x + sichtbar.breite).toBeLessThanOrEqual(
        feld.x + feld.breite + 1e-6,
      )
    }
  })
})

// **Der Plan einer Partie mit einer getappten Karte** — für alles, was die
// Lage danach liest.
const plan = tischplan(
  einePartie({
    fuer: 'a',
    sitze: [
      einSitz({ kennung: 'a', name: 'Ada', karten: 9 }),
      einSitz({ kennung: 'b', name: 'Bo', karten: 9 }),
    ],
    objekte: [
      einObjekt({
        kennung: 'bär',
        zone: 'tisch',
        eigner: 'a',
        beherrscher: 'a',
        geheim: false,
        name: 'Bär',
        getappt: true,
        zaehler: { '+1/+1': 2 },
      }),
    ],
  }),
)

const baer = () => {
  const lage = plan.lagen.find((eine) => eine.kennung === 'bär')
  if (!lage) throw new Error('keine Lage für den Bären')

  return lage
}

describe('wer die Lage einer getappten Karte liest', () => {
  it('trifft sie an ihren Enden', () => {
    const sichtbar = umriss(baer())
    const mitteY = sichtbar.y + sichtbar.hoehe / 2

    expect(objektAn(plan, { x: sichtbar.x + 2, y: mitteY })).toBe('bär')
    expect(
      objektAn(plan, { x: sichtbar.x + sichtbar.breite - 2, y: mitteY }),
    ).toBe('bär')
  })

  it('und nicht über ihr, wo aufrecht ihr Kopf wäre', () => {
    const lage = baer()

    expect(
      objektAn(plan, { x: lage.x + lage.breite / 2, y: lage.y + 2 }),
    ).toBeUndefined()
  })

  it('legt die Marken an ihren sichtbaren unteren Rand', () => {
    const sichtbar = umriss(baer())
    const marken = erstes(markenpunkte(plan)).marken

    for (const marke of marken) {
      expect(marke.y).toBeLessThanOrEqual(sichtbar.y + sichtbar.hoehe)
      expect(marke.y).toBeGreaterThan(sichtbar.y + sichtbar.hoehe / 2)
    }
  })
})

describe('was an einer getappten Karte hängt', () => {
  it('rahmt sie quer ein', () => {
    const sichtbar = umriss(baer())

    expect(rahmenUm(plan, 'bär', 2)).toEqual({
      x: sichtbar.x - 2,
      y: sichtbar.y - 2,
      breite: sichtbar.breite + 4,
      hoehe: sichtbar.hoehe + 4,
    })
  })

  it('und rahmt nichts, was nicht daliegt', () => {
    expect(rahmenUm(plan, 'weg', 2)).toBeUndefined()
  })

  // **Die Zahl eines Haufens steht an der sichtbaren Ecke seiner obersten
  // Karte** — bei einer queren ist das nicht die Ecke der aufrechten Lage.
  it('schreibt die Zahl eines queren Haufens an seine sichtbare Ecke', () => {
    const haufen = tischplan(
      einePartie({
        fuer: 'a',
        sitze: [einSitz({ kennung: 'a', name: 'Ada', karten: 9 })],
        objekte: ['w1', 'w2'].map((kennung) =>
          einObjekt({
            kennung,
            zone: 'tisch',
            eigner: 'a',
            beherrscher: 'a',
            geheim: false,
            name: 'Wald',
            getappt: true,
            haufen: 'w',
          }),
        ),
      }),
    )
    const oberste = haufen.lagen.reduce((eine, andere) =>
      andere.hoehenrang > eine.hoehenrang ? andere : eine,
    )
    const sichtbar = umriss(oberste)
    const schild = erstes(haufenschilder(haufen))

    expect(schild.x).toBeCloseTo(
      sichtbar.x + sichtbar.breite - Schildabstand,
      genau,
    )
    expect(schild.y).toBeCloseTo(
      sichtbar.y + sichtbar.hoehe - Schildabstand,
      genau,
    )
  })
})

// **Der Rand liegt, wie die Karte liegt** — gesehen am 2026-09-24 in WebKit:
// die getappten Karten lagen quer, ihr Rand stand hochkant um sie herum.
describe('der Rand einer Karte', () => {
  const ausdehnung = (drehung: number) => {
    const ecken = kartenecken({
      x: 100,
      y: 200,
      breite: Kartenbreite,
      hoehe: Kartenhoehe,
      drehung,
    })
    const xe = ecken.map((ecke) => ecke.x)
    const ye = ecken.map((ecke) => ecke.y)

    return {
      ecken: ecken.length,
      breite: Math.max(...xe) - Math.min(...xe),
      hoehe: Math.max(...ye) - Math.min(...ye),
    }
  }

  it('liegt quer um eine getappte Karte', () => {
    const quer = ausdehnung(Math.PI / 2)

    expect(quer.ecken).toBe(4)
    expect(quer.breite).toBeCloseTo(Kartenhoehe, genau)
    expect(quer.hoehe).toBeCloseTo(Kartenbreite, genau)
  })

  it('und aufrecht um eine aufrechte', () => {
    const aufrecht = ausdehnung(0)

    expect(aufrecht.breite).toBeCloseTo(Kartenbreite, genau)
    expect(aufrecht.hoehe).toBeCloseTo(Kartenhoehe, genau)
  })
})
