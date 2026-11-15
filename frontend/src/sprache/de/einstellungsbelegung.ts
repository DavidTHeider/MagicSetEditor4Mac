// **Die Belegung von ⌘0 bis ⌘9** — UX-02.AK-03
// (#1674).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihr über die
// Zeilengrenze aus §2 ging; sie wird dort eingesetzt, und der Katalog kommt
// als einer an (AGENTS.md §9).
export const einstellungsbelegung = {
  belegungUeberschrift: 'Belegung von ⌘0 bis ⌘9',
  belegungHinweis:
    'Welche Seite eine Ziffer mit ⌘ öffnet. Die Vorgabe sind die Bereiche der Seitenleiste; „Wie vor 0.6.0“ holt die Belegung bis 0.5.0 zurück, deren Tabelle im Änderungsprotokoll steht.',
  belegungFuer: 'Seite für {kuerzel}',
  belegungFrei: '—',
  belegungVor06: 'Wie vor 0.6.0',
  belegungVorgabe: 'Zurück zur Vorgabe',
  kuerzel: {
    palette: 'Befehlspalette öffnen',
    suche: 'Im Set suchen',
    zurueck: 'Änderung zurücknehmen',
    vor: 'Änderung wiederherstellen',
    sichern: 'Set speichern',
    sichernUnter: 'Set speichern unter …',
    oeffnen: 'Set öffnen …',
    neu: 'Neues Set anlegen',
    schliessen: 'Set schließen',
    drucken: 'Zum Druckbogen',

    // **Die Regionen eines Arbeitsplatzes** — Werkstück U2 von Phase 5.1
    // (#1413): ⌥⌘[ und ⌥⌘], gedeutet über die Taste, nicht das Zeichen.
    navigator: 'Navigator ein- oder ausblenden',
    inspector: 'Inspector ein- oder ausblenden',

    // **Zurück und Vorwärts zwischen den Seiten** — UX-03.AK-02 (#1675):
    // ⌘[ und ⌘], die Kürzel des Finders für den vorigen und nächsten Ordner.
    seiteZurueck: 'Zur vorigen Seite',
    seiteVor: 'Zur nächsten Seite',
  },
} as const
