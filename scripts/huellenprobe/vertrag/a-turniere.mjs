// **Turniermatch: binden, melden, aus der Partie melden, übernehmen, aussteigen**
// — eine Gruppe des Vertragsrundgangs (B01, #2220).
//
// **Ein eigenes Turnier mit vier Teilnehmenden und zwei Paarungen.** Das
// Turnier der Gruppe `turniere.mjs` taugt nicht: seine einzige Paarung hat ein
// Ergebnis (`TurnierErgebnisSetzen`), und eine Paarung mit Ergebnis lässt sich
// nicht mehr binden (`bindbar`, `internal/service/matchleitung.go`).
//
//  - Paarung 1 geht **von Hand**: binden, beide Seiten melden dieselbe Kette
//    und denselben Stand, übernehmen.
//  - Paarung 2 geht **aus der Partie**: binden, in der Partie aufgeben (das
//    beendet sie mit einem Sieger, CR 104.3a), beide Seiten melden daraus,
//    übernehmen.
//
// **Gebunden wird die zweite Paarung erst nach der Übernahme der ersten.** Eine
// Bindung trägt die Fassung des Turniers, und jedes Eintragen eines Ergebnisses
// hebt sie: eine früher gebundene Paarung bekäme bei der Übernahme einen
// Konflikt („die Paarung hat sich inzwischen geändert").
//
// **Eine beendete Partie bleibt** (`domain.Loeschbar`): `PartieLoeschen` weist
// sie ab, sie räumt erst die Datenpflege nach ihrer Frist. Die erste Partie
// läuft nie zu Ende und wird gelöscht.
//
// Braucht `k.set` aus der ersten Gruppe (nur für den Spielnamen).

// teilnehmer nennt, wer sich anmeldet und unter welchem Schlüssel in `k` seine Kennung liegt.
const teilnehmer = [['Anna', 'aT1'], ['Bert', 'aT2'], ['Carla', 'aT3'], ['Dirk', 'aT4']]

// sitz ordnet die beiden Teilnehmenden einer Paarung den Sitzen einer Partie zu.
const sitze = (paarung) => [
  { teilnehmerID: paarung.a, sitz: 'anna', peer: '', deckabdruck: '' },
  { teilnehmerID: paarung.b, sitz: 'bert', peer: '', deckabdruck: '' },
]

// partie ist der Entwurf einer Partie mit zwei Sitzen und demselben Deck. Das
// Regelwerk bleibt leer: es gilt der Sandkasten in Fassung 1 (`regelwerkAus`).
const partie = (k) => ({
  spiel: k.set.spiel,
  regelwerk: { kennung: '', fassung: 0, strenge: '', abgewaehlt: [] },
  wirt: 'anna',
  sitze: ['anna', 'bert'].map((kennung) => ({
    kennung, name: kennung === 'anna' ? 'Anna' : 'Bert', deckID: k.aMdeck, commander: '', leiheID: '', profilID: '',
  })),
})

// Der Stand einer Meldung von Hand. Zwei zu eins ist ein Match über drei
// Spiele — das Turnier hier ist keines im Einzelspiel.
const stand = { spieleA: 2, spieleB: 1, unentschieden: 0 }

