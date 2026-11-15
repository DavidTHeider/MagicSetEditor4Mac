// **Sammlung, Massenvorgänge, Listeneinfuhr, Formate und eine Decklisteneinfuhr**
// — eine Gruppe des Vertragsrundgangs (B01, #2220).
//
// Die Gruppe legt **eigene Exemplare** an (Druckkennungen `vertrag-a-…`), eigene
// Formate und ein eigenes Deck und löscht alles am Ende wieder. Das Exemplar
// `vertrag-001` der Deckgruppe (`decks.mjs`) und deren Deck bleiben unberührt.
//
// **Die Fassung gehört in jeden Ruf, der ein Exemplar ändert oder löscht**
// (§7, Regel 2): sie kommt aus `Sammlung`, und nach jedem Schreiben wird neu
// gelesen, denn jedes Schreiben hebt sie um eins.
//
// Braucht `k.set` aus der ersten Gruppe (nur für den Spielnamen).
import { createHash } from 'node:crypto'

// exemplar baut ein Exemplar, das sich von allen anderen im Druck unterscheidet:
// der Bestand erlaubt je Druck, Zustand, Sprache und Ausführung nur eine Zeile
// und gäbe sonst eine Dublette zurück.
const exemplar = (printing, name, nummer) => ({
  id: 0, fassung: 0, printing, druckID: '', vorlaeufig: false, menge: 1,
  zustand: 'nearMint', sprache: 'de', finish: 'nonfoil', erwerbsdatum: '2026-10-01',
  erwerbspreis: '1.00', erwerbswaehrung: 'EUR', notiz: '', name,
  setcode: 'vta', nummer, lagerort: 'Kiste A', wunsch: false, bezeichnung: '',
})

const a = '00000000-0000-4000-8000-0000000000a1'
const b = '00000000-0000-4000-8000-0000000000a2'
const c = '00000000-0000-4000-8000-0000000000a3'
const liste = 'vertrag-a-liste'

// lesen holt die Exemplare dieser Gruppe samt Fassung und legt sie nach Druck ab.
const lesen = {
  name: 'Sammlung',
  // **Nach den eigenen Kennungen** — die Drucke sind UUIDs (`domain.druckkennung`).
  nimm: `a.exemplare.filter((e) => ${JSON.stringify([a, b, c, liste])}.includes(e.printing))`,
  nach: (w, k) => { k.aExemplare = Object.fromEntries(w.map((e) => [e.printing, e])) },
}

// zeile nennt ein Exemplar so, wie die Massenvorgänge es verlangen.
const zeile = (k, printing) => ({ id: k.aExemplare[printing].id, fassung: k.aExemplare[printing].fassung })

// stand rechnet den Abdruck eines **leeren** Decks, wie `domain.Deckstand` ihn
// bildet: SHA-256 über `<Fassung>\n` und die Plätze — ohne Plätze bleibt nur
// das. Gelesen am 2026-10-09 in `internal/domain/deckstand.go`. Die Vorschau
// (`DecklisteVorschau`) gibt ihn sonst zurück, braucht aber Scryfall für jede
// Zeile ihres Textes und weist einen leeren Text ab.
const stand = (fassung) => createHash('sha256').update(`${fassung}\n`).digest('hex')

