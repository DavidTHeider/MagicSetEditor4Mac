// Der Kopf einer Seite: Titel, Erklärung, Handlungen.
//
// **Die Erklärung steht neben dem Titel und nicht in einem Hilfetext.** Eine
// Seite, die erst nach einem Klick sagt, wofür sie da ist, ist eine Seite, die
// niemand aufschlägt.
//
// **Der erste Satz davon** — UX-15.AK-01 bis AK-03
// (#1771).
// Gemessen am 2026-09-22 über die Seitenkataloge: die Kopferklärungen reichen
// bis **340** Zeichen, vier bis fünf Zeilen über dem, was man sucht. Der
// Satz oben bleibt wahr — er sagt, **dass** die Erklärung sichtbar ist, nicht
// **wie viel** davon.
//
// **Die Handlungen stehen rechts oben** — dieselbe Stelle auf jeder Seite. Wer
// sie einmal gefunden hat, sucht sie nicht wieder.
import { Feldhilfe } from './Feldhilfe'
import { kurzform, langform, ohneAuszeichnung } from '../sprache/kurzform'

import type { ReactNode } from 'react'

interface SeitenkopfEigenschaften {
  titel: string
  erklaerung?: string

  // ebene sagt, ob dieser Kopf eine Seite anfängt oder einen Abschnitt.
  //
  // **Eine Setlasche ist ein Abschnitt**
  // (#812):
  // der `Setarbeitsplatz` trägt die erste Ebene mit dem Namen des Sets, und
  // eine zweite `<h1>` darunter hieße für ein Vorleseprogramm, hier fingen
  // zwei Dokumente an.
  //
  // **Die Optik bleibt gleich.** Was sich ändert, ist die Ebene, nicht die
  // Größe — eine Lasche sieht aus wie eine Lasche.
  ebene?: 1 | 2

  children?: ReactNode
}

export function Seitenkopf({
  titel,
  erklaerung,
  ebene = 1,
  children,
}: SeitenkopfEigenschaften) {
  const Ueberschrift = ebene === 1 ? 'h1' : 'h2'

  // **Einmal gerechnet**: ein zweiter Ruf mit Rückfall daneben wäre ein
  // Zweig, den kein Lauf erreicht (§3).
  const langer =
    erklaerung === undefined
      ? undefined
      : langform(ohneAuszeichnung(erklaerung))

  return (
    <header className="mb-6 flex flex-wrap items-start justify-between gap-4">
      <div className="min-w-0">
        <Ueberschrift className="text-2xl font-semibold tracking-tight">
          {titel}
        </Ueberschrift>

        {erklaerung !== undefined && (
          <p className="mt-1 flex max-w-prose items-start gap-2 text-leise">
            <span>{kurzform(erklaerung)}</span>

            {langer !== undefined && <Feldhilfe zu={titel} text={langer} />}
          </p>
        )}
      </div>

      {/* **Nie breiter als der Kopf** (QA01, #2382): `shrink-0` hält die
          Zeile neben dem Titel zusammen, und ohne Obergrenze war sie so
          breit wie alle Knöpfe nebeneinander — bei sehr großer Schrift
          stand der letzte halb außerhalb, und `flex-wrap` griff nie. */}
      {children !== undefined && (
        <div className="flex max-w-full shrink-0 flex-wrap gap-2">
          {children}
        </div>
      )}
    </header>
  )
}
