// **Die Meldung des Fensters und ihr Kreuz** — herausgelöst aus
// `Anwendung.tsx`, als Befund G5 des Zweitblicks über Phase 6.3
// (#1665)
// eine weitere Meldung ins Band brachte und die Datei an die Zeilengrenze
// (§2).
//
// # Was hinzukam
//
// **Das gescheiterte »Darstellung zurücksetzen« aus der Menüleiste.** Es
// stand nur unter *Einstellungen › Erscheinungsbild* — wer die Menüleiste
// nimmt, steht gerade nicht dort. Der Stand kommt aus dem Gestaltungsraum,
// denn dort läuft die Rücksetzung; welche Meldung den Vortritt hat, bleibt
// eine Frage an `bandmeldung.ts`.
//
// # Ein Kreuz räumt alles
//
// **Wie bisher** (#799): welche Meldung gerade dasteht, weiß der Knopf nicht
// und soll es nicht wissen müssen. Was nicht dasteht, verschwindet nicht.
import { useCallback } from 'react'

import { bandmeldung, type Meldungslage } from '../bandmeldung'
import { useGestaltung } from '../gestaltungsraum'
import type { Bandmeldung } from '../rahmen/Meldungsband'

export interface Meldungsbandstand {
  meldung: Bandmeldung | undefined
  raeumen: () => void
}

export function useMeldungsband(
  lage: Omit<Meldungslage, 'darstellungsfehler'>,
  raeumen: () => void,
): Meldungsbandstand {
  const gestaltung = useGestaltung()
  const darstellungRaeumen = gestaltung?.ruecksetzfehlerRaeumen

  return {
    meldung: bandmeldung({
      ...lage,
      darstellungsfehler: gestaltung?.ruecksetzfehler,
    }),
    raeumen: useCallback(() => {
      raeumen()
      darstellungRaeumen?.()
    }, [raeumen, darstellungRaeumen]),
  }
}
