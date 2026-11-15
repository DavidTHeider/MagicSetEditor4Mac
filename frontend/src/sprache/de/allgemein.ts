// Rahmen, Navigation, Sprachwahl.
export const allgemein = {
  // **Zählwörter** (#1218) — eine Zahl und ihr Wort in der passenden Form.
  zaehlwoerter: {
    karten: { eine: 'Karte', viele: 'Karten' },

    // **Und die Seiten** (#1219):
    // der Druckbogen meldete »1 Karten auf 1 Seiten«.
    seiten: { eine: 'Seite', viele: 'Seiten' },

    // **Und die drei Gegenstände der Übersicht** — Werkstück U11 von Phase 5.1
    // (#1422):
    // „Nichts — 1 Sets durchgesehen“ wäre derselbe Fehler wie beim Druckbogen.
    sets: { eine: 'Set', viele: 'Sets' },
    decks: { eine: 'Deck', viele: 'Decks' },
    pakete: { eine: 'Paket', viele: 'Pakete' },
    // **Eigene Pakete** — die mitgelieferten liegen im Bündel und werden
    // nicht durchgesehen (G01, #2248).
    eigenePakete: { eine: 'eigenes Paket', viele: 'eigene Pakete' },
    stile: { eine: 'Stil', viele: 'Stile' },
    positionen: { eine: 'Position', viele: 'Positionen' },

    // **Und die Exemplare einer Auswahl** (UX-03, #1675) — die WebKit-Probe
    // las „1 Exemplare → Deck“.
    exemplare: { eine: 'Exemplar', viele: 'Exemplare' },

    // **Und die Drucke einer Sichtfreigabe** (UXR-F18, #2033).
    drucke: { eine: 'Druck', viele: 'Drucke' },
  },
  bestandGesichert: 'Der Bestand wurde gesichert nach',

  // **Der Satz am Zahlenfeld** — Werkstück F von Phase 4.10
  // (#1325).
  //
  // **Er nennt den Bereich und nicht nur den Fehler.** §6 verlangt, dass ein
  // Hinweis sagt, **was zu tun ist**; „ungültig“ sagt es nicht.
  zahlenbereich: 'Eine Zahl zwischen {kleinstens} und {groesstens}.',

  // **Drei Geschwister, weil die Domäne vier Fälle kennt** — Werkstück U3 von
  // Phase 5.1 (#1414).
  //
  // **Eine Obergrenze gibt es nicht überall.** `domain.Format` prüft die vier
  // Zahlen eines Formats nur gegen `< 0` (`zahlenPruefen`, `format.go`), und
  // `service.Sammlungssichten.Dubletten` weist `behalten < 0` ab. Eine
  // erfundene Obergrenze wiese einen gültigen Wert ab (§11).
  //
  // **Und „ganz“ ist keine Verzierung**: die Felder dahinter sind in Go `int`.
  // Eine Kommazahl käme dort als Umwandlungsfehler an und nicht als Befund am
  // Feld.
  zahlenAb: 'Eine Zahl ab {kleinstens}.',
  ganzeZahlBereich: 'Eine ganze Zahl zwischen {kleinstens} und {groesstens}.',
  ganzeZahlAb: 'Eine ganze Zahl ab {kleinstens}.',
  nurZiffern:
    'Nur Ziffern, ohne Punkt und Komma — der Betrag steht in der kleinsten Einheit.',

  // **Ein Doppelklick im Finder, der nicht aufging** (#891).
  dateiNichtGeoeffnet: 'Die Datei {pfad} ließ sich nicht öffnen. {grund}',
  anwendungsname: 'MagicSetEditor4Mac',
  sprache: {
    bezeichnung: 'Sprache',
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
    start: 'Start',
    gestalten: 'Gestalten',
    sammeln: 'Sammeln',
    spielen: 'Spielen',
    handeln: 'Handeln',
    ressourcen: 'Ressourcen',
    // **`programm` ist zugleich der Name der Dienstzone und ihre
    // Überschrift** — sie ist beides, und zwei Schlüssel dafür wären zwei
    // Wörter für dieselbe Sache.
    programm: 'Programm',

    // **Die Überschrift über den sechs Bereichen — wörtlich aus dem Entwurf**
    // (`docs/prototyp/project/texte.js`, Zeile 17).
    ueberschrift: 'Arbeitsbereiche',
  },

  navigation: {
    bezeichnung: 'Hauptnavigation',
    sprungZumInhalt: 'Zum Inhalt springen',
    seitenleisteEinblenden: 'Seitenleiste einblenden',
    seitenleisteAusblenden: 'Seitenleiste ausblenden',
    seitenleisteKuerzel: '⌥⌘S',

    // Der Weg über dem Kopf der Seite — UX-03.
    weg: 'Wo du bist',
  },

  // **Wenn etwas nicht zeichnet** — F01 (#2340): ein Satz, eine Kennung, ein
  // Rückweg. Dieselbe Kennung steht im Protokoll.
  fehlergrenze: {
    bereichTitel: 'Dieser Bereich ließ sich nicht zeichnen',
    bereichSatz:
      'Kennung {kennung}. Der übrige Bestand ist unberührt; unter dieser Kennung steht der Fehler im Protokoll.',
    neuLaden: 'Bereich neu laden',
    diagnoseOeffnen: 'Diagnose öffnen',
    wurzelTitel: 'Das Fenster ließ sich nicht zeichnen',
    wurzelSatz:
      'Kennung {kennung}. Dein Bestand ist unberührt; unter dieser Kennung steht der Fehler im Protokoll.',
    fensterNeuLaden: 'Fenster neu laden',
  },
  meldungen: {
    // Die Ersatzmethode der Fassade. Sie erscheint nur, wenn eine Bindung
    // fehlt — im Betrieb also nie, in einem unvollständigen Test sofort.
    methodeFehlt: 'Die Methode {name} ist nicht belegt.',
    vorschauLeer: 'Die Vorschau kam ohne Bild und ohne Fehler zurück.',
    kontoLeer: 'Das Konto kam ohne Nutzer und ohne Fehler zurück.',
  },
  // Das Band über dem Inhalt: was zu keiner Seite gehört (#180).
  meldungsband: {
    bereich: 'Meldungen des Fensters',
    // **Ein Fehler nennt seinen Vorgang** — S08.NAT-02 (#1680): „Eintrag
    // nicht gefunden." über der Turnierseite sah aus wie ihr Fehler.
    setNichtGesichert: 'Das Set „{name}“ wurde nicht gesichert — {grund}',
    setNichtGesichertOhneName: 'Das Set wurde nicht gesichert — {grund}',
    bestandNichtGesichert: 'Der Bestand wurde nicht gesichert — {grund}',
    verlaufGescheitert:
      'Zurücknehmen oder Wiederherstellen ging nicht — {grund}',
    darstellungGescheitert:
      'Die Darstellung ließ sich nicht zurücksetzen — {grund}',
  },
  // Die neue Fassung, wenn eine da ist (#1100).
  vertraeglichkeit: {
    vertraeglich: 'Diese Fassung kann deinen Bestand öffnen.',
    unvertraeglich: 'Diese Fassung kann deinen Bestand nicht öffnen.',
    ungeklaert:
      'Ob diese Fassung deinen Bestand öffnen kann, sagt sie nicht — sie ist älter als diese Prüfung. Sichere deinen Bestand, bevor du wechselst, und spiele ihn zurück, wenn etwas klemmt.',
    bereit: 'Der Wechsel ist vorbereitet: gesichert und zurückgelesen.',
    vorbereiten: 'Wechsel vorbereiten',
    hindernisse: {
      keinManifest: 'Die Freigabe sagt nichts über sich.',
      plattform: 'Sie ist für einen anderen Rechner gebaut.',
      wanderungFehlt: 'Dein Bestand hat eine Änderung, die sie nicht kennt.',
      summeAnders: 'Eine Änderung hat bei ihr einen anderen Inhalt.',
      vorDerGrundlinie:
        'Dein Bestand steht vor ihrer Grundlinie — öffne ihn zuerst mit einer Fassung dazwischen.',
      platz: 'Es ist zu wenig Platz frei.',
      vorgaengeLaufen: 'Es läuft noch etwas.',
      keineSicherung: 'Es ist keine Sicherung gemacht worden.',
      sicherungUnlesbar: 'Die Sicherung ließ sich nicht zurücklesen.',
      halter:
        'Ein anderes Programm hält den Bestand offen — beende es vor dem Wechsel.',
    },
  },
  fassungsband: {
    bereich: 'Neue Fassung',
    neu: 'Fassung {fassung} ist da.',
    seite: 'Was sich geändert hat',
    laden: 'Herunterladen',
    laeuft: 'Wird geladen …',
    fortschritt: '{gelesen} von {ganz} geladen',
    fortschrittOhneLaenge: '{gelesen} geladen',
    geladen:
      'Fassung {fassung} liegt unter {pfad}. Zum Installieren im Finder öffnen und die Anwendung in den Programme-Ordner ziehen.',
    weglegen: 'Diese Meldung schließen',
  },

  // **Das Band nach einem Abbruch** (F04, #2343): die letzte Sitzung endete
  // nicht sauber, und der Weg führt in die Diagnose.
  wiederanlauf: {
    bereich: 'Letzte Sitzung',
    satz: 'Die letzte Sitzung, begonnen am {zeit}, endete unerwartet.',
    satzOhneZeit: 'Die letzte Sitzung endete unerwartet.',
    mitSpur:
      'Eine Absturzspur liegt vor; das Diagnosepaket nimmt sie mit, wenn du es anlegst.',
    ohneSpur:
      'Eine Absturzspur gibt es nicht — wahrscheinlich wurde die Anwendung von außen beendet.',
    ausblenden: 'Ausblenden',
  },

  // Der Arbeitsplatz eines Sets: seine Laschen und sein leerer Zustand (#288).
  arbeitsplatz: {
    // **Die Leiste am Setkopf** — AP-9, Teil 2 (#2366).
    leiste: {
      bezeichnung: 'Kennzahlen des Sets',
      befund: 'Mit Befund',
      nichtGeprueft: 'nicht geprüft',
      nichtGelesen: 'nicht gelesen',
      geprueft: 'geprüft {wann}',
    },
    laschen: 'Bereiche des Sets',

    // **Die Gruppen der Laschenleiste** (#1166). Sie benennen, was jemand
    // gerade vorhat — nicht, wie die Anwendung gebaut ist.
    gruppen: {
      bauen: 'Karten bauen',
      // **„Ausgabe und Druck" ist seit UXR-09 keine Gruppe mehr** (#1917):
      // Ausgeben, Bilder und Druckbogen öffnet der Knopf *Ausgeben* im Kopf.
    },

    // **Die Schubladen** (UX-08, #1680) — Knopf, Titel, Laschenleiste und
    // der Satz darunter, wie im Entwurf. Der Satz sagt, **warum** etwas in
    // der Schublade steht und nicht neben der Arbeit.
    schubladen: {
      eigenschaften: {
        knopf: 'Eigenschaften',
        titel: 'Eigenschaften des Sets',
        laschen: 'Bereiche der Eigenschaften',
        hinweis: 'Was für alle Karten dieses Sets gilt.',
      },
      werkzeuge: {
        knopf: 'Werkzeuge',
        titel: 'Werkzeuge',
        laschen: 'Werkzeuge des Sets',
        hinweis: 'Statistik, Zufallspack und Konsole für dieses Set.',
      },
      schliessen: 'Schließen',
    },

    keinSet: 'Kein Set offen',
    keinSetErklaerung:
      'Wähle unter „Sets“ ein Set, lege eines an oder lies eines ein — dann stehen hier seine Karten, seine Statistik und der Druckbogen.',
    zuDenSets: 'Zu den Sets',
    zuDenPaketen: 'Zu den Paketen',
    zuDenKarten: 'Zu den Karten',
    ohneKartenTitel: 'Noch keine Karte in diesem Set',
    ohneKarten:
      'Diese Lasche arbeitet mit den Karten des Sets. Karten entstehen unter „Karten“ — angelegt, eingefügt oder eingelesen.',
    ohneNamen: 'Offenes Set',
  },
  palette: {
    bezeichnung: 'Befehlspalette',
    eingabe: 'Wohin?',
    keineTreffer: 'Nichts gefunden. Ein anderes Wort probieren?',
    oeffnen: 'Befehlspalette öffnen',
    seiten: 'Seiten',
    befehle: 'Befehle',
    befehlEingabe: 'Wohin — oder welcher Befehl?',

    // **Tun und Finden** — UX-02.AK-02 (#1674), nach dem Artboard *Befehle*.
    tun: 'Tun',
    finden: 'Finden',
    karteErstellen: 'Karte erstellen',
    setWaehlen: 'Ein Set anlegen oder öffnen',
    deckSpielen: 'Schnell spielen mit {name}',
    tauschenMit: 'Karten tauschen mit {name}',
    sicherungErstellen: 'Sicherung erstellen',
    schnittstellenEinrichten:
      'Schnittstellen einrichten (Kommandozeile, MCP, REST)',
    woKarte: 'Gestalten · {set}',
    woSets: 'Gestalten · Sets',
    woPartie: 'Spielen · Partie beginnen',
    woHandel: 'Handeln · Gegenüber',
    woEinstellungen: 'Einstellungen · {gruppe}',
    alsKarte: 'Karte · {set}',
    alsSet: 'Set · {karten}',
    alsDeck: 'Deck · {spiel}',
    alsGegenueber: 'Gegenüber · bestätigt',
    fuss: '↑↓ wählen · ↵ ausführen · esc schließen',
  },
} as const
