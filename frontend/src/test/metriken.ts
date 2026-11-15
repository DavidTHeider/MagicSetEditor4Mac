// **Eine vollständige Metrikantwort für Läufe** — OPS-05 (#1722). Eine
// Antwort, die nur die Felder trägt, die ein Test braucht, sähe anders aus
// als die, die Go schickt (§3: eine Vorrichtung, großzügiger als der
// Betrieb, prüft sich selbst).
import type { MetrikenAntwort } from '../apimetriken'

export function volleMetriken(): MetrikenAntwort {
  return {
    reihen: [],
    grenze: 48,
    laufzeiten: [
      {
        flaeche: 'dienst',
        art: 'lesend',
        anzahl: 4,
        abgewiesen: 1,
        gescheitert: 0,
        summeMs: 10,
        eimer: [0, 3, 1, 0, 0, 0, 0],
      },
    ],
    eimerMs: [1, 5, 25, 100, 500, 2500],
    spur: {
      an: false,
      abtastung: 0,
      grenze: 0,
      aufbewahrungS: 0,
      gesehen: 0,
      aufgenommen: 0,
      verdraengt: 0,
      abgelaufen: 0,
      punkte: [],
    },
    speicher: { treffer: 7, geholt: 2, veraltet: 0, fehlgeschlagen: 1 },
    datenbank: {
      offen: 2,
      inBenutzung: 1,
      frei: 1,
      gewartet: 3,
      wartezeitMs: 4.5,
    },
    prozess: {
      heapBytes: 8388608,
      systemBytes: 33554432,
      goroutinen: 12,
      cpuMs: 1500,
      fussabdruckBytes: 41943040,
      web: [],
      webUngemessen: '',
    },
    seit: '2026-09-26T10:00:00Z',
    bis: '2026-09-26T11:00:00Z',
    aufwandMikros: 250,
    ungemessen: ['fenster', 'kommandozeile'],
    fehler: null,
  }
}
