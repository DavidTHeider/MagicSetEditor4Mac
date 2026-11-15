// **Decks und Sammlung** — eine Gruppe des Vertragsrundgangs (B01, #2220).
//
// **Druckkennungen sind Scryfall-UUIDs** (`domain.druckkennung`); die hier
// sind Testwerte dieser Gestalt, kein echter Druck.
//
// Braucht `k.set` und `k.karten` aus der ersten Gruppe; legt `k.deck`,
// `k.platz` und `k.exemplar` an. Gelöscht wird hier nur, was diese Gruppe
// selbst angelegt hat.

const exemplar = {
  id: 0, fassung: 0, printing: '00000000-0000-4000-8000-000000000001', druckID: '', vorlaeufig: false, menge: 1,
  zustand: 'nearMint', sprache: 'de', finish: 'nonfoil', erwerbsdatum: '2026-10-01',
  erwerbspreis: '1.00', erwerbswaehrung: 'EUR', notiz: '', name: 'Vertragskarte',
  setcode: 'vtr', nummer: '1', lagerort: 'Kiste 1', wunsch: false, bezeichnung: '',
}

export const decks = {
  name: 'Decks und Sammlung',
  rufe: [
    { name: 'Decks', argumente: (k) => [k.set.spiel] },
    { name: 'Formate', argumente: (k) => [k.set.spiel], nimm: 'a.formate.map((f) => f.id)', nach: (w, k) => { k.format = w[0] ?? '' } },
    { name: 'Bauarten' },
    {
      name: 'DeckAnlegen', argumente: (k) => ['Vertragsdeck', k.set.spiel, 'aus dem Rundgang'],
      nimm: 'a.deck.id', nach: (w, k) => { k.deck = w },
    },
    { name: 'DeckplatzLegen', argumente: (k) => [k.deck, k.karten[0], '', 2, 'hauptdeck'], nimm: 'a.platz.id', nach: (w, k) => { k.platz = w } },
    // **Ein Verschieben gibt einen neuen Platz** — die Kennung läuft mit.
    { name: 'DeckplatzVerschieben', argumente: (k) => [k.deck, k.platz, 'sideboard'], nimm: 'a.platz.id', nach: (w, k) => { k.platz = w } },
    { name: 'DeckplatzVerschieben', argumente: (k) => [k.deck, k.platz, 'hauptdeck'], nimm: 'a.platz.id', nach: (w, k) => { k.platz = w } },
    {
      name: 'FormatSetzen',
      argumente: (k) => [{ id: '', fassung: '', name: 'Vertragsformat', spiel: k.set.spiel, mindestgroesse: 1, hoechstgroesse: 0, sideboardgroesse: 15, hoechstzahlGleich: 4, sideboardOffen: false, scryfallname: '' }],
      nimm: 'a.format.id', nach: (w, k) => { k.format = w },
    },
    { name: 'DeckFormatWaehlen', argumente: (k) => [k.deck, k.format] },
    { name: 'DeckPruefen', argumente: (k) => [k.deck] },
    { name: 'DeckFehlkarten', argumente: (k) => [k.deck] },
    { name: 'DeckAbgleichen', argumente: (k) => [k.deck] },
    { name: 'DecklisteAusgeben', argumente: (k) => [k.deck] },
    { name: 'DeckKopieren', argumente: (k) => [k.deck, 'Vertragsdeck Kopie'], nimm: 'a.deck.id', nach: (w, k) => { k.deckKopie = w } },
    { name: 'Deck', argumente: (k) => [k.deck], nimm: 'a.deck.fassung', nach: (w, k) => { k.deckfassung = w } },
    { name: 'DeckUmbenennen', argumente: (k) => [k.deck, k.deckfassung, 'Vertragsdeck neu', 'umbenannt'] },
    { name: 'Sammlung' },
    { name: 'ExemplarAnlegen', argumente: () => [exemplar], nimm: 'a.id', nach: (w, k) => { k.exemplar = w } },
    { name: 'ExemplarDecks', argumente: (k) => [String(k.exemplar)] },
    // **Zugeordnet wird ein Exemplar einem Druckplatz** — ein Platz mit einer
    // eigenen Karte hat keinen Druck, dem ein Exemplar entspräche.
    { name: 'DeckplatzLegen', argumente: (k) => [k.deck, '', '00000000-0000-4000-8000-000000000001', 1, 'hauptdeck'], nimm: 'a.platz.id', nach: (w, k) => { k.druckplatz = w } },
    { name: 'DeckZuordnen', argumente: (k) => [k.deck, k.druckplatz, String(k.exemplar), 1], nimm: 'a.zuordnung.id', nach: (w, k) => { k.zuordnung = w } },
    { name: 'DeckZuordnungen', argumente: (k) => [k.deck] },
    { name: 'DeckZuordnungLoesen', argumente: (k) => [k.deck, k.zuordnung] },
    { name: 'Portfolio' },
    { name: 'Preise', argumente: () => ['00000000-0000-4000-8000-000000000001'] },
    { name: 'DeckFehlkartenWuenschen', argumente: (k) => [k.deck] },
    { name: 'DeckplatzNehmen', argumente: (k) => [k.deck, k.platz] },
    { name: 'DeckLoeschen', argumente: (k) => [k.deckKopie] },
  ],
}
