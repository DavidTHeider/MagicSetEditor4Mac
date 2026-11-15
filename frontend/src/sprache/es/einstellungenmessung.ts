// **The measurement settings** — OPS-05
// (#1722).
export const messnamen = {
  'messung.spur': 'Traza',
  'messung.abtastung': 'Proporción de llamadas registradas, en porcentaje',
  'messung.spurgrenze': 'Como máximo este número de entradas',
  'messung.aufbewahrung': 'Las entradas se conservan este número de minutos',
  'messung.profile': 'Perfiles de rendimiento que se conservan',
  'zwischenspeicher.grenze': 'Límite por caché en disco, en MB',
} as const

export const messwerte = {
  'messung.spur.aus': 'Desactivada',
  'messung.spur.an': 'Activada',
} as const

export const messhinweise = {
  'messung.spur':
    'Una línea por cada llamada al servicio y al servidor de herramientas, con identificador de operación, ruta, duración y resultado, sin contenidos. Vive en la memoria del proceso y se aplica desde el próximo inicio.',
  'messung.abtastung':
    'Con 10, la traza registra una de cada diez llamadas, contadas con exactitud, no al azar. Vacío equivale a 100.',
  'messung.spurgrenze':
    'Cuando la traza se llena, se descarta la entrada más antigua. De 10 a 10000; vacío equivale a 1000.',
  'messung.aufbewahrung':
    'Las entradas más antiguas se descartan. De 1 a 1440 minutos; vacío equivale a 60.',
  'messung.profile':
    'Los más recientes se quedan junto a la biblioteca; los más antiguos se van con el siguiente perfil. De 1 a 20; vacío equivale a 3.',
  'zwischenspeicher.grenze':
    'Para las respuestas y las imágenes de Scryfall, cada una por separado. Al iniciar se descartan las más antiguas hasta que quepa. De 1 a 8192; vacío equivale a 512.',
} as const
