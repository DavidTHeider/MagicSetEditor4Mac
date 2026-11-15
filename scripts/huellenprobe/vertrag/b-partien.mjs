// **Partien** — eine Gruppe des Vertragsrundgangs (B01, #2220).
//
// Braucht `k.set` und `k.karten` aus der ersten Gruppe. Sie baut sich **ein
// eigenes Deck** (zwei Karten) und eine Partie mit zwei Sitzen, `ada` und
// `bo`, am selben Fenster — der Aufbau, den die Kommandozeile mit
// `mse partie-anlegen probe ada:Ada:<deck> bo:Bo:<deck>` kennt und den
// `scripts/huellenprobe/nat04.mjs` benutzt.
//
// # Welcher Befehl, und warum nicht die Starthand
//
// `spieler.starthand` zieht sieben Karten. Ein Deck mit zwei Karten hat sie
// nicht — wer aus einer leeren Bibliothek zieht, verliert (CR 704.5b), und die
// Partie ist beim **nächsten** Befehl **beendet**; die Wege danach
// (Pausieren, Freigeben) gingen ins Leere. An der Kommandozeile geprüft
// (2026-10-09): erst `starthand`, dann `muenzwurf` — danach „beendet".
// `spieler.muenzwurf` allein zieht nichts und genügt, den Weg `PartieBefehl`
// zu gehen.
//
// # Reden und Zeigen ohne Gast
//
// `PartieSagen` und `PartieZeigen` gehen an die Gäste einer **freigegebenen**
// Partie; ohne Gast gehen sie an niemanden und sind trotzdem gelungen
// (`internal/service/partierunde.go`, `anJeden` über keinen Platz). Deshalb
// steht die Freigabe davor — und am Ende `PartieFreigabeBeenden`. Die
// Freigabe selbst öffnet keinen Port (`internal/app/partiefreigabe.go`): sie
// merkt nur, welche Partie offensteht.
//
// # Was eine Einzelanwendung nicht kann
//
// - `PartieBeitreten`: setzt sich an den Tisch eines Gegenübers
//   (`internal/app/partiebeitritt.go:122`, `Oeffne`).
// - `PartieWuenschen`: gilt einem Gast an einem fremden Tisch
//   (`internal/app/partiebeitritt.go:230`, `gastAmTisch`).
//
// # Wie weit das geprüft ist
//
// Die Abfolge ist gegen den App-Bau gefahren worden (`app.NeueBindungen` über
// einem frisch gestarteten `app.NeueApp`, Wegwerfprobe vom 2026-10-09, nicht
// eingecheckt), die Argumentverbünde strikt gegen die Go-Typen und die
// `nimm`-Ausdrücke gegen die serialisierten Antworten. **Nicht geprüft** ist,
// was nur das native Fenster hat: die Wails-Naht selbst, die echten
// Dateidialoge (dort stand ein eingesetzter Dateiwähler) und jeder
// Unterprozess.

import { b, fremder } from './hilfen-b.mjs'

const partie = (k) => b(k).partie
const deck = (k) => b(k).partiedeck

const sitz = (k, kennung, name) => ({ kennung, name, deckID: String(deck(k)), commander: '', leiheID: '', profilID: '' })

export const bPartien = {
  name: 'Partien',
  rufe: [
    // --- ein eigenes Deck, damit nichts eines anderen Eintrags verändert wird ---
    {
      name: 'DeckAnlegen', argumente: (k) => ['Vertragspartie-Deck', k.set.spiel, 'aus dem Rundgang'],
      nimm: 'a.deck.id', nach: (w, k) => { b(k).partiedeck = w },
    },
    { name: 'DeckplatzLegen', argumente: (k) => [deck(k), k.karten[0], '', 2, 'hauptdeck'] },

    // --- die Partie ---
    {
      name: 'PartieAnlegen',
      argumente: (k) => [{ spiel: k.set.spiel, regelwerk: {}, wirt: '', sitze: [sitz(k, 'ada', 'Ada'), sitz(k, 'bo', 'Bo')] }],
      nimm: 'a.partie.id', nach: (w, k) => { b(k).partie = w },
    },
    { name: 'PartieBefehl', argumente: (k) => [partie(k), 'vertrag-befehl-1', 'ada', 'spieler.muenzwurf', {}, ''] },
    { name: 'PartiePausieren', argumente: (k) => [partie(k), 'ada'] },
    { name: 'PartieFortsetzen', argumente: (k) => [partie(k), 'ada'] },

    // **Eine Einwilligung** gilt einem Sitz derselben Partie; die Liste
    // ersetzt die alte, und zurückgenommen wird je Sitz.
    { name: 'PartieOffenlegen', argumente: (k) => [partie(k), 'ada', ['bo']] },
    { name: 'PartieOffenlegungZuruecknehmen', argumente: (k) => [partie(k), 'ada'] },

    // **Eine Kopie ist ein Übungszustand** und läuft, auch aus einer beendeten
    // heraus; gelöscht werden darf, was nicht beendet und nicht eingewilligt
    // ist (`internal/domain/partieverwaltung.go`, `Loeschbar`).
    { name: 'PartieKopieren', argumente: (k) => [partie(k)], nimm: 'a.kennung', nach: (w, k) => { b(k).partiekopie = w } },
    { name: 'PartieLoeschen', argumente: (k) => [b(k).partiekopie] },

    // --- am Tisch ---
    { name: 'PartieFreigeben', argumente: (k) => [partie(k), 'ada', ['bo']] },
    { name: 'PartieSagen', argumente: () => ['Hallo vom Rundgang'] },
    // Ein Zeiger auf freier Fläche: `objekt` darf leer sein, die Stelle sind
    // Millimeter auf dem Tisch.
    { name: 'PartieZeigen', argumente: () => ['zeiger', '', 10, 20] },
    { name: 'PartieFreigabeBeenden' },

    // **Nicht gelingbar**: siehe den Kopf. Der Fremde ist kein Kontakt, die
    // Absage kommt sofort und nicht nach einer Wartezeit.
    {
      // **Seit #2377 bringt ein Gast Deck oder Leihe mit** — leer: ohne.
      name: 'PartieBeitreten', argumente: (k) => [fremder(k, 'tisch').kennung, partie(k), { deckID: '', leiheID: '' }], erwartet: 'fachfehler',
      grund: 'öffnet eine Leitung zu einem Gegenüber, das es hier nicht gibt (internal/app/partiebeitritt.go:122)',
    },
    {
      name: 'PartieWuenschen', argumente: () => ['vertrag-wunsch-1', 'spieler.muenzwurf', {}], erwartet: 'fachfehler',
      grund: 'ein Wunsch geht an den Wirt eines fremden Tisches, an dem dieses Programm nicht sitzt (internal/app/partiebeitritt.go:230)',
    },

    // --- zu Ende: beenden, vergessen, das Deck wegräumen ---
    { name: 'PartieBeenden', argumente: (k) => [partie(k), 'ada'] },
    { name: 'PartieVergessen', argumente: (k) => [partie(k)] },
    { name: 'DeckLoeschen', argumente: (k) => [deck(k)] },
  ],
}
