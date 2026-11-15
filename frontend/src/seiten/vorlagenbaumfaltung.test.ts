// **Wann der Vorlagenbaum faltet** — UXR-18
// (#1926),
// Regel R6 aus `docs/UIREVIEW.md`. Die Rechnung ohne DOM; wie der Baum sie
// zeichnet, prüft `vorlagenbaumgross.test.tsx`.
import { describe, expect, it } from 'vitest'

import { Faltgrenze, Stueck, plane } from './vorlagenbaumfaltung'

import type { Zweig } from '../haken/useVorlagenbaum'

function zweig(
  name: string,
  dateien: string[] = ['style'],
  beiwerk: string[] = [],
): Zweig {
  return {
    paket: `/pakete/${name}`,
    name,
    dateien,
    beiwerk,
    fehler: undefined,
    nurLesbar: false,
  }
}

// viele gibt `anzahl` Pakete mit je einer Datei — zwei Zeilen je Paket.
function viele(anzahl: number): Zweig[] {
  return Array.from({ length: anzahl }, (_, stelle) =>
    zweig(`p${String(stelle)}.mse-style`),
  )
}

function bilder(anzahl: number): string[] {
  return Array.from({ length: anzahl }, (_, stelle) => `b${String(stelle)}.png`)
}

const leer = { gewaehlt: null, klappwahl: {}, stuecke: {} }

describe('Die Faltung des Vorlagenbaums', () => {
  it('lässt einen kleinen Baum ganz offen, mit dem Beiwerk in der Reihe', () => {
    const plan = plane([zweig('a', ['style'], ['rahmen.png'])], leer)

    expect(plan.zweige).toHaveLength(1)
    expect(plan.zweige[0]).toMatchObject({
      offen: true,
      dateien: ['style'],
      beiwerkZeile: false,
      beiwerk: ['rahmen.png'],
      naechsteDatei: undefined,
      naechstesBild: undefined,
    })
    expect(plan.naechstesPaket).toBeUndefined()
  })

  it('hält eine Zeile unter der Grenze offen und faltet ab ihr', () => {
    // Je Paket zwei Zeilen: das Paket und seine Datei.
    const genau = plane(viele(Faltgrenze / 2), leer)
    expect(genau.zweige.every((z) => z.offen)).toBe(true)

    const darueber = plane(
      [...viele(Faltgrenze / 2 - 1), zweig('x', ['style', 'game'])],
      leer,
    )
    expect(darueber.zweige.every((z) => !z.offen)).toBe(true)
  })

  it('zählt das Beiwerk mit — ein Paket mit vielen Bildern faltet', () => {
    const plan = plane([zweig('a', ['style'], bilder(Faltgrenze))], leer)

    expect(plan.zweige[0]).toMatchObject({
      offen: false,
      beiwerkZeile: true,
      beiwerkOffen: false,
      beiwerk: [],
    })
  })

  it('öffnet im großen Baum das Paket der Wahl und sein Beiwerk, wenn sie Beiwerk ist', () => {
    const zweige = [zweig('a', ['style'], bilder(Faltgrenze)), ...viele(3)]

    const textwahl = plane(zweige, {
      ...leer,
      gewaehlt: { paket: '/pakete/a', datei: 'style' },
    })
    expect(textwahl.zweige[0]).toMatchObject({
      offen: true,
      dateien: ['style'],
      beiwerkZeile: true,
      beiwerkOffen: false,
    })
    expect(textwahl.zweige[1]?.offen).toBe(false)

    const bildwahl = plane(zweige, {
      ...leer,
      gewaehlt: { paket: '/pakete/a', datei: 'b150.png' },
    })
    expect(bildwahl.zweige[0]).toMatchObject({
      offen: true,
      beiwerkOffen: true,
      naechstesBild: 'b151.png',
    })
    // **Die Wahl steht immer da**, auch hinter dem ersten Stück.
    expect(bildwahl.zweige[0]?.beiwerk).toHaveLength(151)
    expect(bildwahl.zweige[0]?.beiwerk.at(-1)).toBe('b150.png')
  })

  it('folgt der Hand: zugeklappt bleibt zu, aufgeklappt bleibt offen', () => {
    const klein = plane(viele(2), {
      ...leer,
      klappwahl: { 'p:/pakete/p0.mse-style': false },
    })
    expect(klein.zweige.map((z) => z.offen)).toEqual([false, true])

    const gross = plane(viele(Faltgrenze), {
      ...leer,
      gewaehlt: { paket: '/pakete/p1.mse-style', datei: 'style' },
      klappwahl: {
        'p:/pakete/p0.mse-style': true,
        'p:/pakete/p1.mse-style': false,
      },
    })
    expect(gross.zweige.slice(0, 3).map((z) => z.offen)).toEqual([
      true,
      false,
      false,
    ])
  })

  it('öffnet Beiwerk auf Wunsch und zeigt es stückweise', () => {
    const plan = plane([zweig('a', ['style'], bilder(Stueck * 3))], {
      ...leer,
      klappwahl: { 'p:/pakete/a': true, 'b:/pakete/a': true },
      stuecke: { 'b:/pakete/a': Stueck * 2 },
    })

    expect(plan.zweige[0]?.beiwerk).toHaveLength(Stueck * 2)
    expect(plan.zweige[0]?.naechstesBild).toBe(`b${String(Stueck * 2)}.png`)
  })

  it('zeigt Pakete und Dateien stückweise', () => {
    const plan = plane(
      [
        zweig(
          'gross',
          bilder(Stueck + 5).map((b) => b.slice(0, -4)),
        ),
        ...viele(Stueck * 2),
      ],
      { ...leer, klappwahl: { 'p:/pakete/gross': true } },
    )

    expect(plan.zweige).toHaveLength(Stueck)
    expect(plan.naechstesPaket).toBe(`/pakete/p${String(Stueck - 1)}.mse-style`)
    expect(plan.zweige[0]?.dateien).toHaveLength(Stueck)
    expect(plan.zweige[0]?.naechsteDatei).toBe(`b${String(Stueck)}`)

    const weiter = plane(viele(Stueck * 2), {
      ...leer,
      stuecke: { pakete: Stueck * 2 },
    })
    expect(weiter.zweige).toHaveLength(Stueck * 2)
    expect(weiter.naechstesPaket).toBeUndefined()
  })

  it('zieht das Paket der Wahl ins Stück, auch wenn es weit hinten steht', () => {
    const plan = plane(viele(Stueck * 3), {
      ...leer,
      gewaehlt: { paket: '/pakete/p250.mse-style', datei: 'style' },
    })

    expect(plan.zweige).toHaveLength(251)
    expect(plan.zweige.at(-1)).toMatchObject({ offen: true })
    expect(plan.naechstesPaket).toBe('/pakete/p251.mse-style')
  })
})
