// **The state of a file run** — FMT-03.AK-02 (#1692).
export const laufstand = {
  vonGesamt: '{erledigt} di {gesamt} {einheit}',
  einheitenVon: {
    karte: { eine: 'carta', viele: 'carte' },
    set: { eine: 'set', viele: 'set' },
    paket: { eine: 'pacchetto', viele: 'pacchetti' },
    bild: { eine: 'immagine', viele: 'immagini' },
    schritt: { eine: 'passaggio', viele: 'passaggi' },
  },
  einheiten: {
    karte: { eine: 'carta', viele: 'carte' },
    set: { eine: 'set', viele: 'set' },
    paket: { eine: 'pacchetto', viele: 'pacchetti' },
    bild: { eine: 'immagine', viele: 'immagini' },
    schritt: { eine: 'passaggio', viele: 'passaggi' },
  },
  laeuft: 'In corso …',
  abbrechen: 'Annulla',
  wirdAngehalten: 'Interruzione in corso …',
  abgebrochen:
    'Annullato. Ciò che era già completo resta — non rimane nulla di scritto a metà.',
  bisDahin: 'Fino a quel punto: {stand}.',
} as const
