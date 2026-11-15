// **Ein Lesefehler nach einem geglückten Schreiben nimmt nichts zurück** —
// Z9 des Zweitblicks über Phase 6.12
// (#1905).
//
// Nach dem Schreiben wird neu gelesen; scheiterte das, standen danach gar
// keine Einstellungen mehr da. Ein eben abgeschalteter Bereich erschien
// wieder, obwohl er im Bestand aus war und seine Läufe beendet waren.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { useEinstellungen } from './useEinstellungen'

afterEach(() => {
  raeumeApi()
})

function Schalter() {
  const stand = useEinstellungen()

  return (
    <>
      <p>{`spielen:${stand.wert('schalter.spielen')}`}</p>
      <p>{`sprache:${stand.wert('sprache')}`}</p>
      <p>{stand.ursache === undefined ? 'ohne Fehler' : 'mit Fehler'}</p>
      <button
        type="button"
        onClick={() => {
          stand.setze('schalter.spielen', 'aus')
        }}
      >
        aus
      </button>
    </>
  )
}

// **Die Werte wie im Bestand**: ein Schalter kennt `an` und `aus`, eine
// Sprache ihre Kürzel. Nachgetragen wird nur, was in der Auswahl steht.
function eintrag(schluessel: string, wert: string, bereich: string) {
  const werte = bereich === 'schalter' ? ['an', 'aus'] : ['erben', 'de', 'en']

  return { schluessel, wert, wirksam: wert, werte, bereich }
}

it('behält den geschriebenen Wert und den übrigen Stand, und sagt den Fehler', async () => {
  let gelesen = 0

  belegeApi({
    Einstellungen: () => {
      gelesen += 1

      return gelesen === 1
        ? Promise.resolve({
            einstellungen: [
              eintrag('schalter.spielen', 'an', 'schalter'),
              eintrag('sprache', 'en', 'anwendung'),
            ],
          })
        : Promise.reject(new Error('der Bestand ist zu'))
    },
    EinstellungSetzen: () => Promise.resolve({}),
  })

  render(<Schalter />)

  await screen.findByText('spielen:an')
  await userEvent.click(screen.getByRole('button', { name: 'aus' }))

  await waitFor(() => {
    expect(screen.getByText('mit Fehler')).toBeInTheDocument()
  })
  expect(screen.getByText('spielen:aus')).toBeInTheDocument()
  expect(screen.getByText('sprache:en')).toBeInTheDocument()
})
