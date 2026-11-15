// **The script console** — `seiten.karten.konsole`, split out because
// `seitenkarten.ts` reached the line limit (§2, U13 of phase 5.1, #1424).
export const konsole = {
  ueberschrift: 'Script console',
  erklaerung:
    'An MSE script expression in the context of the open card. The environment holds the same names as when drawing: "card", "set", "stylesheet" and the game package functions. It changes nothing.',
  ausdruck: 'Expression',
  platzhalter: 'card.name',
  auswerten: 'Evaluate',
  ergebnis: 'Result',
  art: 'Type: {art}',
  keineKarte: 'A card must be open for this.',
  zuDenKarten: 'Go to the cards',
  vorbehalte: 'Reservations while building the environment',
  stelle: '(line {zeile}, column {spalte})',
  namen: 'Available names',
  namenErklaerung:
    'What an expression may name on this card — the library and whatever the template adds. A click inserts the name.',
  beispiele: 'Examples',
  beispieleImmer: 'Run everywhere',
  beispieleImmerErklaerung:
    'A click inserts the expression. These need nothing but the language itself.',
  beispieleVorlage: 'Need a template',
  beispieleVorlageErklaerung:
    '“card” and “set” are put into the environment by the style package. If none is installed for this card, the answer says so — the examples above still run.',
  beispielListe: {
    rechnung: 'Arithmetic',
    grossSchreiben: 'A string in capitals',
    laenge: 'How many characters something has',
    ausschnitt: 'A piece of a string',
    bedingung: 'A condition',
    name: 'The name of this card',
    grossName: 'Its name in capitals',
    regeltext: 'The length of its rules text',
    setangabe: 'The title of the set',
  },
  vervollstaendigung: 'Suggestions',
  uebernehmen: 'Accept',
  namenFiltern: 'Narrow the names',
  namenGezeigt: '{gezeigt} of {ganz} are shown — type to narrow them down.',
  namenLeer: 'No name matches that.',
  kuerzel: '⌘↩ evaluates.',
  abbrechen: 'Cancel',
  laeuft: 'Running …',
  dauer: '{ms} ms',
  abgebrochen:
    'Cancelled — the run ended before a result was there. The expression is still in the field.',
  baum: 'Result as a tree',
  verlauf: 'History of this session',
  verlaufErklaerung:
    'The last twenty expressions, in this session only. A click loads the expression back; whatever sounds like a token or password is redacted here and cannot be loaded back.',
  verlaufLeer: 'Nothing evaluated yet.',
  verlaufLoeschen: 'Clear history',
  ausgangGelungen: 'succeeded',
  ausgangGescheitert: 'failed',
  ausgangAbgebrochen: 'cancelled',
}
