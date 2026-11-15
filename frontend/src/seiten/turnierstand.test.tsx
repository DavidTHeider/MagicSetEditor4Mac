// **Was die Turnierseite über einen gegebenen Stand sagt** — Werkstück 5 von
// Phase 4.7
// (#1294).
//
// **Getrennt von `turnierlauf.test.tsx`**, damit beide unter der
// Zeilengrenze aus §2 bleiben. Der Schnitt folgt der Frage: dort steht, was
// die Seite über mehrere Handgriffe **tut**, hier, was sie über einen Stand
// **sagt** — ein Bye, eine ungespielte Paarung, ein Einzelspielturnier — und
// wie sie sich verabschiedet, wenn etwas schiefgeht.
import { act, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Turniere } from './Turniere'
import { aufgeschlagen, zumTurniergebiet } from './turnierwerkbank'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { turnierhandlung } from '../test/turnierhandlung'
import {
  paarungsprobe,
  rundenprobe,
  teilnehmerprobe,
  turnierprobe,
} from './turnierprobe'
import type { TurnierDTO } from '../apiturniere'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

// gebietUndKnopf wechselt ins Gebiet »Runden« und gibt den Knopf.
//
// **Seit Werkstück I von Phase 4.10 hat die Turnierwerkbank drei Gebiete**
// (#1325).
async function gebietUndKnopf(name: string) {
  await zumTurniergebiet('runden')

  return screen.getByRole('button', { name })
}

