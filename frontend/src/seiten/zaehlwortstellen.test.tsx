// **Fünf Stellen setzten Zahl und Wort aneinander** —
// #1219.
//
// # Der Befund
//
// Bei genau einem stand die Mehrzahl da: `magic · 1 Karten · magic-m15`,
// `1 Karten geöffnet`, `1 Karten auf 1 Seiten`. Gefunden bei der Sichtprüfung
// im laufenden Fenster — **ein Test konnte es nicht sehen**, denn die
// Zeichenfolge stand ja da, und ob ein Satz im Katalog *stimmt*, prüft weder
// `keintextimjsx` noch der Vertragstest.
//
// # Warum jeder Fall mit **eins** fährt
//
// Weil die Mehrzahl auch den kaputten Stand besteht. `2 Karten` steht dort
// wie hier; der Unterschied zeigt sich allein bei einer einzigen Karte, und
// ein Lauf, der zwei nimmt, prüft die Zusage nicht.
//
// # Was hier **nicht** steht
//
// Der Druckbogen und die MTG-JSON-Ausgabe. Beide melden erst nach einem
// Schreibvorgang; ihre Sätze fahren in `ausgebensetwechsel.test.tsx` und
// `druckbogen.test.tsx` mit — dort mit der Mehrzahl, weil dort der Weg zählt
// und nicht die Form. **Gesagt ist es hiermit**: die beiden gehen über
// denselben Helfer, und wer ihn ändert, ändert sie mit.
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { SetDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { raeumeApi } from '../test/api'
import { Setkarte } from './Setkarte'
import { stillerUmschlaglauf } from '../test/umschlaglauf'

afterEach(() => {
  raeumeApi()
})

const worte = kataloge.de.allgemein.zaehlwoerter

// einSet ist ein Set mit **genau einer** Karte.
function einSet(): SetDTO {
  return {
    id: '1',
    name: 'Mit einer Karte',
    spiel: 'magic',
    stil: 'magic-m15',
    karten: 1,
    herkunft: '',
    herkunftFehlt: false,
  }
}

// ruhenderLauf ist ein Speicherlauf, der nichts tut.
const ruhenderLauf = {
  speichern: vi.fn(),
  laeuft: () => false,
  speicherung: undefined,
  fehler: undefined,
  fuer: undefined,
  raeumen: vi.fn(),
}

// ruhendeLoeschung ist eine Löschung, die nichts tut.
const ruhendeLoeschung = {
  loeschen: vi.fn(() => Promise.resolve()),
  laeuft: () => false,
  befunde: () => [],
  fehler: () => undefined,
}

// einzahl sucht „1 Karte" und weist „1 Karten" ab.
function einzahl(text: string) {
  expect(text).toMatch(new RegExp(`1 ${worte.karten.eine}\\b`))
  expect(text).not.toMatch(new RegExp(`1 ${worte.karten.viele}`))
}

describe('Eine einzelne Karte heißt „1 Karte"', () => {
  it('in der Setliste', () => {
    render(
      <Sprachraum sprache="de">
        <ul>
          <Setkarte
            satz={einSet()}
            waehlen={vi.fn()}
            speichern={ruhenderLauf}
            loeschung={ruhendeLoeschung}
            umschlag={stillerUmschlaglauf()}
          />
        </ul>
      </Sprachraum>,
    )

    einzahl(screen.getByText(/magic ·/).textContent)
  })
})
