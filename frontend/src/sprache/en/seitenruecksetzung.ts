// The **staged reset** — work item O4 of phase 5.2
// (#1428).

export const ruecksetzung = {
  titel: 'Reset',
  hinweis:
    'Six kinds, from the lightest to the heaviest. What each one touches is ' +
    'shown beforehand — with counts and names.',

  artWaehlen: 'What should be reset?',

  arten: {
    layout: 'The window’s appearance only',
    profileinstellungen: 'This profile’s settings',
    nutzereinstellungen: 'My own settings',
    profil: 'This profile and everything it owns',
    nutzer: 'Me and all my profiles',
    anwendung: 'The whole application',
  } as Record<string, string>,

  nachsehen: 'Preview what would be affected',
  nachsehenLaeuft: 'Looking …',

  nichtsBetroffen: 'There is nothing here that could be reset.',
  gruppe: '{was}: {anzahl}',

  gruppen: {
    einstellungen: 'Settings',
    profile: 'Profiles',
    sets: 'Sets',
    karten: 'Cards',
    decks: 'Decks',
    exemplare: 'Copies in the collection',
    token: 'Access tokens',
    freigaben: 'Authorisation records',
    zugaenge: 'Profile access grants',
    bilder: 'Images',
  } as Record<string, string>,

  sicherung:
    'Data will be lost here. Back the database up first — in the area above.',

  hindernisse: {
    offeneRueckkehr:
      'Quantities in this profile are unsettled after a rollback. Check them ' +
      'and sign off the rollback before you delete.',
    fremdesZiel: 'That is not yours.',
    letzterNutzer:
      'This is the last person in this database — deleting them would leave ' +
      'nobody behind.',
    laufenderAuftrag: 'Something is still running. Wait until it has finished.',
    nachweisFehlt:
      'Resetting the whole application requires an explicit authorisation on ' +
      'this machine.',
  } as Record<string, string>,

  nameTippen: 'Type the name to confirm',
  freigabeSatz:
    'Resetting the whole application needs your authorisation on this computer. It lasts five minutes and covers one fresh start.',
  freigeben: 'Grant authorisation',
  freigebenLaeuft: 'Granting authorisation …',
  freigegeben: 'Authorised until {zeit}.',
  kennungFuerKommandozeile:
    'The same authorisation works on the command line (mse zuruecksetzen anwendung):',

  tun: 'Reset',
  tunLaeuft: 'Resetting …',

  bilanz: '{anzahl} settings taken back.',
  beiseite: 'The previous database is at {ort}.',
}
