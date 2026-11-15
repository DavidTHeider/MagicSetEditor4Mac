// **Der Umriss eines Teils auf der Zeichenfläche** — Werkstück 5 von Phase 3.7
// (#1168).
//
// # Warum eine eigene Datei
//
// Weil eine **Gruppe keinen eigenen Umriss hat**: sie trägt Kinder, und die
// werden gezeichnet. Das ist eine Rekursion, und `Symbolflaeche.tsx` steht an
// der Zeilengrenze (§2).
//
// # Von hinten nach vorn
//
// **Das erste Teil der Liste liegt im Bild oben.** Der Rasterer zeichnet
// rückwärts — `for stelle := len(teile) - 1; stelle >= 0; stelle--`
// (`internal/render/symbolraster.go`), und das ist MSEs
// `FOR_EACH_CONST_REVERSE(p, g->parts)` in `SymbolViewer::combineSymbolPart`
// (`src/render/symbol/viewer.cpp`, Zeile 206, Fassung `f9d9356` vom
// 2021-07-06), samt dem Kommentar „bottom to top".
//
// **SVG malt in Dokumentreihenfolge**, das Spätere über das Frühere. Wer die
// Liste geradeaus abarbeitete, bekäme im Editor die **umgekehrte**
// Überdeckung wie im gerenderten Bild — dieselbe Zeichnung, zwei Antworten.
// Gefunden in der laufenden Anwendung am 2026-08-27, nicht von einem Test.

import type { SymbolteilDTO } from '../antworten'
import { alsPfad } from './symbolformen'

interface TeilumrissEigenschaften {
  teil: SymbolteilDTO
  groesse: number

  // klassen ist die Farbgebung — sie kommt vom obersten Teil und gilt für
  // alles darunter: was zu einer gewählten Gruppe gehört, ist mitgewählt.
  klassen: string
}

export function Teilumriss({
  teil,
  groesse,
  klassen,
}: TeilumrissEigenschaften) {
  // **Nur eine Gruppe, keine Symmetrie.** Eine Symmetrie trägt ebenfalls
  // Kinder, aber sie zeichnet sie **abgebildet** — die Kopien kommen von der
  // Go-Seite und stehen in `Symmetriesicht`. Ihre Kinder hier roh zu malen
  // hieße, das Original ein zweites Mal und unverdreht hinzulegen.
  if (teil.art === 'Gruppe') {
    return (
      <>
        {[...teil.teile].reverse().map((kind, rueckwaerts) => (
          <Teilumriss
            key={kind.name + String(rueckwaerts)}
            teil={kind}
            groesse={groesse}
            klassen={klassen}
          />
        ))}
      </>
    )
  }

  return <path d={alsPfad(teil, groesse)} className={klassen} strokeWidth={2} />
}
