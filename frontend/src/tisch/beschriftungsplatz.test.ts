// **Wo die Beschriftung des Tischs Platz hat** — gefunden am 2026-09-24 an
// der Frage, ob der Tisch professionell aussieht: der Sitzkopf lag auf dem
// Namen der ersten Zone, und der Name einer Zone auf ihrer ersten Karte.
//
// # Wogegen hier geprüft wird
//
// **Dass jede Beschriftung ein eigenes Band hat**, wie im Artboard *Tisch*:
// der Sitzkopf an der Außenkante des Sitzes (`tsitz-kopf`), der Name einer
// Zone in einer Leiste über ihren Karten (`tzone-name`). Geprüft wird an
// einem Plan mit zwei Sitzen in beiden Richtungen und mit Karten in jeder
// Zone — ein leerer Tisch zeigte keine Karte, die ein Name verdecken könnte.
import { describe, expect, it } from 'vitest'

import { einObjekt, einePartie } from '../seiten/partieprobe'
import { Kopfhoehe, Namenshoehe } from './masse'
import { sitzschilder } from './sitzschilder'
import { tischplan } from './tischplan'
import { zonenrahmen } from './zonenfeld'
import { zoneAn } from './treffer'
import { zonenschilder } from './zonenschilder'

interface Rechteck {
  x: number
  y: number
  breite: number
  hoehe: number
}

const ueberschneiden = (eines: Rechteck, anderes: Rechteck) =>
  eines.x < anderes.x + anderes.breite &&
  anderes.x < eines.x + eines.breite &&
  eines.y < anderes.y + anderes.hoehe &&
  anderes.y < eines.y + eines.hoehe

// Eine offene Karte je Zone und Sitz — damit ein Name etwas verdecken könnte.
const zonen = ['hand', 'tisch', 'friedhof', 'exil']
const plan = tischplan(
  einePartie({
    objekte: ['sitz1', 'sitz2'].flatMap((sitz) =>
      zonen.map((zone) =>
        einObjekt({
          kennung: `${sitz}/${zone}/0`,
          beherrscher: sitz,
          eigner: sitz,
          zone,
          karte: 'Wald',
          name: 'Wald',
          geheim: false,
        }),
      ),
    ),
  }),
)

describe('der Platz der Beschriftung', () => {
  it('lässt über jeder Zone eine Leiste für ihren Namen', () => {
    for (const feld of plan.felder) {
      const rahmen = zonenrahmen(feld)

      expect(rahmen.y).toBe(feld.y - Namenshoehe)
      expect(rahmen.hoehe).toBe(feld.hoehe + Namenshoehe)
    }
  })

  it('legt keine Leiste auf eine fremde Zone', () => {
    for (const feld of plan.felder)
      for (const anderes of plan.felder)
        if (anderes !== feld)
          expect(ueberschneiden(zonenrahmen(feld), anderes)).toBe(false)
  })

  it('setzt den Namen in seine Leiste und nicht auf eine Karte', () => {
    const namen = zonenschilder(plan, (zone) => zone)

    for (const feld of plan.felder) {
      const name = namen.find(
        (eines) => eines.kennung === `${feld.sitz}:${feld.zone}:name`,
      )

      expect(name?.y).toBe(feld.y - Namenshoehe / 2)
      expect(name?.ankerY).toBe(0.5)
    }
  })

  it('legt keine Karte in eine Namensleiste', () => {
    expect(plan.lagen.length).toBeGreaterThan(0)

    for (const feld of plan.felder) {
      const leiste: Rechteck = {
        x: feld.x,
        y: feld.y - Namenshoehe,
        breite: feld.breite,
        hoehe: Namenshoehe,
      }

      for (const lage of plan.lagen)
        expect(ueberschneiden(leiste, lage)).toBe(false)
    }
  })

  it('gibt jedem Sitz an seiner Außenkante ein Band für den Kopf', () => {
    const koepfe = sitzschilder(plan, (sitz) => sitz)

    for (const flaeche of plan.sitze) {
      const kopf = koepfe.find(
        (eines) => eines.kennung === `${flaeche.sitz}:sitz:kopf`,
      )
      const band: Rechteck = {
        x: flaeche.x,
        y: (kopf?.y ?? 0) - Kopfhoehe / 2,
        breite: flaeche.breite,
        hoehe: Kopfhoehe,
      }

      expect(kopf?.ankerY).toBe(0.5)

      // **Außen**: oben am Sitz, der zur Mitte nach unten zeigt, unten am
      // eigenen, der nach oben zeigt.
      const mitte = flaeche.y + flaeche.hoehe / 2
      expect((kopf?.y ?? 0) < mitte).toBe(flaeche.zurMitte === 1)

      for (const feld of plan.felder.filter(
        (eines) => eines.sitz === flaeche.sitz,
      ))
        expect(ueberschneiden(band, zonenrahmen(feld))).toBe(false)
    }
  })

  // **Der Plan reicht über jede Leiste** — sonst passte die Kamera ihn so
  // ein, dass die oberste Leiste über den Rand ragt. Ein Tisch mit einem
  // einzigen Sitz legt den Stapel **über** ihn, und dessen Leiste läge dann
  // oberhalb des Plans.
  it('hält jede Leiste innerhalb des Plans, auch bei einem Sitz', () => {
    const allein = tischplan(
      einePartie({
        sitze: [einePartie().sitze[0]].filter((sitz) => sitz !== undefined),
      }),
    )

    for (const feld of [...plan.felder, ...allein.felder]) {
      const rahmen = zonenrahmen(feld)
      const umfang = allein.felder.includes(feld) ? allein : plan

      expect(rahmen.y).toBeGreaterThanOrEqual(0)
      expect(rahmen.y + rahmen.hoehe).toBeLessThanOrEqual(umfang.hoehe)
    }
  })

  // **Und wer auf die Leiste einer Zone zeigt, zeigt auf die Zone** — sie
  // ist ein Teil von ihr, und ein Ablegen dort ginge sonst ins Leere.
  it('rechnet die Leiste zu ihrer Zone', () => {
    for (const feld of plan.felder) {
      const ort = { x: feld.x + feld.breite / 2, y: feld.y - Namenshoehe / 2 }

      expect(zoneAn(plan, ort)).toEqual({ zone: feld.zone, sitz: feld.sitz })
    }
  })
})
