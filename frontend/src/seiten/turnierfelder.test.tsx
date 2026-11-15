// **Die Felder der Turnierseite, die der große Ablauf nicht anfasst** —
// Werkstück 5 von Phase 4.7
// (#1294).
//
// **Eine eigene Datei**, weil `turnierlauf.test.tsx` mit ihnen über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort steht, was die
// Seite über die Zeit **tut**, hier, was ihre einzelnen Felder **mitnehmen**.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Turniere } from './Turniere'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { ergebnisOeffnen, zumTurniergebiet } from './turnierwerkbank'
import { kataloge } from '../sprache/kataloge'
import { assistent } from '../test/turnierassistent'
import {
  paarungsprobe,
  rangprobe,
  rundenprobe,
  teilnehmerprobe,
  turnierprobe,
} from './turnierprobe'
import type { TurnierDTO } from '../apiturniere'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

// aufgeschlagen stellt die Seite hin und schlägt das Turnier auf.
//
// **Dieselbe Hilfe wie in `turnierlauf.test.tsx`** — sie steht zweimal da,
// weil beide Dateien sie brauchen und ein geteilter Baustein für zwei
// Aufrufer ein Baustein zu viel wäre.
async function aufgeschlagen(
  turnier: TurnierDTO,
  weitere: Record<string, unknown> = {},

  // gebiet ist die Lasche, die danach offen sein soll — Werkstück I von Phase
  // 4.10 (#1325).
  gebiet: 'teilnehmer' | 'runden' | 'rangliste' = 'teilnehmer',
) {
  belegeApi({
    Turniere: () => Promise.resolve({ turniere: [turnier], fehler: null }),
    Turnier: () => Promise.resolve({ turnier, fehler: null }),
    TurnierRangliste: () =>
      Promise.resolve({ rangliste: [rangprobe()], fehler: null }),
    ...weitere,
  })

  render(
    <Sprachraum sprache="de">
      <Turniere />
    </Sprachraum>,
  )

  await userEvent.click(
    await screen.findByRole('button', {
      name: `${texte.waehlen}: ${turnier.name}`,
    }),
  )

  await zumTurniergebiet(gebiet)

  return screen.findByRole('region', { name: turnier.name })
}

