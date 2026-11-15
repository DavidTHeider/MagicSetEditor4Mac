// **Was der Weg über dem Kopf zeigt und wohin er führt** — UX-03.AK-02
// (#1675).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Brotkrume } from './Brotkrume'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { zeichnetNichts } from '../test/leererbaum'

function zeige(krumen: Parameters<typeof Brotkrume>[0]['krumen']) {
  const wechseln = vi.fn()

  const { container } = render(
    <Sprachraum sprache="de">
      <Brotkrume krumen={krumen} wechseln={wechseln} />
    </Sprachraum>,
  )

  return { wechseln, container }
}

describe('Der Weg über dem Kopf', () => {
  // **Ein leerer Weg ist kein Orientierungspunkt.** Ein `<nav>` ohne Inhalt
  // zählt ein Vorleseprogramm auf, und darin steht nichts.
  it('zeichnet nichts, wo es keinen Weg gibt', () => {
    const { container } = zeige([])

    expect(zeichnetNichts(container)).toBe(true)
  })

  // **Die letzte Station ist, wo man steht** — und sie ist kein Knopf.
  it('führt die letzte Station ohne Knopf', () => {
    zeige([
      { beschriftung: 'Gestalten', ziel: 'sets' },
      { beschriftung: 'Stil' },
    ])

    expect(screen.queryByRole('button', { name: 'Stil' })).toBeNull()
    expect(screen.getByText('Stil')).toHaveAttribute('aria-current', 'page')
  })

  // **Nur die letzte Station ist, wo man steht.** Eine Station ohne Ziel
  // mitten im Weg — die Partie über dem Tisch, UX-03.AK-02 — ist kein Knopf,
  // denn sie führte auf dieselbe Seite; aber sie ist auch nicht die Seite.
  // Zwei `aria-current` sagten einem Vorleseprogramm, man stünde an zwei
  // Stellen zugleich.
  it('nennt nur die letzte Station die aktuelle', () => {
    zeige([
      { beschriftung: 'Spielen', ziel: 'decks' },
      { beschriftung: 'Partie gegen Bo' },
      { beschriftung: 'Am Tisch' },
    ])

    expect(screen.queryByRole('button', { name: 'Partie gegen Bo' })).toBeNull()
    expect(screen.getByText('Partie gegen Bo')).not.toHaveAttribute(
      'aria-current',
    )
    expect(screen.getByText('Am Tisch')).toHaveAttribute('aria-current', 'page')

    // **Und sie sieht auch nicht so aus** — halbfett ist, wo man steht. Die
    // WebKit-Probe zeigte am 2026-09-24 zwei halbfette Stationen hintereinander.
    expect(screen.getByText('Partie gegen Bo')).not.toHaveClass('font-semibold')
    expect(screen.getByText('Am Tisch')).toHaveClass('font-semibold')
  })

  it('führt auf die Seite einer Station', async () => {
    const { wechseln } = zeige([
      { beschriftung: 'Gestalten', ziel: 'sets' },
      { beschriftung: 'Stil' },
    ])

    await userEvent.click(screen.getByRole('button', { name: 'Gestalten' }))

    expect(wechseln).toHaveBeenCalledWith('sets')
  })

  // **Der Orientierungspunkt trägt einen Namen** — sonst hieße er „Navigation"
  // wie die Hauptnavigation daneben, und wer zwischen ihnen springt, weiß
  // nicht, wo er landet.
  it('nennt sich beim Namen', () => {
    zeige([{ beschriftung: 'Gestalten', ziel: 'sets' }])

    expect(
      screen.getByRole('navigation', {
        name: kataloge.de.allgemein.navigation.weg,
      }),
    ).toBeInTheDocument()
  })

  // **Zwei Stationen mit demselben Namen sind zwei Stationen** — auf
  // *Handeln* heißen Bereich und Seite gleich. Gesehen am 2026-09-24 in
  // WebKit: nach *Handeln* begann jeder Weg mit „Handeln", auch auf
  // *Gestalten › Karten*. Der Schlüssel war der Name, und React ließ beim
  // Wechsel eine der beiden gleich benannten Stationen stehen.
  it('lässt beim Wechsel keine gleichnamige Station stehen', () => {
    const wechseln = vi.fn()
    const { rerender } = render(
      <Sprachraum sprache="de">
        <Brotkrume
          krumen={[
            { beschriftung: 'Handeln', ziel: 'handeln' },
            { beschriftung: 'Handeln' },
          ]}
          wechseln={wechseln}
        />
      </Sprachraum>,
    )

    rerender(
      <Sprachraum sprache="de">
        <Brotkrume
          krumen={[
            { beschriftung: 'Gestalten', ziel: 'sets' },
            { beschriftung: 'Karten' },
          ]}
          wechseln={wechseln}
        />
      </Sprachraum>,
    )

    expect(
      screen.getAllByRole('listitem').map((station) => station.textContent),
    ).toEqual(['Gestalten', '›Karten'])
  })
})
