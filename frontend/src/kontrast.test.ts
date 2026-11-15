// **Kontrast gerechnet** — UX-09.AK-03
// (#1681):
// WCAG 2.2, relative Leuchtdichte und Verhältnis.
import { describe, expect, it } from 'vitest'

import { kontrast, lesbarePaare, lesbarkeitsbefunde } from './kontrast'

describe('kontrast', () => {
  it('rechnet Schwarz auf Weiß als 21 : 1 und Gleiches als 1 : 1', () => {
    expect(kontrast('#000000', '#ffffff')).toBeCloseTo(21, 5)
    expect(kontrast('#fff', '#ffffff')).toBeCloseTo(1, 5)
  })

  // **Gegengerechnet mit einer zweiten Umsetzung** (Python, derselbe
  // Formelweg): Tinte auf der Fläche von Pergament, ohne Faser — mit Faser
  // steht in `docs/KONTRASTE.md` 13,20.
  it('trifft die Zahl einer bekannten Paarung', () => {
    expect(kontrast('#241c14', '#fdf9f0')).toBeCloseTo(15.97, 2)
  })

  it('weist eine Farbe ab, die keine Hexzahl ist', () => {
    expect(kontrast('rot', '#ffffff')).toBeUndefined()
    expect(kontrast('#12345', '#ffffff')).toBeUndefined()
  })
})

describe('lesbarkeitsbefunde', () => {
  const pergament: Record<string, string> = {
    '--mse-grund': '#f2e8d5',
    '--mse-flaeche': '#fdf9f0',
    '--mse-text': '#241c14',
    '--mse-leise': '#5c4a34',
    '--mse-kante': '#786248',
    '--mse-marke': '#6b4423',
    '--mse-betont': '#e8dcc2',
    '--mse-ring': '#6b4423',
    '--mse-marketext': '#fdf9f0',
    '--mse-warnung': '#8c2f16',
  }
  const lies = (werte: Record<string, string>) => (merkmal: string) =>
    werte[merkmal] ?? ''

  it('findet an Pergament nichts', () => {
    expect(lesbarkeitsbefunde(lies(pergament))).toEqual([])
  })

  it('nennt ein unlesbares Paar mit Zahl und Schwelle', () => {
    const befunde = lesbarkeitsbefunde(
      lies({ ...pergament, '--mse-text': '#e0d4bb' }),
    )

    expect(befunde).toContainEqual(
      expect.objectContaining({
        vorn: '--mse-text',
        hinten: '--mse-grund',
        schwelle: 4.5,
      }),
    )
    expect(
      befunde.every((befund) => befund.verhaeltnis < befund.schwelle),
    ).toBe(true)
  })

  // **Ränder brauchen 3 : 1, Text 4,5 : 1** (WCAG 1.4.11 und 1.4.3).
  it('misst eine Kante gegen 3 : 1 und nicht gegen 4,5 : 1', () => {
    const befunde = lesbarkeitsbefunde(
      lies({ ...pergament, '--mse-kante': '#b8a88c' }),
    )

    expect(befunde).toContainEqual(
      expect.objectContaining({ vorn: '--mse-kante', schwelle: 3 }),
    )
  })

  // **Was nicht zu lesen ist, wird nicht geraten** — ein Merkmal ohne Wert
  // (etwa unter jsdom) lässt sein Paar aus.
  it('lässt ein Paar aus, dessen Farbe es nicht kennt', () => {
    expect(lesbarkeitsbefunde(() => '')).toEqual([])
  })
})

