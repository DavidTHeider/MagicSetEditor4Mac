// **Was die Zahlenfelder der Symmetrie annehmen** — Werkstück U3 von Phase
// 5.1 (#1414).
//
// # Warum Mitte und Griff hier **ungeprüft** sind
//
// Weil es nichts zu prüfen gibt. `app.symmetrieAendern` setzt beide ohne
// jeden Vergleich (`internal/app/symbolsymmetrie.go`), und MSE2 tut dasselbe:
// `SymmetryMoveAction::move` (`src/data/action/symbol_part.cpp`, Zeilen
// 461–476, Fassung f9d9356 vom 2021-07-06) rastet auf ein Gitter und auf
// Vielfache von 2π/24, aber **begrenzt nichts**. Eine Mitte außerhalb des
// Einheitsquadrats ist ein zulässiger Drehpunkt.
//
// # Was trotzdem falsch war
//
// `Number('')` ist `0`. Wer das Feld leerte, um eine neue Zahl zu tippen,
// **schob damit die Mitte auf null** — eine Eingabe, die niemand gemacht hat.
// Das ist kein Befund am Feld, sondern ein Ruf, der unterbleibt.
import { fireEvent, render, screen } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Symmetriefeld } from './Symmetriefeld'
import type { SymbolteilDTO } from '../antworten'

const texte = kataloge.de.seiten.karten.symbol

const symmetrie: SymbolteilDTO = {
  art: 'Symmetrie',
  name: 'Quadrat',
  kombination: '',
  punkte: [],
  spiegelart: 'rotation',
  kopien: 2,
  mitteX: 0.5,
  mitteY: 0.5,
  griffX: 0.25,
  griffY: 0,
  teile: [],
}

function zeige() {
  const setzen = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Symmetriefeld
        teil={symmetrie}
        anlegen={vi.fn()}
        aufloesen={vi.fn()}
        setzen={setzen}
      />
    </Sprachraum>,
  )

  return setzen
}

// **`fireEvent.change` und nicht `userEvent.type`**: das Feld ist gesteuert
// und hält seinen Wert am Teil, das der Aufrufer reicht. Ein Tippen liefe
// gegen einen Wert, der sich nie ändert — geprüft werden soll aber die rohe
// Eingabe, nicht die Textmarke.
it('setzt die Mitte nicht, solange das Feld leer ist', () => {
  const setzen = zeige()

  fireEvent.change(screen.getByLabelText(texte.mitteX), {
    target: { value: '' },
  })

  expect(setzen).not.toHaveBeenCalled()
})

// **Eine Zahl außerhalb des Einheitsquadrats geht durch** — sie ist ein
// zulässiger Drehpunkt, und eine hier erfundene Grenze wiese sie ab (§11).
it('nimmt eine Mitte außerhalb des Einheitsquadrats an', () => {
  const setzen = zeige()

  fireEvent.change(screen.getByLabelText(texte.mitteX), {
    target: { value: '2' },
  })

  expect(setzen).toHaveBeenLastCalledWith(
    expect.objectContaining({ mitteSetzen: true, mitteX: 2 }),
  )
})

it('setzt auch den Griff nicht aus einem leeren Feld', () => {
  const setzen = zeige()

  fireEvent.change(screen.getByLabelText(texte.griffY), {
    target: { value: '' },
  })

  expect(setzen).not.toHaveBeenCalled()
})

// **Und die zweite Achse ebenso** — X und Y sind zwei Felder mit derselben
// Zusage, und ein Test über eines von beiden belegt nichts über das andere
// (§10: die Reichweite einer Messung ist die Reichweite ihrer Aussage).
it('nimmt eine Zahl im zweiten Feld des Paares an', () => {
  const setzen = zeige()

  fireEvent.change(screen.getByLabelText(texte.mitteY), {
    target: { value: '0.75' },
  })

  expect(setzen).toHaveBeenLastCalledWith(
    expect.objectContaining({ mitteSetzen: true, mitteY: 0.75 }),
  )
})

// **Die gewählte Spiegelart meldet sich als gedrückt** — AP-3
// (#2360),
// B-08. Bis dahin trug sie die Füllung des Hauptknopfs und sonst nichts: wer
// nicht sah, hörte keinen Unterschied, und wer sah, hielt sie für die
// führende Handlung.
it('meldet die gewählte Spiegelart als gedrückt und füllt sie nicht', () => {
  zeige()

  const drehung = screen.getByRole('button', {
    name: texte.drehung,
    pressed: true,
  })
  expect(drehung).not.toHaveClass('bg-marke')
  expect(drehung).toHaveClass('border-marke')
  expect(
    screen.getByRole('button', { name: texte.achse, pressed: false }),
  ).not.toHaveClass('bg-marke')
})
