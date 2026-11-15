// **Die Suche im offenen Set, die Recherche daneben** — AP-11
// (#2368),
// Befund B-17, Entwurf `Ruhe09`.
//
// „Nur in diesem Set“ stand im offenen Set auf „Alle Sets“, und die
// KI-Recherche stand gleichrangig unter der Textsuche — zwei Aufgaben auf
// einer Seite. Jetzt beginnt die Suche im Set, aus dem sie kommt, und die
// Recherche geht hinter „Recherche …“ auf.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Suche } from './Suche'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten

function zeige(offenesSet?: string) {
  belegeApi({
    Sets: () =>
      Promise.resolve({
        sets: [
          {
            id: '7',
            name: 'Werkstatt',
            spiel: 'magic',
            stil: 'm15',
            herkunft: '',
            karten: 3,
            herkunftFehlt: false,
          },
          {
            id: '8',
            name: 'Zweites',
            spiel: 'magic',
            stil: 'm15',
            herkunft: '',
            karten: 1,
            herkunftFehlt: false,
          },
        ],
      }),
    Suchfelder: () => Promise.resolve({ felder: [] }),
  })

  render(
    <Sprachraum sprache="de">
      <Suche
        oeffnen={vi.fn()}
        {...(offenesSet === undefined ? {} : { offenesSet })}
      />
    </Sprachraum>,
  )
}

describe('der Umfang der Suche', () => {
  it('beginnt im offenen Set, „Alle Sets“ daneben', async () => {
    zeige('7')

    const wahl = await screen.findByRole('combobox', { name: texte.suche.set })
    expect(
      await screen.findByRole('option', { name: 'Werkstatt' }),
    ).toBeInTheDocument()
    expect(wahl).toHaveValue('7')
    expect(
      screen.getByRole('option', { name: texte.suche.alleSets }),
    ).toBeInTheDocument()
  })

  it('beginnt ohne offenes Set bei allen Sets', async () => {
    zeige()

    expect(
      await screen.findByRole('combobox', { name: texte.suche.set }),
    ).toHaveValue('')
  })
})

describe('die Recherche neben der Suche', () => {
  it('steht hinter „Recherche …“ nach dem Suchfeld und geht auf und zu', async () => {
    zeige('7')

    expect(
      screen.queryByRole('heading', { name: texte.recherche.rechercheTitel }),
    ).toBeNull()

    const knopf = screen.getByRole('button', { name: texte.suche.recherche })
    const feld = screen.getByRole('searchbox', { name: texte.suche.eingabe })
    expect(
      feld.compareDocumentPosition(knopf) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy()
    expect(knopf).toHaveAttribute('aria-expanded', 'false')

    await userEvent.click(knopf)
    expect(knopf).toHaveAttribute('aria-expanded', 'true')
    expect(
      screen.getByRole('heading', { name: texte.recherche.rechercheTitel }),
    ).toBeVisible()

    await userEvent.click(knopf)
    expect(knopf).toHaveAttribute('aria-expanded', 'false')
    expect(
      screen.queryByRole('heading', { name: texte.recherche.rechercheTitel }),
    ).toBeNull()
  })
})
