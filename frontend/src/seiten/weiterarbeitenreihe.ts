// **Eine Reihe statt vier Listen** — UXR-03
// (#1911).
//
// # Wofür
//
// Die Ablage antwortet in vier Teilen — angeheftet, zuletzt geöffnet,
// eingeführt, ausgegeben —, je bis zu sechs (`service.Hoechstens`). Die
// Startseite zeigte sie als vier Listen, und ein Set, das jemand geöffnet und
// ausgegeben hat, stand zweimal da. Das Artboard *Main* zeigt **eine Reihe**:
// „Angeheftetes steht oben. Die Reihe darunter folgt der Zeit."
//
// # Was die Reihe ist, und was nicht
//
// **Jedes Objekt einmal**, an der Stelle seines **jüngsten** Schritts, und mit
// diesem Schritt — was zuletzt mit ihm geschah, steht in der Zeile. Was
// angeheftet ist, steht oben und nicht noch einmal in der Zeit.
//
// **Das ist Anzeige und keine Fachlogik**: was in welchem Teil steht und was
// es nicht mehr gibt, entscheidet Go (`service.Weiterarbeiten`). Hier wird nur
// zusammengelegt, was schon da ist — kein Objekt kommt dazu, keines fällt weg.
//
// # Die Zeit wird gerechnet und nicht verglichen
//
// Ein Zeitpunkt kommt als RFC 3339 aus Go, mit Bruchteilen und Versatz, wie
// sie gerade sind. Zwei Zeichenketten zu vergleichen ordnete `…10:00:00.5Z`
// vor `…10:00:00Z` und jeden Versatz falsch; `Date.parse` rechnet beides.
// **Gleich alte Schritte bleiben in der Reihenfolge der Antwort.**
import type {
  ArbeitsobjektDTO,
  ArbeitsschrittDTO,
  WeiterarbeitenAntwort,
} from '../apiweiterarbeiten'

export interface Reiheneintrag {
  objekt: ArbeitsobjektDTO

  // schritt fehlt bei Angeheftetem: es steht oben, weil jemand es dort will,
  // nicht weil zuletzt etwas mit ihm geschah.
  schritt?: ArbeitsschrittDTO
}

// schluessel ist der Name eines Objekts über alle Arten — dieselbe Gestalt
// wie der Vorgangsschlüssel beim Anheften (§7, Regel 4).
export function schluessel(objekt: ArbeitsobjektDTO): string {
  return `${objekt.art}:${objekt.kennung}`
}

export function reihe(daten: WeiterarbeitenAntwort): Reiheneintrag[] {
  const gesehen = new Set(daten.angeheftet.map(schluessel))
  const zeitlich: Reiheneintrag[] = []

  const schritte = [
    ...daten.geoeffnet,
    ...daten.eingefuehrt,
    ...daten.ausgegeben,
  ]
    .map((schritt, stelle) => ({
      schritt,
      stelle,
      zeit: Date.parse(schritt.zeitpunkt),
    }))
    .sort((a, b) => b.zeit - a.zeit || a.stelle - b.stelle)

  for (const { schritt } of schritte) {
    const name = schluessel(schritt.objekt)

    if (!gesehen.has(name)) {
      gesehen.add(name)
      zeitlich.push({ objekt: schritt.objekt, schritt })
    }
  }

  return [...daten.angeheftet.map((objekt) => ({ objekt })), ...zeitlich]
}
