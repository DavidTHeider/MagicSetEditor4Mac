// **Der Sprung im Quelltextfeld** — „Zur Stelle", „Zur Definition", eine
// Verwendung (Werkstück U7 von Phase 5.1,
// #1418).
//
// **Eine eigene Datei**, weil `Quelltextfeld.tsx` damit über die Zeilengrenze
// aus §2 ging — und weil der Sprung eine eigene Frage beantwortet: wie aus
// Zeile und Spalte ein Versatz wird, und wann das Feld ihm folgt.
import { useEffect } from 'react'

// Sprungziel ist eine Stelle im Text, an die der Zeiger soll. `stand` zählt
// die Sprünge, damit derselbe Ort zweimal hintereinander zweimal springt.
export interface Sprungziel {
  zeile: number
  spalte: number
  stand: number
}

// stelleVon rechnet Zeile und Spalte (ab 1) in einen Versatz im Text um —
// hinter dem Ende wird am Ende festgehalten.
export function stelleVon(text: string, zeile: number, spalte: number): number {
  const zeilen = text.split('\n')
  const gewaehlt = Math.min(Math.max(zeile, 1), zeilen.length)
  let versatz = 0

  for (let nummer = 1; nummer < gewaehlt; nummer += 1) {
    versatz += String(zeilen[nummer - 1]).length + 1
  }

  return (
    versatz +
    Math.min(Math.max(spalte, 1) - 1, String(zeilen[gewaehlt - 1]).length)
  )
}

// useSprung setzt den Zeiger, sobald ein neuer Stand des Ziels kommt.
//
// **Ein Effekt am Stand des Ziels.** Er greift ins Feld — Fokus, Auswahl,
// Rollen —, und das geht erst, wenn es gezeichnet ist. **Nur der Stand
// zählt**: dasselbe Ziel mit neuem Stand springt erneut, ein neu gezeichneter
// Text am alten Ziel nicht.
// **Ein Griff zum Feld, keine Referenz**: ein Argument des Hakens gilt dem
// Compiler-Linter als unveränderlich, und `scrollTop` wird geschrieben.
export function useSprung(
  holeFeld: () => HTMLTextAreaElement | null,
  sprungziel: Sprungziel | null,
  springen: (stelle: number) => void,
  aufZeiger: ((stelle: number) => void) | undefined,
): void {
  useEffect(() => {
    const feld = holeFeld()

    if (sprungziel === null || feld === null) {
      return
    }

    const stelle = stelleVon(feld.value, sprungziel.zeile, sprungziel.spalte)
    feld.focus()
    feld.setSelectionRange(stelle, stelle)
    feld.scrollTop = Math.max(
      0,
      ((sprungziel.zeile - 3) * feld.scrollHeight) /
        Math.max(1, feld.value.split('\n').length),
    )
    springen(stelle)
    aufZeiger?.(stelle)
    // eslint-disable-next-line react-hooks/exhaustive-deps -- nur der Stand des Ziels löst den Sprung aus
  }, [sprungziel?.stand])
}
