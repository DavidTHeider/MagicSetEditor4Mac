// **Wie der Wächter des Aktionskatalogs den Quelltext liest** — ausgelagert
// aus `aktionskatalog.test.ts`, damit die Zusagen dort für sich stehen und
// die Datei unter der Zeilengrenze bleibt (§2).
//
// Gelesen wird mit dem Übersetzer von TypeScript, wie in
// `keintextimjsx.test.ts`: eine Stelle ist ein Knopf, eine `Entscheidung`
// oder ein Menüeintrag, dessen Aufschrift ein Katalogschlüssel ist.
import { readFileSync } from 'node:fs'
import { dirname, join, normalize } from 'node:path'

import { dateienUnter } from './dateien'
import ts from 'typescript'

import { type Aktion } from '../aktionen'

export const quellen = join(import.meta.dirname, '..')

// endungen ordnet einen Schlüssel seiner Aktion zu — an seinem **Ende**, wie
// in `loeschwort.test.ts`: `formatLoeschen` löscht, `loeschenFrage` fragt.
const endungen: readonly (readonly [RegExp, Aktion])[] = [
  [/loeschen(ja|trotzdem)?$/i, 'loeschen'],
  [/entfernen(ja)?$/i, 'entfernen'],
  [/(anlegen|hinzufuegen)$/i, 'anlegen'],
  [/(sichern|speichern)(unter)?$/i, 'sichern'],
  [/duplizieren$/i, 'duplizieren'],
  [/kopieren$/i, 'kopieren'],
  [/(bearbeiten|umbenennen|aendern)$/i, 'bearbeiten'],
  [/zuruecknehmen$/i, 'zuruecknehmen'],
]

export function aktionVon(schluessel: string): Aktion | undefined {
  return endungen.find(([muster]) => muster.test(schluessel))?.[1]
}

export interface Stelle {
  datei: string
  zeile: number
  art: 'Knopf' | 'button' | 'Aktionsknopf' | 'Entscheidung' | 'Menüeintrag'
  schluessel: string
  aktion: string | undefined
  mitKuerzel: boolean

  // kuerzel ist das Kürzel einer Objekthandlung (`objektkuerzel`) — AP-7,
  // nachgetragen nach Befund 4 des Zweitblicks über Phase 6 von 0.9.0.
  kuerzel?: string
}

// letzterSchluessel ist der Katalogschlüssel eines Ausdrucks: `eigene.loeschen`
// ergibt `loeschen`, `einsetzen(eigene.kopieren, …)` ebenso.
function letzterSchluessel(ausdruck: ts.Expression): string | undefined {
  if (ts.isPropertyAccessExpression(ausdruck)) {
    return ausdruck.name.text
  }

  if (ts.isCallExpression(ausdruck) && ausdruck.arguments[0] !== undefined) {
    return letzterSchluessel(ausdruck.arguments[0])
  }

  // **Bei einem Wechsel zählt der Zweig, der eine Aktion nennt** — „Kopiert"
  // nach dem Klick ist die Rückmeldung, „Kopieren" davor die Aktion.
  if (ts.isConditionalExpression(ausdruck)) {
    const zweige = [ausdruck.whenTrue, ausdruck.whenFalse].map(
      letzterSchluessel,
    )

    return (
      zweige.find(
        (schluessel) => schluessel !== undefined && aktionVon(schluessel),
      ) ?? zweige.find((schluessel) => schluessel !== undefined)
    )
  }

  return undefined
}

function eigenschaft(
  knoten: ts.JsxOpeningLikeElement,
  name: string,
): ts.JsxAttribute | undefined {
  return knoten.attributes.properties.find(
    (eine): eine is ts.JsxAttribute =>
      ts.isJsxAttribute(eine) && eine.name.getText() === name,
  )
}

function zeichenkette(
  attribut: ts.JsxAttribute | undefined,
): string | undefined {
  const wert = attribut?.initializer
  return wert !== undefined && ts.isStringLiteral(wert) ? wert.text : undefined
}

