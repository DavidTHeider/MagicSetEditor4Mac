// The **collection** page — piece 2 and 3 of phase 3.4
// (#1165).
export const sammlung = {
  titel: 'Collection',
  erklaerung:
    'What you actually own — one row per printing, condition, language and finish. This has nothing to do with your sets: a designed card is a draft, a copy is a piece of cardboard.',

  aufnehmenOeffnen: 'Add …',
  einfuhrOeffnen: 'Import and export …',
  einfuhrTitel: 'Import and export',
  leerTitel: 'Nothing recorded yet',
  leer: 'A copy needs the Scryfall ID of its printing — the UUID from a card page address. A whole list goes in at once.',

  druck: 'Printing',
  menge: 'Quantity',
  zustand: 'Condition',
  sprache: 'Language',
  ausfuehrung: 'Finish',
  foil: 'Foil',
  nonfoil: 'Nonfoil',
  etched: 'Etched',
  erwerbsdatum: 'Acquired on',
  erwerbspreis: 'Paid',
  erwerbswaehrung: 'Currency',
  preisMehrdeutig:
    'Two readings: {erste} or {zweite}. Please write it unambiguously.',
  notiz: 'Note',

  // **Der lesbare Bezug und die Sicht darauf** (#1200).
  name: 'Card name',
  setcode: 'Set',
  nummer: 'Number',
  sicht: 'View of the collection',
  suche: 'Search',
  suchePlatzhalter: 'name, set, id or note',
  sortierung: 'Sort by',
  sortierBezeichnung: 'Card',
  sortierMenge: 'Quantity',
  sortierErwerbsdatum: 'Acquired on',
  ausfuehrungAlle: 'All finishes',
  nurFoil: 'Foil only',
  nurNonfoil: 'Non-foil only',
  nurEtched: 'Etched only',

  navigator: 'View and tools',
  inspector: 'Card Copy',
  exemplarmenue: 'Card Copy',
  inDecks: 'In these decks',

  inspectorLeer:
    'Pick a copy in the list — quantity, condition, purchase and price appear here, and this is where you change it.',
  inspectorAusserhalb:
    'This copy is no longer in the current view. A change still hits this very copy.',
  arbeitsplatz: 'Collection workspace',
  listeBezeichnung: 'Copies',
  spalteWahl: 'Pick',
  spalteBezeichnung: 'Name',
  spalteMenge: 'Quantity',
  spalteZustand: 'Condition',
  spalteSprache: 'Language',
  spalteSet: 'Set',
  spalteLagerort: 'Storage',
  gruppeOhne: 'Not specified',
  gruppeSumme: '{anzahl} copies · {menge} pieces',
  kachelWaehlen: 'Select {bezeichnung}',

  ansicht: 'View',
  ansichtTabelle: 'Table',
  ansichtGalerie: 'Gallery',
  ansichtGruppen: 'Groups',
  gruppierung: 'Group by',
  gruppierungKeine: 'No grouping',
  gruppierungSet: 'Set',
  gruppierungLagerort: 'Storage',
  gruppierungZustand: 'Condition',
  gruppierungSprache: 'Language',
  ohneTreffer: 'No copy matches the search and filter.',
  ohneTrefferZuruecksetzen: 'Reset search and filter',
  gezeigt: '{gezeigt} of {gesamt}',
  gekappt: '{gezeichnet} of {passend} drawn',
  mehrZeigen: 'Show more',

  // **Die Listeneinfuhr** (#1292).
  listeTitel: 'Paste a list',
  listeErklaerung:
    'One card per line, e.g. “4 Lightning Bolt (M11) 149”. Blank lines, “Deck”, “Sideboard” and lines starting with # or // are skipped.',
  listeFeld: 'Card list',
  listePlatzhalter: '4 Lightning Bolt (M11) 149\n2x Counterspell',
  listeDeuten: 'Match',
  listeEinlesen: 'Add',
  listeOffen:
    '{offen} of {gesamt} lines need a decision — pick the printing you own.',
  listeAlleEindeutig: 'All {gesamt} lines are unambiguous.',
  listeDruckWahl: 'Which printing?',
  listeOhneWahl: '— not decided yet —',
  listeDrucke: '{drucke} printings',
  listeBericht: '{angelegt} new, {erhoeht} increased, {stueck} copies added.',
  listeNichtsGewaehlt: 'No line decided — pick at least one printing.',
  listeAusgelassen: '{ausgelassen} existing lines skipped.',
  listeKeinDruck: 'No card has this name — check the spelling.',
  listeZuViele:
    'More than {drucke} printings have this name — name the set in brackets, such as “(M11)”, and one remains.',
  listeOhneKarte: '{anzahl} lines find no card and are skipped.',
  listeOhneKarteEins: 'One line finds no card and is skipped.',
  vorschauTitel: 'Before importing',
  vorschauErklaerung:
    'Some lines match lines your collection already has — same printing, condition, language, finish and intent. Choose what happens to them; nothing is written until then.',
  vorschauZaehlung:
    '{neu} new · {zusammen} match existing lines · {wiederholt} repeat a line of this import',
  vorschauZusammen: '{bezeichnung}, {zustand}: {vorher} → {nachher}',
  vorschauModus: 'Existing lines',
  vorschauAddieren: 'Add quantities',
  vorschauAuslassen: 'Skip existing lines',
  vorschauEinlesen: 'Import',
  vorschauAbbrechen: 'Don’t import',

  // **Die Ausfuhr** (#1292).
  ausfuhrTitel: 'Export and back up',
  ausfuhrErklaerung:
    '“CSV” carries every field and reads back in full. “List” is the pasteable shape for a deck builder — it carries only quantity, card, printing and finish.',
  ausfuhrAlsCSV: 'As CSV',
  ausfuhrAlsListe: 'As list',
  ausfuhrErgebnis: 'Output',
  ausfuhrZurueck: 'Read CSV back',
  ausfuhrLeer: 'Nothing exported yet.',

  // **Die Ordnung** (#1292, Vermerk 0057).
  lagerort: 'Storage',
  wunsch: 'Wanted, not owned',

  // **Die Ordnungssicht** (#1292).
  ordnungTitel: 'Order',
  ordnungLagerorte: 'Storage locations',
  ordnungDubletten: 'Duplicates',
  ordnungWuensche: 'Wishlist',
  ordnungBehalten: 'Keep',
  ordnungLeerLagerorte: 'No copy yet — without one there is no storage place.',
  ordnungLeerDubletten: 'No print is here more than {behalten} times.',
  ordnungLeerWuensche:
    'Nothing on the wishlist. A copy gets there as “Wanted, not owned” — or from a deck via “Add what’s missing to the wishlist”.',
  ordnungStueck: '{stueck} copies in {zeilen} rows',
  ordnungOhneOrt: 'unspecified',
  ordnungUeberzaehlig: '{ueberzaehlig} too many',

  // **Besitze ich diese Karte?** (#1292).
  besitzJa: 'You own {stueck} of these.',
  besitzNein: 'You do not own this printing.',
  besitzGewuenscht: '({gewuenscht} on the wishlist)',

  // **Die Massenvorgänge** (#1292).
  masseTitel: 'Bulk actions',
  masseWaehlen: 'Select',
  masseGewaehlt: '{gewaehlt} of {gesamt} selected',
  masseLagerortSetzen: 'Set storage location to',
  masseZustandSetzen: 'Set condition to',
  masseAendern: 'Change selected',
  masseLoeschen: 'Delete selected',
  masseLoeschenFrage: 'Delete {anzahl} copies?',
  masseLoeschenErklaerung:
    'What goes here you type again: quantity, condition, storage location and purchase price are nowhere else. The collection has no history like a set.',
  masseLoeschenJa: 'Delete',
  masseBericht: '{geaendert} changed.',
  masseZuruecknehmen: 'Undo',
  masseZuruecknehmenVon: 'Undo the storage location of {anzahl} lines',
  masseZurueckgenommen: '{anzahl} undone.',
  massePreise: 'Fetch prices for all',
  massePreisschub: '{geholt} fetched, {offen} remaining',
  massePreiseFehlgeschlagen: '{anzahl} failed',
  massePreiseSteht:
    'This batch moved nothing — another attempt would give the same.',
  massePreiseFertig: 'All prices are in.',

  aufnehmen: 'Add',
  aenderungSichern: 'Save',
  loeschen: 'Delete',
  loeschenFrage: 'Delete this copy?',
  loeschenErklaerung:
    'Quantity, condition, storage location and purchase price are nowhere else — the collection has no history like a set.',
  loeschenJa: 'Delete',
  abbrechen: 'Cancel',
  bearbeiten: 'Edit',
  zumTausch: 'To trade',
  insDeck: 'To deck',
  auswahlTitel: 'Selection',
  auswahlInsDeck: '{anzahl} → deck',

  neuesTitel: 'Add a copy',
  weitereAngaben: 'More details',

  dubletteFrage:
    '{bezeichnung} is already in the collection — {menge} in the same condition, language and finish.',
  dubletteErklaerung:
    'Increasing adds the new quantity; the other details of the existing row stay. Cancel changes nothing — your entry stays.',
  dubletteErhoehen: 'Increase to {summe}',
  dubletteZurZeile: 'Go to the existing row',
  dubletteAbbrechen: 'Cancel',

  zustaende: {
    mint: 'Mint',
    nearMint: 'Near Mint',
    excellent: 'Excellent',
    good: 'Good',
    lightPlayed: 'Light Played',
    played: 'Played',
    poor: 'Poor',
  },

  preiseTitel: 'Prices',
  preiseHolen: 'Fetch prices',
  marktUndVerlauf: 'Offers and history',
  // **The marketplace** — work item A4 of phase 5.3 (#1443).
  marktTitel: 'Marketplace offers',
  marktErklaerung:
    'What individual sellers are asking for this printing — price, quantity, condition and language. The lookup goes out to the network. Nothing is bought.',
  marktHolen: 'Fetch offers',
  marktHolenBei: 'Fetch offers from {quelle}',
  marktLaeuft: 'Fetching …',
  marktLeer: 'Nothing is on sale there for this printing right now.',
  marktZugeordnet:
    'This printing is now linked to card version {fassung} — the next lookup will find it straight away.',
  marktStueck: '{menge} available',
  marktBuendel: 'in bundles of {buendel}',
  marktImUrlaub: 'not shipping right now',
  marktGradiert: 'graded',
  marktSigniert: 'signed',
  marktVeraendert: 'altered',
  marktBefunde: 'Not every offer could be read:',

  // **Die Begleitbefunde des Preisholens** (#1202).
  preisbefunde: 'Not every price could be read. These series are missing:',
  preiseLaufen: 'Fetching …',
  keinePreise: 'No price on record for this printing.',
  preisAlter: '{stunden} h ago · {quelle}',
  preisFrisch: 'just now · {quelle}',
  preisVeraltet: 'stale, {tage} days old · {quelle}',
  preisQuellstand: 'source as of {tag}',

  vorbehalt:
    'Prices are estimates from Scryfall and its affiliates. No guarantee is made for any of them; the price at the store is the one that counts.',

  verlauf: 'History',
  keinVerlauf:
    'Only one snapshot for this series — a history builds up as you fetch prices again.',

  bulkTitel: 'Fetch prices from a whole source',
  leiste: {
    bezeichnung: 'Collection figures',
    exemplare: 'Copies',
    eintraege: '{zahl} entries',
    wert: 'Value',
    keinPreis: 'no price',
    ohnePreis: 'Without price',
    nichtGelesen: 'not read',
  },
  drucksuche: {
    karte: 'Card',
    hinweis:
      'Name, optionally with set and number — e.g. Lightning Bolt (M11) 149',
    suchen: 'Search',
    drucke: 'Printings',
    keinTreffer: 'No card of this name found.',
    weitere: 'There are more printings — set and number narrow the choice.',
    treffer: '{name} · {set} ({code}) · No. {nummer}',
  },
  bulkErklaerung:
    'MTGJSON does not answer per card but with a file. “Today” brings the current day, “Ninety days” the history — and over a gigabyte.',
  bulkQuelle: 'Source',
  bulkUmfang: 'Scope',
  bulkHeute: 'Today',
  bulkNeunzigTage: 'Ninety days',
  bulkHolen: 'Fetch',
  bulkHolenBei: 'Fetch prices from {quelle}',
  bulkLaeuft: 'Fetching …',
  bulkBilanz: '{staende} prices for {drucke} printings stored.',
  bulkOhneDrucke:
    'No printing in this collection carries an MTGJSON identifier — nothing was requested.',
  uebersichtTitel: 'Overview',
  uebersichtLeer: 'No prices fetched yet.',
  ohnePreis: 'without a price: {zahl}',
  // **Ein Exemplar und nicht „1 Exemplare“.** Eine Zahl mit falscher
  // Form liest sich wie ein Fehler im Programm — gesehen im laufenden
  // Fenster.
  einExemplar: '{zahl} copy',
  exemplare: '{zahl} copies',

  keineUmrechnung:
    'One sum per currency. No conversion — that would need an exchange rate, and this program does not have one.',
  // **A price history from another source** — A5, #1444.
  verlaufTitel: 'Price history from another source',
  verlaufErklaerung:
    'Fetches a series over time and stores it. The lookup goes out to the network and uses quota; the condition is shown the way the source names it.',
  verlaufHolen: 'Fetch history',
  verlaufHolenBei: 'Fetch history from {quelle}',
  verlaufLaeuft: 'Fetching …',
  verlaufBilanz: '{staende} price points stored — card version {fassung}.',
} as const
