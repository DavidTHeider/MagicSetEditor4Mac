// **Die neun Gruppen der Einstellungsseite** — UX-04.AK-01
// (#1676).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge. Ein Katalog darf sich auf mehrere Dateien
// verteilen, solange er als **einer** ankommt (§9).
export const einstellungsgruppennamen = {
  gruppen: {
    allgemein: 'Generali',
    bereiche: 'Aree e funzioni',
    erscheinungsbild: 'Aspetto',
    nutzer: 'Utente e profilo',
    fach: 'Funzioni specifiche',
    ki: 'IA',
    verbindungen: 'Connessioni',
    dienste: 'Servizi',
    daten: 'Dati e backup',
    diagnose: 'Diagnostica',
  },

  untertitel: {
    allgemein: 'Lingua, avvio, scorciatoie',
    bereiche: 'Aree di lavoro, funzioni, maturità',
    erscheinungsbild: 'Tema, caratteri, finestra',
    nutzer: 'Accesso, sblocco',
    fach: 'Carte, collezione, gioco',
    ki: 'Provider, modelli, costi',
    verbindungen: 'Rete, controparti, CLI, MCP, REST',
    dienste: 'Servizi di carte, prezzi, mercato',
    daten: 'Archivio, percorsi, backup',
    diagnose: 'Log, misurazione, tecnica',
  },

  abschnitte: {
    fenster: 'Finestra e barra laterale',
    spalten: 'Colonne e viste',
    handel: 'Limiti degli scambi',
    protokoll: 'Registro',
    messung: 'Misurazione',
    speicher: 'Cache e indice di ricerca',
    adressen: 'Indirizzi',
    ordner: 'Cartelle e nomi di file',
    eingriffe: 'Tornare indietro, pulire, ripristinare',
    mitnehmen: 'Portare con sé le impostazioni',
  },
} as const
