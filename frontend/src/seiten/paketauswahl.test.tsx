// **Die Pakete mit Befund, gefiltert** — AP-8
// (#2365),
// Befund B-14: die Sammelzeile der Aufmerksamkeit öffnet die Paketliste mit
// genau ihren Paketen. Der Filter nennt sich und lässt sich lösen.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import type { PaketDTO } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { PaketlisteMitBefunden } from '../test/abhaengigkeiten'
import { belegeApi, raeumeApi } from '../test/api'
import { Anschlussbuehne, uebergabeknopf } from '../test/anschlussbuehne'
import { anlegbar, katalog, texte } from '../test/paketliste'

afterEach(() => {
  raeumeApi()
})

// paket gibt das Paket der Vorrichtung unter einem anderen Namen.
function paket(name: string): PaketDTO {
  const [vorlage] = katalog().pakete
  if (vorlage === undefined) {
    throw new Error('die Vorrichtung trägt kein Paket')
  }

  return { ...vorlage, name, anzeigename: name }
}

it('zeigt genau die übergebenen Pakete und lässt den Filter lösen', async () => {
  belegeApi({
    PaketArten: anlegbar,
    Pakete: vi.fn(() =>
      Promise.resolve({
        pakete: [paket('probe-eins'), paket('probe-zwei'), paket('probe-drei')],
        befunde: [],
      }),
    ),
  })

  render(
    <Sprachraum sprache="de">
      <Anschlussbuehne
        anschluss={{
          art: 'paketauswahl',
          name: '2 eigene Pakete mit Befund im Skript',
          namen: ['probe-eins', 'probe-drei'],
        }}
      >
        <PaketlisteMitBefunden />
      </Anschlussbuehne>
    </Sprachraum>,
  )

  await screen.findByText('probe-zwei')
  await userEvent.click(screen.getByRole('button', { name: uebergabeknopf }))

  expect(
    await screen.findByText('2 eigene Pakete mit Befund im Skript'),
  ).toBeInTheDocument()
  expect(screen.queryByText('probe-zwei')).toBeNull()
  expect(screen.getByText('probe-eins')).toBeInTheDocument()
  expect(screen.getByText('probe-drei')).toBeInTheDocument()

  await userEvent.click(
    screen.getByRole('button', { name: texte.seiten.pakete.auswahlLoesen }),
  )

  expect(await screen.findByText('probe-zwei')).toBeInTheDocument()
  expect(screen.queryByText('2 eigene Pakete mit Befund im Skript')).toBeNull()
})
