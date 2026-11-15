// **The state of a file run** — FMT-03.AK-02 (#1692).
export const laufstand = {
  vonGesamt: '{erledigt} of {gesamt} {einheit}',
  einheitenVon: {
    karte: { eine: 'card', viele: 'cards' },
    set: { eine: 'set', viele: 'sets' },
    paket: { eine: 'package', viele: 'packages' },
    bild: { eine: 'image', viele: 'images' },
    schritt: { eine: 'step', viele: 'steps' },
  },
  einheiten: {
    karte: { eine: 'card', viele: 'cards' },
    set: { eine: 'set', viele: 'sets' },
    paket: { eine: 'package', viele: 'packages' },
    bild: { eine: 'image', viele: 'images' },
    schritt: { eine: 'step', viele: 'steps' },
  },
  laeuft: 'Running …',
  abbrechen: 'Cancel',
  wirdAngehalten: 'Stopping …',
  abgebrochen:
    'Cancelled. What was already complete stays — there is nothing half written.',
  bisDahin: 'Up to then: {stand}.',
} as const
