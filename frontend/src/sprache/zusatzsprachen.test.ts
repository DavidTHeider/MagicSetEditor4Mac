// **Die zusätzlichen Sprachen haben die Gestalt der englischen** — LOC-01
// (#1687).
//
// # Was er hält
//
// Jede Datei unter `it/`, `es/`, `fr/`, `ja/`, `zh/` und `ru/` exportiert
// dieselben Namen wie ihre englische Schwester, mit denselben Schlüsseln an
// denselben Stellen; jeder Wert ist eine nicht leere Zeichenkette mit
// **denselben Platzhaltern** (`{name}`) wie das Deutsche — ein verlorener
// Platzhalter ist ein Satz, der seinen Wert verschluckt. Eine fehlende Datei
// ist ein Befund.
//
// **Eine Ausnahme mit Grund**: eine Zählform darf eine dritte Form `wenige`
// tragen — das Russische braucht sie für 2 bis 4 (`Intl.PluralRules`,
// Kategorie `few`).
//
// # Was er nicht hält
//
// **Ob die Übersetzung stimmt.** Diese Kataloge sind **maschinell übersetzt**
// und tragen diesen Stand (LOC-01.AK-03); Schlüsselgleichheit belegt keine
// Qualität.
//
// `ZUSATZ_DATEIEN` (durch Komma getrennt) grenzt den Lauf auf Dateien ein —
// für einen Übersetzer, der nur seinen Teil geschrieben hat.
import { describe, expect, it } from 'vitest'

type Modul = Record<string, unknown>

const englisch = import.meta.glob<Modul>('./en/*.ts', { eager: true })
const zusatz = import.meta.glob<Modul>(
  ['./{it,es,fr,ja,zh,ru}/*.ts', '!./**/*.test.ts'],
  {
    eager: true,
  },
)

const sprachenDa = [
  ...new Set(Object.keys(zusatz).map((pfad) => pfad.split('/')[1] ?? '')),
]
const eingegrenzt = (process.env.ZUSATZ_DATEIEN ?? '')
  .split(',')
  .map((name) => name.trim())
  .filter((name) => name !== '')

const platzhalter = (text: string) =>
  [...text.matchAll(/\{[A-Za-z]+\}/g)].map((m) => m[0]).sort()

// vergleiche prüft einen Wert gegen sein englisches Vorbild und gibt die Befunde.
function vergleiche(vorbild: unknown, wert: unknown, pfad: string): string[] {
  if (typeof vorbild === 'string') {
    if (typeof wert !== 'string' || wert.trim() === '') {
      return [`${pfad}: keine Zeichenkette`]
    }

    const soll = platzhalter(vorbild).join(' ')
    const ist = platzhalter(wert).join(' ')

    return soll === ist
      ? []
      : [`${pfad}: Platzhalter ${ist || '—'} statt ${soll || '—'}`]
  }

  if (typeof vorbild !== 'object' || vorbild === null) {
    return []
  }

  if (typeof wert !== 'object' || wert === null) {
    return [`${pfad}: fehlt`]
  }

  const befunde: string[] = []
  const soll = vorbild as Record<string, unknown>
  const ist = wert as Record<string, unknown>

  for (const schluessel of Object.keys(soll)) {
    befunde.push(
      ...vergleiche(soll[schluessel], ist[schluessel], `${pfad}.${schluessel}`),
    )
  }

  for (const schluessel of Object.keys(ist)) {
    const zaehlform =
      'eine' in soll && 'viele' in soll && schluessel === 'wenige'

    if (!(schluessel in soll) && !zaehlform) {
      befunde.push(`${pfad}.${schluessel}: gibt es im Englischen nicht`)
    }
  }

  return befunde
}

describe.each(sprachenDa)('Sprache %s', (sprache) => {
  const dateien = Object.keys(englisch)
    .map((pfad) => pfad.slice('./en/'.length))
    .filter((name) => eingegrenzt.length === 0 || eingegrenzt.includes(name))

  it.each(dateien)('%s hat die Gestalt der englischen', (name) => {
    const vorbild = englisch[`./en/${name}`]
    const modul = zusatz[`./${sprache}/${name}`]

    expect(modul, `${sprache}/${name} fehlt`).toBeDefined()

    const befunde = Object.keys(vorbild ?? {}).flatMap((export_) =>
      vergleiche(
        vorbild?.[export_],
        modul?.[export_],
        `${sprache}/${name}:${export_}`,
      ),
    )

    expect(befunde).toEqual([])
  })
})
