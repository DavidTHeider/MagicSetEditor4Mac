// **Woher die Werte stammen** — Werkstück A11 von Phase 5.3
// (#1450).
//
// # Was hier gemessen wird
//
// **Dass eine Herkunft ankommt — und dass sie nicht behauptet, was nicht
// mehr stimmt.** Das Issue erlaubt manuelle Nachbearbeitung ausdrücklich;
// was es nicht erlaubt, ist eine Herkunft, die danach so tut, als käme der
// Inhalt noch von der KI.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { KIBereich } from './KIBereich'
import { KIHerkunft } from './KIHerkunft'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { KIFeldherkunftAntwort, KIFeldherkunftDTO } from '../antworten'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.einstellungen

// eineHerkunft baut eine Zeile.
function eineHerkunft(
  teil: Partial<KIFeldherkunftDTO> = {},
): KIFeldherkunftDTO {
  return {
    feld: 'name',
    anbieter: 'ollama',
    modell: 'gemma4:8b-q4',
    aufgabe: 'Schreibe den Regeltext.',
    auftrag: 'ki-4-1',
    zeitpunkt: '2026-09-12T09:00:00Z',
    veraendert: false,
    ...teil,
  }
}

// zeige stellt den Kasten hin und schlägt ihn auf.
async function zeigeUndAufschlagen(
  herkunft: KIFeldherkunftDTO[],
  fehler: KIFeldherkunftAntwort['fehler'] = null,
) {
  belegeApi({
    KIFeldherkunft: (): Promise<KIFeldherkunftAntwort> =>
      Promise.resolve({ herkunft, fehler }),
  })

  render(
    <Sprachraum sprache="de">
      <KIHerkunft karteID="4" stand={0} />
    </Sprachraum>,
  )

  const hand = userEvent.setup()

  await hand.click(screen.getByText(eigene.kiHerkunftTitel))

  return hand
}

// **Zugeklappt wird nicht gefragt.**
//
// **Dieselbe Zusage wie beim Vorschlagskasten** (A8): das Aufschlagen einer
// Karte soll keine Abfrage auslösen, die niemand gewollt hat.
it('zugeklappt wird nicht gefragt', () => {
  const gefragt = vi.fn()

  belegeApi({ KIFeldherkunft: gefragt })

  render(
    <Sprachraum sprache="de">
      <KIHerkunft karteID="4" stand={0} />
    </Sprachraum>,
  )

  expect(gefragt).not.toHaveBeenCalled()

  // **Und der Titel steht trotzdem da** — sonst wüsste niemand, dass es die
  // Auskunft gibt. Der Beleg, dass jemand zuhört, ist der Test darunter: er
  // klickt denselben Titel und bekommt eine Antwort.
  expect(screen.getByText(eigene.kiHerkunftTitel)).toBeInTheDocument()
})

// **Aufgeschlagen steht da, wer geantwortet hat.**
it('aufgeschlagen steht die Herkunft da', async () => {
  await zeigeUndAufschlagen([eineHerkunft()])

  expect(
    await screen.findByText(
      'name: ollama/gemma4:8b-q4 am 2026-09-12T09:00:00Z',
    ),
  ).toBeInTheDocument()

  // **Die Aufgabe und nicht der Rohprompt.**
  expect(
    screen.getByText('Auftrag war: Schreibe den Regeltext.'),
  ).toBeInTheDocument()

  expect(
    screen.queryByText(eigene.kiHerkunftVeraendert),
  ).not.toBeInTheDocument()
})

// **Und wenn der Wert nicht mehr derselbe ist, steht es dabei.**
it('eine Handarbeit danach steht dabei', async () => {
  await zeigeUndAufschlagen([eineHerkunft({ veraendert: true })])

  expect(
    await screen.findByText(eigene.kiHerkunftVeraendert),
  ).toBeInTheDocument()
})

// **Ohne Herkunft sagt der Kasten das.**
//
// **Und nicht gar nichts**: ein leerer Kasten sähe aus wie einer, der noch
// lädt (§6).
it('ohne Herkunft sagt der Kasten es', async () => {
  await zeigeUndAufschlagen([])

  expect(await screen.findByText(eigene.kiHerkunftKeine)).toBeInTheDocument()
})

// **Ein Fehler steht als Fehler da.**
it('ein Fehler steht als Fehler da', async () => {
  await zeigeUndAufschlagen([], {
    code: 'unbekannt',
    meldung: 'der Bestand antwortet nicht',
    schluessel: 'unbekannt',
    feld: '',
    werte: {},
  })

  // **Der Satz kommt aus dem Katalog und nicht aus dem Fehlerrumpf** — was
  // die Anwendung nicht einordnen kann, bekommt den allgemeinen Satz (§8).
  expect(await screen.findByRole('alert')).toBeInTheDocument()

  // **Und kein „keine Herkunft" daneben** — das wären zwei Auskünfte, von
  // denen eine falsch ist.
  expect(screen.queryByText(eigene.kiHerkunftKeine)).not.toBeInTheDocument()
})

