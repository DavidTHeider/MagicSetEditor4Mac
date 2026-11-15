// Frame, navigation, language selection.
export const allgemein = {
  // **Count words** (#1218) — a number and its word in the matching form.
  zaehlwoerter: {
    karten: { eine: 'card', viele: 'cards' },

    // **Und die Seiten** (#1219):
    // der Druckbogen meldete »1 Karten auf 1 Seiten«.
    seiten: { eine: 'page', viele: 'pages' },

    // And the three objects of the overview (#1422).
    sets: { eine: 'set', viele: 'sets' },
    decks: { eine: 'deck', viele: 'decks' },
    pakete: { eine: 'package', viele: 'packages' },
    eigenePakete: { eine: 'own package', viele: 'own packages' },
    stile: { eine: 'style', viele: 'styles' },
    positionen: { eine: 'position', viele: 'positions' },
    exemplare: { eine: 'copy', viele: 'copies' },

    // **Und die Drucke einer Sichtfreigabe** (UXR-F18, #2033).
    drucke: { eine: 'printing', viele: 'printings' },
  },
  bestandGesichert: 'The database was backed up to',
  zahlenbereich: 'A number between {kleinstens} and {groesstens}.',
  zahlenAb: 'A number from {kleinstens} upwards.',
  ganzeZahlBereich: 'A whole number between {kleinstens} and {groesstens}.',
  ganzeZahlAb: 'A whole number from {kleinstens} upwards.',
  nurZiffern:
    'Digits only, without point or comma — the amount is in the smallest unit.',

  // **Ein Doppelklick im Finder, der nicht aufging** (#891).
  dateiNichtGeoeffnet: 'The file {pfad} could not be opened. {grund}',
  anwendungsname: 'MagicSetEditor4Mac',
  sprache: {
    bezeichnung: 'Language',
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
    start: 'Home',
    gestalten: 'Design',
    sammeln: 'Collect',
    spielen: 'Play',
    handeln: 'Trade',
    ressourcen: 'Resources',
    programm: 'Application',

    ueberschrift: 'Work areas',
  },

  navigation: {
    bezeichnung: 'Main navigation',
    sprungZumInhalt: 'Skip to content',
    seitenleisteEinblenden: 'Show sidebar',
    seitenleisteAusblenden: 'Hide sidebar',
    seitenleisteKuerzel: '⌥⌘S',
    weg: 'You are here',
  },

  // When something does not draw — F01 (#2340).
  fehlergrenze: {
    bereichTitel: 'This area could not be drawn',
    bereichSatz:
      'Reference {kennung}. The rest of the database is untouched; the error is in the log under this reference.',
    neuLaden: 'Reload area',
    diagnoseOeffnen: 'Open diagnostics',
    wurzelTitel: 'The window could not be drawn',
    wurzelSatz:
      'Reference {kennung}. Your database is untouched; the error is in the log under this reference.',
    fensterNeuLaden: 'Reload window',
  },
  meldungen: {
    methodeFehlt: 'The method {name} is not wired up.',
    vorschauLeer:
      'The preview came back without an image and without an error.',
    kontoLeer: 'The account came back without a user and without an error.',
  },
  // Das Band über dem Inhalt: was zu keiner Seite gehört (#180).
  meldungsband: {
    bereich: 'Window messages',
    setNichtGesichert: 'The set “{name}” was not saved — {grund}',
    setNichtGesichertOhneName: 'The set was not saved — {grund}',
    bestandNichtGesichert: 'The database was not backed up — {grund}',
    verlaufGescheitert: 'Undo or redo did not work — {grund}',
    darstellungGescheitert: 'The appearance could not be reset — {grund}',
  },
  // Die neue Fassung, wenn eine da ist (#1100).
  vertraeglichkeit: {
    vertraeglich: 'This version can open your library.',
    unvertraeglich: 'This version cannot open your library.',
    ungeklaert:
      'This version does not say whether it can open your library — it predates this check. Back your library up before switching, and restore it if anything goes wrong.',
    bereit: 'The switch is prepared: backed up and read back.',
    vorbereiten: 'Prepare switch',
    hindernisse: {
      keinManifest: 'The release says nothing about itself.',
      plattform: 'It was built for a different machine.',
      wanderungFehlt: 'Your library has a change it does not know.',
      summeAnders: 'One change has different contents there.',
      vorDerGrundlinie:
        'Your library predates its baseline — open it with a version in between first.',
      platz: 'There is not enough free space.',
      vorgaengeLaufen: 'Something is still running.',
      keineSicherung: 'No backup has been made.',
      sicherungUnlesbar: 'The backup could not be read back.',
      halter:
        'Another program is holding the library open — quit it before switching.',
    },
  },
  fassungsband: {
    bereich: 'New version',
    neu: 'Version {fassung} is out.',
    seite: 'What changed',
    laden: 'Download',
    laeuft: 'Downloading …',
    fortschritt: '{gelesen} of {ganz} downloaded',
    fortschrittOhneLaenge: '{gelesen} downloaded',
    geladen:
      'Version {fassung} is at {pfad}. To install, open it in the Finder and drag the application into the Applications folder.',
    weglegen: 'Dismiss this message',
  },
  wiederanlauf: {
    bereich: 'Last session',
    satz: 'The last session, started {zeit}, ended unexpectedly.',
    satzOhneZeit: 'The last session ended unexpectedly.',
    mitSpur:
      'A crash trace is available; the diagnostic bundle includes it when you create one.',
    ohneSpur:
      'There is no crash trace — the app was probably ended from outside.',
    ausblenden: 'Dismiss',
  },

  // Der Arbeitsplatz eines Sets: seine Laschen und sein leerer Zustand (#288).
  arbeitsplatz: {
    leiste: {
      bezeichnung: 'Set figures',
      befund: 'With findings',
      nichtGeprueft: 'not checked',
      nichtGelesen: 'not read',
      geprueft: 'checked {wann}',
    },
    laschen: 'Areas of the set',

    gruppen: {
      bauen: 'Build cards',
    },
    schubladen: {
      eigenschaften: {
        knopf: 'Properties',
        titel: 'Set properties',
        laschen: 'Areas of the properties',
        hinweis: 'What applies to every card in this set.',
      },
      werkzeuge: {
        knopf: 'Tools',
        titel: 'Tools',
        laschen: 'Tools of the set',
        hinweis: 'Statistics, random pack and console for this set.',
      },
      schliessen: 'Close',
    },

    keinSet: 'No set open',
    keinSetErklaerung:
      'Pick a set under “Sets”, create one or read one in — its cards, statistics and print sheet will appear here.',
    zuDenSets: 'Go to Sets',
    zuDenPaketen: 'Go to Packages',
    zuDenKarten: 'Go to Cards',
    ohneKartenTitel: 'No card in this set yet',
    ohneKarten:
      'This tab works with the set’s cards. Cards come from “Cards” — created, pasted or imported.',
    ohneNamen: 'Open set',
  },
  palette: {
    bezeichnung: 'Command palette',
    eingabe: 'Where to?',
    keineTreffer: 'Nothing found. Try another word?',
    oeffnen: 'Command palette',
    seiten: 'Pages',
    befehle: 'Commands',
    befehlEingabe: 'Where to — or which command?',
    tun: 'Do',
    finden: 'Find',
    karteErstellen: 'Create card',
    setWaehlen: 'Create or open a set',
    deckSpielen: 'Quick play with {name}',
    tauschenMit: 'Trade cards with {name}',
    sicherungErstellen: 'Create backup',
    schnittstellenEinrichten: 'Set up interfaces (command line, MCP, REST)',
    woKarte: 'Design · {set}',
    woSets: 'Design · Sets',
    woPartie: 'Play · Start a game',
    woHandel: 'Trade · Counterpart',
    woEinstellungen: 'Settings · {gruppe}',
    alsKarte: 'Card · {set}',
    alsSet: 'Set · {karten}',
    alsDeck: 'Deck · {spiel}',
    alsGegenueber: 'Counterpart · confirmed',
    fuss: '↑↓ choose · ↵ run · esc close',
  },
} as const
