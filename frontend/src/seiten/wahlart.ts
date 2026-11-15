// **Welches Bedienelement eine Einstellung bekommt** — UXR-F2
// (#2008),
// Regel R7 aus `docs/UIREVIEW.md`: *Schalter für Sofortwirkung, Segmente bis
// vier Werte, Auswahl darüber.*
//
// # Die Zahl der Werte entscheidet
//
// Gezählt wird, was zur Wahl steht — **mit dem Erbwert**: wer `erben` wählen
// kann, wählt es als ein Wert unter anderen. Drei Werte mit `erben` sind darum
// drei Segmente, und kein Schalter mit einem versteckten dritten Zustand.
//
// | Werte | Bedienelement |
// | --- | --- |
// | keine | Textfeld — ein Verzeichnis oder eine Adresse lässt sich nicht aufzählen |
// | zwei, die an und aus sind | Schalter |
// | zwei bis vier sonst | Segmente |
// | ein Wert oder mehr als vier | Auswahlliste |
//
// # Warum nicht jedes Paar ein Schalter ist
//
// **Ein Schalter sagt „an" oder „aus", und das muss für beide Werte stimmen.**
// `senkrecht` und `waagerecht`, `text` und `json`, `system` und `reduziert`
// sind zwei Zustände und keiner davon der ausgeschaltete; als Schalter wüsste
// niemand, was „an" heißt. Sie stehen als zwei Segmente — R7 lässt Segmente
// bis vier Werte zu, und zwei gehören dazu.
//
// **Die Paare stehen hier und nicht beim Wert**: Go kennt kein „ist ein
// Schalter", nur eine Werteliste. Ein neues Paar trägt jemand hier ein — und
// `wahlkatalog.test.tsx` sagt, welche Einstellung dadurch ihr Element
// wechselt.

// Schalterpaar sagt, welcher von zwei Werten angeschaltet heißt.
export interface Schalterpaar {
  an: string
  aus: string
}

// Bedienung ist, wie eine Einstellung gezeichnet wird — und beim Schalter,
// welcher Wert „an" ist.
export type Bedienung =
  | { art: 'text' }
  | { art: 'schalter'; paar: Schalterpaar }
  | { art: 'segmente' }
  | { art: 'auswahl' }

// segmentgrenze ist die größte Zahl von Werten, die noch Segmente sind — R7.
export const segmentgrenze = 4

// schalterpaare sind die Wörter, mit denen eine Einstellung an und aus heißt.
const schalterpaare: readonly Schalterpaar[] = [
  { an: 'an', aus: 'aus' },
  { an: 'sichtbar', aus: 'verborgen' },
]

// paarVon gibt das Paar, wenn die Werte genau eines sind.
function paarVon(werte: readonly string[]): Schalterpaar | undefined {
  if (werte.length !== 2) {
    return undefined
  }

  return schalterpaare.find(
    (paar) => werte.includes(paar.an) && werte.includes(paar.aus),
  )
}

// bedienung wählt das Bedienelement nach den Werten.
export function bedienung(werte: readonly string[]): Bedienung {
  if (werte.length === 0) {
    return { art: 'text' }
  }

  const paar = paarVon(werte)
  if (paar !== undefined) {
    return { art: 'schalter', paar }
  }

  return werte.length >= 2 && werte.length <= segmentgrenze
    ? { art: 'segmente' }
    : { art: 'auswahl' }
}
