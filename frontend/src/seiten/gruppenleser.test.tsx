// **Jedes Kästchen und jedes Optionsfeld nennt den Grund** — H01
// (#2255),
// Befund 4 des Nachlaufs über Phase I.
//
// Mehrfachwahl und Bildwahl einer Stiloption sind Gruppen; der Grund stand
// nur an ihrem `fieldset`. Ob ein Vorleser die Beschreibung der Gruppe beim
// einzelnen Feld spricht, ist nicht gemessen — die des Feldes spricht er.
import { render, screen } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import type { StiloptionDTO } from '../api'
import { Schreibrechtraum } from '../haken/schreibrecht'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Stiloption } from './Stiloption'

afterEach(raeumeApi)

const satz = kataloge.de.bausteine.nurlesen.knopf
const punkt =
  'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=='

function option(teil: Partial<StiloptionDTO>): StiloptionDTO {
  return {
    name: 'probe',
    art: 'choice',
    beschreibung: '',
    wert: 'a',
    moeglichkeiten: [
      { wert: 'a', beschriftung: 'a' },
      { wert: 'b', beschriftung: 'b' },
    ],
    eigeneFarbe: false,
    gesetzt: true,
    hatVorgabe: false,
    vorgabe: '',
    hatWahlbilder: false,
    ...teil,
  }
}

function alsLeser(eine: StiloptionDTO) {
  render(
    <Sprachraum sprache="de">
      <Schreibrechtraum nurLesend>
        <Stiloption setID="1" laeuft={false} setzen={vi.fn()} option={eine} />
      </Schreibrechtraum>
    </Sprachraum>,
  )
}

it('Mehrfachwahl: jedes Kästchen nennt den Grund', () => {
  alsLeser(option({ art: 'multiple choice', wert: 'a' }))

  const kaestchen = screen.getAllByRole('checkbox')
  expect(kaestchen).toHaveLength(2)
  for (const eines of kaestchen) {
    expect(eines).toBeDisabled()
    expect(eines).toHaveAccessibleDescription(satz)
  }
})

it('Bildwahl: jedes Optionsfeld nennt den Grund', async () => {
  belegeApi({
    Stilwahlbilder: () =>
      Promise.resolve({
        bilder: [
          { wahl: 'a', bild: punkt },
          { wahl: 'b', bild: punkt },
        ],
        befunde: [],
        fehler: null,
      }),
  })
  alsLeser(option({ hatWahlbilder: true }))

  const felder = await screen.findAllByRole('radio')
  expect(felder).toHaveLength(2)
  for (const eines of felder) {
    expect(eines).toBeDisabled()
    expect(eines).toHaveAccessibleDescription(satz)
  }
})
