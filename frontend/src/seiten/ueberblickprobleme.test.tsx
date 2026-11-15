// **Die Probleme auf den Kacheln** — AP-8, Teil 2
// (#2365),
// Befund B-13, Entwurf `Start09`: „Unvollständig heißt nie 0.“ Eine Zahl
// steht nur da, wenn jedes Set aktuell geprüft ist; sonst sagt die Kachel,
// wie viele geprüft sind, und die Zeile darunter bietet den Anstoß an —
// geprüft wird nie von selbst (PERF02, `LEISTUNGSSCHNITTE.md` §11).
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Anschluss } from '../anschluss'
import type { Anschlussstelle } from '../anschlussraum'
import type { SetproblemstandDTO } from '../apiprobleme'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Start } from './Start'

const texte = kataloge.de.seiten.start

const uebergeben: Anschluss[] = []
vi.mock('../anschlussraum', async (echt) => ({
  ...(await echt<typeof import('../anschlussraum')>()),
  useAnschluss: (): Anschlussstelle => ({
    offen: undefined,
    uebergib: (anschluss) => {
      uebergeben.push(anschluss)
    },
    verbrauche: () => undefined,
  }),
}))

afterEach(() => {
  raeumeApi()
})

function stand(
  setID: string,
  teil: Partial<SetproblemstandDTO> = {},
): SetproblemstandDTO {
  return {
    setID,
    name: `Set ${setID}`,
    urteil: 'aktuell',
    zahl: 0,
    sicher: true,
    geprueft: 5,
    gesamt: 5,
    beendet: '2026-10-17T03:00:00Z',
    alter: 240,
    karten: [],
    ...teil,
  }
}

function zeige(
  staende: SetproblemstandDTO[],
  geprueft: SetproblemstandDTO[] = staende,
  haengt = false,
) {
  const SetproblemePruefen = vi.fn(() =>
    Promise.resolve({ staende: geprueft, fehler: null }),
  )
  let gelesen = staende
  belegeApi({
    Sets: () =>
      Promise.resolve({
        sets: staende.map((eines) => ({
          id: eines.setID,
          name: eines.name,
          spiel: 'magic',
          stil: 'magic-m15',
          karten: 5,
          herkunft: '',
          herkunftFehlt: false,
        })),
        fehler: null,
      }),
    Decks: () => Promise.resolve({ decks: [], fehler: null }),
    Sammlungsfenster: () =>
      Promise.resolve({
        exemplare: [],
        gesamt: 0,
        insgesamt: 0,
        stueck: 0,
        fehler: null,
      }),
    Setprobleme: () =>
      haengt
        ? new Promise(() => undefined)
        : Promise.resolve({ staende: gelesen, fehler: null }),
    SetproblemePruefen: vi.fn(async () => {
      const antwort = await SetproblemePruefen()
      gelesen = antwort.staende

      return antwort
    }),
  })
  const wechseln = vi.fn()
  const waehleSet = vi.fn()
  render(
    <Sprachraum sprache="de">
      <Start wechseln={wechseln} waehleSet={waehleSet} bestandsstand={0} />
    </Sprachraum>,
  )

  return { SetproblemePruefen, wechseln, waehleSet }
}

const ueberblick = () =>
  within(screen.getByRole('list', { name: texte.ueberblick.titel }))

