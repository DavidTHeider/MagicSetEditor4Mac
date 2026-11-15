// **Das Feld für den zweiten Faktor** — Werkstück 7 von Phase 4.12
// (#1351).
//
// **Für sich geprüft und nicht nur über die Seite**: was es ohne Eingabe tut
// und was die Eingabetaste tut, sähe dort niemand.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Nachweisfeld } from './Nachweisfeld'
import { Sprachraum } from '../sprache/kontext'

function zeige(abgeschickt: (nachweis: string) => void, angehalten = false) {
  render(
    <Sprachraum sprache="de">
      <Nachweisfeld angehalten={angehalten} abgeschickt={abgeschickt} />
    </Sprachraum>,
  )
}

describe('das Nachweisfeld', () => {
  // **Der Knopf ruht, solange nichts dasteht** — und der Ruf bleibt aus.
  it('ruft nichts, solange das Feld leer ist', async () => {
    const abgeschickt = vi.fn()

    zeige(abgeschickt)

    expect(screen.getByRole('button', { name: 'Bestätigen' })).toBeDisabled()
    expect(abgeschickt).not.toHaveBeenCalled()

    // **Und die Eingabetaste auf einem leeren Feld ebenso.** Der Beleg, dass
    // jemand zuhört, steht darunter: derselbe Weg mit einer Eingabe wirkt.
    await userEvent.type(screen.getByLabelText('Code'), '{Enter}')

    expect(abgeschickt).not.toHaveBeenCalled()
  })

  // **Die Eingabetaste schickt ab** (§6, Tastaturbedienbarkeit).
  it('schickt über die Eingabetaste ab', async () => {
    const abgeschickt = vi.fn()

    zeige(abgeschickt)

    await userEvent.type(screen.getByLabelText('Code'), '123456{Enter}')

    expect(abgeschickt).toHaveBeenCalledWith('123456')
  })

  // **Und der Knopf tut dasselbe — mit beschnittenen Rändern.**
  it('schickt über den Knopf ab und beschneidet die Ränder', async () => {
    const abgeschickt = vi.fn()

    zeige(abgeschickt)

    await userEvent.type(screen.getByLabelText('Code'), '  123456  ')
    await userEvent.click(screen.getByRole('button', { name: 'Bestätigen' }))

    expect(abgeschickt).toHaveBeenCalledWith('123456')
  })

  // **Angehalten heißt angehalten** — während ein Vorgang läuft, tippt
  // niemand weiter (§7, Regel 4).
  it('lässt sich anhalten', () => {
    zeige(vi.fn(), true)

    expect(screen.getByLabelText('Code')).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Bestätigen' })).toBeDisabled()
  })

  // **Der Browser füllt hier nicht aus.** Ein Einmalkennwort ist einmalig,
  // und ein gemerktes wäre bereits verbraucht.
  it('lässt den Browser nicht ausfüllen', () => {
    zeige(vi.fn())

    expect(screen.getByLabelText('Code')).toHaveAttribute(
      'autocomplete',
      'one-time-code',
    )
  })
})
