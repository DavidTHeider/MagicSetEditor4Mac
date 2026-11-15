// **Handel und Deckleihe** — eine Gruppe des Vertragsrundgangs (B01, #2220).
//
// Braucht `k.deck` aus der Gruppe der Decks (für die Leihe) und legt sonst
// alles selbst an. Das Gegenüber ist ein **erfundener Peer** (`hilfen-b.mjs`):
// er ist eine gültige Kennung und sonst nichts — niemand antwortet, und es
// geht nichts hinaus.
//
// # Ein Tausch, den eine Anwendung allein zu Ende bringt
//
// Der Weg bis `beidseitigVerbucht` braucht keine zweite Anwendung, solange
// die Stimme des Gegenübers **dieser Rechner einträgt**: `HandelSchritt` mit
// einer genannten Partei tut genau das (`internal/app/handelsleitung.go`,
// „Wer eine andere Partei nennt, trägt deren Stimme ein"). Der Tausch hier
// verlangt Karten **vom Gegenüber** (`von` ist seine Kennung) — damit gibt
// dieser Rechner nichts weg, und das Festhalten entfällt; beim Verbuchen
// kommt eine Zeile dazu. Die Freigabe für `verbuchen` und `bestaetigen`
// stellt `HandelFreigabeErteilen` aus — **nur am Fenster**, und genau dort
// läuft der Rundgang.
//
// # Was eine Einzelanwendung nicht kann
//
// - `HandelVorschlagen`: ein Gegenangebot darf nur die Gegenseite machen
//   (`internal/domain/handelsmatrix.go`, `gegenangebot` für die Gegenseite;
//   aus `entwurf` führt keines). Die Gegenseite ist ohne Leitung nicht da.
//   Am Kommandozeilenbau gegengeprüft: „aus »entwurf« führt kein
//   Gegenangebot".
// - `HandelZustellen`: es baut eine Leitung zu einem Gegenüber
//   (`internal/app/handelszustellung.go`, `Oeffne`) — ohne bestätigten
//   Kontakt, Zweck und Adresse gibt es keine.
// - `DeckleiheAnnehmen`: die Leihe muss an **dieses** Profil gerichtet sein
//   und von einem **anderen** unterschrieben (`internal/service/deckleihe.go`);
//   sich selbst zu verleihen ist keine Leihe. Was hier geprüft wird, ist die
//   Ablehnung einer Leihe, die einem Fremden gilt.
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

import { join } from 'node:path'

import { b, fremder } from './hilfen-b.mjs'

const gegenueber = (k) => fremder(k, 'handel').kennung
const ich = (k) => b(k).meineKennung
const h = (k, name) => b(k).handel[name]

// merke legt eine Antwort unter einem Namen in den Kontext.
const merke = (name) => (w, k) => { (b(k).handel ??= {})[name] = w }

// position ist eine Zeile eines Angebots — Karte, Menge und Güte wie in der
// Kommandozeile (`scripts/huellenprobe/trade01.mjs`).
const position = (von, aenderung = {}) => ({
  von, printing: 'bolt-m10', druckId: 0, menge: 1, guete: 'nearMint', sprache: 'de', finish: 'nonfoil',
  abbild: '', bewertung: 0, ...aenderung,
})

// wunsch eröffnet einen Tausch mit dem Gegenüber; die Kennung des Vorgangs
// vergibt das Programm.
const wunsch = (k, aenderung = {}) => ({
  vorgang: '', art: 'tausch', gegenueber: [gegenueber(k)], positionen: [position(gegenueber(k))], bewertungGeteilt: false,
  ...aenderung,
})

// schritt geht einen Schritt an einem Vorgang — `wer` leer ist dieser Rechner,
// `gesehen` leer heißt „nicht gefragt" (seit #2496 das vierte Argument).
const schritt = (name, was, wer = () => '') => ({
  name: 'HandelSchritt', argumente: (k) => [h(k, name), was, wer(k), ''],
})

