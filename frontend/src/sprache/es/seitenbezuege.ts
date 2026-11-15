// Los textos de los **estados obtenidos** en la página de paquetes — G04 (#2251).
// Su fuente alemana es `de/seitenbezuege.ts`.

export const bezuege = {
  ueberschrift: 'Estados obtenidos',
  erklaerung:
    'Un paquete obtenido se queda en un commit y no envejece por sí solo. Pregunta a las fuentes si hay uno más nuevo; solo entonces sale algo por la red.',
  fragen: 'Preguntar a las fuentes',
  leer: 'Todavía no has obtenido ningún paquete. Obtén uno con «Obtener paquete …»; después verás aquí cuánto ha avanzado su fuente.',
  bezogen: 'Obtenido: {stand}',
  geprueft: 'Comprobado: {stand}',
  ohnePruefung: 'No está en la lista de fuentes comprobadas',
  pakete: '{zahl} paquetes',
  paketEins: 'un paquete',
  keinPaket: 'ningún paquete',
  lage: {
    ungefragt: 'Aún no se ha preguntado.',
    aktuell: 'En el estado más reciente de la fuente.',
    neuerEins:
      'Estado más reciente {spitze} del {datum}: un commit por delante.',
    neuer:
      'Estado más reciente {spitze} del {datum}: {vorsprung} commits por delante.',
    abweichend:
      'La fuente no va por delante: restablecida o divergente (punta {spitze}).',
    unbekannt: 'No se pudo preguntar a la fuente: {grund}',
  },
  ansehenBezeichnung: 'Revisar la actualización … ({name})',
  zurueckBezeichnung: 'Volver al estado guardado ({name}, {stand})',
  ansehen: 'Revisar la actualización …',
  aufgehoben: 'Guardado: {stand}',
  zurueck: 'Volver al estado guardado',
  vorschauTitel: 'Actualización de {name}',
  vonZu: '{von} → {zu}: {vorsprung} por delante, {rueckstand} por detrás',
  paketzeile:
    '{paket}: {neu} nuevos, {geaendert} cambiados, {entfernt} eliminados',
  keinePakete: 'No cambia ningún paquete de plantillas.',
  uebrige: '{zahl} archivos más no pertenecen a ningún paquete.',
  sets: 'Estos sets nombran un paquete afectado: {sets}',
  keineSets: 'Ningún set nombra un paquete afectado.',
  aufheben: 'El estado actual se guarda y se puede recuperar.',
  vorigeWeicht: 'El estado ya guardado {stand} se sustituye.',
  gruende: {
    setVerliert: 'Un paquete que nombra un set pierde archivos.',
    nichtVoraus:
      'La fuente no va por delante: la actualización sería un paso atrás o hacia un lado.',
    gekuerzt:
      'La fuente indica más de 300 archivos cambiados; la lista de arriba está recortada.',
  },
  bestaetigen: 'He visto los motivos y aun así quiero actualizar.',
  ausfuehren: 'Actualizar a {zu}',
  verwerfen: 'No actualizar',
}