// **Die Felder, die der große Ablauf nicht anfasst.**
//
// **Ein eigener Abschnitt und kein Beiwerk**: jedes davon ist ein Zweig, den
// sonst niemand betritt — und ein Feld, das niemand tippt, ist eines, von dem
// niemand weiß, ob es ankommt.
describe('die übrigen Felder', () => {
  it('nimmt Spiel und Notiz beim Anlegen mit', async () => {
    const TurnierAnlegen = vi.fn(() =>
      Promise.resolve({ turnier: turnierprobe(), fehler: null }),
    )

    belegeApi({
      Turniere: () => Promise.resolve({ turniere: [], fehler: null }),
      Turnier: () => Promise.resolve({ turnier: null, fehler: null }),
      TurnierRangliste: () => Promise.resolve({ rangliste: [], fehler: null }),
      TurnierAnlegen,
    })

    render(
      <Sprachraum sprache="de">
        <Turniere />
      </Sprachraum>,
    )

    const bereich = await assistent()

    await userEvent.type(within(bereich).getByLabelText(texte.name), 'Freitag')
    await userEvent.clear(within(bereich).getByLabelText(texte.spiel))
    await userEvent.type(within(bereich).getByLabelText(texte.spiel), 'vs')
    await userEvent.type(
      within(bereich).getByLabelText(texte.notiz),
      'Am Küchentisch',
    )
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.weiter }),
    )
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.anlegen }),
    )

    await waitFor(() => {
      expect(TurnierAnlegen).toHaveBeenCalledWith(
        'Freitag',
        'vs',
        'Am Küchentisch',
        false,
        '',
      )
    })
  })

  it('nimmt die beiden anderen Spielzahlen mit', async () => {
    const TurnierErgebnisSetzen = vi.fn(() => Promise.resolve({ fehler: null }))

    const werkbank = await aufgeschlagen(
      turnierprobe({
        teilnehmer: [
          teilnehmerprobe(),
          teilnehmerprobe({ id: '4', name: 'Bert', stelle: 2 }),
        ],
        runden: [rundenprobe()],
      }),
      { TurnierErgebnisSetzen },
      'runden',
    )

    const gegen = `Ada ${texte.gegen} Bert`

    // **Das Eintragen steht hinter *Ergebnis*** — seit UXR-12 (#1920).
    await ergebnisOeffnen(werkbank, gegen)

    for (const [aufschrift, wert] of [
      [texte.spieleB, '2'],
      [texte.unentschiedenSpiele, '1'],
    ] as const) {
      const feld = within(werkbank).getByLabelText(`${aufschrift} — ${gegen}`)

      await userEvent.clear(feld)
      await userEvent.type(feld, wert)
    }

    await userEvent.click(
      within(werkbank).getByRole('button', {
        name: `${texte.eintragen}: ${gegen}`,
      }),
    )

    await waitFor(() => {
      expect(TurnierErgebnisSetzen).toHaveBeenCalledWith('7', '1', '12', {
        spieleA: 0,
        spieleB: 2,
        unentschieden: 1,
      })
    })
  })

  // **Eine unbrauchbare Eingabe wird nicht zur Null** — berichtigt in
  // Werkstück U3 von Phase 5.1
  // (#1414).
  //
  // # Was hier vorher stand
  //
  // „macht aus einer unbrauchbaren Zahl eine Null" — und schickte damit
  // 0:0:0 los. Genau das weist `domain.Ergebnis.Pruefe` ab („ein Ergebnis
  // ohne ein einziges Spiel ist kein Ergebnis"): der Fehler entstand im
  // Fenster und wurde am anderen Ende sichtbar, als Satz über einer Paarung,
  // die man gar nicht so eingetragen hatte.
  //
  // Seither steht der Befund am Feld, und der Knopf ruht (§6).
  it('macht aus einer unbrauchbaren Zahl keine Null', async () => {
    const TurnierErgebnisSetzen = vi.fn(() => Promise.resolve({ fehler: null }))

    const werkbank = await aufgeschlagen(
      turnierprobe({
        teilnehmer: [
          teilnehmerprobe(),
          teilnehmerprobe({ id: '4', name: 'Bert', stelle: 2 }),
        ],
        runden: [rundenprobe()],
      }),
      { TurnierErgebnisSetzen },
      'runden',
    )

    await ergebnisOeffnen(werkbank, `Ada ${texte.gegen} Bert`)
    const feld = within(werkbank).getByLabelText(
      `${texte.spieleA} — Ada ${texte.gegen} Bert`,
    )

    await userEvent.clear(feld)
    await userEvent.type(feld, 'zwei')

    expect(feld).toHaveAttribute('aria-invalid', 'true')
    expect(within(werkbank).getByRole('alert')).toHaveTextContent(
      texte.ergebnisZahlen,
    )

    const eintragen = within(werkbank).getByRole('button', {
      name: `${texte.eintragen}: Ada ${texte.gegen} Bert`,
    })

    expect(eintragen).toBeDisabled()

    await userEvent.click(eintragen)

    expect(TurnierErgebnisSetzen).not.toHaveBeenCalled()
  })

  // **Ein Teilnehmer, den die Paarung nicht kennt, steht mit seiner
  // Kennung da.** Das ist die ehrliche Vorgabe: eine leere Zelle sähe aus,
  // als fehlte der Gegner.
  it('zeigt eine unbekannte Kennung statt einer Lücke', async () => {
    const werkbank = await aufgeschlagen(
      turnierprobe({
        teilnehmer: [teilnehmerprobe()],
        runden: [rundenprobe({ paarungen: [paarungsprobe({ b: '99' })] })],
      }),
      {},
      'runden',
    )

    expect(
      within(werkbank).getByText(`Ada ${texte.gegen} 99`),
    ).toBeInTheDocument()
  })

  // **Und ein Vorschlag mit unbekannten Kennungen ebenso.**
  it('zeigt im Vorschlag eine unbekannte Kennung', async () => {
    const werkbank = await aufgeschlagen(
      turnierprobe(),
      {
        TurnierPaarungsvorschlag: () =>
          Promise.resolve({
            paarungen: [
              paarungsprobe({ a: '98', b: '99' }),
              paarungsprobe({ id: '14', a: '97', b: '' }),
            ],
            fehler: null,
          }),
      },
      'runden',
    )

    await userEvent.click(
      within(werkbank).getByRole('button', { name: texte.vorschlag }),
    )

    const vorschlag = await screen.findByRole('list', {
      name: texte.vorschlagTitel,
    })

    expect(
      within(vorschlag).getByText(`98 ${texte.gegen} 99`),
    ).toBeInTheDocument()

    expect(within(vorschlag).getByText(`97 — ${texte.bye}`)).toBeInTheDocument()
  })

  // **Ein Ausgestiegener steht in der Rangliste mit seinem Vermerk.**
  it('nennt den Ausgestiegenen auch in der Rangliste', async () => {
    const werkbank = await aufgeschlagen(
      turnierprobe(),
      {
        TurnierRangliste: () =>
          Promise.resolve({
            rangliste: [rangprobe({ ausgestiegen: true })],
            fehler: null,
          }),
      },
      'rangliste',
    )

    expect(
      within(werkbank).getByText(new RegExp(texte.ausgestiegen)),
    ).toBeInTheDocument()
  })
})

