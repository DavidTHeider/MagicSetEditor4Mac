// **Die eine axe-Abfrage** — herausgelöst am 2026-09-07, als eine zweite
// Datei sie brauchte (Werkstück 5 von Phase 4.13,
// #1352).
//
// # Warum sie an einer Stelle steht
//
// **Weil zwei Fassungen auseinanderliefen.** Die Liste der WCAG-Marken ist die
// Zusage; stünde sie zweimal da, prüfte die eine Seite bald etwas anderes als
// die andere — und die schwächere sähe genauso grün aus.
//
// # Was axe unter jsdom nicht sieht
//
// `color-contrast` gilt dort als **unentschieden**: ohne Layout gibt es keine
// gerechnete Farbe. Die Kontraste werden im Browser nachgemessen und stehen in
// `docs/KONTRASTE.md`; dieser Weg löst das **nicht** ein.
import axe from 'axe-core'

// verstoesse gibt die Verstöße als lesbare Zeilen — nicht als Objektgraphen.
//
// Eine fehlgeschlagene Zusicherung, die einen axe-Bericht ausgibt, ist
// unlesbar; hier steht die Regel und das Element, und das reicht zum Finden.
export async function verstoesse(bereich: HTMLElement): Promise<string[]> {
  const bericht = await axe.run(bereich, {
    runOnly: {
      type: 'tag',
      values: ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa', 'wcag22aa'],
    },
  })

  return bericht.violations.map(
    (verstoss) =>
      `${verstoss.id}: ${verstoss.help} — ${verstoss.nodes.map((knoten) => knoten.target.join(' ')).join(', ')}`,
  )
}
