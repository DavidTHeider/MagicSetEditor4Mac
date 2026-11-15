// **Das Prüfmaß aus `docs/UIREVIEW.md`, unter jsdom gezählt** — UXR-11
// (#1919).
//
// # Wozu
//
// Die Hüllenprobe `uirevue` misst am nativen Fenster — mit dem Probebestand,
// und der hat kein Deck. Wie viele Bedienelemente eine Fläche mit **echten
// Mengen** zeichnet, sagt sie nicht. Diese Zählung nimmt **dieselben
// Wähler** wie das Drehbuch (`scripts/huellenprobe/uirevue.mjs`) und zählt in
// einem Teilbaum, den ein Test mit einem realistischen Bestand gebaut hat.
//
// # Was sie nicht kann
//
// **„Im Bild" nicht**: jsdom rechnet kein Layout. Gezählt wird, was
// **gezeichnet** ist — das ist R6, und für R2 eine obere Schranke. Verborgen
// heißt hier: unter einem `hidden` oder in einem geschlossenen Bereich, den
// React mit `display: none` aufhebt.

const bedienwaehler =
  'button, [role=button], a[href], [role=tab], [role=menuitem], summary'
const feldwaehler =
  'input:not([type=hidden]), select, textarea, [role=combobox], [role=switch]'
const regionswaehler =
  'section, form, aside, [role=region], [role=tabpanel], [role=dialog]'

export interface Zaehlung {
  bedien: number
  felder: number
  haupt: number
  hauptJeRegion: number
  ueberschriften: number
}

// gezeichnet sagt, ob ein Element nicht unter einem verborgenen steht.
function gezeichnet(element: Element): boolean {
  return element.closest('[hidden], [style*="display: none"]') === null
}

// zaehle misst einen Teilbaum nach dem Prüfmaß.
export function zaehle(wurzel: Element): Zaehlung {
  const bedien = [...wurzel.querySelectorAll(bedienwaehler)].filter(gezeichnet)
  const haupt = bedien.filter((knopf) =>
    /(^|\s)bg-marke(\s|$)/.test(knopf.className),
  )
  const regionen = [...wurzel.querySelectorAll(regionswaehler)]
  const jeRegion = regionen.map(
    (region) =>
      haupt.filter((knopf) => knopf.closest(regionswaehler) === region).length,
  )

  return {
    bedien: bedien.length,
    felder: [...wurzel.querySelectorAll(feldwaehler)].filter(gezeichnet).length,
    haupt: haupt.length,
    hauptJeRegion: Math.max(0, ...jeRegion),
    ueberschriften: [...wurzel.querySelectorAll('h1, h2, h3, h4')].filter(
      gezeichnet,
    ).length,
  }
}
