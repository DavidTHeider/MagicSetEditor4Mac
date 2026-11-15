// **Ein Hauptgewicht, eine Lasche, ein gewählter Stand** — AP-3
// (#2360),
// Befunde B-08 und B-10, Entwurf `Zustaende09`.
//
// Das Audit zählte bis zu **vier** gefüllte Flächen in einer Sicht, weil ein
// gedrückter Umschalter die Füllung des Hauptknopfs trug, und **zwei**
// Laschengestalten — einen Kasten in `Laschen` und im Setarbeitsplatz, eine
// Unterlinie in der Setschublade.
//
// # Was er hält
//
// - **Kein gewählter Stand füllt mit der Marke** — weder `aria-pressed:` noch
//   `aria-selected:`, `aria-checked:`, `peer-checked:` oder ein Datenzustand.
//   Die Füllung gehört der einen führenden Handlung; gewählt ist ein Segment
//   mit Rand (`Knopf`, `Segmente`). Den Knopf selbst hält `knopf.test.tsx`.
// - **Jede Lasche einer Leiste nimmt [lasche]** aus `bausteine/laschenstil.ts`
//   — erkannt am Element mit `role="tab"`, dessen `className` genau dieser
//   Name ist.
// - **Ein Pfeil für jede Auswahl** (B-10): das native `<select>` und die
//   Auswahl mit Suche zeichnen ihn über dieselbe Utility `auswahlpfeil`, und
//   die Suchauswahl kein eigenes Zeichen daneben.
//
// # Was er nicht hält
//
// - **Eine senkrechte Laschenliste** (`aria-orientation="vertical"` an ihrer
//   `tablist`): die Gruppen der Einstellungen stehen als Liste, nicht als
//   Leiste (UX-04.AK-01, `Gruppenliste`), und eine Unterlinie unter jeder
//   Zeile wäre die falsche Form.
// - **Wie der Pfeil aussieht.** Dass beide Felder am Fenster denselben
//   tragen, misst `ap3gestalt` an den berechneten Stilen.
// - **Wie viele gefüllte Knöpfe eine Sicht zeigt.** Das zählt am laufenden
//   Fenster `uirevue` und `uirevue09zustaende` (R1 je Blick).
// - **Ob ein Element ohne `aria-pressed` oder `aria-selected` eine Auswahl
//   ist**, die jemand als Hauptknopf zeichnet — das ist eine Frage ans Review.
import { readFileSync } from 'node:fs'
import { join, relative } from 'node:path'

import ts from 'typescript'
import { describe, expect, it } from 'vitest'

import { dateienUnter } from './test/dateien'
import { ohneKommentare } from './test/kommentare'

const quellen = join(import.meta.dirname)

const auswahlfuellung =
  /(?<![\w-])(?:aria-(?:pressed|selected|checked|current)|peer-checked|group-aria-[a-z]+|checked|data-\[[^\]\s]+\]):bg-marke(?![\w-])/g

function produktdateien(muster: string[]): string[] {
  return dateienUnter(quellen, muster, {
    ohne: ['**/*.test.ts', '**/*.test.tsx', 'test/**'],
  })
}

// attribut liest den Wert eines Attributs, wenn er eine Zeichenkette ist.
function attribut(
  element: ts.JsxOpeningLikeElement,
  name: string,
): ts.JsxAttribute | undefined {
  return element.attributes.properties.find(
    (a): a is ts.JsxAttribute =>
      ts.isJsxAttribute(a) && a.name.getText() === name,
  )
}

function zeichenkette(a: ts.JsxAttribute | undefined): string | undefined {
  return a?.initializer !== undefined && ts.isStringLiteral(a.initializer)
    ? a.initializer.text
    : undefined
}

