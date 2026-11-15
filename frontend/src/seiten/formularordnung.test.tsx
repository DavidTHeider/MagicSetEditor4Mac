// **Die Ordnung des Formulars bleibt beim Blättern stehen** — #2309, Folge
// aus dem Nachlauf über J05 (#2269), Befund 7.
//
// # Was gemessen ist
//
// Im Kanon ändert sich die Leseordnung zweier aufeinanderfolgender Karten
// desselben Stils in **601 von 1.033** Schritten auch für die Felder, die
// beide zeichnen (`internal/app/feldordnungkanon_intern_test.go`), und die
// eigene Vorschau braucht beim Blättern im Median **129 ms**, im P95 150 ms
// (`docs/belege/huellenprobe/blaettern.md`). Übernehmen hieß: in zwei von
// drei Schritten sortiert sich das Formular nach 130 ms um. Zurückhalten
// hieß: bei jedem Schritt 130 ms ein leeres Formular.
//
// **Am nativen Fenster nachgemessen** (`blaettern`, »Black«, 30 Schritte mit
// ↓): vorher sortierte sich das Formular in 21 von 30 Schritten um, nachher
// in keinem, und unsichtbar stand es in keinem.
//
// # Was gilt
//
// **Eine Ordnung je Stil**: die Lage, in der ein Feld in diesem Stil zuerst
// gesehen wurde, behält seinen Platz. Ein Feld, das eine Karte neu bringt,
// reiht sich nach seiner eigenen Lage ein.
import { act, render, screen } from '@testing-library/react'
import { useState } from 'react'
import { afterEach, expect, it } from 'vitest'

import type { FeldbeschreibungDTO, KarteDTO, VorschauAntwort } from '../api'
import type { FeldflaecheDTO } from '../formen'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karteneditor } from './Karteneditor'
import { grundansicht, type Vorschauansicht } from './vorschauansicht'

afterEach(() => {
  raeumeApi()
})

function form(name: string): FeldbeschreibungDTO {
  return {
    name,
    art: 'text',
    mehrzeilig: false,
    wahlen: [],
    beschreibung: '',
    vorgabe: '',
    teilfelder: [],
    reihe: '',
    stelle: 0,
    editierbar: true,
    gespeichert: true,
    aufschrift: name,
  }
}

const formen = ['name', 'staerke', 'kosten', 'text'].map(form)

function flaeche(feld: string, oben: number): FeldflaecheDTO {
  return { feld, links: 0, oben, breite: 10, hoehe: 5 }
}

function antwortMit(flaechen: FeldflaecheDTO[]): VorschauAntwort {
  return {
    vorschau: {
      stil: '',
      gezeichnet: '',
      bild: '/vorschau/BBBB.png',
      breite: 10,
      hoehe: 14,
      dpi: 150,
      befunde: [],
      flaechen,
      nennungen: [],
      vorgaben: [],
    },
  }
}

function karte(id: string): KarteDTO {
  return {
    id,
    nummer: 1,
    fassung: 1,
    werte: [{ feld: 'name', wert: `Karte ${id}` }],
  } as unknown as KarteDTO
}

function Seite({ anfang }: { anfang: Vorschauansicht }) {
  const ansicht = useState(anfang)

  return (
    <Sprachraum sprache="de">
      <Karteneditor
        setID="1"
        karte={karte('2')}
        formen={formen}
        geaendert={() => undefined}
        ansicht={ansicht}
      />
    </Sprachraum>
  )
}

function reihenfolge(): string[] {
  const formular = screen.getByRole('textbox', { name: 'name' }).closest('form')

  return [...(formular?.querySelectorAll('[aria-label]') ?? [])]
    .map((e) => e.getAttribute('aria-label') ?? '')
    .filter((name) => formen.some((f) => f.name === name))
}

// bekannt ist die Lage, in der der Stil seine Felder zuerst zeigte.
const bekannt = [flaeche('name', 0), flaeche('kosten', 10), flaeche('text', 20)]

it('behält die Ordnung des Stils, wenn die nächste Karte ihre Felder anders legt', async () => {
  // **Die nächste Karte legt den Text über die Kosten.**
  belegeApi({
    Vorschau: () =>
      Promise.resolve(
        antwortMit([
          flaeche('name', 0),
          flaeche('text', 8),
          flaeche('kosten', 16),
        ]),
      ),
  })

  render(
    <Seite anfang={{ ...grundansicht, flaechenJeStil: { '': bekannt } }} />,
  )

  expect(reihenfolge()).toEqual(['name', 'kosten', 'text'])

  await act(async () => {
    await Promise.resolve()
  })

  expect(reihenfolge()).toEqual(['name', 'kosten', 'text'])
})

it('reiht ein Feld, das der Stil bisher nicht zeigte, nach seiner Lage ein', async () => {
  // **Die Stärke kommt mit dieser Karte dazu** — zwischen Name und Kosten.
  belegeApi({
    Vorschau: () =>
      Promise.resolve(
        antwortMit([
          flaeche('name', 0),
          flaeche('staerke', 5),
          flaeche('kosten', 10),
          flaeche('text', 20),
        ]),
      ),
  })

  render(
    <Seite anfang={{ ...grundansicht, flaechenJeStil: { '': bekannt } }} />,
  )

  await act(async () => {
    await Promise.resolve()
  })

  expect(reihenfolge()).toEqual(['name', 'staerke', 'kosten', 'text'])
})
