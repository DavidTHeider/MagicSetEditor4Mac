// The texts of the individual pages.
import { bestandspflege } from './seitenbestand'
import { datenpflege } from './seitenpflege'
import { kosten } from './seitenkosten'
import { ruecksetzung } from './seitenruecksetzung'
import { ruecksprung } from './seitenruecksprung'
import { einstellungen } from './seiteneinstellungen'
import { nutzer } from './seitennutzer'
import { karten } from './seitenkarten'
import { sprachstaende } from './seitensprachstaende'
import { handeln } from './seitenhandeln'
import { technik } from './seitentechnik'
import { recherche } from './seitenrecherche'
import { decks } from './seitendecks'
import { turniere } from './seitenturniere'
import { handlungen } from './seitenhandlungen'
import { spielen } from './seitenspielen'
import { sammlung } from './seitensammlung'
import { adressbuch } from './seitenadressbuch'
import { vorlagen } from './seitenvorlagen'
import { sets } from './seitensets'
import { ausgeben } from './seitenausgeben'
import { werkzeugseiten } from './seitenwerkzeuge'
import { start } from './seitenstart'
import { paketplan } from './seitenpaketplan'
import { bezuege } from './seitenbezuege'

export const seiten = {
  start,
  sets,
  setangaben: {
    kopf: 'Datos del archivo',
    kopfHinweis:
      'Lo que Magic Set Editor anotó sobre este set al guardarlo. Se conserva y no se modifica.',
    mseFassung: 'Escrito con MSE',
    spielfassung: 'Versión del paquete de juego',
    stilfassung: 'Versión del paquete de estilo',
    versionskontrolle: 'Control de versiones',
    apprenticecode: 'Código de Apprentice',
    titel: 'Datos del set',
    leerTitel: 'Sin campos de datos',
    leer: 'El paquete de juego de este set no declara ningún bloque `set field:`. Elige un juego que aporte alguno.',
  },
  karten,
  sprachstaende,
  decks,
  turniere,
  spielen,
  handlungen,
  sammlung,
  adressbuch,
  vorlagen,
  ...werkzeugseiten,
  bilder: {
    titel: 'Imágenes',
    erklaerung:
      'Cada carta se dibuja y se escribe por separado. El proceso se ejecuta aquí, en la ventana: muestra en qué punto va y se puede detener en cualquier momento.',
    ordnerWaehlen: 'Elegir una carpeta …',
    keinOrdner: 'Todavía no se ha elegido ninguna carpeta.',
    gemerkt: 'Elegida la última vez — la misma carpeta para todos los sets.',
    nichtGemerkt: 'La carpeta vale para esta salida, pero no se ha recordado:',
    vorlage: 'Plantilla de nombre',
    konflikt: 'Si ya existe un archivo',
    konfliktHilfe:
      'Dos cartas con el mismo nombre dan el mismo nombre de archivo. «Numerar de nuevo» es el valor predeterminado de Magic Set Editor 2: cuenta hacia arriba dentro de una ejecución y sobrescribe lo que dejó una anterior.',

    // Die Schlüssel kommen aus `internal/domain/einstellung.go`.
    konflikte: {
      altbehalten: 'Conservar el existente',
      ueberschreiben: 'Sobrescribir',
      nummerieren: 'Numerar',
      nummerierenneu: 'Numerar de nuevo',
    },
    vorlageHilfe:
      '{nummer} se convierte en el número de carta de tres cifras y {name} en el nombre de la carta. Los nombres iguales reciben un número añadido.',
    format: 'Formato',
    formatPNG: 'PNG',
    formatJPG: 'JPG',
    formatBMP: 'BMP',
    formatTIFF: 'TIFF',
    formatSVG: 'SVG',
    guete: 'Calidad',
    gueteFast: 'Rápida: para un vistazo rápido',
    gueteMedium: 'Según la regla de Magic Set Editor 2',
    gueteHigh: 'Alta: sin parpadeo en las líneas finas',
    gueteUltra: 'Máxima: para imprimir a plena resolución',
    breite: 'Ancho en píxeles',
    breiteHilfe:
      'Sin ppp: ninguna plantilla indica cuántos milímetros mide una carta. 750 píxeles en una carta de 63 mm son 300 ppp.',
    ausgeben: 'Escribir imágenes',
    abbrechen: 'Detener',
    laeuft: 'escritas de',
    fertig: 'imágenes escritas en',
    abgebrochen: 'Detenido.',
    uebersprungen:
      '{anzahl} cartas omitidas: el archivo ya existía y está seleccionado «Conservar el existente»:',
    befunde:
      'Incompleto al dibujar. Las imágenes existen; les faltan estas partes:',
  },
  druckbogen: {
    titel: 'Hoja de impresión',
    erklaerung:
      'Las cartas conservan su tamaño: 63 × 88 mm, tal como se cortan. Las marcas de corte están en el borde del papel, no sobre la carta.',
    papier: 'Papel',
    papierA4: 'A4',
    papierLetter: 'Letter',
    anschnitt: 'Sangrado en milímetros',
    anschnittHilfe:
      'El sangrado es lo que la imagen sobresale de la línea de corte: evita que una carta cortada muestre un borde blanco. Las cartas se disponen con el doble del sangrado entre ellas.',
    aufloesung: 'Resolución',
    dpi150: '150 ppp',
    dpi300: '300 ppp',
    dpi600: '600 ppp',
    schnittmarken: 'Marcas de corte',
    legen: 'Componer hoja de impresión …',
    fertig: 'en',
    seiten: 'páginas:',
    befunde: 'No todo cupo. La hoja existe; estos puntos requieren revisión:',
  },
  // The **technical area** (#1164) — commands to copy.
  handeln,
  technik,

  // **Die Recherche** (A16, #1515) — sie steht im Sucharbeitsplatz und
  // im Karteneditor, nicht auf einer eigenen Seite.
  recherche,
  pakete: {
    plan: paketplan,
    bezuege,
    sammeldaten: {
      ueberschrift: 'Datos masivos de Scryfall',
      erklaerung:
        'Un archivo con todos los datos de las cartas. Una vez que está aquí, las funciones de Scryfall funcionan sin red, con datos de hasta un día de antigüedad.',
      fehlt: 'Todavía no se han obtenido.',
      vorhanden: 'Presentes, {mb} MB.',
      angeboten: 'Scryfall ofrece {mb} MB, a fecha de {stand}.',
      geholt: 'Descargado el {zeit}.',
      holen: 'Obtener datos masivos',
      laedt: 'Obteniendo …',
    },

    titel: 'Paquetes',
    bezugUeberschrift: 'Obtener una plantilla',
    bezugErklaerung:
      'Las plantillas de MSE están en repositorios. Lo que se obtiene es un estado (una versión publicada o un commit), nunca una rama: una rama cambia, y la misma entrada daría mañana un paquete distinto.',
    bezugAdresse: 'Dirección del repositorio',
    bezugStand: 'Estado (versión publicada o commit)',
    bezugUnterverzeichnis: 'Subcarpeta (raíz del paquete)',
    bezugName: 'Nombre en la carpeta de paquetes',
    bezugPruefen: 'Mostrar de dónde procede …',
    bezugStarten: 'Obtener',
    bezugAbbrechen: 'No obtener',
    bezugGeholt: 'Obtenido en',
    bezugPruefsumme: 'Suma de comprobación',
    bezugUebergangen: 'No se desempaquetó: los ejecutables no se incluyen:',
    bezugUnvollstaendig:
      'Faltan la dirección y el estado. Sin ambos no hay nada que obtener.',
    bezugHerkunft: 'Se obtendrá {quelle}, estado {stand}.',
    bezugLizenz:
      'Estos contenidos no proceden de MagicSetEditor4Mac y no están bajo nuestra licencia. Los marcos de las cartas, los símbolos y las fuentes pueden estar protegidos; usarlos es responsabilidad tuya. Los ejecutables no se desempaquetan.',
    mse2Ueberschrift: 'Importar desde una instalación de MSE2',
    mse2Erklaerung:
      'Quien usa MSE2 ya tiene sus plantillas en el equipo. Abajo se indican los lugares habituales; no se lee ninguno antes de que lo elijas.',
    mse2Uebernehmen: 'Importar desde aquí',
    mse2VonHier: 'Importar desde aquí — {ort}',
    mse2Waehlen: 'Elegir una carpeta …',
    mse2KeinVorschlag:
      'No se encontró ninguna instalación de MSE2 en los lugares habituales. Elige tú mismo la carpeta con las plantillas; en MSE2 se llama «data».',
    installerErklaerung:
      'O un archivo .mse-installer, que es como MSE2 comparte plantillas. Lo que ya está instalado se queda como está.',
    installerWaehlen: 'Leer instalador …',
    mse2Ergebnis: '{anzahl} plantillas importadas.',
    mse2Uebergangen:
      'Ya existían y por eso se omitieron; no se sobrescribió nada:',
    abhaengigUeberschrift: 'Dependencias',
    abhaengigAllesGut: 'Todas las dependencias se cumplen.',
    abhaengigFehlt:
      '{paket} necesita {braucht}, que no está instalado. Obtenlo o impórtalo desde una instalación de MSE2.',
    abhaengigZuAlt:
      '{paket} necesita {braucht} en la versión {verlangt}; la instalada es {da}. Sin la más reciente, la plantilla dibuja en silencio otra cosa.',
    abhaengigZyklus: '{paket} y {braucht} se necesitan mutuamente en círculo.',
    abhaengigUnlesbar:
      '{paket} necesita {braucht}: no se pueden comparar las versiones; requerida {verlangt}, instalada {da}.',
    quellenErklaerung:
      'Cuatro colecciones, un clic para cada una: la dirección y el estado se rellenan; no se obtiene nada hasta que aceptes. Las fuentes de las plantillas se incluyen; sin ellas ninguna carta se dibuja bien. Ninguna de las cuatro lleva licencia: los contenidos pertenecen a sus titulares de derechos, no a nosotros.',
    quellenGroesse: 'unos {megabyte} MB',

    // Die Schlüssel kommen aus `internal/bezug/quellen.go`. Ein Wächter
    // in `internal/werkzeuge` hält beide Seiten zusammen: eine Quelle
    // ohne Satz hier macht die Pipeline rot.
    quellen: {
      basicM15:
        'Los marcos M15 habituales. El camino más rápido hasta una carta dibujada.',
      fullNonMagic: 'Plantillas para juegos distintos de Magic.',
      m15: 'Las plantillas públicas de M15: 324 paquetes. El mejor comienzo para la mayoría.',
      fullMagic:
        'Casi todas las plantillas públicas de Magic: 526 paquetes. Esta colección es la vara de medir de este proyecto.',
    },
    katalogUeberschrift: 'Paquetes instalados',
    katalogLeer:
      'No hay ningún paquete instalado. Sin juego ni estilo, la vista previa recurre a un sustituto.',
    katalogName: 'Nombre',
    katalogArt: 'Tipo',
    katalogVersion: 'Versión',
    katalogHandlung: 'Acción',
    katalogBefunde: 'Al leer las carpetas de paquetes:',
    anlegen: 'Crear un paquete',
    anlegenArt: 'Tipo de paquete nuevo',
    anlegenArtSpiel: 'Juego (.mse-game)',
    anlegenArtStil: 'Estilo (.mse-style)',
    anlegenArtInclude: 'Include (.mse-include)',
    anlegenArtExport: 'Plantilla de exportación (.mse-export-template)',
    anlegenName: 'Nombre corto',
    anlegenNameHinweis:
      'Letras, cifras, guiones y guiones bajos; se convierte en el nombre de la carpeta',
    anlegenVollname: 'Nombre completo',
    anlegenVollnameHinweis: 'Vacío significa: igual que el nombre corto',
    anlegenSpiel: 'Juego del paquete nuevo',
    anlegenSpielHinweis:
      'Un estilo distribuye los campos de su juego; una plantilla de exportación los escribe',
    anlegenSpielLeer: 'Elige uno',
    anlegenKnopf: 'Crear',
    anlegenGelungen: 'Creado: {name}',
    loeschen: 'Eliminar',
    loeschenAn: 'Eliminar — {name}',
    handlungenAn: 'Acciones — {name}',
    handlungenKurz: '…',
    katalogSuche: 'Buscar',
    auswahlLoesen: 'Mostrar todos los paquetes',
    leiste: {
      bezeichnung: 'Cifras de los paquetes',
      installiert: 'Instalados',
      befund: 'Con hallazgos',
      nichtGelesen: 'no leído',
      auswahl: '{pakete} con hallazgos',
    },
    katalogSucheHinweis: 'Nombre, tipo, origen o grupo',
    katalogArtAlle: 'Todos los tipos',
    katalogOrdnung: 'Ordenar por',
    katalogOrdnungGruppen: 'Grupos',
    katalogOrdnungName: 'Nombre',
    katalogOrdnungVersion: 'Versión',
    katalogOrdnungHerkunft: 'Origen',
    katalogGefunden: '{gefunden} de {gesamt} paquetes',
    katalogNichtsGefunden:
      'Ningún paquete coincide con esta búsqueda. Prueba otra grafía o restablece el filtro de tipo.',
    katalogLeerTitel: 'Todavía no hay ningún paquete',
    katalogNichtsTitel: 'Sin resultados',
    katalogOhneGruppe: 'Sin grupo',
    katalogGruppeAn: 'Grupo {name}',
    katalogGruppe: 'Grupo',
    katalogOhneVersion: 'sin versión',
    katalogZustandFehlt: 'Falta una dependencia',
    katalogZustandZuAlt: 'Dependencia demasiado antigua',
    katalogZustandUnlesbar: 'Versiones no comparables',
    katalogZustandZyklus: 'Ciclo de dependencias',
    katalogEinVorbehalt: 'un hallazgo',
    katalogVorbehalte: '{anzahl} hallazgos',
    umschlagTitel: 'Una plantilla desde un sobre',
    umschlagErklaerung:
      'Un sobre contiene los **archivos** de una plantilla; tú eliges el nombre con su extensión, por ejemplo `mygame.mse-game` o `mystyle.mse-style`. La plantilla queda aquí con ese nombre. Si el paquete ya existe, no se sobrescribe nada: recibes un rechazo y eliges otro nombre.',
    umschlagName: 'Cómo debe llamarse el paquete',
    umschlagOhneNamen: 'Indica primero cómo debe llamarse el paquete.',
    umschlagAnnehmen: 'Aceptar plantilla',
    umschlagAngenommen: 'La plantilla ya está en tu carpeta de paquetes.',
  },
  // **Suchen und Ersetzen** (#1168).
  ersetzen: {
    titel: 'Reemplazar',
    erklaerung:
      'Reemplaza texto en los campos de texto de este set. Primero la vista previa, luego aplicar; y un solo deshacer revierte toda la operación.',
    suchtext: 'Buscar',
    ersatz: 'Reemplazar por',
    grossKlein: 'Distinguir mayúsculas y minúsculas',
    ganzesWort: 'Solo palabras completas',
    vorschauen: 'Mostrar qué cambia',
    ersetzen: 'Reemplazar ahora ({treffer})',
    nichtsGefunden: 'No se encontró nada. ¿Probar con otra palabra?',
    zusammenfassung: 'Coincidencias: {treffer} · cartas afectadas: {karten}',
    getan: 'Reemplazado: coincidencias: {treffer} · cartas: {karten}',
    ruecknahme: 'Con ⌘Z se deshace toda la operación.',
    stellen: 'Lugares afectados',
    kartenfeld: 'Carta {nummer} · {feld}',
    weitere: 'Los demás lugares no se enumeran, pero se reemplazan igualmente.',
  },

  statistik: {
    worueber: 'Contar por',
    belegungWahl: 'Cobertura por campo',
    gruppeSpielpaket: 'Según el paquete de juego',
    gruppeFelder: 'Por valor de campo',
    verteilungVon:
      'Distribución de «{feld}»: proporción de las cartas que lo llevan',
    nachDimension: 'Estadísticas por «{dimension}»',
    titel: 'Estadísticas',
    karten: 'cartas',
    felder: 'campos distintos',
    leere: 'cartas sin un solo valor',
    belegung: 'Cobertura por campo: proporción de todas las cartas del set',
    belegungLeer: 'Este set aún no lleva ningún valor de campo.',
    verteilungLeer: 'Este campo no lleva ningún valor en este set.',
    anteil: 'Proporción',
    dimensionenErklaerung:
      'Un paquete de juego indica por sí mismo qué se puede contar: categoría ' +
      'de color, valor de maná, rareza. Cada dimensión calcula su valor por ' +
      'carta y por eso tarda un momento.',
    dimensionLeer: 'Ninguna carta lleva un valor en esta dimensión.',
    gruppeHinweis: 'Elige una barra para ver sus cartas en la lista de cartas.',
    gruppeGewaehlt: 'Elegido: «{gruppe}» · cartas: {karten}',
    gruppeZeigen: 'Mostrar cartas',
    gruppeAufheben: 'Quitar la selección',
    gruppeLeer: 'Ninguna carta lleva este valor.',
    filterBezeichnung: 'Filtro de las estadísticas',
    filterKopf: 'De las estadísticas: «{quelle}» = «{gruppe}»',
    filterZurStatistik: 'Volver a las estadísticas',
    filterAufheben: 'Quitar el filtro',
  },
  // Die Lasche „Style“: Stilpaket und Stiloptionen des Sets (#288).
  stil: {
    titel: 'Estilo',
    vorschauTitel: 'Vista previa',
    vorschauLeer:
      'Este set aún no tiene ninguna carta: crea una para ver el estilo.',
    eigeneFarbe: 'color personalizado',
    vorgabe: 'Predeterminado',
  },
  suche: {
    titel: 'Búsqueda',
    eingabe: '¿Qué buscas?',
    recherche: 'Investigar …',
    feld: 'Solo en este campo',
    alleFelder: 'Todos los campos',
    set: 'Solo en este set',
    alleSets: 'Todos los sets',
    leer: 'Todavía no se ha buscado nada. Escribe algo para encontrar cartas.',
    nichts: 'No se encontró nada. ¿Probar con otra palabra?',
    gekuerzt:
      'Hay más resultados de los que se muestran aquí. Acota la búsqueda para verlos.',
    treffer: 'Resultados',
  },
  ausgeben,
  hilfe: {
    titel: 'Ayuda',
    erklaerung:
      'Qué significan los términos con los que trabaja esta aplicación y qué hacer en el primer inicio.',
    glossar: 'Términos',

    // **Die Beschriftung der Laschenleiste** — seit U11 sind es Laschen
    // und keine Sprungmarken mehr; der Name ist geblieben, weil er sagt,
    // was die Leiste tut.
    aufDieserSeite: 'En esta página',
    seitenuebersicht: 'Qué hace cada página',
  },

  bestandspflege,
  datenpflege,
  kosten,
  ruecksetzung,
  ruecksprung,

  einstellungen,

  nutzer,
} as const
