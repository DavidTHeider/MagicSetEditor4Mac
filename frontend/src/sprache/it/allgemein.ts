// Frame, navigation, language selection.
export const allgemein = {
  // **Count words** (#1218) — a number and its word in the matching form.
  zaehlwoerter: {
    karten: { eine: 'carta', viele: 'carte' },

    // **Und die Seiten** (#1219):
    // der Druckbogen meldete »1 Karten auf 1 Seiten«.
    seiten: { eine: 'pagina', viele: 'pagine' },

    // And the three objects of the overview (#1422).
    sets: { eine: 'set', viele: 'set' },
    decks: { eine: 'mazzo', viele: 'mazzi' },
    pakete: { eine: 'pacchetto', viele: 'pacchetti' },
    eigenePakete: { eine: 'pacchetto proprio', viele: 'pacchetti propri' },
    stile: { eine: 'stile', viele: 'stili' },
    positionen: { eine: 'posizione', viele: 'posizioni' },
    exemplare: { eine: 'esemplare', viele: 'esemplari' },

    // **Und die Drucke einer Sichtfreigabe** (UXR-F18, #2033).
    drucke: { eine: 'stampa', viele: 'stampe' },
  },
  bestandGesichert: 'Backup del database salvato in',
  zahlenbereich: 'Un numero compreso tra {kleinstens} e {groesstens}.',
  zahlenAb: 'Un numero da {kleinstens} in su.',
  ganzeZahlBereich:
    'Un numero intero compreso tra {kleinstens} e {groesstens}.',
  ganzeZahlAb: 'Un numero intero da {kleinstens} in su.',
  nurZiffern:
    'Solo cifre, senza punto né virgola — l’importo è espresso nell’unità più piccola.',

  // **Ein Doppelklick im Finder, der nicht aufging** (#891).
  dateiNichtGeoeffnet: 'Impossibile aprire il file {pfad}. {grund}',
  anwendungsname: 'MagicSetEditor4Mac',
  sprache: {
    bezeichnung: 'Lingua',
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
    gestalten: 'Progetta',
    sammeln: 'Colleziona',
    spielen: 'Gioca',
    handeln: 'Scambia',
    ressourcen: 'Risorse',
    programm: 'Applicazione',

    ueberschrift: 'Aree di lavoro',
  },

  navigation: {
    bezeichnung: 'Navigazione principale',
    sprungZumInhalt: 'Vai al contenuto',
    seitenleisteEinblenden: 'Mostra la barra laterale',
    seitenleisteAusblenden: 'Nascondi la barra laterale',
    seitenleisteKuerzel: '⌥⌘S',
    weg: 'Sei qui',
  },

  // When something does not draw — F01 (#2340).
  fehlergrenze: {
    bereichTitel: 'Non è stato possibile disegnare quest’area',
    bereichSatz:
      'Riferimento {kennung}. Il resto del database è intatto; l’errore si trova nel registro con questo riferimento.',
    neuLaden: 'Ricarica l’area',
    diagnoseOeffnen: 'Apri la diagnostica',
    wurzelTitel: 'Non è stato possibile disegnare la finestra',
    wurzelSatz:
      'Riferimento {kennung}. Il tuo database è intatto; l’errore si trova nel registro con questo riferimento.',
    fensterNeuLaden: 'Ricarica la finestra',
  },
  meldungen: {
    methodeFehlt: 'Il metodo {name} non è collegato.',
    vorschauLeer: 'L’anteprima è tornata senza immagine e senza errore.',
    kontoLeer: 'L’account è tornato senza utente e senza errore.',
  },
  // Das Band über dem Inhalt: was zu keiner Seite gehört (#180).
  meldungsband: {
    bereich: 'Messaggi della finestra',
    setNichtGesichert: 'Il set «{name}» non è stato salvato — {grund}',
    setNichtGesichertOhneName: 'Il set non è stato salvato — {grund}',
    bestandNichtGesichert: 'Il backup del database non è riuscito — {grund}',
    verlaufGescheitert: 'Impossibile annullare o ripetere — {grund}',
    darstellungGescheitert: 'Impossibile ripristinare l’aspetto — {grund}',
  },
  // Die neue Fassung, wenn eine da ist (#1100).
  vertraeglichkeit: {
    vertraeglich: 'Questa versione può aprire la tua libreria.',
    unvertraeglich: 'Questa versione non può aprire la tua libreria.',
    ungeklaert:
      'Questa versione non dice se può aprire la tua libreria — è precedente a questo controllo. Fai un backup della libreria prima di passare a un’altra versione e ripristinala se qualcosa va storto.',
    bereit: 'Il passaggio è pronto: backup eseguito e riletto con successo.',
    vorbereiten: 'Prepara il passaggio',
    hindernisse: {
      keinManifest: 'La release non dice nulla su di sé.',
      plattform: 'È stata compilata per un altro computer.',
      wanderungFehlt: 'La tua libreria contiene una modifica che non conosce.',
      summeAnders: 'Una modifica ha lì un contenuto diverso.',
      vorDerGrundlinie:
        'La tua libreria è precedente alla sua linea di base — aprila prima con una versione intermedia.',
      platz: 'Non c’è abbastanza spazio libero.',
      vorgaengeLaufen: 'Qualcosa è ancora in esecuzione.',
      keineSicherung: 'Non è stato eseguito nessun backup.',
      sicherungUnlesbar: 'Impossibile rileggere il backup.',
      halter:
        'Un altro programma tiene aperta la libreria: chiudilo prima del cambio.',
    },
  },
  fassungsband: {
    bereich: 'Nuova versione',
    neu: 'È disponibile la versione {fassung}.',
    seite: 'Cosa è cambiato',
    laden: 'Scarica',
    laeuft: 'Download in corso …',
    fortschritt: '{gelesen} di {ganz} scaricati',
    fortschrittOhneLaenge: '{gelesen} scaricati',
    geladen:
      'La versione {fassung} si trova in {pfad}. Per installarla, aprila nel Finder e trascina l’applicazione nella cartella Applicazioni.',
    weglegen: 'Chiudi questo messaggio',
  },
  wiederanlauf: {
    bereich: 'Ultima sessione',
    satz: 'L’ultima sessione, iniziata il {zeit}, si è chiusa in modo inatteso.',
    satzOhneZeit: 'L’ultima sessione si è chiusa in modo inatteso.',
    mitSpur:
      'È disponibile una traccia del crash; il pacchetto diagnostico la include quando lo crei.',
    ohneSpur:
      'Non c’è alcuna traccia del crash: probabilmente l’app è stata chiusa dall’esterno.',
    ausblenden: 'Nascondi',
  },

  // Der Arbeitsplatz eines Sets: seine Laschen und sein leerer Zustand (#288).
  arbeitsplatz: {
    leiste: {
      bezeichnung: 'Cifre del set',
      befund: 'Con rilievi',
      nichtGeprueft: 'non verificato',
      nichtGelesen: 'non letto',
      geprueft: 'verificato {wann}',
    },
    laschen: 'Aree del set',

    gruppen: {
      bauen: 'Costruisci le carte',
    },
    schubladen: {
      eigenschaften: {
        knopf: 'Proprietà',
        titel: 'Proprietà del set',
        laschen: 'Aree delle proprietà',
        hinweis: 'Ciò che vale per tutte le carte di questo set.',
      },
      werkzeuge: {
        knopf: 'Strumenti',
        titel: 'Strumenti',
        laschen: 'Strumenti del set',
        hinweis: 'Statistiche, bustina casuale e console per questo set.',
      },
      schliessen: 'Chiudi',
    },

    keinSet: 'Nessun set aperto',
    keinSetErklaerung:
      'Scegli un set in «Set», creane uno o importane uno — qui compariranno le sue carte, le statistiche e il foglio di stampa.',
    zuDenSets: 'Vai ai set',
    zuDenPaketen: 'Vai ai pacchetti',
    zuDenKarten: 'Vai alle carte',
    ohneKartenTitel: 'Ancora nessuna carta in questo set',
    ohneKarten:
      'Questa scheda lavora con le carte del set. Le carte nascono in «Carte»: create, incollate o importate.',
    ohneNamen: 'Set aperto',
  },
  palette: {
    bezeichnung: 'Palette comandi',
    eingabe: 'Dove vuoi andare?',
    keineTreffer: 'Nessun risultato. Prova un’altra parola?',
    oeffnen: 'Palette comandi',
    seiten: 'Pagine',
    befehle: 'Comandi',
    befehlEingabe: 'Dove vuoi andare, o quale comando?',
    tun: 'Esegui',
    finden: 'Trova',
    karteErstellen: 'Crea carta',
    setWaehlen: 'Crea o apri un set',
    deckSpielen: 'Partita rapida con {name}',
    tauschenMit: 'Scambia carte con {name}',
    sicherungErstellen: 'Crea backup',
    schnittstellenEinrichten:
      'Configura le interfacce (riga di comando, MCP, REST)',
    woKarte: 'Progetta · {set}',
    woSets: 'Progetta · Set',
    woPartie: 'Gioca · Inizia una partita',
    woHandel: 'Scambia · Controparte',
    woEinstellungen: 'Impostazioni · {gruppe}',
    alsKarte: 'Carta · {set}',
    alsSet: 'Set · {karten}',
    alsDeck: 'Mazzo · {spiel}',
    alsGegenueber: 'Controparte · confermata',
    fuss: '↑↓ scegli · ↵ esegui · esc chiudi',
  },
} as const
