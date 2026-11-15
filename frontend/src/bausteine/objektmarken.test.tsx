// **Marken an jedem Fachobjekt** — UX-12
// (#1684).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Markenfilter } from './Markenfilter'
import { Objektmarken } from './Objektmarken'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

const texte = kataloge.de.bausteine.objektmarken

afterEach(() => {
  raeumeApi()
})

describe('Die Marken eines Objekts', () => {
  it('setzt und nimmt eine Marke und sagt es dem Aufrufer', async () => {
    const setzen = vi.fn().mockResolvedValue({ fehler: null })
    const entfernen = vi.fn().mockResolvedValue({ fehler: null })
    const geaendert = vi.fn()
    belegeApi({ ObjektmarkeSetzen: setzen, ObjektmarkeEntfernen: entfernen })

    render(
      <Sprachraum sprache="de">
        <Objektmarken
          art="exemplar"
          kennung="31"
          marken={['tausch']}
          geaendert={geaendert}
        />
      </Sprachraum>,
    )

    await userEvent.type(
      screen.getByRole('textbox', { name: texte.setzenFuer }),
      'Binder 2',
    )
    await userEvent.click(screen.getByRole('button', { name: texte.setzen }))

    await waitFor(() => {
      expect(geaendert).toHaveBeenCalledTimes(1)
    })
    await antwortAbwarten()
    expect(setzen).toHaveBeenCalledWith('exemplar', '31', 'Binder 2')

    await userEvent.click(
      screen.getByRole('button', { name: '»tausch« entfernen' }),
    )
    await waitFor(() => {
      expect(geaendert).toHaveBeenCalledTimes(2)
    })
    await antwortAbwarten()
    expect(entfernen).toHaveBeenCalledWith('exemplar', '31', 'tausch')
  })

  it('zeigt einen Fehler am Feld und sagt ohne Marke, wozu sie da ist', async () => {
    belegeApi({
      ObjektmarkeSetzen: vi.fn().mockResolvedValue({
        fehler: {
          code: 'validierung',
          meldung: 'Eine Marke ist höchstens 40 Zeichen lang.',
          schluessel: '',
        },
      }),
    })

    render(
      <Sprachraum sprache="de">
        <Objektmarken art="deck" kennung="11" marken={[]} geaendert={vi.fn()} />
      </Sprachraum>,
    )

    expect(screen.getByText(texte.keine)).toBeInTheDocument()
    await userEvent.type(
      screen.getByRole('textbox', { name: texte.setzenFuer }),
      'x',
    )
    await userEvent.click(screen.getByRole('button', { name: texte.setzen }))

    expect(await screen.findByRole('alert')).toHaveTextContent('40 Zeichen')
  })
})

