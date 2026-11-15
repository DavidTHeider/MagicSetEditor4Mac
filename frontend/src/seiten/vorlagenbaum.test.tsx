import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Vorlagenbaum } from './Vorlagenbaum'

const texte = kataloge.de.seiten.vorlagen
const geladen = { geladen: true, fehlschlag: false } // die Liste ist da

const dateienJePaket: Record<string, string[]> = {
  '/pakete/probe.mse-style': ['style', 'skript.txt'],
  '/pakete/leer.mse-style': [],
}

// fehlschlag ist die Antwort der Werkstatt, wenn sie ein Paket nicht aufzählt.
function fehlschlag(code: 'validierung' | 'nichtGefunden', meldung: string) {
  return {
    dateien: [],
    beiwerk: [],
    fehler: { code, schluessel: code, meldung },
  }
}

// antwortJePaket ist, was die Werkstatt für ein Paket nennt — ohne Beiwerk.
function antwortJePaket(paket: string) {
  return {
    dateien: dateienJePaket[paket] ?? ['game'],
    beiwerk: [],
    fehler: null,
  }
}

function zeige(
  waehlen = vi.fn(),
  gewaehlt: { paket: string; datei: string } | null = null,
  pakete = [
    '/pakete/probe.mse-style',
    '/pakete/zweit.mse-game',
    '/pakete/leer.mse-style',
  ],
) {
  belegeApi({
    Paketdateien: (paket: string) => Promise.resolve(antwortJePaket(paket)),
  })

  render(
    <Sprachraum sprache="de">
      <Vorlagenbaum
        {...geladen}
        pakete={pakete}
        gewaehlt={gewaehlt}
        waehlen={waehlen}
        geaendert={false}
      />
    </Sprachraum>,
  )

  return waehlen
}

