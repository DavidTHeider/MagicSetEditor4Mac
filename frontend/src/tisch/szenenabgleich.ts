// **Was sich seit dem letzten Bild geändert hat** — Werkstück G3 von Phase
// 5.5 (#1460).
//
// # Warum abgeglichen und nicht neu gebaut
//
// **Weil ein Sprite teuer ist und eine Textur teurer.** Ein Tisch mit vier
// Spielern trägt nach der Leistungsvorgabe von #1460 hundertfünfzig Karten,
// hundert Marken und fünfzig Tokens; sie bei jedem React-Render neu anzulegen
// hieße, dreihundert Objekte je Zustandswechsel wegzuwerfen und
// wiederzubeschaffen.
//
// **Und weil es ohne Abgleich keine Bewegung gäbe.** Ein Sprite, das
// stehenbleibt, kann von seinem alten Platz zum neuen fliegen; eines, das neu
// entsteht, erscheint.

// Kenntlich ist alles, was eine Kennung trägt.
export interface Kenntlich {
  kennung: string
}

export interface Abgleich<T extends Kenntlich> {
  // neu ist, was es beim letzten Bild noch nicht gab.
  neu: T[]

  // bleibt ist, was schon da war — es bekommt einen neuen Platz, kein neues
  // Leben.
  bleibt: T[]

  // weg sind die Kennungen dessen, was verschwunden ist.
  weg: string[]
}

export function abgeglichen<T extends Kenntlich>(
  vorher: Set<string>,
  jetzt: T[],
): Abgleich<T> {
  const neu: T[] = []
  const bleibt: T[] = []
  const gesehen = new Set<string>()

  for (const eines of jetzt) {
    gesehen.add(eines.kennung)
    if (vorher.has(eines.kennung)) bleibt.push(eines)
    else neu.push(eines)
  }

  return {
    neu,
    bleibt,
    weg: [...vorher].filter((eine) => !gesehen.has(eine)),
  }
}