export const aTurniere = {
  name: 'Turniermatch (A)',
  rufe: [
    // --- Ein Deck, mit dem sich eine Partie aufbauen lässt: Druckkennungen
    // ohne Netz (`internal/app/partie_intern_test.go`, `einDeckImBestand`).
    {
      name: 'DeckAnlegen', argumente: (k) => ['Vertragsdeck Match', k.set.spiel, 'aus dem Rundgang'],
      nimm: 'a.deck.id', nach: (w, k) => { k.aMdeck = w },
    },
    // **Eigene Karten statt Drucke**: ein Druck ohne Scryfall gibt keinen
    // Deckabzug mit Karten, und die Partie bräuchte einen (zweiter Rundgang).
    { name: 'DeckplatzLegen', argumente: (k) => [k.aMdeck, k.karten[0], '', 8, 'hauptdeck'] },
    { name: 'DeckplatzLegen', argumente: (k) => [k.aMdeck, k.karten[1], '', 2, 'hauptdeck'] },

    // --- Ein Turnier ohne Einzelspiel, vier Teilnehmende, eine Runde
    {
      name: 'TurnierAnlegen', argumente: (k) => ['Vertragsmatch', k.set.spiel, 'aus dem Rundgang', false, ''],
      nimm: 'a.turnier.id', nach: (w, k) => { k.aMturnier = w },
    },
    ...teilnehmer.map(([wer, ablage]) => ({
      name: 'TurnierAnmelden', argumente: (k) => [k.aMturnier, wer, k.aMdeck],
      nimm: 'a.teilnehmer.id', nach: (w, k) => { k[ablage] = w },
    })),
    {
      name: 'TurnierRundeLegen', argumente: (k) => [k.aMturnier],
      nimm: 'a.runde.paarungen.map((x) => ({ id: x.id, a: x.a, b: x.b }))', nach: (w, k) => { k.aMpaarungen = w },
    },

    // --- Zwei Partien an einem Tisch: nur die erste wird je binden und löschen
    {
      name: 'PartieAnlegen', argumente: (k) => [partie(k)],
      nimm: 'a.partie.id', nach: (w, k) => { k.aMpartie1 = w },
    },
    {
      name: 'PartieAnlegen', argumente: (k) => [partie(k)],
      nimm: 'a.partie.id', nach: (w, k) => { k.aMpartie2 = w },
    },

    // --- Paarung 1: von Hand
    {
      name: 'TurnierMatchBinden',
      argumente: (k) => [k.aMturnier, k.aMpaarungen[0].id, k.aMpartie1, sitze(k.aMpaarungen[0])],
    },
    { name: 'TurnierMatchMelden', argumente: (k) => [k.aMpaarungen[0].id, 'anna', 'vertrag-kette-eins', stand] },
    { name: 'TurnierMatchMelden', argumente: (k) => [k.aMpaarungen[0].id, 'bert', 'vertrag-kette-eins', stand] },
    { name: 'TurnierMatchUebernehmen', argumente: (k) => [k.aMturnier, k.aMpaarungen[0].id] },

    // --- Paarung 2: aus der Partie
    {
      name: 'TurnierMatchBinden',
      argumente: (k) => [k.aMturnier, k.aMpaarungen[1].id, k.aMpartie2, sitze(k.aMpaarungen[1])],
    },
    {
      // **Aufgeben geht jederzeit** (`zug.aufgeben`, CR 104.3a): Anna verlässt
      // die Partie, Bert bleibt und gewinnt. Die Befehlskennung ist eigene Sache
      // des Aufrufers (§7, Regel 4); `erwartet` leer heißt „gegen den Stand,
      // der gerade gilt".
      name: 'PartieBefehl',
      argumente: (k) => [k.aMpartie2, 'vertrag-aufgabe-1', 'anna', 'zug.aufgeben', {}, ''],
    },
    { name: 'TurnierMatchAusDerPartie', argumente: (k) => [k.aMpaarungen[1].id, 'anna'] },
    { name: 'TurnierMatchAusDerPartie', argumente: (k) => [k.aMpaarungen[1].id, 'bert'] },
    { name: 'TurnierMatchUebernehmen', argumente: (k) => [k.aMturnier, k.aMpaarungen[1].id] },

    // --- Aussteigen und wieder einsteigen (ein offenes Turnier erlaubt beides)
    { name: 'TurnierAussteigen', argumente: (k) => [k.aMturnier, k.aT4, true] },
    { name: 'TurnierAussteigen', argumente: (k) => [k.aMturnier, k.aT4, false] },

    // --- Ans Ende, was wegnimmt. Die Bindungen gehen mit dem Turnier
    // (`ON DELETE CASCADE`, Migration 0067).
    { name: 'TurnierLoeschen', argumente: (k) => [k.aMturnier] },
    { name: 'PartieLoeschen', argumente: (k) => [k.aMpartie1] },
    { name: 'DeckLoeschen', argumente: (k) => [k.aMdeck] },
  ],
}
