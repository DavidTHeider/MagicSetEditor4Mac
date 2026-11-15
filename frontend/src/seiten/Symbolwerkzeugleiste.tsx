// **Die Werkzeugleiste des Symboleditors** — Grundformen, Nachzeichnen und
// der Verlauf.
//
// # Warum eine eigene Datei
//
// `Symbol.tsx` stand an der Zeilengrenze aus §2, als das Nachzeichnen dazukam
// (#1340).
// Der Schnitt folgt der Frage: *was* die Knöpfe tun, steht dort; *wie* sie
// aussehen, hier.
//
// # Was MSE hier hat
//
// `src/gui/symbol/window.cpp` (Fassung `f9d9356` vom 2021-07-06): eine
// Werkzeugleiste mit Grundformen, Punkten, Symmetrie und Auswahl, dazu Undo
// und Redo im Menü *Edit*. **Hier sind Undo und Redo auch Knöpfe** — sie
// stehen im Menü ebenso, aber ein Knopf sagt zusätzlich, **ob** es etwas
// zurückzunehmen gibt.

import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Knopf } from '../bausteine/Knopf'
import { useTexte } from '../sprache/kontext'

// grundformen sind die vier Knöpfe der Formleiste.
//
// **Die Eckenzahl steht dabei**, weil sie die Form bestimmt: MSE hat dafür ein
// Zählfeld von 2 bis 10 (`symmetry_editor.cpp`, Zeile 57). Hier stehen die
// gebräuchlichen als eigene Knöpfe; ein Zählfeld folgt, wenn jemand es
// braucht — und das ist eine Vertagung, keine Auslassung.
// **Mit Zeichen** — AP-12 (#2369), B-26: die Grundformen standen als leise
// Textknöpfe da. Das Zeichen ist Schmuck und für Hilfsmittel stumm; der Name
// trägt den Knopf.
export const grundformen = [
  { art: 'Kreis', ecken: 0, eng: true, text: 'kreis', zeichen: '○' },
  { art: 'Rechteck', ecken: 0, eng: false, text: 'rechteck', zeichen: '□' },
  { art: 'Vieleck', ecken: 3, eng: false, text: 'dreieck', zeichen: '△' },
  { art: 'Stern', ecken: 5, eng: false, text: 'stern', zeichen: '☆' },
] as const

interface LeisteEigenschaften {
  hinzufuegen: (art: string, ecken: number, eng: boolean) => void
  nachzeichnen: () => void

  zuruecknehmen: () => void
  wiederherstellen: () => void
  kannZuruecknehmen: boolean
  kannWiederherstellen: boolean
}

export function Symbolwerkzeugleiste({
  hinzufuegen,
  nachzeichnen,
  zuruecknehmen,
  wiederherstellen,
  kannZuruecknehmen,
  kannWiederherstellen,
}: LeisteEigenschaften) {
  const eigene = useTexte().seiten.karten.symbol

  return (
    <>
      <div
        className="flex flex-wrap gap-2"
        role="group"
        aria-label={eigene.verlauf}
      >
        <Aktionsknopf
          aktion="zuruecknehmen"
          mitKuerzel
          gewicht="leise"
          aufKlick={zuruecknehmen}
          angehalten={!kannZuruecknehmen}
        >
          {eigene.zuruecknehmen}
        </Aktionsknopf>
        <Knopf
          gewicht="leise"
          aufKlick={wiederherstellen}
          angehalten={!kannWiederherstellen}
        >
          {eigene.wiederherstellen}
        </Knopf>
      </div>

      <div
        className="flex flex-wrap gap-2"
        role="group"
        aria-label={eigene.grundform}
      >
        {grundformen.map((form) => (
          <Knopf
            key={form.art + String(form.ecken)}
            gewicht="leise"
            aufKlick={() => {
              hinzufuegen(form.art, form.ecken, form.eng)
            }}
          >
            <span aria-hidden="true" className="mr-1.5">
              {form.zeichen}
            </span>
            {eigene[form.text]}
          </Knopf>
        ))}

        {/* **Der Hilfesatz steht daneben und nicht als `aria-label`.** Er
            ersetzte sonst die Beschriftung: der Knopf hieße für ein
            Vorleseprogramm anders als für das Auge, und wer „Bild
            nachzeichnen" sucht, fände ihn nicht. Dieselbe Form wie bei
            `gruppierenHilfe` in der Teilliste. */}
        <Knopf gewicht="leise" aufKlick={nachzeichnen}>
          {eigene.nachzeichnen}
        </Knopf>
      </div>

      <p className="text-sm text-leise">{eigene.nachzeichnenHilfe}</p>
    </>
  )
}
