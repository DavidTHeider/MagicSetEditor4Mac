// **Jede Farbklasse hat ihr Merkmal** — gefunden bei der Sichtprüfung zu
// #1643.
//
// # Der Fund, für den es diesen Wächter gibt
//
// **`border-trennlinie` gab es nicht.** Die Klasse stand an **46** Stellen in
// dreizehn Bauteilen, seit Phase 5.3 — und Tailwind erzeugt für einen Namen,
// den es nicht kennt, **keine** Regel. Der Rand erbte damit die **Textfarbe**:
// gemessen am 2026-09-19 an der laufenden Anwendung `rgb(240, 230, 210)`
// statt `#9b8768`. Jedes Eingabefeld dieser Bauteile war heller umrandet als
// jedes andere im Programm.
//
// **Kein Test sah es.** Ein DOM-Test prüft Rollen und Texte, axe prüft
// Kontraste von Text, und die Kontrastmessung in `docs/KONTRASTE.md` geht
// über **benutzte** Farben — eine Klasse ohne Regel benutzt keine.
//
// # Was er hält und was nicht
//
// **Er hält, dass jedes Farbmerkmal in einer Klasse auch eines ist.** Was er
// **nicht** hält: ob die Farbe die richtige ist, und ob sie genug Kontrast
// hat — das steht in `docs/KONTRASTE.md` und im Auge dessen, der hinsieht.
//
// **Er liest die ganze Datei, Kommentare eingeschlossen.** Das ist Absicht:
// ein Kommentar, der einen toten Namen ausschreibt, ist die Quelle des
// nächsten Kopierfehlers. Wer einen erwähnen muss, nennt ihn in Worten.
import { readFileSync, readdirSync, statSync } from 'node:fs'
import { join } from 'node:path'

import { describe, expect, it } from 'vitest'

// quelle ist der Ort der Bauteile.
const quelle = import.meta.dirname

// mitEinbindungen gibt ein Blatt samt den lokalen Blättern zurück, die es
// einbindet — so, wie Tailwind es liest: ein Merkmal in
// `erscheinungsmerkmale.css` ist eines (AP-2, #2359).
function mitEinbindungen(datei: string): string {
  return readFileSync(join(quelle, datei), 'utf8').replace(
    /@import '\.\/([^']+)';/g,
    (_, name: string) => mitEinbindungen(name),
  )
}

// stilblatt hält die Merkmale, die es gibt.
const stilblatt = mitEinbindungen('stil.css')

// merkmale sind die Namen hinter `--color-`.
const merkmale = new Set(
  [...stilblatt.matchAll(/--color-([a-z0-9-]+)\s*:/g)].map(
    (treffer) => treffer[1],
  ),
)

// farbklasse trifft `bg-`, `text-`, `border-`, `divide-`, `ring-`,
// `outline-`, `fill-` und `stroke-` mit einem einfachen Namen dahinter.
//
// **Ohne Zahlen und ohne Klammern**: `text-sm` ist eine Größe, `border-2`
// eine Breite, `bg-[#abc]` eine ausgeschriebene Farbe, und `text-red-600`
// gehört Tailwinds eigener Palette — der Nachblick `(?!-)` hält sie
// draußen, weil auf ihren Namen immer noch eine Stufe folgt.
//
// **Mit Bindestrichen im Namen**, seit dem Nachlauf zum Zweitblick über
// Phase 5.6: ein Merkmal darf `marketext` heißen und ebenso
// `markt-text` — ein Muster, das nur einfache Wörter sähe, verspräche
// mehr, als es hält (§10).
//
// **Und nicht in einer Werteliste** (AP-4): hinter `[`, `,` oder `_` steht
// eine Eigenschaft wie in `transition-[color,border-color]`, keine Klasse.
const farbklasse =
  /(?<![[,_])\b(?:bg|text|border|divide|ring|outline|fill|stroke)-([a-z]+(?:-[a-z]+)*)\b(?!-)/g

// keineFarben sind die Namen, die nach einem Merkmal aussehen und keines
// sind — Tailwinds eigene Wörter.
const keineFarben = new Set([
  'sm',
  'xs',
  'base',
  'lg',
  'xl',
  'left',
  'right',
  'center',
  'justify',
  'wrap',
  'nowrap',
  'balance',
  'pretty',
  'ellipsis',
  'clip',
  'transparent',
  'current',
  'inherit',
  'white',
  'black',
  'none',
  'solid',
  'dashed',
  'dotted',
  'double',
  'hidden',
  'offset',
  'inset',
  'top',
  'bottom',
  'x',
  'y',
  't',
  'r',
  'b',
  'l',

  // **Drei, die nach einer Farbe aussehen und keine sind**: `border-collapse`
  // ist eine Tabelleneigenschaft von Tailwind, `text-overflow` und
  // `text-writing` stehen als **CSS-Eigenschaften** in Zeichenketten.
  'collapse',
  'overflow',
  'writing',
])

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

describe('Jede Farbklasse hat ihr Merkmal', () => {
  it('nennt kein Merkmal, das es nicht gibt', () => {
    const befunde: string[] = []

    for (const pfad of dateien(quelle)) {
      const inhalt = readFileSync(pfad, 'utf8')

      for (const treffer of inhalt.matchAll(farbklasse)) {
        const name = treffer[1] ?? ''

        if (keineFarben.has(name) || merkmale.has(name)) {
          continue
        }

        // **Nur Namen, die wie ein eigenes Merkmal aussehen.** Tailwinds
        // Palette (`slate-200`) trägt eine Zahl und fällt schon oben heraus;
        // was hier landet, ist ein Wort aus diesem Projekt.
        befunde.push(`${pfad.slice(quelle.length + 1)}: ${treffer[0]}`)
      }
    }

    expect(
      befunde,
      'Diese Klassen nennen ein Farbmerkmal, das `stil.css` nicht kennt — ' +
        'Tailwind erzeugt dafür keine Regel, und der Rand erbt die Textfarbe.',
    ).toEqual([])
  })

  it('kennt die Merkmale, die es gibt', () => {
    expect(merkmale.has('kante')).toBe(true)
    expect(merkmale.has('trennlinie')).toBe(false)
  })

  // **Und es sieht auch einen Namen mit Bindestrich** — die Gegenprobe zum
  // Muster, ohne die der Wächter mehr verspräche, als er hält.
  it('sieht auch ein Merkmal mit Bindestrich', () => {
    const treffer = [...'border-neue-kante'.matchAll(farbklasse)]

    expect(treffer.map((eins) => eins[1])).toEqual(['neue-kante'])

    // **Und Tailwinds Palette bleibt draußen.**
    expect([...'text-red-600'.matchAll(farbklasse)]).toEqual([])
  })

  // **Eine Werteliste in Klammern ist keine Klasse** — AP-4 (#2361):
  // `transition-[color,background-color,border-color]` nennt Eigenschaften,
  // und `border-color` darin ist keine Randfarbe „color“. Eine Klasse steht
  // hinter Leerraum, einem Anführungszeichen oder einer Variante.
  it('liest keine Eigenschaft aus einer Werteliste', () => {
    expect([
      ...'transition-[color,background-color,border-color,opacity]'.matchAll(
        farbklasse,
      ),
    ]).toEqual([])
    expect(
      [...'hover:border-marke "text-leise'.matchAll(farbklasse)].map(
        (eins) => eins[1],
      ),
    ).toEqual(['marke', 'leise'])
  })
})
