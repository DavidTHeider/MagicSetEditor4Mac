// **The state of a file run** — FMT-03.AK-02 (#1692).
export const laufstand = {
  vonGesamt: '{erledigt} sur {gesamt} {einheit}',
  einheitenVon: {
    karte: { eine: 'carte', viele: 'cartes' },
    set: { eine: 'set', viele: 'sets' },
    paket: { eine: 'paquet', viele: 'paquets' },
    bild: { eine: 'image', viele: 'images' },
    schritt: { eine: 'étape', viele: 'étapes' },
  },
  einheiten: {
    karte: { eine: 'carte', viele: 'cartes' },
    set: { eine: 'set', viele: 'sets' },
    paket: { eine: 'paquet', viele: 'paquets' },
    bild: { eine: 'image', viele: 'images' },
    schritt: { eine: 'étape', viele: 'étapes' },
  },
  laeuft: 'En cours …',
  abbrechen: 'Annuler',
  wirdAngehalten: 'Arrêt en cours …',
  abgebrochen:
    'Annulé. Ce qui était déjà terminé est conservé — rien n’est écrit à moitié.',
  bisDahin: 'Jusque-là : {stand}.',
} as const
