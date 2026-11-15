// **Wo jemand seine Karten hingelegt hat** — PLAY-04.AK-03
// (#1712).
//
// # Warum im Browserspeicher und nicht im Bestand
//
// **Weil eine Lage in Millimetern Darstellung ist und keine Regel.** CR
// 400.5 lässt die Anordnung frei; der Spielkern hat dazu nichts zu sagen,
// und ein Gegenüber bekommt davon nichts zu sehen. Wer sie in den Partiestand
// schriebe, machte aus einer Gewohnheit einen Spielzustand — und müsste sie
// über die Leitung schicken, versionieren und bei einem Konflikt auflösen.
//
// **Dieselbe Überlegung wie beim Tischbild** (`tisch.bild`): was dem Auge
// gehört, gehört nicht in die Partie.
//
// # Und warum je Partie
//
// **Weil eine Lage zu einer Karte an einem Tisch gehört.** Dieselbe Karte in
// der nächsten Partie liegt woanders, und ein gemeinsamer Speicher hätte
// Lagen von Tischen, die es nicht mehr gibt.
import { useCallback, useState } from 'react'

import type { Stelle } from '../tisch/eigenelagen'

// Lagenschluessel ist der Name, unter dem die Lagen einer Partie liegen.
export function lagenschluessel(partie: string): string {
  return `tisch.lagen.${partie}`
}

export interface Lagengedaechtnis {
  // lagen sind die abgelegten Stellen je Objektkennung.
  lagen: Readonly<Record<string, Stelle>>

  // lege hält fest, wo ein Objekt liegen soll.
  lege: (kennung: string, stelle: Stelle) => void
}

// gelesen holt die Lagen einer Partie aus dem Browserspeicher.
//
// **Mit `try` um jeden Griff.** In einem privaten Fenster, bei gesperrten
// Seitendaten und in einer Vorschau wirft schon der Zugriff — und eine
// Anwendung, die daran scheitert, verlöre den Tisch wegen einer Gewohnheit.
//
// **Und nur, was die Gestalt einer Lage hat.** Gültiges JSON in falscher
// Gestalt — `null`, eine Liste, eine Lage ohne `y` — ließ den Tisch sonst
// beim Zeichnen werfen (Nachlauf über UX-12, Befund 4); übergangen wird der
// Eintrag, die übrigen gelten.
function gelesen(partie: string): Record<string, Stelle> {
  try {
    const roh = localStorage.getItem(lagenschluessel(partie))

    return roh === null ? {} : nurLagen(JSON.parse(roh))
  } catch {
    return {}
  }
}

function nurLagen(wert: unknown): Record<string, Stelle> {
  if (typeof wert !== 'object' || wert === null || Array.isArray(wert)) {
    return {}
  }

  return Object.fromEntries(
    Object.entries(wert).filter((eintrag): eintrag is [string, Stelle] =>
      istStelle(eintrag[1]),
    ),
  )
}

function istStelle(wert: unknown): wert is Stelle {
  return (
    typeof wert === 'object' &&
    wert !== null &&
    typeof (wert as Record<string, unknown>).x === 'number' &&
    typeof (wert as Record<string, unknown>).y === 'number'
  )
}

// useEigeneLagen gibt die abgelegten Lagen einer Partie.
//
// **`ordnetSelbst` entscheidet, ob sie gelten** — und die Frage steht hier
// und nicht am Tisch: „welche Lagen gelten" ist **eine** Frage, und zwei
// Stellen, die sie beantworten, liefen auseinander (§10).
//
// **Gemerkt wird trotzdem weiter.** Wer die Anordnung einschaltet und wieder
// aus, findet seinen Tisch vor, wie er ihn gelegt hat — ein Schalter, der
// beim Umlegen Arbeit vernichtete, wäre eine Falle.
export function useEigeneLagen(
  partie: string,
  ordnetSelbst: boolean,
): Lagengedaechtnis {
  const [lagen, setzeLagen] = useState<Record<string, Stelle>>(() =>
    gelesen(partie),
  )

  const lege = useCallback(
    (kennung: string, stelle: Stelle) => {
      setzeLagen((bisher) => {
        const naechste = { ...bisher, [kennung]: stelle }

        try {
          localStorage.setItem(
            lagenschluessel(partie),
            JSON.stringify(naechste),
          )
        } catch {
          // **Ein Speicher, der nicht will, hält den Tisch nicht auf.** Die
          // Lage gilt dann für diese Sitzung — das ist weniger, als jemand
          // wollte, und mehr als nichts.
        }

        return naechste
      })
    },
    [partie],
  )

  return { lagen: ordnetSelbst ? {} : lagen, lege }
}
