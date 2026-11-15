// **Die Texte des Dienstkatalogs** — Werkstück A1 von Phase 5.3
// (#1440).
//
// **Aus `seiteneinstellungen.ts` herausgelöst am 2026-09-21**, als die
// Datei über die Zeilengrenze aus §2 wuchs. Der Schnitt folgt der Frage:
// dort steht, wie die Einstellungsseite aufgebaut ist, hier, **was ein
// Dienst kann** — die Schlüssel der Fragen und Zustände kommen aus
// `internal/domain/dienstvertrag.go` und `dienstlage.go`.
export const dienstkatalog = {
  katalog: 'Was die Dienste hier können',
  katalogHinweis:
    'Je Dienst die Fragen, für die es hier einen gebauten Weg gibt — und ob er **jetzt** offen ist. Diese Auskunft schickt keine Anfrage; sie liest Schalter, Schlüsselbund und Rolle. Ein Lauf gegen den echten Dienst gilt für den Tag und die Fassung, an denen er lief.',
  katalogFragen: {
    suche: 'Kartensuche',
    kartendaten: 'Kartendaten',
    bild: 'Kartenbilder',
    regeln: 'Regelhinweise',
    manakosten: 'Manakosten prüfen',
    setangaben: 'Setangaben',
    preise: 'Preise',
    sammeldaten: 'Sammeldaten',
  },
  katalogStatus: {
    bereit: 'bereit',
    aus: 'aus',
    zugangFehlt: 'Zugang fehlt',
    gestoert: 'gestört',
    gesperrt: 'gesperrt',
    nichtGewaehlt: 'nicht gewählt',
  },
  katalogAngabeFehlt: '{angabe} ist nicht hinterlegt',
  katalogNachweis:
    'Gegen den echten Dienst gelaufen am {datum} mit Fassung {fassung} —',
  katalogNachweisAktuell: 'mit dieser Fassung.',
  katalogNachweisFremd: 'nicht mit dieser Fassung.',
  // **Der Weg zum Lauf steht unter „Diagnose"** — UX-08 S11.NAT-02 (#1680):
  // ein Pfad in den Quelltext ist keine Entscheidung, die hier jemand trifft.
  katalogDiagnose: 'Diagnose',
  katalogLauf: 'Der Lauf lässt sich wiederholen: {lauf}',
  katalogImLauf: 'im Lauf',
  katalogNichtImLauf: 'nicht im Lauf',
  katalogVertrag: 'Vertrag und Bedingungen',
  katalogFelder: {
    basisadresse: 'Adresse',
    endpunkte: 'Endpunkte',
    zugang: 'Zugang',
    tarif: 'Tarif',
    grenze: 'Grenze',
    frist: 'Frist',
    wirkung: 'Wirkung',
    bedingungen: 'Bedingungen',
    abgerufen: 'Nachgelesen am',
    beleg: 'Beleg',
  },
  katalogWirkung: {
    externSenden: 'schickt Daten an einen fremden Dienst',
    externSchreiben: 'schreibt bei einem fremden Dienst',
  },
}
