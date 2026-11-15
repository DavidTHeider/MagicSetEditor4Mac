// Die Texte der **Kartenseite** — die größte Seite der Anwendung.
//
// Aus `seiten.ts` herausgelöst, als die Datei über 400 Zeilen wuchs (§2).
// Der Schnitt folgt der Frage: dort stehen die Texte **aller** Seiten, hier
// die der einen, die am meisten trägt — und weiter wachsen wird
// (#674).

import { verlaufEn } from './seitenverlauf'
import { massenvorschauEn } from './seitenmassen'
import { autoersetzung } from './seitenautoersetzung'
import { schluesselwoerter } from './seitenschluesselwoerter'
import { werkstatt } from './seitenwerkstatt'
import { konsole } from './seitenkonsole'
import { symbolbibliothek } from './seitensymbolbibliothek'

export const karten = {
  ...werkstatt,
  autoersetzung,
  schluesselwoerter,
  symbolbibliothek,
  anlegen: 'Añadir carta',
  nachOben: 'Subir',
  nachUnten: 'Bajar',
  duplizieren: 'Duplicar carta',
  loeschen: 'Eliminar carta',
  loeschenFrage: '¿Eliminar esta carta? Con ⌘Z se puede deshacer.',
  loeschenJa: 'Eliminar',
  loeschenNein: 'Cancelar',
  loeschenTrotzdem: 'Eliminar de todos modos',
  symbole: 'Insertar símbolos',
  hergeleitet: 'Derivado',
  hergeleitetWert: 'Derivado: {wert}',
  bildWaehlen: 'Elegir imagen',
  stilfelder: 'Campos del estilo',
  stilfelderHilfe:
    'Estos campos vienen con el estilo de la carta. Su valor solo vale para este estilo.',
  symbolZeichnen: 'Dibujar símbolo',
  kartenskripteTitel: 'Cartas desde un script',
  kartenskripteErklaerung:
    'Acciones ya preparadas del paquete de juego: un clic añade varias cartas a la vez. ' +
    'Un solo deshacer las revierte todas.',
  kartenskriptEine: 'Una carta añadida.',
  kartenskriptMehrere: '{anzahl} cartas añadidas.',
  bildEinsetzen: 'Insertar imagen',
  pack: {
    ueberschrift: 'Pack aleatorio',
    erklaerung:
      'Un pack de refuerzo con las cartas de este set, según las reglas del paquete de juego. La misma semilla da la misma extracción: así se puede reproducir un resultado.',
    laedtGrund:
      'Cada tipo de pack pregunta a cada carta si encaja; eso tarda un momento la primera vez y después queda listo.',
    art: 'Tipo de pack',
    same: 'Semilla',
    sameZahl: 'La semilla es un número entero de 0 a 18446744073709551615.',
    ziehen: 'Extraer',
    neuerSame: 'Nueva semilla',
    gezogen: 'Extraído',
    keineArten:
      'El juego de este set no aporta tipos de pack, o su paquete no está instalado.',
    keineKarten: 'Este tipo de pack no produjo ninguna carta.',
    artMitZahl: '{name} ({zahl} cartas)',
    nurBaustein: 'Componente: lo extraen otros tipos de pack',
    vorbehalte: 'Tipos de pack cuyo filtro no se pudo evaluar',
  },
  symbol: {
    ueberschrift: 'Símbolo',
    erklaerung:
      'El símbolo de este set: dibujado, no cargado. Aparece de inmediato en la vista previa y entra en el archivo al guardar el set. Cada forma se puede colocar y mover con el teclado.',

    grundform: 'Añadir una forma básica',

    verlauf: 'Historial del dibujo',
    nachzeichnen: 'Vectorizar una imagen',
    groesser: 'Más grande',
    nachzeichnenHilfe:
      'Un PNG, JPEG o GIF con una silueta clara se convierte en formas con puntos de Bézier que puedes seguir editando aquí.',
    zuruecknehmen: 'Deshacer gesto',
    wiederherstellen: 'Rehacer gesto',
    kreis: 'Círculo',
    rechteck: 'Rectángulo',
    dreieck: 'Triángulo',
    stern: 'Estrella',

    teile: 'Partes',
    keineTeile:
      'Todavía no hay nada dibujado. Una forma básica es el comienzo; después se puede cambiar en cualquier punto.',
    teilLoeschen: 'Eliminar parte',
    teilMitPunkten: '{name} — {zahl} puntos',
    teilMitKopien: '{name} — simetría, {zahl} copias',
    gruppieren: 'Agrupar',
    gruppierenHilfe:
      'Dos o más partes seleccionadas se convierten en un grupo. Mantén pulsada ⌘ para seleccionar varias.',
    aufloesen: 'Desagrupar',
    nachOben: 'Subir',
    nachUnten: 'Bajar',
    nachObenFuer: 'Subir {name}',
    nachUntenFuer: 'Bajar {name}',
    teilLoeschenFuer: 'Eliminar {name}',
    teilMitKindern: '{name} — {zahl} partes',
    teilKind: '{name}, dentro de un grupo',
    wahlHilfe:
      'Mantén pulsada ⌘ para seleccionar varias partes. El orden es el orden de dibujo: lo que está primero queda encima.',

    zeichnung: 'Área de dibujo',
    punkt: 'Punto {nummer} de {name}',
    griffDavor: 'Tirador antes del punto {nummer}',
    griffDanach: 'Tirador después del punto {nummer}',
    punktHinzufuegen: 'Insertar punto',
    punktLoeschen: 'Eliminar punto',

    tastatur: 'Con el teclado',
    tastaturhilfe:
      'Tab se mueve entre los puntos. Las flechas desplazan una centésima, y con Mayús una décima. Intro inserta un punto nuevo tras el seleccionado; Retroceso lo elimina.',

    kombination: 'Combinar',
    ueberdecken: 'Superponer',
    vereinen: 'Unir',
    abziehen: 'Restar',
    schneiden: 'Intersecar',
    unterschied: 'Diferencia',
    umranden: 'Contorno',

    symmetrie: 'Simetría',
    symmetrieAnlegen: 'Añadir simetría',
    symmetrieAufloesen: 'Disolver simetría',
    symmetrieErklaerung:
      'Una simetría multiplica lo que hay dentro de ella. Las copias aparecen tenues en el lienzo; solo se edita siempre el original.',
    symmetrieHinweis: 'Selecciona una parte para convertirla en una simetría.',
    symmetrieart: 'Tipo',
    drehung: 'Rotación',
    achse: 'Reflexión',
    kopien: 'Copias',
    kopienHinweis:
      'Una reflexión necesita pares: un número impar deja una copia sola, y MSE redondea hacia abajo.',
    mitteX: 'Centro, horizontal',
    mitteY: 'Centro, vertical',
    griffX: 'Tirador, horizontal',
    griffY: 'Tirador, vertical',
    mitte: 'Centro de la simetría',
    griff: 'Tirador de la simetría',

    sichern: 'Poner en el set',
    gesichert:
      'Guardado como {name}. Entra en el archivo cuando se guarda el set.',
  },
  konsole,

  wortlisten: {
    ueberschrift: 'Opciones',
    liste: 'Elegir de {liste}',
    keine: 'ninguna de estas',
    gruppe: '{name} (grupo)',
  },
  regeln: {
    ueberschrift: 'Rulings de Scryfall',
    holen: 'Obtener rulings',
    keine: 'No hay rulings para esta carta.',
    offiziell: 'oficial',
    anmerkung: 'Nota de Scryfall',
    anmerkungErklaerung:
      'Las notas proceden del equipo de Scryfall y a menudo explican un formato no oficial. No son reglas.',
    quelle: 'Ver esta carta en Scryfall',
    alter: 'Estado: hace {stunden} horas',
    manaPruefen: 'Comprobar coste de maná',
    manaGut: 'El coste de maná es correcto.',
    manaSchlecht: 'Desconocido: {symbole}',
    manaLeer: 'Esta carta no tiene coste de maná.',
  },
  abgleich: {
    ueberschrift: 'Datos de la carta de Scryfall',
    erklaerung: 'Muestra lo que cambiaría. Solo se escribe lo que marques.',
    holen: 'Obtener propuesta',
    uebernehmen: 'Aplicar lo marcado',
    nichtsZuTun: 'Nada que cambiar: la carta coincide con Scryfall.',
    alt: 'actual',
    neu: 'nuevo',
    leer: '(vacío)',
    alter: 'Estado: hace {stunden} horas',
    frisch: 'Estado: recién obtenido',
    quelle: 'Ver esta carta en Scryfall',
  },
  scryfall: {
    ueberschrift: 'Imagen de Scryfall',
    erklaerung:
      'El nombre de la carta determina la imagen. El artista también se escribe en la carta: Scryfall exige que su nombre figure junto a la ilustración.',
    name: 'Nombre de la carta en Scryfall',
    art: 'Tipo de imagen',
    artKlein: 'Pequeña (vista previa)',
    artNormal: 'Carta completa',
    artAusschnitt: 'Solo la ilustración',
    holen: 'Obtener imagen',
    ausschnittHinweis:
      'El recorte muestra solo la ilustración. Se rechaza cuando Scryfall no indica el artista.',
    quelle: 'Ver esta carta en Scryfall',
  },
  scryfallkarte: {
    name: 'Nombre de la carta en Scryfall',
    platzhalter: 'p. ej. Lightning Bolt',
    mitBild: 'con ilustración',
    anlegen: 'Crear desde Scryfall',
  },
  zuschnitt: {
    ueberschrift: 'Elegir el recorte',
    erklaerung:
      'El marco indica qué parte de la imagen va a la carta. Arrastra para moverlo y arrastra una esquina para cambiar su tamaño.',
    quelle: 'Imagen de origen con el recorte elegido',
    vorschau: 'Vista previa al tamaño del campo',
    links: 'Izquierda',
    oben: 'Arriba',
    breite: 'Ancho',
    hoehe: 'Alto',
    seitenverhaeltnis: 'Mantener la proporción del campo',
    uebernehmen: 'Insertar',
    abbrechen: 'Cancelar',
    masse: '{breite} × {hoehe} píxeles',
  },
  feldWaehlen: 'Editar este campo',
  leerTitel: 'Todavía no hay cartas',
  listeBezeichnung: 'Cartas del set',
  mehr: 'Más',
  gruppeAnlage: 'Carta',
  gruppeAnsicht: 'Vista',
  ausScryfallAufklapp: 'Crear desde Scryfall …',
  ausScryfallErklaerung:
    'En lugar de una carta en blanco: escribe un nombre o busca en Scryfall; los campos vienen entonces de allí.',
  dieseKarte: 'Esta carta',
  kartenmenue: 'Carta',
  auswahlLoesen: 'Mostrar todas las cartas',
  bezugAufklapp: 'Vincular a una impresión …',
  dieseKarteLeer: 'Abre una carta para editarla.',
  keineKarteTitel: 'Ninguna carta seleccionada',
  titel: 'Cartas',
  leer: 'Crea una o pega cartas copiadas.',
  ohneName: 'Sin nombre',
  ohneFelder: 'Esta carta no tiene ningún valor de campo.',
  nummer: 'Número',
  vorschau: 'Vista previa de la carta',
  keineKarte:
    'Elige una de la lista; sus campos y su vista previa aparecerán aquí.',
  bearbeiten: 'Editar',

  notizen: {
    ueberschrift: 'Notas de la carta',
    platzhalter: 'Solo para ti: nunca aparece en la carta.',
  },

  // **What the game package does not declare** (#1271).
  buchfuehrung: 'Lo que la aplicación sabe al respecto',
  tiefe: 'Cuánto ves',
  tiefeNormal: 'Normal',
  tiefeErweitert: 'Avanzado',
  tiefeExperte: 'Experto',
  tiefeErweitertTitel: 'Lo que rellena un script',
  tiefeErweitertHinweis:
    'El paquete de juego calcula esto: dice por qué hay algo en la carta. No se puede editar.',
  befundVerdeckt: 'Hay un hallazgo bajo «{tiefe}».',
  befundSpringen: 'Ir allí',
  buchfuehrungHinweis:
    'Estas entradas no pertenecen a la carta, sino a su administración: qué plantilla usa y qué está configurado en ella. Cámbialas solo si sabes por qué.',

  // **Note and marks together, collapsed** (#1271). They sat below the list
  // and took its height: four rows were left.
  arbeitsstand: 'Progreso de esta carta',
  arbeitsstandHinweis:
    'Nota y marcas: pertenecen a tu trabajo con la carta, no a la carta. Un archivo que compartas no las incluye.',

  marken: {
    ueberschrift: 'Marcas',
    erklaerung:
      'Palabras para tu propio trabajo: están junto a la carta, no dentro de ella. Un archivo que compartas no las incluye.',
    platzhalter: 'p. ej. terminada',
    setzen: 'Añadir marca',
    entfernen: 'Quitar la marca {marke}',
    keine: 'Esta carta aún no tiene ninguna marca.',
    schonGesetzt:
      'La carta ya lleva «{marke}»: se puso en otro lugar mientras tanto.',
    filter: 'Filtrar por marca',
    alle: 'Todas',
  },

  scryfallsuche: {
    frage: 'Buscar en Scryfall',
    platzhalter: 'p. ej. Lightning Bolt, o t:creature c:r cmc<=3',
    erklaerung:
      'Basta con el nombre de una carta. Si conoces el lenguaje de búsqueda de Scryfall, puedes usarlo aquí.',
    suchen: 'Buscar',
    nichtsGefunden: 'No se encontró nada. ¿Probar con otra palabra?',
    mehr: 'Scryfall tiene más resultados. Acota la consulta: no se carga nada automáticamente.',
    alter: 'Estado: hace {stunden} horas',
    frisch: 'Estado: recién obtenido',
    befund: 'Hallazgo: {text}',
  },

  // **Der Verweis auf einen echten Druck** (#1167).
  verweis: {
    aufschrift: 'Vinculada a',
    keiner:
      'Esta carta no está vinculada a ninguna impresión. La sincronización con Scryfall elegirá entonces una cualquiera: busca arriba y vincúlala.',
    druck: 'de {setcode} {nummer}',
    gesetzt: '— vinculada el {wann}',
    setzen: 'Vincular a {name} ({setcode} {nummer})',
    loesen: 'Desvincular',
  },
  // **Die Zwischenablage** (#1168).
  ablage: {
    kopieren: 'Copiar ({zahl})',
    ausschneiden: 'Cortar ({zahl})',
    einfuegen: 'Pegar',
  },

  // **Die Spalten der Kartenliste** (#1168).
  listenspalten: {
    ueberschrift: 'Columnas de la lista',
    jeSpiel: 'La elección vale para todos los sets del mismo juego.',
    filtern: 'Buscar columna',
    gezeigt: '{gezeigt} de {alle} columnas',
    ausgeblendet: 'Hay {anzahl} más ocultas — búscalas con el campo de arriba.',
    genauer: 'Coinciden más — precisa la búsqueda para verlas.',
    nachOben: 'mover más a la izquierda',
    nachUnten: 'mover más a la derecha',
    sortieren: 'Ordenar por',
    garNicht: 'Orden del set',
    aufsteigend: 'Ascendente',
    absteigend: 'Descendente',
    ohneVorlage:
      'No se puede leer el paquete de juego. Las columnas proceden de las propias cartas, sin títulos y sin el orden del paquete.',
    spalteAusblenden: 'Ocultar columna',
  },

  verlauf: verlaufEn,
  massen: {
    ...massenvorschauEn,
    keinFeld: '— elige un campo —',
    keinWert: '— elige un valor —',
    ueberschrift: 'Cambio masivo',
    erklaerung:
      'Establece un campo en todas las cartas marcadas, en una sola operación. Si algo sale mal, no se cambia ninguna.',
    auswaehlen: 'Marcar para el cambio masivo',
    mehrereWaehlen: 'Elegir varias',
    ausgewaehlt: 'marcadas',
    keineAuswahl:
      'No hay ninguna carta marcada. Marca varias en la lista para cambiarlas juntas.',
    feld: 'Campo',
    wert: 'Valor nuevo',
    anwenden: 'Aplicar',
    betroffen: 'El cambio falló en estas cartas. No se cambió ninguna:',
    gescheitert: 'fallido',
    ohneWert: 'aún sin valor',
  },
  stilTitel: 'Estilo',
  stilWaehlen: 'Estilo de este set',
  stilLeer:
    'No hay ningún estilo instalado. Sin uno, la vista previa recurre a un sustituto.',
  stilFrage: '¿Cambiar a este estilo?',
  stilSpielwechsel: 'El nuevo estilo pertenece a otro juego.',
  stilUngedeutet:
    'El nuevo juego no conoce estos campos. Sus valores se quedan en la base de datos, pero ya no se dibujarán:',
  stilVerloreneOptionen:
    'El nuevo estilo no tiene estas opciones de estilo. Sus valores se pierden:',
  stilUebernehmen: 'Cambiar de estilo',
  stilAbbrechen: 'Mantener estilo',
  stilKeiner:
    'El set nombra un estilo que no está instalado; elige uno de los instalados.',
  stiloptionen: 'Opciones de estilo',
  stiloptionenSichtbar: '{zahl} de {gesamt}',
  stiloptionenWeitere: 'Más opciones ({zahl})',
  marktTitel: 'Dónde encontrar esta carta',
  marktErklaerung:
    'Estos enlaces se construyen en tu equipo, a partir del nombre de la carta o de un identificador del proveedor. No sale nada hasta que hagas clic, en el navegador que elijas. No hace falta ninguna clave; si allí hay algo a la venta, lo decide el proveedor.',
  marktGenau: 'lleva exactamente a esta carta',
  marktSuche: 'lleva a una búsqueda',
  marktNurAnbieter: 'lleva solo al proveedor',
  umschlagAnnehmen: 'Carta desde un sobre',
  umschlagAngenommen: 'La carta ya está en este set.',
}