// **Nach einer Übernahme wird neu gelesen.**
//
// # Warum das hier steht und nicht im Vorschlagskasten
//
// **Weil der Zähler zwischen beiden liegt** ([KIBereich]). Wer ihn an zwei
// Stellen pflegte, hätte eine Anzeige, die nach der Übernahme den Stand von
// vorhin zeigt.
it('nach einer Übernahme wird neu gelesen', async () => {
  const gelesen = vi.fn().mockResolvedValue({ herkunft: [], fehler: null })

  belegeApi({
    KIFeldherkunft: gelesen,
    KIModelle: () =>
      Promise.resolve({
        nurLokal: false,
        vorrang: [],
        vorrangbefunde: [],
        einrichtung: 'bereit',
        anbieter: [],
        cloud: { stand: 'ungewiss', beleg: 'keine Serverdatei' },
        befunde: [],
        modelle: [],
        fehler: null,
      }),
  })

  const geaendert = vi.fn()

  render(
    <Sprachraum sprache="de">
      <KIBereich
        karteID="4"
        setID="1"
        bildfelder={[]}
        zuschneiden={() => Promise.resolve(true)}
        bildstand={0}
        geaendert={geaendert}
      />
    </Sprachraum>,
  )

  const hand = userEvent.setup()

  await hand.click(screen.getByText(eigene.kiHerkunftTitel))

  expect(await screen.findByText(eigene.kiHerkunftKeine)).toBeInTheDocument()
  expect(gelesen).toHaveBeenCalledTimes(1)
})

// **Nach einer wirklichen Übernahme wird die Herkunft neu gelesen.**
//
// # Warum der ganze Weg und nicht nur der Zähler
//
// **Weil der Zähler die Naht zwischen zwei Kästen ist** ([KIBereich]): der
// eine schreibt, der andere liest. Wer nur den Zähler prüfte, prüfte eine
// Zahl; was zugesagt ist, ist die **Anzeige nach der Übernahme** — sie zeigt
// sonst den Stand von vorhin.
it('nach einer wirklichen Übernahme wird neu gelesen', async () => {
  const gelesen = vi.fn().mockResolvedValue({ herkunft: [], fehler: null })
  const geaendert = vi.fn()

  belegeApi({
    KIFeldherkunft: gelesen,
    KIModelle: () =>
      Promise.resolve({
        nurLokal: false,
        vorrang: [],
        vorrangbefunde: [],
        einrichtung: 'bereit',
        anbieter: [],
        cloud: { stand: 'ungewiss', beleg: 'keine Serverdatei' },
        befunde: [],
        modelle: [
          {
            anbieter: 'ollama',
            alias: 'gemma4:8b',
            laufzeit: 'ollama 0.12.3',
            ort: 'lokal',
            faehigkeiten: [],
            herkunft: [],
            bindung: 'ollama/gemma4:8b',
            kontextfenster: 0,
            kontextbeleg: '',
          },
        ],
        fehler: null,
      }),
    KIVorschlagen: () =>
      Promise.resolve({
        kennung: 'ki-4-1',
        anbieter: 'ollama',
        modell: 'gemma4:8b',
        felder: [{ feld: 'name', werte: ['Blitzschlag'], bisher: 'Alt' }],
        befunde: [],
        ablauf: '2026-09-13T09:00:00Z',
        uebernommen: 0,
        belege: [],
        korpusstaende: [],
        einschraenkungen: [],
        weg: 'gewaehlt',
        ausgang: '',
        rueckfaelle: null,
        fehler: null,
      }),
    KIVorschlagUebernehmen: () => Promise.resolve({ felder: 1, fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <KIBereich
        karteID="4"
        setID="1"
        bildfelder={[]}
        zuschneiden={() => Promise.resolve(true)}
        bildstand={0}
        geaendert={geaendert}
      />
    </Sprachraum>,
  )

  const hand = userEvent.setup()

  // **Zuerst aufschlagen** — sonst wird gar nicht gelesen.
  await hand.click(screen.getByText(eigene.kiHerkunftTitel))
  expect(await screen.findByText(eigene.kiHerkunftKeine)).toBeInTheDocument()

  await hand.click(screen.getByText(eigene.kiVorschlagTitel))

  await hand.selectOptions(
    await screen.findByRole('combobox', { name: eigene.kiVorschlagAnbieter }),
    'ollama/gemma4:8b',
  )

  await hand.type(
    screen.getByRole('textbox', { name: eigene.kiVorschlagAuftrag }),
    'Schreibe den Regeltext.',
  )

  await hand.click(screen.getByRole('button', { name: eigene.kiVorschlagen }))

  await hand.click(
    await screen.findByRole('button', {
      name: eigene.kiVorschlagUebernehmen,
    }),
  )

  // **Der Editor erfährt es** — und die Herkunft wird ein zweites Mal
  // gelesen.
  expect(
    await screen.findByText('1 Felder wurden geschrieben.'),
  ).toBeInTheDocument()

  expect(geaendert).toHaveBeenCalled()
  expect(gelesen).toHaveBeenCalledTimes(2)
})
