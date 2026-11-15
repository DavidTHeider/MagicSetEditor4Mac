// **Die neun Gruppen der Einstellungsseite** — UX-04.AK-01
// (#1676).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge. Ein Katalog darf sich auf mehrere Dateien
// verteilen, solange er als **einer** ankommt (§9).
export const einstellungsgruppennamen = {
  gruppen: {
    allgemein: 'General',
    bereiche: 'Areas and functions',
    erscheinungsbild: 'Appearance',
    nutzer: 'User and profile',
    fach: 'Domain features',
    ki: 'AI',
    verbindungen: 'Connections',
    dienste: 'Services',
    daten: 'Data and backup',
    diagnose: 'Diagnostics',
  },

  untertitel: {
    allgemein: 'Language, start, shortcuts',
    bereiche: 'Workspaces, functions, maturity',
    erscheinungsbild: 'Theme, type, window',
    nutzer: 'Sign-in, unlocking',
    fach: 'Cards, collection, play',
    ki: 'Providers, models, cost',
    verbindungen: 'Network, counterparts, CLI, MCP, REST',
    dienste: 'Card services, prices, market',
    daten: 'Store, paths, backup',
    diagnose: 'Logs, measurement, technical',
  },

  abschnitte: {
    fenster: 'Window and sidebar',
    spalten: 'Columns and views',
    handel: 'Trading limits',
    protokoll: 'Log',
    messung: 'Measurement',
    speicher: 'Caches and search index',
    adressen: 'Addresses',
    ordner: 'Folders and file names',
    eingriffe: 'Go back, tidy up, reset',
    mitnehmen: 'Take settings with you',
  },
} as const
