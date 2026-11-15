// **Ein Zurückspielen räumt, was die verlassenen Seiten aufhoben** — Befund
// M9 (a) des Zweitblicks über Phase 6.3, verlagert nach Phase 6.5
// (#1666).
//
// # Der Befund
//
// Seit UX-03.AK-02 bleiben besuchte Seiten in einer `<Activity>` stehen
// (Vermerk 0133). Ein Zurückspielen tauscht den **ganzen** Bestand —
// `useBestandspflege` nennt es Regel 1 aus §7 „in ihrer größten Form" —, und
// geräumt wurden Set und Karte, aber nicht die aufgehobenen Seiten: ein halb
// getippter Deckname aus dem ersetzten Bestand stand danach wieder da.
//
// # Warum über der ganzen Anwendung
//
// **Dieselbe Begründung wie beim Profilwechsel** (`profilwechsel.test.tsx`,
// „vergisst, was die verlassenen Seiten aufhoben"): die Zusage gilt über den
// Seiten, und nur ein Lauf, der sie alle aufbaut, sieht sie.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { App } from './App'
import { zuGebiet } from './test/gebiete'
import { kataloge } from './sprache/kataloge'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'
import { leisteneintrag } from './test/leistenweg'

import type { Sicherungsauswahl } from './api'

afterEach(() => {
  raeumeApi()
})

const worte = kataloge.de.seiten.bestandspflege
const decks = kataloge.de.seiten.decks

const eineSicherung: Sicherungsauswahl = {
  pfad: '/Users/probe/mse-sicherung.msesicherung',
  angelegtAm: '2026-08-01T09:12:33.000Z',
  schemastand: '0012_paketkurzname.sql',
  sets: 3,
  karten: 107,
  bilder: 'ersetzt',
  abgebrochen: false,
  verschluesselt: false,
  paketeEnthalten: [],
  paketeFehlen: [],
}

describe('ein Zurückspielen', () => {
  it('vergisst, was die verlassenen Seiten aufhoben', async () => {
    belegeApi({
      ...belegeFenster('de'),
      Decks: () => Promise.resolve({ decks: [], fehler: null }),
      Formate: () => Promise.resolve({ formate: [], fehler: null }),
      Bauarten: () => Promise.resolve({ bauarten: [], fehler: null }),
      SicherungWaehlen: () => Promise.resolve(eineSicherung),
      SicherungZurueckspielen: () =>
        Promise.resolve({
          ersetzt: true,
          beiseite: '/Users/probe/vorher.db',
          unveraendert: false,
        }),
    })

    render(<App />)

    await userEvent.click(await leisteneintrag('decks'))
    await userEvent.click(
      await screen.findByRole('button', { name: decks.neu }),
    )
    await userEvent.type(
      await screen.findByRole('textbox', { name: decks.name }),
      'Aus dem alten Bestand',
    )

    await userEvent.click(await leisteneintrag('einstellungen'))
    await zuGebiet(kataloge.de.seiten.einstellungen.gruppen.daten)
    await userEvent.click(
      await screen.findByRole('button', { name: worte.waehlen }),
    )
    await userEvent.click(
      await screen.findByRole('button', { name: worte.bestaetigen }),
    )

    // **Die Seite, auf der zurückgespielt wurde, bleibt stehen** — mit dem
    // Ort des alten Bestands, den sonst niemand mehr nennt.
    expect(await screen.findByText(/vorher\.db/)).toBeInTheDocument()

    await userEvent.click(await leisteneintrag('decks'))

    // **Die Deckseite beginnt von vorn**: der Kasten steht zu, und wer ihn
    // aufklappt, findet ein leeres Feld.
    const neu = await screen.findByRole('button', { name: decks.neu })
    expect(neu).toHaveAttribute('aria-expanded', 'false')

    await userEvent.click(neu)

    await waitFor(() => {
      expect(screen.getByRole('textbox', { name: decks.name })).toHaveValue('')
    })
  })
})
