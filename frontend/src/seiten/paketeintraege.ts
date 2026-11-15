// **Die Handlungen an einem Paket** — für das Menü der Zeile **und** den
// Rechtsklick an ihr (AP-5,
// #2362).
// Eine Liste an einer Stelle: sonst bekäme eine der beiden Wege einen
// Eintrag, den der andere nicht kennt — und keine Aktion soll es nur im
// Kontextmenü geben (U2, #1413).
//
// **Das Menü löscht nicht selbst**: Löschen und Zurück öffnen den Plan im
// Inspector, und er ist die Rückfrage (DOM-01, #1696).
import type { PaketDTO } from '../api'
import type { Menueeintrag } from '../bausteine/Menue'
import type { Katalog } from '../sprache/kataloge'

export interface Paketgriffe {
  markenOffen: string | null
  setzeMarkenOffen: (kennung: string | null) => void
  planen: (vorgang: 'entfernen' | 'zurueck', paket: PaketDTO) => void
  laeuft: boolean
}

export function paketkennung(paket: PaketDTO): string {
  return `${paket.art}/${paket.name}`
}

export function paketeintraege(
  paket: PaketDTO,
  texte: Katalog,
  griffe: Paketgriffe,
): Menueeintrag[] {
  const kennung = paketkennung(paket)

  return [
    {
      schluessel: 'marken',
      schreibt: true,
      name: texte.bausteine.objektmarken.ueberschrift,
      aufWahl: () => {
        griffe.setzeMarkenOffen(griffe.markenOffen === kennung ? null : kennung)
      },
    },
    {
      schluessel: 'loeschen',
      name: texte.seiten.pakete.plan.loeschen,
      aktion: 'loeschen',
      angehalten: griffe.laeuft,
      aufWahl: () => {
        griffe.planen('entfernen', paket)
      },
    },
    {
      schluessel: 'zurueck',
      name: texte.seiten.pakete.plan.zurueck,
      angehalten: griffe.laeuft,
      aufWahl: () => {
        griffe.planen('zurueck', paket)
      },
    },
  ]
}
