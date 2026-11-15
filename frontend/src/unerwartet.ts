// **Ein Fehler ohne Schlüssel bekommt eine Kennung, und sein Rohtext geht ins
// Protokoll** — F03 (#2342),
// Befunde A-03 und A-24.
//
// # Was hier fremd heißt
//
// **Ein Fehler mit eigenem Typ** — `TypeError`, `RangeError`, Wails'
// `RuntimeError`: sein Satz kommt aus der Laufzeit, englisch und für
// niemanden am Fenster gemacht. **Ein schlichtes `Error` bleibt**, wie es ist:
// einen Satz darin verbietet `keinfehlertext.test.ts` im eigenen Code, und
// was dennoch kommt, ist ein Befund dort und nicht hier. **Ein schlichtes
// `Error` aus einer Bindung** — Wails' Weg für eine Antwort ohne JSON —
// kommt hier gar nicht an: `anbindung.ts` macht schon an der Grenze einen
// Fehler »unbekannt« mit Kennung daraus (Zweitblick über Phase 1 von 0.9.0,
// Befund 6).
//
// # Einmal je Fehler
//
// Dieselbe Ursache zeigen oft zwei Stellen — die Meldung und ihre Kennung.
// Gemeldet wird sie **einmal**, und beide zeigen dieselbe Kennung.
import { Bindungsfehler } from './bindungsfehler'
import { api } from './fassade'
import { Katalogfehler } from './katalogfehler'
import { neueKennung } from './kennung'

const kennungen = new WeakMap<Error, string>()

// istFremd sagt, ob eine Ursache ein Fehler mit eigenem Typ ist.
export function istFremd(ursache: unknown): ursache is Error {
  return (
    ursache instanceof Error &&
    ursache.name !== 'Error' &&
    !(ursache instanceof Bindungsfehler) &&
    !(ursache instanceof Katalogfehler)
  )
}

// gemeldeteKennung gibt die Kennung eines fremden Fehlers und meldet ihn
// beim ersten Mal. **Die Meldung ist Bestenfalls-Arbeit** (§7, Regel 3):
// scheitert sie, steht der Satz trotzdem da.
export function gemeldeteKennung(ursache: Error): string {
  const bekannt = kennungen.get(ursache)
  if (bekannt !== undefined) {
    return bekannt
  }

  const kennung = neueKennung()
  kennungen.set(ursache, kennung)

  try {
    const gemeldet = api().FensterfehlerMelden({
      kennung,
      ebene: 'meldung',
      bereich: 'fenster',
      meldung: String(ursache),
    })
    gemeldet.catch(() => undefined)
  } catch {
    // Ohne Fassade gibt es kein Protokoll; die Kennung steht trotzdem da.
  }

  return kennung
}
