// **Die Turnierseite** — Werkstück 5 von Phase 4.7
// (#1294),
// nach Vermerk 0062.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Turniere } from './Turniere'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { zumTurniergebiet } from './turnierwerkbank'
import { kataloge } from '../sprache/kataloge'
import { turnierhandlung } from '../test/turnierhandlung'
import { assistent } from '../test/turnierassistent'
import {
  rangprobe,
  rundenprobe,
  teilnehmerprobe,
  turnierprobe,
} from './turnierprobe'
import type { TurnierDTO } from '../apiturniere'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

// zeige stellt die Seite mit einem Turnier hin.
function zeige(
  turnier: TurnierDTO | null = turnierprobe(),
  weitere: Record<string, unknown> = {},
) {
  belegeApi({
    Turniere: () =>
      Promise.resolve({
        turniere: turnier === null ? [] : [turnier],
        fehler: null,
      }),
    Turnier: () => Promise.resolve({ turnier, fehler: null }),
    TurnierRangliste: () => Promise.resolve({ rangliste: [], fehler: null }),
    ...weitere,
  })

  render(
    <Sprachraum sprache="de">
      <Turniere />
    </Sprachraum>,
  )
}

describe('die Turnierseite', () => {
  it('sagt es, wenn noch kein Turnier da ist', async () => {
    zeige(null)

    expect(await screen.findByText(texte.leer)).toBeInTheDocument()
  })

  it('meldet einen Ladefehler', async () => {
    zeige(null, {
      Turniere: () =>
        Promise.resolve({
          turniere: [],
          fehler: {
            code: 'unbekannt',
            schluessel: 'fehler.unbekannt',
            meldung: 'Der Bestand schweigt.',
          },
        }),
    })

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Der Bestand schweigt.',
    )
  })

  it('legt ein Turnier an', async () => {
    const TurnierAnlegen = vi.fn(() =>
      Promise.resolve({ turnier: turnierprobe(), fehler: null }),
    )

    zeige(turnierprobe(), { TurnierAnlegen })

    const bereich = await assistent()

    await userEvent.type(within(bereich).getByLabelText(texte.name), 'Freitag')
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.weiter }),
    )
    await userEvent.click(within(bereich).getByLabelText(texte.einzelspiel))
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.anlegen }),
    )

    await waitFor(() => {
      expect(TurnierAnlegen).toHaveBeenCalledWith(
        'Freitag',
        'magic',
        '',
        true,
        '',
      )
    })
  })

  it('meldet einen Fehler beim Anlegen am Feld', async () => {
    zeige(turnierprobe(), {
      TurnierAnlegen: () =>
        Promise.resolve({
          turnier: null,
          fehler: {
            code: 'validierung',
            schluessel: 'fehler.turniernameLeer',
            meldung: 'Ein Turnier braucht einen Namen.',
            feld: 'name',
          },
        }),
    })

    const bereich = await assistent()

    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.weiter }),
    )
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.anlegen }),
    )

    await waitFor(() => {
      expect(screen.getAllByRole('alert').length).toBeGreaterThan(0)
    })
  })

  it('schlägt ein Turnier auf und zeigt seine Teile', async () => {
    zeige(
      turnierprobe({
        teilnehmer: [
          teilnehmerprobe(),
          teilnehmerprobe({ id: '4', name: 'Bert', ausgestiegen: true }),
        ],
        runden: [rundenprobe()],
      }),
      {
        TurnierRangliste: () =>
          Promise.resolve({ rangliste: [rangprobe()], fehler: null }),
      },
    )

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.waehlen}: Küchentisch`,
      }),
    )

    const werkbank = await screen.findByRole('region', { name: 'Küchentisch' })

    // **Ein Turnier mit Runden schlägt bei den Runden auf** — Werkstück I von
    // Phase 4.10. Wer die Teilnehmer sehen will, geht hin.
    await zumTurniergebiet('teilnehmer')

    expect(
      within(werkbank).getByRole('list', { name: texte.teilnehmer }),
    ).toBeInTheDocument()

    expect(within(werkbank).getByText(texte.ausgestiegen)).toBeInTheDocument()

    // **Die Rangliste ist eine Tabelle mit Kopfzellen** — im Gebiet
    // »Rangliste« (Werkstück I von Phase 4.10).
    await zumTurniergebiet('rangliste')

    expect(
      within(werkbank).getByRole('columnheader', { name: texte.matchpunkte }),
    ).toBeInTheDocument()

    // **Und die Paarungen einer Runde stehen als geordnete Liste** — die
    // Reihenfolge der Tische wird damit angesagt.
    await zumTurniergebiet('runden')

    expect(
      within(werkbank).getByRole('list', {
        name: texte.runde.replace('{nummer}', '1'),
      }),
    ).toBeInTheDocument()
  })

  it('löscht ein Turnier erst nach Rückfrage', async () => {
    const TurnierLoeschen = vi.fn(() => Promise.resolve({ fehler: null }))

    zeige(turnierprobe(), { TurnierLoeschen })

    // **Gelöscht wird am offenen Turnier** — seit UXR-12 (#1920) steht der
    // Knopf nicht mehr an jeder Zeile der Liste.
    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.waehlen}: Küchentisch`,
      }),
    )
    await turnierhandlung('Küchentisch', texte.loeschen)

    // **Erst abbrechen** — die Rückfrage ist keine Formsache.
    await userEvent.click(
      screen.getByRole('button', { name: texte.loeschenNein }),
    )

    expect(TurnierLoeschen).not.toHaveBeenCalled()

    await turnierhandlung('Küchentisch', texte.loeschen)
    await userEvent.click(
      screen.getByRole('button', { name: texte.loeschenJa }),
    )

    await waitFor(() => {
      expect(TurnierLoeschen).toHaveBeenCalledWith('7')
    })
  })
})
