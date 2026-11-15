// The texts of the **Play** area — work item G1 of phase 5.5
// (#1458).
import { abzugworte } from './spielenabzug'
import { zugworte } from './spielenzug'
import { menueworte } from './spielenmenue'
import { vorhangworte } from './spielenvorhang'
import { partienwoerter } from './spielenpartien'
import { regelwoerter } from './spielenregeln'
import { anfangworte } from './spielenanfang'
import { beginnworte } from './spielenbeginn'
import { gastworte } from './spielengast'
import { netzworte } from './spielennetz'
import { leihewoerter } from './spielenleihe'

export const spielen = {
  ...abzugworte,
  ...zugworte,
  ...menueworte,
  ...vorhangworte,
  ...partienwoerter,
  ...regelwoerter,
  ...anfangworte,
  ...beginnworte,
  ...gastworte,
  ...netzworte,
  ...leihewoerter,

  titel: 'Play',
  erklaerung:
    'A game is a table with frozen decks. What happens here is a stream of events — each hangs on the previous one, and none is overwritten. Move a card and you immediately see what the others see of it.',
  anfangTitel: 'Start a game',
  anfangErklaerung:
    'Every seat brings a deck, and it is frozen: changing your deck afterwards does not change this game. The game comes from the first deck chosen.',
  leerDecksTitel: 'No deck yet',
  leerDecks: 'Playing needs a deck. Create one under Decks.',
  sitz: 'Seat',
  sitzName: 'Name at the table',
  sitzDeck: 'Deck',
  sitzLeihe: 'Borrowed from …{von} · {karten} cards',
  deckWaehlen: 'Choose a deck',
  anschlussWartet:
    '“{name}” came along and takes the first seat. The second one is free.',
  sitzHinzu: 'Add a seat',
  sitzWeg: 'Remove this seat',
  sitzGruppe: 'Seat {nummer}',
  spiel: 'Game',
  anfangen: 'Start the game',
  tischTitel: 'At the table',
  partieGegen: 'Game against {namen}',
  partieZwischen: 'Game between {namen}',
  partieVon: 'Solo game: {namen}',
  fassung: 'Revision',
  wirt: 'Host',
  wirtHinweis:
    'The host holds the full state of this game and therefore sees more than their own view shows. It says so here because this is to be known before joining, not after.',
  tischAufHaufen: '{zahl} in a pile',
  tischMarkeEine: '{zahl} {art} counter',
  tischMarkeViele: '{zahl} {art} counters',
  tischArt: {
    spielstein: 'Token',
    emblem: 'Emblem',
    faehigkeit: 'Ability on the stack',
  },

  sicht: 'View for',
  sichtZuschauer: 'Spectator (no seat)',
  sichtWechseln: 'Switch seat',
  sichtLaedt: 'Loading the chosen view …',
  verlassen: 'Close the game',
  zonen: 'What lies where',
  zoneBibliothek: 'Library',
  zoneHand: 'Hand',
  zoneTisch: 'Battlefield',
  zoneFriedhof: 'Graveyard',
  zoneExil: 'Exile',
  zoneStapel: 'Stack',
  zoneKommando: 'Command zone',
  zonekurzBibliothek: 'Lib.',
  zonekurzFriedhof: 'Grave',
  zonekurzExil: 'Exile',
  zonekurzKommando: 'Cmd.',
  zoneLeer: 'empty',

  unbekannterTyp: 'Unknown type: {woerter}',
  geheim: 'face down',
  getappt: 'tapped',
  befehle: 'Actions',
  regelnTitel: 'What is enforced',
  regelwerk: 'Rule set',
  wirksamAb: 'effective',
  hoechsteStrenge: 'Highest strictness',
  strenge: 'Strictness',
  andereSeite: 'Other face',
  grossAnsehen: 'View large',
  ersteSeite: 'First face',
  stelle: 'To position',
  stelleOben: 'top',
  stelleUnten: 'bottom',
  stelleZufaellig: 'shuffled in',
  mischpflichtTitel: 'Open duties to shuffle',

  legendeTitel: 'Open choice: the legend rule',
  legendeSatz:
    '{wer} must pick the legendary permanent named “{name}” that stays (CR 704.5j).',
  legendeBehalten: 'Keep this one',

  reihungTitel: 'Open choice: the order of the abilities',
  reihungSatz:
    '{wer} controls several abilities that triggered at the same time and says in which order they go on the stack (CR 603.3b). Whatever is put on first ends up at the bottom and resolves last.',
  reihungLegen: 'Put on next:',
  stand: 'State',
  standLaeuft: 'running',
  standPausiert: 'paused',
  standBeendet: 'finished',
  pausieren: 'Pause',
  fortsetzen: 'Resume',
  beenden: 'Finish',
  beendenFrage:
    'Finish this game? Finishing is final — a finished game cannot be resumed.',
  beendenJa: 'Finish',
  beendenNein: 'Cancel',
  ereignisse: 'What just happened',
  ereignisseLeer: 'Nothing yet.',
  ereignisseLuecke:
    'Some moves may be missing here: more updates arrived while this page was hidden than can be caught up.',
  ereignisPrivat: 'For you only',
  ereignisSatz: '{sitz}: {handlung} — {objekt}',
  ereignisSatzOhneKarte: '{sitz}: {handlung}',

  // **Die Rückschau** (G5, #1462).
  verlauf: 'The story so far',
  verlaufAnsehen: 'Show history',
  verlaufNachladen: 'Reload',
  verlaufLaeuft: 'Fetching …',
  verlaufUngeholt:
    'The history has not been fetched yet. It shows every move of this game, as far as this seat may see it.',
  verlaufLeer: 'Nothing has happened in this game yet.',
  verlaufGrenze:
    'What another seat saw for itself alone is not here — not after the fact either.',

  // **Was von einer Partie übrig bleibt** (G5, #1462).
  protokoll: 'What remains of this game',
  protokollErklaerung:
    'Every move of this game is stored on this computer, and each seat sees its own part. You can release your part for later — to a single opponent or to everyone.',
  protokollAnsehen: 'Show releases',
  protokollNiemand: 'Nobody has released anything.',
  protokollFuer: 'Released to',
  protokollAlle: 'everyone',
  protokollSitz: 'Seat',
  protokollEmpfaenger: 'Recipients, comma separated — or * for everyone',
  protokollFreigeben: 'Release',
  protokollZuruecknehmen: 'Withdraw',
  protokollNurSpaeter:
    'A release only takes effect after the game. At the table everyone still sees only what is theirs.',
  protokollKeineRueckholung:
    'Withdrawing closes the door — it does not retrieve what somebody has already read.',
  protokollAufbewahrung:
    'A finished record is offered for review in data maintenance after 90 days. It is removed only when you say so.',
  protokollLoeschen: 'Delete record',
  protokollLoeschenErklaerung:
    'This removes the state, history and releases of this game from this computer. It says nothing about copies elsewhere.',
  protokollGeloescht: 'The record has been deleted.',
  protokollLoeschenFrage: 'Delete the record of this game?',
  protokollLoeschenJa: 'Delete record',

  // **Die Diagnose der Leitung** (G5, #1462).
  diagnose: 'Connection status',
  diagnoseErklaerung:
    "This shows what this table talks over and how much is waiting. None of it says anything about cards — yours or anyone else's.",
  diagnoseAnsehen: 'Check connection',
  diagnoseStill: 'No table is running.',
  diagnoseAlsWirt: 'This computer hosts the table.',
  diagnoseAlsGast: "This computer is seated at somebody else's table.",
  diagnosePausiert: 'The connection dropped — the seat remains.',
  diagnoseWeg: 'Route',
  diagnoseGegenueber: 'Peer',
  diagnoseSeit: 'Up for',
  diagnoseWartend: 'Waiting',
  diagnoseVerworfen: 'Dropped',
  diagnoseStand: 'At sequence',
  diagnoseRunde: 'Last round trip',
  diagnoseBildzeit: 'Frame interval in this window',
  diagnoseUnbekannt: 'unknown',
  diagnoseOhneLeitung: 'No connection is up.',

  // **Die Freigabe für Gäste** (G5, #1462).
  freigabe: 'Let others play at this table',
  freigabeErklaerung:
    'Being reachable and letting others play are two decisions. Opening the table lets a peer make moves in this game — you hand out the seats, not them.',
  freigabeOffen: 'This table is open.',
  freigabeZu: 'This table is not open.',
  freigabeStarten: 'Open table',
  freigabeBeenden: 'Close table',
  freigabeNachsehen: 'Check again',
  freigabeSitze: 'Free for',
  freigabeNiemand: 'Nobody is seated yet.',
  freigabePartner: 'That is',
  freigabePartnerFremd: 'This peer is not among your contacts.',
  freigabeBesetzt: 'Seated:',
  freigabeOhneHorchposten:
    'For a peer to arrive, this machine also has to be reachable — that lives under Settings.',

  // **Der gezeichnete Tisch** — Werkstück G3 von Phase 5.5
  // (#1460).
  tischBild: 'The table',
  tischListe: 'As a list',
  tischGrafisch: 'As a table',
  tischDarstellung: 'Display',
  tischEinpassen: 'Show everything',
  tischNaeher: 'Zoom in',
  tischWeiter: 'Zoom out',
  tischDrehen: 'Rotate the table',
  tischAuswahl: 'To the selection',
  tischEigenerSitz: 'To my seat',
  tischBlick: 'Table view',
  tischZiel: 'Jump to',
  tischOhneWebgl:
    'This machine provides no accelerated graphics. The table appears below as a list, and everything that works on the drawn table works there.',
  tischOhneRenderer:
    'The table could not be drawn. The list below holds every card, and everything else works there.',
  tischVerloren:
    'The graphics connection was lost. The table returns once you leave it and open it again; until then the list below holds every card.',
  tischHeimatlos:
    'These objects lie in a zone this program does not know and therefore do not appear on the table:',
  tischVerdeckt: 'face down',
  tischGetappt: 'tapped',
  tischHaengtAn: 'attached to',
  tischbild: 'Background',
  tischbildHolz: 'Wood',
  tischbildDunklesHolz: 'Dark wood',
  tischbildPicknick: 'Picnic',
  tischbildDungeon: 'Dungeon',
  tischbildTaverne: 'Tavern',
  tischbildTurniertisch: 'Tournament table',
  tischEigenes: 'Own image …',
  tischEigenesWeg: 'Remove own image',
  gespraech: 'Table talk',
  gespraechVerlauf: 'What was said at the table',
  gespraechEingabe: 'Say something',
  gespraechSenden: 'Say',
  gespraechLeer: 'Nobody has said anything yet.',
  gespraechNiemandZeigt: 'Nobody is pointing at anything right now.',
  gespraechZeigt: '{sitz}: {tut}',
  gespraechZeigtAuf: '{sitz}: {tut} {objekt}',
  zeigerZeiger: 'pointing',
  zeigerGreift: 'dragging',
  zeigerSchwebt: 'looking at',
  zeigerPing: 'calling',
  spielart: 'Mode',
  spielartGoldfish: 'Practice alone',
  spielartSelfplay: 'Several seats',
  format: 'Format',
  fehltDeck:
    'Choose a deck for every seat — the first one determines the game.',
  fehltEigenesDeck:
    'One seat needs a deck of your own — a loan brings no game with it, and the game comes from the first deck of your own.',
  fehltName:
    'Every seat needs a name; at the table it tells the players apart.',

  // **Die digitale Deckleihe** (G6, #1463).
  leihe: 'Lend or borrow a deck',
  leiheOeffnen: 'Lend or borrow …',
  leiheVerleihenTitel: 'Lend a deck',
  leiheAnnehmenTitel: 'Accept a loan',
  leiheListeTitel: 'Your loans',
  leiheErklaerung:
    'Lending lets an opponent play with a frozen snapshot of your deck. The deck stays here, and not a single card of your collection changes.',
  leiheKeinEigentum:
    'This is a usage grant, not a transfer. Nothing changes owner, digitally or physically.',
  leiheKeinKopierschutz:
    'And it is not copy protection: whoever has seen the cards has seen them. What the loan governs is what this program will start a game with.',
  leiheAnsehen: 'Show loans',
  leiheKeine: 'There are no loans.',
  leiheEntleiher: 'Who do you lend to?',
  leiheEntleiherLeer:
    'No confirmed peer yet. Add one in the box “Add or confirm a counterpart” below.',
  leiheDeck: 'Which deck',
  leiheDeckWaehlen: 'Choose a deck …',
  leiheKeinDeck: 'No deck yet — one is made under Decks.',
  leiheFachleute: 'Lend for one particular game only',
  leiheFachleuteErklaerung:
    'Without it, the loan applies to the peer’s next game. To bind it to exactly one, enter that game’s id — the peer finds it in their list of games.',
  leiheWoran: 'Game ID — empty means the next one',
  leiheStunden: 'Valid for how many hours',
  leiheZiel: 'Where to put the envelope',
  leiheAnbieten: 'Lend and save',
  leiheAbgelegt: 'The envelope is here:',
  leiheUmschlag: 'Envelope of a loan',
  leiheAnnehmen: 'Accept loan',
  leiheWiderrufen: 'Revoke',
  leiheZurueckgeben: 'Return',
  leiheVerliehen: 'lent to',
  leiheGeliehen: 'borrowed from',
  leiheBis: 'valid until',
  leiheKarten: 'cards',
  leiheVerbraucht: 'redeemed for',
  leiheOffen: 'not yet redeemed',
  leiheZurueckgenommen: 'revoked',
  leiheModusZuende: 'A running game plays out.',
  leiheModusAbbruch: 'A running game is stopped.',
  leiheRuecknahmeGrenze:
    'Revoking prevents new starts. It removes nothing from a running game and deletes no copy at the other end.',

  // **Wie die Partie ausgegangen ist** (M17, #1618).
  ausgangTitel: 'The result',
  ausgangSieger: 'The winner is',
  ausgangUnentschieden:
    'A draw — everyone still in the game lost at the same time.',
  grundAufgabe: 'conceded',
  standLeiste: 'Player standings',
  standDu: 'You',
  standLeben: 'Life',
  standGift: 'Poison',
  standMana: 'Mana',
  standStrahlung: 'Radiation',
  standSteuer: 'Tax',
  standCommanderschaden: '{name}: {schaden}',
  standCommanderTod: 'at 21 the game is over',
  standLeergezogen: 'drew from an empty library',
  grundLeben: 'was at zero life or below',
  grundBibliothek: 'was to draw from an empty library',
  grundGift: 'had ten poison counters',
  aufgeben: 'Concede',
  aufgebenFrage:
    'Concede? The game ends at once for this seat, and its cards leave with it.',
  aufgebenJa: 'Yes, concede',
  aufgebenNein: 'Continue',
  aufgebenHinweis:
    'You can concede at any time, and it takes effect at once. You leave the game, and your cards go with you.',

  // **Der Weg zurück in die eigene Partie** (#1600).

  // **Ein Sitz an einem Profil** (G8, #1465).
  sitzprofil: 'Profile',
  sitzprofilKeines: 'no profile',
  sitzprofilErklaerung:
    'If you keep two profiles, you can give each seat one. The game then starts in the view of the seat that belongs to your current profile; you reach the others from the view picker at the table. What the other view showed is gone. If you give no seat a profile, your current profile takes the first seat.',
  sitzprofilBefunde: 'There is a note about this:',
}
