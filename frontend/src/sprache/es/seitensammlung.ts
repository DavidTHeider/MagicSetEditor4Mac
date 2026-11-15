// The **collection** page — piece 2 and 3 of phase 3.4
// (#1165).
export const sammlung = {
  titel: 'Colección',
  erklaerung:
    'Lo que realmente tienes: una fila por impresión, estado, idioma y acabado. Esto no tiene nada que ver con tus sets: una carta diseñada es un borrador; un ejemplar es un trozo de cartón.',

  aufnehmenOeffnen: 'Añadir …',
  einfuhrOeffnen: 'Importar y exportar …',
  einfuhrTitel: 'Importar y exportar',
  leerTitel: 'Aún no hay nada registrado',
  leer: 'Un ejemplar necesita el identificador de Scryfall de su impresión: el UUID que aparece en la dirección de la página de una carta. Una lista entera se importa de una vez.',

  druck: 'Impresión',
  menge: 'Cantidad',
  zustand: 'Estado',
  sprache: 'Idioma',
  ausfuehrung: 'Acabado',
  foil: 'Foil',
  nonfoil: 'No foil',
  etched: 'Etched',
  erwerbsdatum: 'Fecha de adquisición',
  erwerbspreis: 'Precio de compra',
  erwerbswaehrung: 'Moneda',
  preisMehrdeutig:
    'Dos lecturas posibles: {erste} o {zweite}. Escríbelo de forma inequívoca.',
  notiz: 'Nota',

  // **Der lesbare Bezug und die Sicht darauf** (#1200).
  name: 'Nombre de la carta',
  setcode: 'Set',
  nummer: 'Número',
  sicht: 'Vista de la colección',
  suche: 'Buscar',
  suchePlatzhalter: 'nombre, set, identificador o nota',
  sortierung: 'Ordenar por',
  sortierBezeichnung: 'Carta',
  sortierMenge: 'Cantidad',
  sortierErwerbsdatum: 'Fecha de adquisición',
  ausfuehrungAlle: 'Todos los acabados',
  nurFoil: 'Solo foil',
  nurNonfoil: 'Solo no foil',
  nurEtched: 'Solo etched',

  navigator: 'Vista y herramientas',
  inspector: 'Ejemplar',
  exemplarmenue: 'Ejemplar',
  inDecks: 'En estos mazos',

  inspectorLeer:
    'Elige un ejemplar de la lista: aquí aparecen su cantidad, estado, adquisición y precio, y aquí lo modificas.',
  inspectorAusserhalb:
    'Este ejemplar ya no está en la vista actual. Un cambio sigue afectando a este mismo ejemplar.',
  arbeitsplatz: 'Espacio de trabajo de la colección',
  listeBezeichnung: 'Ejemplares',
  spalteWahl: 'Elegir',
  spalteBezeichnung: 'Nombre',
  spalteMenge: 'Cantidad',
  spalteZustand: 'Estado',
  spalteSprache: 'Idioma',
  spalteSet: 'Set',
  spalteLagerort: 'Ubicación',
  gruppeOhne: 'Sin especificar',
  gruppeSumme: '{anzahl} ejemplares · {menge} unidades',
  kachelWaehlen: 'Seleccionar {bezeichnung}',

  ansicht: 'Vista',
  ansichtTabelle: 'Tabla',
  ansichtGalerie: 'Galería',
  ansichtGruppen: 'Grupos',
  gruppierung: 'Agrupar por',
  gruppierungKeine: 'Sin agrupar',
  gruppierungSet: 'Set',
  gruppierungLagerort: 'Ubicación',
  gruppierungZustand: 'Estado',
  gruppierungSprache: 'Idioma',
  ohneTreffer: 'Ningún ejemplar coincide con la búsqueda y el filtro.',
  ohneTrefferZuruecksetzen: 'Restablecer búsqueda y filtro',
  gezeigt: '{gezeigt} de {gesamt}',
  gekappt: '{gezeichnet} de {passend} dibujados',
  mehrZeigen: 'Mostrar más',

  // **Die Listeneinfuhr** (#1292).
  listeTitel: 'Pegar una lista',
  listeErklaerung:
    'Una carta por línea, p. ej. «4 Lightning Bolt (M11) 149». Se omiten las líneas en blanco, «Deck», «Sideboard» y las líneas que empiezan por # o //.',
  listeFeld: 'Lista de cartas',
  listePlatzhalter: '4 Lightning Bolt (M11) 149\n2x Counterspell',
  listeDeuten: 'Asociar',
  listeEinlesen: 'Añadir',
  listeOffen:
    '{offen} de {gesamt} líneas necesitan una decisión: elige la impresión que tienes.',
  listeAlleEindeutig: 'Las {gesamt} líneas son inequívocas.',
  listeDruckWahl: '¿Qué impresión?',
  listeOhneWahl: '— aún sin decidir —',
  listeDrucke: '{drucke} impresiones',
  listeBericht:
    '{angelegt} nuevas, {erhoeht} aumentadas, {stueck} ejemplares añadidos.',
  listeNichtsGewaehlt: 'Ninguna línea decidida: elige al menos una impresión.',
  listeAusgelassen: '{ausgelassen} líneas existentes omitidas.',
  listeKeinDruck: 'Ninguna carta se llama así: revisa la ortografía.',
  listeZuViele:
    'Más de {drucke} impresiones se llaman así: indica el set entre paréntesis, por ejemplo «(M11)», y quedará una.',
  listeOhneKarte: '{anzahl} líneas no encuentran ninguna carta y se omiten.',
  listeOhneKarteEins: 'Una línea no encuentra ninguna carta y se omite.',
  vorschauTitel: 'Antes de importar',
  vorschauErklaerung:
    'Algunas líneas coinciden con líneas que tu colección ya tiene: misma impresión, estado, idioma, acabado y propósito. Elige qué pasa con ellas; hasta entonces no se escribe nada.',
  vorschauZaehlung:
    '{neu} nuevas · {zusammen} coinciden con líneas existentes · {wiederholt} repiten una línea de esta importación',
  vorschauZusammen: '{bezeichnung}, {zustand}: {vorher} → {nachher}',
  vorschauModus: 'Líneas existentes',
  vorschauAddieren: 'Sumar cantidades',
  vorschauAuslassen: 'Omitir las existentes',
  vorschauEinlesen: 'Importar',
  vorschauAbbrechen: 'No importar',

  // **Die Ausfuhr** (#1292).
  ausfuhrTitel: 'Exportar y guardar',
  ausfuhrErklaerung:
    '«CSV» lleva todos los datos y se puede volver a leer por completo. «Lista» es la forma que se puede pegar en un creador de mazos: solo lleva cantidad, carta, impresión y acabado.',
  ausfuhrAlsCSV: 'Como CSV',
  ausfuhrAlsListe: 'Como lista',
  ausfuhrErgebnis: 'Salida',
  ausfuhrZurueck: 'Volver a leer el CSV',
  ausfuhrLeer: 'Aún no se ha exportado nada.',

  // **Die Ordnung** (#1292, Vermerk 0057).
  lagerort: 'Ubicación',
  wunsch: 'Deseada, no poseída',

  // **Die Ordnungssicht** (#1292).
  ordnungTitel: 'Orden',
  ordnungLagerorte: 'Ubicaciones',
  ordnungDubletten: 'Duplicados',
  ordnungWuensche: 'Lista de deseos',
  ordnungBehalten: 'Conservar',
  ordnungLeerLagerorte:
    'Todavía no hay ningún ejemplar; sin uno no hay ubicación.',
  ordnungLeerDubletten: 'Ningún impreso está más de {behalten} veces.',
  ordnungLeerWuensche:
    'No hay nada en la lista de deseos. Un ejemplar llega a ella como «Deseada, no poseída», o desde un mazo con «Añadir lo que falta a la lista de deseos».',
  ordnungStueck: '{stueck} ejemplares en {zeilen} filas',
  ordnungOhneOrt: 'sin especificar',
  ordnungUeberzaehlig: '{ueberzaehlig} de más',

  // **Besitze ich diese Karte?** (#1292).
  besitzJa: 'Tienes {stueck} de estas.',
  besitzNein: 'No tienes esta impresión.',
  besitzGewuenscht: '({gewuenscht} en la lista de deseos)',

  // **Die Massenvorgänge** (#1292).
  masseTitel: 'Acciones en bloque',
  masseWaehlen: 'Seleccionar',
  masseGewaehlt: '{gewaehlt} de {gesamt} seleccionados',
  masseLagerortSetzen: 'Establecer la ubicación en',
  masseZustandSetzen: 'Establecer el estado en',
  masseAendern: 'Modificar los seleccionados',
  masseLoeschen: 'Eliminar los seleccionados',
  masseLoeschenFrage: '¿Eliminar {anzahl} ejemplares?',
  masseLoeschenErklaerung:
    'Lo que se elimine aquí tendrás que volver a escribirlo: la cantidad, el estado, la ubicación y el precio de compra no están en ningún otro sitio. La colección no tiene historial como un set.',
  masseLoeschenJa: 'Eliminar',
  masseBericht: '{geaendert} modificados.',
  masseZuruecknehmen: 'Deshacer',
  masseZuruecknehmenVon: 'Deshacer la ubicación de {anzahl} líneas',
  masseZurueckgenommen: '{anzahl} deshechos.',
  massePreise: 'Obtener los precios de todos',
  massePreisschub: '{geholt} obtenidos, {offen} pendientes',
  massePreiseFehlgeschlagen: '{anzahl} fallidos',
  massePreiseSteht: 'Este lote no ha avanzado: otro intento daría lo mismo.',
  massePreiseFertig: 'Ya están todos los precios.',

  aufnehmen: 'Añadir',
  aenderungSichern: 'Guardar',
  loeschen: 'Eliminar',
  loeschenFrage: '¿Eliminar este ejemplar?',
  loeschenErklaerung:
    'La cantidad, el estado, la ubicación y el precio de compra no están en ningún otro sitio: la colección no tiene historial como un set.',
  loeschenJa: 'Eliminar',
  abbrechen: 'Cancelar',
  bearbeiten: 'Editar',
  zumTausch: 'Al intercambio',
  insDeck: 'Al mazo',
  auswahlTitel: 'Selección',
  auswahlInsDeck: '{anzahl} → mazo',

  neuesTitel: 'Añadir un ejemplar',
  weitereAngaben: 'Más datos',

  dubletteFrage:
    '{bezeichnung} ya está en la colección: {menge} en el mismo estado, idioma y acabado.',
  dubletteErklaerung:
    'Aumentar suma la cantidad nueva; los demás datos de la fila existente se conservan. Cancelar no cambia nada: tu entrada se mantiene.',
  dubletteErhoehen: 'Aumentar a {summe}',
  dubletteZurZeile: 'Ir a la fila existente',
  dubletteAbbrechen: 'Cancelar',

  zustaende: {
    mint: 'Mint',
    nearMint: 'Near Mint',
    excellent: 'Excellent',
    good: 'Good',
    lightPlayed: 'Light Played',
    played: 'Played',
    poor: 'Poor',
  },

  preiseTitel: 'Precios',
  preiseHolen: 'Obtener precios',
  marktUndVerlauf: 'Ofertas e historial',
  // **The marketplace** — work item A4 of phase 5.3 (#1443).
  marktTitel: 'Ofertas del marketplace',
  marktErklaerung:
    'Lo que piden vendedores individuales por esta impresión: precio, cantidad, estado e idioma. La consulta sale a la red. No se compra nada.',
  marktHolen: 'Obtener ofertas',
  marktHolenBei: 'Obtener ofertas de {quelle}',
  marktLaeuft: 'Obteniendo …',
  marktLeer: 'Ahora mismo no hay nada a la venta allí para esta impresión.',
  marktZugeordnet:
    'Esta impresión está ahora asociada a la versión de carta {fassung}: la próxima consulta la encontrará de inmediato.',
  marktStueck: '{menge} disponibles',
  marktBuendel: 'en lotes de {buendel}',
  marktImUrlaub: 'no envía por ahora',
  marktGradiert: 'graduada',
  marktSigniert: 'firmada',
  marktVeraendert: 'alterada',
  marktBefunde: 'No todas las ofertas se pudieron interpretar:',

  // **Die Begleitbefunde des Preisholens** (#1202).
  preisbefunde:
    'No todos los precios se pudieron interpretar. Faltan estas series:',
  preiseLaufen: 'Obteniendo …',
  keinePreise: 'No hay ningún precio registrado para esta impresión.',
  preisAlter: 'hace {stunden} h · {quelle}',
  preisFrisch: 'ahora mismo · {quelle}',
  preisVeraltet: 'obsoleto, de hace {tage} días · {quelle}',
  preisQuellstand: 'datos de la fuente del {tag}',

  vorbehalt:
    'Los precios son estimaciones de Scryfall y sus socios. No se garantiza ninguno; el que vale es el precio de la tienda.',

  verlauf: 'Historial',
  keinVerlauf:
    'Solo hay un dato para esta serie: el historial se irá formando cada vez que vuelvas a obtener precios.',

  bulkTitel: 'Obtener los precios de una fuente completa',
  leiste: {
    bezeichnung: 'Cifras de la colección',
    exemplare: 'Ejemplares',
    eintraege: '{zahl} entradas',
    wert: 'Valor',
    keinPreis: 'sin precio',
    ohnePreis: 'Sin precio',
    nichtGelesen: 'no leído',
  },
  drucksuche: {
    karte: 'Carta',
    hinweis:
      'Nombre, opcionalmente con set y número — p. ej. Lightning Bolt (M11) 149',
    suchen: 'Buscar',
    drucke: 'Impresiones',
    keinTreffer: 'No se encontró ninguna carta con ese nombre.',
    weitere: 'Hay más impresiones — con set y número la elección es más corta.',
    treffer: '{name} · {set} ({code}) · N.º {nummer}',
  },
  bulkErklaerung:
    'MTGJSON no responde carta a carta, sino con un archivo. «Hoy» trae el día en curso, «Noventa días» el historial, y más de un gigabyte.',
  bulkQuelle: 'Fuente',
  bulkUmfang: 'Alcance',
  bulkHeute: 'Hoy',
  bulkNeunzigTage: 'Noventa días',
  bulkHolen: 'Obtener',
  bulkHolenBei: 'Obtener precios de {quelle}',
  bulkLaeuft: 'Obteniendo …',
  bulkBilanz: '{staende} precios almacenados para {drucke} impresiones.',
  bulkOhneDrucke:
    'Ninguna impresión de esta colección lleva un identificador de MTGJSON: no se ha pedido nada.',
  uebersichtTitel: 'Resumen',
  uebersichtLeer: 'Aún no se ha obtenido ningún precio.',
  ohnePreis: 'sin precio: {zahl}',
  // **Ein Exemplar und nicht „1 Exemplare“.** Eine Zahl mit falscher
  // Form liest sich wie ein Fehler im Programm — gesehen im laufenden
  // Fenster.
  einExemplar: '{zahl} ejemplar',
  exemplare: '{zahl} ejemplares',

  keineUmrechnung:
    'Una suma por moneda. No se convierte: haría falta un tipo de cambio y este programa no lo tiene.',
  // **A price history from another source** — A5, #1444.
  verlaufTitel: 'Historial de precios de otra fuente',
  verlaufErklaerung:
    'Obtiene una serie a lo largo del tiempo y la guarda. La consulta sale a la red y consume cuota; el estado se muestra tal como lo nombra la fuente.',
  verlaufHolen: 'Obtener historial',
  verlaufHolenBei: 'Obtener historial de {quelle}',
  verlaufLaeuft: 'Obteniendo …',
  verlaufBilanz:
    '{staende} puntos de precio almacenados — versión de carta {fassung}.',
} as const
