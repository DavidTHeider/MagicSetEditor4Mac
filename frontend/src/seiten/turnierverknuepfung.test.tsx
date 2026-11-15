// **Teilnehmer, Decklisten, Paarungen und Partien sind verknüpft** —
// DOM-04.AK-02 (#1699).
//
// Nach dem Artboard *Turniere*: der Kopf nennt System, Runde und Zahl der
// Teilnehmenden; die Decklisten stehen in einer eigenen Lasche und schlagen
// ihr Deck auf; eine an eine Partie gebundene Paarung öffnet sie am Tisch.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Anschlussraum, useAnschluss } from '../anschlussraum'
import { Spielen } from './Spielen'
import { Turniere } from './Turniere'
import { Turniermatch } from '../bausteine/Turniermatch'
import { aufgeschlagen, zumTurniergebiet } from './turnierwerkbank'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { mitDecks } from './spielenprobe'
import { rundenprobe, teilnehmerprobe, turnierprobe } from './turnierprobe'

import type { Anschluss } from '../anschluss'
import type { Seitenschluessel } from '../navigation'
import type { MatchAntwort } from '../apiturniere'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

// Unterwegs zeigt, was gerade reist — als Text, damit ein Lauf es lesen kann.
function Unterwegs() {
  const anschluss = useAnschluss()

  return (
    <output aria-label="unterwegs">
      {JSON.stringify(anschluss.offen ?? null)}
    </output>
  )
}

// Raum stellt einen Baum in den Anschlussraum und meldet jeden Wechsel.
function Raum({
  gewechselt,
  children,
}: {
  gewechselt: (seite: Seitenschluessel) => void
  children: React.ReactNode
}) {
  const [aktuell, setzeAktuell] = useState<Seitenschluessel>('turniere')

  return (
    <Sprachraum sprache="de">
      <Anschlussraum
        aktuell={aktuell}
        wechseln={(seite) => {
          gewechselt(seite)
          setzeAktuell(seite)
        }}
      >
        <Unterwegs />
        {children}
      </Anschlussraum>
    </Sprachraum>
  )
}

// unterwegs liest, was reist.
function unterwegs(): Anschluss | null {
  return JSON.parse(
    screen.getByRole('status', { name: 'unterwegs' }).textContent,
  ) as Anschluss | null
}

const zwei = [
  teilnehmerprobe({
    id: '3',
    name: 'Anna',
    deckID: '5',
    deckabzug: [
      {
        id: '1',
        bezeichnung: 'Blitz',
        printing: '',
        druckID: '',
        menge: 4,
        bereich: 'hauptdeck',
      },
    ],
  }),
  teilnehmerprobe({ id: '4', name: 'Bert', stelle: 2 }),
]

describe('der Kopf eines Turniers', () => {
  it('nennt System, Runde und Zahl der Teilnehmenden', async () => {
    const werkbank = await aufgeschlagen(
      turnierprobe({
        teilnehmer: zwei,
        runden: [rundenprobe()],
        empfohleneRunden: 2,
      }),
    )

    expect(
      within(werkbank).getByText(
        einsetzen(texte.kopfzeile, { runde: '1', von: '2', anzahl: '2' }),
      ),
    ).toBeInTheDocument()
  })

  // **Ohne Empfehlung kein „von"** — Appendix E der MTR (Fassung
  // 2026-02-27) empfiehlt für wenige Teilnehmende nichts, und „Runde 1 von 0" fand die Hüllenprobe.
  it('nennt ohne Empfehlung nur die Runde', async () => {
    const werkbank = await aufgeschlagen(
      turnierprobe({
        teilnehmer: zwei,
        runden: [rundenprobe()],
        empfohleneRunden: 0,
      }),
    )

    expect(
      within(werkbank).getByText(
        einsetzen(texte.kopfzeileOhneEmpfehlung, { runde: '1', anzahl: '2' }),
      ),
    ).toBeInTheDocument()
  })

  it('sagt vor der ersten Runde nur, wer dabei ist', async () => {
    const werkbank = await aufgeschlagen(turnierprobe({ teilnehmer: zwei }))

    expect(
      within(werkbank).getByText(
        einsetzen(texte.kopfzeileOhneRunde, { anzahl: '2' }),
      ),
    ).toBeInTheDocument()
  })
})

