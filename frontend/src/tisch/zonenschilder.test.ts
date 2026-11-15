// **Was an einem Zonenfeld steht** — PLAY-03.AK-06
// (#1711).
//
// Die Aufnahme `048-nativer-spieltisch.png` zeigt dünne Striche ohne Namen und
// ohne Zahl. `tischplan` **rechnet** die Zahl seit jeher aus und legt sie an
// eine Stelle — gezeichnet hat sie niemand.
import { describe, expect, it } from 'vitest'

import { einObjekt, einePartie, einSitz } from '../seiten/partieprobe'
import { tischplan } from './tischplan'
import {
  Kurzschildgroesse,
  Schildabstand,
  Schildgroesse,
  zonenschilder,
} from './zonenschilder'
import { Namenshoehe } from './masse'

// grossbuchstabe ist ein Probewort je Zone — hier zählt die Verdrahtung, nicht
// der Katalog.
const grossbuchstabe = (zone: string) => zone.toUpperCase()

// einPlan gibt einen Plan mit zwei Sitzen und den übergebenen Objekten.
const einPlan = (...teile: Parameters<typeof einObjekt>[0][]) =>
  tischplan(
    einePartie({
      sitze: [
        einSitz({ kennung: 'ich', name: 'Ich', deckabdruck: 'a', karten: 2 }),
        einSitz({ kennung: 'du', name: 'Du', deckabdruck: 'b', karten: 2 }),
      ],
      fuer: 'ich',
      objekte: teile.map((eines) =>
        einObjekt({ eigner: 'ich', beherrscher: 'ich', ...eines }),
      ),
    }),
  )

describe('zonenschilder', () => {
  it('gibt jedem Feld seinen Namen', () => {
    const plan = einPlan()
    const namen = zonenschilder(plan, grossbuchstabe).filter((eines) =>
      eines.kennung.endsWith(':name'),
    )

    expect(namen).toHaveLength(plan.felder.length)
    expect(namen.map((eines) => eines.text)).toContain('BIBLIOTHEK')
  })

  // **Der Name sitzt in der Leiste über seinem Feld** — seit dem 2026-09-24.
  // Bis dahin hing er von der Oberkante ins Feld und lag damit auf der
  // ersten Karte (`beschriftungsplatz.test.ts`).
  it('setzt den Namen in die Leiste über seinem Feld', () => {
    const plan = einPlan()
    const feld = plan.felder[0]

    if (!feld) throw new Error('kein Feld')

    const schild = zonenschilder(plan, grossbuchstabe).find(
      (eines) => eines.kennung === `${feld.sitz}:${feld.zone}:name`,
    )

    expect(schild).toMatchObject({
      x: feld.x + Schildabstand,
      y: feld.y - Namenshoehe / 2,
      ankerY: 0.5,
    })
  })

  // **Eine Zahl steht, wo `tischplan` eine führt — und nur dort.**
  //
  // **Die erste Fassung dieses Laufs war falsch**, und der Irrtum gehört
  // aufgeschrieben: sie verlangte für eine offene Zone gar keine Zahl und
  // prüfte dafür die **ganze** Liste. Die enthält aber auch die vier Ablagen
  // jedes Sitzes, und die sind Haufen — `tischplan` legt ihnen eine Zahl an,
  // auch bei null. Geprüft wird deshalb die **gemeinte** Zone.
  it('gibt der offenen Zone keine Zahl', () => {
    const plan = einPlan({ kennung: 'a', zone: 'tisch' })

    expect(
      zonenschilder(plan, grossbuchstabe).map((eines) => eines.kennung),
    ).not.toContain('ich:tisch:zahl')
  })

  // **Und ein Haufen bekommt seine, auch wenn er leer ist.** Eine Bibliothek
  // ohne Karten ist eine Auskunft — CR 104.3c entscheidet die Partie daran.
  it('nennt jede Zahl, die der Plan führt', () => {
    const plan = einPlan()
    const zahlen = zonenschilder(plan, grossbuchstabe).filter((eines) =>
      eines.kennung.endsWith(':zahl'),
    )

    expect(zahlen).toHaveLength(plan.zahlen.length)
    expect(zahlen.map((eines) => eines.kennung)).toContain(
      'ich:bibliothek:zahl',
    )
  })

  // **Und sie steht an ihrem Platz** — Befund (c): `tischplan` rechnet sie
  // seit jeher aus, und gezeichnet hat sie niemand. **Am rechten Ende der
  // Leiste**, wie `tzone-zahl` im Artboard; bis zum 2026-09-24 stand sie
  // über der Unterkante, zwischen den Karten.
  it('setzt die Zahl ans rechte Ende der Leiste ihres Feldes', () => {
    const plan = einPlan()
    // **Die Hand und nicht eine Ablage** — über einem Stapel sitzt die
    // Anzahl auf ihm (Lauf unten).
    const feld = plan.felder.find(
      (eines) => eines.sitz === 'du' && eines.zone === 'hand',
    )

    if (!feld) throw new Error('keine Hand')

    plan.zahlen.length = 0
    plan.zahlen.push({ zone: 'hand', sitz: 'du', x: 40, y: 90, zahl: 53 })

    expect(
      zonenschilder(plan, grossbuchstabe).find((eines) =>
        eines.kennung.endsWith(':zahl'),
      ),
    ).toEqual({
      kennung: 'du:hand:zahl',
      x: feld.x + feld.breite - Schildabstand,
      y: feld.y - Namenshoehe / 2,
      ankerX: 1,
      text: '53',
      groesse: Schildgroesse,
      ankerY: 0.5,
      art: 'zahl',
    })
  })

  // **Eine Zahl ohne Feld hat keine Leiste** — und damit keinen Platz. Sie
  // entfällt, statt an einer erfundenen Stelle zu stehen.
  it('übergeht eine Zahl, zu der es kein Feld gibt', () => {
    const plan = einPlan()

    plan.zahlen.length = 0
    plan.zahlen.push({ zone: 'nirgends', sitz: 'du', x: 40, y: 90, zahl: 7 })

    expect(
      zonenschilder(plan, grossbuchstabe).filter((eines) =>
        eines.kennung.endsWith(':zahl'),
      ),
    ).toEqual([])
  })

  // **Über einem Stapel steht die Kurzform, kleiner** — und die Anzahl
  // sitzt auf dem Stapel statt neben dem Namen. Zusammen passten Name und
  // Anzahl nicht in 63 Einheiten.
  it('beschriftet einen Stapel mit der Kurzform und legt die Anzahl darauf', () => {
    const plan = einPlan()
    const feld = plan.felder.find(
      (eines) => eines.sitz === 'du' && eines.zone === 'bibliothek',
    )

    if (!feld) throw new Error('keine Bibliothek')

    const schilder = zonenschilder(plan, (zone, kurz) =>
      kurz ? `${zone}.kurz` : zone,
    )

    expect(
      schilder.find((eines) => eines.kennung === 'du:bibliothek:name'),
    ).toMatchObject({
      text: 'bibliothek.kurz',
      groesse: Kurzschildgroesse,
      art: 'kurzname',
    })
    expect(
      schilder.find((eines) => eines.kennung === 'du:bibliothek:zahl'),
    ).toMatchObject({
      x: feld.x + feld.breite - Schildabstand,
      y: feld.y + feld.hoehe - Schildabstand,
      ankerX: 1,
      ankerY: 1,
    })
    expect(Kurzschildgroesse).toBeLessThan(Schildgroesse)
  })
})
