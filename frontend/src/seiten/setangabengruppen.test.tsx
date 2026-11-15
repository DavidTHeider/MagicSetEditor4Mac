// **Die Setangaben in Gruppen, eine offen** — UXR-07
// (#1915).
//
// # Wofür
//
// `magic.mse-game` erklärt **89** Setfelder, darunter zehn Zwischenüberschriften
// (`type: info` — „Set Information“, „Language“, „Card Numbering“ …; gezählt
// am 2026-10-02 in `set_fields` des Full-Magic-Pack). Die Lasche zeichnete sie
// alle untereinander, jede Überschrift als leeres, schreibgeschütztes Feld —
// R3 aus `docs/UIREVIEW.md` erlaubt zwölf, und die Messung am Probespiel sah
// es nicht, weil `probe.mse-game` zwei Felder führt.
//
// # Was MSE tut
//
// Es zeichnet eine Auskunft als **Kopfzeile** über die Breite des Editors:
// `InfoValueViewer::draw` füllt den Grund und setzt den Wert, und der ist
// ohne Skript die Aufschrift (`InfoValue::update`,
// `src/data/field/information.cpp`; `src/render/value/information.cpp`,
// Fassung f9d9356 vom 2021-07-06). Eingeklappt wird dort nichts — das ist
// **adapted**, nach R3: die erste Gruppe offen, jede weitere hinter ihrer
// Überschrift.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { FeldbeschreibungDTO } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { verstoesse } from '../test/axe'
import { leererSetkopf } from '../test/setkopfprobe'
import { Setangaben } from './Setangaben'

afterEach(() => {
  raeumeApi()
})

function form(
  name: string,
  teile: Partial<FeldbeschreibungDTO> = {},
): FeldbeschreibungDTO {
  return {
    name,
    aufschrift: name,
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
    ...teile,
  }
}

const auskunft = (name: string) =>
  form(name, { art: 'info', editierbar: false, gespeichert: false })

function zeige(felder: FeldbeschreibungDTO[], SetwertSetzen = vi.fn()) {
  belegeApi({
    Vorbehalte: () => Promise.resolve({ vorbehalte: [], fehler: null }),
    Setangaben: () =>
      Promise.resolve({
        felder,
        werte: [{ feld: 'card number style', wert: 'alt' }],
        kopf: leererSetkopf,
        fehler: null,
      }),
    SetwertSetzen,
  })

  render(
    <Sprachraum sprache="de">
      <Setangaben setID="1" verlaufsstand={0} />
    </Sprachraum>,
  )

  return SetwertSetzen
}

// wieMagic hat die Gestalt von `magic.mse-game`: eine Überschrift vorn, und
// hinter jeder weiteren ihre Felder.
const wieMagic = [
  auskunft('Set Information'),
  form('title'),
  form('set code'),
  auskunft('Card Numbering'),
  form('card number style'),
  form('last main partition'),
  auskunft('Card Stamps'),
  form('default stamp'),
]

describe('Die Setangaben in Gruppen', () => {
  it('zeigt die erste Gruppe offen, unter ihrer Überschrift', async () => {
    zeige(wieMagic)

    expect(
      await screen.findByRole('heading', { name: 'Set Information' }),
    ).toBeInTheDocument()
    expect(screen.getByLabelText('title')).toBeVisible()
    expect(screen.getByLabelText('set code')).toBeVisible()
  })

  // **Eine Überschrift ist kein Feld** — sie stand als leeres,
  // schreibgeschütztes Eingabefeld da.
  it('zeichnet keine Überschrift als Feld', async () => {
    zeige(wieMagic)

    await screen.findByLabelText('title')

    expect(screen.queryByLabelText('Set Information')).not.toBeInTheDocument()
    expect(screen.getAllByRole('textbox')).toHaveLength(2)
  })

  it('legt jede weitere Gruppe hinter ihre Überschrift', async () => {
    zeige(wieMagic)

    const nummern = await screen.findByRole('button', {
      name: 'Card Numbering',
    })
    expect(nummern).toHaveAttribute('aria-expanded', 'false')
    expect(screen.getByRole('button', { name: 'Card Stamps' })).toHaveAttribute(
      'aria-expanded',
      'false',
    )
    expect(screen.queryByLabelText('card number style')).not.toBeInTheDocument()
  })

  // **Eingeklappt, nicht entfernt**: aufgeklappt steht das Feld mit seinem
  // Wert da und schreibt wie jedes andere.
  it('schreibt ein Feld einer aufgeklappten Gruppe', async () => {
    const SetwertSetzen = zeige(
      wieMagic,
      vi.fn(() => Promise.resolve({ fehler: null })),
    )

    await userEvent.click(
      await screen.findByRole('button', { name: 'Card Numbering' }),
    )
    const feld = screen.getByLabelText('card number style')
    expect(feld).toHaveValue('alt')

    await userEvent.clear(feld)
    await userEvent.type(feld, 'neu')
    await userEvent.tab()

    expect(SetwertSetzen).toHaveBeenCalledWith(
      '1',
      'card number style',
      'neu',
      expect.any(Array),
    )
  })

  // **Ohne Überschrift vorn** bilden die Felder davor die erste Gruppe — und
  // sie bleibt offen, ohne eine Überschrift zu erfinden.
  it('lässt Felder vor der ersten Überschrift offen und ohne Kopf', async () => {
    zeige([
      form('title'),
      auskunft('Card Numbering'),
      form('last main partition'),
    ])

    expect(await screen.findByLabelText('title')).toBeVisible()
    expect(screen.queryByRole('heading', { level: 3 })).not.toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Card Numbering' }),
    ).toHaveAttribute('aria-expanded', 'false')
  })

  // **Ein Fehler erscheint am verursachenden Feld** (§6) — auch wenn es in
  // einer eingeklappten Gruppe steht: die öffnet sich dann von selbst.
  it('öffnet die Gruppe des Feldes, das ein Fehler nennt', async () => {
    zeige(
      wieMagic,
      vi.fn(() =>
        Promise.resolve({
          fehler: {
            code: 'validierung' as const,
            schluessel: 'fehler.validierung',
            meldung: 'Diesen Stempel gibt es nicht.',
            feld: 'default stamp',
          },
        }),
      ),
    )

    const titel = await screen.findByLabelText('title')
    await userEvent.type(titel, 'x')
    await userEvent.tab()

    expect(
      await screen.findByText('Diesen Stempel gibt es nicht.'),
    ).toBeVisible()
    expect(screen.getByRole('button', { name: 'Card Stamps' })).toHaveAttribute(
      'aria-expanded',
      'true',
    )
    expect(screen.getByLabelText('default stamp')).toHaveAttribute(
      'aria-invalid',
      'true',
    )
  })

  it('hält die Regeln der Barrierefreiheit', async () => {
    zeige(wieMagic)

    await userEvent.click(
      await screen.findByRole('button', { name: 'Card Numbering' }),
    )
    await screen.findByLabelText('card number style')

    expect(await verstoesse(document.body)).toEqual([])
  })
})
