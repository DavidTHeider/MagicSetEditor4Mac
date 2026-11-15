// **Wie Mantine an diese Anwendung angeschlossen ist** — UX-07
// (#1679).
import { describe, expect, it } from 'vitest'

import { mantinethema, mantineumgebung } from './mantine'

describe('mantineumgebung', () => {
  // **Ohne `test` bleibt unter jsdom jedes Overlay auf `display: none`**, und
  // damit ist es für jede rollenbasierte Abfrage unsichtbar. Gemessen am
  // 2026-09-20: ohne diese Umgebung fallen Menü, Auswahlliste und
  // Hinweisfähnchen aus, mit ihr bestehen alle sieben Proben.
  it('nennt den Prüfmodus beim Namen', () => {
    expect(mantineumgebung('test')).toBe('test')
  })

  // **Jeder andere Modus ist der gewöhnliche.** Er steht hier als eigener
  // Fall, weil sonst die Hälfte dieser Entscheidung nie betreten würde — und
  // genau darum ist es eine Funktion und kein Ausdruck im Baum.
  it('lässt jeden anderen Modus den gewöhnlichen sein', () => {
    expect(mantineumgebung('production')).toBe('default')
    expect(mantineumgebung('development')).toBe('default')
    expect(mantineumgebung('')).toBe('default')
  })
})

describe('das Thema', () => {
  // **Die Schriften zeigen auf dieselben Merkmale wie `stil.css`.** Ein
  // zweiter Schriftstapel daneben liefe auseinander — und ein Wächter, der
  // das hält, ist billiger als die Sichtprüfung, die es fände.
  it('nimmt die Schriften dieser Anwendung', () => {
    expect(mantinethema.fontFamily).toBe('var(--font-text)')
    expect(mantinethema.headings?.fontFamily).toBe('var(--font-anzeige)')
  })

  // **Kein eigener Fokusring.** `stil.css` setzt ihn für die ganze Anwendung,
  // und `internal/werkzeuge/fokusring_test.go` hält ihn.
  it('bringt keinen zweiten Fokusring mit', () => {
    expect(mantinethema.focusRing).toBe('auto')
  })

  // **Und keine Farben** — Mantines Palette ist ein Zahlenraster, unsere sind
  // acht benannte, gemessene Merkmale (`docs/KONTRASTE.md`). Eine Abbildung
  // darauf wäre eine Erfindung, und dieser Lauf hält sie fern.
  it('bringt keine eigene Palette mit', () => {
    expect(mantinethema.colors).toBeUndefined()
    expect(mantinethema.primaryColor).toBeUndefined()
  })
})
