// **Jede Seite über dem Bestand lädt nach, wenn ein anderer schreibt** —
// Befund 3 des Gesamtreviews über 0.6.0 (#1660).
//
// Der fremde Stand (`useBestandsstand`, #156) ging an Start, Pakete und
// Handeln, nicht an Sets, Sammlung, Decks, Turniere und Spielen. Legte die
// Kommandozeile ein Deck an, während die Seite offen war, stand es dort erst
// nach einem Wechsel. Der Horchposten aus 6.0 war für spätere Seiten
// unterschiedlich angeschlossen — dieselbe Klasse wie im Handel
// (`handelsraumnachladen.test.tsx`).
import { render, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { ReactElement } from 'react'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { stillerSpeicherlauf } from '../test/speicherlauf'
import { Decks } from './Decks'
import { Sammlung } from './Sammlung'
import { Sets } from './Sets'
import { Spielen } from './Spielen'
import { Turniere } from './Turniere'

afterEach(() => {
  raeumeApi()
})

// fall ist eine Seite und die Bindung, die ihre Liste holt.
interface Fall {
  bindung: string
  antwort: unknown
  seite: (bestandsstand: number) => ReactElement
}

// **Stabile Rückrufe**: ein neues `vi.fn()` je Zeichnen löste über einen
// Effekt selbst ein Nachladen aus, und der Lauf bestand aus dem falschen
// Grund (gesehen an der Gegenprobe über `useDecks`).
const ruf = {
  waehlen: vi.fn(),
  speichern: stillerSpeicherlauf(),
  anstossVerbraucht: vi.fn(),
  neuAngestossen: vi.fn(),
  geloescht: vi.fn(),
  wechseln: vi.fn(),
}

const faelle: Record<string, Fall> = {
  Sets: {
    bindung: 'Sets',
    antwort: { sets: [] },
    seite: (bestandsstand) => <Sets {...ruf} bestandsstand={bestandsstand} />,
  },
  Sammlung: {
    bindung: 'Sammlungsfenster',
    antwort: { zeilen: [], gesamt: 0, insgesamt: 0 },
    seite: (bestandsstand) => (
      <Sammlung bestandsstand={bestandsstand} bestandGeaendert={vi.fn()} />
    ),
  },
  Decks: {
    bindung: 'Decks',
    antwort: { decks: [] },
    seite: (bestandsstand) => (
      <Decks wechseln={ruf.wechseln} bestandsstand={bestandsstand} />
    ),
  },
  Turniere: {
    bindung: 'Turniere',
    antwort: { turniere: [] },
    seite: (bestandsstand) => <Turniere bestandsstand={bestandsstand} />,
  },
  Spielen: {
    bindung: 'Partien',
    antwort: { partien: [] },
    seite: (bestandsstand) => (
      <Spielen wechseln={ruf.wechseln} bestandsstand={bestandsstand} />
    ),
  },
}

describe('Eine fremde Änderung', () => {
  for (const [name, fall] of Object.entries(faelle)) {
    it(`lädt ${name} nach, solange die Seite offen ist`, async () => {
      const geholt = vi.fn(() => Promise.resolve(fall.antwort))
      belegeApi({ [fall.bindung]: geholt })

      const ansicht = (stand: number) => (
        <Sprachraum sprache="de">{fall.seite(stand)}</Sprachraum>
      )
      const { rerender } = render(ansicht(0))

      await waitFor(() => {
        expect(geholt).toHaveBeenCalled()
      })
      const vorher = geholt.mock.calls.length

      rerender(ansicht(1))

      await waitFor(() => {
        expect(geholt.mock.calls.length).toBeGreaterThan(vorher)
      })
    })
  }
})
