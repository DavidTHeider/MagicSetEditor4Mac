// **Welcher Menüeintrag für einen Leser gesperrt ist** — #2288
// (#2288).
//
// Dasselbe wie `schreibknoepfe.test.ts`, für die Zeilen eines `Menue`: die
// Sperre hängt an `schreibt`, und **welche Zeile schreibt, entscheidet der
// Träger** in `docs/AUTORISIERUNG.md`. Pakete gehören dem Rechner — ihr
// Löschen und Zurücksetzen bleibt offen; Kopieren und Teilen lesen nur.
//
// **Was dieser Wächter hält**: jeder genannte Eintrag trägt `schreibt: true`,
// jeder offene nicht. **Was er nicht hält**: ob die Tafel vollständig ist —
// ein neuer Eintrag erscheint hier nicht von selbst.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const quelle = (datei: string) => readFileSync(join(__dirname, datei), 'utf8')

// eintrag gibt den Verbund, in dem der Schlüssel steht — **mit gezählten
// Klammern**: eine Aufschrift wie `einsetzen(…, { zahl })` trägt selbst eine.
function eintrag(text: string, schluessel: string): string {
  const stelle = text.indexOf(`schluessel: '${schluessel}'`)
  expect(stelle, `der Schlüssel ${schluessel} steht nicht da`).toBeGreaterThan(
    -1,
  )

  const beginn = text.lastIndexOf('{', stelle)
  let tiefe = 0

  for (let bei = beginn; bei < text.length; bei += 1) {
    tiefe += text[bei] === '{' ? 1 : text[bei] === '}' ? -1 : 0
    if (tiefe === 0) {
      return text.slice(beginn, bei + 1)
    }
  }

  return text.slice(beginn)
}

const gesperrt: readonly (readonly [string, string, string])[] = [
  ['seiten/Kartenhandlungen.tsx', 'duplizieren', 'Karten anlegen'],
  [
    'seiten/Kartenhandlungen.tsx',
    'ausschneiden',
    'Karten über die Zwischenablage bewegen',
  ],
  [
    'seiten/Kartenhandlungen.tsx',
    'einfuegen',
    'Karten über die Zwischenablage bewegen',
  ],
  ['seiten/Kartenhandlungen.tsx', 'nachOben', 'Karten umordnen'],
  ['seiten/Kartenhandlungen.tsx', 'nachUnten', 'Karten umordnen'],
  ['seiten/Kartenhandlungen.tsx', 'bezug', 'den Bezug einer Karte setzen'],
  ['seiten/Kartenhandlungen.tsx', 'loeschen', 'eine Karte löschen'],
  [
    'seiten/Kartenwerkzeuge.tsx',
    'scryfall',
    'Kartendaten von Scryfall übernehmen',
  ],
  ['seiten/Setkarte.tsx', 'loeschen', 'Ein Set löschen'],
  ['haken/useExemplarhandlungen.ts', 'bearbeiten', 'Ein Exemplar ändern'],
  ['haken/useExemplarhandlungen.ts', 'loeschen', 'Ein Exemplar ändern'],
  ['seiten/paketeintraege.ts', 'marken', 'Marken an jedem Fachobjekt führen'],
  [
    'bausteine/Partienverwaltung.tsx',
    'duplizieren',
    'Eine Partie löschen oder kopieren',
  ],
  [
    'bausteine/Partienverwaltung.tsx',
    'loeschen',
    'Eine Partie löschen oder kopieren',
  ],
  [
    'bausteine/Partienverwaltung.tsx',
    'marken',
    'Marken an jedem Fachobjekt führen',
  ],
  [
    'bausteine/Handelsaktionen.tsx',
    'kopieren',
    'Einen Handel als neuen Entwurf kopieren',
  ],
  ['bausteine/Handelsaktionen.tsx', 'loeschen', 'einen Handel löschen'],
  [
    'bausteine/Handelsaktionen.tsx',
    'archiv',
    'Einen beendeten Handel archivieren',
  ],
  ['bausteine/Deckplaetze.tsx', 'mehr', 'Eine Karte ins Deck legen'],
  ['bausteine/Deckplaetze.tsx', 'weniger', 'Einen Platz aus dem Deck nehmen'],
  ['bausteine/Deckplaetze.tsx', 'bereich', 'Eine Karte ins Deck legen'],
  ['bausteine/Deckplaetze.tsx', 'nehmen', 'Einen Platz aus dem Deck nehmen'],
  // **Das Menü am offenen Deck** — AP-6 (#2363); bis dahin vier Knöpfe in
  // `Deckliste.tsx` mit eigener Sperre. Nachgetragen nach Befund 4 des
  // Zweitblicks über Phase 6 von 0.9.0.
  ['haken/useDeckhandlung.ts', 'umbenennen', 'Ein Deck umbenennen'],
  ['haken/useDeckhandlung.ts', 'duplizieren', 'Ein Deck kopieren'],
  ['haken/useDeckhandlung.ts', 'loeschen', 'Ein Deck löschen'],
]

const offen: readonly (readonly [string, string, string])[] = [
  [
    'seiten/Kartenhandlungen.tsx',
    'kopieren',
    'AblageKopieren ist lesend (SCHREIBWEGE.md)',
  ],
  [
    'seiten/Kartenhandlungen.tsx',
    'teilen',
    'Etwas über die Ziele des Systems teilen — leser',
  ],
  ['haken/useDeckhandlung.ts', 'teilen', 'Ein Deck als Liste teilen — leser'],
  ['seiten/paketeintraege.ts', 'loeschen', 'Pakete gehören dem Rechner'],
  ['seiten/paketeintraege.ts', 'zurueck', 'Pakete gehören dem Rechner'],
]

describe('Die Schreibeinträge der Menüs für einen Leser', () => {
  it.each(gesperrt)('%s: %s ist gesperrt (%s)', (datei, schluessel) => {
    expect(eintrag(quelle(datei), schluessel)).toContain('schreibt: true')
  })

  it.each(offen)('%s: %s bleibt offen (%s)', (datei, schluessel) => {
    expect(eintrag(quelle(datei), schluessel)).not.toContain('schreibt')
  })
})