function stellenIn(datei: string): Stelle[] {
  const quelle = ts.createSourceFile(
    datei,
    readFileSync(join(quellen, datei), 'utf8'),
    ts.ScriptTarget.Latest,
    true,
    datei.endsWith('.tsx') ? ts.ScriptKind.TSX : ts.ScriptKind.TS,
  )

  const stellen: Stelle[] = []
  const zeileVon = (knoten: ts.Node) =>
    quelle.getLineAndCharacterOfPosition(knoten.getStart(quelle)).line + 1

  const besuche = (knoten: ts.Node): void => {
    if (ts.isJsxElement(knoten)) {
      const art = knoten.openingElement.tagName.getText()

      if (art === 'Knopf' || art === 'button' || art === 'Aktionsknopf') {
        const kinder = knoten.children.filter(
          (kind) => !(ts.isJsxText(kind) && kind.getText().trim() === ''),
        )
        const einziges = kinder.length === 1 ? kinder[0] : undefined

        if (
          einziges !== undefined &&
          ts.isJsxExpression(einziges) &&
          einziges.expression !== undefined
        ) {
          const schluessel = letzterSchluessel(einziges.expression)

          if (schluessel !== undefined) {
            stellen.push({
              datei,
              zeile: zeileVon(knoten),
              art,
              schluessel,
              aktion: zeichenkette(
                eigenschaft(knoten.openingElement, 'aktion'),
              ),
              mitKuerzel:
                eigenschaft(knoten.openingElement, 'mitKuerzel') !== undefined,
            })
          }
        }
      }
    }

    if (
      (ts.isJsxSelfClosingElement(knoten) || ts.isJsxOpeningElement(knoten)) &&
      knoten.tagName.getText() === 'Entscheidung'
    ) {
      const bestaetigen = eigenschaft(knoten, 'bestaetigen')?.initializer
      if (
        bestaetigen !== undefined &&
        ts.isJsxExpression(bestaetigen) &&
        bestaetigen.expression !== undefined
      ) {
        const schluessel = letzterSchluessel(bestaetigen.expression)
        if (schluessel !== undefined) {
          stellen.push({
            datei,
            zeile: zeileVon(knoten),
            art: 'Entscheidung',
            schluessel,
            aktion: zeichenkette(eigenschaft(knoten, 'aktion')),
            mitKuerzel: false,
          })
        }
      }
    }

    // **Ein Menüeintrag** hat die Gestalt `{ schluessel, name, … }`.
    if (ts.isObjectLiteralExpression(knoten)) {
      const feld = (name: string) =>
        knoten.properties.find(
          (eines): eines is ts.PropertyAssignment =>
            ts.isPropertyAssignment(eines) && eines.name.getText() === name,
        )?.initializer

      const name = feld('name')
      if (feld('schluessel') !== undefined && name !== undefined) {
        const schluessel = letzterSchluessel(name)
        const aktion = feld('aktion')

        if (schluessel !== undefined) {
          stellen.push({
            datei,
            zeile: zeileVon(knoten),
            art: 'Menüeintrag',
            schluessel,
            aktion:
              aktion !== undefined && ts.isStringLiteral(aktion)
                ? aktion.text
                : undefined,
            // **Ein Menüeintrag kündigt sein Kürzel an wie ein Knopf**
            // (UXR-06) — dieselbe Zusage, derselbe Lauf, der die Taste drückt.
            mitKuerzel: feld('mitKuerzel') !== undefined,
            ...kuerzelVon(feld('kuerzel')),
          })
        }
      }
    }

    ts.forEachChild(knoten, besuche)
  }

  besuche(quelle)

  return stellen
}

// kuerzelVon nimmt das Kürzel eines Eintrags auf, wenn er eines nennt.
function kuerzelVon(wert: ts.Expression | undefined): { kuerzel?: string } {
  return wert !== undefined && ts.isStringLiteral(wert)
    ? { kuerzel: wert.text }
    : {}
}

// alleStellen liest jede Datei der Oberfläche — **auch `.ts`**: seit AP-5
// stehen Einträge in `paketeintraege.ts` und `dateieintraege.ts` (Befund 4
// des Zweitblicks über Phase 6 von 0.9.0).
export function alleStellen(): Stelle[] {
  const dateien = [
    ...dateienUnter(quellen, '**/*.tsx', {
      ohne: ['**/*.test.tsx', 'test/**', 'sprache/**'],
    }),
    ...dateienUnter(quellen, '**/*.ts', {
      ohne: ['**/*.test.ts', '**/*.d.ts', 'test/**', 'sprache/**'],
    }),
  ]

  return dateien.flatMap(stellenIn)
}

export function beschreiben(stelle: Stelle, befund: string): string {
  return `${stelle.datei}:${String(stelle.zeile)} ${stelle.art} {…${stelle.schluessel}} — ${befund}`
}

// naheOrte sind die Dateien, die eine Bindung rufen, die Dateien, die deren
// Haken benutzen — und was eine davon als Baustein einsetzt. **Weiter geht es
// nicht**: die Frage steht bei allen zweiunddreißig höchstens eine Stufe unter
// dem Ruf, und ein Ort, der weiter weg liegt, hätte mit ihm nichts zu tun.
export function naheOrte(bindung: string): string[] {
  const dateien = dateienUnter(quellen, ['**/*.ts', '**/*.tsx'], {
    ohne: ['**/*.test.ts', '**/*.test.tsx', 'test/**', 'api*.ts'],
  })
  const inhalt = new Map(
    dateien.map((datei) => [datei, readFileSync(join(quellen, datei), 'utf8')]),
  )

  const direkt = dateien.filter((datei) =>
    inhalt.get(datei)?.includes(`.${bindung}`),
  )
  const haken = direkt
    .filter((datei) => datei.startsWith('haken/'))
    .map((datei) => datei.slice('haken/'.length).replace(/\.ts$/, ''))
  const rufer = [
    ...direkt,
    ...dateien.filter((datei) =>
      haken.some((name) => inhalt.get(datei)?.includes(`/${name}'`)),
    ),
  ]

  const eingesetzt = rufer.flatMap((datei) =>
    [...(inhalt.get(datei) ?? '').matchAll(/from '(\.{1,2}\/[^']+)'/g)].map(
      (treffer) => `${normalize(join(dirname(datei), treffer[1] ?? ''))}.tsx`,
    ),
  )

  return [...rufer, ...eingesetzt]
}
