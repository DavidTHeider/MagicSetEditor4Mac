// **Ein Preis im Feld** — Zweitblick über Phase 6.3, S6: in der
// Schreibweise des Fensters und mit den Stellen der Währung.
import { describe, expect, it } from 'vitest'

import { preisImFeld } from './preiseingabe'

describe('preisImFeld', () => {
  // **Ein Betrag ohne Nachkommastellen** bekommt die der Währung.
  it('schreibt einen ganzen Betrag mit den Stellen der Währung', () => {
    expect(preisImFeld('19', 'EUR', 'de')).toBe('19,00')
    expect(preisImFeld('19.5', 'EUR', 'en')).toBe('19.50')
  })

  // **Mehr Stellen, als die Währung hat, bleiben stehen**, wie sie sind —
  // sonst rundete das Feld einen Wert, der im Bestand steht.
  it('lässt stehen, was mehr Stellen trägt als die Währung', () => {
    expect(preisImFeld('1.234', 'EUR', 'de')).toBe('1.234')
  })
})
