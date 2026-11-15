// **Die Suche sagt, dass sie lädt, und was scheiterte** — G11
// (#2277).
//
// B03 verzögerte `Suchfelder` um vier Sekunden und ließ sie scheitern: die
// Seite stand beide Male wie im Normalfall da — die Auswahl der Felder nur
// leer.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, belegeFenster, raeumeApi } from '../test/api'
import { Suche } from './Suche'

const zustaende = kataloge.de.bausteine.zustaende

const fehlschlag = {
  code: 'unbekannt' as const,
  schluessel: 'unbekannt',
  meldung: 'roh',
  werte: {},
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Suche oeffnen={() => undefined} />
    </Sprachraum>,
  )
}

afterEach(raeumeApi)

describe('Die Suche beim Öffnen', () => {
  it('sagt, dass sie lädt, solange die Felder ausstehen', async () => {
    belegeApi({
      ...belegeFenster('de'),
      Sets: () => Promise.resolve({ sets: [] }),
      Suchfelder: () => new Promise(() => undefined),
    })

    zeige()

    expect(await screen.findByRole('status')).toHaveTextContent(zustaende.laedt)
  })

  it('meldet, wenn die Felder scheitern, und lädt sie erneut', async () => {
    let gefragt = 0
    belegeApi({
      ...belegeFenster('de'),
      Sets: () => Promise.resolve({ sets: [] }),
      Suchfelder: () => {
        gefragt++

        return Promise.resolve(
          gefragt === 1
            ? { felder: [], fehler: fehlschlag }
            : { felder: ['name'] },
        )
      },
    })

    zeige()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.unbekannt,
    )

    await userEvent.click(
      screen.getByRole('button', { name: zustaende.wiederholen }),
    )

    expect(await screen.findByRole('option', { name: 'name' })).toBeVisible()
    expect(screen.queryByRole('alert')).toBeNull()
  })
})
