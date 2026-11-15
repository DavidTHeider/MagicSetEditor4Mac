// Die Stiloptionen eines Sets (#41).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { type Api } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Stiloptionen } from './Stiloptionen'

afterEach(() => {
  raeumeApi()
})

// optionen gibt eine Auswahl, einen Schalter und ein Textfeld.
function optionen() {
  return {
    optionen: [
      {
        name: 'hintergrund',
        art: 'choice',
        beschreibung: 'Der Grund hinter der Karte.',
        wert: 'hell',
        moeglichkeiten: [
          { wert: 'hell', beschriftung: 'hell' },
          { wert: 'dunkel', beschriftung: 'dunkel' },
        ],
        eigeneFarbe: false,
        gesetzt: false,
        hatVorgabe: false,
        hatWahlbilder: false,
        vorgabe: '',
      },
      {
        name: 'rahmen zeigen',
        art: 'boolean',
        beschreibung: '',
        wert: 'no',
        moeglichkeiten: [],
        eigeneFarbe: false,
        gesetzt: false,
        hatVorgabe: false,
        hatWahlbilder: false,
        vorgabe: '',
      },
      {
        name: 'aufschrift',
        art: 'text',
        beschreibung: '',
        wert: 'alt',
        moeglichkeiten: [],
        eigeneFarbe: false,
        gesetzt: false,
        hatVorgabe: false,
        hatWahlbilder: false,
        vorgabe: '',
      },
    ],
  }
}

function belegeOptionen(teile: Partial<Api>) {
  return belegeApi({
    Stiloptionen: () => Promise.resolve(optionen()),
    ...teile,
  })
}

function zeige(geaendert = vi.fn()) {
  const stand = render(
    <Sprachraum sprache="de">
      <Stiloptionen setID="1" geaendert={geaendert} />
    </Sprachraum>,
  )

  return { geaendert, stand }
}

describe('Stiloptionen', () => {
  // **Die Optionen kommen aus dem Paket** — samt ihrer Beschreibung, und die
  // bleibt unübersetzt (§9).
  it('zeigt Auswahl, Schalter und Text mit der Beschreibung des Pakets', async () => {
    belegeOptionen({})
    zeige()

    expect(
      await screen.findByText('Der Grund hinter der Karte.'),
    ).toBeInTheDocument()
    expect(screen.getByLabelText(/hintergrund/)).toHaveValue('hell')
    expect(screen.getByLabelText(/rahmen zeigen/)).not.toBeChecked()
    expect(screen.getByLabelText(/aufschrift/)).toHaveValue('alt')
  })

  // **Vor die Verneinung gehört ein Beleg, dass die Antwort angekommen ist**
  // (§3, #661).
  //
  // Der erste Anlauf lud die leere Liste und sah sofort nach: „kein
  // Auswahlfeld da". Das trifft aber auch zu, **solange geladen wird** — und
  // solange lädt, gibt die Seite ohnehin nichts her. `waitFor` gab beim ersten
  // Versuch zurück, und die leere Antwort wurde nie geprüft.
  //
  // **Nachgemessen**: gab die Seite für eine geladene, leere Liste ein
  // Auswahlfeld her, blieb der Test grün.
  //
  // Deshalb erst ein Set **mit** Optionen. `useLaden` lässt das Vorige stehen,
  // während der neue Schlüssel lädt — die Auswahl verschwindet also erst, wenn
  // die leere Antwort eingetragen ist. Genau das prüft die Zusicherung jetzt.
  it('zeigt nichts, wenn der Stil keine Optionen führt', async () => {
    belegeOptionen({
      Stiloptionen: vi
        .fn()
        .mockResolvedValueOnce(optionen())
        .mockResolvedValue({ optionen: [] }),
    })

    const { geaendert, stand } = zeige()

    expect(await screen.findByLabelText(/hintergrund/)).toBeInTheDocument()

    stand.rerender(
      <Sprachraum sprache="de">
        <Stiloptionen setID="2" geaendert={geaendert} />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(screen.queryByRole('combobox')).not.toBeInTheDocument()
    })
  })

  it('schreibt eine Auswahl und meldet die Änderung', async () => {
    const StiloptionSetzen = vi.fn(() => Promise.resolve({}))
    belegeOptionen({ StiloptionSetzen })
    const { geaendert } = zeige()

    await screen.findByLabelText(/hintergrund/)
    await userEvent.selectOptions(
      screen.getByLabelText(/hintergrund/),
      'dunkel',
    )

    expect(StiloptionSetzen).toHaveBeenCalledWith('1', 'hintergrund', 'dunkel')
    await waitFor(() => {
      expect(geaendert).toHaveBeenCalled()
    })
  })

  // **`yes` und `no`, wie die Pakete es schreiben** — nicht `true`/`false`.
  it('schreibt einen Schalter als yes', async () => {
    const StiloptionSetzen = vi.fn(() => Promise.resolve({}))
    belegeOptionen({ StiloptionSetzen })
    zeige()

    await userEvent.click(await screen.findByLabelText(/rahmen zeigen/))

    expect(StiloptionSetzen).toHaveBeenCalledWith('1', 'rahmen zeigen', 'yes')
  })

  it('schreibt einen abgeschalteten Schalter als no', async () => {
    const StiloptionSetzen = vi.fn(() => Promise.resolve({}))
    belegeOptionen({
      Stiloptionen: () =>
        Promise.resolve({
          optionen: [
            {
              name: 'rahmen zeigen',
              art: 'boolean',
              beschreibung: '',
              wert: 'yes',
              moeglichkeiten: [],
              eigeneFarbe: false,
              gesetzt: false,
              hatVorgabe: false,
              hatWahlbilder: false,
              vorgabe: '',
            },
          ],
        }),
      StiloptionSetzen,
    })
    zeige()

    await userEvent.click(await screen.findByLabelText(/rahmen zeigen/))

    expect(StiloptionSetzen).toHaveBeenCalledWith('1', 'rahmen zeigen', 'no')
  })

  it('schreibt ein Textfeld erst, wenn es den Wert wirklich ändert', async () => {
    const StiloptionSetzen = vi.fn(() => Promise.resolve({}))
    belegeOptionen({ StiloptionSetzen })
    zeige()

    const feld = await screen.findByLabelText(/aufschrift/)

    await userEvent.click(feld)
    await userEvent.tab()

    expect(StiloptionSetzen).not.toHaveBeenCalled()

    await userEvent.type(feld, 'neu')
    await userEvent.tab()

    expect(StiloptionSetzen).toHaveBeenCalledWith('1', 'aufschrift', 'altneu')
  })

  it('meldet einen Fehler beim Laden', async () => {
    belegeOptionen({
      Stiloptionen: () =>
        Promise.resolve({
          optionen: [],
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'fehler.unbekannt',
            meldung: 'Der Stil ist weg.',
          },
        }),
    })
    zeige()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Der Stil ist weg.',
    )
  })

  it('meldet einen Fehler beim Schreiben', async () => {
    belegeOptionen({
      StiloptionSetzen: () =>
        Promise.resolve({
          fehler: {
            code: 'validierung' as const,
            schluessel: 'fehler.validierung',
            meldung: 'Diese Option gibt es nicht.',
          },
        }),
    })
    zeige()

    await userEvent.selectOptions(
      await screen.findByLabelText(/hintergrund/),
      'dunkel',
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Diese Option gibt es nicht.',
    )
  })
})
