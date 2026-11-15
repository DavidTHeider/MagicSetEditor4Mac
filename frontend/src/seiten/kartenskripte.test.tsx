// **Karten aus einem Skript, im Fenster** — Werkstück B von Phase 4.11
// (#1337).
//
// # Was MSE2 hier tut
//
// Es zeigt die Skripte als Aufklappmenü an „Karte hinzufügen"
// (`src/gui/set/cards_panel.cpp`, Zeilen 160–176, Fassung f9d9356 vom
// 2021-07-06). Hier stehen sie als eigener Kasten — ein Menü über einem Knopf
// wäre ein Menü in einem Menü, und §6 verbietet verschachtelte Dialoge.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Karten } from './Karten'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { ausKarten } from '../test/kartenproben'
import { belegeApi, raeumeApi } from '../test/api'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten

// zwei sind zwei Skripte, wie ein Spielpaket sie mitbringt.
//
// **Das zweite ohne Beschreibung.** `description:` ist in MSE2 optional —
// `IMPLEMENT_REFLECTION_NO_SCRIPT(AddCardsScript)` liest den Schlüssel, wenn
// er dasteht, und lässt das Feld sonst leer
// (`src/data/add_cards_script.cpp`, Zeilen 18–23, Fassung f9d9356 vom
// 2021-07-06); das Menü hängt sie als Hilfetext an und verträgt einen leeren
// (`src/gui/set/cards_panel.cpp`, Zeile 174).
//
// **Gemessen am 2026-08-31**: in `full-magic-pack` trägt jeder der zwölf
// Blöcke eine Beschreibung — die Reichweite dieser Zahl ist dieses Pack und
// nicht „jedes Paket". Der Fall ohne bleibt trotzdem echt, weil ihn der Leser
// erzeugt, sobald der Schlüssel fehlt.
const zwei = [
  {
    name: 'Zwei Farben',
    kuerzel: 'Z',
    beschreibung: 'Legt je eine rote und eine blaue Karte an',
  },
  { name: 'Leerer Zug', kuerzel: '', beschreibung: '' },
]

function zeige(zusatz = {}) {
  belegeApi({
    ...ausKarten(() =>
      Promise.resolve({
        karten: [
          { id: '1', nummer: 0, werte: [{ feld: 'name', wert: 'Erste' }] },
        ],
      }),
    ),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Kartenskripte: () => Promise.resolve({ skripte: zwei, fehler: null }),
    ...zusatz,
  })

  render(
    <Sprachraum sprache="de">
      <Karten
        setID="1"
        karteID={undefined}
        karteGewaehlt={vi.fn()}
        bestandGeaendert={vi.fn()}
        verlaufsstand={0}
        verlaufsfehler={undefined}
      />
    </Sprachraum>,
  )
}

// aufklappen öffnet den Skriptkasten.
//
// **Seit dem 2026-09-03 ist er zu, wenn die Lasche aufgeht** — Werkstück C
// von Phase 4.10: dreizehn Knöpfe schoben die Kartenliste unter die Falz.
// Was diese Tests prüfen, steht darin und ändert sich nicht; erreicht wird es
// über einen Klick mehr.
//
// **Seit UXR-05 (#1913) über das Menü „Mehr"** neben „Karte anlegen" statt
// über einen eigenen Aufklapper — die Zeile über dem Arbeitsplatz trägt eine
// Handlung und ein Menü.
async function aufklappen() {
  await userEvent.click(await screen.findByRole('button', { name: texte.mehr }))
  await userEvent.click(
    await screen.findByRole('menuitem', { name: texte.kartenskripteTitel }),
  )
}

// **Die Skripte stehen als Knöpfe da, mit Namen und Beschreibung.**
it('zeigt die Kartenskripte des Spielpakets', async () => {
  zeige()
  await aufklappen()

  expect(
    await screen.findByRole('button', {
      name: 'Zwei Farben — Legt je eine rote und eine blaue Karte an',
    }),
  ).toBeInTheDocument()

  // **Auch eines ohne Beschreibung** — dann steht nur der Name da, ohne den
  // Gedankenstrich, der ins Leere zeigte.
  expect(screen.getByRole('button', { name: 'Leerer Zug' })).toBeInTheDocument()
})

// **Ein Klick führt das Skript und sagt, was entstanden ist.**
it('führt ein Skript aus und nennt die Zahl', async () => {
  const KartenskriptAusfuehren = vi.fn(() =>
    Promise.resolve({ karten: ['2', '3'], befunde: [], fehler: null }),
  )

  zeige({ KartenskriptAusfuehren })
  await aufklappen()

  await userEvent.click(
    await screen.findByRole('button', {
      name: 'Zwei Farben — Legt je eine rote und eine blaue Karte an',
    }),
  )

  await waitFor(() => {
    expect(KartenskriptAusfuehren).toHaveBeenCalledWith('1', 'Zwei Farben')
  })

  // **Eine Zahl und kein „fertig".** Wer zwölf Grundlande erwartet und zwei
  // bekommt, soll es sehen.
  expect(await screen.findByText('2 Karten angelegt.')).toBeInTheDocument()
})

// **Eine einzelne Karte bekommt den Singular.**
it('sagt bei einer Karte den Einzahlsatz', async () => {
  zeige({
    KartenskriptAusfuehren: () =>
      Promise.resolve({ karten: ['2'], befunde: [], fehler: null }),
  })
  await aufklappen()

  await userEvent.click(
    await screen.findByRole('button', { name: /^Leerer Zug/ }),
  )

  expect(await screen.findByText(texte.kartenskriptEine)).toBeInTheDocument()
})

// **Ein Fehler steht am Kasten und hält die Seite nicht an.**
it('zeigt einen Fehler des Laufs', async () => {
  zeige({
    KartenskriptAusfuehren: () =>
      Promise.resolve({
        karten: [],
        befunde: [],
        fehler: {
          code: 'nichtGefunden' as const,
          schluessel: '',
          meldung: 'dieses Spiel kennt kein Kartenskript "Zwei Farben"',
        },
      }),
  })
  await aufklappen()

  await userEvent.click(
    await screen.findByRole('button', { name: /^Zwei Farben/ }),
  )

  expect(await screen.findByText(/kennt kein Kartenskript/)).toBeInTheDocument()
})

// **Ohne Skripte steht der Kasten nicht da.**
//
// §6: leere Zustände erklären den nächsten Schritt, statt „Keine Daten" zu
// zeigen — und für ein Set ohne Spielpaket gibt es keinen nächsten Schritt in
// diesem Kasten.
it('zeigt keinen Kasten, wenn das Spielpaket keine Skripte hat', async () => {
  zeige({ Kartenskripte: () => Promise.resolve({ skripte: [], fehler: null }) })

  await screen.findByText('Erste')

  // **Kein Eintrag und kein Kasten.** Seit UXR-05 (#1913) öffnet das Menü
  // „Mehr" den Kasten; ein Eintrag über nichts wäre genau die leere Zeile,
  // die §6 verbietet. **Gefragt wird im offenen Menü** — ohne es fände die
  // Abfrage den Eintrag ohnehin nicht, und die Verneinung belegte nichts.
  await userEvent.click(screen.getByRole('button', { name: texte.mehr }))

  expect(
    await screen.findByRole('menuitem', { name: texte.umschlagAnnehmen }),
  ).toBeInTheDocument()
  expect(
    screen.queryByRole('menuitem', { name: texte.kartenskripteTitel }),
  ).not.toBeInTheDocument()
})
