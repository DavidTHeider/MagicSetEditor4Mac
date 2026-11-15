// **Was die Seiten außer Set und Karten lesen** — für die Wächter über
// `seiten` (Barrierefreiheit, Auszeichnung, Sprache).
//
// # Der Befund
//
// Ein Erkundungslauf über alle 25 Seiten fand acht Methoden, die die
// Vorrichtung nicht belegte: Verlauf, Packarten, Setangaben,
// Schlüsselwörter, Sammlung, Portfolio, Partien und Turniere. Auf
// Setangaben, Schlüsselwörter, Zufallspack, Sammlung, Spielen und Turniere
// stand deshalb „Die Methode … ist nicht belegt" — und jeder Wächter prüfte
// dort eine Fehlermeldung statt der Seite. Dieselbe Falle, die der Kopf von
// `belegeProbeset` für die Nutzerseite beschreibt; sie war wieder
// zugeschnappt (G02, #2249).
//
// **Befüllt und nicht leer**: eine leere Liste zeigte nur den Leerzustand,
// und die Zeilen, an denen Beschriftungen hängen, prüfte niemand.
import type { Api, FeldbeschreibungDTO } from '../api'
import { exemplar } from './sammlungsseite'
import { wort } from './schluesselwoerter'
import { partienbestand } from './spielbestand'
import { liste as turnierliste } from './turnierbestand'

function feld(name: string, aufschrift: string): FeldbeschreibungDTO {
  return {
    name,
    aufschrift,
    art: 'text',
    mehrzeilig: false,
    wahlen: [],
    beschreibung: '',
    vorgabe: '',
    teilfelder: [],
    reihe: '',
    stelle: 0,
    editierbar: true,
    gespeichert: true,
  }
}

export function seitenteile(): Partial<Api> {
  return {
    Verlauf: () => Promise.resolve({ schritte: [], fehler: null }),
    Packarten: () =>
      Promise.resolve({
        arten: [
          { name: 'booster', waehlbar: true, karten: 1 },
          { name: 'common', waehlbar: false, karten: 1 },
        ],
        befunde: [],
        fehler: null,
      }),
    Setangaben: () =>
      Promise.resolve({
        felder: [feld('title', 'Titel'), feld('set code', 'Setkürzel')],
        werte: [{ feld: 'title', wert: 'Probeset' }],
        kopf: {
          mseFassung: '2.0.0',
          spielfassung: '',
          stilfassung: '',
          versionskontrolle: '',
          apprenticecode: '',
        },
        fehler: null,
      }),
    Schluesselwoerter: () =>
      Promise.resolve({
        woerter: [
          wort({
            id: '1',
            wort: 'Flugwurf',
            muster: 'Flugwurf',
            erinnerung: 'Diese Karte fliegt.',
            modus: 'custom',
          }),
          wort({ id: 'f', wort: 'Flying', modus: 'core', fest: true }),
        ],
        modi: ['core', 'custom'],
        fehler: null,
      }),
    Parameterhilfen: () =>
      Promise.resolve({ arten: [], verweise: [], befunde: [], fehler: null }),
    Sammlung: () => Promise.resolve({ exemplare: [exemplar()], fehler: null }),
    Portfolio: () =>
      Promise.resolve({ zeilen: [], ohnePreis: 0, fehler: null }),
    Partien: () =>
      Promise.resolve({ partien: partienbestand.slice(0, 3), fehler: null }),
    Turniere: () => Promise.resolve({ turniere: turnierliste, fehler: null }),
  }
}
