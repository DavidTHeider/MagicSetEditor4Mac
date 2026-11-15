// The texts of the **set list** — open, create, import.
//
// Split out of `seiten.ts` when the file grew past 400 lines (§2).

export const sets = {
  auswahlLoesen: 'Show all sets',
  erklaerung:
    'Every open set lives in the store. From here you reach its cards — or write it back to a file.',
  titel: 'Sets',
  leer: 'No set open yet. Open a .mse-set file to get started.',
  setOeffnen: 'Open set …',
  neuTitel: 'A new set',
  neuKopf: 'New set',
  einlesenTitel: 'Read a set from elsewhere',
  neuHinweis:
    'An empty set — game and style decide what its cards look like. The style can be changed later.',
  neuName: 'Name of the set',
  neuStil: 'Style for the new set',
  neuAnlegen: 'Create set',
  neuSchritte: 'Steps to a new set',
  neuSchrittArt: 'Kind',
  neuSchrittAussehen: 'Look',
  neuSchrittName: 'Name',
  neuArtFrage: 'What are the cards for? We look up the package name for you.',
  neuArtZusatz: '· {stile} available',
  neuEigenes: 'Something entirely your own — create a game package',
  neuWeiter: 'Next',
  neuSchrittZurueck: 'Back',
  neuUeberspringen: 'Skip',
  neuWahl: '{spiel} · {stil}',
  neuOhneSpiele:
    'No game package is installed. Without one there are no card fields — packages come from the packages page.',
  neuOhneStile:
    'No style package is installed for {spiel}. Without a style no set can be created — packages come from the packages page.',
  neuZuDenPaketen: 'Go to packages',
  neuZurueck: 'Press ⌘[ to come back here; what you entered is still there.',
  beispielsets: 'Shipped to try out',
  geoeffnet: 'opened',
  schonOffen: 'This file is already open — here it is.',
  teilerfolg:
    'The file was not read to the end. {anzahl} sets are in the database nonetheless — they are in the list below.',
  beimOeffnenBefunde:
    'Not readable while opening. The set is there all the same; these places are missing from it:',
  // **The preview before reading in** — FMT-03.AK-01 (#1692). It appears
  // only when there is something to say.
  vorschau: {
    ueberschrift: 'Before reading in: {name}',
    umfang: '{karten} · game {spiel} · style {stil}',
    umfangOhneStil: '{karten} · game {spiel}',
    schonOffen:
      'This file is already in your library — Open shows the existing set instead of adding a second one.',
    befunde: 'These places cannot be read; the set would come without them:',
    bedarf:
      'Packages are missing. You can still read it in — the cards are drawn with a stand-in until the packages are there:',
    einlesen: 'Read in',
    aufschlagen: 'Open',
    verwerfen: "Don't read in",
  },
  // **Deletion** (#1218) — final, confirmed by name.
  loeschen: 'Delete …',
  loeschenTitel: 'Delete this set',
  loeschenWarnung:
    '{name} disappears from the store, along with {karten}. There is no trash and no undo. Whatever lives in a file stays there — what is deleted is the copy in the store.',
  loeschenSammlung:
    'Your collection stays: a copy belongs to a printing, not to a set.',
  loeschenFeld: 'Type the name of the set to confirm',
  loeschenPasstNicht: 'The name does not match {name} yet.',
  loeschenJa: 'Delete for good',
  loeschenNein: 'Cancel',
  loeschenBefunde:
    'The set is gone. These leftovers could not be cleared away:',

  spalten: {
    name: 'Name',
    spiel: 'Game',
    karten: 'Cards',
    herkunft: 'Origin',
  },
  mtgjsonUeberschrift: 'MTG JSON',
  mtgjsonErklaerung:
    'Card definitions from mtgjson.com — a single file or a directory. Which card fields they map to depends on the game.',
  mtgjsonSpiel: 'Game',
  mtgjsonDatei: 'Open file …',
  mtgjsonVerzeichnis: 'Open directory …',
  mtgjsonKeinSpiel:
    'No game has a mapping for MTG JSON fields. Without one, nothing can be read in.',
  scryfallsatzUeberschrift: 'A set from Scryfall',
  scryfallsatzErklaerung:
    'The set code decides which cards arrive — with their texts, without images. Name, mana cost, types, rules text, flavour text, power, toughness, rarity and illustrator come along; fetch the images per card. Pick the style afterwards.',
  scryfallsatzKeinSpiel:
    'Without a game that has a mapping, no set can be imported.',
  scryfallsatzSpiel: 'Game for the mapping',
  scryfallsatzCode: 'Set code on Scryfall',
  scryfallsatzPlatzhalter: 'e.g. lea',
  scryfallsatzEinlesen: 'Import set',
  vorlageAusfuhr: 'Via an export template …',
  vorlageWaehlen: 'Export template',
  vorlageAusfuehren: 'Run',
  vorlageSchliessen: 'Close',
  vorlageKeine:
    'No export template is installed for this game. Templates come from downloaded packages.',
  vorlageGeschrieben: 'Written to',
  vorlageBefunde:
    'Something came up while running — the file was written anyway:',
  vorlageGescheitert:
    'This template does not run through. What it is missing is below:',
  mtgjsonSichern: 'Save as MTG JSON …',
  mtgjsonGesichert: 'cards saved to',
  mtgjsonVerluste:
    'Not expressible on the way out. The file is there; these details are missing from it:',
  mtgjsonSpalte: 'Export',
  speichern: 'Save',
  speichernUnter: 'Save as …',
  speicherSpalte: 'Save',
  gespeichert: 'Saved to',
  beimSpeichernBefunde:
    'Not carried over when saving. The set file was written; these files beside it are missing:',
  oeffnen: "Show this set's cards",
  handlungenAn: 'Actions — {name}',
  handlungenKurz: '…',
  vorbehalte: {
    mitZahl: 'Not understood ({anzahl})',
    ueberschrift: 'Not interpreted',
    erklaerung:
      'The import did not understand these lines. They travel along unchanged and will not be lost when saving.',
    keine: 'Nothing — this set was read completely.',
  },

  // **The portable envelope** — work item N4 of phase 5.4
  // (#1457).
  umschlagAblegen: 'Save as an envelope',
  umschlagGeschnuert:
    'Saved. The envelope is signed — whoever receives it can check that it came from here.',
  umschlagTitel: 'Take in an envelope',
  umschlagErklaerung:
    'An envelope is a set someone else saved. It is inspected first and taken in afterwards — nothing in the database changes before that.',
  umschlagWaehlen: 'Inspect an envelope …',
  umschlagUrheber: 'Signed by',
  umschlagAbdruckHinweis:
    'The fingerprint says WHO signed — not whether you should trust them. If you know the person, compare it with them.',
  umschlagInhalt: 'It holds',
  umschlagUebernehmen: 'Take it in',
  umschlagVerwerfen: 'Never mind',
  umschlagUebernommen:
    'Taken in. The set is newly created — it has its own identifiers and overwrites nothing.',
}