describe('Der Markenfilter', () => {
  it('wählt Marken und speichert die Wahl als Einstellung', async () => {
    const filtern = vi
      .fn()
      .mockResolvedValue({ kennungen: ['31'], fehler: null })
    const setzeFilter = vi.fn()
    const setzeEinstellung = vi.fn()
    belegeApi({ MarkenFiltern: filtern })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Markenfilter
          art="exemplar"
          je={{ '31': ['tausch', 'binder 2'], '32': ['tausch'] }}
          filter={{ marken: [], verknuepfung: 'und' }}
          setzeFilter={setzeFilter}
          gespeichert=""
          setzeGespeichert={setzeEinstellung}
          getroffen={undefined}
        />
      </Sprachraum>,
    )

    // **Die eigenen Marken dieser Sicht, jede einmal.**
    expect(screen.getByText(texte.gespeicherteKeine)).toBeInTheDocument()
    await userEvent.click(
      screen.getByRole('button', { name: 'Nach »tausch« filtern' }),
    )
    expect(setzeFilter).toHaveBeenCalledWith({
      marken: ['tausch'],
      verknuepfung: 'und',
    })

    rerender(
      <Sprachraum sprache="de">
        <Markenfilter
          art="exemplar"
          je={{ '31': ['tausch', 'binder 2'], '32': ['tausch'] }}
          filter={{ marken: ['tausch', 'binder 2'], verknuepfung: 'oder' }}
          setzeFilter={setzeFilter}
          gespeichert=""
          setzeGespeichert={setzeEinstellung}
          getroffen={new Set(['31'])}
        />
      </Sprachraum>,
    )

    expect(
      screen.getByText('1 passen zu den gewählten Marken.'),
    ).toBeInTheDocument()

    await userEvent.type(
      screen.getByRole('textbox', { name: texte.speichernName }),
      'Zum Tausch',
    )
    await userEvent.click(screen.getByRole('button', { name: texte.speichern }))

    expect(setzeEinstellung).toHaveBeenCalledWith(
      JSON.stringify([
        {
          name: 'Zum Tausch',
          sicht: 'exemplar',
          marken: ['tausch', 'binder 2'],
          verknuepfung: 'oder',
        },
      ]),
    )

    await userEvent.click(
      screen.getByRole('button', { name: texte.filterLeeren }),
    )
    expect(setzeFilter).toHaveBeenLastCalledWith({
      marken: [],
      verknuepfung: 'oder',
    })
    expect(filtern).not.toHaveBeenCalled()
  })

  it('wendet einen gespeicherten Filter an und löscht ihn wieder', async () => {
    const setzeFilter = vi.fn()
    const setzeEinstellung = vi.fn()
    const gespeichert = JSON.stringify([
      {
        name: 'Zum Tausch',
        sicht: 'exemplar',
        marken: ['tausch'],
        verknuepfung: 'oder',
      },
      { name: 'Rot', sicht: 'deck', marken: ['rot'], verknuepfung: 'und' },
    ])

    render(
      <Sprachraum sprache="de">
        <Markenfilter
          art="exemplar"
          je={{}}
          filter={{ marken: [], verknuepfung: 'und' }}
          setzeFilter={setzeFilter}
          gespeichert={gespeichert}
          setzeGespeichert={setzeEinstellung}
          getroffen={undefined}
        />
      </Sprachraum>,
    )

    expect(screen.getByText(texte.filterKeine)).toBeInTheDocument()

    // **Nur die Filter dieser Sicht** — „Rot" gehört den Decks.
    expect(
      screen.queryByRole('button', { name: 'Filter »Rot« anwenden' }),
    ).toBeNull()

    await userEvent.click(
      screen.getByRole('button', { name: 'Filter »Zum Tausch« anwenden' }),
    )
    expect(setzeFilter).toHaveBeenCalledWith({
      marken: ['tausch'],
      verknuepfung: 'oder',
    })

    await userEvent.click(
      screen.getByRole('button', { name: 'Filter »Zum Tausch« löschen' }),
    )
    expect(setzeEinstellung).toHaveBeenCalledWith(
      JSON.stringify([
        { name: 'Rot', sicht: 'deck', marken: ['rot'], verknuepfung: 'und' },
      ]),
    )
    await antwortAbwarten()
  })

  it.each(['[{', '{}'])(
    'liest eine kaputte Einstellung (%s) als keine',
    (gespeichert) => {
      render(
        <Sprachraum sprache="de">
          <Markenfilter
            art="deck"
            je={{}}
            filter={{ marken: [], verknuepfung: 'und' }}
            setzeFilter={vi.fn()}
            gespeichert={gespeichert}
            setzeGespeichert={vi.fn()}
            getroffen={undefined}
          />
        </Sprachraum>,
      )

      expect(screen.getByText(texte.gespeicherteKeine)).toBeInTheDocument()
    },
  )

  // **Gültiges JSON, aber falsche Einträge** — sie werden übergangen, die
  // übrigen gelten, und beim Speichern bleiben sie weg (Befund 4 des
  // Zweitblicks über UX-12): Go wiese die ganze Liste sonst ab.
  it('übergeht falsch geformte Einträge und speichert ohne sie', async () => {
    const setzeEinstellung = vi.fn()
    const gut = {
      name: 'Rot',
      sicht: 'deck',
      marken: ['rot'],
      verknuepfung: 'und',
    }
    const gespeichert = JSON.stringify([
      null,
      7,
      { name: 'ohne Marken', sicht: 'deck', verknuepfung: 'und' },
      { name: 'Zahl', sicht: 'deck', marken: [1], verknuepfung: 'und' },
      { name: 'Verknüpfung', sicht: 'deck', marken: [], verknuepfung: 'x' },
      { name: 3, sicht: 'deck', marken: [], verknuepfung: 'oder' },
      gut,
    ])

    render(
      <Sprachraum sprache="de">
        <Markenfilter
          art="deck"
          je={{ '1': ['rot'] }}
          filter={{ marken: ['rot'], verknuepfung: 'oder' }}
          setzeFilter={vi.fn()}
          gespeichert={gespeichert}
          setzeGespeichert={setzeEinstellung}
          getroffen={undefined}
        />
      </Sprachraum>,
    )

    expect(
      screen.getByRole('button', { name: 'Filter »Rot« anwenden' }),
    ).toBeInTheDocument()
    expect(
      screen.getAllByRole('button', { name: /^Filter ».*« anwenden$/ }),
    ).toHaveLength(1)

    await userEvent.type(
      screen.getByRole('textbox', { name: texte.speichernName }),
      'Neu',
    )
    await userEvent.click(screen.getByRole('button', { name: texte.speichern }))
    expect(setzeEinstellung).toHaveBeenCalledWith(
      JSON.stringify([
        gut,
        { name: 'Neu', sicht: 'deck', marken: ['rot'], verknuepfung: 'oder' },
      ]),
    )
  })
})

describe('Der Markenfilter, im Einzelnen', () => {
  it('nimmt eine gewählte Marke wieder ab, wechselt die Verknüpfung und speichert nicht ohne Namen', async () => {
    const setzeFilter = vi.fn()
    const setzeGespeichert = vi.fn()

    render(
      <Sprachraum sprache="de">
        <Markenfilter
          art="deck"
          je={{ '11': ['rot'] }}
          filter={{ marken: ['rot'], verknuepfung: 'und' }}
          setzeFilter={setzeFilter}
          gespeichert=""
          setzeGespeichert={setzeGespeichert}
          getroffen={new Set(['11'])}
        />
      </Sprachraum>,
    )

    await userEvent.click(
      screen.getByRole('button', { name: 'Nach »rot« filtern' }),
    )
    expect(setzeFilter).toHaveBeenLastCalledWith({
      marken: [],
      verknuepfung: 'und',
    })

    await userEvent.click(screen.getByRole('radio', { name: texte.oder }))
    expect(setzeFilter).toHaveBeenLastCalledWith({
      marken: ['rot'],
      verknuepfung: 'oder',
    })

    // **Die Eingabetaste in einem leeren Feld speichert nichts.**
    await userEvent.type(
      screen.getByRole('textbox', { name: texte.speichernName }),
      '{Enter}',
    )
    expect(setzeGespeichert).not.toHaveBeenCalled()
  })
})
