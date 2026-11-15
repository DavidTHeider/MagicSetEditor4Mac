// **Ein Griff auf die Zwischenablage für Tests** (#1168).
//
// **Geteilt und nicht je Datei nachgebaut**: eine Vorrichtung, die an vier
// Stellen steht, ist an vier Stellen zu ändern — und drei davon vergisst man.
import { vi } from 'vitest'

import type { Kartenablagegriff } from '../haken/useKartenablage'

export function ablagedoppel(
  abweichung: Partial<Kartenablagegriff> = {},
): Kartenablagegriff {
  return {
    kopieren: vi.fn(),
    ausschneiden: vi.fn(),
    einfuegen: vi.fn(),
    laeuft: false,
    anzahl: 0,
    ...abweichung,
  }
}
