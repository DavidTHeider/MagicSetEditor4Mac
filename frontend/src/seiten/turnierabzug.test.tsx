// **Der eingereichte Deckabzug ist zu sehen** — Befund 2 des Nachlaufs über
// Phase 4.7.
//
// Er wurde beim Anmelden gebaut, im Bestand abgelegt und im DTO ausgeliefert
// — und das Fenster zeigte davon eine nackte Zahl neben dem Namen. Wozu es
// den Abzug gibt (Vermerk 0062, Entscheidung 3: Text und keine Verweise,
// damit er jedes Löschen überlebt), war nirgends zu sehen.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import { Turnierdecklisten } from '../bausteine/Turnierdecklisten'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { teilnehmerprobe, turnierprobe } from './turnierprobe'

const texte = kataloge.de.seiten.turniere

// Werkbank zeigt die Lasche *Decklisten* — seit DOM-04.AK-02 (#1699) steht
// der Abzug dort und nicht mehr in der Zeile des Teilnehmers.
function Werkbank({ turnier }: { turnier: ReturnType<typeof turnierprobe> }) {
  return <Turnierdecklisten turnier={turnier} />
}

describe('der Deckabzug in den Decklisten', () => {
  it('zeigt die eingereichte Liste, wenn man sie aufklappt', async () => {
    belegeApi({})

    const turnier = turnierprobe({
      teilnehmer: [
        teilnehmerprobe({
          name: 'Ada',
          deckabzug: [
            {
              id: '1',
              bezeichnung: 'Lightning Bolt',
              printing: '',
              druckID: '',
              menge: 4,
              bereich: 'hauptdeck',
            },
            {
              id: '2',
              bezeichnung: 'Counterspell',
              printing: '',
              druckID: '',
              menge: 2,
              bereich: 'sideboard',
            },
          ],
        }),
      ],
    })

    render(
      <Sprachraum sprache="de">
        <Werkbank turnier={turnier} />
      </Sprachraum>,
    )

    // **Zugeklappt steht die Zahl da** — und sie sagt jetzt, was sie zählt.
    const aufklappen = screen.getByRole('button', {
      name: `${texte.deckabzug}: Ada`,
    })

    expect(aufklappen).toHaveTextContent('6')

    await userEvent.click(aufklappen)

    const liste = screen.getByRole('list', { name: `${texte.deckabzug}: Ada` })

    expect(within(liste).getByText(/Lightning Bolt/)).toBeInTheDocument()
    expect(within(liste).getByText(/Counterspell/)).toBeInTheDocument()

    raeumeApi()
  })

  // **Ein echter Druck trägt keine Bezeichnung** — wie er heißt, weiß nur
  // Scryfall (siehe `Turnierleitung.abziehen`). Dann steht das Wort „Karte"
  // und die Kennung des Drucks daneben, statt einer leeren Zeile.
  it('nennt einen unbenannten Druck bei seiner Kennung', async () => {
    belegeApi({})

    render(
      <Sprachraum sprache="de">
        <Werkbank
          turnier={turnierprobe({
            teilnehmer: [
              teilnehmerprobe({
                name: 'Cem',
                deckabzug: [
                  {
                    id: '3',
                    bezeichnung: '',
                    printing: 'lea-1',
                    druckID: '',
                    menge: 1,
                    bereich: 'hauptdeck',
                  },
                ],
              }),
            ],
          })}
        />
      </Sprachraum>,
    )

    await userEvent.click(
      screen.getByRole('button', { name: `${texte.deckabzug}: Cem` }),
    )

    const liste = screen.getByRole('list', { name: `${texte.deckabzug}: Cem` })

    expect(
      within(liste).getByText(`1× ${texte.deckabzugUnbenannt} (lea-1)`),
    ).toBeInTheDocument()

    raeumeApi()
  })

  it('bietet nichts an, wo kein Abzug ist', () => {
    belegeApi({})

    render(
      <Sprachraum sprache="de">
        <Werkbank
          turnier={turnierprobe({
            // **Mit Deck, aber ohne Zeilen** — dann steht er in den
            // Decklisten, und einen Abzug zum Aufklappen gibt es nicht.
            teilnehmer: [
              teilnehmerprobe({ name: 'Bert', deckID: '9', deckabzug: [] }),
            ],
          })}
        />
      </Sprachraum>,
    )

    expect(
      screen.queryByRole('button', { name: `${texte.deckabzug}: Bert` }),
    ).toBeNull()

    raeumeApi()
  })
})
