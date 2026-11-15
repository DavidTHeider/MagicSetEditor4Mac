// **Die Texte des Dienstkatalogs** — die englische Seite zu
// [`de/einstellungenkatalog.ts`](../de/einstellungenkatalog.ts).
export const dienstkatalog = {
  katalog: 'What the services can do here',
  katalogHinweis:
    'Per service, the questions for which a path is built here — and whether it is open **right now**. This does not send a request; it reads switches, keychain and role. A run against the real service holds for the day and the version it ran with.',
  katalogFragen: {
    suche: 'Card search',
    kartendaten: 'Card data',
    bild: 'Card images',
    regeln: 'Rulings',
    manakosten: 'Check mana costs',
    setangaben: 'Set details',
    preise: 'Prices',
    sammeldaten: 'Bulk data',
  },
  katalogStatus: {
    bereit: 'ready',
    aus: 'off',
    zugangFehlt: 'credentials missing',
    gestoert: 'unavailable',
    gesperrt: 'blocked',
    nichtGewaehlt: 'not chosen',
  },
  katalogAngabeFehlt: '{angabe} is not stored',
  katalogNachweis:
    'Run against the real service on {datum} with version {fassung} —',
  katalogNachweisAktuell: 'with this version.',
  katalogNachweisFremd: 'not with this version.',
  katalogDiagnose: 'Diagnostics',
  katalogLauf: 'The run can be repeated: {lauf}',
  katalogImLauf: 'in the run',
  katalogNichtImLauf: 'not in the run',
  katalogVertrag: 'Contract and terms',
  katalogFelder: {
    basisadresse: 'Address',
    endpunkte: 'Endpoints',
    zugang: 'Access',
    tarif: 'Pricing',
    grenze: 'Limit',
    frist: 'Cache period',
    wirkung: 'Effect',
    bedingungen: 'Terms',
    abgerufen: 'Read on',
    beleg: 'Source',
  },
  katalogWirkung: {
    externSenden: 'sends data to a third-party service',
    externSchreiben: 'writes at a third-party service',
  },
}
