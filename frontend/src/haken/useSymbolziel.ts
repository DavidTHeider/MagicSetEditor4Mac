// **Welches Symbol der Editor bearbeitet** — und dass es sein Set nicht
// überlebt.
//
// # Warum ein Haken und kein Zustand in `Anwendung.tsx`
//
// Weil zum Zustand eine **Regel** gehört (§7, Regel 1): ein Wechsel des
// Bezugsobjekts räumt ihn. Diese Regel stand als Effekt neben dem Zustand,
// und eine Gegenprobe darauf blieb grün — geprüft waren die reinen
// Funktionen, nicht die Verbindung. Hier stehen beide zusammen und sind
// zusammen prüfbar (Nachlauf über Phase 4.11).

import { useState } from 'react'

import {
  geraeumtesZiel,
  zielImSet,
  type Symbolziel,
  type Symbolzielstand,
} from '../seiten/symbolablegen'

export interface Symbolzielsteuerung {
  // ziel ist, was der Editor jetzt bearbeitet — `undefined` heißt Setsymbol.
  ziel: Symbolziel | undefined

  // oeffne trägt ein Kartensymbol ein.
  oeffne: (karteID: string, feld: string) => void

  // schliesse geht zurück zum Setsymbol.
  schliesse: () => void
}

export function useSymbolziel(setID: string | undefined): Symbolzielsteuerung {
  const [stand, setzeStand] = useState<Symbolzielstand>()
  const [voriges, setzeVoriges] = useState(setID)

  // **Ein fremdes Ziel wird geräumt, nicht nur verborgen.**
  //
  // `zielImSet` blendet es aus; ohne das Räumen kehrte es zurück, sobald
  // dieselbe Set-Kennung wieder aufgeschlagen wurde — nach einem
  // Zurückspielen des Bestands womöglich für eine **andere** Karte
  // (Nachlauf über Phase 4.11).
  //
  // **Beim Zeichnen und nicht im Effekt.** Ein Effekt zeichnete erst einmal
  // mit dem alten Ziel und räumte danach — React nennt das eine kaskadierende
  // Anzeige und der Prüfer weist es ab. Dies ist die Form, die React für
  // „Zustand an eine geänderte Eigenschaft anpassen" vorsieht: der vorige
  // Wert steht daneben, und der Vergleich entscheidet.
  if (voriges !== setID) {
    setzeVoriges(setID)
    setzeStand((bisher) => geraeumtesZiel(bisher, setID))
  }

  return {
    ziel: zielImSet(stand, setID),

    oeffne: (karteID: string, feld: string) => {
      setzeStand(zielFuerStand(setID, karteID, feld))
    },

    schliesse: () => {
      setzeStand(undefined)
    },
  }
}

// zielFuerStand baut den Stand — **wenn ein Set offen ist**.
//
// **Ohne offenes Set gibt es keines.** Über die Oberfläche kommt der Fall
// nicht vor: ein Symbolfeld steht auf einer Karte, und eine Karte steht in
// einem Set. Er steht hier, weil der Typ ihn verlangt.
function zielFuerStand(
  setID: string | undefined,
  karteID: string,
  feld: string,
): Symbolzielstand | undefined {
  if (setID === undefined) {
    return undefined
  }

  return { setID, karteID, feld }
}