// **Die Empfehlung aus Appendix E ist zu sehen** — Befund 2 des Zweitblicks
// über Phase 4.7.
//
// **Sie war zugesagt und nicht gebaut.** Die Zahl stand nur als
// Schleifengrenze in einem Test; ein richtiger Quellenwert in einer
// Vorrichtung ist keine Anwendungsfunktion.
describe('die empfohlene Rundenzahl', () => {
  it('steht neben den Runden', async () => {
    const werkbank = await aufgeschlagen(
      turnierprobe({
        empfohleneRunden: 3,
        teilnehmer: [
          teilnehmerprobe(),
          teilnehmerprobe({ id: '4', name: 'Bert', stelle: 2 }),
        ],
        runden: [rundenprobe()],
      }),
      {},
      'runden',
    )

    expect(
      within(werkbank).getByText(
        texte.empfehlung
          .replace('{gelegt}', '1')
          .replaceAll('{empfohlen}', '3')
          .replace('{spieler}', '2'),
      ),
    ).toBeInTheDocument()
  })

  // **Ohne Empfehlung steht dort nichts.** Unter fünf Teilnehmern nennt
  // Appendix E keine Swiss-Runden, und eine erfundene Zahl wäre schlimmer
  // als keine.
  it('fehlt, wenn es keine gibt', async () => {
    const werkbank = await aufgeschlagen(
      turnierprobe({ empfohleneRunden: 0 }),
      {},
      'runden',
    )

    expect(
      within(werkbank).queryByText(/Runden — die Turnierregeln/),
    ).toBeNull()
  })

  // **Ein Ausgestiegener zählt nicht mit.**
  it('zählt nur, wer noch spielt', async () => {
    const werkbank = await aufgeschlagen(
      turnierprobe({
        empfohleneRunden: 3,
        teilnehmer: [
          teilnehmerprobe(),
          teilnehmerprobe({
            id: '4',
            name: 'Bert',
            stelle: 2,
            ausgestiegen: true,
          }),
        ],
      }),
      {},
      'runden',
    )

    expect(
      within(werkbank).getByText(
        texte.empfehlung
          .replace('{gelegt}', '0')
          .replaceAll('{empfohlen}', '3')
          .replace('{spieler}', '1'),
      ),
    ).toBeInTheDocument()
  })
})
