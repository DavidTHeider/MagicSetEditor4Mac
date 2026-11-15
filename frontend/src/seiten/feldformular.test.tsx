// **Wie die Felder einer Karte angeordnet sind** —
// #1271.
import { render, screen, within } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import type { FeldbeschreibungDTO } from '../api'
import { Feldformular } from './Feldformular'
import { Sprachraum } from '../sprache/kontext'

// form gibt eine Feldbeschreibung, wie das Spielpaket sie liefert.
function form(name: string): FeldbeschreibungDTO {
  return {
    name,
    aufschrift: name,
    art: 'text',
    mehrzeilig: false,
    wahlen: [],
    beschreibung: '',
    vorgabe: '',
    editierbar: true,
    gespeichert: true,
    teilfelder: [],
    reihe: '',
    stelle: 0,
  }
}

function zeige(felder: string[], erklaert: string[]) {
  const nachName = new Map(erklaert.map((name) => [name, form(name)]))

  render(
    <Sprachraum sprache="de">
      <Feldformular
        felder={felder.map((feld) => ({ feld, wert: '' }))}
        nachName={nachName}
        tiefe="experte"
        probleme={[]}
        aufTiefe={() => undefined}
        formular={{ current: null }}
        zeile={(feld) => <p key={feld.feld}>{feld.feld}</p>}
      />
    </Sprachraum>,
  )
}

// **Was das Spielpaket erklärt, steht oben** — und was es nicht erklärt, in
// einem eigenen Block darunter.
//
// In MSE2 sind `stylesheet`, `styling_data` und `has_styling` **Mitglieder von
// `Card`** (`src/data/card.hpp`, Zeilen 44–50, Fassung f9d9356 vom
// 2021-07-06) und keine `card field:` des Spiels.
//
// # Hier stand ein `<details>`, und das ist seit UX-05 weg
//
// **Der Ausklapper war der halbe Weg zu drei Tiefen**
// (#1677):
// eine Anzeigefrage, die der Browser beantwortet, mit genau zwei Antworten.
// Seit es den Modus gibt, entscheidet **er**, was dasteht — und was in einer
// Tiefe nicht steht, steht auch nicht zugeklappt im Baum. Das ist mehr als
// eine Umbenennung: ein Vorleseprogramm ging vorher durch Felder, die
// niemand sehen wollte.
it('stellt die Buchführung in einen eigenen Block', () => {
  zeige(
    ['name', 'rule text', 'stylesheet', 'styling data'],
    ['name', 'rule text'],
  )

  const block = screen.getByRole('region', {
    name: 'Was die Anwendung dazu weiß',
  })

  expect(within(block).getByText('stylesheet')).toBeInTheDocument()
  expect(within(block).getByText('styling data')).toBeInTheDocument()

  // **Der Karteninhalt steht dagegen außerhalb.**
  expect(within(block).queryByText('name')).not.toBeInTheDocument()
  expect(screen.getByText('name')).toBeInTheDocument()
  expect(screen.getByText('rule text')).toBeInTheDocument()
})

// **Ohne solche Felder steht gar nichts da.**
//
// Ein leerer Block wäre eine Überschrift, die erklärt werden müsste.
it('lässt den Block weg, wenn alles erklärt ist', () => {
  zeige(['name', 'rule text'], ['name', 'rule text'])

  expect(
    screen.queryByText('Was die Anwendung dazu weiß'),
  ).not.toBeInTheDocument()
})

// **Die Reihenfolge bleibt die, die hereinkommt** — sie ist die Leseordnung
// der Karte und in `angebotenefelder.ts` begründet.
it('lässt die Reihenfolge, wie sie kommt', () => {
  zeige(['zuletzt', 'zuerst'], ['zuletzt', 'zuerst'])

  const absaetze = screen.getAllByText(/zu/)

  expect(absaetze[0]).toHaveTextContent('zuletzt')
})

// **Und der Rückruf bekommt jedes Feld genau einmal.**
it('zeichnet jedes Feld einmal', () => {
  const zeile = vi.fn((feld: { feld: string }) => <p key={feld.feld}>x</p>)

  render(
    <Sprachraum sprache="de">
      <Feldformular
        felder={[
          { feld: 'name', wert: '' },
          { feld: 'stylesheet', wert: '' },
        ]}
        nachName={new Map([['name', form('name')]])}
        tiefe="experte"
        probleme={[]}
        aufTiefe={() => undefined}
        formular={{ current: null }}
        zeile={zeile}
      />
    </Sprachraum>,
  )

  expect(zeile).toHaveBeenCalledTimes(2)
})
