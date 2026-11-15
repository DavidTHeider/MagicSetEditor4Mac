// The texts of the **fetched states** on the packages page — G04 (#2251).
// Their German source is `de/seitenbezuege.ts`.

export const bezuege = {
  ueberschrift: 'Fetched states',
  erklaerung:
    'A fetched package stays at one commit and does not age by itself. Ask the sources whether there is a newer one — only then does anything go over the network.',
  fragen: 'Ask the sources',
  leer: 'No package fetched yet. Fetch one with “Fetch package …” — afterwards this shows how far its source has moved.',
  bezogen: 'Fetched: {stand}',
  geprueft: 'Checked: {stand}',
  ohnePruefung: 'Not in the list of checked sources',
  pakete: '{zahl} packages',
  paketEins: 'one package',
  keinPaket: 'no packages',
  lage: {
    ungefragt: 'Not asked yet.',
    aktuell: 'At the newest state of the source.',
    neuerEins: 'Newer state {spitze} from {datum} — one commit ahead.',
    neuer: 'Newer state {spitze} from {datum} — {vorsprung} commits ahead.',
    abweichend: 'The source is not ahead — reset or diverged (head {spitze}).',
    unbekannt: 'The source could not be asked: {grund}',
  },
  ansehenBezeichnung: 'Review update … ({name})',
  zurueckBezeichnung: 'Back to the kept state ({name}, {stand})',
  ansehen: 'Review update …',
  aufgehoben: 'Kept: {stand}',
  zurueck: 'Back to the kept state',
  vorschauTitel: 'Update of {name}',
  vonZu: '{von} → {zu}: {vorsprung} ahead, {rueckstand} behind',
  paketzeile: '{paket}: {neu} new, {geaendert} changed, {entfernt} removed',
  keinePakete: 'No template package changes.',
  uebrige: '{zahl} more files belong to no package.',
  sets: 'These sets name an affected package: {sets}',
  keineSets: 'No set names an affected package.',
  aufheben: 'The current state is kept and can be brought back.',
  vorigeWeicht: 'The already kept state {stand} is replaced.',
  gruende: {
    setVerliert: 'A package that a set names loses files.',
    nichtVoraus:
      'The source is not ahead — the update would be a step back or sideways.',
    gekuerzt:
      'The source lists more than 300 changed files; the list above is cut short.',
  },
  bestaetigen: 'I have seen the reasons and still want to update.',
  ausfuehren: 'Update to {zu}',
  verwerfen: 'Do not update',
}