// freigabe stellt den Nachweis aus; `bestaetigen` gilt je Partei.
const freigabe = (name, was, wer, ziel) => ({
  name: 'HandelFreigabeErteilen', argumente: (k) => [h(k, name), was, wer(k)], nimm: 'a.kennung', nach: merke(ziel),
})

const vorgang = 'a.vorgang.vorgang'

// exemplar ist die Besitzzeile, die der Tausch mit Festhalten weggäbe.
const exemplar = {
  id: 0, fassung: 0, printing: 'vertrag-h01', druckID: '', vorlaeufig: false, menge: 1,
  zustand: 'nearMint', sprache: 'de', finish: 'nonfoil', erwerbsdatum: '2026-10-01',
  erwerbspreis: '1.00', erwerbswaehrung: 'EUR', notiz: '', name: 'Vertragskarte Handel',
  setcode: 'vtr', nummer: '2', lagerort: 'Kiste 2', wunsch: false, bezeichnung: '',
}

const leihgegenueber = (k) => fremder(k, 'leihe').kennung
const leihpfad = (k, nummer) => join(k.ort.ort, 'vertrag', `b-leihe-${String(nummer)}.msepackage`)
const leihwunsch = (k, woran) => ({
  entleiher: leihgegenueber(k), deckID: String(k.deck), bindungsart: 'partie', woran, rechte: ['spielen'], stunden: 24, modus: 'zuende',
})

// neueLeihe nimmt die Leihe aus der Liste, die diese Gruppe eben ausgestellt
// hat: die an den eigenen Fremden, die noch nicht bekannt ist.
const neueLeihe = (name) => (w, k) => {
  const bekannt = (b(k).leihenBekannt ??= new Set())
  const neu = w.find((l) => l.verliehen && l.entleiher === leihgegenueber(k) && !bekannt.has(l.kennung))
  bekannt.add(neu.kennung)
  merke(name)(neu.kennung, k)
}

const leihenform = 'a.leihen.map((l) => ({ kennung: l.kennung, entleiher: l.entleiher, verliehen: l.verliehen }))'

