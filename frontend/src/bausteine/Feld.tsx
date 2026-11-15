// **Ein Feld: Beschriftung, Bedienelement, Hinweis, Befund** — Werkstück U2
// von Phase 5.1
// (#1413),
// nach [Vermerk 0077](../../../docs/entscheidungen/0077-eigene-primitive-statt-radix.md).
//
// # Wozu
//
// Fünf Seiten und zwei Bausteine hängen einen Befund an ein Feld, und jede
// tut es auf ihre Art: `Setangaben` mit `fehlerkennung()`, `Pflichtfeld` mit
// eigenem `useId`, `Zahlenfeld` mit einem Bereichssatz. Was gleich ist, steht
// jetzt hier — die Verbindung von Beschriftung, Hinweis und Befund mit dem
// Element, das sie meinen.
//
// # Warum ein Render-Rückruf
//
// Das Bedienelement ist mal ein `input`, mal ein `select`, mal ein
// `textarea`, und U3 stellt sie nacheinander um. Der Rückruf gibt die
// Verbindung — Kennung, `aria-describedby`, `aria-invalid` — und der Aufrufer
// setzt sie an sein Element. So bleibt das Element seins, und die Zusage
// bleibt hier.
//
// # Was es zusagt
//
//   - `label` und `htmlFor` — die native Beschriftung, die jedes
//     Vorleseprogramm kennt (die Radix-Doku sagt dasselbe).
//   - Hinweis **und** Befund stehen in `aria-describedby`, wenn sie dastehen.
//   - Der Befund ist ein `role="alert"` **am Feld** (§6: Fehler erscheinen am
//     verursachenden Feld), und `aria-invalid` sagt es dem Element.
import { useId, type ReactNode } from 'react'

// Feldverbindung ist, was das Bedienelement bekommt.
export interface Feldverbindung {
  id: string
  'aria-describedby': string | undefined
  'aria-invalid': boolean
}

interface FeldEigenschaften {
  beschriftung: string

  // hinweis steht unter dem Feld, solange kein Befund da ist — was das Feld
  // erwartet, in einem Satz.
  hinweis?: string

  // befund ist der Satz der Anwendung: was nicht stimmt und was zu tun ist.
  befund?: string

  // beschreibungsID nennt einen Hilfesatz, der **ohnehin** auf der Seite
  // steht — Werkstück U3 von Phase 5.1
  // (#1414).
  //
  // **Er tritt neben Hinweis und Befund, nicht an ihre Stelle.** Was das Feld
  // bedeutet, gilt weiter, auch wenn der Wert gerade nicht passt; ein
  // `aria-describedby`, das ihn beim Fehler herauswirft, nimmt dem Feld seine
  // Erklärung genau dann, wenn sie gebraucht wird.
  beschreibungsID?: string

  // platzFuerBefund hält die Zeile des Befunds frei, solange keiner da ist
  // (#2024). **Für ein Feld, das sich beim Verlassen meldet**: die Meldung
  // schob sonst die Knöpfe darunter hinab, bevor die Maustaste los war, und
  // der erste Klick ging ins Leere.
  platzFuerBefund?: boolean

  children: (verbindung: Feldverbindung) => ReactNode
}

export function Feld({
  beschriftung,
  hinweis,
  befund,
  beschreibungsID,
  platzFuerBefund = false,
  children,
}: FeldEigenschaften) {
  const feldID = useId()
  const hinweisID = useId()
  const befundID = useId()

  // **Beide Kennungen, wenn beide stehen** — ein Vorleseprogramm liest erst
  // den Befund, dann den Hinweis; wer nur den Befund nennt, nimmt dem Feld
  // seine Erklärung genau dann, wenn sie gebraucht wird.
  const beschriebenVon = [
    befund === undefined ? undefined : befundID,
    hinweis === undefined ? undefined : hinweisID,
    beschreibungsID,
  ]
    .filter((kennung) => kennung !== undefined)
    .join(' ')

  return (
    <div className="flex flex-col gap-1">
      <label htmlFor={feldID} className="text-sm font-medium">
        {beschriftung}
      </label>

      {children({
        id: feldID,
        'aria-describedby': beschriebenVon === '' ? undefined : beschriebenVon,
        'aria-invalid': befund !== undefined,
      })}

      {hinweis !== undefined && (
        <p id={hinweisID} className="text-sm text-leise">
          {hinweis}
        </p>
      )}

      {befund === undefined && platzFuerBefund && (
        // **So hoch wie eine Zeile der Meldung** — `text-sm` hat 1,25rem.
        <div data-befundplatz="" aria-hidden="true" className="h-5" />
      )}

      {befund !== undefined && (
        <div role="alert">
          <p id={befundID} className="text-sm text-fehler">
            {befund}
          </p>
        </div>
      )}
    </div>
  )
}
