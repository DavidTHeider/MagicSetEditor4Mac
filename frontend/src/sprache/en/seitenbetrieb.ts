// **Run times, consumption and the performance profile** — OPS-05
// (#1722).

export const betrieb = {
  ueberschrift: 'Run times and consumption',
  zeitraum: 'Collected from {seit} to {bis}.',
  aufwand: 'Cost of collecting: {mikros} µs',
  speicher: 'Memory',
  speicherWert:
    '{fussabdruck} footprint — of which {heap} Go heap, {system} from the system',
  speicherOhneFussabdruck:
    '{heap} Go heap, {system} from the system — the footprint is not measured',
  web: 'Web processes',
  webWert: '{summe} in total — {teile}',
  webTeil: '{art} {groesse}',
  webArten: { webinhalt: 'Web content', gpu: 'Graphics', netz: 'Network' },
  ohneWeb: 'This process holds no web view.',
  webUngemessen: 'Not measured: {grund}',
  goroutinen: 'Concurrent routines',
  rechenzeit: 'CPU time since start',
  datenbank: 'Connections to the store',
  datenbankWert:
    '{offen} open, {inBenutzung} in use, waited {gewartet} times, {dauer} in total',
  ohneBestand: 'This process holds no store.',
  zwischenspeicher: 'Caches of outside services',
  zwischenspeicherWert:
    '{treffer} hits, {geholt} fetched, {veraltet} served stale, {fehlgeschlagen} failed',
  laufzeiten: 'Run times per surface',
  laufzeitenLeer:
    'No calls are measured here — only the service and the tool server are measured.',
  ungemessen: 'Not measured: {liste}.',
  ungemessenNamen: {
    fenster:
      'the window bindings (they have no common place where a call ends)',
    kommandozeile: 'the command line (it lives for one command)',
    speicher: 'the footprint of this process (the system did not provide it)',
  },
  spalteArt: 'Kind',
  spalteAnzahl: 'Calls',
  spalteFehler: 'Error rate',
  spalteMittel: 'Mean',
  spalteVerteilung: 'Distribution',
  arten: { lesend: 'reading', schreibend: 'writing' },
  eimer: 'up to {grenze} ms',
  verteilung: '{eimer}: {anzahl} calls',
  eimerOffen: 'above',
  spurAus:
    'The trace is off. Switch it on under Settings › Diagnostics — it applies from the next start.',
  spurAn:
    'Trace on: {aufgenommen} of {gesehen} calls, {abtastung} % sampled. At most {grenze} entries for {minuten} min — {verdraengt} displaced, {abgelaufen} expired.',
}

export const leistungsprofil = {
  ueberschrift: 'Performance profile',
  erklaerung:
    'A CPU profile of this window over the chosen time, in the format of »go tool pprof«. It is kept next to the store; how many stay is set under Settings › Diagnostics.',
  sekunden: 'Duration in seconds',
  grenze: 'A whole number from 1 to 60.',
  aufzeichnen: 'Record profile',
  laeuft: 'Recording — the window stays usable.',
  fertig: 'Written: {datei} ({groesse})',
}

export const zwischenspeicher = {
  ueberschrift: 'Caches',
  erklaerung:
    'What this application keeps so it does not fetch or compute it twice — on disk and in memory. None of it is your data: once emptied, every cache fills again on next use.',
  spalteSpeicher: 'Cache',
  spalteOrt: 'Place',
  spalteQuote: 'Hit rate',
  spalteEintraege: 'Entries',
  spalteGroesse: 'Size',
  spalteGrenze: 'Limit',
  keineQuote: 'no lookup yet',
  keineGrenze: 'one file that gets replaced',
  namen: {
    'scryfall.antworten': 'Scryfall answers',
    'scryfall.bilder': 'Card images from Scryfall',
    'scryfall.sammeldaten': 'Scryfall bulk file',
    'render.ebenen': 'Sampled layers',
    'pakete.bilder': 'Template package images',
    'skript.baeume': 'Parsed scripts',
    'vorschau.bilder': 'Card previews',
    'tisch.texturen': 'Card images at the table',
  },
  orte: { platte: 'Disk', arbeitsspeicher: 'Memory' },
  leerenHinweis:
    'Emptying throws everything away. Without a network, Scryfall answers and images only come back once it returns.',
  leeren: 'Empty all',
  geleert: 'Emptied — they fill again on next use.',
}
