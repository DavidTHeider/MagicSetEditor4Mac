// Die Texte der **Turniere** — Werkstück 5 von Phase 4.7
// (#1294),
// nach Vermerk 0062.
//
// **Ein eigener Baustein**, wie `seitendecks.ts` und aus demselben Grund:
// `seiten.ts` steht an der Zeilengrenze aus §2.
export const turniere = {
  alleTurniere: 'All tournaments ({anzahl})',
  titel: 'Tournaments',
  erklaerung:
    'A tournament is a course over time: rounds, pairings, results. The standings are computed from it and stored nowhere — correct a result and you see the new standings at once.',

  leer: 'No tournament yet. “Create tournament …” above starts one, “Import tournament …” brings one in from a file.',

  neu: 'Create tournament',
  neuOeffnen: 'Create tournament …',
  name: 'Name',
  spiel: 'Game',
  notiz: 'Note',
  einzelspiel: 'One game per match',
  einzelspielHinweis:
    'Then game-win and opponents’ game-win percentage say nothing — they copy the match figures. They are computed anyway.',
  anlegen: 'Create',

  // **The wizard** — DOM-04.AK-01 (#1699).
  assistent: 'A new tournament in four steps',
  schrittTurnier: 'Tournament',
  schrittRegeln: 'Rules',
  schrittTeilnehmende: 'Participants',
  schrittErsteRunde: 'First round',
  weiter: 'Next',
  zurueck: 'Back',
  verlassen: 'Continue later',
  vorlage: 'Template',
  ohneVorlage: 'No template',
  vorlageHinweis:
    'A template brings game, format, rules and note. The names too if you like, never the decks: a submitted list belongs to one tournament.',
  mitTeilnehmenden: 'Take over the players’ names',
  nameWarum:
    'Required — so you find it in the list again. Two tournaments of one game have different names.',
  spielWarum: 'Required — it decides which formats and decks fit.',
  format: 'Format',
  ohneFormat: 'No format',
  formatWarum: 'Optional — you can pick from this game’s formats.',
  regelnAusVorlage: 'The rules come from the template.',
  schweizerSystem:
    'Pairings follow the Swiss system: players with the same points play each other, and nobody meets the same opponent twice. The app suggests how many rounds based on the number of players.',
  zuWenige:
    'A round needs two people to play each other — register at least two.',
  ersteRundeHinweis:
    'This is how the first round would look. Nothing is paired until you press the button.',
  ersteRundeAuslosen: 'Pair round 1',

  waehlen: 'Open',
  loeschen: 'Delete tournament',
  handlungenAn: 'Actions — {name}',
  handlungenKurz: '…',
  turniermenue: 'Tournament',
  loeschenFrage:
    'Delete this tournament? Rounds, pairings and results go with it. The players’ decks stay.',
  loeschenJa: 'Delete',
  loeschenNein: 'Cancel',

  teilnehmer: 'Participants',
  teilnehmerLeer: 'Nobody registered yet.',
  anmelden: 'Register',
  gebiete: 'Areas of this tournament',
  gebietTeilnehmer: 'Participants',
  gebietRunden: 'Rounds',
  gebietRangliste: 'Standings',

  // **Closing and log** — DOM-04.AK-04 (#1699).
  abschliessen: 'Close',
  abgeschlossenAm: 'Closed on {datum}. Players, rounds and results are final.',
  wiedereroeffnen: 'Reopen',
  nichtWiedereroeffnen: 'Don’t reopen',
  wiedereroeffnenFrage: 'Reopen this tournament?',
  wiedereroeffnenErklaerung:
    'Afterwards players, rounds and results can change again. The log records that it happened.',
  gebietProtokoll: 'Log',
  protokollLeer: 'Nothing has happened yet.',
  protokollAngelegt: 'Created as “{angaben}”',
  protokollAngemeldet: '{angaben} registered',
  protokollAusgestiegen: '{angaben} dropped',
  protokollEingestiegen: '{angaben} registered again',
  protokollRunde: 'Round {angaben} paired',
  protokollErgebnis: 'Result {paarung}: {angaben}',
  protokollZurueckgenommen: 'Result {paarung} withdrawn',
  protokollAbgeschlossen: 'Closed',
  protokollWiedereroeffnet: 'Reopened',
  protokollEingelesen: 'Imported as “{angaben}”',

  // **File and report** — DOM-04.AK-03 (#1699).
  ausgeben: 'Export …',
  berichtSichern: 'Save report …',
  einlesen: 'Import tournament …',
  gesichertUnter: 'Saved to {pfad}.',
  eingelesen: 'Imported and opened.',
  einlesebefunde: 'Noticed while importing:',

  // **Header and links after the artboard** — DOM-04.AK-02 (#1699).
  kopfzeile: 'Swiss system · Round {runde} of {von} · {anzahl} players',
  kopfzeileOhneRunde: 'Swiss system · {anzahl} players',
  kopfzeileOhneEmpfehlung: 'Swiss system · Round {runde} · {anzahl} players',
  gebietDecklisten: 'Decklists',
  decklistenLeer:
    'Nobody has submitted a list yet — choosing a deck when registering submits its list.',
  deckAufschlagen: 'Open deck',
  partieOeffnen: 'Open game',
  deck: 'Deck',
  ohneDeck: 'No deck',
  deckHinweis:
    'If a deck is named, its list is copied rather than referenced: a decklist in a tournament is a submitted document.',
  deckabzug: 'Submitted list',
  deckabzugLeer: 'No deck submitted.',
  deckabzugUnbenannt: 'Card',
  aussteigen: 'Drop',
  einsteigen: 'Re-enter',
  ausgestiegen: 'dropped',
  ausgestiegenHinweis:
    'They are no longer paired and stay in the standings: their percentage still counts towards their opponents’.',

  runden: 'Rounds',
  rundenLeer: 'No round paired yet.',
  runde: 'Round {nummer}',
  empfehlung:
    '{gelegt} of {empfohlen} rounds — the tournament rules recommend {empfohlen} for {spieler} players. A suggestion, not a rule.',
  vorschlag: 'Preview next round',
  vorschlagTitel: 'This is how the next round would look',
  vorschlagHinweis: 'Previewed, not created. Only the button below creates it.',
  rundeLegen: 'Create round',
  bye: 'Bye (2–0)',
  gegen: 'vs',
  offen: 'not played yet',
  spiele: 'Games',
  spieleA: 'Games for the first',
  spieleB: 'Games for the second',
  unentschiedenSpiele: 'Draws',
  eintragen: 'Record',
  zuruecknehmen: 'Undo',
  ergebnis: 'Result',

  // **The effects of a correction** — DOM-04.AK-02 (#1699).
  folgenFrage: 'This correction changes the standings.',
  folgenPlatz:
    '{name}: place {vorher} → {nachher}, {punkteVorher} → {punkteNachher} points',
  folgenSpaeter:
    'Round {runden} was paired on the old standings and stays as it is.',
  folgenEintragen: 'Enter anyway',
  folgenNicht: 'Don’t enter',
  ergebnisZahlen: 'Games are counted in whole numbers from zero upwards.',
  ergebnisOhneSpiel:
    'A result without a single game is not a result — whoever has not played has none yet.',
  ergebnisZuViele: 'A match here runs over at most {hoechstens} games.',

  rangliste: 'Standings',
  ranglisteLeer: 'Nothing played yet.',
  platz: 'Place',
  bilanz: 'Record',
  matchpunkte: 'Points',
  gegnermatchquote: 'OMW',
  spielquote: 'GW',
  gegnerspielquote: 'OGW',
  quotenHinweis:
    'Four tiebreakers in the order of the Magic Tournament Rules: points, opponents’ match-win, game-win, opponents’ game-win. Each percentage has a floor of 0.33.',

  // **Eine Paarung, an eine Partie gebunden** (G7, #1464).
  match: 'Played here',
  matchErklaerung:
    'If this pairing is played in this program, bind it to the game. The result is then reported by both sides and only entered once both say the same.',
  matchNichtGebunden: 'No game is attached to this pairing.',
  matchAnsehen: 'Check',
  matchPartie: 'Game ID',
  matchSitzA: 'Seat of the first player',
  matchSitzB: 'Seat of the second player',
  matchBinden: 'Bind to this game',
  matchGebunden: 'Bound to game',
  matchSitz: 'Seat',
  matchDeckabdruck: 'Deck fingerprint',
  matchKeineMeldung: 'Nobody has reported yet.',
  matchMeldungVon: 'reported by',
  matchMelden: 'Report for this seat',
  matchAusDerPartie: 'Report the result from the game',
  matchAusDerPartieHilfe:
    'Score and proof come from the finished game — you only say who is reporting.',
  matchNachweis: 'Proof (last chain link)',
  matchUebernehmen: 'Enter result',
  matchUebernommen: 'The result is in the tournament.',
  matchGrenze:
    'The proof says both sides saw the same history. It does not say the game was played correctly — and a private round is not official tournament sanctioning.',
}
