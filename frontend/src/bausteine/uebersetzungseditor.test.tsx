// **Der Editor einer eigenen Übersetzung** — LOC-02.AK-01
// (#1688).
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Uebersetzungseditor } from './Uebersetzungseditor'
import { aufAenderung } from '../sprache/eigenesprache'
import { flach } from '../sprache/flach'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { sprachantwort, sprachpaketprobe } from '../test/sprachprobe'

import type { Api } from '../api'

const texte = kataloge.de.seiten.einstellungen.eigenuebersetzung
const pfade = Object.keys(flach(kataloge.de)).sort()
const fehlend = pfade.slice(0, 30)
const uebersetzt = pfade[30] ?? ''
const veraltet = pfade[31] ?? ''

afterEach(() => {
  raeumeApi()
})

const stand = sprachantwort({
  paket: sprachpaketprobe({ kennung: 'da', name: 'Dansk' }),
  texte: {
    [uebersetzt]: { text: 'Oversat', quelle: '' },
    [veraltet]: { text: 'Gammel', quelle: 'alt' },

    // **Ein Pfad aus einer neueren Fassung** — ohne deutschen Satz und ohne
    // Basis, aber mit seiner Übersetzung.
    'zz.neu': { text: 'Ny', quelle: 'Neu' },
  },
  fehlendePfade: fehlend,
  veraltetePfade: [veraltet],
})

function zeige(teile: Partial<Api> = {}) {
  const zurueck = vi.fn()

  belegeApi({ Sprachpaket: () => Promise.resolve(stand), ...teile })

  render(
    <Sprachraum sprache="de">
      <Uebersetzungseditor kennung="da" zurueck={zurueck} />
    </Sprachraum>,
  )

  return zurueck
}

