// The mapping of ⌘0 to ⌘9 — UX-02.AK-03 (#1674), split out like the German file.
export const einstellungsbelegung = {
  belegungUeberschrift: 'Mapping of ⌘0 to ⌘9',
  belegungHinweis:
    'Which page a digit opens with ⌘. The default is the areas of the sidebar; “As before 0.6.0” brings back the mapping up to 0.5.0, whose table is in the change log.',
  belegungFuer: 'Page for {kuerzel}',
  belegungFrei: '—',
  belegungVor06: 'As before 0.6.0',
  belegungVorgabe: 'Back to the default',
  kuerzel: {
    palette: 'Open the command palette',
    suche: 'Search the set',
    zurueck: 'Undo a change',
    vor: 'Redo a change',
    sichern: 'Save the set',
    sichernUnter: 'Save the set as …',
    oeffnen: 'Open a set …',
    neu: 'Create a new set',
    schliessen: 'Close the set',
    drucken: 'Go to the print sheet',
    navigator: 'Show or hide the navigator',
    inspector: 'Show or hide the inspector',

    // Back and forward between pages — UX-03.AK-02 (#1675).
    seiteZurueck: 'Previous page',
    seiteVor: 'Next page',
  },
} as const
