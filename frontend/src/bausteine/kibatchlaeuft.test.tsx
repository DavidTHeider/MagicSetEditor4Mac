// **Was der Knopf sagt, solange etwas läuft** — Werkstück A13.1b von Phase
// 5.3 (#1452).
//
// # Eine dritte Datei
//
// **Weil `kibatchwege.test.tsx` sonst über die Zeilengrenze aus §2 ginge** —
// und weil diese beiden Läufe eine eigene Vorrichtung brauchen: eine Antwort,
// die **hängt**. Ohne sie ist der laufende Zustand zu schnell vorbei, um ihn
// zu sehen, und §7 Regel 4 bliebe ungemessen.
//
// **Sie haben einen Fehler gefunden**: gestartet wurde unter
// `batch/${setID}`, gefragt wurde nach `setID` — zwei Schlüssel, und der
// laufende Zustand war nie sichtbar.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it } from 'vitest'

import { KIBatchkasten } from './KIBatchkasten'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { KIBatchAntwort } from '../antwortenbatch'

const eigene = kataloge.de.seiten.einstellungen

afterEach(() => {
  raeumeApi()
})

// zweiKarten ist die Liste, aus der gewählt wird.
function zweiKarten() {
  return Promise.resolve({
    karten: [
      { id: '1', nummer: 1, werte: [{ feld: 'name', wert: 'Blitzschlag' }] },
      { id: '2', nummer: 2, werte: [] },
    ],
    fehler: null,
  })
}

// einLokalesModell kostet nichts und fragt deshalb nicht nach einer Freigabe.
function einLokalesModell() {
  return Promise.resolve({
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
        laufzeit: 'Ollama 0.12.3',
        ort: 'lokal',
        faehigkeiten: [],
        herkunft: [],
        bindung: 'ollama/gemma4:8b',
        kontextfenster: 0,
        kontextbeleg: '',
      },
    ],
    fehler: null,
  })
}

// einZiel ist ein geglückter Vorschlag.
function einZiel(): Promise<KIBatchAntwort> {
  return Promise.resolve({
    kennung: 'kibatch-1',
    anbieter: 'ollama',
    ziele: [
      {
        zielkarte: 1,
        kennung: 'a1',
        modell: 'gemma4:8b',
        felder: [{ feld: 'name', werte: ['Donnerschlag'], bisher: '' }],
        befunde: [],
        ablauf: '2026-09-13T14:00:00Z',
        fehler: null,
      },
    ],
    fehler: null,
  })
}

// aufklappenUndFuellen stellt den Kasten hin, klappt auf und füllt aus.
async function aufklappenUndFuellen() {
  const hand = userEvent.setup()

  render(
    <Sprachraum sprache="de">
      <KIBatchkasten setID="7" geaendert={() => undefined} />
    </Sprachraum>,
  )

  await hand.click(screen.getByText(eigene.kiBatchTitel))
  await hand.click(await screen.findByLabelText('Blitzschlag'))
  await hand.selectOptions(
    screen.getByLabelText(eigene.kiVorschlagAnbieter),
    'ollama/gemma4:8b',
  )
  await hand.type(screen.getByLabelText(eigene.kiBatchAuftrag), 'Schreibe.')

  return hand
}

// **Solange etwas läuft, sagt der Knopf es — und nimmt keinen zweiten Ruf.**
//
// **§7, Regel 4**: ein zweiter Auftrag, während der erste unterwegs ist, geht
// gegen einen Stand, den das Backend noch nicht kennt. Gemessen wird an einer
// Antwort, die **hängt**: ohne sie ist der laufende Zustand zu schnell vorbei,
// um ihn zu sehen.
it('sperrt den Knopf, solange der Batch unterwegs ist', async () => {
  let loesen: (antwort: KIBatchAntwort) => void = () => undefined

  const haengt = new Promise<KIBatchAntwort>((fertig) => {
    loesen = fertig
  })

  belegeApi({
    Karten: zweiKarten,
    KIModelle: einLokalesModell,
    KIBatchVorschlagen: () => haengt,
  })

  const hand = await aufklappenUndFuellen()
  await hand.click(screen.getByRole('button', { name: eigene.kiBatchFahren }))

  expect(
    await screen.findByRole('button', { name: eigene.kiBatchLaeuft }),
  ).toBeDisabled()

  loesen(await einZiel())

  expect(await screen.findByText(eigene.kiBatchLagen)).toBeInTheDocument()
})

// **Und dasselbe für die Gruppe.**
it('sperrt den Knopf, solange die Gruppe unterwegs ist', async () => {
  let loesen: (antwort: unknown) => void = () => undefined

  const haengt = new Promise((fertig) => {
    loesen = fertig
  })

  belegeApi({
    Karten: zweiKarten,
    KIModelle: einLokalesModell,
    KIBatchVorschlagen: einZiel,
    KIGruppeUebernehmen: () => haengt as Promise<never>,
  })

  const hand = await aufklappenUndFuellen()
  await hand.click(screen.getByRole('button', { name: eigene.kiBatchFahren }))

  await hand.click(
    await screen.findByRole('button', { name: /Als Gruppe übernehmen/ }),
  )

  expect(
    await screen.findByRole('button', { name: eigene.kiBatchGruppeLaeuft }),
  ).toBeDisabled()

  loesen({ karten: 1, felder: 1, namen: [], fehler: null })

  expect(await screen.findByRole('status')).toBeInTheDocument()
})
