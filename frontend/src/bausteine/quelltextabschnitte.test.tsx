// **Was `abschnitte` mit Merkmalen tut, die nicht zum Text passen** —
// Phase 4.8 (#1321).
//
// # Warum dieser Test entstanden ist
//
// Die beiden Wächter in `zeilenteile` — „das Merkmal liegt hinter dem Ende der
// Zeile" und „es überlappt das vorige" — wurden bis dahin **nur zufällig**
// betreten: irgendein Test der Vorlagenseite traf sie über eine verspätete
// Farbantwort. Sichtbar wurde das an der Abdeckung, und zwar daran, dass sie
// **schwankte**: 5.513 von 5.514 Anweisungen in einem Lauf, 5.514 im nächsten.
//
// Eine Abdeckung, die zwischen zwei Läufen springt, sagt nicht, dass der Zweig
// geprüft ist — sie sagt, dass niemand ihn **meint** (§10).
//
// # Und der Fall ist echt
//
// Die Farben werden **asynchron** geholt. Wer tippt, während die Antwort
// unterwegs ist, bekommt Merkmale zu einem Text, den es nicht mehr gibt: die
// Antwort gehört zu einem älteren Stand (§7, Regel 2). Der Text wird dann
// ungefärbt gezeigt und nicht zerschnitten.
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { abschnitte } from './quelltextanzeige'
import type { Merkmalstelle } from '../api'

function merkmal(zeile: number, spalte: number, breite: number): Merkmalstelle {
  return { art: 'zeichenkette', zeile, spalte, breite }
}

function zeige(wert: string, merkmale: Merkmalstelle[]) {
  render(
    <div data-testid="text">
      {abschnitte(wert, merkmale, null, null, null)}
    </div>,
  )

  return screen.getByTestId('text')
}

describe('Merkmale, die nicht zum Text passen', () => {
  // **Der ganze Text bleibt stehen** — nichts fehlt, nichts wird doppelt
  // gezeigt. Das ist die Zusage; die Farbe ist die Zugabe.
  it('lässt eine Zeile stehen, wenn ein Merkmal hinter ihrem Ende beginnt', () => {
    const anzeige = zeige('kurz', [merkmal(1, 99, 3)])

    expect(anzeige).toHaveTextContent('kurz')
  })

  it('lässt eine Zeile stehen, wenn ein Merkmal das vorige überlappt', () => {
    // Das erste färbt die Spalten 1–4, das zweite begänne bei 2 — mitten
    // darin. Wer es trotzdem setzte, schnitte die Zeile ein zweites Mal
    // auseinander und zeigte Zeichen doppelt.
    const anzeige = zeige('abcdef', [merkmal(1, 1, 4), merkmal(1, 2, 2)])

    expect(anzeige).toHaveTextContent('abcdef')
  })

  // **Die Gegenprobe zu beiden**: ein Merkmal, das passt, färbt auch.
  it('färbt ein Merkmal, das in die Zeile passt', () => {
    const anzeige = zeige('abcdef', [merkmal(1, 2, 3)])

    expect(anzeige).toHaveTextContent('abcdef')
    expect(anzeige.querySelectorAll('span').length).toBeGreaterThan(0)
  })
})
