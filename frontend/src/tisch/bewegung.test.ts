// **Wie sich etwas bewegt** — Werkstück G3 von Phase 5.5
// (#1460).
import { describe, expect, it } from 'vitest'

import { bewegungErwuenscht, lohntBewegung, weich, zwischen } from './bewegung'

const ort = (x: number, y: number, drehung = 0) => ({ x, y, drehung })

describe('weich', () => {
  it('fängt bei null an und hört bei eins auf', () => {
    expect(weich(0)).toBe(0)
    expect(weich(1)).toBe(1)
  })

  it('ist in der Mitte auf halbem Weg', () => {
    expect(weich(0.5)).toBeCloseTo(0.5, 9)
  })

  it('beschleunigt am Anfang und bremst am Ende', () => {
    // **Langsam los, schnell in der Mitte, langsam an**: der erste Zehntel
    // legt weniger zurück als der mittlere.
    expect(weich(0.1)).toBeLessThan(0.1)
    expect(weich(0.9)).toBeGreaterThan(0.9)
  })

  it('bleibt in den Grenzen, auch wo der Anteil es nicht ist', () => {
    expect(weich(-5)).toBe(0)
    expect(weich(5)).toBe(1)
  })
})

describe('zwischen', () => {
  it('gibt am Anfang den Ausgangspunkt', () => {
    expect(zwischen(ort(10, 20), ort(110, 220), 0)).toEqual(ort(10, 20))
  })

  it('gibt am Ende das Ziel', () => {
    const angekommen = zwischen(ort(10, 20), ort(110, 220), 1)

    expect(angekommen.x).toBeCloseTo(110, 9)
    expect(angekommen.y).toBeCloseTo(220, 9)
  })

  it('liegt in der Mitte auf halbem Weg', () => {
    const halb = zwischen(ort(0, 0), ort(100, 200), 0.5)

    expect(halb.x).toBeCloseTo(50, 9)
    expect(halb.y).toBeCloseTo(100, 9)
  })

  // **Von 350° nach 10° sind es zwanzig Grad und nicht dreihundertvierzig.**
  it('nimmt bei der Drehung den kürzeren Weg', () => {
    expect(zwischen(ort(0, 0, 350), ort(0, 0, 10), 1).drehung).toBeCloseTo(
      370,
      6,
    )
    expect(zwischen(ort(0, 0, 10), ort(0, 0, 350), 1).drehung).toBeCloseTo(
      -10,
      6,
    )
  })

  it('dreht ein Tappen um genau einen Viertelkreis', () => {
    expect(zwischen(ort(0, 0, 0), ort(0, 0, 90), 1).drehung).toBeCloseTo(90, 6)
  })
})

describe('lohntBewegung', () => {
  it('sagt ja, wo sich etwas sichtbar bewegt', () => {
    expect(lohntBewegung(ort(0, 0), ort(50, 0))).toBe(true)
    expect(lohntBewegung(ort(0, 0), ort(0, 50))).toBe(true)
    expect(lohntBewegung(ort(0, 0, 0), ort(0, 0, 90))).toBe(true)
  })

  it('sagt nein bei einem Rundungsfehler', () => {
    // **Ein Ticker, der für einen halben Millimeter anspringt, kostet sechzig
    // Bilder für nichts.**
    expect(lohntBewegung(ort(0, 0), ort(0.5, 0.5))).toBe(false)
    expect(lohntBewegung(ort(0, 0), ort(0, 0))).toBe(false)
  })
})

describe('bewegungErwuenscht', () => {
  const fenster = (reduziert: boolean) => ({
    matchMedia: () => ({ matches: reduziert }),
  })

  it('bewegt, wo niemand widerspricht', () => {
    expect(bewegungErwuenscht(fenster(false), true)).toBe(true)
  })

  // **Wer `prefers-reduced-motion` setzt, hat einen Grund** — Schwindel,
  // Migräne, Aufmerksamkeit. Die Systemeinstellung ist die erste Abschaltung.
  it('bewegt nicht, wo das System es ablehnt', () => {
    expect(bewegungErwuenscht(fenster(true), true)).toBe(false)
  })

  it('bewegt nicht, wo die Anwendung es abschaltet', () => {
    expect(bewegungErwuenscht(fenster(false), false)).toBe(false)
  })

  it('bewegt, wo niemand gefragt werden kann', () => {
    // **Ein Fenster ohne `matchMedia` gibt es** — jedes Dokument, das nicht in
    // einem hängt. Dann gilt die Vorgabe der Anwendung.
    expect(bewegungErwuenscht(undefined, true)).toBe(true)
    expect(bewegungErwuenscht({}, true)).toBe(true)
  })

  // **Und die Einstellung der Anwendung** — UX-10.AK-03 (#1682). „Bewegung:
  // Immer wenig" steht am Wurzelelement, wie für die Oberfläche; bis zum
  // 2026-09-25 flogen die Karten trotzdem, wenn das System nichts sagte.
  it('bewegt nicht, wo die Einstellung „Immer wenig" sagt', () => {
    const mitEinstellung = (bewegung?: string) => ({
      ...fenster(false),
      document: {
        documentElement: {
          dataset: bewegung === undefined ? {} : { bewegung },
        },
      },
    })

    expect(bewegungErwuenscht(mitEinstellung('reduziert'), true)).toBe(false)
    expect(bewegungErwuenscht(mitEinstellung(), true)).toBe(true)
  })
})
