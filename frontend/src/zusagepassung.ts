// **Ob ein Exemplar eine Zeile des Tauschs erfüllt** —
// #2492.
//
// **Dieselbe Gleichheit wie im Dienst** (`domain.Zusagevertrag`): Zustand und
// Sprache genau, die Ausführung mit „leer heißt nonfoil"
// (`domain.Exemplar.Ausfuehrung`). Der Dienst hält nur fest, was eine Zeile
// so erfüllt; ein Fenster, das mehr anböte, böte eine Wahl an, die scheitert.
// Der Druck ist schon derselbe — gefragt wird die Sammlung nach ihm.
import type { Handelsposition } from './apihandelsleitung'

interface Variante {
  zustand: string
  sprache: string
  finish: string
}

const ausfuehrung = (finish: string) => (finish === '' ? 'nonfoil' : finish)

export function erfuellt(zeile: Variante, position: Handelsposition): boolean {
  return (
    zeile.zustand === position.guete &&
    zeile.sprache === position.sprache &&
    ausfuehrung(zeile.finish) === ausfuehrung(position.finish)
  )
}
