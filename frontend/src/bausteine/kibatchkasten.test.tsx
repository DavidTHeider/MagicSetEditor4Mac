// **Der Batchkasten und die atomare Gruppe am Fenster** — Werkstück A13.1b
// von Phase 5.3
// (#1452).
//
// # Was hier gemessen wird
//
// **Dass der erste Schritt nichts schreibt** — er holt Vorschläge, und die
// Karten bleiben, wie sie waren.
//
// **Dass die Übernahme EIN Ruf ist und nicht einer je Karte.** Das ist die
// ganze Zusage: „selektive Übernahme als definierte atomare Gruppe". Ein
// zweiter Ruf wäre eine zweite Transaktion.
//
// **Und dass ein gescheitertes Ziel dasteht** — mit seinem Grund, und ohne
// Häkchen: es gibt nichts zu übernehmen.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { KIBatchkasten } from './KIBatchkasten'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { KIBatchAntwort } from '../antwortenbatch'

const eigene = kataloge.de.seiten.einstellungen

afterEach(() => {
  raeumeApi()
})

// eineKartenliste gibt zwei Karten mit Namen.
function eineKartenliste() {
  return Promise.resolve({
    karten: [
      { id: '1', nummer: 1, werte: [{ feld: 'name', wert: 'Blitzschlag' }] },
      { id: '2', nummer: 2, werte: [{ feld: 'name', wert: 'Eisregen' }] },
    ],
    fehler: null,
  })
}

// einKatalog führt ein lokales Modell — es kostet nichts.
function einKatalog() {
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

// einBatch gibt ein geglücktes und ein gescheitertes Ziel.
//
// **Beides in einer Antwort**, weil genau das der Fall ist, für den es den
// Batch gibt: „Ein Ziel, das scheitert, hält ihn nicht auf."
function einBatch(): Promise<KIBatchAntwort> {
  return Promise.resolve({
    kennung: 'kibatch-1',
    anbieter: 'ollama',
    ziele: [
      {
        zielkarte: 1,
        kennung: 'a1',
        modell: 'gemma4:8b',
        felder: [
          { feld: 'name', werte: ['Donnerschlag'], bisher: 'Blitzschlag' },
        ],
        befunde: [],
        ablauf: '2026-09-13T14:00:00Z',
        fehler: null,
      },
      {
        zielkarte: 2,
        kennung: '',
        modell: '',
        felder: [],
        befunde: [],
        ablauf: '',
        fehler: {
          code: 'nichtGefunden',
          schluessel: 'fehler.nichtGefunden',
          meldung: 'Diese Karte gibt es nicht',
          feld: '',
          grund: '',
        },
      },
    ],
    fehler: null,
  })
}

// aufklappenUndFahren klappt auf, wählt, füllt und drückt.
async function aufklappenUndFahren() {
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
  await hand.type(
    screen.getByLabelText(eigene.kiBatchAuftrag),
    'Schreibe den Regeltext.',
  )
  await hand.click(screen.getByRole('button', { name: eigene.kiBatchFahren }))

  return hand
}

// **Der Batch zeigt je Ziel, was daraus wurde — und schreibt nichts.**
it('zeigt je Ziel seinen Ausgang', async () => {
  const gruppe = vi.fn(() =>
    Promise.resolve({ karten: 1, felder: 1, namen: [], fehler: null }),
  )

  belegeApi({
    Karten: eineKartenliste,
    KIModelle: einKatalog,
    KIBatchVorschlagen: einBatch,
    KIGruppeUebernehmen: gruppe,
  })

  await aufklappenUndFahren()

  expect(await screen.findByText(eigene.kiBatchLagen)).toBeInTheDocument()
  expect(screen.getByText('Donnerschlag')).toBeInTheDocument()
  expect(screen.getByText(/Diese Karte gibt es nicht/)).toBeInTheDocument()

  // **Der erste Schritt hat nichts geschrieben.**
  expect(gruppe).not.toHaveBeenCalled()
})

// **Die Übernahme ist EIN Ruf über alle Ziele.**
it('übernimmt als eine Gruppe und nicht je Karte', async () => {
  const gruppe = vi.fn(() =>
    Promise.resolve({ karten: 1, felder: 1, namen: [], fehler: null }),
  )

  belegeApi({
    Karten: eineKartenliste,
    KIModelle: einKatalog,
    KIBatchVorschlagen: einBatch,
    KIGruppeUebernehmen: gruppe,
  })

  const hand = await aufklappenUndFahren()

  await hand.click(
    await screen.findByRole('button', { name: /Als Gruppe übernehmen/ }),
  )

  expect(await screen.findByRole('status')).toHaveTextContent(
    '1 Karten geschrieben.',
  )

  // **Einmal, nicht zweimal.** Zwei Rufe wären zwei Transaktionen — und
  // dazwischen läge der Teilbestand, den A13.1b ausschließt.
  expect(gruppe).toHaveBeenCalledTimes(1)
  expect(gruppe).toHaveBeenCalledWith([{ kennung: 'a1', felder: ['name'] }])
})

// **Ein Ziel ohne Haken fällt aus der Gruppe — und sperrt sie nicht.**
it('schickt kein Ziel ohne ausgewähltes Feld mit', async () => {
  const gruppe = vi.fn(() =>
    Promise.resolve({ karten: 0, felder: 0, namen: [], fehler: null }),
  )

  belegeApi({
    Karten: eineKartenliste,
    KIModelle: einKatalog,
    KIBatchVorschlagen: einBatch,
    KIGruppeUebernehmen: gruppe,
  })

  const hand = await aufklappenUndFahren()

  // Das einzige Feld des einzigen geglückten Ziels abhaken.
  await hand.click(await screen.findByLabelText(/name/))

  expect(
    screen.getByRole('button', { name: /Als Gruppe übernehmen/ }),
  ).toBeDisabled()
})
