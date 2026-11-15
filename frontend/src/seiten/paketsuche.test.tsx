// **Ein Sprung zu den Paketen mit der Suche schon gefüllt** — DOM-01.AK-04
// (#1696):
// der Bedarf eines Sets nennt ein Paket, und die Paketseite sucht es.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { PaketlisteMitBefunden } from '../test/abhaengigkeiten'
import { belegeApi, raeumeApi } from '../test/api'
import { Anschlussbuehne, uebergabeknopf } from '../test/anschlussbuehne'
import { anlegbar, katalog, texte } from '../test/paketliste'

afterEach(() => {
  raeumeApi()
})

it('sucht das Paket, das ein Set braucht', async () => {
  belegeApi({
    PaketArten: anlegbar,
    Pakete: vi.fn(() => Promise.resolve(katalog())),
  })

  render(
    <Sprachraum sprache="de">
      <Anschlussbuehne
        anschluss={{
          art: 'paketsuche',
          name: 'probe-schlicht.mse-style',
          paket: 'probe-schlicht.mse-style',
        }}
      >
        <PaketlisteMitBefunden />
      </Anschlussbuehne>
    </Sprachraum>,
  )

  await screen.findByText('probe-schlicht')
  await userEvent.click(screen.getByRole('button', { name: uebergabeknopf }))

  // **Ohne Endung** — gesucht wird im Namen, und der trägt sie nicht.
  expect(
    await screen.findByDisplayValue('probe-schlicht'),
  ).toHaveAccessibleName(texte.seiten.pakete.katalogSuche)
})
