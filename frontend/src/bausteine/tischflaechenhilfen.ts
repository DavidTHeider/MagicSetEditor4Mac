// **Was die Tischfläche rechnet, ohne etwas zu zeigen** — Tasten, Zeiger,
// Mitte und der Satz zu einem Hindernis.
//
// Eine eigene Datei, seit `Tischflaeche.tsx` mit der eigenen Leinwand je
// Anlauf (UI-N01, #2089) über 400 Zeilen ging (§2).
import { nachMillimetern } from '../tisch/kamera'
import type { Kamera } from '../tisch/kamera'
import type { useTexte } from '../sprache/kontext'
import type { useTisch } from '../tisch/useTisch'

// Spieltexte ist der Katalog der Seite — mit seinen Feldern.
type Spieltexte = ReturnType<typeof useTexte>['seiten']['spielen']

// griff bedient den Tisch über die Tastatur.
//
// **Jede Maushandlung hat hier ihre Taste** (§6): schwenken mit den Pfeilen,
// näher und weiter mit Plus und Minus, alles zeigen mit der Null.
export function griff(taste: string, tisch: ReturnType<typeof useTisch>): void {
  const schritt = 40
  const mitte = mitteVon(tisch.masse)

  const schwenke: Record<string, [number, number]> = {
    ArrowLeft: [schritt, 0],
    ArrowRight: [-schritt, 0],
    ArrowUp: [0, schritt],
    ArrowDown: [0, -schritt],
  }

  const richtung = schwenke[taste]
  if (richtung) {
    tisch.schwenke(richtung[0], richtung[1])

    return
  }

  if (taste === '+') tisch.zoome(1.25, mitte)
  if (taste === '-') tisch.zoome(0.8, mitte)
  if (taste === '0') tisch.passeEin()
}

// amZeiger rechnet einen Zeigerpunkt in den Ort auf dem Tisch um.
//
// **Über die Kamera und nicht über das Element.** Der Punkt kommt in
// Bildschirmpunkten relativ zur Leinwand; was er auf dem Tisch bedeutet, hängt
// an Maßstab, Schwenk und Drehung — und die kennt nur die Kamera.
export function amZeiger(
  ereignis: React.PointerEvent<HTMLElement>,
  kamera: Kamera,
  masse: { breite: number; hoehe: number },
) {
  const kasten = ereignis.currentTarget.getBoundingClientRect()

  return nachMillimetern(
    kamera,
    { x: ereignis.clientX - kasten.left, y: ereignis.clientY - kasten.top },
    masse,
  )
}

export const mitteVon = (masse: { breite: number; hoehe: number }) => ({
  x: masse.breite / 2,
  y: masse.hoehe / 2,
})

export function hindernissatz(eigene: Spieltexte, hindernis: string): string {
  switch (hindernis) {
    case 'kein-webgl':
      return eigene.tischOhneWebgl
    case 'verloren':
      return eigene.tischVerloren
    default:
      return eigene.tischOhneRenderer
  }
}
