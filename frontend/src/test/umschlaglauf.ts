// Ein Umschlaglauf für Tests, der nichts tut.
//
// **Dieselbe Gestalt wie `stillerSpeicherlauf` daneben** und aus demselben
// Grund: `Setkarte` bekommt den Lauf von außen (§7, Regel 4), und ein Test,
// der nur die Karte zeigt, braucht trotzdem einen.
import { vi } from 'vitest'

import type { Umschlaglauf } from '../haken/useUmschlag'

export function stillerUmschlaglauf(
  teile: Partial<Umschlaglauf> = {},
): Umschlaglauf {
  return {
    schnueren: vi.fn(),
    waehlen: vi.fn(),
    uebernehmen: vi.fn(),
    raeumen: vi.fn(),
    laeuft: () => false,
    geschnuert: undefined,
    fuer: undefined,
    angesehen: undefined,
    uebernommen: undefined,
    fehler: undefined,
    ...teile,
  }
}
