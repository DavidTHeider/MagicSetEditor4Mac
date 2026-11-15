// Die Texte der **Kartenseite** — die größte Seite der Anwendung.
//
// Aus `seiten.ts` herausgelöst, als die Datei über 400 Zeilen wuchs (§2).
// Der Schnitt folgt der Frage: dort stehen die Texte **aller** Seiten, hier
// die der einen, die am meisten trägt — und weiter wachsen wird
// (#674).

import { verlaufEn } from './seitenverlauf'
import { massenvorschauEn } from './seitenmassen'
import { autoersetzung } from './seitenautoersetzung'
import { schluesselwoerter } from './seitenschluesselwoerter'
import { werkstatt } from './seitenwerkstatt'
import { konsole } from './seitenkonsole'
import { symbolbibliothek } from './seitensymbolbibliothek'

export const karten = {
  ...werkstatt,
  autoersetzung,
  schluesselwoerter,
  symbolbibliothek,
  anlegen: 'Add card',
  nachOben: 'Move up',
  nachUnten: 'Move down',
  duplizieren: 'Duplicate card',
  loeschen: 'Delete card',
  loeschenFrage: 'Delete this card? ⌘Z will undo it.',
  loeschenJa: 'Delete',
  loeschenNein: 'Cancel',
  loeschenTrotzdem: 'Delete anyway',
  symbole: 'Insert symbols',
  hergeleitet: 'Derived',
  hergeleitetWert: 'Derived: {wert}',
  bildWaehlen: 'Choose image',
  stilfelder: 'Style fields',
  stilfelderHilfe:
    'These fields come with the card’s style. Their value applies to this style only.',
  symbolZeichnen: 'Draw symbol',
  kartenskripteTitel: 'Cards from a script',
  kartenskripteErklaerung:
    'Ready-made moves from the game package — one click adds several cards at once. ' +
    'A single undo takes them all back.',
  kartenskriptEine: 'One card added.',
  kartenskriptMehrere: '{anzahl} cards added.',
  bildEinsetzen: 'Insert image',
  pack: {
    ueberschrift: 'Random pack',
    erklaerung:
      "A booster from this set's cards, following the rules of the game package. The same seed gives the same draw — that is how a result can be reproduced.",
    laedtGrund:
      'Every pack type asks every card whether it fits — that takes a moment the first time and is ready afterwards.',
    art: 'Pack type',
    same: 'Seed',
    sameZahl: 'The seed is a whole number from 0 to 18446744073709551615.',
    ziehen: 'Draw',
    neuerSame: 'New seed',
    gezogen: 'Drawn',
    keineArten:
      'The game of this set brings no pack types — or its package is not installed.',
    keineKarten: 'This pack type produced no card.',
    artMitZahl: '{name} ({zahl} cards)',
    nurBaustein: 'Building block — drawn by other pack types',
    vorbehalte: 'Pack types whose filter could not be evaluated',
  },
  symbol: {
    ueberschrift: 'Symbol',
    erklaerung:
      "This set's symbol — drawn, not loaded. It appears in the preview at once and goes into the file when the set is saved. Every shape can be placed and moved with the keyboard.",

    grundform: 'Add a basic shape',

    verlauf: 'Drawing history',
    nachzeichnen: 'Trace an image',
    groesser: 'Larger',
    nachzeichnenHilfe:
      'A PNG, JPEG or GIF with a clear silhouette becomes shapes with Bézier points you can keep editing here.',
    zuruecknehmen: 'Undo gesture',
    wiederherstellen: 'Redo gesture',
    kreis: 'Circle',
    rechteck: 'Rectangle',
    dreieck: 'Triangle',
    stern: 'Star',

    teile: 'Parts',
    keineTeile:
      'Nothing drawn yet. A basic shape is the start — every point can be changed afterwards.',
    teilLoeschen: 'Delete part',
    teilMitPunkten: '{name} — {zahl} points',
    teilMitKopien: '{name} — symmetry, {zahl} copies',
    gruppieren: 'Group',
    gruppierenHilfe:
      'Two or more selected parts become a group. Hold ⌘ to select several.',
    aufloesen: 'Ungroup',
    nachOben: 'Move up',
    nachUnten: 'Move down',
    nachObenFuer: 'Move {name} up',
    nachUntenFuer: 'Move {name} down',
    teilLoeschenFuer: 'Delete {name}',
    teilMitKindern: '{name} — {zahl} parts',
    teilKind: '{name}, inside a group',
    wahlHilfe:
      'Hold ⌘ to select several parts. The order is the drawing order: what stands first lies on top.',

    zeichnung: 'Drawing area',
    punkt: 'Point {nummer} of {name}',
    griffDavor: 'Handle before point {nummer}',
    griffDanach: 'Handle after point {nummer}',
    punktHinzufuegen: 'Insert point',
    punktLoeschen: 'Delete point',

    tastatur: 'With the keyboard',
    tastaturhilfe:
      'Tab moves between points. Arrow keys move by one hundredth, with Shift by one tenth. Enter inserts a new point after the selected one, Backspace deletes it.',

    kombination: 'Combine',
    ueberdecken: 'Overlap',
    vereinen: 'Merge',
    abziehen: 'Subtract',
    schneiden: 'Intersect',
    unterschied: 'Difference',
    umranden: 'Border',

    symmetrie: 'Symmetry',
    symmetrieAnlegen: 'Add symmetry',
    symmetrieAufloesen: 'Dissolve symmetry',
    symmetrieErklaerung:
      'A symmetry multiplies whatever lies inside it. The copies appear faintly on the canvas; only the original is ever edited.',
    symmetrieHinweis: 'Select a part to turn it into a symmetry.',
    symmetrieart: 'Kind',
    drehung: 'Rotation',
    achse: 'Reflection',
    kopien: 'Copies',
    kopienHinweis:
      'A reflection needs pairs — an odd number leaves one copy alone, and MSE rounds down.',
    mitteX: 'Centre, horizontal',
    mitteY: 'Centre, vertical',
    griffX: 'Handle, horizontal',
    griffY: 'Handle, vertical',
    mitte: 'Centre of the symmetry',
    griff: 'Handle of the symmetry',

    sichern: 'Put into the set',
    gesichert: 'Stored as {name}. It goes into the file when the set is saved.',
  },
  konsole,

  wortlisten: {
    ueberschrift: 'Choices',
    liste: 'Choose from {liste}',
    keine: 'none of these',
    gruppe: '{name} (group)',
  },
  regeln: {
    ueberschrift: 'Rulings from Scryfall',
    holen: 'Fetch rulings',
    keine: 'There are no rulings for this card.',
    offiziell: 'official',
    anmerkung: 'Scryfall note',
    anmerkungErklaerung:
      'Notes come from the Scryfall team and often explain an unofficial format. They are not rules.',
    quelle: 'View this card on Scryfall',
    alter: 'As of: {stunden} hours old',
    manaPruefen: 'Check mana cost',
    manaGut: 'The mana cost is fine.',
    manaSchlecht: 'Unknown: {symbole}',
    manaLeer: 'This card has no mana cost.',
  },
  abgleich: {
    ueberschrift: 'Card data from Scryfall',
    erklaerung: 'Shows what would change. Only what you tick is written.',
    holen: 'Fetch proposal',
    uebernehmen: 'Apply ticked',
    nichtsZuTun: 'Nothing to change — the card matches Scryfall.',
    alt: 'current',
    neu: 'new',
    leer: '(empty)',
    alter: 'State: {stunden} hours old',
    frisch: 'State: just fetched',
    quelle: 'View this card on Scryfall',
  },
  scryfall: {
    ueberschrift: 'Image from Scryfall',
    erklaerung:
      'The card name picks the image. The artist is written into the card as well — Scryfall requires the name next to the art.',
    name: 'Card name on Scryfall',
    art: 'Image type',
    artKlein: 'Small (preview)',
    artNormal: 'Full card',
    artAusschnitt: 'Art only',
    holen: 'Fetch image',
    ausschnittHinweis:
      'The art crop shows the illustration alone. It is refused when Scryfall does not name the artist.',
    quelle: 'View this card on Scryfall',
  },
  scryfallkarte: {
    name: 'Card name on Scryfall',
    platzhalter: 'e.g. Lightning Bolt',
    mitBild: 'with illustration',
    anlegen: 'Create from Scryfall',
  },
  zuschnitt: {
    ueberschrift: 'Choose the crop',
    erklaerung:
      'The frame says which part of the image ends up on the card. Drag to move it, drag a corner to resize.',
    quelle: 'Source image with the chosen crop',
    vorschau: 'Preview at field size',
    links: 'Left',
    oben: 'Top',
    breite: 'Width',
    hoehe: 'Height',
    seitenverhaeltnis: "Keep the field's aspect ratio",
    uebernehmen: 'Insert',
    abbrechen: 'Cancel',
    masse: '{breite} × {hoehe} pixels',
  },
  feldWaehlen: 'Edit this field',
  leerTitel: 'No cards yet',
  listeBezeichnung: 'Cards of the set',
  mehr: 'More',
  gruppeAnlage: 'Card',
  gruppeAnsicht: 'View',
  ausScryfallAufklapp: 'Create from Scryfall …',
  ausScryfallErklaerung:
    'Instead of a blank card: type a name or search Scryfall — the fields then come from there.',
  dieseKarte: 'This card',
  kartenmenue: 'Card',
  auswahlLoesen: 'Show all cards',
  bezugAufklapp: 'Link to a printing …',
  dieseKarteLeer: 'Open a card to edit it.',
  keineKarteTitel: 'No card selected',
  titel: 'Cards',
  leer: 'Create one or paste copied cards.',
  ohneName: 'Unnamed',
  ohneFelder: 'This card has no field values at all.',
  nummer: 'Number',
  vorschau: 'Preview of the card',
  keineKarte:
    'Pick one from the list — its fields and preview will appear here.',
  bearbeiten: 'Edit',

  notizen: {
    ueberschrift: 'Card notes',
    platzhalter: 'For you only — this never appears on the card.',
  },

  // **What the game package does not declare** (#1271).
  buchfuehrung: 'What the application knows about it',
  tiefe: 'How much you see',
  tiefeNormal: 'Normal',
  tiefeErweitert: 'Advanced',
  tiefeExperte: 'Expert',
  tiefeErweitertTitel: 'What a script fills in',
  tiefeErweitertHinweis:
    'The game package computes these — they say why something is on the card. They cannot be edited.',
  befundVerdeckt: 'A finding sits under “{tiefe}”.',
  befundSpringen: 'Go there',
  buchfuehrungHinweis:
    'These entries do not belong to the card but to its bookkeeping — which template it uses and what is set on it. Only change them if you know why.',

  // **Note and marks together, collapsed** (#1271). They sat below the list
  // and took its height: four rows were left.
  arbeitsstand: 'Progress on this card',
  arbeitsstandHinweis:
    'Note and marks — they belong to your work on the card, not to the card. A file you pass on does not carry them.',

  marken: {
    ueberschrift: 'Marks',
    erklaerung:
      'Words for your own work — they sit beside the card, not inside it. A file you pass on does not carry them.',
    platzhalter: 'e.g. done',
    setzen: 'Add mark',
    entfernen: 'Remove mark {marke}',
    keine: 'No mark on this card yet.',
    schonGesetzt:
      'The card already carries »{marke}« — set elsewhere in the meantime.',
    filter: 'Filter by mark',
    alle: 'All',
  },

  scryfallsuche: {
    frage: 'Search on Scryfall',
    platzhalter: 'e.g. Lightning Bolt — or t:creature c:r cmc<=3',
    erklaerung:
      "A card name is enough. If you know Scryfall's search language, you can use it here.",
    suchen: 'Search',
    nichtsGefunden: 'Nothing found. Try another word?',
    mehr: 'Scryfall has more results. Narrow the question — nothing is fetched automatically.',
    alter: 'State: {stunden} hours old',
    frisch: 'State: just fetched',
    befund: 'Finding: {text}',
  },

  // **Der Verweis auf einen echten Druck** (#1167).
  verweis: {
    aufschrift: 'Linked to',
    keiner:
      'This card is not linked to a printing. Syncing with Scryfall will then pick an arbitrary one — search above and link it.',
    druck: 'from {setcode} {nummer}',
    gesetzt: '— linked on {wann}',
    setzen: 'Link to {name} ({setcode} {nummer})',
    loesen: 'Unlink',
  },
  // **Die Zwischenablage** (#1168).
  ablage: {
    kopieren: 'Copy ({zahl})',
    ausschneiden: 'Cut ({zahl})',
    einfuegen: 'Paste',
  },

  // **Die Spalten der Kartenliste** (#1168).
  listenspalten: {
    ueberschrift: 'List columns',
    jeSpiel: 'The choice applies to every set of the same game.',
    filtern: 'Find column',
    gezeigt: '{gezeigt} of {alle} columns',
    ausgeblendet: '{anzahl} more are hidden — find them with the field above.',
    genauer: 'More match — narrow the search to see them.',
    nachOben: 'move further left',
    nachUnten: 'move further right',
    sortieren: 'Sort by',
    garNicht: 'Order of the set',
    aufsteigend: 'Ascending',
    absteigend: 'Descending',
    ohneVorlage:
      'The game package cannot be read. The columns come from the cards themselves \u2014 without captions and without the package\u2019s order.',
    spalteAusblenden: 'Hide column',
  },

  verlauf: verlaufEn,
  massen: {
    ...massenvorschauEn,
    keinFeld: '— choose a field —',
    keinWert: '— choose a value —',
    ueberschrift: 'Bulk change',
    erklaerung:
      'Sets one field on every ticked card — in a single operation. If anything goes wrong, not one is changed.',
    auswaehlen: 'Tick for the bulk change',
    mehrereWaehlen: 'Select several',
    ausgewaehlt: 'ticked',
    keineAuswahl:
      'No card ticked. Tick several in the list to change them together.',
    feld: 'Field',
    wert: 'New value',
    anwenden: 'Apply',
    betroffen: 'The change failed at these cards. Not one card was changed:',
    gescheitert: 'failed',
    ohneWert: 'no value yet',
  },
  stilTitel: 'Style',
  stilWaehlen: 'Style of this set',
  stilLeer:
    'No style installed. Without one the preview falls back to a stand-in.',
  stilFrage: 'Switch to this style?',
  stilSpielwechsel: 'The new style belongs to a different game.',
  stilUngedeutet:
    'The new game does not know these fields. Their values stay in the database — they will no longer be drawn:',
  stilVerloreneOptionen:
    'The new style has no such styling options. Their values are lost:',
  stilUebernehmen: 'Switch style',
  stilAbbrechen: 'Keep style',
  stilKeiner:
    'The set names a style that is not installed — choose one of the installed ones.',
  stiloptionen: 'Styling options',
  stiloptionenSichtbar: '{zahl} of {gesamt}',
  stiloptionenWeitere: 'More options ({zahl})',
  marktTitel: 'Where to find this card',
  marktErklaerung:
    'These links are built on your machine — from the card name or a provider identifier. Nothing goes out until you click, in the browser of your choice. No key is needed; whether anything is for sale there is up to the provider.',
  marktGenau: 'goes to exactly this card',
  marktSuche: 'goes to a search',
  marktNurAnbieter: 'goes to the provider only',
  umschlagAnnehmen: 'Card from an envelope',
  umschlagAngenommen: 'The card is now in this set.',
}
