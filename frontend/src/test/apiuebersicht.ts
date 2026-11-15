// **Die Vorbelegungen der Übersicht und des Dienstkatalogs** — ausgelagert aus
// `api.ts`, weil die Datei mit ihnen über 400 Zeilen ging (AGENTS.md §2).
// Was hier steht, ist dieselbe Zusage wie dort: eine Fassade, die eine
// Seite beim Aufschlagen fragt, antwortet mit dem ehrlichen leeren Fall.

// ohneWeiterarbeiten ist die Vorbelegung für `Weiterarbeiten` — Werkstück U11
// von Phase 5.1 (#1422).
//
// **Aus demselben Grund wie [ohneZugaenge]**: die Startseite fragt sie beim
// Aufschlagen. Leer heißt: nichts zuletzt, nichts angeheftet — der ehrliche
// Fall eines frischen Bestands.
export function ohneWeiterarbeiten() {
  return Promise.resolve({
    zuletzt: null,
    geoeffnet: [],
    eingefuehrt: [],
    ausgegeben: [],
    angeheftet: [],
    fehler: null,
  })
}

// ohneAufmerksamkeit und ohneAuftraege sind die Vorbelegungen der Übersicht
// (#1422): nichts braucht Aufmerksamkeit, nichts läuft — der ehrliche Fall
// eines frischen Bestands, und die Startseite fragt beide beim Aufschlagen.
export function ohneAufmerksamkeit() {
  return Promise.resolve({
    hinweise: [],
    durchgesehen: { sets: 0, decks: 0, pakete: 0 },
    befunde: [],
    fehler: null,
  })
}

export function ohneAuftraege() {
  return Promise.resolve({ auftraege: [], laufend: 0, fehler: null })
}

// ohneBetriebsband ist die Vorbelegung des Betriebsbands (OPS02, #2348):
// nichts meldet sich, und alles ist gemessen — „Was Aufmerksamkeit braucht"
// fragt es beim Aufschlagen.
export function ohneBetriebsband() {
  return Promise.resolve({
    hinweise: [],
    ungemessen: [],
    aufwandMikros: 0,
    fehler: null,
  })
}

// uebersichtsvorgaben sind die Wege, die die Startseite beim Aufschlagen holt.
export const uebersichtsvorgaben = {
  Weiterarbeiten: ohneWeiterarbeiten,
  Aufmerksamkeit: ohneAufmerksamkeit,
  Auftraege: ohneAuftraege,
  Betriebsband: ohneBetriebsband,
}

// stillerVermerk ist die Vorbelegung für `DeckBenutzt`: die Deckseite ruft ihn
// beim Aufschlagen eines Decks, und ein Test, der Decks meint, soll nicht
// über den Vermerk stolpern.
export function stillerVermerk() {
  return Promise.resolve({ fehler: null })
}

// ohneDienstkatalog ist die Vorbelegung für `Dienstkatalog` — Werkstück A1
// von Phase 5.3 (#1440).
//
// **Aus demselben Grund wie [ohneZugaenge]**: die Einstellungsseite fragt ihn
// beim Aufschlagen des Dienstegebiets. Leer heißt: das Gebiet zeigt keinen
// Katalog — ein Test, der ihn meint, belegt ihn selbst.
export function ohneDienstkatalog() {
  return Promise.resolve({ eintraege: [], fehler: null })
}

// ohneDeckbezuege ist die Vorbelegung für `ExemplarDecks` — Werkstück U9 von
// Phase 5.1 (#1420).
//
// **Aus demselben Grund wie [stillerVermerk]**: der Inspector der Sammlung
// fragt bei jedem gewählten Exemplar nach seinen Decks, und ein Test, der die
// Sammlung meint, soll nicht darüber stolpern. Leer ist der Normalfall — ein
// Exemplar gehört keinem Deck, bis jemand es zuordnet.
export function ohneDeckbezuege() {
  return Promise.resolve({ decks: [], fehler: null })
}