// laschenOhneGestalt nennt jede Lasche einer waagerechten Leiste, deren
// `className` nicht [lasche] ist — mit Datei und Zeile.
export function laschenOhneGestalt(datei: string, text: string): string[] {
  const quelle = ts.createSourceFile(
    datei,
    text,
    ts.ScriptTarget.Latest,
    true,
    ts.ScriptKind.TSX,
  )
  const funde: string[] = []

  const leisteVon = (knoten: ts.Node): ts.JsxElement | undefined => {
    const leiste = ts.findAncestor(
      knoten.parent,
      (k) =>
        ts.isJsxElement(k) &&
        zeichenkette(attribut(k.openingElement, 'role')) === 'tablist',
    )

    return leiste !== undefined && ts.isJsxElement(leiste) ? leiste : undefined
  }
  const senkrecht = (knoten: ts.Node): boolean => {
    const leiste = leisteVon(knoten)

    return (
      leiste !== undefined &&
      zeichenkette(attribut(leiste.openingElement, 'aria-orientation')) ===
        'vertical'
    )
  }

  const besuche = (knoten: ts.Node): void => {
    if (
      (ts.isJsxOpeningElement(knoten) || ts.isJsxSelfClosingElement(knoten)) &&
      zeichenkette(attribut(knoten, 'role')) === 'tab' &&
      !senkrecht(knoten)
    ) {
      const klasse = attribut(knoten, 'className')?.initializer
      const gestalt =
        klasse !== undefined &&
        ts.isJsxExpression(klasse) &&
        klasse.expression !== undefined &&
        ts.isIdentifier(klasse.expression) &&
        klasse.expression.text === 'lasche'

      if (!gestalt) {
        const zeile =
          quelle.getLineAndCharacterOfPosition(knoten.getStart()).line + 1
        funde.push(`${datei}:${String(zeile)}`)
      }
    }

    ts.forEachChild(knoten, besuche)
  }
  besuche(quelle)

  return funde
}

describe('die Werkzeuge des Wächters', () => {
  it('erkennt eine Auswahlfüllung, aber nicht den Rand und nicht den Hauptknopf', () => {
    expect(
      'aria-pressed:bg-marke peer-checked:bg-marke data-[an=ja]:bg-marke'.match(
        auswahlfuellung,
      ),
    ).toHaveLength(3)
    expect(
      'bg-marke aria-pressed:border-marke aria-selected:bg-betont hover:bg-marke/10'.match(
        auswahlfuellung,
      ),
    ).toBeNull()
  })

  it('nimmt eine senkrechte Liste aus und nennt eine Lasche ohne Gestalt', () => {
    const text = [
      'const a = (',
      '  <div role="tablist" aria-orientation="vertical">',
      '    <button role="tab" className="x" />',
      '  </div>',
      ')',
      'const b = (',
      '  <div role="tablist">',
      '    <button role="tab" className={lasche} />',
      '    <button role="tab" className="rounded-md" />',
      '  </div>',
      ')',
    ].join('\n')

    expect(laschenOhneGestalt('probe.tsx', text)).toEqual(['probe.tsx:9'])
  })
})

describe('Hauptgewicht und Auswahl', () => {
  it('füllt keinen gewählten Stand mit der Marke', () => {
    const funde = produktdateien(['**/*.ts', '**/*.tsx', '**/*.css']).flatMap(
      (datei) =>
        [
          ...ohneKommentare(
            datei,
            readFileSync(join(quellen, datei), 'utf8'),
          ).matchAll(auswahlfuellung),
        ].map((treffer) => `${datei}: ${treffer[0]}`),
    )

    expect(funde).toEqual([])
  })

  it('gibt jeder Lasche einer Leiste dieselbe Gestalt', () => {
    const funde = produktdateien(['**/*.tsx']).flatMap((datei) =>
      laschenOhneGestalt(
        relative(quellen, join(quellen, datei)),
        readFileSync(join(quellen, datei), 'utf8'),
      ),
    )

    expect(funde).toEqual([])
  })
})

describe('ein Pfeil für jede Auswahl', () => {
  it('zeichnet das native Feld und die Suchauswahl über dieselbe Utility', () => {
    const stil = ohneKommentare(
      'erscheinungsmerkmale.css',
      readFileSync(join(quellen, 'erscheinungsmerkmale.css'), 'utf8'),
    )
    const suche = ohneKommentare(
      'bausteine/Suchauswahl.tsx',
      readFileSync(join(quellen, 'bausteine/Suchauswahl.tsx'), 'utf8'),
    )

    expect(stil).toMatch(/@utility auswahlpfeil\s*\{/)
    expect(stil).toMatch(
      /select:not\(\[multiple\]\):not\(\[size\]\)\s*\{[^}]*@apply auswahlpfeil;/,
    )
    expect(suche).toMatch(/(?<![\w-])auswahlpfeil(?![\w-])/)
    expect(suche).not.toContain('▾')
  })
})
