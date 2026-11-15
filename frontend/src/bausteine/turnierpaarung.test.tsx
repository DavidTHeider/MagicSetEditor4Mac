// **Das Eintragen klappt nach dem Schreiben zu** — UXR-12
// (#1920).
//
// # Warum das hier steht und nicht auf der Seite
//
// Auf der Seite baut das Nachladen nach einem Schreiben die Werkbank
// **meistens** neu auf, und das Eintragen ist dann ohnehin zu. Meistens:
// `waehle` setzt das offene Turnier erst leer und dann neu, und schafft
// Reacts Scheduler es nicht zwischen beide, zeichnet er beides in einem Zug
// — dann steht dieselbe Zeile da, mit offenem Eintragen. Wer acht Ergebnisse
// einträgt, hätte so acht offene Kästen. Geprüft wird deshalb die Zeile
// selbst, an einem Stand, der nach dem Schreiben **nichts** neu aufbaut.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Turnierpaarung } from './Turnierpaarung'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { paarungsprobe } from '../seiten/turnierprobe'
import type { Api } from '../api'
import type { Turnierstand } from '../haken/useTurniere'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

// stillerStand führt jede Arbeit aus und lädt nichts nach — ein Nein
// verschluckt er, wie `useTurniere` es über die Seite meldet.
function stillerStand(): Turnierstand {
  const ausfuehren = async (_marke: string, arbeit: () => Promise<void>) => {
    try {
      await arbeit()
    } catch {
      // gemeldet würde es über der Seite
    }
  }

  return {
    vorgang: { laeuft: () => false },
    fuehre: ausfuehren,
    lies: ausfuehren,
  } as unknown as Turnierstand
}

function zeige(ergebnisSetzen: Api['TurnierErgebnisSetzen']) {
  belegeApi({ TurnierErgebnisSetzen: ergebnisSetzen })

  render(
    <Sprachraum sprache="de">
      <Turnierpaarung
        stand={stillerStand()}
        turnierID="7"
        fassung="1"
        paarung={paarungsprobe()}
        namen={
          new Map([
            ['3', 'Ada'],
            ['4', 'Bert'],
          ])
        }
        hoechstens={0}
        spaeterGelegt={false}
      />
    </Sprachraum>,
  )
}

const wer = `Ada ${texte.gegen} Bert`

it('klappt nach dem Eintragen zu', async () => {
  const setzen = vi.fn<Api['TurnierErgebnisSetzen']>(() =>
    Promise.resolve({ fehler: null }),
  )
  zeige(setzen)

  const ergebnis = screen.getByRole('button', {
    name: `${texte.ergebnis}: ${wer}`,
  })
  await userEvent.click(ergebnis)
  await userEvent.type(
    screen.getByLabelText(`${texte.spieleA} — ${wer}`),
    '{Backspace}2',
  )
  await userEvent.click(
    screen.getByRole('button', { name: `${texte.eintragen}: ${wer}` }),
  )

  await waitFor(() => {
    expect(setzen).toHaveBeenCalled()
  })
  await waitFor(() => {
    expect(ergebnis).toHaveAttribute('aria-expanded', 'false')
  })
})

// **Ein Nein lässt es offen** — dort steht, was zu tun ist.
it('bleibt offen, wenn das Eintragen scheitert', async () => {
  const setzen = vi.fn<Api['TurnierErgebnisSetzen']>(() =>
    Promise.reject(new Error('Der Bestand schweigt.')),
  )
  zeige(setzen)

  const ergebnis = screen.getByRole('button', {
    name: `${texte.ergebnis}: ${wer}`,
  })
  await userEvent.click(ergebnis)
  await userEvent.type(
    screen.getByLabelText(`${texte.spieleA} — ${wer}`),
    '{Backspace}2',
  )
  await userEvent.click(
    screen.getByRole('button', { name: `${texte.eintragen}: ${wer}` }),
  )

  await waitFor(() => {
    expect(setzen).toHaveBeenCalled()
  })
  expect(ergebnis).toHaveAttribute('aria-expanded', 'true')
})
