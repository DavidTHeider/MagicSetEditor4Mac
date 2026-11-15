// **Die Startseite** — herausgelöst, weil `seiten.ts` mit ihr über 400 Zeilen
// ginge (§2).
//
// **Ein Katalog darf sich auf mehrere Dateien verteilen**, solange er als
// **einer** ankommt — so wie `seitenkarten.ts`. Was gleich bleibt, ist die
// Zahl der Kataloge; an ihr hängt der Vertragstest gegen die Go-Seite (§9).
export const start = {
  ueberblick: {
    titel: 'Your overview',
    mitProblemen: '{anzahl} with problems',
    geprueftVon: '{geprueft} of {gesamt} checked',
    mitBefund: '{anzahl} with findings',
    nichtGeprueft: 'not fully checked',
    problemeGeprueft: 'Problems checked {wann} · {geprueft} of {gesamt} cards',
    problemeOffen: 'Problems: {offen} of {gesamt} sets not checked',
    pruefen: 'Check now',
    setsMitProblemen: '{sets} with problems',
    kartenMitBefund: '{karten} with findings in {sets}',
    aeltesterPreis: 'oldest price {wann}',
    ohnePreis: '{anzahl} without a price',
    mitFehlenden: '{anzahl} with missing cards',
    decksMitFehlenden: '{decks} with missing cards',
    alter: {
      gerade: 'just now',
      minuten: '{minuten} min ago',
      stunden: '{stunden} h ago',
      tage: '{tage} days ago',
      tag: '1 day ago',
    },
  },
  // Resume — work item U11 of phase 5.1 (#1422).
  // Attention and background tasks (#1422).
  aufmerksamkeit: {
    titel: 'Needs attention',
    nichts: 'Nothing — {sets}, {decks} and {pakete} checked.',
    durchgesehen: '{sets}, {decks} and {pakete} checked.',
    arten: {
      setVorbehalte: '{anzahl} not interpreted',
      deckFehlend: '{anzahl} missing from the collection',
      paketBefund: '{anzahl} in scripts',
    },
    paketeMitBefund: '{pakete} with script findings',
    gruendeZeigen: 'Show reasons',
    gruendeVerbergen: 'Hide reasons',
    setBefunde: '{karten} with findings',
    befunde: 'Not checked:',
    alle: 'Show all {anzahl}',
    weniger: 'Show fewer',
    gruende: {
      vorbehalt: '{0}, line {1}',
    },
    amOrt: '{ort}: {satz}',
    // The operations band (OPS02, #2348).
    betrieb: {
      arten: {
        platzKnapp: 'Only {frei} free next to the library',
        sicherungAlt: 'Last backup {tage} days ago',
        dienstBeendet: 'The login service ended with code {ausgang}',
        webprozessNeu: 'The web view was reloaded {anzahl} times',
      },
      rat: {
        platzKnapp:
          'Below about 2 GB or 5 % free, every backup gets tight — clean up in “Data and backup”.',
        sicherungAlt:
          'Older than 14 days — back up the library in “Data and backup”.',
        dienstBeendet:
          'After an error, launchd restarts it — the diagnostics say why it ended.',
        webprozessNeu:
          'WebKit had ended its process; anything unsaved in an open form may be missing — check the diagnostics.',
      },
    },
  },
  auftraege: {
    titel: 'Background tasks',
    staende: {
      laeuft: 'running',
      fertig: 'done',
      gescheitert: 'failed',
      abgebrochen: 'cancelled',
    },
    arten: {
      preise: 'Fetch prices for the collection',
      sammeldaten: 'Load bulk data from Scryfall',
      fassung: 'Fetch a new version',
      bogen: 'Lay out a print sheet',
      set: 'Read a set',
      mtgjson: 'Read MTG JSON',
      vorlage: 'Export through a template',
      installer: 'Read an installer',
      weitergabe: 'Pass on as an installer',
      probleme: 'Check sets for problems',
    },
    abbrechen: 'Cancel {art}',
    dauer: '{sekunden} s',
    laufend: { eine: 'task running', viele: 'tasks running' },
  },
  weiterarbeiten: {
    titel: 'Resume',
    weiterAn: 'Continue with {art} “{name}”',
    zuletztWann: 'Last {vorgang} on {zeit}',
    angeheftet: 'pinned',
    angeheftetLeer: 'Nothing pinned — pin a row to keep it up here.',
    leer: 'This is where your recent work shows up. It is empty so far — the first step is a set or a template package.',
    handlungSets: 'Create or open a set',
    handlungPakete: 'Fetch a package',
    anheften: 'Pin',
    loesen: 'Unpin',
    anheftenVon: 'Pin “{name}”',
    loesenVon: 'Unpin “{name}”',
    arten: { set: 'Set', deck: 'Deck', paket: 'Package' },
    vorgaenge: {
      geoeffnet: 'opened',
      eingefuehrt: 'imported',
      ausgegeben: 'exported',
    },
  },
  titel: 'Home',
  ueberschrift: 'MagicSetEditor4Mac',
  was: 'Design, review and print cards for Magic and other games — using the templates Magic Set Editor 2 uses.',
  ersteSchritte: 'First steps',
  einfuehrung: {
    willkommen:
      'Three steps and the first card is done. You can bring this window back any time from the settings.',
    wiederzeigen: 'Show again on start',
    schliessen: 'Let’s go',
    bezeichnung: 'Introduction',
  },
  schritte: {
    setOeffnen: {
      titel: 'Open a set',
      erklaerung:
        'Open an existing .mse-set file — or import a set from MTG JSON.',
      handlung: 'Go to sets',
    },
    kartenBearbeiten: {
      titel: 'Edit cards',
      erklaerung:
        'View and change the cards of the open set; the preview shows every change at once.',
      handlung: 'Go to cards',
      ersteinSet: 'This needs an open set first.',
      stattdessen: 'Open a set first',
    },
    paketeHolen: {
      titel: 'Get templates',
      erklaerung:
        'Eight template packages are already here. Want more games or styles? Fetch them here.',
      handlung: 'Go to packages',
    },
  },
}
