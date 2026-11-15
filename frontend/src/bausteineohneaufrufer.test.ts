// **Kein Baustein ohne Aufrufer** — Befund 6 des Nachlaufs zum Zweitblick über
// Phase 5.1
// (#1405).
//
// # Warum es diesen Wächter gibt
//
// **Weil dieselbe Sache dreimal gefunden wurde und zweimal von Hand.** Der
// Zweitblick fand `Popover`, `Kontextmenue` und `Tooltip` ohne Verbraucher;
// der Nachlauf fand `Sprungmarken` — vier Bausteine, die im Baum standen und
// die niemand benutzte. §2 sagt dazu einen Satz: *ein Paket ohne Aufrufer ist
// keine Anwendungsfunktion.*
//
// **Ein Wächter statt eines vierten Fundes.** Was er prüft, ist die Klasse:
// jede Datei unter `src/bausteine/` und `src/rahmen/`, die etwas exportiert,
// wird von **mindestens einer** Datei außerhalb der Tests importiert.
//
// # Was er nicht prüft
//
//   - **Nicht, ob der Aufrufer sinnvoll ist.** Ein Import genügt ihm; ob das
//     Bauteil an der richtigen Stelle steht, sieht nur ein Mensch.
//   - **Nicht die Seiten.** `src/seiten/` hängt an `navigation.ts`, und dafür
//     gibt es einen eigenen Wächter (`navigation.test.ts`).
//   - **Nicht die Haken.** Sie sind keine Bauteile; wer einen ohne Aufrufer
//     schreibt, hat dasselbe Problem, aber der Schnitt hier folgt dem Ort,
//     an dem die vier Fälle lagen. **Sie hält seit SEC04
//     `hakenohneaufrufer.test.ts`** (#2357).
import { readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

// **`src` relativ zum Arbeitsverzeichnis** — Vitest läuft in `frontend/`,
// wie `wartemuster.test.ts` es auch annimmt.
const wurzel = 'src'

// quellen sind alle Dateien, die einen Import tragen können — ohne Tests.
function quellen(): string[] {
  const gefunden: string[] = []

  const gehe = (ort: string) => {
    for (const eintrag of readdirSync(ort, { withFileTypes: true })) {
      const pfad = join(ort, eintrag.name)

      if (eintrag.isDirectory()) {
        gehe(pfad)

        continue
      }

      const istQuelle = /\.tsx?$/u.test(eintrag.name)
      const istTest = /\.test\.tsx?$/u.test(eintrag.name)

      if (istQuelle && !istTest && !pfad.includes('/test/')) {
        gefunden.push(pfad)
      }
    }
  }

  gehe(wurzel)

  return gefunden
}

describe('die Bausteine dieses Fensters', () => {
  it('haben alle einen Aufrufer', () => {
    const alle = quellen()

    const bauteile = alle.filter(
      (pfad) => pfad.includes('/bausteine/') || pfad.includes('/rahmen/'),
    )

    expect(bauteile.length).toBeGreaterThan(50)

    const ohne = bauteile.filter((pfad) => {
      const name =
        pfad
          .split('/')
          .pop()
          ?.replace(/\.tsx?$/u, '') ?? ''

      // **Ein großgeschriebener Name ist ein Bauteil**, ein kleiner eine
      // Rechnung — beide brauchen einen Aufrufer, und beide werden über
      // ihren Dateinamen importiert.
      return !alle.some(
        (andere) =>
          andere !== pfad &&
          new RegExp(`from '[^']*/${name}'`, 'u').test(
            readFileSync(andere, 'utf8'),
          ),
      )
    })

    expect(ohne).toEqual([])
  })
})
