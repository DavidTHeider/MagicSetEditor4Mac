// **Die übrigen Wege des Batchkastens** — Werkstück A13.1b von Phase 5.3
// (#1452).
//
// # Eine zweite Datei
//
// **Weil `kibatchkasten.test.tsx` sonst über die Zeilengrenze aus §2 ginge**
// — und weil der Schnitt der Frage folgt: dort steht der Weg, der glückt,
// hier stehen die, die **nicht** glücken. Zum Ergebnis gehört auch das Nein
// (§10).
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

// einModell führt eines an einem Ort — `lokal` kostet nichts, `cloud` schon.
function einModell(ort: string) {
  return () =>
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
          anbieter: ort === 'lokal' ? 'ollama' : 'openai',
          alias: 'gemma4:8b',
          laufzeit: 'Ollama 0.12.3',
          ort,
          faehigkeiten: [],
          herkunft: [],
          bindung: (ort === 'lokal' ? 'ollama' : 'openai') + '/gemma4:8b',
          kontextfenster: 0,
          kontextbeleg: '',
        },
      ],
      fehler: null,
    })
}

// einBatchMitBefund gibt ein Ziel, an dem das Schema etwas beanstandet hat.
function einBatchMitBefund(): Promise<KIBatchAntwort> {
  return Promise.resolve({
    kennung: 'kibatch-1',
    anbieter: 'ollama',
    ziele: [
      {
        zielkarte: 1,
        kennung: 'a1',
        modell: 'gemma4:8b',
        felder: [{ feld: 'name', werte: ['Donnerschlag'], bisher: '' }],
        befunde: ['das Feld »farbe« gibt es in dieser Vorlage nicht'],
        ablauf: '2026-09-13T14:00:00Z',
        fehler: null,
      },
    ],
    fehler: null,
  })
}

// aufklappen stellt den Kasten hin und klappt ihn auf.
async function aufklappen() {
  const hand = userEvent.setup()

  render(
    <Sprachraum sprache="de">
      <KIBatchkasten setID="7" geaendert={() => undefined} />
    </Sprachraum>,
  )

  await hand.click(screen.getByText(eigene.kiBatchTitel))

  return hand
}

// fuellen wählt eine Karte, ein Modell und tippt den Auftrag.
async function fuellen(hand: ReturnType<typeof userEvent.setup>) {
  await hand.click(await screen.findByLabelText('Blitzschlag'))
  await hand.selectOptions(
    screen.getByLabelText(eigene.kiVorschlagAnbieter),
    'ollama/gemma4:8b',
  )
  await hand.type(screen.getByLabelText(eigene.kiBatchAuftrag), 'Schreibe.')
}

// **Ein gescheiterter Batch meldet den Grund — und zeigt keine Lagen.**
it('meldet, wenn der Batch scheitert', async () => {
  belegeApi({
    Karten: zweiKarten,
    KIModelle: einModell('lokal'),
    KIBatchVorschlagen: () => Promise.reject(new Error('keine Laufzeit')),
  })

  const hand = await aufklappen()
  await fuellen(hand)
  await hand.click(screen.getByRole('button', { name: eigene.kiBatchFahren }))

  expect(await screen.findByRole('alert')).toHaveTextContent(/keine Laufzeit/)
  expect(screen.queryByText(eigene.kiBatchLagen)).not.toBeInTheDocument()
})

// **Eine gescheiterte Gruppe meldet den Grund — und die Lagen bleiben.**
//
// **Das ist die Zusage**: nichts wurde geschrieben, und ein zweiter Versuch
// ist möglich. Ein Kasten, der die Lagen wegräumte, nähme ihn weg.
it('meldet, wenn die Gruppe scheitert', async () => {
  belegeApi({
    Karten: zweiKarten,
    KIModelle: einModell('lokal'),
    KIBatchVorschlagen: einBatchMitBefund,
    KIGruppeUebernehmen: () => Promise.reject(new Error('der Stand ist alt')),
  })

  const hand = await aufklappen()
  await fuellen(hand)
  await hand.click(screen.getByRole('button', { name: eigene.kiBatchFahren }))

  await hand.click(
    await screen.findByRole('button', { name: /Als Gruppe übernehmen/ }),
  )

  expect(await screen.findByRole('alert')).toHaveTextContent(
    /der Stand ist alt/,
  )
  expect(screen.getByText(eigene.kiBatchLagen)).toBeInTheDocument()
})

