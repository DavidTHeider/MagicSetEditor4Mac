// **Die Zeichenreihenfolge der Symbolfläche** — Werkstück 5 von Phase 3.7
// (#1168).
//
// **Beide Läufe hier haben einen Fehler gefunden, den kein Test sah** — die
// laufende Anwendung hat ihn gezeigt (§3). Sie stehen deshalb zusammen und
// nicht bei der Bedienung nebenan.
import { screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { raeumeApi } from '../test/api'
import { mitSymbol, teil, zeige } from './symbolgruppenprobe'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.symbol

// **Die Fläche ist eine benannte Gruppe und kein Bild** — G17 aus dem
// Zweitblick über Phase 6.3: als `role="img"` sind ihre Kinder für
// Hilfsmittel nur Darstellung, und die Punkte, die mit den Pfeiltasten
// gehen, verschwanden für VoiceOver darin.
it('zeigt Hilfsmitteln die Punkte als Knöpfe in einer Gruppe', async () => {
  mitSymbol()
  zeige()

  await screen.findByRole('button', { name: 'Dreieck — 3 Punkte' })

  const flaeche = screen.getByRole('group', { name: texte.zeichnung })
  expect(within(flaeche).getAllByRole('button').length).toBeGreaterThan(0)
  expect(screen.queryByRole('img', { name: texte.zeichnung })).toBeNull()
})

// **Die Zeichenreihenfolge auf der Fläche** — Werkstück 5, gefunden in der
// laufenden Anwendung (§3).
describe('Die Zeichenreihenfolge der Fläche', () => {
  // **Das erste Teil der Liste liegt im Bild oben.**
  //
  // Der Rasterer zeichnet rückwärts (`internal/render/symbolraster.go`,
  // Zeile 101), so wie MSE (`FOR_EACH_CONST_REVERSE`,
  // `src/render/symbol/viewer.cpp`, Zeile 206, Fassung `f9d9356` vom
  // 2021-07-06). SVG malt das Spätere über das Frühere — die Fläche muss die
  // Liste deshalb **rückwärts** ausgeben, sonst zeigt der Editor die
  // umgekehrte Überdeckung wie das gerenderte Bild.
  it('malt das erste Teil zuletzt', async () => {
    mitSymbol()
    zeige()

    await screen.findByRole('button', { name: 'Dreieck — 3 Punkte' })

    const flaeche = screen.getByRole('group', { name: texte.zeichnung })
    const pfade = [...flaeche.querySelectorAll('path')]
      .map((pfad) => pfad.getAttribute('d') ?? '')
      .filter((d) => d !== '')

    // Das Dreieck hat drei Punkte, das Rechteck vier: der letzte gemalte Pfad
    // hat drei Strecken, der erste vier.
    expect(pfade).toHaveLength(2)
    expect(zuege(pfade[1] ?? '')).toBe(3)
    expect(zuege(pfade[0] ?? '')).toBe(4)
  })

  // **Eine Gruppe hat keinen eigenen Umriss — ihre Kinder werden gemalt.**
  //
  // Ohne das war die Fläche nach dem Gruppieren **leer**: `alsPfad` gibt für
  // ein Teil ohne Punkte eine leere Zeichenkette, und eine Gruppe hat keine.
  it('malt die Kinder einer Gruppe', async () => {
    mitSymbol({
      Symbol: vi.fn().mockResolvedValue({
        teile: [teil('Gruppe', 'Gruppe', [teil('Dreieck'), teil('Rechteck')])],
        name: 'symbol1.mse-symbol',
        fehler: null,
      }),
    })

    zeige()

    await screen.findByRole('button', { name: 'Gruppe — 2 Teile' })

    const flaeche = screen.getByRole('group', { name: texte.zeichnung })
    const gemalt = [...flaeche.querySelectorAll('path')].filter(
      (pfad) => (pfad.getAttribute('d') ?? '') !== '',
    )

    expect(gemalt).toHaveLength(2)
  })

  // **Die Anfasser kommen nach allen Umrissen.**
  //
  // MSE zeichnet sie zuletzt — `drawEditingHints(dc)` nach den Teilen
  // (`src/render/symbol/viewer.cpp`, Zeilen 109–116, Fassung `f9d9356` vom
  // 2021-07-06). Seit die Fläche rückwärts malt, ist das keine Feinheit: ein
  // Anfasser im selben `<g>` wie sein Umriss verschwände unter der Füllung
  // des Teils, das **danach** gemalt wird.
  it('malt die Anfasser über alle Umrisse', async () => {
    mitSymbol()
    zeige()

    await screen.findByRole('button', { name: 'Dreieck — 3 Punkte' })

    const flaeche = screen.getByRole('group', { name: texte.zeichnung })
    const knoten = [...flaeche.querySelectorAll('path, circle')]

    const arten = knoten.map((knoten) => knoten.tagName)
    const letzterUmriss = arten.lastIndexOf('path')
    const ersterAnfasser = arten.indexOf('circle')

    expect(ersterAnfasser).toBeGreaterThan(letzterUmriss)
  })
})

// zuege zählt die Strecken eines Pfades — `L` und `C` je eine.
function zuege(pfad: string): number {
  return (pfad.match(/[LC]/g) ?? []).length
}
