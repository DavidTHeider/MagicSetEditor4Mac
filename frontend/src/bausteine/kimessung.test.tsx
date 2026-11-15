// **Messen, was ein Modell kann** — Werkstück A9 von Phase 5.3
// (#1448).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { KIMessung } from './KIMessung'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.einstellungen

// zeige rendert den Kasten und gibt den Spion auf »gemeldet«.
//
// **`kostet` steht auf wahr**, denn der Anbieter ist OpenAI: eine Messung
// dort schickt einen Auftrag hinaus, der Geld kostet. Eine Vorrichtung, die
// ihn umsonst messen ließe, wäre großzügiger als der Betrieb (§3) — und genau
// so ist #1539
// entstanden.
function zeige(messbar: string[], gemessen: string[] = [], kostet = true) {
  const gemeldet = vi.fn()

  render(
    <Sprachraum sprache="de">
      <KIMessung
        anbieter="openai"
        kostet={kostet}
        messbar={messbar}
        gemessen={gemessen}
        gemeldet={gemeldet}
      />
    </Sprachraum>,
  )

  return gemeldet
}

// einBudget ist die Auskunft, mit der eine Freigabe möglich ist.
function einBudget() {
  const betrag = (wert: number) => ({ wert, waehrung: 'EUR' })

  return {
    konto: 'openai',
    zeitraum: 'tag',
    grenze: betrag(50_000),
    reserviert: betrag(0),
    verbraucht: betrag(0),
    unklar: betrag(0),
    verfuegbar: betrag(50_000),
    fehler: null,
  }
}

// misst drückt den Messknopf und gibt die Freigabe, die er verlangt.
//
// **Gewartet wird auf das Feld** (§3): der Kasten zeichnet sich sofort, sein
// Betragsfeld erst, wenn das Kostenjournal geantwortet hat.
async function misst(hand: ReturnType<typeof userEvent.setup>) {
  await hand.click(screen.getByRole('button', { name: eigene.kiMessen }))

  await hand.type(await screen.findByLabelText(eigene.kiFreigabeBetrag), '1200')
  await hand.click(
    screen.getByRole('button', { name: eigene.kiFreigabeErteilen }),
  )
}

// eineFreigabe ist die Antwort auf `KIFreigabeErteilen`.
function eineFreigabe() {
  return Promise.resolve({
    kennung: 'n-1',
    laeuftab: '2026-09-12T12:05:00Z',
    fehler: null,
  })
}

// eineProbe baut eine Antwort auf eine Messung.
function eineProbe() {
  return {
    proben: [
      {
        anbieter: 'openai',
        alias: 'gpt-probe',
        faehigkeit: 'textgenerierung',
        datum: '2026-09-12',
        beleg: 'gemessen',
        modell: 'gpt-probe-2026-09-01',
      },
      {
        anbieter: 'openai',
        alias: 'gpt-probe',
        faehigkeit: 'strukturiert',
        datum: '2026-09-12',
        beleg: 'gemessen',
        modell: 'gpt-probe-2026-09-01',
      },
    ],
    fehler: null,
  }
}

// **Ohne messbare und ohne gemessene Modelle steht der Kasten nicht da.**
//
// **Und das ist kein leerer Zustand, sondern gar keiner**: ein Anbieter, der
// seine Modelle beschreibt, hat hier nichts zu tun.
it('steht nicht da, wo es nichts zu messen gibt', () => {
  belegeApi({})
  zeige([])

  expect(screen.queryByText(eigene.kiMessbarTitel)).not.toBeInTheDocument()
})

// **Ein messbares Modell steht mit einem Knopf da.**
it('zeigt ein messbares Modell mit einem Knopf', () => {
  belegeApi({})
  zeige(['gpt-probe'])

  expect(screen.getByText('gpt-probe')).toBeInTheDocument()
  expect(
    screen.getByRole('button', { name: eigene.kiMessen }),
  ).toBeInTheDocument()
})

// **Eine geglückte Messung meldet, was jetzt belegt ist — und lädt nach.**
it('meldet eine geglückte Messung und lädt nach', async () => {
  const pruefen = vi.fn(() => Promise.resolve(eineProbe()))

  belegeApi({
    KIModellPruefen: pruefen,
    Kostenstand: () => Promise.resolve(einBudget()),
    KIFreigabeErteilen: eineFreigabe,
  })

  const gemeldet = zeige(['gpt-probe'])

  await misst(userEvent.setup())

  // **Gewartet wird auf die Meldung und nicht auf die Rolle** (§3): während
  // die Messung läuft, steht »Es wird gemessen …« als erstes `status` da —
  // ein `findByRole` darauf träfe schon vor der Antwort zu.
  expect(await screen.findByText(/strukturiert/)).toBeInTheDocument()

  // **Die Kennung reist mit** (#1539) — ohne sie weist die Kostenwache ab.
  expect(pruefen).toHaveBeenCalledWith('openai', 'gpt-probe', 'n-1')

  // **Erst melden, dann nachladen** (§7, Regel 3).
  expect(gemeldet).toHaveBeenCalled()
})

