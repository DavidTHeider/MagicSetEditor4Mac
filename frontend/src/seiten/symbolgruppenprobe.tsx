// **Vorrichtung für den Symboleditor mit Gruppen** — Werkstück 5 von Phase 3.7
// (#1168).
//
// **Eine eigene Datei**, weil zwei Läufe sie brauchen und `symbolgruppen`
// sonst über die Zeilengrenze ginge (§2): `symbolgruppen.test.tsx` prüft die
// Bedienung, `symbolflaechenordnung.test.tsx` die Malreihenfolge.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi } from '../test/api'
import { leererSetkopf } from '../test/setkopfprobe'
import { Symbol } from './Symbol'

// Teilprobe ist die Gestalt eines Teils, so wie Go sie schickt.
//
// **Ausgeschrieben und nicht aus der Fabrik erschlossen**: ein `teile`, das
// auf `ReturnType<typeof teil>` zeigt, ist auf sich selbst bezogen, und
// TypeScript weist es ab.
export interface Teilprobe {
  art: string
  name: string
  kombination: string
  punkte: ReturnType<typeof punkt>[]
  spiegelart: string
  kopien: number
  mitteX: number
  mitteY: number
  griffX: number
  griffY: number
  teile: Teilprobe[]
}

// teil baut ein Teil der Antwort.
export function teil(
  name: string,
  art = 'Form',
  kinder: Teilprobe[] = [],
  ecken = 3,
): Teilprobe {
  return {
    art,
    name,
    kombination: 'overlap',
    punkte:
      art === 'Form'
        ? Array.from({ length: ecken }, (_, i) => punkt(0.1 + i * 0.2, 0.8))
        : [],
    spiegelart: '',
    kopien: 0,
    mitteX: 0,
    mitteY: 0,
    griffX: 0,
    griffY: 0,
    teile: kinder,
  }
}

function punkt(x: number, y: number) {
  return {
    x,
    y,
    griffDavorX: 0,
    griffDavorY: 0,
    griffDanachX: 0,
    griffDanachY: 0,
    kurveDanach: false,
    sperre: 'free',
  }
}

// zweiTeile ist ein Set, das mit zwei Formen aufschlägt.
//
// **Drei Punkte und vier** — daran lassen sich die beiden im gemalten Pfad
// auseinanderhalten, ohne dass ein Test in die Reihenfolge der DOM-Knoten
// greift.
export function zweiTeile() {
  return {
    teile: [teil('Dreieck'), teil('Rechteck', 'Form', [], 4)],
    name: 'symbol1.mse-symbol',
    fehler: null,
  }
}

export function zeige() {
  render(
    <Sprachraum sprache="de">
      <Symbol setID="7" />
    </Sprachraum>,
  )
}

// mitSymbol belegt die Fassade so, dass das Set ein Symbol mit zwei Teilen hat.
export function mitSymbol(weiteres: Parameters<typeof belegeApi>[0] = {}) {
  belegeApi({
    Setangaben: vi.fn().mockResolvedValue({
      felder: [],
      werte: [{ feld: 'symbol', wert: 'symbol1.mse-symbol' }],
      kopf: leererSetkopf,
      fehler: null,
    }),
    Symbol: vi.fn().mockResolvedValue(zweiTeile()),
    ...weiteres,
  })
}

// teilknopf gibt den Wahlknopf eines Teils.
export function teilknopf(name: string, zahl = 3) {
  return screen.getByRole('button', {
    name: `${name} — ${String(zahl)} Punkte`,
  })
}

// beideWaehlen wählt Dreieck und Rechteck — das zweite mit ⌘.
//
// **Über eine eigene Sitzung von `userEvent`.** Die kurzen Aufrufe
// (`userEvent.click`) legen je eine an; ein `{Meta>}` aus dem einen Aufruf
// erreicht den Klick des nächsten nicht, und der Test sähe einen gewöhnlichen
// Klick — die erste Fassung war genau daran rot.
export async function beideWaehlen() {
  const nutzer = userEvent.setup()

  await screen.findByRole('button', { name: 'Dreieck — 3 Punkte' })
  await nutzer.click(teilknopf('Dreieck'))

  await nutzer.keyboard('{Meta>}')
  await nutzer.click(teilknopf('Rechteck', 4))
  await nutzer.keyboard('{/Meta}')
}