describe('Der Editor einer eigenen Übersetzung', () => {
  it('zeigt die fehlenden Sätze seitenweise', async () => {
    zeige()

    expect(
      await screen.findByRole('heading', { name: 'Dansk (da)' }),
    ).toBeInTheDocument()
    expect(screen.getAllByRole('listitem')).toHaveLength(25)
    expect(screen.getByText('Seite 1 von 2')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: texte.vor }))
    expect(screen.getAllByRole('listitem')).toHaveLength(5)

    await userEvent.click(
      screen.getByRole('button', { name: texte.zurueckBlaettern }),
    )
    expect(screen.getByText('Seite 1 von 2')).toBeInTheDocument()
  })

  it('übernimmt einen Satz und meldet dem Fenster die Änderung', async () => {
    const gehoert: string[] = []
    const aufhoeren = aufAenderung((kennung) => gehoert.push(kennung))
    const SprachpaketAendern = vi.fn(() =>
      Promise.resolve(
        sprachantwort({ ...stand, fehlendePfade: fehlend.slice(1) }),
      ),
    )
    zeige({ SprachpaketAendern })

    const feld = await screen.findByRole('textbox', {
      name: `Übersetzung von ${fehlend[0] ?? ''}`,
    })
    await userEvent.type(feld, 'Sæt')
    await userEvent.click(
      within(feld.closest('li') as HTMLElement).getByRole('button', {
        name: new RegExp(texte.uebernehmen),
      }),
    )

    expect(SprachpaketAendern).toHaveBeenCalledWith(
      'da',
      { name: null, stand: null, lizenz: null, herkunft: null },
      { [fehlend[0] ?? '']: 'Sæt' },
    )
    await waitFor(() => {
      expect(gehoert).toEqual(['da'])
    })
    aufhoeren()
  })

  // **Der Befund steht am Satz** (§6).
  it('zeigt einen falschen Satz an seinem Feld', async () => {
    zeige({
      SprachpaketAendern: () =>
        Promise.resolve(
          sprachantwort({
            fehler: {
              code: 'validierung',
              schluessel: 'platzhalterAbweichend',
              meldung: 'x',
              feld: fehlend[0] ?? '',
              werte: { pfad: fehlend[0] ?? '', erwartet: '{name}' },
            },
          }),
        ),
    })

    const feld = await screen.findByRole('textbox', {
      name: `Übersetzung von ${fehlend[0] ?? ''}`,
    })
    await userEvent.type(feld, 'x')
    await userEvent.click(
      within(feld.closest('li') as HTMLElement).getByRole('button', {
        name: new RegExp(texte.uebernehmen),
      }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent('{name}')
    expect(feld).toHaveAttribute('aria-invalid', 'true')
  })

  it('zeigt Übersetztes und Veraltetes und nimmt zurück', async () => {
    const SprachpaketAendern = vi.fn(() => Promise.resolve(stand))
    zeige({ SprachpaketAendern })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.filterVeraltet }),
    )
    expect(screen.getByText(texte.veraltetMarke)).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: texte.filterUebersetzt }),
    )
    expect(screen.getByDisplayValue('Oversat')).toBeInTheDocument()
    expect(screen.getByDisplayValue('Ny')).toBeInTheDocument()

    const zeile = screen
      .getByDisplayValue('Oversat')
      .closest('li') as HTMLElement
    await userEvent.click(
      within(zeile).getByRole('button', {
        name: new RegExp(texte.zuruecknehmen),
      }),
    )

    expect(SprachpaketAendern).toHaveBeenCalledWith('da', expect.anything(), {
      [uebersetzt]: '',
    })
  })

  it('sucht und sagt, wenn nichts passt', async () => {
    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: texte.filterAlle }),
    )
    await userEvent.type(
      screen.getByRole('searchbox', { name: texte.suche }),
      'Oversat',
    )
    expect(screen.getAllByRole('listitem')).toHaveLength(1)

    await userEvent.type(
      screen.getByRole('searchbox', { name: texte.suche }),
      'zzz',
    )
    expect(screen.getByText(texte.keinTreffer)).toBeInTheDocument()
  })

  it('sichert die Angaben und zeigt ihren Befund', async () => {
    const SprachpaketAendern = vi.fn(() =>
      Promise.resolve(
        sprachantwort({
          fehler: {
            code: 'validierung',
            schluessel: 'validierung',
            meldung: 'Ein Sprachpaket braucht einen Namen.',
            feld: 'name',
            werte: {},
          },
        }),
      ),
    )
    zeige({ SprachpaketAendern })

    const lizenz = await screen.findByRole('textbox', { name: texte.lizenz })
    await userEvent.type(lizenz, 'CC0-1.0')
    await userEvent.type(
      screen.getByRole('textbox', { name: texte.herkunft }),
      'Ana',
    )
    await userEvent.selectOptions(
      screen.getByRole('combobox', { name: texte.stand }),
      'geprueft',
    )
    await userEvent.clear(screen.getByRole('textbox', { name: texte.name }))
    await userEvent.click(
      screen.getByRole('button', { name: new RegExp(texte.angabenSichern) }),
    )

    expect(SprachpaketAendern).toHaveBeenCalledWith(
      'da',
      { name: '', stand: 'geprueft', lizenz: 'CC0-1.0', herkunft: 'Ana' },
      {},
    )
    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })

  it('meldet, wenn sich die Übersetzung nicht laden lässt, und geht zurück', async () => {
    const zurueck = zeige({
      Sprachpaket: () =>
        Promise.resolve(
          sprachantwort({
            fehler: {
              code: 'nichtGefunden',
              schluessel: 'nichtGefunden',
              meldung: 'fort',
              feld: '',
              werte: {},
            },
          }),
        ),
    })

    expect(await screen.findByRole('alert')).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'da' })).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: texte.zurueck }))
    expect(zurueck).toHaveBeenCalled()
  })

  it('meldet einen Fehler ohne Feld bei den Angaben', async () => {
    zeige({
      SprachpaketAendern: () =>
        Promise.resolve(
          sprachantwort({
            fehler: {
              code: 'unbekannt',
              schluessel: 'unbekannt',
              meldung: 'x',
              werte: {},
            },
          }),
        ),
    })

    await userEvent.click(
      await screen.findByRole('button', {
        name: new RegExp(texte.angabenSichern),
      }),
    )

    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })

  it('meldet einen abgewiesenen Ruf am Editor', async () => {
    zeige({
      SprachpaketAendern: () =>
        Promise.reject(new Error('die Brücke schweigt')),
    })

    await userEvent.click(
      await screen.findByRole('button', {
        name: new RegExp(texte.angabenSichern),
      }),
    )

    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })
})
