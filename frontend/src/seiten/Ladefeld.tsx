// Der gemeinsame Rahmen für Laden, Fehler und Leere.
//
// **Eine Stelle für drei Zustände**, weil sie sonst auf jeder Seite anders
// aussehen — und weil der Fehler dann irgendwann auf einer Seite fehlt.
//
// `role="status"` an der Lademeldung ist nicht Zierde: wer mit einem Vorleser
// arbeitet, erfährt sonst nichts davon, dass etwas unterwegs ist.
import type { ReactNode } from 'react'

import { useTexte } from '../sprache/kontext'

interface LadefeldEigenschaften {
  laedt: boolean
  fehler: string | undefined
  leer: boolean

  // grund sagt beim Laden, **warum es dauert** — Werkstück C von Phase 4.10
  // (#1325).
  //
  // **Nur wo es wirklich dauert.** Die meisten Läufe sind schnell, und eine
  // Erklärung für etwas, das man nicht bemerkt, ist Lärm. `Packarten`
  // braucht beim **ersten** Aufruf mit *Blue* (3.693 Karten) am nativen
  // Fenster 16,6 s (2026-10-19, Mac17,3, `zufallspack09`; vorher rund 100 s,
  // #2466), danach Millisekunden — der Speicher greift. Sekunden „Wird
  // geladen …" ohne Grund sehen aus wie ein Fehler (§8).
  grund?: string

  // vorhanden sagt, dass schon Inhalt dasteht — beim **Nachlesen** und nicht
  // beim ersten Laden (U11, #2125). Dann wird das Laden nur dem Vorleser
  // gesagt: eine sichtbare Zeile schob den Inhalt nach unten und gleich
  // wieder zurück.
  vorhanden?: boolean

  // children ist der **leere Zustand** — der Text, der den nächsten Schritt
  // erklärt (§6). Er steht beim Aufrufer, weil nur der weiß, was als Nächstes
  // dran ist.
  children: ReactNode
}

export function Ladefeld({
  laedt,
  fehler,
  leer,
  grund,
  vorhanden = false,
  children,
}: LadefeldEigenschaften) {
  const texte = useTexte()

  if (laedt && vorhanden) {
    return (
      <p role="status" className="sr-only">
        {texte.bausteine.zustaende.laedt}
      </p>
    )
  }

  if (laedt) {
    return (
      <div role="status" className="mt-3 text-leise">
        <p>{texte.bausteine.zustaende.laedt}</p>

        {grund !== undefined && (
          <p className="mt-1 max-w-prose text-sm">{grund}</p>
        )}
      </div>
    )
  }

  if (fehler !== undefined) {
    // **Die rohe Ursache steht dabei, aber nicht allein.** Der Satz sagt, was
    // los ist; die Ursache hilft dem, der sie melden will.
    return (
      <p role="alert" className="mt-3 text-fehler">
        {texte.bausteine.zustaende.fehler} {fehler}
      </p>
    )
  }

  if (leer) {
    return <>{children}</>
  }

  return null
}
