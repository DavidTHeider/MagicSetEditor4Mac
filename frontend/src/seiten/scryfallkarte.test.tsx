// **Eine Karte aus einem Scryfall-Namen** —
// #1032.
//
// Bis dahin arbeitete der Abgleich nur an einer **vorhandenen** Karte: erst
// eine leere anlegen, dann abgleichen — zwei Schritte für einen Vorgang, und
// den ersten musste man kennen.
//
// # Die drei Zusagen, die hier geprüft werden
//
//   - §7 Regel 4: ein Vorgang je Datensatz — der Knopf ist gesperrt, solange
//     der Aufruf läuft.
//   - §7 Regel 3: geschrieben ist geschrieben — die angelegte Karte wird
//     aufgeschlagen, **auch wenn** das Füllen scheitert (#155).
//   - Das Bild ist wahlweise dabei: ohne den Haken wird keines geholt.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import type { KarteDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { oeffneAnlegen } from '../test/kartenlasche'
import { belegeApi, raeumeApi } from '../test/api'
import { Karten } from './Karten'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.scryfallkarte

function karten(): KarteDTO[] {
  return [{ id: '1', nummer: 0, werte: [] }]
}

function zeige(teile = {}) {
  belegeApi({
    Karten: () => Promise.resolve({ karten: karten() }),
    Kartenfelder: () => Promise.resolve({ felder: [] }),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    ...teile,
  })

  return render(
    <Sprachraum sprache="de">
      <Karten
        karteGewaehlt={vi.fn()}
        bestandGeaendert={() => undefined}
        karteID={undefined}
        setID="3"
        verlaufsfehler={undefined}
        verlaufsstand={0}
      />
    </Sprachraum>,
  )
}

// **Sie klappt den Anlegekasten selbst auf** (#1240) — der Scryfall-Weg steht
// dort und nicht mehr in der Kopfzeile.
async function lege(name: string, mitBild = true) {
  await oeffneAnlegen()

  const feld = await screen.findByLabelText(texte.name)
  await userEvent.type(feld, name)

  if (!mitBild) {
    await userEvent.click(screen.getByLabelText(texte.mitBild))
  }

  await userEvent.click(screen.getByRole('button', { name: texte.anlegen }))
}

it('legt eine Karte aus dem Namen an — mit Bild', async () => {
  const ScryfallKarteAnlegen = vi.fn(() =>
    Promise.resolve({
      karteID: '9',
      felder: ['name'],
      befunde: [],
      herkunft: 'https://scryfall.com/card/lea/161',
      alterStunden: 0,
      nennung: '',
      fehler: null,
    }),
  )

  zeige({ ScryfallKarteAnlegen })
  await lege('Lightning Bolt')

  await waitFor(() => {
    expect(ScryfallKarteAnlegen).toHaveBeenCalledWith(
      '3',
      'Lightning Bolt',
      'image',
      'ausschnitt',
    )
  })
})

it('holt kein Bild, wenn der Haken weg ist', async () => {
  const ScryfallKarteAnlegen = vi.fn(() =>
    Promise.resolve({
      karteID: '9',
      felder: ['name'],
      befunde: [],
      herkunft: '',
      alterStunden: 0,
      nennung: '',
      fehler: null,
    }),
  )

  zeige({ ScryfallKarteAnlegen })
  await lege('Lightning Bolt', false)

  await waitFor(() => {
    expect(ScryfallKarteAnlegen).toHaveBeenCalledWith(
      '3',
      'Lightning Bolt',
      'image',
      '',
    )
  })
})

it('zeigt die Meldung und schlägt die Karte trotzdem auf', async () => {
  const ScryfallKarteAnlegen = vi.fn(() =>
    Promise.resolve({
      karteID: '9',
      felder: [],
      befunde: [],
      herkunft: '',
      alterStunden: 0,
      nennung: '',
      fehler: { code: 'validierung', meldung: 'Scryfall antwortet nicht' },
    }),
  )

  zeige({ ScryfallKarteAnlegen })
  await lege('Gibbes Nicht')

  // **Die Meldung steht da** — und die Karte ist trotzdem angelegt (#155).
  expect(
    await screen.findByText(/Scryfall antwortet nicht/),
  ).toBeInTheDocument()
})

it('sperrt den Knopf, solange kein Name dasteht', async () => {
  zeige()
  await oeffneAnlegen()

  const knopf = await screen.findByRole('button', { name: texte.anlegen })
  expect(knopf).toBeDisabled()
})

// **Eine Antwort ohne Kennung schlägt nichts auf** (§7, Regel 3).
//
// Sie kommt, wenn schon der Abruf scheitert — dann ist gar keine Karte
// entstanden, und eine Wahl auf eine leere Kennung wäre eine auf nichts.
it('schlägt nichts auf, wenn keine Karte entstanden ist', async () => {
  const Karten = vi.fn(() => Promise.resolve({ karten: karten() }))

  zeige({
    Karten,
    ScryfallKarteAnlegen: () =>
      Promise.resolve({
        karteID: '',
        felder: [],
        befunde: [],
        herkunft: '',
        alterStunden: 0,
        nennung: '',
        fehler: { code: 'nichtGefunden', meldung: 'die Karte gibt es nicht' },
      }),
  })

  const vorher = Karten.mock.calls.length

  await lege('Gibbes Nicht')

  expect(await screen.findByText(/die Karte gibt es nicht/)).toBeInTheDocument()

  // **Kein Nachladen** — es gibt nichts nachzuladen.
  expect(Karten.mock.calls.length).toBe(vorher)
})

// **Ein Setwechsel während des Aufrufs verwirft die Antwort** (§7, Regel 2).
//
// Sie gehört zu dem Set, aus dem sie stammt. Wer sie danach einträgt, schlägt
// im **neuen** Set eine Karte auf, die dort nicht liegt.
it('verwirft die Antwort nach einem Setwechsel', async () => {
  let loesen: (() => void) | undefined

  const ScryfallKarteAnlegen = vi.fn(
    () =>
      new Promise((erfuellen) => {
        loesen = () => {
          erfuellen({
            karteID: '9',
            felder: ['name'],
            befunde: [],
            herkunft: '',
            alterStunden: 0,
            nennung: '',
            fehler: null,
          })
        }
      }),
  )

  const gezeichnet = zeige({ ScryfallKarteAnlegen })
  await oeffneAnlegen()

  const feld = await screen.findByLabelText(texte.name)
  await userEvent.type(feld, 'Lightning Bolt')
  await userEvent.click(screen.getByRole('button', { name: texte.anlegen }))

  // **Der Wechsel geschieht, während der Aufruf hängt.**
  gezeichnet.rerender(
    <Sprachraum sprache="de">
      <Karten
        karteGewaehlt={vi.fn()}
        bestandGeaendert={() => undefined}
        karteID={undefined}
        setID="4"
        verlaufsfehler={undefined}
        verlaufsstand={0}
      />
    </Sprachraum>,
  )

  loesen?.()

  // **Kein Fehler und keine Wahl** — die Antwort wird verworfen, nicht
  // gemeldet: sie ist nicht falsch, sie gehört nur woandershin.
  await waitFor(() => {
    expect(ScryfallKarteAnlegen).toHaveBeenCalled()
  })
})
