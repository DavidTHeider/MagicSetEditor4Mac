// **Marken an einem Paket** — UX-12 (#1684): eine lokale Anmerkung des
// Profils, erreicht über das Menü der Zeile.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { PaketlisteMitBefunden } from '../test/abhaengigkeiten'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

const texte = kataloge.de.bausteine.objektmarken

afterEach(() => {
  raeumeApi()
})

const paket = {
  art: 'mse-style',
  name: 'probe-schlicht',
  anzeigename: 'probe-schlicht',
  version: '2026-08-15',
  pfad: '/pakete/probe-schlicht.mse-style',
  herkunft: 'aus der Saat',
  pruefsumme: 'sha256:abc',
  braucht: [],
  gruppe: '',
  rang: 1000000,
  artrang: 30,
  herkunftsart: 'mitgeliefert' as const,
  hauptdatei: 'style',
  geaendert: false,
  gebrauchtVon: [],
  benutzt: 0,
}

it('zeigt die Marken eines Pakets und öffnet den Editor über das Menü', async () => {
  const setzen = vi.fn().mockResolvedValue({ fehler: null })
  belegeApi({
    PaketArten: () =>
      Promise.resolve([{ art: 'mse-style', brauchtSpiel: true }]),
    Pakete: vi.fn(() => Promise.resolve({ pakete: [paket], befunde: [] })),
    MarkenJeObjekt: vi.fn().mockResolvedValue({
      marken: { 'mse-style/probe-schlicht': ['liebling'] },
      fehler: null,
    }),
    ObjektmarkeSetzen: setzen,
  })

  render(
    <Sprachraum sprache="de">
      <PaketlisteMitBefunden />
    </Sprachraum>,
  )

  expect(await screen.findByText('liebling')).toBeInTheDocument()

  // **Erst wählen** — seit UXR-17 (#1925) steht das Menü nur an der
  // aktiven Zeile.
  await userEvent.click(screen.getByRole('row', { name: 'probe-schlicht' }))
  await userEvent.click(
    screen.getByRole('button', { name: 'Handlungen — probe-schlicht' }),
  )
  await userEvent.click(
    screen.getByRole('menuitem', { name: texte.ueberschrift }),
  )

  await userEvent.type(
    screen.getByRole('textbox', { name: texte.setzenFuer }),
    'Alt',
  )
  await userEvent.click(screen.getByRole('button', { name: texte.setzen }))

  await waitFor(() => {
    expect(setzen).toHaveBeenCalledWith(
      'paket',
      'mse-style/probe-schlicht',
      'Alt',
    )
  })
  await antwortAbwarten()

  // **Ein zweiter Klick auf den Eintrag schließt den Editor wieder.**
  await userEvent.click(
    screen.getByRole('button', { name: 'Handlungen — probe-schlicht' }),
  )
  await userEvent.click(
    screen.getByRole('menuitem', { name: texte.ueberschrift }),
  )
  expect(screen.queryByRole('textbox', { name: texte.setzenFuer })).toBeNull()
})
