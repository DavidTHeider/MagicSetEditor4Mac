// **Was ein losgelassener Zug für die Seite bedeutet** — PLAY-04.AK-03
// (#1712).
//
// # Die Weiche
//
// **Ein Zug ist entweder ein Befehl an den Wirt oder eine Lage.** Schieben,
// häufen und enthäufen sind Fachhandlungen und gehen denselben Weg wie die
// Handlungsleiste (§16). Eine Lage in Millimetern ist **keine**: CR 400.5
// lässt die Anordnung frei, der Wirt hat dazu nichts zu entscheiden, und ein
// Gegenüber bekommt davon nichts zu sehen.
//
// # Warum das hier steht und nicht im JSX
//
// **Weil es eine Regel ist und keine Darstellung.** Im Rückruf am
// `Tischflaeche` wäre sie nur über einen gezogenen Zeiger zu erreichen — und
// der kommt in einem Lauf ohne Grafikkarte nicht an: der Tisch weicht dann
// auf die Knopfliste aus, und die Zeigerkette liegt still. Ein Lauf, der das
// nicht sehen kann, prüft die Regel nie (§3).
import { befehlFuer } from './tischbefehle'

import type { Zugziel } from '../bausteine/tischziehen'

export interface Zughandlung {
  // lege hält eine Lage fest — sie bleibt auf diesem Gerät.
  lege: (kennung: string, stelle: { x: number; y: number }) => void

  // befehle stellt eine Fachhandlung an den Wirt.
  befehle: (art: string, nutzlast: Record<string, unknown>) => void

  // waehle setzt die offene Auswahl zurück — nach einem Zug ist sie nicht
  // mehr das, was jemand gemeint hat (§7, Regel 1).
  waehle: (kennung: string | undefined) => void
}

// zugBehandeln führt aus, was ein losgelassener Zug bedeutet.
export function zugBehandeln(ziel: Zugziel, handlung: Zughandlung): void {
  if (ziel.art === 'ablegen') {
    handlung.lege(ziel.kennung, { x: ziel.x, y: ziel.y })
    handlung.waehle(undefined)

    return
  }

  befehlFuer(ziel).forEach(([art, nutzlast]) => {
    handlung.befehle(art, nutzlast)
  })

  handlung.waehle(undefined)
}