// **Ein Befund des Schemas steht neben dem Feld und nicht statt seiner.**
it('zeigt die Befunde eines Ziels', async () => {
  belegeApi({
    Karten: zweiKarten,
    KIModelle: einModell('lokal'),
    KIBatchVorschlagen: einBatchMitBefund,
  })

  const hand = await aufklappen()
  await fuellen(hand)
  await hand.click(screen.getByRole('button', { name: eigene.kiBatchFahren }))

  expect(
    await screen.findByText(/gibt es in dieser Vorlage nicht/),
  ).toBeInTheDocument()

  // **Und das Feld steht trotzdem da** — der Befund ersetzt es nicht.
  expect(screen.getByText('Donnerschlag')).toBeInTheDocument()
})

// **Ein Cloudmodell fragt vorher nach der Freigabe.**
//
// **Und schickt vorher nichts los**: der Zähler unten steht auf null, solange
// niemand freigegeben hat.
it('fragt vor einem Cloudbatch nach der Freigabe', async () => {
  const batch = vi.fn(einBatchMitBefund)

  belegeApi({
    Karten: zweiKarten,
    KIModelle: einModell('cloud'),
    Kostenstand: () =>
      Promise.resolve({
        konto: 'openai',
        zeitraum: 'tag',
        grenze: { wert: 0, waehrung: 'EUR' },
        reserviert: { wert: 0, waehrung: 'EUR' },
        verbraucht: { wert: 0, waehrung: 'EUR' },
        unklar: { wert: 0, waehrung: 'EUR' },
        verfuegbar: { wert: 0, waehrung: 'EUR' },
        fehler: null,
      }),
    KIBatchVorschlagen: batch,
  })

  const hand = await aufklappen()

  await hand.click(await screen.findByLabelText('Blitzschlag'))
  await hand.selectOptions(
    screen.getByLabelText(eigene.kiVorschlagAnbieter),
    'openai/gemma4:8b',
  )
  await hand.type(screen.getByLabelText(eigene.kiBatchAuftrag), 'Schreibe.')
  await hand.click(screen.getByRole('button', { name: eigene.kiBatchFahren }))

  expect(await screen.findByText(eigene.kiFreigabeTitel)).toBeInTheDocument()
  expect(batch).not.toHaveBeenCalled()

  // **Wer verzichtet, kommt zum Knopf zurück** — und es ging nichts hinaus.
  await hand.click(
    screen.getByRole('button', { name: eigene.kiFreigabeVerzichten }),
  )

  expect(
    screen.getByRole('button', { name: eigene.kiBatchFahren }),
  ).toBeInTheDocument()
  expect(batch).not.toHaveBeenCalled()
})

// **Ein Set ohne Karten sagt es — und der Knopf bleibt gesperrt.**
it('sagt es, wenn das Set keine Karte hat', async () => {
  belegeApi({
    Karten: () => Promise.resolve({ karten: [], fehler: null }),
    KIModelle: einModell('lokal'),
  })

  await aufklappen()

  expect(await screen.findByText(eigene.kiBatchOhneKarten)).toBeInTheDocument()
  expect(
    screen.getByRole('button', { name: eigene.kiBatchFahren }),
  ).toBeDisabled()
})

// **Eine Karte ohne Namen bekommt trotzdem eine Zeile.**
it('nennt eine namenlose Karte beim Rückfall', async () => {
  belegeApi({ Karten: zweiKarten, KIModelle: einModell('lokal') })

  await aufklappen()

  expect(
    await screen.findByLabelText(eigene.kiBatchOhneName),
  ).toBeInTheDocument()
})

