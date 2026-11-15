// **Zwei geglückte Schreibvorgänge, und beide bleiben** — N6 des Nachlaufs
// über Phase 6.12
// (#1905).
//
// Nach Z9 galt ein geschriebener Wert erst mit seinem Nachladen. Zwei
// Schreibvorgänge auf verschiedene Schlüssel nacheinander: das erste
// Nachladen war überholt und wurde verworfen, das zweite brachte nur seinen
// eigenen Wert — der erste ging verloren, obwohl er im Bestand stand.
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
      <p>{`netz:${stand.wert('schalter.netz')}`}</p>
      <button
        type="button"
        onClick={() => {
          stand.setze('schalter.spielen', 'aus')
        }}
      >
        spielen aus
      </button>
      <button
        type="button"
        onClick={() => {
          stand.setze('schalter.netz', 'aus')
        }}
      >
        netz aus
      </button>
    </>
  )
}

function eintrag(schluessel: string, wert: string) {
  return {
    schluessel,
    wert,
    wirksam: wert,
    werte: ['an', 'aus'],
    bereich: 'schalter',
  }
}

it('behält beide geschriebenen Werte, auch wenn beide Nachladen scheitern', async () => {
  let gelesen = 0
  const offen: ((grund: Error) => void)[] = []

  belegeApi({
    Einstellungen: () => {
      gelesen += 1

      return gelesen === 1
        ? Promise.resolve({
            einstellungen: [
              eintrag('schalter.spielen', 'an'),
              eintrag('schalter.netz', 'an'),
            ],
          })
        : new Promise((_, ablehnen) => {
            offen.push(ablehnen)
          })
    },
    EinstellungSetzen: () => Promise.resolve({}),
  })

  render(<Schalter />)

  await screen.findByText('spielen:an')
  await userEvent.click(screen.getByRole('button', { name: 'spielen aus' }))
  await userEvent.click(screen.getByRole('button', { name: 'netz aus' }))

  await waitFor(() => {
    expect(offen).toHaveLength(2)
  })

  for (const ablehnen of offen) {
    ablehnen(new Error('der Bestand ist zu'))
  }

  await waitFor(() => {
    expect(screen.getByText('netz:aus')).toBeInTheDocument()
  })
  expect(screen.getByText('spielen:aus')).toBeInTheDocument()
})
