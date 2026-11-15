// **Das Textfeld einer Einstellung ohne Werteliste** — herausgelöst aus
// `Wahl.tsx` für UXR-F2
// (#2008),
// als die Wahl ein Schalter, Segmente und eine Auswahlliste dazubekam und
// über die Zeilengrenze aus §2 ging. Der Schnitt folgt der Frage: dort steht,
// **welches** Bedienelement eine Einstellung bekommt, hier, wie ein Textfeld
// mit Entwurf sich verhält.
import { useRef, useState } from 'react'

// erbenWert ist die Wahl „wie beim Menschen" — `domain.Erben`.
export const erbenWert = 'erben'

// eingabeklassen sind Rand, Grund und Fokusring der Eingabefelder und der
// Auswahlliste einer Einstellung.
export const eingabeklassen =
  'rounded-md border border-kante bg-flaeche px-2 py-1 text-text ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring disabled:opacity-60 '

// Freitext ist das Textfeld einer Einstellung ohne Werteliste.
//
// **Mit Entwurf, geschrieben beim Verlassen oder mit ⏎** — OPS-02
// (#1719).
// Bis dahin schrieb jeder Tastendruck; seit das Setzen eines Ordners die
// Platte fragt, wiese sie jeden halben Pfad ab, und das Feld ließe sich nicht
// füllen. Dasselbe galt für jede geprüfte Angabe — eine Zeitzone, einen
// Proxy —, nur fiel es dort nicht auf.
export function Freitext({
  id,
  schluessel,
  wert,
  erbe,
  angehalten,
  aendern,
}: {
  id: string
  schluessel: string
  wert: string
  // erbe ist der Platzhalter einer Einstellung, die erben kann — sonst
  // nichts.
  erbe: string | undefined
  angehalten: boolean
  aendern: (wert: string) => void
}) {
  // **Wer erbt, sieht ein leeres Feld** und den geltenden Wert als
  // Platzhalter; **wer leert, erbt wieder** — für eine Breite ist ein leerer
  // Wert keiner (`pruefeBreiteOderErben`), und sonst gäbe es keinen Weg
  // zurück.
  const gezeigt = erbe !== undefined && wert === erbenWert ? '' : wert

  const [entwurf, setzeEntwurf] = useState(gezeigt)

  // **Einmal geschrieben ist geschrieben** — ⏎ und das Verlassen danach
  // schicken denselben Entwurf nicht zweimal.
  const zuletzt = useRef(gezeigt)

  const uebernehmen = () => {
    if (entwurf === zuletzt.current) {
      return
    }

    zuletzt.current = entwurf
    aendern(erbe !== undefined && entwurf.trim() === '' ? erbenWert : entwurf)
  }

  return (
    <input
      id={id}
      // **Der Schlüssel steht am Bedienelement** — Befund der Sichtprüfung
      // vom 2026-09-21
      // (#1676).
      // Nach dem Umbau auf neun Gruppen stand eine Einstellung **zweimal**
      // auf der Seite, und kein Lauf sah es: jeder sucht ein Element über
      // seine Aufschrift und findet eines. Über den Schlüssel lässt sich
      // zählen — `einstellungsdoppelung.test.tsx` tut es.
      data-einstellung={schluessel}
      type="text"
      value={entwurf}
      placeholder={erbe}
      disabled={angehalten}
      onChange={(ereignis) => {
        setzeEntwurf(ereignis.target.value)
      }}
      onBlur={uebernehmen}
      onKeyDown={(ereignis) => {
        if (ereignis.key === 'Enter') {
          uebernehmen()
        }

        // **Esc verwirft den Entwurf** — S13.NAT-02 (#1706): „Abbruch
        // ändert nichts". Das Verlassen danach findet nichts zu schreiben.
        if (ereignis.key === 'Escape') {
          setzeEntwurf(zuletzt.current)
        }
      }}
      className={eingabeklassen}
    />
  )
}
