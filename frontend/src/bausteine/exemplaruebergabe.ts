// **Was ein Exemplar mitnimmt, wenn es weiterreist** — aus
// `Anschlussknoepfe.tsx` herausgelöst, als dieselben zwei Wege auch im
// Rechtsklick an der Zeile und im Menü „Exemplar“ standen (AP-6, Teil 2,
// #2363).
// Eine Stelle für alle drei: was ins Deck geht und was in einen Handel,
// steht nur hier.
import type { Anschluss } from '../anschluss'
import type { ExemplarDTO } from '../api'

// **Der Zustand reist unverändert mit** — warum, steht in
// `Anschlussknoepfe.tsx`.
export function zumTausch(exemplar: ExemplarDTO): Anschluss {
  return {
    art: 'handel',
    name: exemplar.name,
    printing: exemplar.printing,
    guete: exemplar.zustand,
    sprache: exemplar.sprache,
    finish: exemplar.finish,
  }
}

export function insDeck(exemplar: ExemplarDTO): Anschluss {
  return { art: 'deck', name: exemplar.name, printing: exemplar.printing }
}