describe('Der Vorlagenbaum', () => {
  afterEach(raeumeApi)

  it('zeigt jedes Paket mit seinen Dateien aufgeklappt', async () => {
    zeige()

    expect(
      await screen.findByRole('button', { name: 'style' }),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'skript.txt' }),
    ).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'game' })).toBeInTheDocument()
    expect(screen.getByText(texte.keineDateien)).toBeInTheDocument()

    const paket = screen.getByRole('button', { name: 'probe.mse-style' })
    expect(paket).toHaveAttribute('aria-expanded', 'true')
  })

  it('reicht Paket und Datei beim Klick weiter', async () => {
    const waehlen = zeige()

    await userEvent.click(await screen.findByRole('button', { name: 'game' }))

    expect(waehlen).toHaveBeenCalledWith('/pakete/zweit.mse-game', 'game')
  })

  it('markiert die gewählte Datei', async () => {
    zeige(vi.fn(), { paket: '/pakete/probe.mse-style', datei: 'style' })

    expect(
      await screen.findByRole('button', { name: 'style' }),
    ).toHaveAttribute('aria-current', 'true')
    expect(screen.getByRole('button', { name: 'game' })).not.toHaveAttribute(
      'aria-current',
    )
  })

  it('klappt ein Paket zu und wieder auf', async () => {
    zeige()

    const paket = await screen.findByRole('button', { name: 'probe.mse-style' })
    await screen.findByRole('button', { name: 'style' })

    await userEvent.click(paket)

    expect(paket).toHaveAttribute('aria-expanded', 'false')
    expect(
      screen.queryByRole('button', { name: 'style' }),
    ).not.toBeInTheDocument()

    await userEvent.click(paket)

    expect(
      await screen.findByRole('button', { name: 'style' }),
    ).toBeInTheDocument()
  })

  it('filtert über Paket- und Dateinamen', async () => {
    zeige()

    await screen.findByRole('button', { name: 'game' })

    const suche = screen.getByLabelText(texte.baumSuche)
    await userEvent.type(suche, 'skript')

    expect(
      screen.getByRole('button', { name: 'skript.txt' }),
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: 'style' }),
    ).not.toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: 'game' }),
    ).not.toBeInTheDocument()

    await userEvent.clear(suche)
    await userEvent.type(suche, 'zweit')

    expect(screen.getByRole('button', { name: 'game' })).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: 'probe.mse-style' }),
    ).not.toBeInTheDocument()

    await userEvent.clear(suche)
    await userEvent.type(suche, 'gibtesnicht')

    expect(screen.getByText(texte.baumLeer)).toBeInTheDocument()

    await userEvent.keyboard('{Escape}')

    expect(suche).toHaveValue('')
  })

  it('läuft mit den Pfeiltasten durch den Baum', async () => {
    zeige()

    const paket = await screen.findByRole('button', { name: 'probe.mse-style' })
    await screen.findByRole('button', { name: 'game' })

    paket.focus()
    await userEvent.keyboard('{ArrowDown}')

    expect(screen.getByRole('button', { name: 'style' })).toHaveFocus()

    await userEvent.keyboard('{ArrowDown}{ArrowDown}')

    expect(screen.getByRole('button', { name: 'zweit.mse-game' })).toHaveFocus()

    await userEvent.keyboard('{ArrowUp}')

    expect(screen.getByRole('button', { name: 'skript.txt' })).toHaveFocus()

    await userEvent.keyboard('{End}')

    expect(screen.getByRole('button', { name: 'leer.mse-style' })).toHaveFocus()

    await userEvent.keyboard('{Home}')

    expect(paket).toHaveFocus()

    // **Eine andere Taste tut nichts im Baum** — `Tab` verlässt ihn, statt
    // auf den nächsten Knopf zu springen (eine Tabstelle, siehe unten).
    await userEvent.keyboard('{Tab}')

    expect(screen.getByRole('button', { name: 'style' })).not.toHaveFocus()
  })

  // **Eine Tabstelle für den ganzen Baum** (#1290, hier für U7): `Tab` landet
  // auf der ersten Datei des ersten bearbeitbaren Pakets, die Pfeile laufen
  // weiter, und wer den Baum verlässt und zurückkommt, findet seinen Halt.
  it('hat eine Tabstelle und wandert von dort mit den Pfeilen', async () => {
    zeige()

    await screen.findByRole('button', { name: 'game' })

    screen.getByLabelText(texte.baumSuche).focus()
    await userEvent.tab()

    expect(screen.getByRole('button', { name: 'style' })).toHaveFocus()
    expect(
      screen.getByRole('button', { name: 'probe.mse-style' }),
    ).toHaveAttribute('tabindex', '-1')

    await userEvent.keyboard('{ArrowUp}')
    expect(
      screen.getByRole('button', { name: 'probe.mse-style' }),
    ).toHaveFocus()

    await userEvent.tab()
    expect(screen.getByRole('button', { name: 'game' })).not.toHaveFocus()
    expect(screen.getByRole('button', { name: 'style' })).not.toHaveFocus()

    await userEvent.tab({ shift: true })
    expect(
      screen.getByRole('button', { name: 'probe.mse-style' }),
    ).toHaveFocus()
  })

  it('macht die gewählte Datei zur Tabstelle', async () => {
    zeige(vi.fn(), { paket: '/pakete/zweit.mse-game', datei: 'game' })

    await screen.findByRole('button', { name: 'game' })
    screen.getByLabelText(texte.baumSuche).focus()
    await userEvent.tab()

    expect(screen.getByRole('button', { name: 'game' })).toHaveFocus()
  })

  it('stellt mitgelieferte Pakete unter die bearbeitbaren', async () => {
    belegeApi({
      Paketdateien: (paket: string) =>
        Promise.resolve(
          paket === '/saat/probe.mse-style'
            ? fehlschlag('validierung', 'liegt nicht unter der Wurzel')
            : { dateien: ['style'], beiwerk: [], fehler: null },
        ),
    })

    render(
      <Sprachraum sprache="de">
        <Vorlagenbaum
          {...geladen}
          pakete={['/saat/probe.mse-style', '/pakete/eigen.mse-style']}
          gewaehlt={null}
          waehlen={vi.fn()}
          geaendert={false}
        />
      </Sprachraum>,
    )

    await screen.findByText(texte.nurLesbar)
    const namen = screen
      .getAllByRole('button', { expanded: true })
      .map((knopf) => knopf.textContent.trim())

    expect(namen).toEqual(['▾eigen.mse-style', '▾probe.mse-style'])
  })

  it('meldet den Fehler eines Pakets an seinem Zweig', async () => {
    belegeApi({
      Paketdateien: (paket: string) =>
        Promise.resolve(
          paket === '/pakete/kaputt.mse-style'
            ? {
                dateien: [],
                beiwerk: [],
                fehler: {
                  code: 'nichtGefunden',
                  schluessel: 'nichtGefunden',
                  meldung: 'weg',
                },
              }
            : { dateien: ['style'], beiwerk: [], fehler: null },
        ),
    })

    render(
      <Sprachraum sprache="de">
        <Vorlagenbaum
          {...geladen}
          pakete={['/pakete/kaputt.mse-style', '/pakete/probe.mse-style']}
          gewaehlt={null}
          waehlen={vi.fn()}
          geaendert={false}
        />
      </Sprachraum>,
    )

    const meldung = await screen.findByRole('alert')
    expect(meldung).toHaveTextContent(kataloge.de.fehler.nichtGefunden)
    expect(
      await screen.findByRole('button', { name: 'style' }),
    ).toBeInTheDocument()

    const zweig = screen
      .getByRole('button', { name: 'kaputt.mse-style' })
      .closest('li')
    expect(zweig).not.toBeNull()
    expect(within(zweig as HTMLElement).getByRole('alert')).toBe(meldung)
  })

  it('nennt ein mitgeliefertes Paket als nur lesbar statt als Fehler', async () => {
    belegeApi({
      Paketdateien: (paket: string) =>
        Promise.resolve(
          paket === '/saat/probe.mse-style'
            ? fehlschlag('validierung', 'liegt nicht unter der Wurzel')
            : { dateien: ['style'], beiwerk: [], fehler: null },
        ),
    })

    render(
      <Sprachraum sprache="de">
        <Vorlagenbaum
          {...geladen}
          pakete={['/saat/probe.mse-style', '/pakete/eigen.mse-style']}
          gewaehlt={null}
          waehlen={vi.fn()}
          geaendert={false}
        />
      </Sprachraum>,
    )

    expect(await screen.findByText(texte.nurLesbar)).toBeInTheDocument()
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    expect(screen.queryByText(texte.keineDateien)).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'style' })).toBeInTheDocument()
  })

  it('sagt einmal, wenn alles mitgeliefert ist', async () => {
    belegeApi({
      Paketdateien: () =>
        Promise.resolve(
          fehlschlag('validierung', 'liegt nicht unter der Wurzel'),
        ),
    })

    render(
      <Sprachraum sprache="de">
        <Vorlagenbaum
          {...geladen}
          pakete={['/saat/a.mse-style', '/saat/b.mse-style']}
          gewaehlt={null}
          waehlen={vi.fn()}
          geaendert={false}
        />
      </Sprachraum>,
    )

    expect(
      await screen.findByText(texte.baumNurMitgeliefert),
    ).toBeInTheDocument()
    expect(screen.getAllByText(texte.nurLesbar)).toHaveLength(2)
  })

  it('sagt ohne Pakete, was zu tun ist', async () => {
    zeige(vi.fn(), null, [])

    await waitFor(() => {
      expect(screen.getByText(texte.baumOhnePakete)).toBeInTheDocument()
    })
  })

  it('nennt die ungespeicherte Datei', async () => {
    belegeApi({
      Paketdateien: () =>
        Promise.resolve({ dateien: ['style'], beiwerk: [], fehler: null }),
    })

    render(
      <Sprachraum sprache="de">
        <Vorlagenbaum
          {...geladen}
          pakete={['/pakete/probe.mse-style']}
          gewaehlt={{ paket: '/pakete/probe.mse-style', datei: 'style' }}
          waehlen={vi.fn()}
          geaendert
        />
      </Sprachraum>,
    )

    expect(
      await screen.findByRole('button', {
        name: `style — ${texte.ungespeichert}`,
      }),
    ).toBeInTheDocument()
  })
})
