// **Lesende Bindungen mit echten Argumenten** — C01
// (#2226).
//
// Die Sammelgruppe „Übrige" ruft jede Bindung ohne eigenen Eintrag mit
// Nullwerten (`ersatz.mjs`). Für eine lesende heißt das meist eine fachliche
// Ablehnung — die Antwortform des Erfolgs bleibt dann ungeprüft, und genau sie
// zählt die Quote über die lesende Registry (Schwelle 90 %). Hier bekommen
// diese Bindungen Argumente, mit denen sie antworten: die reinen Rechnungen
// des Symboleditors über einer kleinen Zeichnung, Pakete, Texturen, Decks,
// Partien, Handel, Statistik und Ausgabe über dem Bestand der Gruppen davor.
//
// **Was der erste Lauf am nativen Fenster zeigte** (2026-10-10, 84,9 %):
// sechs Rufe scheiterten an ihrer Vorrichtung, nicht an der Anwendung — die
// Saat liegt im Bündel und nicht im verwalteten Paketverzeichnis, aus dem
// `Paketdateien` liest; eine Exportvorlage heißt nach ihrem Paket samt
// Endung; das Schlüsselwort und die Partie der Gruppen davor gab es nicht
// mehr (gelöscht, beendet); und Decks übernimmt man nur aus einem **anderen**
// Profil. Jede bekommt hier ihren eigenen Gegenstand.
//
// **Was hier nicht steht** — mit Grund: die Suche im Index, die Recherche,
// ähnliche Karten und das Kontextpaket brauchen einen gebauten Suchindex und
// ein Einbettungsmodell; ein KI-Bildzuschnitt ein gewartetes Bild; ein
// Kontostand ein Konto bei einem fremden Dienst; `TeilenAnbieten` öffnet das
// Teilenblatt von macOS, `DateiImFinderZeigen` und `DateiMitProgrammOeffnen`
// ein Fenster außerhalb der Probe. Sie bleiben in „Übrige" und zählen gegen
// die Quote.
import { join } from 'node:path'

import { b, einladungVon, fremder, schreibePng } from './hilfen-b.mjs'

const punkt = (x, y) => ({ x, y, griffDavorX: 0, griffDavorY: 0, griffDanachX: 0, griffDanachY: 0, kurveDanach: false, sperre: 'free' })
const form = (teilname, punkte) => ({
  art: 'Form', name: teilname, kombination: 'merge', punkte,
  spiegelart: '', kopien: 0, mitteX: 0, mitteY: 0, griffX: 0, griffY: 0, teile: [],
})
const quadrat = form('Quadrat', [punkt(0.2, 0.2), punkt(0.8, 0.2), punkt(0.8, 0.8), punkt(0.2, 0.8)])
const dreieck = form('Dreieck', [punkt(0.5, 0.1), punkt(0.9, 0.9), punkt(0.1, 0.9)])
const lesesitz = (k, kennung, name) => ({ kennung, name, deckID: String(c(k).partiedeck), commander: '', leiheID: '', profilID: '' })
const merke = (name) => (w, k) => { (k.c ??= {})[name] = w }
const c = (k) => k.c ?? {}

