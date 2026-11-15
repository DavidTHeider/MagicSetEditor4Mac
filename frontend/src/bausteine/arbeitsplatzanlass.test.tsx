// **Ein Anlass öffnet den Inspector bei Enge** — UXR-17
// (#1925).
//
// Gemessen an der Hüllenprobe `fmt02e`: bei 1280 × 820 Punkten weicht der
// Inspector der Pakete (`beiEnge="inspector"`), und **eine gewählte Zeile
// blendet ihn nicht ein** — wer ein Paket anklickte oder im Menü der Zeile
// *Löschen …* wählte, sah nichts geschehen; der Plan stand in einer Spalte,
// die niemand sieht. Jetzt gibt die Seite einen Anlass, und bei Enge geht
// der Inspector darauf über dem Zentrum auf. **Bei Platz gilt der Wunsch**:
// wer ihn geschlossen hat, dem öffnet ihn keine Wahl.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, describe, expect, it } from 'vitest'

import { Arbeitsplatz } from './Arbeitsplatz'
import { effekteNachholen } from '../test/effekte'
import { gibJsdomEineBreite } from '../test/sichtbreite'
import { Sprachraum } from '../sprache/kontext'

let raeumeBreite: () => void = () => undefined

afterEach(() => {
  raeumeBreite()
})

function Probe() {
  const [anlass, setzeAnlass] = useState(0)

  return (
    <Sprachraum sprache="de">
      <button
        type="button"
        onClick={() => {
          setzeAnlass((zahl) => zahl + 1)
        }}
      >
        Wählen
      </button>
      <Arbeitsplatz
        bezeichnung="Paketarbeitsplatz"
        navigator={{ bezeichnung: 'Suche', inhalt: <p>Die Filter</p> }}
        zentrum={<p>Die Liste</p>}
        inspector={{ bezeichnung: 'Paket', inhalt: <p>Das Paket</p> }}
        beiEnge="inspector"
        inspectorAnlass={anlass}
      />
    </Sprachraum>
  )
}

const inspector = () => screen.queryByRole('complementary', { name: 'Paket' })

describe('Der Anlass für den Inspector', () => {
  it('lässt ihn bei Enge über dem Zentrum aufgehen', async () => {
    raeumeBreite = gibJsdomEineBreite(1000)
    render(<Probe />)
    await effekteNachholen()

    // **Der erste Stand ist kein Anlass** — wer die Seite aufschlägt, hat
    // noch nichts gewählt.
    expect(inspector()).toBeNull()

    await userEvent.click(screen.getByRole('button', { name: 'Wählen' }))
    expect(inspector()).toHaveAttribute('data-ueber', 'zentrum')

    // **Geschlossen, und der nächste Anlass öffnet wieder.**
    await userEvent.click(
      screen.getByRole('button', { name: 'Inspector ausblenden' }),
    )
    expect(inspector()).toBeNull()

    await userEvent.click(screen.getByRole('button', { name: 'Wählen' }))
    expect(inspector()).toHaveAttribute('data-ueber', 'zentrum')
  })

  it('lässt bei Platz den Wunsch stehen', async () => {
    raeumeBreite = gibJsdomEineBreite(1400)
    render(<Probe />)
    await effekteNachholen()

    await userEvent.click(
      screen.getByRole('button', { name: 'Inspector ausblenden' }),
    )
    await userEvent.click(screen.getByRole('button', { name: 'Wählen' }))

    expect(inspector()).toBeNull()
  })
})
