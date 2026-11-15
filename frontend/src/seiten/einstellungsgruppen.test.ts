// **Die Gruppen der Einstellungen** — UX-04.AK-01
// (#1676),
// und seit OPS-10 (#1908) die Schalter als zehnte.
import { describe, expect, it } from 'vitest'

import {
  einstellungsgruppen,
  fachschluessel,
  gruppeFuer,
} from './einstellungsgruppen'

describe('die zehn Gruppen', () => {
  // **Die neun aus dem Kriterium**: „Allgemein, Erscheinungsbild,
  // Nutzer/Profil, Fachfunktionen, KI, Verbindungen, Dienste, Daten/Sicherung
  // und Diagnose" — **und die Schalter** aus OPS-10.AK-04 (#1908), deren
  // Abweichung vom Entwurf in `PROTOTYP-0.6.0.md` steht.
  it('sind zehn', () => {
    expect(einstellungsgruppen).toHaveLength(10)
  })

  it('heißen so, wie der Entwurf sie zeichnet — und die Schalter dazu', () => {
    expect([...einstellungsgruppen]).toEqual([
      'allgemein',
      'bereiche',
      'erscheinungsbild',
      'nutzer',
      'fach',
      'ki',
      'verbindungen',
      'dienste',
      'daten',
      'diagnose',
    ])
  })
})

describe('gruppeFuer', () => {
  // **Jede Einstellung landet in genau einer.** Dass die Liste vollständig
  // ist, hält `internal/werkzeuge/einstellungsgruppen_test.go` gegen
  // `domain.Einstellungsreihe()` — hier steht, dass die Zuordnung tut, was
  // sie sagt.
  it('ordnet nach der Sache und nicht nach dem Bereich', () => {
    // **Sprache und Erscheinungsbild liegen beide im Bereich `anwendung`**
    // und trotzdem in zwei Gruppen: der Entwurf trennt „Sprache, Start,
    // Kürzel" von „Thema, Schrift, Hintergrund".
    expect(gruppeFuer('sprache')).toBe('allgemein')
    expect(gruppeFuer('erscheinungsbild')).toBe('erscheinungsbild')
  })

  it('stellt das Fenster und den Arbeitsplatz zur Darstellung', () => {
    expect(gruppeFuer('fensterstart')).toBe('erscheinungsbild')
    expect(gruppeFuer('arbeitsplatz.navigator')).toBe('erscheinungsbild')
  })

  // **Die KI-Anbieter stehen bei der KI und nicht bei den Diensten** —
  // AK-04 verlangt „einen eigenen eindeutig benannten Bereich".
  it('trennt die KI-Anbieter von den Kartendiensten', () => {
    expect(gruppeFuer('dienst.anthropic')).toBe('ki')
    expect(gruppeFuer('ki.laufzeit')).toBe('ki')
    expect(gruppeFuer('dienst.scryfall')).toBe('dienste')
  })

  it('stellt die Kostengrenze zur KI und den Handel zu den Diensten', () => {
    expect(gruppeFuer('richtlinie.kostengrenze')).toBe('ki')
    expect(gruppeFuer('richtlinie.handel.markt')).toBe('dienste')
  })

  it('stellt Pfade und Bildausfuhr zu den Daten', () => {
    expect(gruppeFuer('vorgabe.hochladepfad')).toBe('daten')
    expect(gruppeFuer('bildkonflikt')).toBe('daten')
  })

  it('stellt die Anmeldung zu Nutzer und Profil', () => {
    expect(gruppeFuer('anbieter.github.kennung')).toBe('nutzer')
  })

  it('stellt das Netz zu den Verbindungen', () => {
    expect(gruppeFuer('netz.proxy.http')).toBe('verbindungen')
  })

  // **Der Kontingentschalter steht bei den Diensten, die er bremst** —
  // QA-08.AK-03 (#1739): neben ihrem Zeugnis, unter den Verbindungen, und
  // eingeklappt, weil ihn nur braucht, wer eine Umgebung kontrolliert.
  it('stellt den Kontingentschalter zu den Verbindungen und klappt ihn ein', () => {
    expect(gruppeFuer('grenzen.kontingente')).toBe('verbindungen')
    expect(fachschluessel.has('grenzen.kontingente')).toBe(true)
  })

  // **Ein unbekannter Schlüssel fällt nicht heraus** (§8): ein Bestand von
  // morgen kann einen tragen, den dieses Fenster nicht kennt — und ihn
  // stillschweigend zu verstecken wäre schlimmer, als ihn unter „Allgemein"
  // zu zeigen.
  it('gibt einem unbekannten Schlüssel die erste Gruppe', () => {
    expect(gruppeFuer('gibtesnicht')).toBe('allgemein')
  })
})
