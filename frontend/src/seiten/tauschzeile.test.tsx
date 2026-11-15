// **Wie eine Zeile des Abgleichs dasteht** — Werkstück T4 von Phase 5.6
// (#1469).
//
// # Was hier gemessen wird
//
// **Der ganze Text einer Zeile und nicht ein Wort darin.** „Steht Shivan
// Dragon da" trifft auch dann zu, wenn es zweimal dasteht — und genau das
// stand am 2026-09-18 am Fenster.
import { screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { eineListe } from './peerkontaktprobe'
import {
  bestaetigt,
  einAbgleich,
  einKandidat,
  einStand,
  setzeEin,
  zeigeHandel,
} from './tauschprobe'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.handeln

describe('Wie eine Zeile des Abgleichs dasteht', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  // **Eine Zeile nennt die Karte einmal** — gefunden bei der Sichtprüfung am
  // 2026-09-18: bei einem genauen Wunsch sind Angebot und Wunsch **dieselbe**
  // Zeile, und die Anzeige schrieb sie zweimal hin:
  //
  //     1 × Shivan Dragon · LEA #232 · Shivan Dragon · LEA #232 · normal · genau
  //
  // **Ein DOM-Lauf fängt das nicht**: „steht Shivan Dragon da" trifft auch
  // zweimal zu. Deshalb steht hier der **ganze** Text der Zeile.
  it('nennt die Karte einmal, wenn Angebot und Wunsch dieselbe sind', async () => {
    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Tauschabgleich: () =>
        Promise.resolve(
          einAbgleich({
            gebe: [
              einKandidat({
                wunsch: 'Shivan Dragon · LEA #232',
                angebot: 'Shivan Dragon · LEA #232',
              }),
            ],
            bekomme: [],
          }),
        ),
    })

    await setzeEin(await zeigeHandel(), einStand)

    const gebe = await screen.findByRole('list', { name: eigene.tauschGebe })
    const zeile = within(gebe).getAllByRole('listitem')[0]

    expect(zeile).toHaveTextContent('1 × Shivan Dragon · LEA #232')

    // **Und nichts darüber hinaus**: kein zweites Mal, kein »normal«, kein
    // »genau« — was die Vorgabe ist, sagt nichts.
    expect(zeile?.textContent).toBe('1 × Shivan Dragon · LEA #232')
  })

  // **Was von der Vorgabe abweicht, steht dran** — der Rang, die Art und der
  // Wunsch, wenn er ein anderer ist als das Angebot.
  it('nennt Rang, Art und den anderen Wunsch, wo sie etwas sagen', async () => {
    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Tauschabgleich: () =>
        Promise.resolve(
          einAbgleich({
            gebe: [
              einKandidat({
                angebot: 'Shivan Dragon · M10 #234',
                wunsch: 'Shivan Dragon',
                rang: 'hoch',
                art: 'jederdruck',
                fehlend: 2,
              }),
            ],
            bekomme: [],
          }),
        ),
    })

    await setzeEin(await zeigeHandel(), einStand)

    const gebe = await screen.findByRole('list', { name: eigene.tauschGebe })
    const zeile = within(gebe).getAllByRole('listitem')[0]

    expect(zeile?.textContent).toBe(
      '1 × Shivan Dragon · M10 #234 · Shivan Dragon · hoch · jederdruck · −2',
    )
  })
})
