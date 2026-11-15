// **Ein Vorschlag, den ein Wechsel überholt hat, wird verworfen** (§7,
// Regel 2) — die zweite Hälfte von Befund 1 des Nachlaufs über Phase 4.7.
//
// Der Vorschlag läuft seit der Korrektur nicht mehr durch den Schreibweg,
// sondern über `lies`. Was er von dort mitbekommen hat, ist der Wächter über
// den Stand: wer das Turnier wechselt, während ein Vorschlag unterwegs ist,
// bekommt weder dessen Ergebnis noch dessen Fehler zu sehen.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Turniere } from './Turniere'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { zumTurniergebiet, alleTurniere } from './turnierwerkbank'
import { kataloge } from '../sprache/kataloge'
import { teilnehmerprobe, turnierprobe } from './turnierprobe'
import type { PaarungenAntwort } from '../apiturniere'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

describe('ein überholter Paarungsvorschlag', () => {
  it('landet weder als Liste noch als Fehler', async () => {
    const erste = turnierprobe({
      id: '7',
      name: 'Küchentisch',
      teilnehmer: [teilnehmerprobe()],
    })

    const zweite = turnierprobe({ id: '8', name: 'Anderswo' })

    let scheitern: (() => void) | undefined

    const TurnierPaarungsvorschlag = vi.fn(
      () =>
        new Promise<PaarungenAntwort>((_, misslingt) => {
          scheitern = () => {
            misslingt(new Error('kein Vorschlag'))
          }
        }),
    )

    belegeApi({
      Turniere: () =>
        Promise.resolve({ turniere: [erste, zweite], fehler: null }),
      Turnier: (id: string) =>
        Promise.resolve({ turnier: id === '8' ? zweite : erste, fehler: null }),
      TurnierRangliste: () => Promise.resolve({ rangliste: [], fehler: null }),
      TurnierPaarungsvorschlag,
    })

    render(
      <Sprachraum sprache="de">
        <Turniere />
      </Sprachraum>,
    )

    await alleTurniere()
    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.waehlen}: Küchentisch`,
      }),
    )

    // **Der Vorschlag liegt im Gebiet »Runden«** — Werkstück I von Phase
    // 4.10 (#1325).
    await zumTurniergebiet('runden')

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.vorschlag,
      }),
    )

    await waitFor(() => {
      expect(TurnierPaarungsvorschlag).toHaveBeenCalled()
    })

    // **Jetzt woanders hin** — und erst dann scheitert der Vorschlag.
    await alleTurniere()
    await userEvent.click(
      screen.getByRole('button', { name: `${texte.waehlen}: Anderswo` }),
    )

    // **Nach dem Wechsel steht das erste Gebiet** (§7, Regel 1) — wer die
    // Runden will, geht hin.
    await zumTurniergebiet('runden')

    scheitern?.()

    await waitFor(() => {
      expect(
        screen.getByRole('region', { name: 'Anderswo' }),
      ).toBeInTheDocument()
    })

    expect(screen.queryByText(/kein Vorschlag/)).toBeNull()
    expect(
      screen.queryByRole('list', { name: texte.vorschlagTitel }),
    ).toBeNull()
  })

  // **Und wenn er glückt, auch nicht.**
  //
  // Beide Wege brauchen den Wächter, und beide brauchen ihren Test: einer,
  // der nur den Fehlerweg prüft, ließe den Erfolgsweg offen — und dann
  // erschiene eine Paarung des einen Turniers unter dem Namen des anderen.
  it('landet auch dann nicht, wenn er glückt', async () => {
    const erste = turnierprobe({
      id: '7',
      name: 'Küchentisch',
      teilnehmer: [teilnehmerprobe()],
    })

    const zweite = turnierprobe({ id: '8', name: 'Anderswo' })

    let gelingen: (() => void) | undefined

    const TurnierPaarungsvorschlag = vi.fn(
      () =>
        new Promise<PaarungenAntwort>((fertig) => {
          gelingen = () => {
            fertig({
              paarungen: [{ id: '', a: '3', b: '', ergebnis: null }],
              fehler: null,
            })
          }
        }),
    )

    belegeApi({
      Turniere: () =>
        Promise.resolve({ turniere: [erste, zweite], fehler: null }),
      Turnier: (id: string) =>
        Promise.resolve({ turnier: id === '8' ? zweite : erste, fehler: null }),
      TurnierRangliste: () => Promise.resolve({ rangliste: [], fehler: null }),
      TurnierPaarungsvorschlag,
    })

    render(
      <Sprachraum sprache="de">
        <Turniere />
      </Sprachraum>,
    )

    await alleTurniere()
    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.waehlen}: Küchentisch`,
      }),
    )

    // **Der Vorschlag liegt im Gebiet »Runden«** (Werkstück I von Phase 4.10).
    await zumTurniergebiet('runden')

    await userEvent.click(
      await screen.findByRole('button', { name: texte.vorschlag }),
    )

    await waitFor(() => {
      expect(TurnierPaarungsvorschlag).toHaveBeenCalled()
    })

    await alleTurniere()
    await userEvent.click(
      screen.getByRole('button', { name: `${texte.waehlen}: Anderswo` }),
    )

    // **Nach dem Wechsel steht das erste Gebiet** (§7, Regel 1) — wer die
    // Runden will, geht hin.
    await zumTurniergebiet('runden')

    gelingen?.()

    await waitFor(() => {
      expect(
        screen.getByRole('region', { name: 'Anderswo' }),
      ).toBeInTheDocument()
    })

    expect(
      screen.queryByRole('list', { name: texte.vorschlagTitel }),
    ).toBeNull()
  })
})
