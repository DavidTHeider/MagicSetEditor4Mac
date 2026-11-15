// **Marken und gespeicherte Filter in der Sammlung** — UX-12
// (#1684),
// Artboard *Sammlung*.
import { keineHerkunft } from '../apiherkunft'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import type { ExemplarDTO } from '../api'
import { Sammlung } from './Sammlung'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.bausteine.objektmarken

function exemplar(id: number, bezeichnung: string): ExemplarDTO {
  return {
    id,
    fassung: 1,
    printing: `p${String(id)}`,
    druckID: '',
    vorlaeufig: false,
    menge: 1,
    zustand: 'nearMint',
    sprache: 'en',
    finish: 'nonfoil',
    erwerbsdatum: '',
    erwerbspreis: '',
    erwerbswaehrung: '',
    notiz: '',
    name: bezeichnung,
    setcode: '',
    nummer: '',
    lagerort: '',
    wunsch: false,
    bezeichnung,
    herkunft: keineHerkunft,
  }
}

const zwei = [exemplar(31, 'Aschenwächter'), exemplar(32, 'Bernsteinruf')]

it('filtert nach Marken über Go und setzt eine Marke im Inspector', async () => {
  const filtern = vi.fn().mockResolvedValue({ kennungen: ['31'], fehler: null })
  const setzen = vi.fn().mockResolvedValue({ fehler: null })
  const einstellungSetzen = vi.fn().mockResolvedValue({ fehler: null })
  const je = vi
    .fn()
    .mockResolvedValue({ marken: { '31': ['tausch'] }, fehler: null })

  belegeApi({
    Sammlungsfenster: () =>
      Promise.resolve({
        exemplare: zwei,
        gesamt: 2,
        insgesamt: 2,
        stueck: 2,
        fehler: null,
      }),
    Portfolio: () =>
      Promise.resolve({ zeilen: [], ohnePreis: 0, fehler: null }),
    Preise: () => Promise.resolve({ staende: [], befunde: [], fehler: null }),
    Einstellungen: () => Promise.resolve({ einstellungen: [], fehler: null }),
    MarkenJeObjekt: je,
    MarkenFiltern: filtern,
    ObjektmarkeSetzen: setzen,
    EinstellungSetzen: einstellungSetzen,
  })

  render(
    <Sprachraum sprache="de">
      <Sammlung bestandGeaendert={vi.fn()} />
    </Sprachraum>,
  )

  await screen.findByText('Bernsteinruf')

  await userEvent.click(
    await screen.findByRole('button', { name: 'Nach »tausch« filtern' }),
  )

  await waitFor(() => {
    expect(screen.queryByText('Bernsteinruf')).toBeNull()
  })
  expect(screen.getByText('Aschenwächter')).toBeInTheDocument()
  expect(filtern).toHaveBeenCalledWith('exemplar', ['tausch'], 'und')

  // **Im Inspector: die Marken des gewählten Exemplars und „+ Marke".**
  await userEvent.click(screen.getByText('Aschenwächter'))
  const inspector = within(
    screen.getByRole('complementary', {
      name: kataloge.de.seiten.sammlung.inspector,
    }),
  )
  expect(inspector.getByText('tausch')).toBeInTheDocument()

  await userEvent.type(
    inspector.getByRole('textbox', { name: texte.setzenFuer }),
    'Binder 2',
  )
  await userEvent.click(inspector.getByRole('button', { name: texte.setzen }))

  await waitFor(() => {
    expect(je).toHaveBeenCalledTimes(2)
  })
  await antwortAbwarten()
  expect(setzen).toHaveBeenCalledWith('exemplar', '31', 'Binder 2')

  // **Die Wahl lässt sich unter einem Namen speichern** — als Einstellung.
  await userEvent.type(
    screen.getByRole('textbox', { name: texte.speichernName }),
    'Zum Tausch',
  )
  await userEvent.click(screen.getByRole('button', { name: texte.speichern }))
  await waitFor(() => {
    expect(einstellungSetzen).toHaveBeenCalledWith(
      'filter.gespeichert',
      expect.stringContaining('"name":"Zum Tausch"'),
    )
  })
  await antwortAbwarten()

  // **„Alles zeigen" nimmt die Wahl zurück.**
  await userEvent.click(
    screen.getByRole('button', { name: texte.filterLeeren }),
  )
  expect(await screen.findByText('Bernsteinruf')).toBeInTheDocument()
})
