// **Hinter einer Schriftgröße steht die Zeilenhöhe, nicht die Deckkraft** —
// gefunden am 2026-10-17 an den Handbuchbildern von AP-12
// (#2369).
//
// # Der Fund, für den es diesen Wächter gibt
//
// **Die Befehlszeilen unter *Werkzeuge ohne Fenster* standen in Kästen von
// 240 Punkten Höhe**, der Befehl am unteren Rand, seit Phase 4.10. Die
// Klasse hieß Schriftgröße „sm“ mit „/60“ dahinter und war wohl als
// Deckkraft gemeint, wie bei einer Farbe. **An einer Größe ist die Zahl
// hinter dem Strich die Zeilenhöhe** in Schritten von `--spacing` — im
// gebauten Stilblatt `line-height: calc(var(--spacing) * 60)`, also 15rem.
//
// **Kein Test sah es**: ein DOM-Test prüft Rollen und Texte, und die
// Klasse ist eine gültige Klasse. Gesehen hat es nur das Bild.
//
// # Was er hält und was nicht
//
// **Er hält, dass eine Zeilenhöhe hinter einer Größe höchstens 12 Schritte
// misst** (3rem) — darüber ist sie kein Zeilenabstand mehr, sondern ein
// Kasten. Was er **nicht** hält: ob eine kleinere Zahl die richtige ist,
// und eine Zeilenhöhe über `leading-…` oder in eckigen Klammern.
import { readFileSync, readdirSync, statSync } from 'node:fs'
import { join } from 'node:path'

import { describe, expect, it } from 'vitest'

const quelle = import.meta.dirname

// groesseMitZahl trifft eine Schriftgröße mit einer Zahl hinter dem Strich.
const groesseMitZahl = /\btext-(?:xs|sm|base|lg|[2-9]?xl)\/(\d+)\b/g

// hoechstens ist die größte Zeilenhöhe in Schritten, die ein Zeilenabstand
// sein kann — `text-sm/6` steht im Bestand, 1,5rem.
const hoechstens = 12

// dateien sammelt jede `.ts`/`.tsx` unter `src`, ohne Tests.
function dateien(ort: string): string[] {
  return readdirSync(ort).flatMap((name) => {
    const pfad = join(ort, name)

    if (statSync(pfad).isDirectory()) {
      return dateien(pfad)
    }

    if (!/\.tsx?$/.test(name) || /\.test\.tsx?$/.test(name)) {
      return []
    }

    return [pfad]
  })
}

// befunde nennt jede Größe, deren Zeilenhöhe ein Kasten wäre.
function befunde(inhalt: string): string[] {
  return [...inhalt.matchAll(groesseMitZahl)]
    .filter((treffer) => Number(treffer[1]) > hoechstens)
    .map((treffer) => treffer[0])
}

describe('Hinter einer Schriftgröße steht die Zeilenhöhe', () => {
  it('setzt keine Zeilenhöhe, die ein Kasten wäre', () => {
    const gefunden = dateien(quelle).flatMap((pfad) =>
      befunde(readFileSync(pfad, 'utf8')).map(
        (klasse) => `${pfad.slice(quelle.length + 1)}: ${klasse}`,
      ),
    )

    expect(
      gefunden,
      'Die Zahl hinter einer Schriftgröße ist die Zeilenhöhe in Schritten ' +
        'von --spacing, nicht die Deckkraft — für die Deckkraft gehört sie ' +
        'hinter die Farbe.',
    ).toEqual([])
  })

  // **Die Gegenprobe am Muster**: es sieht die Klasse des Funds und lässt
  // den Zeilenabstand des Bestands und eine Farbe mit Deckkraft durch.
  it('sieht den Fund und lässt Zeilenabstand und Deckkraft durch', () => {
    expect(befunde('font-mono text-sm/60')).toEqual(['text-sm/60'])
    expect(befunde('text-2xl/16 text-xs/13')).toEqual([
      'text-2xl/16',
      'text-xs/13',
    ])
    expect(befunde('text-sm/6 text-leise/60 bg-betont/60')).toEqual([])
  })
})
