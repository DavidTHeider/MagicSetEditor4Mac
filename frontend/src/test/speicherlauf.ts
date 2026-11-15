// Ein Speicherlauf für Tests, der nichts tut.
//
// **Als Vorrichtung und nicht als zweiter Haken.** Seit der Zweitblick über
// Phase 6 die zweite Vorgangsliste angestrichen hat, bekommt `Sets` den Lauf
// von außen (#162); ein Test, der die Seite allein zeigt, braucht trotzdem
// einen — und dieser sagt, dass er nichts tut, statt es zu verschweigen.
import { vi } from 'vitest'

import type { Speicherlauf } from '../haken/useSetSpeichern'

export function stillerSpeicherlauf(
  teile: Partial<Speicherlauf> = {},
): Speicherlauf {
  return {
    speichern: vi.fn(),
    laeuft: () => false,
    speicherung: undefined,
    fehler: undefined,
    fuer: undefined,
    raeumen: vi.fn(),
    ...teile,
  }
}
