// **Was jemand hinlegt, bleibt liegen** — PLAY-04.AK-03
// (#1712).
//
// # Der Anlass
//
// **Heute liegt jede Karte auf einem gerechneten Platz.** `zonenlage.ts`
// legt Reihen, `schlachtfeldlagen` gruppiert nach CR 403.1 — und eine von
// Hand gewählte Lage gibt es nirgends. Wer seine Kreaturen nach
// Angriffsreihen sortiert, findet sie beim nächsten Zug neu sortiert vor.
//
// # Was „aus" heißt und was nicht
//
// **Nicht „ohne Ordnung", sondern „ohne Nachordnung".** Eine neu gezogene
// Karte hat keine eigene Lage, und die Rechnung ist die einzige, die ihr
// einen Platz geben kann. Diese Schicht **verschiebt** deshalb, statt zu
// legen: sie nimmt die gerechnete Lage und setzt die um, für die jemand
// etwas anderes gesagt hat.
//
// # Warum hier und nicht in `zonenlage.ts`
//
// **Weil es zwei Fragen sind.** Dort steht, **wie** eine Zone ihre Karten
// legt — Reihen, Haufen, Stapel; hier, **wessen** Wort schwerer wiegt. Wer
// beides vermischte, hätte in jeder Legefunktion ein `if` über eine
// Einstellung.
import type { Tischlage } from './tischplan'

// Stelle ist ein Punkt in Millimetern.
export interface Stelle {
  x: number
  y: number
}

// eigeneLagen setzt die Karten um, für die eine eigene Lage abgelegt ist.
//
// **Mitgenommen wird, was dazugehört** — das Kriterium sagt „berücksichtigt
// Anhänge und Gruppen":
//
//   - **Ein Haufen zieht ganz mit.** Blieben die unteren Blätter liegen,
//     wäre der Haufen auseinandergerissen — und der Versatz bleibt, denn
//     daran erkennt man ihn als Haufen.
//   - **Was an einer Karte hängt, zieht mit ihr** (CR 303.4): eine Aura
//     liegt bei dem, was sie verzaubert. Bliebe sie stehen, sähe man nicht
//     mehr, woran sie hängt.
//   - **Eine eigene Lage am Anhängsel schlägt die seines Wirts** — wer eine
//     Aura selbst hinlegt, meint es so.
export function eigeneLagen(
  gerechnet: readonly Tischlage[],
  eigene: Readonly<Record<string, Stelle>>,
): Tischlage[] {
  // **Die Verschiebung je Karte, nicht die Zielstelle.** Ein Haufen und ein
  // Anhängsel folgen ihrem Anker um **denselben Weg**; wer die Zielstelle
  // weiterreichte, legte sie aufeinander.
  const versatz = new Map<string, Stelle>()

  for (const lage of gerechnet) {
    const stelle = eigene[lage.kennung]

    if (stelle !== undefined) {
      versatz.set(lage.kennung, { x: stelle.x - lage.x, y: stelle.y - lage.y })
    }
  }

  // **Der Haufen folgt dem Blatt, für das eine Lage abgelegt ist.**
  //
  // **Verschoben und nicht gelegt**: der Versatz zwischen den Blättern
  // bleibt damit von selbst erhalten — `zonenlage` hat ihn gerechnet, und
  // diese Schicht fasst ihn nicht an.
  const jeHaufen = new Map<string, Stelle>()

  // **Erst holen, dann fragen** — ein `has` und danach ein `get` mit
  // Rückfall wären zwei Griffe für eine Frage, und der Rückfall ein Zweig,
  // den kein Aufruf erreicht (§3). Die Abdeckung hat ihn angestrichen.
  for (const lage of gerechnet) {
    const eigener = versatz.get(lage.kennung)

    if (lage.haufen !== '' && eigener !== undefined) {
      jeHaufen.set(lage.haufen, eigener)
    }
  }

  return gerechnet.map((lage) => {
    const eigener =
      versatz.get(lage.kennung) ??
      (lage.haufen === '' ? undefined : jeHaufen.get(lage.haufen)) ??
      (lage.haengtAn === '' ? undefined : versatz.get(lage.haengtAn))

    if (eigener === undefined) {
      return lage
    }

    return { ...lage, x: lage.x + eigener.x, y: lage.y + eigener.y }
  })
}