export const cLesend = {
  name: 'Lesende Bindungen mit echten Argumenten (C01)',
  rufe: [
    // --- die Rechnungen des Symboleditors: Teile hinein, Teile heraus ---
    { name: 'Grundform', argumente: () => [{ art: 'Rechteck', ecken: 0, eng: false }] },
    { name: 'GriffZiehen', argumente: () => [[quadrat], { teil: 0, punkt: 0, davor: false, x: 0.3, y: 0.3 }] },
    { name: 'PunktEinfuegen', argumente: () => [[quadrat], { teil: 0, punkt: 0 }] },
    { name: 'PunktLoeschen', argumente: () => [[quadrat], { teil: 0, punkt: 0 }] },
    { name: 'TeilVerschieben', argumente: () => [[quadrat, dreieck], 0, 1] },
    { name: 'TeileGruppieren', argumente: () => [[quadrat, dreieck], [0, 1]], nimm: 'a.teile', nach: merke('gruppe') },
    { name: 'GruppeAufloesen', argumente: (k) => [c(k).gruppe, [0]] },
    { name: 'SymmetrieAnlegen', argumente: () => [[quadrat], 0], nimm: 'a.teile', nach: merke('symmetrie') },
    {
      name: 'SymmetrieSetzen', argumente: (k) => [c(k).symmetrie, {
        teil: 0, spiegelart: 'rotation', kopien: 3, mitteX: 0.5, mitteY: 0.5, griffX: 0, griffY: 0, mitteSetzen: true, griffSetzen: false,
      }],
    },
    { name: 'Symmetriekopien', argumente: (k) => [c(k).symmetrie, 0] },
    { name: 'SymmetrieAufloesen', argumente: (k) => [c(k).symmetrie, 0] },

    // --- ein Symbol im Set und eines in der Bibliothek ---
    { name: 'SymbolSetzen', argumente: (k) => [k.set.id, [quadrat], ''], nimm: 'a.name', nach: merke('symboldatei') },
    { name: 'Symbol', argumente: (k) => [k.set.id, c(k).symboldatei] },
    { name: 'SetwertSetzen', argumente: (k) => [k.set.id, 'symbol', c(k).symboldatei, []] },
    { name: 'BibliothekssymbolAufnehmen', argumente: (k) => [k.set.id, 'Lesesymbol'], nimm: 'a.id', nach: merke('bibliothek') },
    { name: 'BibliothekssymbolTeile', argumente: (k) => [c(k).bibliothek] },
    { name: 'BibliothekssymbolFassungen', argumente: (k) => [c(k).bibliothek, k.set.id] },
    { name: 'SetwertSetzen', argumente: (k) => [k.set.id, 'symbol', '', []] },
    { name: 'BibliothekssymbolLoeschen', argumente: (k) => [c(k).bibliothek, true] },

    // --- Pakete: ein eigenes Include im verwalteten Verzeichnis ---
    { name: 'Pakete' },
    {
      name: 'PaketAnlegen', argumente: () => ['mse-include', 'vertrag-lesen', 'Lese-Include', ''],
      nimm: 'a.paket.pfad', nach: merke('lesepaket'),
    },
    { name: 'Paketdateien', argumente: (k) => [c(k).lesepaket] },
    { name: 'PaketdateiLesen', argumente: (k) => [c(k).lesepaket, 'include'] },
    { name: 'Paketsymbole', argumente: (k) => [c(k).lesepaket] },
    { name: 'Symbolverwendungen', argumente: (k) => [c(k).lesepaket, 'include'] },
    { name: 'PaketPlan', argumente: () => ['entfernen', 'mse-game', 'probe', ''] },
    { name: 'PaketEntfernen', argumente: () => ['mse-include', 'vertrag-lesen'] },

    // --- Set, Karten, Stil, Statistik und Ausgabe ---
    { name: 'Stilfolgen', argumente: (k) => [k.set.id, 'probe-kunst'] },
    { name: 'Auszaehlen', argumente: (k) => [k.set.id, 'farbe'] },
    { name: 'PackZiehen', argumente: (k) => [k.set.id, 'probepack', '2226'] },
    { name: 'Bildnamen', argumente: (k) => [k.set.id, 'karte', 'png', 'nummerieren', []] },
    { name: 'BogenProbelauf', argumente: (k) => [k.set.id, 'a4', 3, 300, true, []] },
    {
      name: 'Exportvorlagen', argumente: (k) => [String(k.set.id)],
      nimm: 'a.vorlagen.find((v) => v.paket.startsWith("probe-liste"))?.paket', nach: merke('vorlage'),
    },
    { name: 'VorlageProbelauf', argumente: (k) => [k.set.id, c(k).vorlage, {}, []] },
    { name: 'SchluesselwortAnlegen', argumente: (k) => [k.set.id], nimm: 'a.wortID', nach: merke('wort') },
    { name: 'Parameterhilfen', argumente: (k) => [k.set.id, c(k).wort] },
    { name: 'SchluesselwortLoeschen', argumente: (k) => [k.set.id, c(k).wort] },
    { name: 'Rueckentextur', argumente: (k) => [k.set.spiel, 'tisch'] },
    { name: 'Spieltextur', argumente: (k) => [k.karten[0], 'vorne', 'tisch'] },
    {
      name: 'BildAusDatei', argumente: (k) => [k.karten[0], k.textfeld ?? 'name', join(k.ort.ort, 'vertrag', 'lesebild.png')],
      vorher: async (_p, k) => { schreibePng(join(k.ort.ort, 'vertrag', 'lesebild.png')) },
    },
    { name: 'DateiAnsehen', argumente: (k) => [join(k.ort.ort, 'vertrag', 'lesebild.png')] },

    // --- Sammlung, Decks, Regeln, Rücksetzung ---
    { name: 'SammlungAusgeben', argumente: () => ['liste'] },
    { name: 'DecklisteVorschau', argumente: (k) => [Number(k.deck), '1 Feuerball'] },
    // **Aus einem anderen Profil desselben Menschen** — eigens angelegt; das
    // aus `b-nutzer` ist zu diesem Zeitpunkt schon gelöscht.
    {
      name: 'ProfilAnlegen', argumente: (k) => [b(k).ichID, 'Leseprofil'],
      nimm: '({ id: a.profil.id, name: a.profil.name })', nach: merke('leseprofil'),
    },
    { name: 'DeckeZurUebernahme', argumente: (k) => [c(k).leseprofil.id] },
    { name: 'ProfilLoeschen', argumente: (k) => [c(k).leseprofil.id, c(k).leseprofil.name] },
    { name: 'Regelfreigabe', argumente: () => ['magic', 'streng', 2, '', []] },
    { name: 'Ruecksetzungsvorschau', argumente: () => ['layout', '', ''] },

    // --- eine laufende Partie: die aus `b-partien` ist beendet, ihr Deck weg ---
    {
      name: 'DeckAnlegen', argumente: (k) => ['Lesepartie-Deck', k.set.spiel, 'aus dem Rundgang'],
      nimm: 'a.deck.id', nach: merke('partiedeck'),
    },
    { name: 'DeckplatzLegen', argumente: (k) => [c(k).partiedeck, k.karten[0], '', 2, 'hauptdeck'] },
    {
      name: 'PartieAnlegen',
      argumente: (k) => [{ spiel: k.set.spiel, regelwerk: {}, wirt: '', sitze: [lesesitz(k, 'ada', 'Ada'), lesesitz(k, 'bo', 'Bo')] }],
      nimm: 'a.partie.id', nach: merke('partie'),
    },
    { name: 'Partie', argumente: (k) => [c(k).partie, 'ada'] },
    { name: 'PartieVerlauf', argumente: (k) => [c(k).partie, 'ada'] },
    { name: 'PartieOffenlegungen', argumente: (k) => [c(k).partie] },
    { name: 'PartieBeenden', argumente: (k) => [c(k).partie, 'ada'] },

    // --- Handel aus der Gruppe davor: h1 ist verbucht ---
    { name: 'HandelStand', argumente: (k) => [b(k).handel?.h1] },
    { name: 'Handelsbeleg', argumente: (k) => [b(k).handel?.h1] },
    { name: 'HandelVorgaenge', argumente: (k) => [fremder(k, 'handel').kennung] },

    // --- Einladung und Teilen, ohne dass etwas hinausgeht ---
    { name: 'KontakteinladungErstellen', argumente: () => ['Leser', 'paarung'], nimm: 'a.einladung.code', nach: merke('einladung') },
    // **Die eigene prüft das Programm nicht** — es sagt, sie gehöre zum
    // Gegenüber; geprüft wird eine, die ein fremder Peer ausgestellt hat.
    { name: 'KontakteinladungPruefen', argumente: (k) => [einladungVon(fremder(k, 'leseeinladung'))] },
    { name: 'TeilenVorbereiten', argumente: () => ['sammlung', ''], nimm: 'a.teilgut.kennung', nach: merke('teilung') },
    { name: 'TeilenVerwerfen', argumente: (k) => [c(k).teilung] },
  ],
}
