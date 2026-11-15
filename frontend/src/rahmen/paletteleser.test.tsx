// **Ein Leser bekommt in der Palette nur, was er tun darf** — G07
// (#2254).
//
// Ein Leser arbeitet über eine Verbindung, die nur liest
// (`internal/app/nurlesen.go`); jedes Schreiben scheitert mit
// `ErrNurLesend`. Die Palette bot ihm trotzdem „Karte erstellen", „Deck
// spielen" (legt eine Partie an), „Karten tauschen mit" (legt einen Tausch
// an) und jeden schreibenden Befehl — die Bindung wies erst danach ab.
// Dieselbe Klasse wie G10 bei den Menüs (#2288).
//
// **Was bleibt**: Navigieren, Finden und die Sicherung, die nur liest.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Anschlussraum } from '../anschlussraum'
import type { BefehlDTO } from '../apibefehle'
import { Schreibrechtraum } from '../haken/schreibrecht'
import { deckprobe } from '../seiten/deckprobe'
import { eineListe, einKontakt } from '../seiten/peerkontaktprobe'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Palette } from './Palette'

const texte = kataloge.de.allgemein.palette

afterEach(raeumeApi)

function befehl(name: string, schreibend: boolean): BefehlDTO {
  return {
    name,
    faehigkeit: '',
    cli: '',
    rest: '',
    erklaerung: '',
    schreibend,
    idempotent: !schreibend,
    gefaehrlich: false,
    schema: {},
  }
}

function zeige(nurLesend: boolean) {
  belegeApi({
    Befehlskatalog: () =>
      Promise.resolve({
        befehle: [befehl('set.lesen', false), befehl('set.anlegen', true)],
        fehler: null,
      }),
    Sets: () =>
      Promise.resolve({
        sets: [
          {
            id: '7',
            name: 'Werkstatt',
            spiel: 'magic',
            stil: 'm15',
            herkunft: '',
            karten: 3,
            herkunftFehlt: false,
          },
        ],
        fehler: null,
      }),
    Decks: () =>
      Promise.resolve({
        decks: [deckprobe({ id: '3', name: 'Aschenrat' })],
        fehler: null,
      }),
    Peerkontakte: () =>
      Promise.resolve(
        eineListe([
          einKontakt({
            kennung: 'peer-1',
            eigenerName: 'Turfer',
            stand: 'bestaetigt',
          }),
        ]),
      ),
    Suchen: () =>
      Promise.resolve({ treffer: [], gekuerzt: false, fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Schreibrechtraum nurLesend={nurLesend}>
        <Anschlussraum aktuell="start" wechseln={vi.fn()}>
          <Palette
            schliessen={vi.fn()}
            waehlen={vi.fn()}
            waehleBefehl={vi.fn()}
            wege={{ setID: '7', waehleSet: vi.fn(), oeffneKarte: vi.fn() }}
          />
        </Anschlussraum>
      </Schreibrechtraum>
    </Sprachraum>,
  )
}

// tunzeilenSind wartet, bis unter »Tun« genau `erwartet` steht.
//
// **Auf die Wirkung und nicht auf den ersten Knopf** (§3): hier stand das
// Warten auf »Sicherung«, und die steht von Anfang an da — Sets, Decks und
// Tauschpartner kommen später. Unter der Last der Vorprüfung las der Lauf
// zwei Zeilen statt vier und war rot, allein grün.
async function tunzeilenSind(erwartet: string[]) {
  const tun = await screen.findByRole('list', { name: texte.tun })

  await waitFor(() => {
    expect(
      within(tun)
        .getAllByRole('button')
        .map((knopf) => knopf.firstChild?.textContent ?? undefined),
    ).toEqual(erwartet)
  })
}

describe('Die Palette eines Lesers', () => {
  it('bietet unter »Tun« nur, was nicht schreibt', async () => {
    zeige(true)

    await tunzeilenSind([texte.setWaehlen, texte.sicherungErstellen])
  })

  it('bietet einem, der schreiben darf, weiter alles an', async () => {
    zeige(false)

    await tunzeilenSind([
      texte.karteErstellen,
      'Schnell spielen mit Aschenrat',
      'Karten tauschen mit Turfer',
      texte.sicherungErstellen,
    ])
  })

  it('findet für einen Leser nur Befehle, die lesen', async () => {
    zeige(true)

    await userEvent.type(screen.getByRole('textbox'), 'set.')

    const liste = await screen.findByRole('list', { name: texte.befehle })

    // **Der lesende steht da** — der Zuhörer vor der Verneinung (§3).
    await within(liste).findByText('set.lesen')
    expect(within(liste).queryByText('set.anlegen')).toBeNull()
  })
})
