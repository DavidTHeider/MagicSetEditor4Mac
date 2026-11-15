// The texts of the **set list** — open, create, import.
//
// Split out of `seiten.ts` when the file grew past 400 lines (§2).

export const sets = {
  auswahlLoesen: 'Mostrar todos los sets',
  erklaerung:
    'Todos los sets abiertos están en el almacén. Desde aquí llegas a sus cartas o los vuelves a escribir en un archivo.',
  titel: 'Sets',
  leer: 'Todavía no hay ningún set abierto. Abre un archivo .mse-set para empezar.',
  setOeffnen: 'Abrir set …',
  neuTitel: 'Un set nuevo',
  neuKopf: 'Nuevo set',
  einlesenTitel: 'Importar un set de otro sitio',
  neuHinweis:
    'Un set vacío: el juego y el estilo determinan el aspecto de sus cartas. El estilo se puede cambiar más adelante.',
  neuName: 'Nombre del set',
  neuStil: 'Estilo del nuevo set',
  neuAnlegen: 'Crear set',
  neuSchritte: 'Pasos para un set nuevo',
  neuSchrittArt: 'Tipo',
  neuSchrittAussehen: 'Aspecto',
  neuSchrittName: 'Nombre',
  neuArtFrage:
    '¿Para qué son las cartas? Nosotros buscamos el nombre del paquete por ti.',
  neuArtZusatz: '· hay {stile}',
  neuEigenes: 'Algo completamente propio: crear un paquete de juego',
  neuWeiter: 'Siguiente',
  neuSchrittZurueck: 'Atrás',
  neuUeberspringen: 'Omitir',
  neuWahl: '{spiel} · {stil}',
  neuOhneSpiele:
    'No hay ningún paquete de juego instalado. Sin uno no hay campos de carta; los paquetes se obtienen en la página de paquetes.',
  neuOhneStile:
    'No hay ningún paquete de estilo instalado para {spiel}. Sin un estilo no se puede crear ningún set; los paquetes se obtienen en la página de paquetes.',
  neuZuDenPaketen: 'Ir a los paquetes',
  neuZurueck: 'Pulsa ⌘[ para volver aquí; lo que has introducido sigue ahí.',
  beispielsets: 'Incluidos para probar',
  geoeffnet: 'abierto',
  schonOffen: 'Este archivo ya está abierto: aquí lo tienes.',
  teilerfolg:
    'El archivo no se leyó hasta el final. {anzahl} sets están en la base de datos de todos modos: aparecen en la lista de abajo.',
  beimOeffnenBefunde:
    'No se pudo leer al abrir. El set está ahí de todos modos; le faltan estos puntos:',
  // **The preview before reading in** — FMT-03.AK-01 (#1692). It appears
  // only when there is something to say.
  vorschau: {
    ueberschrift: 'Antes de importar: {name}',
    umfang: '{karten} · juego {spiel} · estilo {stil}',
    umfangOhneStil: '{karten} · juego {spiel}',
    schonOffen:
      'Este archivo ya está en tu biblioteca: Abrir muestra el set existente en lugar de añadir un segundo.',
    befunde: 'Estos puntos no se pueden leer; el set llegaría sin ellos:',
    bedarf:
      'Faltan paquetes. Puedes importarlo igualmente: las cartas se dibujan con un sustituto hasta que estén los paquetes:',
    einlesen: 'Importar',
    aufschlagen: 'Abrir',
    verwerfen: 'No importar',
  },
  // **Deletion** (#1218) — final, confirmed by name.
  loeschen: 'Eliminar …',
  loeschenTitel: 'Eliminar este set',
  loeschenWarnung:
    '{name} desaparece del almacén, junto con {karten}. No hay papelera ni deshacer. Lo que está en un archivo se queda allí: lo que se elimina es la copia del almacén.',
  loeschenSammlung:
    'Tu colección se conserva: un ejemplar pertenece a una impresión, no a un set.',
  loeschenFeld: 'Escribe el nombre del set para confirmar',
  loeschenPasstNicht: 'El nombre todavía no coincide con {name}.',
  loeschenJa: 'Eliminar definitivamente',
  loeschenNein: 'Cancelar',
  loeschenBefunde: 'El set ya no existe. No se pudieron limpiar estos restos:',

  spalten: {
    name: 'Nombre',
    spiel: 'Juego',
    karten: 'Cartas',
    herkunft: 'Origen',
  },
  mtgjsonUeberschrift: 'MTG JSON',
  mtgjsonErklaerung:
    'Definiciones de cartas de mtgjson.com, en un solo archivo o en una carpeta. Los campos de carta a los que se asignan dependen del juego.',
  mtgjsonSpiel: 'Juego',
  mtgjsonDatei: 'Abrir archivo …',
  mtgjsonVerzeichnis: 'Abrir carpeta …',
  mtgjsonKeinSpiel:
    'Ningún juego tiene una asignación para los campos de MTG JSON. Sin ella no se puede importar nada.',
  scryfallsatzUeberschrift: 'Un set de Scryfall',
  scryfallsatzErklaerung:
    'El código del set determina qué cartas llegan, con sus textos y sin imágenes. Se importan nombre, coste de maná, tipos, texto de reglas, texto de ambientación, fuerza, resistencia, rareza e ilustrador; las imágenes se obtienen carta por carta. El estilo se elige después.',
  scryfallsatzKeinSpiel:
    'Sin un juego con asignación no se puede importar ningún set.',
  scryfallsatzSpiel: 'Juego para la asignación',
  scryfallsatzCode: 'Código del set en Scryfall',
  scryfallsatzPlatzhalter: 'p. ej. lea',
  scryfallsatzEinlesen: 'Importar set',
  vorlageAusfuhr: 'Mediante una plantilla de exportación …',
  vorlageWaehlen: 'Plantilla de exportación',
  vorlageAusfuehren: 'Ejecutar',
  vorlageSchliessen: 'Cerrar',
  vorlageKeine:
    'No hay ninguna plantilla de exportación instalada para este juego. Las plantillas vienen en los paquetes descargados.',
  vorlageGeschrieben: 'Escrito en',
  vorlageBefunde:
    'Al ejecutarla surgió algo; el archivo se escribió de todos modos:',
  vorlageGescheitert:
    'Esta plantilla no se ejecuta hasta el final. Lo que le falta aparece debajo:',
  mtgjsonSichern: 'Guardar como MTG JSON …',
  mtgjsonGesichert: 'cartas guardadas en',
  mtgjsonVerluste:
    'No se puede expresar al exportar. El archivo existe; le faltan estos detalles:',
  mtgjsonSpalte: 'Exportación',
  speichern: 'Guardar',
  speichernUnter: 'Guardar como …',
  speicherSpalte: 'Guardar',
  gespeichert: 'Guardado en',
  beimSpeichernBefunde:
    'No se transfirió al guardar. El archivo del set se escribió; faltan estos archivos junto a él:',
  oeffnen: 'Mostrar las cartas de este set',
  handlungenAn: 'Acciones — {name}',
  handlungenKurz: '…',
  vorbehalte: {
    mitZahl: 'Sin interpretar ({anzahl})',
    ueberschrift: 'Sin interpretar',
    erklaerung:
      'La importación no entendió estas líneas. Viajan sin cambios y no se pierden al guardar.',
    keine: 'Nada: este set se leyó por completo.',
  },

  // **The portable envelope** — work item N4 of phase 5.4
  // (#1457).
  umschlagAblegen: 'Guardar como sobre',
  umschlagGeschnuert:
    'Guardado. El sobre está firmado: quien lo reciba puede comprobar que procede de aquí.',
  umschlagTitel: 'Incorporar un sobre',
  umschlagErklaerung:
    'Un sobre es un set que ha guardado otra persona. Primero se inspecciona y después se incorpora; antes de eso no cambia nada en la base de datos.',
  umschlagWaehlen: 'Inspeccionar un sobre …',
  umschlagUrheber: 'Firmado por',
  umschlagAbdruckHinweis:
    'La huella dice QUIÉN firmó, no si deberías confiar en esa persona. Si la conoces, compárala con la suya.',
  umschlagInhalt: 'Contiene',
  umschlagUebernehmen: 'Incorporar',
  umschlagVerwerfen: 'Mejor no',
  umschlagUebernommen:
    'Incorporado. El set se ha creado de nuevo: tiene identificadores propios y no sobrescribe nada.',
}
