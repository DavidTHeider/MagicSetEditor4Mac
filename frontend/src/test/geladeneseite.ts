// **Eine Seite, die steht** — für die Wächter über `seiten` (G02, #2249).
import type { Seitenschluessel } from '../navigation'
import type { Sprache } from '../sprache/kataloge'
import { antwortAbwarten } from './api'
import { dokumentDer } from './seitenvorrichtung'

// zurRuhe wartet, bis eine gezeichnete Seite steht — **keine Antwort mehr
// unterwegs** (zwei Runden lang derselbe Baum) und nichts mit `aria-busy`.
//
// **Gewartet wird auf die Wirkung** (§3): eine Antwort kommt hier eine
// Makroaufgabe später, und eine Seite, die erst das Set und dann seine
// Karten liest, braucht mehrere. Die Wächter über `seiten` lasen bis G02
// gleich nach dem Zeichnen oder nach **einer** Runde — sie prüften den
// Ladezustand und nicht die Seite.
export async function zurRuhe(
  wurzel: HTMLElement,
  wofuer: string,
): Promise<HTMLElement> {
  let vorher = ''

  for (let runde = 0; runde < 40; runde++) {
    await antwortAbwarten()

    const jetzt = wurzel.innerHTML

    if (
      jetzt === vorher &&
      wurzel.querySelector('[aria-busy="true"]') === null
    ) {
      return wurzel
    }

    vorher = jetzt
  }

  throw new Error(`${wofuer} kommt nicht zur Ruhe`)
}

// geladenesDokumentDer zeichnet wie [dokumentDer] und wartet [zurRuhe].
export async function geladenesDokumentDer(
  seite: Seitenschluessel,
  sprache: Sprache,
): Promise<HTMLElement> {
  return zurRuhe(dokumentDer(seite, sprache), seite)
}