describe('die Turnierseite über einen gegebenen Stand', () => {
  it('zeigt ein Bye als Ergebnis und nicht als fehlenden Gegner', async () => {
    const werkbank = await aufgeschlagen(
      turnierprobe({
        teilnehmer: [teilnehmerprobe()],
        runden: [
          rundenprobe({
            paarungen: [paarungsprobe({ b: '', ergebnis: null })],
          }),
        ],
      }),
      {},
      'runden',
    )

    expect(within(werkbank).getByText(`Ada — ${texte.bye}`)).toBeInTheDocument()
  })

  it('sagt es, wenn eine Paarung noch nicht gespielt ist', async () => {
    const werkbank = await aufgeschlagen(
      turnierprobe({
        teilnehmer: [
          teilnehmerprobe(),
          teilnehmerprobe({ id: '4', name: 'Bert', stelle: 2 }),
        ],
        runden: [rundenprobe()],
      }),
      {},
      'runden',
    )

    expect(within(werkbank).getByText(texte.offen)).toBeInTheDocument()
  })

  it('lässt bei einem Einzelspielturnier zwei Spalten weg', async () => {
    const werkbank = await aufgeschlagen(
      turnierprobe({ einzelspiel: true }),
      {},
      'runden',
    )

    // **Im Gebiet »Rangliste«** — Werkstück I von Phase 4.10.
    await zumTurniergebiet('rangliste')

    expect(
      within(werkbank).queryByRole('columnheader', { name: texte.spielquote }),
    ).toBeNull()

    expect(
      within(werkbank).getByRole('columnheader', {
        name: texte.gegnermatchquote,
      }),
    ).toBeInTheDocument()
  })

  it('bleibt nach dem Löschen nicht aufgeschlagen', async () => {
    let gelöscht = false

    const turnier = turnierprobe()

    belegeApi({
      Turniere: () =>
        Promise.resolve({ turniere: gelöscht ? [] : [turnier], fehler: null }),
      Turnier: () => Promise.resolve({ turnier, fehler: null }),
      TurnierRangliste: () => Promise.resolve({ rangliste: [], fehler: null }),
      TurnierLoeschen: () => {
        gelöscht = true

        return Promise.resolve({ fehler: null })
      },
    })

    render(
      <Sprachraum sprache="de">
        <Turniere />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.waehlen}: Küchentisch`,
      }),
    )

    expect(
      await screen.findByRole('region', { name: 'Küchentisch' }),
    ).toBeInTheDocument()

    await turnierhandlung('Küchentisch', texte.loeschen)
    await userEvent.click(
      screen.getByRole('button', { name: texte.loeschenJa }),
    )

    await waitFor(() => {
      expect(screen.queryByRole('region', { name: 'Küchentisch' })).toBeNull()
    })
  })

  // **Auch ein Lesen meldet seinen Fehler.**
  //
  // Der Paarungsvorschlag geht seit Befund 1 des Nachlaufs über Phase 4.7
  // **nicht** mehr durch den Schreibweg — er lädt nichts nach. Was er
  // behalten hat, ist die Meldung: ein Vorschlag, der scheitert, sagt es,
  // statt still nichts zu tun.
  it('meldet einen Fehler beim Vorschlagen', async () => {
    const werkbank = await aufgeschlagen(
      turnierprobe({ teilnehmer: [teilnehmerprobe()] }),
      {
        TurnierPaarungsvorschlag: () =>
          Promise.reject(new Error('kein Vorschlag')),
      },
      'runden',
    )

    await userEvent.click(
      within(werkbank).getByRole('button', { name: texte.vorschlag }),
    )

    expect(await screen.findByText(/kein Vorschlag/)).toBeInTheDocument()
  })

  it('meldet einen Fehler beim Schreiben', async () => {
    await aufgeschlagen(
      turnierprobe(),
      {
        TurnierRundeLegen: () =>
          Promise.resolve({
            runde: null,
            fehler: {
              code: 'validierung',
              schluessel: 'fehler.validierung',
              meldung: 'Die laufende Runde hat offene Ergebnisse.',
            },
          }),
      },
      'runden',
    )

    await userEvent.click(await gebietUndKnopf(texte.rundeLegen))

    await waitFor(() => {
      expect(
        screen.getByText('Die laufende Runde hat offene Ergebnisse.'),
      ).toBeInTheDocument()
    })
  })

  it('verwirft die Antwort eines überholten Aufschlagens', async () => {
    const erste = turnierprobe({ id: '7', name: 'Küchentisch' })
    const zweite = turnierprobe({ id: '8', name: 'Anderswo' })

    let loeseAuf: (() => void) | undefined
    let antwort: Promise<{ turnier: TurnierDTO; fehler: null }> | undefined

    const Turnier = vi.fn((id: string) => {
      if (id === '8') {
        return Promise.resolve({ turnier: zweite, fehler: null })
      }

      antwort = new Promise((fertig) => {
        loeseAuf = () => {
          fertig({ turnier: erste, fehler: null })
        }
      })

      return antwort
    })

    belegeApi({
      Turniere: () =>
        Promise.resolve({ turniere: [erste, zweite], fehler: null }),
      Turnier,
      TurnierRangliste: () => Promise.resolve({ rangliste: [], fehler: null }),
    })

    render(
      <Sprachraum sprache="de">
        <Turniere />
      </Sprachraum>,
    )

    // **Das erste aufschlagen — die Antwort hängt.**
    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.waehlen}: Küchentisch`,
      }),
    )

    // **Und währenddessen das zweite.**
    await userEvent.click(
      screen.getByRole('button', { name: `${texte.waehlen}: Anderswo` }),
    )

    expect(
      await screen.findByRole('region', { name: 'Anderswo' }),
    ).toBeInTheDocument()

    // **Jetzt kommt die Antwort von vorhin** — und darf das zweite nicht
    // verdrängen (§7, Regel 2).
    loeseAuf?.()

    await act(async () => {
      await antwort
    })

    expect(screen.queryByRole('region', { name: 'Küchentisch' })).toBeNull()
    expect(screen.getByRole('region', { name: 'Anderswo' })).toBeInTheDocument()
  })
})

// **Eine leere Rangliste** — Werkstück I von Phase 4.10
// (#1325).
// Sie stand vorher immer neben allem anderen; als eigenes Gebiet ist sie das
// Erste, was jemand sieht, der auf die Lasche geht — und dann darf dort nicht
// nichts stehen (§6).
describe('die leere Rangliste', () => {
  it('nennt sie beim Namen, auch wenn niemand drinsteht', async () => {
    const werkbank = await aufgeschlagen(
      turnierprobe(),
      {
        TurnierRangliste: () =>
          Promise.resolve({ rangliste: [], fehler: null }),
      },
      'rangliste',
    )

    expect(
      within(werkbank).getByRole('heading', { name: texte.rangliste }),
    ).toBeInTheDocument()
  })
})
