// **Die Problemliste des Inspectors** — Werkstück U5 von Phase 5.1
// (#1416).
//
// Was an einer Karte nicht stimmt, steht an **einer** Stelle oben im
// Inspector, und jeder Eintrag führt zum Feld: der Feldfehler des Editors
// (ein Schreiben, das der Go-Teil abgewiesen hat) und die Befunde des
// Rasterers („Ebene 3 (name): …" — der Feldname steht in der Klammer,
// `render.Befund.String`). Ein Befund ohne Klammer hat kein Feld und keinen
// Knopf.
//
// # Warum keine zweite Quelle für Fehler
//
// Der Feldfehler steht **weiterhin** am Feld (§6) — hier steht er ein zweites
// Mal, damit er in einer Liste von fünfundzwanzig Feldern nicht unter der
// Falz liegt. Die Liste erfindet nichts: was sie zeigt, kommt aus Editor und
// Vorschau, und was sie tut, ist der Sprung.
import { useState } from 'react'

import { Faltfuss } from '../bausteine/Faltliste'
import { Knopf } from '../bausteine/Knopf'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

export interface Problem {
  feld: string | undefined
  meldung: string
}

// problemeAus baut die Liste aus dem Feldfehler und den Rasterbefunden.
export function problemeAus(
  feldfehler: { feld: string; meldung: string } | undefined,
  befunde: readonly string[],
): Problem[] {
  const liste: Problem[] = []

  if (feldfehler !== undefined) {
    liste.push({ feld: feldfehler.feld, meldung: feldfehler.meldung })
  }

  for (const befund of befunde) {
    liste.push({ feld: feldVon(befund), meldung: befund })
  }

  return liste
}

// feldVon liest das Feld aus einem Rasterbefund — `Ebene 3 (name): …`.
export function feldVon(befund: string): string | undefined {
  return /^Ebene \d+ \((.+?)\): /.exec(befund)?.[1]
}

interface ProblemlisteEigenschaften {
  probleme: readonly Problem[]
  zumFeld: (feld: string) => void
}

// vorn und jeKlick: drei Probleme stehen da, ein Klick zeigt die übrigen —
// UXR-06 (#1914). Fehlen die Schriften eines Pakets, meldet der Rasterer je
// Ebene, und bei `magic-m15` stünden ein Dutzend Zeilen mit je einem Knopf
// über dem Formular, oben im Inspector.
const vorn = 3
const jeKlick = 50

export function Problemliste({ probleme, zumFeld }: ProblemlisteEigenschaften) {
  const worte = useTexte().seiten.karten.probleme
  const [zahl, setzeZahl] = useState(vorn)

  return (
    <section
      aria-label={worte.ueberschrift}
      className="mb-3 rounded-md border border-kante p-2"
    >
      <h3 className="text-sm font-medium">{worte.ueberschrift}</h3>

      {probleme.length === 0 ? (
        <p className="text-sm text-leise">{worte.keine}</p>
      ) : (
        <ul className="mt-1 space-y-1 text-sm">
          {probleme.slice(0, zahl).map((problem, stelle) => {
            const feld = problem.feld

            return (
              <li
                key={`${feld ?? ''}/${String(stelle)}`}
                className="flex flex-wrap items-baseline gap-2 text-fehler"
              >
                <span className="flex-1">{problem.meldung}</span>
                {feld === undefined ? (
                  <span className="text-xs text-leise">{worte.ohneFeld}</span>
                ) : (
                  <Knopf
                    gewicht="leise"
                    aufKlick={() => {
                      zumFeld(feld)
                    }}
                  >
                    {einsetzen(worte.zumFeld, { feld })}
                  </Knopf>
                )}
              </li>
            )
          })}
        </ul>
      )}

      {probleme.length > vorn && (
        <Faltfuss
          gezeigt={Math.min(zahl, probleme.length)}
          gesamt={probleme.length}
          schritt={jeKlick}
          weiter={() => {
            setzeZahl(zahl + jeKlick)
          }}
        />
      )}
    </section>
  )
}
