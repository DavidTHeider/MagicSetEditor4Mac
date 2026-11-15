// **Die Schalter eines Arbeitsplatzes an einem fremden Ort** — P03
// (#2174).
//
// Die Kartenseite stellt „Karte anlegen", „Mehr" und die zwei Schalter in
// **eine** Zeile, als zwei Gruppen. Die Schalter gehören weiter dem
// Arbeitsplatz — sein Zustand, seine Tasten, die Rückgabe des Fokus —; sie
// stehen nur woanders im Dokument. `schalterort` sagt, wo:
//
// - `undefined`: kein fremder Ort, die eigene Zeile wie bisher,
// - `null`: ein Ort ist angekündigt, aber noch nicht da — dann steht nichts,
//   statt für einen Augenblick an der alten Stelle,
// - ein Element: dort.
import { fireEvent, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'

import { Arbeitsplatz } from './Arbeitsplatz'
import { gibJsdomEineBreite } from '../test/sichtbreite'
import { Sprachraum } from '../sprache/kontext'

let raeumeBreite: () => void = () => undefined

beforeEach(() => {
  raeumeBreite = gibJsdomEineBreite(1400)
})

afterEach(() => {
  raeumeBreite()
})

function Probe({ ort }: { ort: HTMLElement | null | undefined }) {
  return (
    <Sprachraum sprache="de">
      <Arbeitsplatz
        bezeichnung="Kartenarbeitsplatz"
        navigator={{ bezeichnung: 'Karten', inhalt: <p>Die Liste</p> }}
        zentrum={<p>Die Vorschau</p>}
        inspector={{ bezeichnung: 'Felder', inhalt: <p>Das Formular</p> }}
        schalterort={ort}
      />
    </Sprachraum>
  )
}

describe('Die Schalter an einem fremden Ort', () => {
  it('stehen dort und nicht im Arbeitsplatz', () => {
    const ort = document.createElement('div')
    ort.setAttribute('data-ort', 'zeile')
    document.body.append(ort)

    render(<Probe ort={ort} />)

    const arbeitsplatz = screen.getByRole('region', {
      name: 'Kartenarbeitsplatz',
    })
    expect(
      within(arbeitsplatz).queryByRole('button', { name: /Navigator/ }),
    ).not.toBeInTheDocument()
    expect(
      within(ort).getByRole('button', { name: /Navigator ausblenden/ }),
    ).toBeInTheDocument()
    expect(
      within(ort).getByRole('button', { name: /Inspector ausblenden/ }),
    ).toBeInTheDocument()

    ort.remove()
  })

  it('klappen ihre Region weiter, mit Maus und Taste', async () => {
    const ort = document.createElement('div')
    document.body.append(ort)

    render(<Probe ort={ort} />)

    await userEvent.click(
      within(ort).getByRole('button', { name: /Navigator ausblenden/ }),
    )
    expect(screen.queryByRole('complementary', { name: 'Karten' })).toBeNull()
    expect(
      within(ort).getByRole('button', { name: /Navigator einblenden/ }),
    ).toHaveAttribute('aria-expanded', 'false')

    fireEvent.keyDown(window, {
      key: '[',
      code: 'BracketLeft',
      altKey: true,
      metaKey: true,
    })
    expect(
      await screen.findByRole('complementary', { name: 'Karten' }),
    ).toBeInTheDocument()

    ort.remove()
  })

  it('stehen nirgends, solange der angekündigte Ort fehlt', () => {
    render(<Probe ort={null} />)

    expect(screen.queryByRole('button', { name: /Navigator/ })).toBeNull()
    expect(screen.queryByRole('button', { name: /Inspector/ })).toBeNull()
    expect(screen.getByText('Das Formular')).toBeInTheDocument()
  })

  it('stehen ohne fremden Ort in der eigenen Zeile, wie bisher', () => {
    render(<Probe ort={undefined} />)

    const arbeitsplatz = screen.getByRole('region', {
      name: 'Kartenarbeitsplatz',
    })
    expect(
      within(arbeitsplatz).getByRole('button', {
        name: /Navigator ausblenden/,
      }),
    ).toBeInTheDocument()
  })
})
