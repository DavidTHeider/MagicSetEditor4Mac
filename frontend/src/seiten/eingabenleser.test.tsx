// **Was ohne Knopf schreibt, steht für einen Leser still** — #2290
// (#2290).
//
// Neben den Kartenfeldern (`feldeingabeleser.test.tsx`) schreiben vier
// Bausteine beim Verlassen oder beim Wählen: Kartennotizen, Schlüsselwörter,
// Stiloptionen und die Regeln der Autoersetzung. Ein Textfeld wird
// schreibgeschützt, was sofort schreibt, gesperrt. Den Grund nennt der
// Streifen oben — **und jede Eingabe selbst** als Beschreibung (H01,
// #2255).
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Schreibrechtraum } from '../haken/schreibrecht'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Autoersetzungen } from './Autoersetzungen'
import { Kartennotizen } from './Kartennotizen'
import { Schluesselwortfeld } from './Schluesselwortfeld'
import { Stiloption } from './Stiloption'

afterEach(raeumeApi)

const satz = kataloge.de.bausteine.nurlesen.knopf

function alsLeser(knoten: React.ReactNode) {
  render(
    <Sprachraum sprache="de">
      <Schreibrechtraum nurLesend>{knoten}</Schreibrechtraum>
    </Sprachraum>,
  )
}

describe('Selbstschreibende Eingaben für einen Leser', () => {
  it('Kartennotizen: der Kasten ist schreibgeschützt', () => {
    alsLeser(
      <Kartennotizen
        setID="1"
        karteID="7"
        werte={[]}
        geaendert={vi.fn()}
        aufFehler={vi.fn()}
      />,
    )

    expect(screen.getByRole('textbox')).toHaveAttribute('readonly')
    expect(screen.getByRole('textbox')).toHaveAccessibleDescription(satz)
  })

  it('Schlüsselwort: das Feld ist schreibgeschützt', () => {
    alsLeser(
      <Schluesselwortfeld
        aufschrift="Muster"
        wert="Fliegend"
        angehalten={false}
        entwurf={{
          schluessel: 'a',
          konflikt: false,
          aendern: vi.fn(),
          verlassen: vi.fn(),
          behalten: vi.fn(),
          verwerfen: vi.fn(),
          uebernehmen: vi.fn(),
          bekannt: [],
        }}
      />,
    )

    const feld = screen.getByRole('textbox', { name: 'Muster' })
    expect(feld).toHaveAttribute('readonly')
    expect(feld).toHaveAccessibleDescription(satz)
  })

  it('Stiloption: die Auswahl ist gesperrt', () => {
    belegeApi({
      Stilwahlbilder: () =>
        Promise.resolve({ bilder: [], befunde: [], fehler: null }),
    })
    alsLeser(
      <Stiloption
        setID="1"
        laeuft={false}
        setzen={vi.fn()}
        option={{
          name: 'rahmen',
          art: 'choice',
          beschreibung: '',
          wert: 'rot',
          moeglichkeiten: [
            { wert: 'rot', beschriftung: 'rot' },
            { wert: 'blau', beschriftung: 'blau' },
          ],
          eigeneFarbe: false,
          gesetzt: true,
          hatVorgabe: false,
          vorgabe: '',
          hatWahlbilder: false,
        }}
      />,
    )

    expect(screen.getByRole('combobox')).toBeDisabled()
    expect(screen.getByRole('combobox')).toHaveAccessibleDescription(satz)
  })

  it('Autoersetzung: Schalter und Regeln sind gesperrt', async () => {
    belegeApi({
      Autoersetzungen: () =>
        Promise.resolve({
          regeln: [
            {
              muster: '(C)',
              ersatz: '©',
              aktiv: true,
              ganzesWort: false,
              eigen: true,
            },
          ],
          an: true,
          fehler: null,
        }),
    })
    alsLeser(<Autoersetzungen setID="1" />)

    const schalter = await screen.findByRole('checkbox', {
      name: kataloge.de.seiten.karten.autoersetzung.schalter,
    })
    expect(schalter).toBeDisabled()
    expect(schalter).toHaveAccessibleDescription(satz)
  })
})
