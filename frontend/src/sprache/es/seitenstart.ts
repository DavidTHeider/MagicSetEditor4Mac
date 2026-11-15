// **Die Startseite** — herausgelöst, weil `seiten.ts` mit ihr über 400 Zeilen
// ginge (§2).
//
// **Ein Katalog darf sich auf mehrere Dateien verteilen**, solange er als
// **einer** ankommt — so wie `seitenkarten.ts`. Was gleich bleibt, ist die
// Zahl der Kataloge; an ihr hängt der Vertragstest gegen die Go-Seite (§9).
export const start = {
  ueberblick: {
    titel: 'Tu resumen',
    mitProblemen: '{anzahl} con problemas',
    geprueftVon: '{geprueft} de {gesamt} revisados',
    mitBefund: '{anzahl} con hallazgos',
    nichtGeprueft: 'no revisado del todo',
    problemeGeprueft:
      'Problemas revisados {wann} · {geprueft} de {gesamt} cartas',
    problemeOffen: 'Problemas: {offen} de {gesamt} sets sin revisar',
    pruefen: 'Revisar ahora',
    setsMitProblemen: '{sets} con problemas',
    kartenMitBefund: '{karten} con hallazgos en {sets}',
    aeltesterPreis: 'precio más antiguo {wann}',
    ohnePreis: '{anzahl} sin precio',
    mitFehlenden: '{anzahl} con cartas que faltan',
    decksMitFehlenden: '{decks} con cartas que faltan',
    alter: {
      gerade: 'ahora mismo',
      minuten: 'hace {minuten} min',
      stunden: 'hace {stunden} h',
      tage: 'hace {tage} días',
      tag: 'hace 1 día',
    },
  },
  // Resume — work item U11 of phase 5.1 (#1422).
  // Attention and background tasks (#1422).
  aufmerksamkeit: {
    titel: 'Requiere atención',
    nichts: 'Nada: {sets}, {decks} y {pakete} revisados.',
    durchgesehen: '{sets}, {decks} y {pakete} revisados.',
    arten: {
      setVorbehalte: '{anzahl} sin interpretar',
      deckFehlend: '{anzahl} faltan en la colección',
      paketBefund: '{anzahl} en scripts',
    },
    paketeMitBefund: '{pakete} con hallazgos en el script',
    gruendeZeigen: 'Mostrar motivos',
    gruendeVerbergen: 'Ocultar motivos',
    setBefunde: '{karten} con hallazgos',
    befunde: 'Sin revisar:',
    alle: 'Mostrar los {anzahl}',
    weniger: 'Mostrar menos',
    gruende: {
      vorbehalt: '{0}, línea {1}',
    },
    amOrt: '{ort}: {satz}',
    // The operations band (OPS02, #2348).
    betrieb: {
      arten: {
        platzKnapp: 'Solo quedan {frei} libres junto a la biblioteca',
        sicherungAlt: 'Última copia de seguridad hace {tage} días',
        dienstBeendet:
          'El servicio de inicio de sesión terminó con el código {ausgang}',
        webprozessNeu: 'La vista web se recargó {anzahl} veces',
      },
      rat: {
        platzKnapp:
          'Con menos de unos 2 GB o del 5 % libre, cada copia va justa — libera espacio en «Datos y copias de seguridad».',
        sicherungAlt:
          'Tiene más de 14 días — haz una copia de la biblioteca en «Datos y copias de seguridad».',
        dienstBeendet:
          'Tras un error, launchd lo reinicia — el diagnóstico dice por qué terminó.',
        webprozessNeu:
          'WebKit había terminado su proceso; lo que no se guardó en un formulario abierto puede faltar — revísalo en el diagnóstico.',
      },
    },
  },
  auftraege: {
    titel: 'Tareas en segundo plano',
    staende: {
      laeuft: 'en curso',
      fertig: 'terminado',
      gescheitert: 'fallido',
      abgebrochen: 'cancelado',
    },
    arten: {
      preise: 'Obtener precios de la colección',
      sammeldaten: 'Cargar datos masivos de Scryfall',
      fassung: 'Obtener una versión nueva',
      bogen: 'Componer una hoja de impresión',
      set: 'Importar un set',
      mtgjson: 'Importar MTG JSON',
      vorlage: 'Exportar con una plantilla',
      installer: 'Leer un instalador',
      weitergabe: 'Compartir como instalador',
      probleme: 'Revisar los sets en busca de problemas',
    },
    abbrechen: 'Cancelar: {art}',
    dauer: '{sekunden} s',
    laufend: { eine: 'tarea en curso', viele: 'tareas en curso' },
  },
  weiterarbeiten: {
    titel: 'Continuar',
    weiterAn: 'Continuar con {art} «{name}»',
    zuletztWann: 'Última vez: {vorgang} el {zeit}',
    angeheftet: 'fijado',
    angeheftetLeer: 'Nada fijado: fija una fila para mantenerla aquí arriba.',
    leer: 'Aquí aparece aquello en lo que has trabajado hace poco. De momento está vacío: el primer paso es un set o un paquete de plantillas.',
    handlungSets: 'Crear o abrir un set',
    handlungPakete: 'Obtener un paquete',
    anheften: 'Fijar',
    loesen: 'Desfijar',
    anheftenVon: 'Fijar «{name}»',
    loesenVon: 'Desfijar «{name}»',
    arten: { set: 'Set', deck: 'Mazo', paket: 'Paquete' },
    vorgaenge: {
      geoeffnet: 'abierto',
      eingefuehrt: 'importado',
      ausgegeben: 'exportado',
    },
  },
  titel: 'Inicio',
  ueberschrift: 'MagicSetEditor4Mac',
  was: 'Diseña, revisa e imprime cartas de Magic y de otros juegos, con las plantillas que usa Magic Set Editor 2.',
  ersteSchritte: 'Primeros pasos',
  einfuehrung: {
    willkommen:
      'Tres pasos y la primera carta está lista. Puedes volver a abrir esta ventana en cualquier momento desde los ajustes.',
    wiederzeigen: 'Mostrar de nuevo al iniciar',
    schliessen: 'Vamos allá',
    bezeichnung: 'Introducción',
  },
  schritte: {
    setOeffnen: {
      titel: 'Abrir un set',
      erklaerung:
        'Abre un archivo .mse-set existente o importa un set desde MTG JSON.',
      handlung: 'Ir a los sets',
    },
    kartenBearbeiten: {
      titel: 'Editar cartas',
      erklaerung:
        'Consulta y modifica las cartas del set abierto; la vista previa muestra cada cambio al instante.',
      handlung: 'Ir a las cartas',
      ersteinSet: 'Para esto primero debe haber un set abierto.',
      stattdessen: 'Abre primero un set',
    },
    paketeHolen: {
      titel: 'Obtener plantillas',
      erklaerung:
        'Ya hay ocho paquetes de plantillas. ¿Quieres más juegos o estilos? Descárgalos aquí.',
      handlung: 'Ir a los paquetes',
    },
  },
}
