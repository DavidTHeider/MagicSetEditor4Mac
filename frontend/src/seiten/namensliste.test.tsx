// **1.644 Namen sind keine Einladung** — Werkstück 4 von Phase 3.3
// (#1164).
//
// # Woher die Zahl kommt
//
// Am laufenden Fenster gezählt, an einer Karte aus `Vanguard.mse-set` mit
// eingebundenem `m15-altered`: die Umgebung kennt 1.644 Namen. Als flache
// Reihe von Knöpfen ist das dieselbe Ratlosigkeit wie ein leeres Feld — nur
// mit mehr Scrollen.
//
// **Die Zahl ist eine Aussage über diese eine Karte**, nicht über „die
// Umgebung" (§10): ein Set ohne Stilpaket kommt auf gut neunzig. Was beide
// Fälle teilen, ist, dass niemand eine Liste dieser Länge liest.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Namensliste } from './Namensliste'

const texte = kataloge.de.seiten.karten.konsole

// viele baut so viele Namen, wie eine echte Vorlage mitbringt.
function viele(anzahl: number): string[] {
  return Array.from({ length: anzahl }, (_, lauf) => `name_${String(lauf)}`)
}

describe('Die Namensliste', () => {
  it('zeigt nicht alle Namen auf einmal und sagt es', () => {
    render(
      <Sprachraum sprache="de">
        <Namensliste namen={viele(1644)} einsetzen={vi.fn()} />
      </Sprachraum>,
    )

    // **Gekappt und nicht still gekappt** (§10): eine Liste, die einfach
    // aufhört, liest sich wie eine vollständige.
    expect(screen.getAllByRole('button').length).toBeLessThan(100)
    expect(screen.getByText(/1644/)).toBeInTheDocument()
  })

  it('grenzt auf das ein, was eingetippt wird', async () => {
    render(
      <Sprachraum sprache="de">
        <Namensliste
          namen={['card', 'card_style', 'set', 'length']}
          einsetzen={vi.fn()}
        />
      </Sprachraum>,
    )

    await userEvent.type(screen.getByLabelText(texte.namenFiltern), 'card')

    expect(screen.getByRole('button', { name: 'card' })).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'card_style' }),
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: 'set' }),
    ).not.toBeInTheDocument()
  })

  // **Ein Klick setzt den Namen ein.** Das ist der Zweck der Liste; ohne diese
  // Zusicherung prüften die Läufe darüber eine Anzeige und keine Bedienung.
  it('reicht den angeklickten Namen weiter', async () => {
    const eingesetzt = vi.fn()

    render(
      <Sprachraum sprache="de">
        <Namensliste namen={['length']} einsetzen={eingesetzt} />
      </Sprachraum>,
    )

    await userEvent.click(screen.getByRole('button', { name: 'length' }))

    expect(eingesetzt).toHaveBeenCalledWith('length')
  })

  // **Eine Eingrenzung ohne Treffer sagt es**, statt eine leere Fläche zu
  // zeigen (§6: leere Zustände erklären den nächsten Schritt).
  it('sagt es, wenn nichts übrig bleibt', async () => {
    render(
      <Sprachraum sprache="de">
        <Namensliste namen={['card', 'set']} einsetzen={vi.fn()} />
      </Sprachraum>,
    )

    await userEvent.type(screen.getByLabelText(texte.namenFiltern), 'zzz')

    expect(screen.getByText(texte.namenLeer)).toBeInTheDocument()
  })
})
