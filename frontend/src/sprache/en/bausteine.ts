import { dateiansicht } from './bausteinedatei'
import { ablage } from './bausteineablage'
import { geste } from './bausteinegeste'
import { laufstand } from './bausteinelauf'
import { objektmarken } from './bausteinemarken'
import { teilen } from './bausteineteilen'

export const bausteine = {
  dateiansicht,
  ablage,
  geste,
  laufstand,
  objektmarken,
  teilen,
  entwurf: {
    fremdGeaendert:
      'Meanwhile this field reads »{wert}« — changed elsewhere. Your draft is not saved yet.',
    behalten: 'Save my draft',
    nehmen: 'Take the new value',
  },
  listenansicht: {
    suche: 'Search',
    suchePlatzhalter: 'Name or game',
    sortierung: 'Order',
    ansicht: 'View',
    liste: 'List',
    kacheln: 'Tiles',
    zahl: '{gezeigt} of {gesamt}',
    nichtsPasst: 'Nothing matches “{suche}”.',
    sucheLeeren: 'Clear search',
    filterZuruecksetzen: 'Reset search and filters',
    weitere: 'Show {anzahl} more',
    wieImBestand: 'As stored',
    zuletztBenutzt: 'Recently used first',
    nachName: 'By name',
    nachKarten: 'By card count',
  },
  speicherstand: {
    schreibt: 'Saving …',
    geschrieben: 'Every change is in the store.',
    fehler: '“{feld}” is not saved — the reason is at the field.',
  },
  gegenueberTitel: 'Add or confirm a counterpart',
  kontaktwahlEintrag: '{name} · {fingerabdruck}',
  kontaktwahlSuchen: 'Search counterparts …',
  feldhilfe: 'Explanation for {feld}',
  mehrZu: 'More on “{satz}”',

  // **Where something came from** — Werkstück 12 of Phase 4.12
  // (#1351).
  herkunftAusProfil: 'Taken over from the profile {quelle}.',
  herkunftAusEinfuhr: 'Imported from {quelle}.',

  pflichtfeldLeer: '{feld} needs a value.',

  waehrung: {
    beispiel: 'e.g. EUR',
    keineGestalt:
      '“{wert}” is not a currency code — one has three letters, like EUR.',
    grossSchreiben:
      '“{wert}” — a currency code is written in capitals: {gross}.',
  },

  drucksprachen: {
    beispiel: 'e.g. en',
    namen: {
      en: 'English',
      es: 'Spanish',
      fr: 'French',
      de: 'German',
      it: 'Italian',
      pt: 'Portuguese',
      ja: 'Japanese',
      ko: 'Korean',
      ru: 'Russian',
      zhs: 'Simplified Chinese',
      zht: 'Traditional Chinese',
      he: 'Hebrew',
      la: 'Latin',
      grc: 'Ancient Greek',
      ar: 'Arabic',
      sa: 'Sanskrit',
      ph: 'Phyrexian',
      qya: 'Quenya',
      dw: 'Dwarvish',
      tlh: 'Klingon',
    },
  },

  // **The password field** — piece 6 of phase 4.12 (#1351).
  kennwort: {
    zeigen: 'Show the password in plain text',
    zeigenKurz: 'Show',
    verbergen: 'Hide the password again',
    verbergenKurz: 'Hide',
  },

  // **The field for the second factor** — piece 7 of phase 4.12 (#1351).
  nachweis: {
    beschriftung: 'Code',
    abschicken: 'Confirm',
  },

  aktionen: {
    abbrechen: 'Cancel',
    oeffnen: 'Open',
    schliessen: 'Close',
    speichern: 'Save',
  },
  stilliste: {
    mitZusatz: '{kennung} · {zusatz}',
    keinerGewaehlt: '— Choose a style —',
  },
  druckwahl: {
    drucke: 'Printings',
    hinweis: 'The peer sees only these printings, with the quantity you own.',
    suchen: 'Search printings',
    gewaehlt: '{anzahl} selected',
    nurSuche: 'Your collection has {anzahl}. Search for a name to select some.',
    keiner: 'No printing matches “{suche}”.',
    genauer: 'More printings match than are shown. Narrow the search.',
    leer: 'Your collection has no printing you could share yet. Add copies first.',
    fehlt:
      'Select at least one printing — a selection without printings shares nothing.',
    behalten: 'Copies of each printing I keep',
    behaltenHinweis:
      'The peer sees what lies above that: with 1 and three copies, two.',
    detailBehalten: 'you keep {behalten}',
  },
  suchauswahl: {
    keiner: 'Nothing matches “{suche}”.',
    treffer: 'Results: {anzahl}',
    begrenzt: 'Showing {gezeigt} of {anzahl} results — narrow the search …',
  },
  zeitzonenwahl: {
    system: 'Same as the system',
  },
  aktionskatalog: {
    aktionen: {
      anlegen: { kurz: 'Create', heisst: 'Creates something new.' },
      bearbeiten: { kurz: 'Edit', heisst: 'Changes it in place.' },
      sichern: { kurz: 'Save', heisst: 'Keeps the current state.' },
      zuruecknehmen: { kurz: 'Undo', heisst: 'Takes back the last step.' },
      duplizieren: {
        kurz: 'Duplicate',
        heisst: 'Puts a second one next to it.',
      },
      kopieren: { kurz: 'Copy', heisst: 'Puts it on the clipboard.' },
      entfernen: {
        kurz: 'Remove',
        heisst: 'Takes it out of this view; it stays elsewhere.',
      },
      loeschen: { kurz: 'Delete', heisst: 'Deletes it for good, everywhere.' },
    },
  },
  nurlesen: {
    hinweis:
      'You are viewing this profile as a reader: you can look at everything and change nothing. Switch to your own profile to work \u2014 or ask whoever manages this one for the \u201ccollaborator\u201d role.',
    knopf: 'Read only — you cannot change anything in this profile.',
  },

  // **The context strip** — piece U4 of phase 5.1 (#1415).
  kontext: {
    bezeichnung: 'Working context',
    profilVon: 'owned by {eigner}',
    rollen: {
      leser: 'Reader',
      mitarbeiter: 'Collaborator',
      verwalter: 'Manager',
    },
    nurLesen: 'read only',
    zuDenNutzern: 'Switch profile …',
    zeile: '{name} · {profil}',
  },

  zustaende: {
    laedt: 'Loading …',
    leer: 'Nothing here',
    fehler: 'This could not be loaded.',
    fehlerMitGrund: 'This could not be loaded. {grund}',
    wiederholen: 'Load again',
    neuLaden: 'Reload',
  },
  // Find and replace **inside the field** (#1164).
  suche: {
    oeffnen: 'Find in text',
    schliessen: 'Close find',
    suchen: 'Find',
    ersetzen: 'Replace with',
    treffer: '{bei} of {ganz}',
    keinTreffer: 'No match',
    weiter: 'Next match',
    zurueck: 'Previous match',
    einsetzen: 'Replace',
    alleEinsetzen: 'Replace all',
    grossKlein:
      'Case matters — in MSE-Script »card« and »Card« are two different things.',
  },
  // Completion **inside** the field (#1164).
  // **Welche Karten hinausgehen** (#1167).
  ausgabeauswahl: {
    aufschrift: 'Which cards',
    ganzesSet: 'The whole set ({zahl} cards)',
    ganzesSetOhneZahl: 'The whole set',
    teilmenge: 'Only selected ({zahl})',
    ohneNamen: 'unnamed',
    eingrenzen: 'Filter cards',
  },

  vorschlaege: {
    bezeichnung: 'Suggestions',
    uebernehmen: 'Accept with ↹',
  },

  // **The regions of a workspace** — piece U2 of phase 5.1 (#1413).
  arbeitsplatz: {
    breiteVon: 'Width of “{region}”',
    breiteWert: '{punkte} points',
    navigatorAuf: 'Show the navigator',
    navigatorZu: 'Hide the navigator',
    inspectorAuf: 'Show the inspector',
    inspectorZu: 'Hide the inspector',
    navigatorEng:
      'Too narrow for all three columns — the navigator opens over the centre',
    inspectorEng:
      'Too narrow for all three columns — the inspector opens over the centre',
  },
  // The command centre (#1423).
  druck: {
    vorlaeufig: 'provisional',
    vorlaeufigErklaerung:
      'This printing has no Scryfall identifier yet. Prices and images stay away until it is resolved.',
  },
  kommandozentrale: {
    titel: 'Command centre',
    erklaerung:
      'A command is a tool of the MCP server, called here in the window — the same answer as over MCP and on the command line. Nothing passes through a shell: what you type stays data.',
    lesend: 'reads',
    schreibend: 'changes',
    gefaehrlich: 'not easily undone',
    faehigkeit: 'Capability',
    cli: 'Command line',
    rest: 'Service',
    parameter: 'Parameters',
    keineParameter: 'This command takes no parameters.',
    pflicht: 'Required',
    jsonHinweis:
      'As JSON — a list in square brackets, an object in curly braces.',
    befundPflicht: 'This field needs a value.',
    befundZahl: 'That is not a number.',
    befundJson: 'That is not JSON.',
    vorschau: 'What will be sent',
    bestaetigen: 'I have read the preview and want to run this as is.',
    ausfuehren: 'Run',
    abbrechen: 'Cancel',
    laeuft: 'The command is running …',
    ergebnis: 'Result',
    ausgabe: 'Command output',
    alsText: 'Text',
    alsJson: 'JSON',
    dauer: '{ms} ms',
    verlaufLeer:
      'Nothing run yet. Secrets — tokens, passwords — never appear here in clear text.',
    verlauf: 'History of this session',
    wiederholen: 'Load {name} again',
    gelungen: 'succeeded',
    gescheitert: 'failed',
    abgebrochen: 'cancelled',
    schliessen: 'Close',
  },
  // „Diagnose öffnen“ an einem Fehler mit Vorgangskennung (NET-02,
  // S13.NAT-04, #1706).
  diagnoseverweis: {
    oeffnen: 'Open diagnostics',
  },
  // „KI einrichten“ an einer KI-Meldung (AI-01.AK-04, #1742).
  kieinrichtung: {
    oeffnen: 'Set up AI',
  },
  kikasten: {
    titel: 'AI for this card',
  },
} as const