describe('die Probleme auf den Kacheln', () => {
  it('nennen Sets mit Problemen und Karten mit Befund, wenn alles geprüft ist', async () => {
    zeige([stand('1', { zahl: 3 }), stand('2', { zahl: 0 })])

    expect(
      await ueberblick().findByRole('button', {
        name: /2 Sets 1 mit Problemen/,
      }),
    ).toBeInTheDocument()
    expect(
      ueberblick().getByRole('button', { name: /10 Karten 3 mit Befund/ }),
    ).toBeInTheDocument()
    expect(
      screen.getByText('Probleme geprüft vor 4 min · 10 von 10 Karten'),
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: texte.ueberblick.pruefen }),
    ).toBeNull()
  })

  it('sagen nie 0, solange ein Set ungeprüft ist, und bieten das Prüfen an', async () => {
    zeige([
      stand('1', { zahl: 0 }),
      stand('2', {
        urteil: 'ungeprueft',
        sicher: false,
        geprueft: 0,
        alter: 0,
        beendet: '',
      }),
    ])

    expect(
      await ueberblick().findByRole('button', {
        name: /2 Sets 1 von 2 geprüft/,
      }),
    ).toBeInTheDocument()
    expect(
      ueberblick().getByRole('button', {
        name: /10 Karten nicht vollständig geprüft/,
      }),
    ).toBeInTheDocument()
    expect(screen.queryByText(/0 mit Problemen|0 mit Befund/)).toBeNull()
    expect(
      screen.getByText('Probleme: 1 von 2 Sets nicht geprüft'),
    ).toBeInTheDocument()
  })

  it('prüfen auf Anstoß und zeigen danach die Zahlen', async () => {
    const offen = stand('2', { urteil: 'veraltet', sicher: false })
    const { SetproblemePruefen } = zeige(
      [stand('1'), offen],
      [stand('1'), stand('2', { zahl: 2 })],
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.ueberblick.pruefen }),
    )

    expect(SetproblemePruefen).toHaveBeenCalledTimes(1)
    expect(
      await ueberblick().findByRole('button', {
        name: /10 Karten 2 mit Befund/,
      }),
    ).toBeInTheDocument()
  })

  it('führen zu den Sets mit Problemen, gefiltert', async () => {
    const { wechseln } = zeige([
      stand('1', { zahl: 3 }),
      stand('2', { zahl: 1 }),
      stand('3'),
    ])

    await userEvent.click(
      await ueberblick().findByRole('button', {
        name: /3 Sets 2 mit Problemen/,
      }),
    )
    expect(uebergeben.at(-1)).toEqual({
      art: 'setauswahl',
      name: '2 Sets mit Problemen',
      setIDs: ['1', '2'],
    })

    await userEvent.click(
      ueberblick().getByRole('button', { name: /15 Karten 4 mit Befund/ }),
    )
    expect(uebergeben.at(-1)).toEqual({
      art: 'setauswahl',
      name: '4 Karten mit Befund in 2 Sets',
      setIDs: ['1', '2'],
    })
    expect(wechseln).not.toHaveBeenCalledWith('sets')
  })

  // **Ist es ein Set, führt die Kachel „Karten“ in seine Kartenliste** —
  // gefiltert auf die Karten mit Befund (AP-8, Teil 2b): die Setliste mit
  // einer Zeile wäre ein Umweg.
  it('führen bei einem Set in seine Karten mit Befund', async () => {
    const { waehleSet } = zeige([
      stand('1', {
        zahl: 2,
        karten: [
          { karteID: '7', befunde: ['a'] },
          { karteID: '9', befunde: ['b'] },
        ],
      }),
      stand('2'),
    ])

    await userEvent.click(
      await ueberblick().findByRole('button', {
        name: /10 Karten 2 mit Befund/,
      }),
    )
    expect(waehleSet).toHaveBeenCalledWith('1', '7')
    expect(uebergeben.at(-1)).toEqual({
      art: 'kartenauswahl',
      name: '2 Karten mit Befund',
      setID: '1',
      karteIDs: ['7', '9'],
    })
  })

  // **Ohne Probleme oder ungeprüft führt die Kachel zur ganzen Liste** — ein
  // leerer Filter wäre eine Sackgasse, einer über Ungeprüftes eine Behauptung.
  it('führen ohne bekannte Probleme zur ganzen Liste', async () => {
    const { wechseln } = zeige([stand('1'), stand('2')])

    await userEvent.click(
      await ueberblick().findByRole('button', {
        name: /2 Sets 0 mit Problemen/,
      }),
    )

    expect(wechseln).toHaveBeenCalledWith('sets')
  })

  // **Die Zahlen warten nicht auf die Prüfstände** — im selben Ladelauf
  // stand die Kennzahl bei 50 Sets kalt erst nach 526 ms (P95,
  // `kennzahlen09`, Grenze 500 ms aus PERF02).
  it('stehen da, auch wenn die Prüfstände noch ausstehen', async () => {
    zeige([stand('1')], [], true)

    expect(
      await ueberblick().findByRole('button', { name: /^1 Set$/ }),
    ).toBeInTheDocument()
  })
})
