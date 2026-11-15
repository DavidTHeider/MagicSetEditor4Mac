// **The state of a file run** — FMT-03.AK-02 (#1692).
export const laufstand = {
  vonGesamt: '{erledigt} de {gesamt} {einheit}',
  einheitenVon: {
    karte: { eine: 'carta', viele: 'cartas' },
    set: { eine: 'set', viele: 'sets' },
    paket: { eine: 'paquete', viele: 'paquetes' },
    bild: { eine: 'imagen', viele: 'imágenes' },
    schritt: { eine: 'paso', viele: 'pasos' },
  },
  einheiten: {
    karte: { eine: 'carta', viele: 'cartas' },
    set: { eine: 'set', viele: 'sets' },
    paket: { eine: 'paquete', viele: 'paquetes' },
    bild: { eine: 'imagen', viele: 'imágenes' },
    schritt: { eine: 'paso', viele: 'pasos' },
  },
  laeuft: 'En curso …',
  abbrechen: 'Cancelar',
  wirdAngehalten: 'Deteniendo …',
  abgebrochen:
    'Cancelado. Lo que ya estaba completo se conserva: no queda nada a medio escribir.',
  bisDahin: 'Hasta ese momento: {stand}.',
} as const
