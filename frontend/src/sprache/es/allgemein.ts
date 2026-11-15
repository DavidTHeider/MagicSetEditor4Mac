// Frame, navigation, language selection.
export const allgemein = {
  // **Count words** (#1218) — a number and its word in the matching form.
  zaehlwoerter: {
    karten: { eine: 'carta', viele: 'cartas' },

    // **Und die Seiten** (#1219):
    // der Druckbogen meldete »1 Karten auf 1 Seiten«.
    seiten: { eine: 'página', viele: 'páginas' },

    // And the three objects of the overview (#1422).
    sets: { eine: 'set', viele: 'sets' },
    decks: { eine: 'mazo', viele: 'mazos' },
    pakete: { eine: 'paquete', viele: 'paquetes' },
    eigenePakete: { eine: 'paquete propio', viele: 'paquetes propios' },
    stile: { eine: 'estilo', viele: 'estilos' },
    positionen: { eine: 'posición', viele: 'posiciones' },
    exemplare: { eine: 'ejemplar', viele: 'ejemplares' },

    // **Und die Drucke einer Sichtfreigabe** (UXR-F18, #2033).
    drucke: { eine: 'impresión', viele: 'impresiones' },
  },
  bestandGesichert: 'Se ha guardado una copia de la base de datos en',
  zahlenbereich: 'Un número entre {kleinstens} y {groesstens}.',
  zahlenAb: 'Un número a partir de {kleinstens}.',
  ganzeZahlBereich: 'Un número entero entre {kleinstens} y {groesstens}.',
  ganzeZahlAb: 'Un número entero a partir de {kleinstens}.',
  nurZiffern:
    'Solo dígitos, sin punto ni coma — el importe se indica en la unidad más pequeña.',

  // **Ein Doppelklick im Finder, der nicht aufging** (#891).
  dateiNichtGeoeffnet: 'No se pudo abrir el archivo {pfad}. {grund}',
  anwendungsname: 'MagicSetEditor4Mac',
  sprache: {
    bezeichnung: 'Idioma',
    de: 'Deutsch',
    en: 'English',
    es: 'Español',
    fr: 'Français',
    it: 'Italiano',
    ja: '日本語',
    ru: 'Русский',
    zh: '简体中文',
  },
  // **Die fünf Arbeitsbereiche, der Einstieg und die Dienstzone** — UX-02
  // (#1674).
  //
  // **`bereiche` und `dienste` sind die zwei Überschriften der Leiste**, nicht
  // Bereiche selbst — sie sagen, was darunter steht.
  bereiche: {
    // **„Home" und nicht „Start"** — so heißt die Seite in dieser Sprache
    // (`en/seitenstart.ts`), und der Bereich trägt ihren Namen.
    start: 'Inicio',
    gestalten: 'Diseñar',
    sammeln: 'Coleccionar',
    spielen: 'Jugar',
    handeln: 'Intercambiar',
    ressourcen: 'Recursos',
    programm: 'Aplicación',

    ueberschrift: 'Áreas de trabajo',
  },

  navigation: {
    bezeichnung: 'Navegación principal',
    sprungZumInhalt: 'Saltar al contenido',
    seitenleisteEinblenden: 'Mostrar la barra lateral',
    seitenleisteAusblenden: 'Ocultar la barra lateral',
    seitenleisteKuerzel: '⌥⌘S',
    weg: 'Estás aquí',
  },

  // When something does not draw — F01 (#2340).
  fehlergrenze: {
    bereichTitel: 'No se pudo dibujar esta área',
    bereichSatz:
      'Referencia {kennung}. El resto de la base de datos está intacto; el error figura en el registro con esta referencia.',
    neuLaden: 'Recargar el área',
    diagnoseOeffnen: 'Abrir el diagnóstico',
    wurzelTitel: 'No se pudo dibujar la ventana',
    wurzelSatz:
      'Referencia {kennung}. Tu base de datos está intacta; el error figura en el registro con esta referencia.',
    fensterNeuLaden: 'Recargar la ventana',
  },
  meldungen: {
    methodeFehlt: 'El método {name} no está conectado.',
    vorschauLeer: 'La vista previa llegó sin imagen y sin error.',
    kontoLeer: 'La cuenta llegó sin usuario y sin error.',
  },
  // Das Band über dem Inhalt: was zu keiner Seite gehört (#180).
  meldungsband: {
    bereich: 'Mensajes de la ventana',
    setNichtGesichert: 'El set «{name}» no se guardó — {grund}',
    setNichtGesichertOhneName: 'El set no se guardó — {grund}',
    bestandNichtGesichert:
      'No se hizo la copia de seguridad de la base de datos — {grund}',
    verlaufGescheitert: 'Deshacer o rehacer no funcionó — {grund}',
    darstellungGescheitert: 'No se pudo restablecer la apariencia — {grund}',
  },
  // Die neue Fassung, wenn eine da ist (#1100).
  vertraeglichkeit: {
    vertraeglich: 'Esta versión puede abrir tu biblioteca.',
    unvertraeglich: 'Esta versión no puede abrir tu biblioteca.',
    ungeklaert:
      'Esta versión no indica si puede abrir tu biblioteca: es anterior a esta comprobación. Haz una copia de seguridad de tu biblioteca antes de cambiar y restáurala si algo sale mal.',
    bereit: 'El cambio está preparado: copia de seguridad hecha y releída.',
    vorbereiten: 'Preparar el cambio',
    hindernisse: {
      keinManifest: 'La versión publicada no dice nada sobre sí misma.',
      plattform: 'Está compilada para otro equipo.',
      wanderungFehlt:
        'Tu biblioteca tiene un cambio que esa versión no conoce.',
      summeAnders: 'Un cambio tiene allí un contenido distinto.',
      vorDerGrundlinie:
        'Tu biblioteca es anterior a su línea base: ábrela primero con una versión intermedia.',
      platz: 'No queda espacio libre suficiente.',
      vorgaengeLaufen: 'Todavía hay algo en marcha.',
      keineSicherung: 'No se ha hecho ninguna copia de seguridad.',
      sicherungUnlesbar: 'No se pudo releer la copia de seguridad.',
      halter:
        'Otro programa mantiene abierta la biblioteca: ciérralo antes del cambio.',
    },
  },
  fassungsband: {
    bereich: 'Nueva versión',
    neu: 'Ya está disponible la versión {fassung}.',
    seite: 'Qué ha cambiado',
    laden: 'Descargar',
    laeuft: 'Descargando …',
    fortschritt: '{gelesen} de {ganz} descargados',
    fortschrittOhneLaenge: '{gelesen} descargados',
    geladen:
      'La versión {fassung} está en {pfad}. Para instalarla, ábrela en el Finder y arrastra la aplicación a la carpeta Aplicaciones.',
    weglegen: 'Cerrar este aviso',
  },
  wiederanlauf: {
    bereich: 'Última sesión',
    satz: 'La última sesión, iniciada el {zeit}, terminó de forma inesperada.',
    satzOhneZeit: 'La última sesión terminó de forma inesperada.',
    mitSpur:
      'Hay un rastro del fallo; el paquete de diagnóstico lo incluye cuando lo creas.',
    ohneSpur:
      'No hay rastro del fallo: probablemente la aplicación se cerró desde fuera.',
    ausblenden: 'Ocultar',
  },

  // Der Arbeitsplatz eines Sets: seine Laschen und sein leerer Zustand (#288).
  arbeitsplatz: {
    leiste: {
      bezeichnung: 'Cifras del set',
      befund: 'Con hallazgos',
      nichtGeprueft: 'sin comprobar',
      nichtGelesen: 'no leído',
      geprueft: 'comprobado {wann}',
    },
    laschen: 'Áreas del set',

    gruppen: {
      bauen: 'Crear cartas',
    },
    schubladen: {
      eigenschaften: {
        knopf: 'Propiedades',
        titel: 'Propiedades del set',
        laschen: 'Áreas de las propiedades',
        hinweis: 'Lo que vale para todas las cartas de este set.',
      },
      werkzeuge: {
        knopf: 'Herramientas',
        titel: 'Herramientas',
        laschen: 'Herramientas del set',
        hinweis: 'Estadísticas, sobre aleatorio y consola para este set.',
      },
      schliessen: 'Cerrar',
    },

    keinSet: 'Ningún set abierto',
    keinSetErklaerung:
      'Elige un set en «Sets», crea uno o impórtalo: aquí aparecerán sus cartas, sus estadísticas y la hoja de impresión.',
    zuDenSets: 'Ir a Sets',
    zuDenPaketen: 'Ir a Paquetes',
    zuDenKarten: 'Ir a Cartas',
    ohneKartenTitel: 'Este set todavía no tiene cartas',
    ohneKarten:
      'Esta pestaña trabaja con las cartas del set. Las cartas se crean en «Cartas»: nuevas, pegadas o importadas.',
    ohneNamen: 'Set abierto',
  },
  palette: {
    bezeichnung: 'Paleta de comandos',
    eingabe: '¿Adónde?',
    keineTreffer: 'No se ha encontrado nada. ¿Probamos con otra palabra?',
    oeffnen: 'Paleta de comandos',
    seiten: 'Páginas',
    befehle: 'Comandos',
    befehlEingabe: '¿Adónde o qué comando?',
    tun: 'Hacer',
    finden: 'Buscar',
    karteErstellen: 'Crear carta',
    setWaehlen: 'Crear o abrir un set',
    deckSpielen: 'Partida rápida con {name}',
    tauschenMit: 'Intercambiar cartas con {name}',
    sicherungErstellen: 'Crear copia de seguridad',
    schnittstellenEinrichten:
      'Configurar las interfaces (línea de comandos, MCP, REST)',
    woKarte: 'Diseñar · {set}',
    woSets: 'Diseñar · Sets',
    woPartie: 'Jugar · Empezar una partida',
    woHandel: 'Intercambiar · Contraparte',
    woEinstellungen: 'Ajustes · {gruppe}',
    alsKarte: 'Carta · {set}',
    alsSet: 'Set · {karten}',
    alsDeck: 'Mazo · {spiel}',
    alsGegenueber: 'Contraparte · confirmada',
    fuss: '↑↓ elegir · ↵ ejecutar · esc cerrar',
  },
} as const
