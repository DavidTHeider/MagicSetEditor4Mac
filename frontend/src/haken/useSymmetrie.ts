// **Die Symmetrie im Symboleditor**
// (#1092).
//
// # Warum ein Haken und nicht vier Aufrufe in der Seite
//
// Weil `Symbol.tsx` an der Zeilengrenze aus §2 steht und weil der Schnitt der
// Frage folgt: *was* eine Symmetrie tut, steht hier; *wie* sie aussieht, in
// `Symmetriefeld.tsx` und auf der Zeichenfläche.
//
// # Die Kopien kommen von der Go-Seite
//
// Die Rechnung dahinter hat drei Fallen, die man ohne den C++-Quelltext
// verfehlt — die Multiplikationsrichtung, die Spiegelachse bei `(a + b)/2`,
// das Abrunden auf eine gerade Zahl. Sie steht **einmal** in `service`; eine
// zweite Fassung hier liefe auseinander, und dann zeichnete der Editor etwas
// anderes als der Rasterer.

import { useCallback, useEffect, useState } from 'react'

import type {
  SymbolAntwort,
  SymbolteilDTO,
  SymbolzugDTO,
  SymmetrieweiseDTO,
} from '../antworten'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import type { Handgriff } from './useZeichenverlauf'
import type { Zeichenschutz } from './useZeichenschutz'
import { useLauf } from './useLauf'

interface SymmetrieEigenschaften {
  teile: SymbolteilDTO[]
  gewaehlt: number
  eintragen: (teile: SymbolteilDTO[], handgriff: Handgriff) => void

  // schutz hält §7 Regeln 2 und 4 über der Zeichnung (Befund 2 des
  // Zweitblicks über Phase 4.11).
  schutz: Zeichenschutz
  klagen: (ursache: unknown) => void
}

// keine ist die leere Kopienliste — **eine einzige, wiederverwendete**.
//
// # Warum das keine Kleinigkeit ist
//
// `setzeKopien([])` legte bei jedem Zeichnen ein **neues** Array an. React
// vergleicht mit `Object.is`; ein neues Array ist nie dasselbe, also folgte ein
// weiteres Zeichnen — und weil `gezeigte` in der Seite mit `?? []` ebenfalls
// jedes Mal neu entstand, lief der Effekt gleich mit. Ein Kreis, der sich
// selbst antreibt.
//
// **Die laufende Anwendung hätte dasselbe getan**: solange kein Symbol geladen
// ist, dreht sich der Editor im Kreis. Gefunden hat es der Testlauf, der nicht
// zurückkam — nicht ein Blick in den Code (§3).
const keine: SymbolzugDTO[][] = []

export interface Symmetriewege {
  // kopien sind die Züge, die die gewählte Symmetrie zeichnet — leer, wenn das
  // gewählte Teil keine ist.
  kopien: SymbolzugDTO[][]

  anlegen: () => void
  aufloesen: () => void
  setzen: (weise: Omit<SymmetrieweiseDTO, 'teil'>) => void

  // mitteAuf und griffAuf sind die beiden Anfasser auf der Zeichenfläche.
  //
  // **Sie stehen hier und nicht in der Seite**, weil sie dieselbe Gestalt
  // haben: acht Felder, von denen zwei gelten. Zweimal ausgeschrieben wären
  // sie zwei Stellen, an denen ein vergessenes `mitteSetzen` die Mitte in die
  // linke obere Ecke zöge.
  mitteAuf: (x: number, y: number) => void
  griffAuf: (x: number, y: number) => void
}

// unveraendert ist ein Aufruf, der nichts ändert — die Grundlage für alle
// Aufrufe, die genau eines ändern.
const unveraendert: Omit<SymmetrieweiseDTO, 'teil'> = {
  spiegelart: '',
  kopien: 0,
  mitteX: 0,
  mitteY: 0,
  griffX: 0,
  griffY: 0,
  mitteSetzen: false,
  griffSetzen: false,
}

