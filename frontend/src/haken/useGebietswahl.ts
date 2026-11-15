// **Welches Gebiet offen ist und wie die Tastatur es wechselt** — UX-04.AK-01
// (#1676).
//
// # Warum ein Haken und nicht zwei Bausteine mit demselben Inhalt
//
// **Seit dem 2026-09-21 gibt es zwei Gestalten derselben Sache**: die
// waagerechte Leiste aus [Laschen] und die senkrechte Liste aus
// [Gruppenliste], die das Artboard `Einstellungen` zeichnet. Verschieden ist,
// **wie** sie aussehen und welche Pfeiltasten sie nehmen; gleich ist alles
// andere — welches Gebiet offen ist, dass ein Wunsch von außen eines
// aufschlägt, dass die Pfeile umlaufen und der Fokus mitwandert.
//
// Ein zweites Bauteil mit demselben Inhalt daneben hätte die Falle, die §10
// beschreibt: eine Regel an zwei Stellen, von denen später nur eine gepflegt
// wird.
//
// # Was er nicht tut
//
// **Er zeichnet nichts.** Er gibt den Zustand und den Tastenhandler her; wie
// eine Zeile aussieht, ob sie einen Untertitel trägt und wo das Feld steht,
// entscheidet das Bauteil.
import { useId, useRef, useState, type KeyboardEvent } from 'react'

// Gebiet ist ein Bereich samt dem, was unter ihm steht.
export interface Gebiet {
  schluessel: string

  // name steht auf der Lasche oder in der Zeile — er kommt aus dem Katalog
  // des Aufrufers (§9).
  name: string

  inhalt: React.ReactNode
}

// ersterTeil nimmt den Schlüssel aus einem Wunsch — alles vor dem `#`.
//
// **Die Nummer dahinter macht den Wunsch unterscheidbar**, damit derselbe Weg
// zweimal wirkt; für die Zuordnung zählt sie nicht.
function ersterTeil(wunsch: string | undefined): string | undefined {
  return wunsch?.split('#')[0]
}

export interface Gebietswahl {
  // offen ist das Gebiet, das gerade dasteht — `undefined`, wenn es keines
  // gibt.
  offen: Gebiet | undefined

  // kennung ist der Stamm für die Kennungen von Zeile und Feld.
  kennung: string

  // leiste ist der Rahmen, in dem der Fokus wandert.
  leiste: React.RefObject<HTMLDivElement | null>

  waehle: (schluessel: string) => void

  // aufTaste bewegt den Fokus zwischen den Gebieten — das Muster, das eine
  // Vorlesehilfe bei `role="tablist"` erwartet.
  aufTaste: (ereignis: KeyboardEvent) => void
}

// useGebietswahl führt die Wahl über eine Liste von Gebieten.
//
// `vor` und `zurueck` sind die Tasten, die weiter und zurück gehen —
// `ArrowRight`/`ArrowLeft` bei einer Leiste, `ArrowDown`/`ArrowUp` bei einer
// Liste.
export function useGebietswahl(
  gebiete: readonly Gebiet[],
  oeffne: string | undefined,
  vor: string,
  zurueck: string,
): Gebietswahl {
  // **Der Anfangswert ist leer, und das ist Absicht.** Ein leerer Schlüssel
  // findet kein Gebiet, und `offen` fällt dann auf das erste zurück —
  // dieselbe Stelle, die auch greift, wenn ein Gebiet zwischen zwei
  // Zeichnungen wegfällt. Zwei Wege auf denselben Zweig sind besser als ein
  // zweiter Zweig daneben, den kein Aufruf erreicht (§3).
  const [gewaehlt, setzeGewaehlt] = useState('')

  // **Beginnt bei `undefined`, nicht bei `oeffne`.** Ein Wunsch, der schon
  // beim ersten Zeichnen dasteht, wäre sonst keine Änderung — und die Wahl
  // ignorierte ihn.
  const [letzterWunsch, setzeLetztenWunsch] = useState<string | undefined>(
    undefined,
  )

  if (oeffne !== letzterWunsch) {
    setzeLetztenWunsch(oeffne)

    const gewuenscht = ersterTeil(oeffne)

    if (
      gewuenscht !== undefined &&
      gebiete.some((gebiet) => gebiet.schluessel === gewuenscht)
    ) {
      setzeGewaehlt(gewuenscht)
    }
  }

  const leiste = useRef<HTMLDivElement>(null)
  const kennung = useId()

  const stelle = gebiete.findIndex((gebiet) => gebiet.schluessel === gewaehlt)

  // **`Math.max` und kein `? :`.** Beide rechnen dasselbe; die
  // Fallunterscheidung wäre ein Zweig, den ein Lauf einzeln betreten müsste,
  // um etwas zu belegen — und zu belegen gibt es hier nichts als eine
  // Untergrenze.
  const offen = gebiete[Math.max(stelle, 0)]

  // **Am Ende geht es von vorn los.** Ein Pfeil, der ins Leere zeigt, wäre
  // eine tote Taste.
  const aufTaste = (ereignis: KeyboardEvent) => {
    const richtung =
      ereignis.key === vor ? 1 : ereignis.key === zurueck ? -1 : 0

    if (richtung === 0) {
      return
    }

    ereignis.preventDefault()

    // **Kein zweiter Zweig für „gibt es nicht".** Der Index ist durch die
    // Modulorechnung immer im Bereich, und `gebiete` ist nicht leer — das
    // entscheidet der frühe Austritt im Bauteil. TypeScript weiß das nicht;
    // ein `if (… === undefined) return` wäre ein Zweig, den kein Aufruf
    // erreicht, und damit eine Behauptung darüber, was im Ernstfall geschieht
    // (§3).
    const naechste =
      gebiete[
        (((Math.max(stelle, 0) + richtung) % gebiete.length) + gebiete.length) %
          gebiete.length
      ]

    /* v8 ignore next 3 -- der Index ist durch die Modulorechnung immer im
       Bereich, und `gebiete` ist nicht leer: das entscheidet der frühe
       Austritt im Bauteil. TypeScript sieht das nicht. */
    if (naechste === undefined) {
      return
    }

    setzeGewaehlt(naechste.schluessel)

    // **Der Fokus wandert mit.** Ohne ihn bliebe er auf der alten Zeile, und
    // die nächste Pfeiltaste zählte von dort — man käme nie weiter als eins.
    leiste.current
      ?.querySelector<HTMLElement>(`[data-lasche="${naechste.schluessel}"]`)
      ?.focus()
  }

  return { offen, kennung, leiste, waehle: setzeGewaehlt, aufTaste }
}