// **Was die Sperre bis zum Zweitblick über Phase 6.3 nicht maß** (S4) —
// die Farben der Meldungen, den Schleier über dem Fensterbild und den Ring
// auf der betonten Fläche.
describe('lesbarkeitsbefunde, jenseits der Merkmale', () => {
  const lies = (werte: Record<string, string>) => (merkmal: string) =>
    werte[merkmal] ?? ''

  // Das Beispiel des Zweitblicks: dunkle eigene Farben im hellen Bild. Alle
  // sechzehn Paare der Merkmale halten, und jede Fehlermeldung stand bei
  // 2,07 : 1 — denn ihr Rot folgt dem Erscheinungsbild, nicht den Merkmalen.
  const dunkelImHellen: Record<string, string> = {
    '--mse-grund': '#1e1b16',
    '--mse-flaeche': '#26221c',
    '--mse-text': '#f0e6d2',
    '--mse-leise': '#c8b89c',
    '--mse-kante': '#9b8768',
    '--mse-marke': '#e0b050',
    '--mse-ring': '#e0b050',
    '--mse-betont': '#2e2820',
    '--mse-marketext': '#1a150f',
    '--mse-warnung': '#f0a07a',
  }

  it('misst die Farben der Meldungen im gewählten Erscheinungsbild mit', () => {
    const hell = lesbarkeitsbefunde(lies(dunkelImHellen), { dunkel: false })
    const dunkel = lesbarkeitsbefunde(lies(dunkelImHellen), { dunkel: true })

    expect(hell).toContainEqual(
      expect.objectContaining({
        vorn: 'status:warnung',
        hinten: '--mse-grund',
      }),
    )
    expect(
      dunkel.filter((befund) => befund.vorn.startsWith('status:')),
    ).toEqual([])
  })

  // Weißer Grund und #767676 halten 4,54 : 1 — über einem schwarzen
  // Bildpunkt unter 88 % Schleier sind es 3,44.
  it('misst über einem Fensterbild gegen den Schleier über Schwarz und Weiß', () => {
    const weiss: Record<string, string> = {
      '--mse-grund': '#ffffff',
      '--mse-flaeche': '#ffffff',
      '--mse-betont': '#ffffff',
      '--mse-text': '#767676',
      '--mse-leise': '#595959',
      '--mse-kante': '#767676',
      '--mse-marke': '#6b4423',
      '--mse-ring': '#6b4423',
      '--mse-marketext': '#ffffff',
      '--mse-warnung': '#8c2f16',
    }

    expect(lesbarkeitsbefunde(lies(weiss), { dunkel: false })).toEqual([])
    const befunde = lesbarkeitsbefunde(lies(weiss), {
      dunkel: false,
      schleier: 0.88,
    })

    expect(befunde).toContainEqual(
      expect.objectContaining({
        vorn: '--mse-text',
        hinten: 'fensterbild:dunkel',
      }),
    )
    expect(
      befunde.find((befund) => befund.hinten === 'fensterbild:dunkel')
        ?.verhaeltnis,
    ).toBeCloseTo(3.44, 1)
  })

  // **Auch die Farben der Meldungen stehen auf dem Grund** — und damit unter
  // dem Schleier (Nachlauf über Phase 6.3, N4): ein Grund von #cecece hält
  // sie ohne Bild, über einem schwarzen Bildpunkt fällt die Warnung auf 3,6.
  it('misst die Farben der Meldungen auch unter dem Schleier', () => {
    const grau: Record<string, string> = {
      '--mse-grund': '#cecece',
      '--mse-flaeche': '#ffffff',
      '--mse-betont': '#ffffff',
      '--mse-text': '#000000',
    }

    expect(
      lesbarkeitsbefunde(lies(grau), { dunkel: false }).filter((befund) =>
        befund.vorn.startsWith('status:'),
      ),
    ).toEqual([])
    expect(
      lesbarkeitsbefunde(lies(grau), { dunkel: false, schleier: 0.9 }),
    ).toContainEqual(
      expect.objectContaining({
        vorn: 'status:warnung',
        hinten: 'fensterbild:dunkel',
      }),
    )
  })

  // **Was nicht zu lesen ist, wird auch unter dem Schleier nicht geraten.**
  it('lässt unter dem Schleier aus, was es nicht kennt', () => {
    expect(
      lesbarkeitsbefunde(() => '', { dunkel: false, schleier: 0.9 }),
    ).toEqual([])
  })

  it('misst den Ring auch auf der betonten Fläche', () => {
    expect(lesbarePaare).toContainEqual(['--mse-ring', '--mse-betont', 3])
  })
})
