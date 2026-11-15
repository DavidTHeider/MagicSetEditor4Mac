// The texts of the **package inspector and the plan before a package
// change** — DOM-01 (#1696). Their German source is `de/seitenpaketplan.ts`.

export const paketplan = {
  arbeitsplatz: 'Package workspace',
  navigator: 'Search and filters',
  inspector: 'Package',
  zentrum: 'Packages',

  herkunft: 'Origin',
  herkunftAlle: 'Any origin',
  herkunftsarten: {
    mitgeliefert: 'bundled',
    bezogen: 'fetched',
    uebernommen: 'imported',
    eigen: 'own',
  },

  spalteStand: 'State',
  standEingebunden: 'installed',
  standGeaendert: 'changed locally',

  inspectorLeer:
    'Choose a package in the list — its origin, version, dependencies and what a change would do appear here.',
  kopf: '{art} · version {version}',
  kopfOhneFassung: '{art} · no version',
  eingebunden: 'Installed',
  keinSet: 'No set of this profile uses it.',
  einSet: 'One set uses it.',
  vieleSets: '{anzahl} sets use it.',
  setsUnbekannt: 'Which sets need it is not recorded in any set header.',
  geaendert:
    'Changed locally since it was installed. Replacing it asks first; the change is kept as the previous version.',
  woher: 'Origin',
  woherMitgeliefert: 'Bundled with the application.',
  woherBezogen: 'Fetched from {herkunft}.',
  woherUebernommen: 'Imported from {herkunft}.',
  woherEigen: 'Created here or placed in the package folder by hand.',
  pruefsumme: 'Checksum',
  lizenz: 'Licence',
  lizenzUnbekannt:
    'not stated — an MSE package carries no licence field; the content belongs to its rights holders',
  ort: 'Location',
  haengtAb: 'Depends on',
  gebrauchtVon: 'Needed by',
  nichts: 'nothing',

  einbinden: 'Install package …',
  beziehen: 'Fetch a package …',
  beziehenTitel: 'Fetch a package',
  loeschen: 'Delete …',
  zurueck: 'Restore previous version …',
  weitergeben: 'Share as installer …',
  weitergegeben: 'Saved as installer: {datei}',

  titel: {
    loeschen: 'Delete {paket}',
    einbinden: 'Install {paket}',
    ersetzen: 'Replace {paket}',
    zurueck: 'Reset {paket} to its previous version',
  },
  fassung: 'Version {von} → {zu}',
  ohneFassung: 'no version',
  sets: 'Affected sets',
  keineSets: 'No set of this profile is affected.',
  konflikte: 'No longer satisfied afterwards',
  keineKonflikte: 'No dependency breaks.',
  planGeaendert: 'The package has been changed locally since it was installed.',
  ueberschreiben:
    'Replace the local change — it is kept as the previous version',
  nichtMoeglich: {
    loeschen:
      'This package is not in the managed folder. Only what this application installed is deleted here.',
    einbinden: 'This package cannot be installed here.',
    ersetzen: 'This package cannot be replaced here.',
    zurueck:
      'There is no previous version. One is kept when you install a newer version over this package.',
  },
  bestaetigen: {
    loeschen: 'Delete',
    einbinden: 'Install',
    ersetzen: 'Replace',
    zurueck: 'Reset',
  },
  verwerfen: 'Change nothing',
  bedarf: {
    ueberschrift: 'Packages this set is missing',
    erklaerung:
      'Without them the preview draws a fallback or leaves out what the package contributes. Install the package from a folder — or search for it among the packages to fetch it or import it from MSE2.',
    suchen: 'Search among packages',
  },
  quelle: 'From {quelle}',
  keineRechte:
    'A package brings templates and MSE-Script. The script has no access to files, the network or programs, and nothing the package needs is fetched along — it is listed above as a conflict.',
}
