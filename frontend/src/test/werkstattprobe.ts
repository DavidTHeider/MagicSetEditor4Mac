// **Die Belegung der Vorlagenwerkstatt** — was jeder Lauf über `Vorlagen.tsx`
// braucht, damit keine Region ins Leere greift (Werkstück U7, #1418).
//
// Eigene Belegungen überschreiben die Vorgaben; wer einen Fehlerweg prüft,
// belegt genau die eine Bindung anders.
import { belegeApi, belegeFenster } from './api'
import { paketDTO } from './vorlagenprobe'

import type { Api } from '../api'

export function belegeWerkstatt(eigene: Partial<Api> = {}): void {
  belegeApi({
    ...belegeFenster(),
    Pakete: () =>
      Promise.resolve({
        pakete: [paketDTO('probe.mse-style')],
        befunde: [],
        fehler: null,
      }),
    Paketdateien: () =>
      Promise.resolve({
        dateien: ['style', 'card_fields'],
        beiwerk: [],
        fehler: null,
      }),
    PaketdateiLesen: () =>
      Promise.resolve({
        inhalt: 'mse version: 2.0.0\n',
        stand: 'alt',
        fehler: null,
      }),
    PaketdateiSchreiben: () => Promise.resolve({ stand: 'neu', fehler: null }),
    PaketdateiPruefen: () =>
      Promise.resolve({ stelle: null, meldung: '', schluessel: '', werte: [] }),
    Vorlagenvorschau: () =>
      Promise.resolve({
        vorschau: null,
        fehler: {
          code: 'nichtGefunden',
          schluessel: 'keinSetMitStil',
          meldung: 'kein Set',
        },
      }),
    Probekarten: () => Promise.resolve({ karten: [], fehler: null }),
    Paketsymbole: () => Promise.resolve({ symbole: [], fehler: null }),
    Symbolverwendungen: () =>
      Promise.resolve({ verwendungen: [], fehler: null }),
    Paketfarben: () => Promise.resolve({ merkmale: [] }),
    ...eigene,
  })
}
