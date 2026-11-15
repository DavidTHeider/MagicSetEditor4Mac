// **Vorrichtung für die Liste „Was gerade läuft"** — TRADE-01
// (#1715).
import type { Handelsvorgang } from '../apihandelsleitung'
import type { Handelsvorgangzeile } from '../apihandelsverzeichnis'

// einVorgang ist ein Handel mit dem Gegenüber der Kontaktprobe.
export function einVorgang(
  anders: Partial<Handelsvorgang> = {},
): Handelsvorgang {
  return {
    vorgang: 'h-1',
    revision: 1,
    eltern: '',
    abdruck: 'abcdef0123456789',
    art: 'tausch',
    initiator: 'bich',
    parteien: ['bich', 'bgegenueber'],
    ich: 'bich',
    zustand: 'angeboten',
    positionen: [],
    bewertungGeteilt: false,
    zustimmungen: [],
    ende: false,
    beschluss: '',
    lage: 'dran',
    naechster: 'zustimmen',
    moegliche: ['zustimmen', 'abbrechen', 'ablaufen'],
    phase: 'angebot',
    ...anders,
  }
}

// eineZeile ist dieser Handel in der Liste.
export function eineZeile(
  anders: Partial<Handelsvorgangzeile> = {},
  vorgang: Partial<Handelsvorgang> = {},
): Handelsvorgangzeile {
  return {
    vorgang: einVorgang(vorgang),
    lage: 'dran',
    naechster: 'zustimmen',
    gegenueber: [{ kennung: 'bgegenueber', name: 'Turfer' }],
    karten: ['Lightning Bolt'],
    marken: [],
    angelegt: '2026-09-28T10:00:00Z',
    geaendert: '2026-09-29T10:00:00Z',
    archiviert: false,
    unbestaetigt: false,
    ...anders,
  }
}

// eineUebersicht ist die Antwort mit diesen Zeilen.
export function eineUebersicht(zeilen: Handelsvorgangzeile[]) {
  return Promise.resolve({ zeilen, fehler: null })
}
