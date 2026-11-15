// **Run times, consumption and the performance profile** — OPS-05
// (#1722).

export const betrieb = {
  ueberschrift: 'Tiempos de ejecución y consumo',
  zeitraum: 'Recopilado de {seit} a {bis}.',
  aufwand: 'Coste de la recopilación: {mikros} µs',
  speicher: 'Memoria',
  speicherWert:
    '{fussabdruck} de huella — de ellos {heap} de heap de Go, {system} del sistema',
  speicherOhneFussabdruck:
    '{heap} de heap de Go, {system} del sistema — la huella no se ha medido',
  web: 'Procesos web',
  webWert: '{summe} en total — {teile}',
  webTeil: '{art} {groesse}',
  webArten: { webinhalt: 'Contenido web', gpu: 'Gráficos', netz: 'Red' },
  ohneWeb: 'Este proceso no tiene ninguna vista web.',
  webUngemessen: 'Sin medir: {grund}',
  goroutinen: 'Rutinas concurrentes',
  rechenzeit: 'Tiempo de CPU desde el inicio',
  datenbank: 'Conexiones con la base de datos',
  datenbankWert:
    '{offen} abiertas, {inBenutzung} en uso, {gewartet} esperas, {dauer} en total',
  ohneBestand: 'Este proceso no mantiene ninguna base de datos.',
  zwischenspeicher: 'Cachés de servicios externos',
  zwischenspeicherWert:
    '{treffer} aciertos, {geholt} obtenidos, {veraltet} servidos obsoletos, {fehlgeschlagen} fallidos',
  laufzeiten: 'Tiempos de ejecución por superficie',
  laufzeitenLeer:
    'Aquí no se ha medido ninguna llamada: solo se miden el servicio y el servidor de herramientas.',
  ungemessen: 'No medido: {liste}.',
  ungemessenNamen: {
    fenster:
      'los bindings de la ventana (no tienen un lugar común donde termine una llamada)',
    kommandozeile: 'la línea de comandos (vive lo que dura un comando)',
    speicher: 'la huella de este proceso (el sistema no la proporcionó)',
  },
  spalteArt: 'Tipo',
  spalteAnzahl: 'Llamadas',
  spalteFehler: 'Tasa de errores',
  spalteMittel: 'Media',
  spalteVerteilung: 'Distribución',
  arten: { lesend: 'lectura', schreibend: 'escritura' },
  eimer: 'hasta {grenze} ms',
  verteilung: '{eimer}: {anzahl} llamadas',
  eimerOffen: 'por encima',
  spurAus:
    'La traza está desactivada. Actívala en Ajustes › Diagnóstico; se aplica a partir del próximo inicio.',
  spurAn:
    'Traza activada: {aufgenommen} de {gesehen} llamadas, {abtastung} % muestreado. Como máximo {grenze} entradas durante {minuten} min: {verdraengt} desplazadas, {abgelaufen} caducadas.',
}

export const leistungsprofil = {
  ueberschrift: 'Perfil de rendimiento',
  erklaerung:
    'Un perfil de CPU de esta ventana durante el tiempo elegido, en el formato de «go tool pprof». Se guarda junto a la base de datos; cuántos se conservan se define en Ajustes › Diagnóstico.',
  sekunden: 'Duración en segundos',
  grenze: 'Un número entero de 1 a 60.',
  aufzeichnen: 'Grabar perfil',
  laeuft: 'Grabando: la ventana sigue siendo utilizable.',
  fertig: 'Escrito: {datei} ({groesse})',
}

export const zwischenspeicher = {
  ueberschrift: 'Cachés',
  erklaerung:
    'Lo que esta aplicación guarda para no obtenerlo ni calcularlo dos veces, en el disco y en la memoria. Nada de esto son tus datos: al vaciarlos, cada caché se vuelve a llenar en el siguiente uso.',
  spalteSpeicher: 'Caché',
  spalteOrt: 'Ubicación',
  spalteQuote: 'Tasa de aciertos',
  spalteEintraege: 'Entradas',
  spalteGroesse: 'Tamaño',
  spalteGrenze: 'Límite',
  keineQuote: 'aún sin consultas',
  keineGrenze: 'un archivo que se reemplaza',
  namen: {
    'scryfall.antworten': 'Respuestas de Scryfall',
    'scryfall.bilder': 'Imágenes de cartas de Scryfall',
    'scryfall.sammeldaten': 'Archivo masivo de Scryfall',
    'render.ebenen': 'Capas muestreadas',
    'pakete.bilder': 'Imágenes de los paquetes de plantillas',
    'skript.baeume': 'Scripts analizados',
    'vorschau.bilder': 'Vistas previas de cartas',
    'tisch.texturen': 'Imágenes de cartas en la mesa',
  },
  orte: { platte: 'Disco', arbeitsspeicher: 'Memoria' },
  leerenHinweis:
    'Vaciar lo descarta todo. Sin red, las respuestas y las imágenes de Scryfall solo vuelven cuando la red regresa.',
  leeren: 'Vaciar todo',
  geleert: 'Vaciado: se vuelven a llenar en el siguiente uso.',
}