// **Ein Haken lässt sich wieder wegnehmen.**
it('nimmt eine Karte aus der Auswahl', async () => {
  belegeApi({ Karten: zweiKarten, KIModelle: einModell('lokal') })

  const hand = await aufklappen()
  const kaestchen = await screen.findByLabelText('Blitzschlag')

  await hand.click(kaestchen)
  expect(screen.getByText('1 von 2 Karten gewählt')).toBeInTheDocument()

  await hand.click(kaestchen)
  expect(screen.getByText('0 von 2 Karten gewählt')).toBeInTheDocument()
})

// **Und ein Lesefehler der Liste steht als Meldung da.**
it('meldet, wenn die Kartenliste nicht kommt', async () => {
  belegeApi({
    Karten: () => Promise.reject(new Error('der Bestand ist zu')),
    KIModelle: einModell('lokal'),
  })

  await aufklappen()

  expect(await screen.findByRole('alert')).toHaveTextContent(
    /der Bestand ist zu/,
  )
})

// **Ein Haken lässt sich auch wieder setzen.**
//
// **Der Weg zurück ist ein eigener Zweig**: abhaken nimmt aus der Liste,
// anhaken hängt an — und nur einer von beiden wäre halb gemessen.
it('setzt einen weggenommenen Feldhaken wieder', async () => {
  const gruppe = vi.fn(() =>
    Promise.resolve({ karten: 1, felder: 1, namen: [], fehler: null }),
  )

  belegeApi({
    Karten: zweiKarten,
    KIModelle: einModell('lokal'),
    KIBatchVorschlagen: einBatchMitBefund,
    KIGruppeUebernehmen: gruppe,
  })

  const hand = await aufklappen()
  await fuellen(hand)
  await hand.click(screen.getByRole('button', { name: eigene.kiBatchFahren }))

  const haken = await screen.findByLabelText(/name/)

  await hand.click(haken)
  expect(
    screen.getByRole('button', { name: /Als Gruppe übernehmen/ }),
  ).toBeDisabled()

  await hand.click(haken)
  await hand.click(
    screen.getByRole('button', { name: /Als Gruppe übernehmen/ }),
  )

  expect(gruppe).toHaveBeenCalledWith([{ kennung: 'a1', felder: ['name'] }])
})

// **Wer freigibt, schickt denselben Auftrag los, den er ausgelöst hat.**
it('fährt den Batch nach der Freigabe', async () => {
  const batch = vi.fn(einBatchMitBefund)

  belegeApi({
    Karten: zweiKarten,
    KIModelle: einModell('cloud'),
    Kostenstand: () =>
      Promise.resolve({
        konto: 'openai',
        zeitraum: 'tag',
        grenze: { wert: 5000, waehrung: 'EUR' },
        reserviert: { wert: 0, waehrung: 'EUR' },
        verbraucht: { wert: 0, waehrung: 'EUR' },
        unklar: { wert: 0, waehrung: 'EUR' },
        verfuegbar: { wert: 5000, waehrung: 'EUR' },
        fehler: null,
      }),
    KIFreigabeErteilen: () =>
      Promise.resolve({
        kennung: 'nw-1',
        laeuftab: '2026-09-13T14:05:00Z',
        fehler: null,
      }),
    KIBatchVorschlagen: batch,
  })

  const hand = await aufklappen()

  await hand.click(await screen.findByLabelText('Blitzschlag'))
  await hand.selectOptions(
    screen.getByLabelText(eigene.kiVorschlagAnbieter),
    'openai/gemma4:8b',
  )
  await hand.type(screen.getByLabelText(eigene.kiBatchAuftrag), 'Schreibe.')
  await hand.click(screen.getByRole('button', { name: eigene.kiBatchFahren }))

  await hand.type(await screen.findByLabelText(eigene.kiFreigabeBetrag), '100')
  await hand.click(
    screen.getByRole('button', { name: eigene.kiFreigabeErteilen }),
  )

  expect(await screen.findByText(eigene.kiBatchLagen)).toBeInTheDocument()

  // **Der Nachweis geht mit** — sonst scheiterte der Ruf an der Kostenwache,
  // nachdem ein Mensch gerade freigegeben hat.
  expect(batch).toHaveBeenCalledWith(
    ['1'],
    'openai',
    'gemma4:8b',
    'Schreibe.',
    'nw-1',
  )
})
