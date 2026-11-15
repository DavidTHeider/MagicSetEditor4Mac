// **Die Fehlerwege des Turnierassistenten** — DOM-04.AK-01
// (#1699).
//
// **Jeder Weg betreten** (§10): ein Format, das sich nicht laden lässt, ein
// Vorschlag, der ausbleibt, ein Auslosen, das scheitert — und ein Bye.
import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { assistent, buehne, melde, schritt } from '../test/turnierassistent'
import { raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { paarungsprobe } from './turnierprobe'
import type { Api } from '../api'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

const schweigt = {
  code: 'unbekannt' as const,
  schluessel: 'fehler.unbekannt',
  meldung: 'Der Bestand schweigt.',
}

// bisZurErstenRunde legt ein Turnier an, meldet zwei an und geht weiter.
async function bisZurErstenRunde(weitere: Partial<Api> = {}) {
  const doppel = buehne([], weitere)
  const bereich = await assistent()

  // **Auf das Feld warten und nicht auf die Region** — unter Last stand die
  // Region, bevor das Feld darin gezeichnet war (Vorprüfung am 2026-09-30).
  await userEvent.type(
    await within(bereich).findByLabelText(texte.name),
    'Freitag',
  )
  await userEvent.click(
    within(bereich).getByRole('button', { name: texte.weiter }),
  )
  await userEvent.click(
    within(bereich).getByRole('button', { name: texte.anlegen }),
  )
  await waitFor(() => {
    expect(schritt()).toContain(texte.schrittTeilnehmende)
  })
  await melde('Anna')
  await melde('Bert')
  await waitFor(() => {
    expect(screen.getByRole('button', { name: texte.weiter })).toBeEnabled()
  })
  await userEvent.click(screen.getByRole('button', { name: texte.weiter }))

  return doppel
}

describe('die Fehlerwege des Assistenten', () => {
  it('sagt am Format, wenn sich die Formate nicht laden lassen', async () => {
    buehne([], {
      Formate: () => Promise.resolve({ formate: [], fehler: schweigt }),
    })
    const bereich = await assistent()

    await waitFor(() => {
      expect(within(bereich).getByLabelText(texte.format)).toHaveAttribute(
        'aria-invalid',
        'true',
      )
    })
    expect(
      within(bereich).getByLabelText(texte.format),
    ).toHaveAccessibleDescription(
      expect.stringContaining('Der Bestand schweigt.'),
    )
  })

  it('zeigt einen Bye und einen Namen, den es nicht mehr gibt', async () => {
    await bisZurErstenRunde({
      TurnierPaarungsvorschlag: () =>
        Promise.resolve({
          paarungen: [
            paarungsprobe({ a: '20', b: '21' }),
            paarungsprobe({ a: 'x9', b: '' }),
          ],
          fehler: null,
        }),
    })

    const vorschlag = await screen.findByRole('list', {
      name: texte.vorschlagTitel,
    })
    expect(
      await within(vorschlag).findByText(`x9 — ${texte.bye}`),
    ).toBeInTheDocument()
  })

  it('bietet ein Wiederholen an, wenn der Vorschlag ausbleibt', async () => {
    await bisZurErstenRunde({
      TurnierPaarungsvorschlag: () =>
        Promise.resolve({ paarungen: [], fehler: schweigt }),
    })

    expect(await screen.findByText(/Der Bestand schweigt/)).toBeInTheDocument()
    expect(
      within(
        screen.getByRole('list', { name: texte.vorschlagTitel }),
      ).queryAllByRole('listitem'),
    ).toHaveLength(0)
  })

  it('bleibt bei der ersten Runde, wenn das Auslosen scheitert', async () => {
    await bisZurErstenRunde({
      TurnierRundeLegen: () =>
        Promise.resolve({ runde: null, fehler: schweigt }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.ersteRundeAuslosen }),
    )

    expect(await screen.findByText('Der Bestand schweigt.')).toBeInTheDocument()
    expect(schritt()).toContain(texte.schrittErsteRunde)
  })
})
