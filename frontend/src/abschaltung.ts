// **Was ein abgeschalteter Bereich im Fenster verbirgt** — OPS-10
// (#1908,
// Vermerk 0158).
//
// # Warum über die Namen und nicht über eine Liste
//
// **Weil die Liste in Go steht** (`internal/domain/schalter.go`), und eine
// zweite hier liefe auseinander. Die Regel ist eine über Namen: ein Schalter,
// der heißt wie ein Bereich, verbirgt dessen Seiten; einer, der heißt wie
// eine Seite, verbirgt diese. `internal/werkzeuge/schalterwege_test.go` hält,
// dass jeder Bereichsschalter einen Bereich dieser Leiste trifft.
//
// # Was er nicht verbirgt
//
// **Einen Knopf auf einer Seite des Kerns** — etwa den KI-Vorschlag an der
// Karte. Er bleibt stehen und bekommt beim Drücken die Abweisung samt dem
// Weg zurück; verborgen werden Seiten, nicht einzelne Handlungen.
import { bereiche } from './bereiche'
import { useEinstellungsraum } from './haken/einstellungsraum'

import type { Seitenschluessel } from './navigation'

const aus = 'aus'

// verborgeneSeiten gibt die Seiten, die abgeschaltet sind.
export function verborgeneSeiten(
  wert: (schluessel: string) => string,
): ReadonlySet<Seitenschluessel> {
  const istAus = (name: string) => wert(`schalter.${name}`) === aus
  const verborgen = new Set<Seitenschluessel>()

  // **Die Dienstzone nie** — sie ist der Weg zurück.
  for (const bereich of bereiche.filter((einer) => !einer.dienst)) {
    for (const seite of bereich.seiten) {
      if (istAus(bereich.schluessel) || istAus(seite)) {
        verborgen.add(seite)
      }
    }
  }

  return verborgen
}

// useVerborgeneSeiten gibt die verborgenen Seiten nach dem geteilten Stand —
// **ohne Raum keine**: ein Lauf, der eine Fläche allein aufbaut, sieht alles.
export function useVerborgeneSeiten(): ReadonlySet<Seitenschluessel> {
  const stand = useEinstellungsraum()

  return verborgeneSeiten((schluessel) => stand?.wert(schluessel) ?? '')
}
