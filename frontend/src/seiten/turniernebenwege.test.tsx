// **Die Nebenwege der Turnierseite** — Werkstück 5 von Phase 4.7
// (#1294).
//
// **Eine eigene Datei**, weil `turnierlauf.test.tsx` mit ihnen über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort steht der
// **Ablauf**, hier stehen die Fälle, die nur einmal vorkommen — eine Notiz,
// ein Lesefehler, eine überholte Ablehnung.
import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Turniere } from './Turniere'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { turnierhandlung } from '../test/turnierhandlung'
import { alleTurniere } from './turnierwerkbank'
import { turnierprobe } from './turnierprobe'
import type { TurnierAntwort } from '../apiturniere'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

describe('zwei Fälle, die nur einmal vorkommen', () => {
  it('zeigt die Notiz eines Turniers, wenn es eine hat', async () => {
    belegeApi({
      Turniere: () =>
        Promise.resolve({
          turniere: [turnierprobe({ notiz: 'Freitagabend' })],
          fehler: null,
        }),
      Turnier: () => Promise.resolve({ turnier: turnierprobe(), fehler: null }),
      TurnierRangliste: () => Promise.resolve({ rangliste: [], fehler: null }),
    })

    render(
      <Sprachraum sprache="de">
        <Turniere />
      </Sprachraum>,
    )

    expect(await screen.findByText('Freitagabend')).toBeInTheDocument()
  })

  // **Der Fehlerzweig des Aufschlagens.**
  //
  // Er ist nicht derselbe wie der beim Schreiben: hier scheitert das
  // **Lesen**, und die Meldung darf nicht am Feld hängen, sondern gehört an
  // die Seite.
  it('meldet einen Fehler beim Aufschlagen', async () => {
    belegeApi({
      Turniere: () =>
        Promise.resolve({ turniere: [turnierprobe()], fehler: null }),
      Turnier: () =>
        Promise.resolve({
          turnier: null,
          fehler: {
            code: 'nichtGefunden',
            schluessel: 'fehler.nichtGefunden',
            meldung: 'Dieses Turnier gibt es nicht mehr.',
          },
        }),
      TurnierRangliste: () => Promise.resolve({ rangliste: [], fehler: null }),
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

    expect(
      await screen.findByText('Dieses Turnier gibt es nicht mehr.'),
    ).toBeInTheDocument()
  })
})

// **Auch der Fehler eines überholten Aufschlagens wird verworfen.**
//
// **Nicht dasselbe wie der Erfolgsfall darüber**: dort wird ein Turnier
// gesetzt, hier eine Meldung gezeigt. Eine Meldung zu einem Turnier, das
// niemand mehr offen hat, sieht aus wie ein Fehler des aktuellen (§7,
// Regel 2).
describe('ein überholter Fehler', () => {
  it('erreicht die Seite nicht', async () => {
    const erste = turnierprobe({ id: '7', name: 'Küchentisch' })
    const zweite = turnierprobe({ id: '8', name: 'Anderswo' })

    let lehneAb: (() => void) | undefined
    let antwort: Promise<TurnierAntwort> | undefined

    const Turnier = vi.fn((id: string): Promise<TurnierAntwort> => {
      if (id === '8') {
        return Promise.resolve({ turnier: zweite, fehler: null })
      }

      antwort = new Promise<TurnierAntwort>((_, ablehnen) => {
        lehneAb = () => {
          ablehnen(new Error('Dieses Turnier gibt es nicht mehr.'))
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

    await alleTurniere()
    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.waehlen}: Küchentisch`,
      }),
    )

    await alleTurniere()
    await userEvent.click(
      screen.getByRole('button', { name: `${texte.waehlen}: Anderswo` }),
    )

    expect(
      await screen.findByRole('region', { name: 'Anderswo' }),
    ).toBeInTheDocument()

    lehneAb?.()

    await act(async () => {
      await antwort?.catch(() => undefined)
    })

    expect(screen.queryByText('Dieses Turnier gibt es nicht mehr.')).toBeNull()
  })
})

// **Ein verspäteter Schreibfehler von Turnier A erreicht Turnier B nicht.**
//
// # Der Befund
//
// Der Auswahlwechsel wurde gezählt, ein gestarteter **Schreibvorgang** hielt
// seine Ausgangsgeneration aber nicht fest. Sein später Fehler landete immer
// im gemeinsamen Seitenzustand — also über dem Turnier, das inzwischen
// aufgeschlagen ist.
//
// **Die Läufe deckten verspätete Lese-Antworten ab, nicht verspätete
// Schreib-Antworten.** Befund 9 des Zweitblicks über Phase 4.7.
describe('ein verspäteter Schreibfehler', () => {
  it('wird nicht dem falschen Turnier zugeschrieben', async () => {
    const erste = turnierprobe({ id: '7', name: 'Küchentisch' })
    const zweite = turnierprobe({ id: '8', name: 'Anderswo' })

    let lehneAb: (() => void) | undefined
    let antwort: Promise<{ fehler: null }> | undefined

    const TurnierLoeschen = vi.fn((): Promise<{ fehler: null }> => {
      antwort = new Promise((_, ablehnen) => {
        lehneAb = () => {
          ablehnen(new Error('Dieses Turnier ließ sich nicht löschen.'))
        }
      })

      return antwort
    })

    belegeApi({
      Turniere: () =>
        Promise.resolve({ turniere: [erste, zweite], fehler: null }),
      Turnier: (id: string) =>
        Promise.resolve({ turnier: id === '8' ? zweite : erste, fehler: null }),
      TurnierRangliste: () => Promise.resolve({ rangliste: [], fehler: null }),
      TurnierLoeschen,
    })

    render(
      <Sprachraum sprache="de">
        <Turniere />
      </Sprachraum>,
    )

    // **Das Löschen von A anstoßen — die Antwort hängt.** Gelöscht wird
    // am offenen Turnier (seit UXR-12, #1920), also wird A erst
    // aufgeschlagen.
    await alleTurniere()
    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.waehlen}: Küchentisch`,
      }),
    )
    await turnierhandlung('Küchentisch', texte.loeschen)
    await userEvent.click(
      screen.getByRole('button', { name: texte.loeschenJa }),
    )

    // **Und währenddessen B aufschlagen.**
    await alleTurniere()
    await userEvent.click(
      screen.getByRole('button', { name: `${texte.waehlen}: Anderswo` }),
    )

    expect(
      await screen.findByRole('region', { name: 'Anderswo' }),
    ).toBeInTheDocument()

    // **Jetzt scheitert das Löschen von A.**
    lehneAb?.()

    await act(async () => {
      await antwort?.catch(() => undefined)
    })

    expect(
      screen.queryByText('Dieses Turnier ließ sich nicht löschen.'),
    ).toBeNull()
  })
})
