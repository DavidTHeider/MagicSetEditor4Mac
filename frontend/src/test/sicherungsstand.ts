// Ein Stand der Sicherung für Tests — OPS01
// (#2347).
import type { Sicherungsstand } from '../antwortenbestand'

// standMit baut einen Stand — ohne Plan, wenn `plan.ordner` leer bleibt.
export function standMit(
  teile: Partial<Sicherungsstand> = {},
): Sicherungsstand {
  return {
    stufe: 'unbekannt',
    gesichert: '',
    tage: 0,
    anlass: '',
    flaeche: '',
    datei: '',
    verschluesselt: false,
    plan: { ordner: '', letzter: '', faellig: false },
    ...teile,
  }
}
