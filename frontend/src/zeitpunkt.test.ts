// **Ein Zeitpunkt, wie ein Mensch ihn liest** —
// gefunden bei der Sichtprüfung am 2026-08-24 (AGENTS.md §3).
//
// Im Fenster stand `2026-08-23T21:01:56.596+00:00` — die rohe Angabe von
// Scryfall, mit Millisekunden und Versatz. Kein Test hat das gezeigt: die
// Go-Seite reicht sie **absichtlich** unverändert durch (§3, „die Wanduhrzeit
// mit Zone, nicht der UTC-Kalendertag"), und der Fenstertest prüfte, dass der
// Wert **ankommt**, nicht wie er aussieht.
import { describe, expect, it } from 'vitest'

import { alsKalendertag, alsZeitpunkt } from './zeitpunkt'

describe('alsZeitpunkt', () => {
  // **Die Zone ist eine eigene Einstellung** — LOC-03.AK-03
  // (#1689).
  // Mittag in UTC ist in Tokio neun Uhr abends und in New York acht Uhr
  // morgens (gemessen am 2026-09-22).
  it.each([
    ['Asia/Tokyo', '22.09.2026, 21:00'],
    ['America/New_York', '22.09.2026, 08:00'],
  ])('zeigt einen Zeitpunkt in der Zone %s', (zone, erwartet) => {
    expect(alsZeitpunkt('2026-09-22T12:00:00Z', 'de-DE', zone)).toBe(erwartet)
  })

  it('macht aus einer ISO-Angabe ein lesbares Datum', () => {
    // **Die Zone ist die des Rechners**, und das ist der Ort, an dem diese
    // Entscheidung hingehört: Go kennt sie nicht, der Browser schon.
    const gelesen = alsZeitpunkt('2026-08-23T21:01:56.596+00:00', 'de')

    expect(gelesen).not.toContain('T')
    expect(gelesen).not.toContain('596')
    expect(gelesen).toContain('2026')
  })

  it('gibt eine unleserliche Angabe unverändert zurück', () => {
    // **Nicht raten.** Was sich nicht als Zeitpunkt lesen lässt, ist besser
    // roh zu sehen als als „Invalid Date" — dann sieht wenigstens jemand,
    // was ankam.
    expect(alsZeitpunkt('morgen', 'de')).toBe('morgen')
  })

  it('gibt eine leere Angabe leer zurück', () => {
    expect(alsZeitpunkt('', 'de')).toBe('')
  })

  it('folgt der Sprache der Sitzung', () => {
    const deutsch = alsZeitpunkt('2026-08-23T21:01:56.596+00:00', 'de')
    const englisch = alsZeitpunkt('2026-08-23T21:01:56.596+00:00', 'en')

    // **Sie unterscheiden sich**, und das ist der Punkt: ein Datum, das in
    // jeder Sprache gleich aussieht, folgt keiner.
    expect(deutsch).not.toBe(englisch)
  })
})

// **Ein Kalendertag bleibt sein Tag** — H04 (#2258): in UTC gezeigt, nicht
// über die Zone des Rechners; was keiner ist, steht da, wie es kam.
describe('alsKalendertag', () => {
  it('zeigt einen Tag in der Darstellung', () => {
    expect(alsKalendertag('2026-09-29', 'de')).toBe('29. September 2026')
  })

  it('lässt stehen, was kein Tag ist', () => {
    expect(alsKalendertag('gestern', 'de')).toBe('gestern')
  })
})
