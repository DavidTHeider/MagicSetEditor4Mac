// **Sensible Teile lassen sich vor dem Sichern herausnehmen** — OPS-04.AK-04
// (#1721).
// Je Teil ein Häkchen „mitnehmen"; was abgewählt ist, geht als `ohne` an Go.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Diagnosepaket } from './Diagnosepaket'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

const texte = kataloge.de.seiten.technik.paket

afterEach(raeumeApi)

it('nimmt einen abgewählten Teil aus dem Paket', async () => {
  const sichern = vi.fn(() => Promise.resolve({ fehler: null }))
  belegeApi({
    Diagnosepaket: vi.fn(() =>
      Promise.resolve({
        teile: [
          {
            teil: 'fassung',
            datei: 'fassung.txt',
            zeilen: 3,
            bytes: 64,
            grund: '',
          },
          {
            teil: 'protokoll',
            datei: 'protokoll.txt',
            zeilen: 9,
            bytes: 900,
            grund: '',
          },
        ],
        bytes: 964,
        grenze: 8388608,
        von: '',
        bis: '',
        ohne: [],
        fehler: null,
      }),
    ),
    DiagnosepaketSichern: sichern,
  })

  render(
    <Sprachraum sprache="de">
      <Diagnosepaket />
    </Sprachraum>,
  )

  await userEvent.click(
    screen.getByRole('button', { name: texte.zusammenstellen }),
  )

  const protokoll = await screen.findByRole('checkbox', {
    name: texte.mitnehmen.replace('{teil}', texte.teile.protokoll),
  })
  expect(protokoll).toBeChecked()

  await userEvent.click(protokoll)
  await userEvent.click(screen.getByRole('button', { name: texte.sichern }))

  await waitFor(() => {
    expect(sichern).toHaveBeenCalledWith(['protokoll'])
  })
})

// **Ein zweiter Klick nimmt ihn wieder mit** — die Wahl ist ein Schalter und
// keine Einbahnstraße.
it('nimmt einen wieder angewählten Teil mit', async () => {
  const sichern = vi.fn(() => Promise.resolve({ fehler: null }))
  belegeApi({
    Diagnosepaket: vi.fn(() =>
      Promise.resolve({
        teile: [
          {
            teil: 'protokoll',
            datei: 'protokoll.txt',
            zeilen: 9,
            bytes: 900,
            grund: '',
          },
        ],
        bytes: 900,
        grenze: 8388608,
        von: '',
        bis: '',
        ohne: [],
        fehler: null,
      }),
    ),
    DiagnosepaketSichern: sichern,
  })

  render(
    <Sprachraum sprache="de">
      <Diagnosepaket />
    </Sprachraum>,
  )

  await userEvent.click(
    screen.getByRole('button', { name: texte.zusammenstellen }),
  )

  const protokoll = await screen.findByRole('checkbox', {
    name: texte.mitnehmen.replace('{teil}', texte.teile.protokoll),
  })

  await userEvent.click(protokoll)
  await waitFor(() => {
    expect(protokoll).not.toBeChecked()
  })

  await userEvent.click(protokoll)
  await waitFor(() => {
    expect(protokoll).toBeChecked()
  })

  await userEvent.click(screen.getByRole('button', { name: texte.sichern }))

  await waitFor(() => {
    expect(sichern).toHaveBeenCalledWith([])
  })
})