export function useSymmetrie({
  teile,
  gewaehlt,
  eintragen,
  klagen,
  schutz,
}: SymmetrieEigenschaften): Symmetriewege {
  const [kopien, setzeKopien] = useState<SymbolzugDTO[][]>([])
  const lauf = useLauf()

  const gewaehltesTeil = teile[gewaehlt]
  const istSymmetrie = gewaehltesTeil?.art === 'Symmetrie'

  // **Die Kopien werden nach jeder Änderung neu geholt**, weil jede von ihnen
  // sie verschiebt: eine andere Art, eine Kopie mehr, eine verschobene Mitte.
  //
  // **Und der Lauf zählt mit** (§7, Regel 2). Wer den Griff zieht, schickt
  // eine Anfrage je Bewegung; sie kommen über dieselbe Naht zurück, aber nicht
  // zwingend in der Reihenfolge, in der sie hinausgingen.
  useEffect(() => {
    if (!istSymmetrie) {
      return
    }

    const nummer = lauf.naechster()

    void (async () => {
      try {
        const antwort = await api().Symmetriekopien(teile, gewaehlt)
        const neue = auspacken(antwort, () => antwort.kopien)

        if (lauf.istAktuell(nummer)) {
          setzeKopien(neue)
        }
      } catch (ursache: unknown) {
        if (lauf.istAktuell(nummer)) {
          klagen(ursache)
        }
      }
    })()
  }, [teile, gewaehlt, istSymmetrie, lauf, klagen])

  // aendern ruft eine der drei schreibenden Bindungen und trägt das Ergebnis
  // ein.
  //
  // **Ein Fehler räumt die Zeichnung nicht** (§7, Regel 3): wer eine
  // Kopienzahl außerhalb der Spanne setzt, bekommt eine Meldung und behält,
  // was er hat.
  const aendern = useCallback(
    (ruf: () => Promise<SymbolAntwort>, marke: string) => {
      void schutz.starte(async (nochAktuell) => {
        try {
          const antwort = await ruf()
          const neue = auspacken(antwort, () => antwort.teile)

          // **Eine überholte Antwort trägt sich nicht ein** (§7, Regel 2).
          if (!nochAktuell()) {
            return
          }

          eintragen(neue, { marke })
        } catch (ursache: unknown) {
          klagen(ursache)
        }
      })
    },
    [eintragen, klagen, schutz],
  )

  const anlegen = useCallback(() => {
    aendern(() => api().SymmetrieAnlegen(teile, gewaehlt), 'symmetrie:anlegen')
  }, [aendern, teile, gewaehlt])

  const aufloesen = useCallback(() => {
    aendern(
      () => api().SymmetrieAufloesen(teile, gewaehlt),
      'symmetrie:aufloesen',
    )
  }, [aendern, teile, gewaehlt])

  const setzen = useCallback(
    (weise: Omit<SymmetrieweiseDTO, 'teil'>) => {
      aendern(
        () => api().SymmetrieSetzen(teile, { ...weise, teil: gewaehlt }),
        `symmetrie:${String(gewaehlt)}`,
      )
    },
    [aendern, teile, gewaehlt],
  )

  const mitteAuf = useCallback(
    (x: number, y: number) => {
      setzen({ ...unveraendert, mitteX: x, mitteY: y, mitteSetzen: true })
    },
    [setzen],
  )

  const griffAuf = useCallback(
    (x: number, y: number) => {
      setzen({ ...unveraendert, griffX: x, griffY: y, griffSetzen: true })
    },
    [setzen],
  )

  // **Der leere Fall wird abgeleitet, nicht gesetzt.**
  //
  // Der erste Anlauf rief `setzeKopien(keine)` im Effekt — und `eslint` weist
  // das mit `react-hooks/set-state-in-effect` ab, zu Recht: es ist dieselbe
  // Klasse wie der Kreis darüber. Ein Wert, der sich aus dem Zustand ergibt,
  // gehört nicht in den Zustand.
  return {
    kopien: istSymmetrie ? kopien : keine,
    anlegen,
    aufloesen,
    setzen,
    mitteAuf,
    griffAuf,
  }
}
