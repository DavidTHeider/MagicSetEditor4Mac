// Die Texte der **abgestuften Rücksetzung** — Werkstück O4 von Phase 5.2
// (#1428).

export const ruecksetzung = {
  titel: 'Zurücksetzen',
  hinweis:
    'Sechs Arten, von der leichtesten zur schwersten. Was jede anfasst, ' +
    'steht vorher da — mit Zahlen und Namen.',

  artWaehlen: 'Was soll zurückgesetzt werden?',

  arten: {
    layout: 'Nur das Aussehen des Fensters',
    profileinstellungen: 'Die Einstellungen dieses Profils',
    nutzereinstellungen: 'Meine eigenen Einstellungen',
    profil: 'Dieses Profil samt allem, was ihm gehört',
    nutzer: 'Mich samt allen meinen Profilen',
    anwendung: 'Die ganze Anwendung',
  } as Record<string, string>,

  nachsehen: 'Nachsehen, was betroffen wäre',
  nachsehenLaeuft: 'Sieht nach …',

  nichtsBetroffen: 'Hier ist nichts, was zurückgesetzt werden könnte.',
  gruppe: '{was}: {anzahl}',

  gruppen: {
    einstellungen: 'Einstellungen',
    profile: 'Profile',
    sets: 'Sets',
    karten: 'Karten',
    decks: 'Decks',
    exemplare: 'Exemplare in der Sammlung',
    token: 'Zugriffstoken',
    freigaben: 'Freigabenachweise',
    zugaenge: 'Zugänge zu Profilen',
    bilder: 'Bilder',
  } as Record<string, string>,

  sicherung:
    'Hier gehen Daten fort. Sichere den Bestand vorher — im Gebiet darüber.',

  hindernisse: {
    offeneRueckkehr:
      'Über Mengen in diesem Profil steht nach einer Rückkehr nichts fest. ' +
      'Sieh sie nach und quittiere die Rückkehr, bevor du löschst.',
    fremdesZiel: 'Das gehört dir nicht.',
    letzterNutzer:
      'Das ist der letzte Mensch in diesem Bestand — ihn zu löschen ließe ' +
      'niemanden übrig.',
    laufenderAuftrag: 'Hier läuft noch etwas. Warte, bis es fertig ist.',
    nachweisFehlt:
      'Die ganze Anwendung zurückzusetzen verlangt eine ausdrückliche ' +
      'Freigabe an diesem Rechner.',
  } as Record<string, string>,

  nameTippen: 'Zur Bestätigung den Namen abtippen',
  freigabeSatz:
    'Die ganze Anwendung zurückzusetzen braucht deine Freigabe an diesem Rechner. Sie gilt fünf Minuten und für einen Neuanfang.',
  freigeben: 'Freigabe erteilen',
  freigebenLaeuft: 'Erteilt die Freigabe …',
  freigegeben: 'Freigegeben bis {zeit}.',
  kennungFuerKommandozeile:
    'An der Kommandozeile gilt dieselbe Freigabe (mse zuruecksetzen anwendung):',

  tun: 'Zurücksetzen',
  tunLaeuft: 'Setzt zurück …',

  bilanz: '{anzahl} Einstellungen zurückgenommen.',
  beiseite: 'Der bisherige Bestand liegt bei {ort}.',
}