export const aSammlung = {
  name: 'Sammlung, Formate, Decklisten (A)',
  rufe: [
    // --- Exemplare
    { name: 'ExemplarAnlegen', argumente: () => [exemplar(a, 'Vertragskarte A', '1')], nimm: 'a.id' },
    { name: 'ExemplarAnlegen', argumente: () => [exemplar(b, 'Vertragskarte B', '2')], nimm: 'a.id' },
    { name: 'ExemplarAnlegen', argumente: () => [exemplar(c, 'Vertragskarte C', '3')], nimm: 'a.id' },
    lesen,
    // **Das gelesene Exemplar zurück, mit einer Änderung** — so schickt auch
    // das Fenster es, samt Fassung.
    {
      name: 'ExemplarAendern',
      argumente: (k) => [{ ...k.aExemplare[a], menge: 2, notiz: 'aus dem Rundgang geändert' }],
    },
    // Zwei Zeilen auf einmal. Ein `null` heißt „unverändert" (Zeiger im DTO).
    {
      name: 'MasseAendern',
      argumente: (k) => [{ zeilen: [zeile(k, b), zeile(k, c)], menge: 3, zustand: null, lagerort: 'Kiste B', wunsch: null }],
    },

    // --- Listeneinfuhr: entschiedene Zeilen, ohne Scryfall (die Deutung braucht es, das Einlesen nicht)
    {
      name: 'ListeEinlesen',
      argumente: () => [[{ printing: liste, menge: 2, zustand: 'nearMint', sprache: 'de', finish: 'nonfoil', name: 'Vertragsliste', setcode: 'vta', nummer: '9' }], 'addieren'],
      nimm: 'a.bericht.angelegt',
    },
    lesen,

    // --- Formate: aus einer Bauart, danach gleich wieder weg (unbenutzt, also löschbar)
    {
      name: 'FormatAusBauart', argumente: (k) => ['konstruiert', 'Vertragsformat A', k.set.spiel, ''],
      nimm: 'a.format.id', nach: (w, k) => { k.aFormat = w },
    },
    { name: 'FormatLoeschen', argumente: (k) => [k.aFormat] },

    // --- Decklisteneinfuhr in ein eigenes, noch leeres Deck
    {
      name: 'DeckAnlegen', argumente: (k) => ['Vertragsdeck A', k.set.spiel, 'aus dem Rundgang'],
      nimm: 'a.deck.id', nach: (w, k) => { k.aDeck = w },
    },
    {
      name: 'Deck', argumente: (k) => [k.aDeck],
      nimm: 'a.deck.fassung', nach: (w, k) => { k.aDeckfassung = w },
    },
    {
      // `deckID` ist eine Zahl, kein Text (`DecklisteUebernehmen(deckID int64, …)`).
      name: 'DecklisteUebernehmen',
      argumente: (k) => [Number(k.aDeck), stand(k.aDeckfassung), [{ printing: '00000000-0000-4000-8000-0000000000a4', menge: 2, bereich: 'hauptdeck' }]],
    },

    // --- Was in einer Einzelanwendung nicht geht
    {
      // `erlaubnisPruefen` weist das eigene Profil ab („aus dem eigenen Profil
      // lässt sich nichts übernehmen", `internal/service/uebernahme.go`, Zeile
      // ~190). Es gibt ein Profil; ein zweites müsste mit `ProfilAnlegen`
      // entstehen und ein Deck darin mit einem `ProfilWechseln` — ein Wechsel,
      // der die Sitzung des Rundgangs umstellt.
      name: 'DeckUebernehmen', argumente: () => ['1', '1'], erwartet: 'fachfehler',
      grund: 'aus dem eigenen Profil lässt sich nichts übernehmen; es gibt nur ein Profil (uebernahme.go, erlaubnisPruefen)',
    },
    {
      // Ohne Zugangsschlüssel für Scrydex endet der Ruf **vor** dem Netz:
      // `scrydexpreiseUeber` fragt den Schlüsselbund und gibt `ZugangFehlt`.
      // Mit einem Schlüssel ginge er zu Scrydex hinaus.
      name: 'Fremdpreisverlauf', argumente: () => ['scrydex', a, 'nonfoil', []], erwartet: 'fachfehler',
      grund: 'braucht ein Scrydex-Konto im Schlüsselbund und das Netz (scrydexpreisnaht.go, scrydexpreiseUeber)',
    },

    // --- Ans Ende, was wegnimmt
    lesen,
    {
      name: 'ExemplarLoeschen',
      argumente: (k) => [k.aExemplare[a].id, k.aExemplare[a].fassung],
    },
    {
      name: 'MasseLoeschen',
      argumente: (k) => [[zeile(k, b), zeile(k, c), zeile(k, liste)]],
    },
    { name: 'DeckLoeschen', argumente: (k) => [k.aDeck] },
  ],
}
