// **Wie eine Bildangabe heißt** — herausgezogen für Werkstück A12 von
// Phase 5.3
// (#1451).
//
// **Zwei Bauteile zeigen dieselben Angaben**: der Vorschaukasten vor der
// Übernahme und der Nachweis danach. Zwei Fassungen liefen beim ersten
// Sonderfall auseinander.
import type { kataloge } from '../sprache/kataloge'

type Einstellungstexte = (typeof kataloge)['de']['seiten']['einstellungen']

// beschriftung übersetzt den Namen einer Angabe — oder gibt ihn roh zurück.
//
// **Roh und nicht leer.** Ein Anbieter darf mehr nennen, als dieser Katalog
// kennt; den Namen dann zu verschlucken hieße, eine Auskunft zu verlieren,
// die jemand gebrauchen kann (§16, Drift).
export function beschriftung(name: string, eigene: Einstellungstexte): string {
  const bekannte: Record<string, string> = {
    groesse: eigene.kiBildAngabeGroesse,
    guete: eigene.kiBildAngabeGuete,
    umschrift: eigene.kiBildAngabeUmschrift,
    medientyp: eigene.kiBildAngabeMedientyp,
  }

  return bekannte[name] ?? name
}
