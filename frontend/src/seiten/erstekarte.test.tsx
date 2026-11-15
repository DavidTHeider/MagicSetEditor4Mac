// **Ein geöffnetes Set schlägt seine erste Karte auf**
// (#668).
//
//	void SetWindow::onChangeSet() {
//	  …
//	  // only after setSet select a card
//	  FOR_EACH(p, panels) {
//	    p->selectFirstCard();
//	  }
//
// `src/gui/set/window.cpp`, Zeilen 305–320, weiter über
// `CardsPanel::selectFirstCard` (`src/gui/set/cards_panel.cpp` 488–491) und
// `ItemList::selectFirst` (`src/gui/control/item_list.cpp` 56–59), Fassung
// f9d9356 vom 2021-07-06.
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Api, KarteDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karten } from './Karten'
import { ausKarten } from '../test/kartenproben'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function karte(id: string, nummer: number, name: string): KarteDTO {
  return { id, nummer, werte: [{ feld: 'name', wert: name }] }
}

// **Ein `Karten`-Beleg wird auf die drei Wege der Seite projiziert** (#891).
//
// Die Seite holt seit dem 2026-08-25 `Kartenzeilen` für die Liste und `Karte`
// für die aufgeschlagene; hier steht weiter, welche Karten es **gibt**.
function zeige(teile: Partial<Api>, karteID?: string) {
  belegeApi({
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Kartenfelder: () => Promise.resolve({ felder: [] }),
    Vorschau: () => Promise.resolve({ vorschau: null }),
    Probevorschau: () => Promise.resolve({ vorschau: null }),
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),
    ...(teile.Karten ? ausKarten(teile.Karten, '7') : {}),
    ...teile,
  })

  return render(
    <Sprachraum sprache="de">
      <Karten
        karteGewaehlt={vi.fn()}
        setID="7"
        karteID={karteID}
        verlaufsstand={0}
        verlaufsfehler={undefined}
        bestandGeaendert={vi.fn()}
      />
    </Sprachraum>,
  )
}

describe('Ein geöffnetes Set', () => {
  it('schlägt die erste Karte auf', async () => {
    zeige({
      Karten: () =>
        Promise.resolve({
          karten: [karte('8', 0, 'Erste'), karte('9', 1, 'Zweite')],
        }),
    })

    expect(await screen.findByDisplayValue('Erste')).toBeInTheDocument()
  })

  // **Wer eine Karte mitgibt, bekommt sie** — nicht die erste. Sonst spränge
  // ein Tiefenverweis auf die falsche Karte.
  it('lässt eine mitgegebene Karte stehen', async () => {
    zeige(
      {
        Karten: () =>
          Promise.resolve({
            karten: [karte('8', 0, 'Erste'), karte('9', 1, 'Zweite')],
          }),
      },
      '9',
    )

    expect(await screen.findByDisplayValue('Zweite')).toBeInTheDocument()
  })

  // **Ein leeres Set behält seinen leeren Zustand** (§6).
  //
  // MSE2 legt in diesem Fall stillschweigend eine Karte an — *„make sure there
  // is always at least one card; some things need this"*
  // (`src/gui/set/window.cpp` 305–312). Das wird **nicht** übernommen: hier
  // sagt der leere Zustand, was als Nächstes zu tun ist, statt eine Karte
  // anzulegen, die niemand wollte.
  it('zeigt bei einem leeren Set weiter seinen leeren Zustand', async () => {
    zeige({ Karten: () => Promise.resolve({ karten: [] }) })

    expect(
      await screen.findByText(texte.seiten.karten.leerTitel),
    ).toBeInTheDocument()
  })
})
