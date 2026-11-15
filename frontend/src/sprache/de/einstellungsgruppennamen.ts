// **Die neun Gruppen der Einstellungsseite** — UX-04.AK-01
// (#1676).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge. Ein Katalog darf sich auf mehrere Dateien
// verteilen, solange er als **einer** ankommt (§9).
export const einstellungsgruppennamen = {
  gruppen: {
    allgemein: 'Allgemein',
    bereiche: 'Bereiche und Funktionen',
    erscheinungsbild: 'Erscheinungsbild',
    nutzer: 'Nutzer und Profil',
    fach: 'Fachfunktionen',
    ki: 'KI',
    verbindungen: 'Verbindungen',
    dienste: 'Dienste',
    daten: 'Daten und Sicherung',
    diagnose: 'Diagnose',
  },

  // **Die Abschnitte innerhalb einer Gruppe** — UXR-19
  // (#1927).
  // Eigene Übersetzungen, Tastenkürzel und die Werkzeuge ohne Fenster
  // tragen ihren Namen schon an ihrem Bauteil; hier steht nur, was neu
  // einen braucht.
  abschnitte: {
    fenster: 'Fenster und Leiste',
    spalten: 'Spalten und Ansichten',
    handel: 'Schranken für den Handel',
    protokoll: 'Protokoll',
    messung: 'Messung',
    speicher: 'Zwischenspeicher und Suchindex',
    adressen: 'Adressen',
    ordner: 'Ordner und Dateinamen',
    eingriffe: 'Zurückgehen, aufräumen, zurücksetzen',
    mitnehmen: 'Einstellungen mitnehmen',
  },
} as const
