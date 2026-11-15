// **Was die Turnierseite über mehrere Handgriffe hinweg tut** — Werkstück 5
// von Phase 4.7
// (#1294).
//
// **Eine eigene Datei**, damit `turniere.test.tsx` unter der Zeilengrenze aus
// §2 bleibt. Der Schnitt folgt der Frage: dort steht, wie die Seite
// **aussieht**, hier, was sie über die Zeit **tut** — und in
// `turnierstand.test.tsx`, was sie über einen gegebenen Stand **sagt**.
import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import {
  aufgeschlagen,
  ergebnisOeffnen,
  zumTurniergebiet,
} from './turnierwerkbank'
import { deckprobe } from './deckprobe'
import { effekteNachholen } from '../test/effekte'
import { raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import {
  ohneFolgen,
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

// gebietUndKnopfRunden wechselt ins Gebiet »Runden« und gibt den Knopf.
async function gebietUndKnopfRunden(name: string) {
  await zumTurniergebiet('runden')

  return screen.getByRole('button', { name })
}

describe('die Turnierseite über mehrere Handgriffe', () => {
  describe('die Turnierseite über mehrere Handgriffe', () => {
    it('meldet einen Teilnehmer an und wieder ab', async () => {
      const TurnierAnmelden = vi.fn(() =>
        Promise.resolve({ teilnehmer: teilnehmerprobe(), fehler: null }),
      )
      const TurnierAussteigen = vi.fn(() => Promise.resolve({ fehler: null }))
      const Decksbeleg = vi.fn(() =>
        Promise.resolve({
          decks: [deckprobe({ id: '5', name: 'Weiß-Blau' })],
          fehler: null,
        }),
      )

      const werkbank = await aufgeschlagen(
        turnierprobe({
          teilnehmer: [
            teilnehmerprobe({
              deckabzug: [
                {
                  id: '11',
                  bezeichnung: 'Lightning Bolt',
                  printing: '',
                  druckID: '',
                  menge: 4,
                  bereich: 'hauptdeck',
                },
              ],
            }),
          ],
        }),
        {
          TurnierAnmelden,
          TurnierAussteigen,
          Decks: Decksbeleg,
        },
      )

      await userEvent.type(within(werkbank).getByLabelText(texte.name), 'Bert')

      // **Das Deck wird gewählt, nicht getippt** — Werkstück I von Phase 4.10
      // (#1325).
      // Gewartet wird auf den Eintrag, nicht auf das Feld (§3).
      await screen.findByRole('option', { name: 'Weiß-Blau' })

      await userEvent.selectOptions(
        within(werkbank).getByLabelText(texte.deck),
        '5',
      )
      await userEvent.click(
        within(werkbank).getByRole('button', { name: texte.anmelden }),
      )

      await waitFor(() => {
        expect(TurnierAnmelden).toHaveBeenCalledWith('7', 'Bert', '5')
      })

      // **Gewartet wird auf die Wirkung, nicht auf den Ruf** (§3). Nach der
      // Anmeldung lädt die Seite das Turnier nach, und einen Augenblick lang
      // steht es gar nicht da — unter Last lange genug für einen Griff ins
      // Leere: rot in `make ci` am 2026-09-23. Und solange der Vorgang läuft,
      // ist der Knopf angehalten; ein Klick darauf ginge verloren.
      await waitFor(() => {
        expect(
          screen.getByRole('button', { name: `${texte.aussteigen}: Ada` }),
        ).toBeEnabled()
      })

      await userEvent.click(
        screen.getByRole('button', { name: `${texte.aussteigen}: Ada` }),
      )

      await waitFor(() => {
        expect(TurnierAussteigen).toHaveBeenCalledWith('7', '3', true)
      })
    })

    it('meldet einen Ausgestiegenen wieder an', async () => {
      const TurnierAussteigen = vi.fn(() => Promise.resolve({ fehler: null }))

      await aufgeschlagen(
        turnierprobe({
          teilnehmer: [teilnehmerprobe({ ausgestiegen: true })],
        }),
        { TurnierAussteigen },
      )

      await userEvent.click(
        screen.getByRole('button', { name: `${texte.einsteigen}: Ada` }),
      )

      await waitFor(() => {
        expect(TurnierAussteigen).toHaveBeenCalledWith('7', '3', false)
      })
    })

    it('sieht die nächste Runde an, bevor sie gelegt wird', async () => {
      const TurnierPaarungsvorschlag = vi.fn(() =>
        Promise.resolve({
          paarungen: [
            paarungsprobe(),
            paarungsprobe({ id: '13', a: '5', b: '' }),
          ],
          fehler: null,
        }),
      )
      const TurnierRundeLegen = vi.fn(() =>
        Promise.resolve({ runde: rundenprobe(), fehler: null }),
      )

      // **Das Nachladen braucht eine Weile** — wie in Wirklichkeit. Ein Doppel,
      // das sofort antwortet, kann den Fall nicht zeigen: dann steht der offene
      // Kontext nie auf `null`, und das Bauteil wird nie abgebaut.
      let loeseAuf: (() => void) | undefined

      const Turnier = vi.fn(() => {
        if (Turnier.mock.calls.length === 1) {
          return Promise.resolve({
            turnier: turnierprobe({ teilnehmer: [teilnehmerprobe()] }),
            fehler: null,
          })
        }

        return new Promise<{ turnier: TurnierDTO; fehler: null }>((fertig) => {
          loeseAuf = () => {
            fertig({
              turnier: turnierprobe({ teilnehmer: [teilnehmerprobe()] }),
              fehler: null,
            })
          }
        })
      })

      const werkbank = await aufgeschlagen(
        turnierprobe({ teilnehmer: [teilnehmerprobe()] }),
        { TurnierPaarungsvorschlag, TurnierRundeLegen, Turnier },
        'runden',
      )

      expect(within(werkbank).getByText(texte.rundenLeer)).toBeInTheDocument()

      const rufeVorher = Turnier.mock.calls.length

      await userEvent.click(
        within(werkbank).getByRole('button', { name: texte.vorschlag }),
      )

      // **Der Vorschlag steht da — und hat nichts angelegt.**
      const vorschlag = await screen.findByRole('list', {
        name: texte.vorschlagTitel,
      })

      expect(within(vorschlag).getByText(/Ada/)).toBeInTheDocument()
      expect(TurnierRundeLegen).not.toHaveBeenCalled()

      // **Und er bleibt stehen.**
      //
      // Der erste Anlauf wartete nur darauf, dass die Liste **einmal**
      // erscheint. Der Vorschlag lief durch den Schreibweg, und der räumt nach
      // getaner Arbeit den offenen Kontext und lädt neu — dabei wurde das
      // Bauteil abgebaut und sein Vorschlag mit ihm. Er blitzte auf und war
      // weg (Befund 1 des Nachlaufs über Phase 4.7).
      //
      // **Zugesichert wird der Weg und die Wirkung**: ein Vorschlag lädt nicht
      // nach, und die Liste steht danach immer noch da. Wer ihn zurück in den
      // Schreibweg legt, sieht beides fehlschlagen — der Ruf kommt, und mit ihm
      // ist die Liste weg.
      expect(Turnier).toHaveBeenCalledTimes(rufeVorher)

      // Ist doch nachgeladen worden, kommt hier die Antwort — und die Liste
      // verschwindet.
      loeseAuf?.()
      await effekteNachholen()

      expect(
        screen.queryByRole('list', { name: texte.vorschlagTitel }),
      ).not.toBeNull()

      await userEvent.click(await gebietUndKnopfRunden(texte.rundeLegen))

      await waitFor(() => {
        expect(TurnierRundeLegen).toHaveBeenCalledWith('7')
      })

      // **Und danach ist er weg** — sonst stünde er neben der Runde, die aus
      // ihm entstand.
      await waitFor(() => {
        expect(
          screen.queryByRole('list', { name: texte.vorschlagTitel }),
        ).toBeNull()
      })
    })

    it('trägt ein Ergebnis ein und nimmt es zurück', async () => {
      const TurnierErgebnisSetzen = vi.fn(() =>
        Promise.resolve({ fehler: null }),
      )

      const werkbank = await aufgeschlagen(
        turnierprobe({
          teilnehmer: [
            teilnehmerprobe(),
            teilnehmerprobe({ id: '4', name: 'Bert', stelle: 2 }),
          ],
          runden: [
            rundenprobe({
              paarungen: [
                paarungsprobe({
                  ergebnis: { spieleA: 2, spieleB: 0, unentschieden: 0 },
                }),
              ],
            }),
          ],
        }),
        { TurnierErgebnisSetzen, TurnierKorrekturfolgen: ohneFolgen },
        'runden',
      )

      // **Das Eintragen steht hinter *Ergebnis*** — seit UXR-12 (#1920).
      await ergebnisOeffnen(werkbank, `Ada ${texte.gegen} Bert`)
      const feld = within(werkbank).getByLabelText(
        `${texte.spieleA} — Ada ${texte.gegen} Bert`,
      )

      await userEvent.clear(feld)
      await userEvent.type(feld, '1')
      await userEvent.click(
        within(werkbank).getByRole('button', {
          name: `${texte.eintragen}: Ada ${texte.gegen} Bert`,
        }),
      )

      await waitFor(() => {
        expect(TurnierErgebnisSetzen).toHaveBeenCalledWith('7', '1', '12', {
          spieleA: 1,
          spieleB: 0,
          unentschieden: 0,
        })
      })

      // **Nach dem Schreiben baut sich die Werkbank neu auf** (`fuehre`
      // lädt nach), und das Eintragen ist wieder zu — wer zurücknehmen will,
      // klappt es auf. Seit UXR-12 (#1920); vorher stand es offen da.
      await waitFor(() => {
        expect(
          screen.getByRole('button', {
            name: `${texte.ergebnis}: Ada ${texte.gegen} Bert`,
          }),
        ).toHaveAttribute('aria-expanded', 'false')
      })
      await userEvent.click(
        screen.getByRole('button', {
          name: `${texte.ergebnis}: Ada ${texte.gegen} Bert`,
        }),
      )

      // **Dieselbe Gestalt wie beim Abmelden**: der Knopf zum Zurücknehmen
      // entsteht erst mit der Antwort, und er ruht, solange sie läuft.
      await waitFor(() => {
        expect(
          screen.getByRole('button', {
            name: `${texte.zuruecknehmen}: Ada ${texte.gegen} Bert`,
          }),
        ).toBeEnabled()
      })

      await userEvent.click(
        screen.getByRole('button', {
          name: `${texte.zuruecknehmen}: Ada ${texte.gegen} Bert`,
        }),
      )

      await waitFor(() => {
        expect(TurnierErgebnisSetzen).toHaveBeenLastCalledWith(
          '7',
          '1',
          '12',
          null,
        )
      })
    })
  })
})