export const bHandel = {
  name: 'Handel und Deckleihe',
  rufe: [
    { name: 'PeeridentitaetSicherstellen', nimm: 'a.kennung', nach: (w, k) => { b(k).meineKennung = w } },

    // --- der Tausch, bis er beidseitig verbucht ist ---
    { name: 'HandelEroeffnen', argumente: (k) => [wunsch(k)], nimm: vorgang, nach: merke('h1') },
    schritt('h1', 'anbieten'),
    // Aus `angeboten` geht eine Nachricht hinaus; sie liegt im Ausgang, bis
    // die Gegenseite sie bestätigt — hier bestätigt der Eintrag selbst.
    { name: 'HandelSchicken', argumente: (k) => [h(k, 'h1'), gegenueber(k)], nimm: 'a.nachrichtId', nach: merke('nachricht') },
    { name: 'HandelPostbestaetigung', argumente: (k) => [h(k, 'nachricht')] },
    schritt('h1', 'zustimmen'),
    schritt('h1', 'zustimmen', gegenueber),
    schritt('h1', 'vorbereiten'),
    schritt('h1', 'uebergeben'),
    schritt('h1', 'uebergeben', gegenueber),
    freigabe('h1', 'verbuchen', () => '', 'nachweisVerbuchen'),
    { name: 'Handelsverbuchung', argumente: (k) => [h(k, 'h1'), h(k, 'nachweisVerbuchen')] },
    { name: 'Handelsabgleich', argumente: (k) => [h(k, 'h1')] },
    freigabe('h1', 'bestaetigen', ich, 'nachweisEigene'),
    { name: 'Handelsbestaetigung', argumente: (k) => [h(k, 'h1'), ich(k), h(k, 'nachweisEigene')] },
    freigabe('h1', 'bestaetigen', gegenueber, 'nachweisGegenueber'),
    { name: 'Handelsbestaetigung', argumente: (k) => [h(k, 'h1'), gegenueber(k), h(k, 'nachweisGegenueber')] },
    { name: 'HandelArchivieren', argumente: (k) => [h(k, 'h1'), true] },
    { name: 'HandelArchivieren', argumente: (k) => [h(k, 'h1'), false] },

    // --- ein Entwurf: ändern, kopieren, löschen, beenden ---
    { name: 'HandelEroeffnen', argumente: (k) => [wunsch(k)], nimm: vorgang, nach: merke('h2') },
    {
      name: 'HandelEntwurfAendern',
      argumente: (k) => [wunsch(k, { vorgang: h(k, 'h2'), positionen: [position(gegenueber(k), { menge: 2 })] })],
    },
    { name: 'HandelKopieren', argumente: (k) => [h(k, 'h2')], nimm: vorgang, nach: merke('h3') },
    { name: 'HandelEntwurfLoeschen', argumente: (k) => [h(k, 'h3')] },
    {
      name: 'HandelVorschlagen', argumente: (k) => [h(k, 'h2'), [position(gegenueber(k))]], erwartet: 'fachfehler',
      grund: 'ein Gegenangebot macht nur die Gegenseite, und aus „entwurf" führt keines (internal/domain/handelsmatrix.go:115)',
    },
    schritt('h2', 'abbrechen'),

    // --- ein Entwurf, der eine eigene Besitzzeile festhält ---
    { name: 'ExemplarAnlegen', argumente: () => [exemplar], nimm: 'a.id', nach: merke('exemplar') },
    {
      name: 'Sammlung', nimm: 'a.exemplare.map((e) => ({ id: e.id, fassung: e.fassung }))',
      nach: (w, k) => merke('exemplarfassung')(w.find((e) => e.id === h(k, 'exemplar')).fassung, k),
    },
    {
      name: 'HandelEroeffnen',
      argumente: (k) => [wunsch(k, { positionen: [position(ich(k), { printing: 'vertrag-h01' })] })],
      nimm: vorgang, nach: merke('h4'),
    },
    { name: 'HandelFesthalten', argumente: (k) => [h(k, 'h4'), h(k, 'exemplar'), 1, h(k, 'exemplarfassung')] },
    // Ein beendeter Handel gibt seine Zusagen frei.
    schritt('h4', 'abbrechen'),
    { name: 'HandelArchivieren', argumente: (k) => [h(k, 'h4'), true] },
    { name: 'HandelArchivieren', argumente: (k) => [h(k, 'h4'), false] },

    // **Nicht gelingbar**: siehe den Kopf. Ein Fremder ohne Kontakt gibt eine
    // sofortige Absage („darf handel nicht") und keine Wartezeit.
    {
      name: 'HandelZustellen', argumente: (k) => [fremder(k, 'zustellen').kennung], erwartet: 'fachfehler',
      grund: 'baut eine Leitung zu einem Gegenüber, das es hier nicht gibt (internal/app/handelszustellung.go:45)',
    },

    // --- die Deckleihe ---
    {
      name: 'DeckleiheAnbieten', argumente: (k) => [leihwunsch(k, 'p1'), leihpfad(k, 1)], nimm: leihenform,
      nach: neueLeihe('leihe1'),
    },
    { name: 'DeckleiheWiderrufen', argumente: (k) => [h(k, 'leihe1')] },
    {
      name: 'DeckleiheAnbieten', argumente: (k) => [leihwunsch(k, 'p2'), leihpfad(k, 2)], nimm: leihenform,
      nach: neueLeihe('leihe2'),
    },
    {
      name: 'DeckleiheAnnehmen', argumente: (k) => [leihpfad(k, 2)], erwartet: 'fachfehler',
      grund: 'der Umschlag gilt einem Fremden, und von sich selbst unterschrieben nimmt dieses Profil nichts an (internal/service/deckleihe.go:146)',
    },
    { name: 'DeckleiheZurueckgeben', argumente: (k) => [h(k, 'leihe2')] },
  ],
}
