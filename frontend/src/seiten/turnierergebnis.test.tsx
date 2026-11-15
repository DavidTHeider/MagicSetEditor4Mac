// **Was ein Spielstand am Feld sagt** — Werkstück U3 von Phase 5.1
// (#1414).
//
// # Die Regeln stehen in `domain.Ergebnis.Pruefe` und nicht hier
//
//	case e.SpieleA < 0 || e.SpieleB < 0 || e.Unentschieden < 0:
//	case e.SpieleA+e.SpieleB+e.Unentschieden == 0:
//	case hoechstens > 0 && e.SpieleA+e.SpieleB+e.Unentschieden > hoechstens:
//
// **Die Obergrenze gilt für die Summe und nicht je Zahl**, und wie hoch sie
// ist, sagt `service.hoechstensSpiele`: bei einem Einzelspielturnier eins,
// sonst null — und null heißt keine Grenze.
//
// **Eine eigene Datei**, damit `turnierlauf.test.tsx` unter der Zeilengrenze
// aus §2 bleibt. Der Schnitt folgt der Frage: dort steht, was die Seite über
// die Zeit **tut**, hier, was sie über eine **Eingabe** urteilt.
import { waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { aufgeschlagen, ergebnisOeffnen } from './turnierwerkbank'
import { raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import {
  ohneFolgen,
  paarungsprobe,
  rundenprobe,
  teilnehmerprobe,
} from './turnierprobe'
import { turnierprobe } from './turnierprobe'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

// mitPaarung stellt ein Turnier mit **einer** gespielten Paarung hin.
async function mitPaarung(einzelspiel = false) {
  const TurnierErgebnisSetzen = vi.fn(() => Promise.resolve({ fehler: null }))

  const werkbank = await aufgeschlagen(
    turnierprobe({
      einzelspiel,
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

  return { werkbank, TurnierErgebnisSetzen }
}

// felder gibt die drei Kästchen einer Paarung.
function felder(werkbank: HTMLElement) {
  const wer = `Ada ${texte.gegen} Bert`

  return {
    a: within(werkbank).getByLabelText(`${texte.spieleA} — ${wer}`),
    b: within(werkbank).getByLabelText(`${texte.spieleB} — ${wer}`),
    remis: within(werkbank).getByLabelText(
      `${texte.unentschiedenSpiele} — ${wer}`,
    ),
    eintragen: within(werkbank).getByRole('button', {
      name: `${texte.eintragen}: ${wer}`,
    }),
  }
}

// **Die rohe Zeichenkette und nicht der geparste Wert.** Bis heute machte
// `Number.isFinite(gelesen) ? gelesen : 0` aus jedem Vertipper eine Null —
// und drei Nullen sind kein Ergebnis, sondern ein Stand, den Go abweist.
it('meldet eine negative Zahl am Feld', async () => {
  const { werkbank } = await mitPaarung()

  const { a, eintragen } = felder(werkbank)

  await userEvent.clear(a)
  await userEvent.type(a, '-1')

  expect(a).toHaveAttribute('aria-invalid', 'true')
  expect(within(werkbank).getByRole('alert')).toHaveTextContent(
    texte.ergebnisZahlen,
  )
  expect(eintragen).toBeDisabled()
})

// **Ein Ergebnis ohne ein einziges Spiel ist kein Ergebnis** — genau der Satz
// aus `Ergebnis.Pruefe`.
it('meldet drei Nullen als kein Ergebnis', async () => {
  const { werkbank } = await mitPaarung()

  const { a, eintragen } = felder(werkbank)

  await userEvent.clear(a)
  await userEvent.type(a, '0')

  expect(within(werkbank).getByRole('alert')).toHaveTextContent(
    texte.ergebnisOhneSpiel,
  )
  expect(eintragen).toBeDisabled()
})

// **Die Summe und nicht die einzelne Zahl.** Bei einem Einzelspielturnier
// geht ein Match über ein Spiel; 1:1 sind zwei.
it('meldet zu viele Spiele in einem Einzelspielturnier', async () => {
  const { werkbank } = await mitPaarung(true)

  const { a, b, eintragen } = felder(werkbank)

  await userEvent.clear(a)
  await userEvent.type(a, '1')
  await userEvent.clear(b)
  await userEvent.type(b, '1')

  expect(a).toHaveAttribute('aria-invalid', 'true')
  expect(b).toHaveAttribute('aria-invalid', 'true')
  expect(eintragen).toBeDisabled()
})

// **Und der Grenzfall geht durch**: ein Spiel in einem Einzelspielturnier ist
// genau die erlaubte Höchstzahl.
it('lässt den Grenzfall eines Einzelspiels durch', async () => {
  const { werkbank, TurnierErgebnisSetzen } = await mitPaarung(true)

  const { a, b, eintragen } = felder(werkbank)

  await userEvent.clear(a)
  await userEvent.type(a, '1')
  await userEvent.clear(b)
  await userEvent.type(b, '0')

  expect(a).toHaveAttribute('aria-invalid', 'false')
  await userEvent.click(eintragen)

  await waitFor(() => {
    expect(TurnierErgebnisSetzen).toHaveBeenCalledWith('7', '1', '12', {
      spieleA: 1,
      spieleB: 0,
      unentschieden: 0,
    })
  })
})

// **Zurücknehmen braucht keinen gültigen Stand.** Es schickt `null` und
// benutzt die drei Kästchen gar nicht — ein Knopf, der an einer Eingabe
// hinge, die er nicht liest, wäre eine Sperre ohne Grund.
it('nimmt ein Ergebnis auch bei ungültiger Eingabe zurück', async () => {
  const { werkbank, TurnierErgebnisSetzen } = await mitPaarung()

  const { a } = felder(werkbank)

  await userEvent.clear(a)
  await userEvent.type(a, '-1')

  await userEvent.click(
    within(werkbank).getByRole('button', {
      name: `${texte.zuruecknehmen}: Ada ${texte.gegen} Bert`,
    }),
  )

  await waitFor(() => {
    expect(TurnierErgebnisSetzen).toHaveBeenCalledWith('7', '1', '12', null)
  })
})
