// **Befehle in der Palette** — Werkstück U12 von Phase 5.1 (#1423).
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { Palette, passendeBefehle } from './Palette'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi } from '../test/api'
import type { BefehlDTO } from '../apibefehle'

const texte = kataloge.de.allgemein.palette

const befehle: BefehlDTO[] = [
  {
    name: 'lesen.sets',
    faehigkeit: 'Sets auflisten',
    cli: 'mse sets',
    rest: 'GET /sets',
    erklaerung: 'Zählt die Sets auf.',
    schreibend: false,
    idempotent: true,
    gefaehrlich: false,
    schema: {},
  },
  {
    name: 'schreiben.set.neu',
    faehigkeit: 'Ein leeres Set anlegen',
    cli: 'mse set-neu',
    rest: 'POST /sets/neu',
    erklaerung: 'Legt ein Set an.',
    schreibend: true,
    idempotent: false,
    gefaehrlich: true,
    schema: {},
  },
  {
    name: 'lesen.ohne',
    faehigkeit: '',
    cli: '',
    rest: '',
    erklaerung: 'ohne Fähigkeit',
    schreibend: false,
    idempotent: true,
    gefaehrlich: false,
    schema: {},
  },
]

describe('die Befehle in der Palette', () => {
  it('filtert über Name, Fähigkeit und Erklärung — nie ohne Eingabe', () => {
    expect(passendeBefehle(befehle, '')).toEqual([])
    expect(passendeBefehle(befehle, 'zzz')).toEqual([])
    // Der volle Name trifft genau — und steht vor allem anderen.
    expect(passendeBefehle(befehle, 'LESEN.SETS').map((b) => b.name)).toEqual([
      'lesen.sets',
    ])
    expect(passendeBefehle(befehle, 'SETS').map((b) => b.name)).toEqual([
      'lesen.sets',
    ])
    // **Der Name schlägt die Erklärung**, und Gleichrangiges steht alphabetisch.
    expect(passendeBefehle(befehle, 'lesen.se').map((b) => b.name)).toEqual([
      'lesen.sets',
    ])
    expect(passendeBefehle(befehle, 'set.neu').map((b) => b.name)).toEqual([
      'schreiben.set.neu',
    ])
    expect(passendeBefehle(befehle, 'lesen').map((b) => b.name)).toEqual([
      'lesen.ohne',
      'lesen.sets',
    ])
    expect(passendeBefehle(befehle, 'leeres').map((b) => b.name)).toEqual([
      'schreiben.set.neu',
    ])
    expect(passendeBefehle(befehle, 'legt').map((b) => b.name)).toEqual([
      'schreiben.set.neu',
    ])
  })

  it('zeigt die Befehle neben den Seiten und wählt einen', async () => {
    belegeApi({
      Befehlskatalog: () => Promise.resolve({ befehle, fehler: null }),
    })
    const waehleBefehl = vi.fn()
    const waehlen = vi.fn()

    render(
      <Sprachraum sprache="de">
        <Palette
          schliessen={vi.fn()}
          waehlen={waehlen}
          waehleBefehl={waehleBefehl}
        />
      </Sprachraum>,
    )

    await userEvent.type(screen.getByRole('textbox'), 'se')

    const seiten = await screen.findByRole('list', { name: texte.seiten })
    expect(
      within(seiten).getByRole('button', { name: /Sets/ }),
    ).toBeInTheDocument()

    const liste = await screen.findByRole('list', { name: texte.befehle })
    await userEvent.click(
      within(liste).getByRole('button', { name: /schreiben\.set\.neu/ }),
    )

    expect(waehleBefehl).toHaveBeenCalledWith(befehle[1])

    await userEvent.click(within(seiten).getByRole('button', { name: /Sets/ }))
    expect(waehlen).toHaveBeenCalledWith('sets')
  })

  // **Der lesbare Name zuerst, der Befehlsname klein dahinter** — AP-11
  // (#2368), B-27: `lesen.aehnlichekarten` vorn ist ein Name für Maschinen.
  it('nennt erst, was ein Befehl tut, und den Befehlsnamen klein dahinter', async () => {
    belegeApi({
      Befehlskatalog: () => Promise.resolve({ befehle, fehler: null }),
    })

    render(
      <Sprachraum sprache="de">
        <Palette
          schliessen={vi.fn()}
          waehlen={vi.fn()}
          waehleBefehl={vi.fn()}
        />
      </Sprachraum>,
    )

    await userEvent.type(screen.getByRole('textbox'), 'se')

    const liste = await screen.findByRole('list', { name: texte.befehle })
    const zeile = within(liste).getByRole('button', {
      name: /schreiben\.set\.neu/,
    })
    expect(zeile.firstChild).toHaveTextContent('Ein leeres Set anlegen')
    expect(within(zeile).getByText('schreiben.set.neu').tagName).toBe('CODE')
  })

  it('kommt ohne Katalog aus — dann gibt es nur Seiten', async () => {
    belegeApi({
      Befehlskatalog: () =>
        Promise.resolve({
          befehle: [],
          fehler: {
            code: 'nichtGefunden',
            meldung: 'leer',
            feld: '',
            schluessel: 'nichtGefunden',
          },
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Palette
          schliessen={vi.fn()}
          waehlen={vi.fn()}
          waehleBefehl={vi.fn()}
        />
      </Sprachraum>,
    )

    await userEvent.type(screen.getByRole('textbox'), 'zzz')
    expect(await screen.findByText(texte.keineTreffer)).toBeInTheDocument()
    expect(
      screen.queryByRole('list', { name: texte.befehle }),
    ).not.toBeInTheDocument()
  })
})