describe('die Decklisten', () => {
  it('stehen in einer eigenen Lasche und schlagen ihr Deck auf', async () => {
    const turnier = turnierprobe({ teilnehmer: zwei })
    belegeApi({
      Turniere: () => Promise.resolve({ turniere: [turnier], fehler: null }),
      Turnier: () => Promise.resolve({ turnier, fehler: null }),
      TurnierRangliste: () => Promise.resolve({ rangliste: [], fehler: null }),
    })
    const gewechselt = vi.fn()

    render(
      <Raum gewechselt={gewechselt}>
        <Turniere />
      </Raum>,
    )

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.waehlen}: ${turnier.name}`,
      }),
    )
    await zumTurniergebiet('decklisten')

    // **Der Abzug steht hier**, und wer keine Liste eingereicht hat, bekommt
    // keinen Knopf zum Aufschlagen.
    expect(
      screen.getByRole('button', { name: `${texte.deckabzug}: Anna` }),
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: `${texte.deckAufschlagen}: Bert` }),
    ).toBeNull()

    await userEvent.click(
      screen.getByRole('button', { name: `${texte.deckAufschlagen}: Anna` }),
    )

    expect(gewechselt).toHaveBeenCalledWith('decks')
    expect(unterwegs()).toEqual({
      art: 'deckansicht',
      name: 'Anna',
      deckID: '5',
    })
  })
})

describe('ohne eingereichte Liste', () => {
  it('sagt, wie eine hineinkommt', async () => {
    const werkbank = await aufgeschlagen(
      turnierprobe({ teilnehmer: [teilnehmerprobe({ name: 'Bert' })] }),
      {},
      'decklisten',
    )

    expect(within(werkbank).getByText(texte.decklistenLeer)).toBeInTheDocument()
  })
})

// gebundenerStand ist eine Paarung, an der eine Partie hängt.
function gebundenerStand(): MatchAntwort {
  return {
    gebunden: true,
    paarungID: '7',
    partieID: 'p1',
    fassung: '3',
    sitze: [
      { teilnehmerID: '10', sitz: 'anna', peer: '', deckabdruck: 'a' },
      { teilnehmerID: '11', sitz: 'bert', peer: '', deckabdruck: 'b' },
    ],
    meldungen: [],
    uebernommen: null,
    fehler: null,
  }
}

describe('eine gebundene Paarung', () => {
  it('öffnet ihre Partie am Tisch', async () => {
    belegeApi({ TurnierMatch: vi.fn().mockResolvedValue(gebundenerStand()) })
    const gewechselt = vi.fn()

    render(
      <Raum gewechselt={gewechselt}>
        <Turniermatch turnierID="1" paarungID="7" teilnehmer={['10', '11']} />
      </Raum>,
    )

    await userEvent.click(
      screen.getByRole('button', { name: texte.matchAnsehen }),
    )
    await userEvent.click(
      await screen.findByRole('button', { name: texte.partieOeffnen }),
    )

    expect(gewechselt).toHaveBeenCalledWith('spielen')
    expect(unterwegs()).toEqual({
      art: 'partieansicht',
      name: texte.match,
      partieID: 'p1',
      sitz: 'anna',
    })
  })

  it('wird auf der Spielseite aufgeschlagen', async () => {
    const partie = vi.fn().mockResolvedValue({
      fehler: {
        code: 'nichtGefunden' as const,
        schluessel: 'nichtGefunden',
        meldung: 'weg',
      },
    })
    mitDecks({ Partie: partie })

    function Abgeber() {
      const anschluss = useAnschluss()

      return (
        <button
          type="button"
          onClick={() => {
            anschluss.uebergib({
              art: 'partieansicht',
              name: 'Freitag',
              partieID: 'p1',
              sitz: 'anna',
            })
          }}
        >
          abgeben
        </button>
      )
    }

    render(
      <Raum gewechselt={vi.fn()}>
        <Abgeber />
        <Spielen wechseln={() => undefined} />
      </Raum>,
    )

    await userEvent.click(screen.getByRole('button', { name: 'abgeben' }))

    await waitFor(() => {
      expect(partie).toHaveBeenCalledWith('p1', 'anna')
    })
  })
})
