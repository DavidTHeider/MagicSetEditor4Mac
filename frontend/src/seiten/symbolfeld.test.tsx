// **Ein Symbolfeld wird gezeichnet, nicht getippt** — Werkstück A von Phase
// 4.11 (#1339,
// #1349).
//
// # Was nachgelesen wurde (§15)
//
// `src/data/field/symbol.hpp`, Zeilen 63–70 (Fassung f9d9356 vom 2021-07-06):
// der Wert eines Symbolfeldes ist ein **Dateiname im Paket**. MSE2 öffnet
// dafür seinen Symboleditor (`src/gui/value/symbol.cpp`), nicht ein Textfeld.
//
// **Ein Dateiname ist nichts, was jemand tippt.** Er entsteht beim Ablegen —
// derselbe Grund, aus dem ein Bildfeld einen Knopf statt einer Eingabe
// bekommt.
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { FeldbeschreibungDTO, KarteDTO, VorschauAntwort } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karteneditor } from './Karteneditor'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function vorschau(): VorschauAntwort {
  return {
    vorschau: {
      stil: '',
      gezeichnet: '',
      bild: '/vorschau/AAAA.png',
      breite: 10,
      hoehe: 14,
      dpi: 150,
      befunde: [],
      flaechen: [],
      nennungen: [],
      vorgaben: [],
    },
  }
}

function form(art: string): FeldbeschreibungDTO {
  return {
    name: 'kartensymbol',
    aufschrift: 'Kartensymbol',
    art,
    mehrzeilig: false,
    wahlen: [],
    beschreibung: '',
    vorgabe: '',
    teilfelder: [],
    reihe: '',
    stelle: 0,
    editierbar: true,
    gespeichert: true,
  }
}

function karte(werte: KarteDTO['werte'] = []): KarteDTO {
  return { id: '5', nummer: 0, werte }
}

function zeige(art: string, aufSymbol = vi.fn()) {
  belegeApi({
    Vorschau: () => Promise.resolve(vorschau()),
    Probevorschau: () => Promise.resolve(vorschau()),
  })

  render(
    <Sprachraum sprache="de">
      <Karteneditor
        setID="1"
        karte={karte([{ feld: 'kartensymbol', wert: 'eigen.mse-symbol' }])}
        formen={[form(art)]}
        geaendert={() => undefined}
        symbolZeichnen={aufSymbol}
      />
    </Sprachraum>,
  )

  return aufSymbol
}

describe('Ein Symbolfeld', () => {
  it('bietet einen Knopf, der den Symboleditor öffnet', async () => {
    const aufSymbol = zeige('symbol')

    const knopf = await screen.findByRole('button', {
      name: `Kartensymbol — ${texte.seiten.karten.symbolZeichnen}`,
    })

    knopf.click()

    expect(aufSymbol).toHaveBeenCalledWith('kartensymbol')
  })

  // **Nur ein Symbolfeld.** Ein Textfeld bekommt keinen — der Knopf öffnet
  // eine Zeichenfläche, und in ein Textfeld gehört keine.
  it('gibt einem Textfeld keinen', async () => {
    zeige('text')

    await screen.findByLabelText('kartensymbol')

    expect(
      screen.queryByRole('button', {
        name: `Kartensymbol — ${texte.seiten.karten.symbolZeichnen}`,
      }),
    ).not.toBeInTheDocument()
  })
})
