// **Vergleichen und Kleinschreiben in der Sprache des Fensters** — F01
// (#2243).
//
// **Die eine Naht** für Menschentext (`docs/LOKALISIERUNG.md`): wer ordnet
// oder für eine Suche faltet, nennt die Sprache. `localeCompare` und
// `toLocaleLowerCase` ohne sie nehmen die des Systems — auf einem englischen
// Mac ordnete eine japanische Oberfläche wie eine englische. Ein Wächter
// (`sprachnaht.test.ts`) hält, dass außer hier keine Stelle sie weglässt.
//
// **Technische Schlüssel** gehören nicht hierher: sie ordnen nach Codepunkten
// und auf jedem Rechner gleich (`ordnung.ts`).

const vergleicherJeSprache = new Map<string, (a: string, b: string) => number>()

// vergleicher gibt die Ordnung einer Sprache — einmal gebaut je Sprache, denn
// ein `Intl.Collator` ist teuer und eine Liste ruft ihn für jedes Paar.
export function vergleicher(sprache: string): (a: string, b: string) => number {
  const bekannt = vergleicherJeSprache.get(sprache)
  if (bekannt !== undefined) {
    return bekannt
  }

  const neu = new Intl.Collator(sprache).compare
  vergleicherJeSprache.set(sprache, neu)

  return neu
}

// falte schreibt für eine Suche klein, in der Sprache des Fensters.
export function falte(text: string, sprache: string): string {
  return text.toLocaleLowerCase(sprache)
}
