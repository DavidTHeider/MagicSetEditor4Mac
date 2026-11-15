// **Welche Zeilen entschieden sind** — Werkstück 3 von Phase 4.5
// (#1292).
//
// **Eine eigene Datei**, weil `Listeneinfuhr.test.tsx` an der Zeilengrenze aus
// §2 stand — und weil der Schnitt der Frage folgt: dort, *wie der Kasten sich
// verhält*, hier, *welche Zeile mitgeht*. Diese Prüfungen brauchen weder DOM
// noch Fassade.
import { expect, it } from 'vitest'

import { entschieden } from './Deutungszeile'

import type { Listendeutungszeile } from '../antwortenlisteneinfuhr'

// eindeutige gibt eine Zeile mit genau einem Druck.
const eindeutige: Listendeutungszeile = {
  roh: '4 Lightning Bolt (M11) 149',
  zeilennummer: 1,
  menge: 4,
  name: 'Lightning Bolt',
  setcode: 'm11',
  nummer: '149',
  finish: 'nonfoil',
  treffer: [
    {
      printing: 'druck-1',
      name: 'Lightning Bolt',
      setcode: 'm11',
      setname: 'Magic 2011',
      nummer: '149',
      seltenheit: 'common',
      bild: '',
    },
  ],
  drucke: 1,
  eindeutig: true,
  befund: '',
  weitere: false,
  fehler: null,
}

// offene gibt eine Zeile mit zwei Drucken.
const offene: Listendeutungszeile = {
  ...eindeutige,
  roh: '2 Counterspell',
  zeilennummer: 2,
  menge: 2,
  name: 'Counterspell',
  setcode: '',
  nummer: '',
  treffer: [
    {
      printing: 'druck-2',
      name: 'Counterspell',
      setcode: 'lea',
      setname: 'Limited Edition Alpha',
      nummer: '54',
      seltenheit: 'common',
      bild: '',
    },
  ],
  drucke: 2,
  eindeutig: false,
}

// **Eine Wahl, die zu keinem Treffer gehört, fällt weg.**
//
// `wahl` hängt an der **Zeilennummer** und überlebt damit theoretisch eine
// Deutung, die andere Treffer hat. Der Kasten räumt sie zwar (§7, Regel 1) —
// aber ein Zweig, der sich auf das Aufräumen an anderer Stelle verlässt, ist
// eine Zusage über zwei Dateien.
it('lässt eine Wahl weg, die zu keinem Treffer gehört', () => {
  expect(
    entschieden(
      { zeilen: [offene], befunde: [], alterSekunden: 0 },
      {
        2: 'druck-gibtesnicht',
      },
    ),
  ).toHaveLength(0)
})

// **Und eine eindeutige Zeile ohne Treffer ebenso.**
//
// `eindeutig` und eine leere Trefferliste widersprechen einander, und die
// Antwort kommt von drüben. Ein Absturz wäre die falsche Reaktion darauf.
it('lässt eine eindeutige Zeile ohne Treffer weg', () => {
  expect(
    entschieden(
      {
        zeilen: [{ ...eindeutige, treffer: [] }],
        befunde: [],
        alterSekunden: 0,
      },
      {},
    ),
  ).toHaveLength(0)
})

// **Ohne Deutung gibt es nichts zu entscheiden.**
it('gibt ohne Deutung keine Zeile', () => {
  expect(entschieden(null, {})).toHaveLength(0)
})

// **Eine eindeutige Zeile geht ohne Zutun mit — samt lesbarem Bezug.**
it('nimmt eine eindeutige Zeile ohne Wahl', () => {
  expect(
    entschieden({ zeilen: [eindeutige], befunde: [], alterSekunden: 0 }, {}),
  ).toEqual([
    {
      printing: 'druck-1',
      menge: 4,
      zustand: 'nearMint',
      sprache: 'en',
      finish: 'nonfoil',
      name: 'Lightning Bolt',
      setcode: 'm11',
      nummer: '149',
    },
  ])
})
