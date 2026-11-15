// **Wie die Ziele des Ausgabeassistenten heißen** — UXR-09
// (#1917).
//
// **Unter dem Namen, unter dem man sie schon kennt.** *Bilder* und
// *Druckbogen* heißen wie ihre Seiten — dieselben Namen tragen Palette,
// Kürzeltafel und Menü —, die Exportvorlage und MTG JSON wie auf der
// Setliste. Neu ist nur die *Setdatei*: sie hatte nie einen eigenen Namen,
// nur einen Satz.
import type { Ausgabeziel } from '../ausgabe'
import type { Katalog } from '../sprache/kataloge'

// zielname gibt den Namen eines Ziels.
export function zielname(texte: Katalog, ziel: Ausgabeziel): string {
  switch (ziel) {
    case 'setdatei':
      return texte.seiten.ausgeben.ziele.setdatei.name
    case 'bilder':
      return texte.seiten.bilder.titel
    case 'druckbogen':
      return texte.seiten.druckbogen.titel
    case 'vorlage':
      return texte.seiten.sets.vorlageWaehlen
    case 'mtgjson':
      return texte.seiten.sets.mtgjsonUeberschrift
  }
}

// zielerklaerung gibt, was unter dem Namen in der Fläche eines Ziels steht —
// **nicht** der Satz aus der Wahl, der steht schon links daneben. Bilder und
// Druckbogen bringen ihre Erklärung als Seite mit.
export function zielerklaerung(texte: Katalog, ziel: Ausgabeziel): string {
  switch (ziel) {
    case 'bilder':
      return texte.seiten.bilder.erklaerung
    case 'druckbogen':
      return texte.seiten.druckbogen.erklaerung
    default:
      return texte.seiten.ausgeben.ziele[ziel].lang
  }
}