// **Eine gescheiterte Messung meldet den Grund — und lädt nicht nach.**
//
// **Regel 3 andersherum**: nach einem gescheiterten Schreiben wird gar nicht
// nachgeladen, sonst überschriebe ein Lesefehler die Meldung, die sagt, was
// zu tun ist.
it('meldet eine gescheiterte Messung und lädt nicht nach', async () => {
  belegeApi({
    KIModellPruefen: () => Promise.reject(new Error('das Budget ist leer')),
    Kostenstand: () => Promise.resolve(einBudget()),
    KIFreigabeErteilen: eineFreigabe,
  })

  const gemeldet = zeige(['gpt-probe'])

  await misst(userEvent.setup())

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'das Budget ist leer',
  )

  expect(gemeldet).not.toHaveBeenCalled()
})

// **Eine gemessene Fähigkeit lässt sich zurücknehmen.**
it('nimmt eine Messung zurück', async () => {
  const vergessen = vi.fn(() => Promise.resolve({ fehler: null }))

  belegeApi({ KIModellVergessen: vergessen })

  const gemeldet = zeige([], ['gpt-probe'])

  await userEvent
    .setup()
    .click(screen.getByRole('button', { name: eigene.kiMessungVergessen }))

  expect(vergessen).toHaveBeenCalledWith('openai', 'gpt-probe')
  expect(gemeldet).toHaveBeenCalled()
})

// **Und ein Fehler dabei steht da.**
it('meldet einen Fehler beim Zurücknehmen', async () => {
  belegeApi({
    KIModellVergessen: () => Promise.reject(new Error('kein Bestand offen')),
  })

  const gemeldet = zeige([], ['gpt-probe'])

  await userEvent
    .setup()
    .click(screen.getByRole('button', { name: eigene.kiMessungVergessen }))

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'kein Bestand offen',
  )

  expect(gemeldet).not.toHaveBeenCalled()
})

// **Ein zweiter Ruf, während der erste läuft, geht nicht** (§7, Regel 4).
//
// **Und hier kostet er zusätzlich Geld** — jede Messung ist eine Anfrage.
it('lässt keinen zweiten Ruf zu, solange der erste läuft', async () => {
  let freigeben: (() => void) | undefined

  const pruefen = vi.fn(
    () =>
      new Promise<ReturnType<typeof eineProbe>>((fertig) => {
        freigeben = () => {
          fertig(eineProbe())
        }
      }),
  )

  belegeApi({
    KIModellPruefen: pruefen,
    Kostenstand: () => Promise.resolve(einBudget()),
    KIFreigabeErteilen: eineFreigabe,
  })
  zeige(['gpt-probe'])

  const knopf = screen.getByRole('button', { name: eigene.kiMessen })

  await misst(userEvent.setup())

  expect(await screen.findByText(eigene.kiMessungLaeuft)).toBeInTheDocument()

  expect(knopf).toBeDisabled()

  freigeben?.()
})

// **Wo nichts kostet, wird auch nichts freigegeben.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil eine Freigabe, die überall verlangt wird, keine Auskunft ist.** Ein
// Modell auf diesem Rechner kostet kein Geld; wer dafür einen Höchstbetrag
// nennen müsste, gewöhnte sich das Wegklicken an — und dann gälte die Frage
// auch dort nichts mehr, wo sie zählt.
it('misst ohne Freigabe, wo nichts kostet', async () => {
  const pruefen = vi.fn(() => Promise.resolve(eineProbe()))
  const erteilen = vi.fn()

  belegeApi({ KIModellPruefen: pruefen, KIFreigabeErteilen: erteilen })

  zeige(['gemma4:8b'], [], false)

  await userEvent
    .setup()
    .click(screen.getByRole('button', { name: eigene.kiMessen }))

  // **Erst der Beleg, dass gemessen wurde** — die Verneinung darunter träfe
  // sonst auch zu, wenn der Knopf gar nichts ausgelöst hätte (§3).
  expect(await screen.findByText(/strukturiert/)).toBeInTheDocument()

  expect(pruefen).toHaveBeenCalledWith('openai', 'gemma4:8b', '')
  expect(erteilen).not.toHaveBeenCalled()
})

// **Wer nicht freigibt, misst nicht.**
it('misst nicht, wenn niemand freigibt', async () => {
  const pruefen = vi.fn()

  belegeApi({
    KIModellPruefen: pruefen,
    Kostenstand: () => Promise.resolve(einBudget()),
  })

  zeige(['gpt-probe'])

  const hand = userEvent.setup()

  await hand.click(screen.getByRole('button', { name: eigene.kiMessen }))

  await hand.click(
    await screen.findByRole('button', { name: eigene.kiFreigabeVerzichten }),
  )

  // **Erst der Beleg, dass der Kasten weg ist** — die Verneinung darunter
  // träfe sonst auch zu, wenn der Knopf nichts ausgelöst hätte (§3).
  await vi.waitFor(() => {
    expect(screen.queryByText(eigene.kiFreigabeTitel)).not.toBeInTheDocument()
  })

  expect(pruefen).not.toHaveBeenCalled()
})
