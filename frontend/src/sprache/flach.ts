// **Ein Katalog, flach und wieder gefaltet** — LOC-02
// (#1688).
//
// Eine eigene Übersetzung führt ihre Sätze flach: Pfad → Satz, derselbe
// Pfad wie in `internal/app/sprachvorlage.json`. Hier wird daraus wieder ein
// Katalog — **über** der Basis, die wiederum über dem Deutschen liegt. Was
// fehlt, zeigt damit die Basis, und eine leere Beschriftung gibt es nicht
// (AK-03).

type Zweig = Record<string, unknown>

// flach gibt jeden Satz eines Katalogs unter seinem Pfad.
export function flach(katalog: object, vorsatz = ''): Record<string, string> {
  const saetze: Record<string, string> = {}

  for (const [schluessel, wert] of Object.entries(katalog)) {
    const pfad = vorsatz + schluessel

    if (typeof wert === 'string') {
      saetze[pfad] = wert
    } else {
      Object.assign(saetze, flach(wert as object, `${pfad}.`))
    }
  }

  return saetze
}

// ueberlagere setzt die Sätze einer Übersetzung auf einen Katalog und gibt
// einen neuen; der Grund bleibt, wie er war.
//
// **Gesetzt wird nur, was der Grund als Satz kennt** — und die dritte
// Zählform `wenige` neben `eine` und `viele` (LOC-01). Ein Pfad aus einer
// neueren Fassung legt keinen Zweig an, den niemand liest.
//
// **Eine Zählform, deren Wörter die Übersetzung setzt, zählt nach ihrer
// Sprache** (`regel`, Befund N5 des Nachlaufs über Phase 6.5) — die übrigen
// nach der Basis, deren Wörter sie tragen.
export function ueberlagere<T extends object>(
  grund: T,
  texte: Record<string, string>,
  regel?: string,
): T {
  const ergebnis = structuredClone(grund) as Zweig
  const gezaehlt = new Set<Zweig>()

  for (const [pfad, text] of Object.entries(texte)) {
    const stelle = pfad.lastIndexOf('.')
    const blatt = pfad.slice(stelle + 1)
    const teile = stelle < 0 ? [] : pfad.slice(0, stelle).split('.')
    let zweig: unknown = ergebnis

    for (const teil of teile) {
      zweig = istZweig(zweig) ? zweig[teil] : undefined
    }

    if (!istZweig(zweig)) {
      continue
    }

    const zaehlform = 'eine' in zweig && 'viele' in zweig

    if (typeof zweig[blatt] === 'string' || (zaehlform && blatt === 'wenige')) {
      zweig[blatt] = text

      if (zaehlform) {
        gezaehlt.add(zweig)
      }
    }
  }

  if (regel !== undefined) {
    for (const form of gezaehlt) {
      Object.defineProperty(form, 'regel', { value: regel, enumerable: false })
    }
  }

  return ergebnis as T
}

// istZweig sagt, ob ein Wert ein Zweig des Katalogs ist.
function istZweig(wert: unknown): wert is Zweig {
  return typeof wert === 'object' && wert !== null
}
