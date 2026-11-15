// **Die Sets mit Problemen, gefiltert** — AP-8, Teil 2
// (#2365),
// Befund B-13: die Kachel des Überblicks öffnet die Setliste mit genau ihren
// Sets. Der Filter nennt sich und lässt sich lösen.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import type { SetDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Anschlussbuehne, uebergabeknopf } from '../test/anschlussbuehne'
import { stillerSpeicherlauf } from '../test/speicherlauf'
import { Sets } from './Sets'

afterEach(() => {
  raeumeApi()
})

const set = (id: string, name: string): SetDTO => ({
  id,
  name,
  spiel: 'magic',
  stil: 'magic-m15',
  karten: 3,
  herkunft: '',
  herkunftFehlt: false,
})

it('zeigt genau die übergebenen Sets und lässt den Filter lösen', async () => {
  belegeApi({
    Sets: () =>
      Promise.resolve({
        sets: [set('1', 'Vier Winde'), set('2', 'Glutpfad'), set('3', 'Moor')],
      }),
  })
  render(
    <Sprachraum sprache="de">
      <Anschlussbuehne
        anschluss={{
          art: 'setauswahl',
          name: '2 Sets mit Problemen',
          setIDs: ['1', '3'],
        }}
      >
        <Sets
          wechseln={vi.fn()}
          waehlen={vi.fn()}
          speichern={stillerSpeicherlauf()}
        />
      </Anschlussbuehne>
    </Sprachraum>,
  )

  await screen.findByText('Glutpfad')
  await userEvent.click(screen.getByRole('button', { name: uebergabeknopf }))

  expect(await screen.findByText('2 Sets mit Problemen')).toBeInTheDocument()
  expect(screen.queryByText('Glutpfad')).toBeNull()
  expect(screen.getByText('Vier Winde')).toBeInTheDocument()
  expect(screen.getByText('Moor')).toBeInTheDocument()

  await userEvent.click(
    screen.getByRole('button', {
      name: kataloge.de.seiten.sets.auswahlLoesen,
    }),
  )

  expect(await screen.findByText('Glutpfad')).toBeInTheDocument()
  expect(screen.queryByText('2 Sets mit Problemen')).toBeNull()
})
