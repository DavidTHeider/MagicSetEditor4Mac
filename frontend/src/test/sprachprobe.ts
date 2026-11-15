// **Die Proben der eigenen Übersetzungen** — LOC-02
// (#1688).
// Eine eigene Datei, weil zwei Tests sie brauchen — dieselbe Aufteilung wie
// bei `deckprobe.ts`.
import type { SprachpaketAntwort, SprachpaketDTO } from '../apisprache'

export function sprachpaketprobe(
  teile: Partial<SprachpaketDTO> = {},
): SprachpaketDTO {
  return {
    kennung: 'pt-BR',
    name: 'Português',
    basis: 'en',
    stand: 'entwurf',
    lizenz: '',
    herkunft: '',
    geaendert: '2026-09-30T00:00:00Z',
    gesamt: 3,
    uebersetzt: 1,
    fehlend: 2,
    veraltet: 0,
    unbekannt: 0,
    ...teile,
  }
}

export function sprachantwort(
  teile: Partial<SprachpaketAntwort> = {},
): SprachpaketAntwort {
  return {
    paket: sprachpaketprobe(),
    texte: {},
    fehlendePfade: [],
    veraltetePfade: [],
    unbekanntePfade: [],
    pfad: '',
    abgebrochen: false,
    fehler: null,
    ...teile,
  }
}
