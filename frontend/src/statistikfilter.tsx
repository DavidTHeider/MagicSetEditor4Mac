// **Der Filter, den die Statistik über die Kartenliste legt** — UXR-F6
// (#2013).
//
// # Was MSE2 hier tut
//
// Wer im Diagramm eine Gruppe wählt, filtert die Kartenliste darunter auf
// ihre Karten (`StatsPanel::onGraphSelect` → `filterCards`,
// `src/gui/set/stats_panel.cpp` Zeilen 526–547, Fassung f9d9356 vom
// 2021-07-06). Dort stehen Diagramm und Liste im selben Bereich; hier stehen
// sie in zwei Flächen — die Statistik in der Schublade *Werkzeuge*, die Liste
// auf den Karten darunter. Diese Datei ist die Naht dazwischen: **eine Wahl,
// zwei Leser**.
//
// # Warum ein Raum und keine Eigenschaft
//
// Die Wahl reist von der Statistik zur Kartenliste und zurück, und dazwischen
// liegen `Setbereich`, `Lascheninhalt` und `Karten` — vier Ebenen, die nichts
// damit zu tun haben. Der Raum steht in `Setbereich` und damit **je Set**: ein
// anderes Set baut ihn neu, und die Wahl geht mit (§7, Regel 1).
//
// **Ohne Raum gilt kein Filter**, und Wählen tut nichts. Das ist die
// Vorgabe für jede Fläche, die ihn nicht braucht — und der Grund, warum die
// Kartenliste auch ohne Statistik nach Plan arbeitet.
import {
  createContext,
  useContext,
  useMemo,
  useState,
  type ReactNode,
} from 'react'

import type { Seitenschluessel } from './navigation'

// Gruppenwahl ist der Balken, der die Kartenliste gerade filtert.
export interface Gruppenwahl {
  // wahl ist der Eintrag „Worüber", zu dem der Balken gehört
  // (`dimension:<name>` oder `feld:<name>`). **Eine Wahl gilt für das Bild,
  // aus dem sie stammt** (§7, Regel 2): wechselt das Bild, endet sie.
  wahl: string

  // quelle ist der Name der Dimension oder des Felds, gruppe der des
  // Balkens — beide stehen in der Zeile über der Kartenliste.
  quelle: string
  gruppe: string

  // kartenIDs sind die Karten des Balkens, wie die Auszählung sie nannte.
  kartenIDs: readonly string[]
}

export interface Statistikfilter {
  wahl: Gruppenwahl | undefined
  waehlen: (wahl: Gruppenwahl) => void
  aufheben: () => void

  // zurStatistik und zurKartenliste sind der Hin- und der Rückweg: die
  // Schublade schließen oder öffnen. Wer den Filter setzt, soll nicht wissen
  // müssen, wo die Seiten stehen.
  zurStatistik: () => void
  zurKartenliste: () => void
}

// nichts ist das Ganze der Vorgabe: ein einziger Griff, der nichts tut.
function nichts(): void {
  // ohne Raum gilt kein Filter
}

const ohneFilter: Statistikfilter = {
  wahl: undefined,
  waehlen: nichts,
  aufheben: nichts,
  zurStatistik: nichts,
  zurKartenliste: nichts,
}

const Filterraum = createContext<Statistikfilter>(ohneFilter)

interface RaumEigenschaften {
  // wechseln führt auf eine andere Seite des Sets.
  wechseln: (seite: Seitenschluessel) => void
  children: ReactNode
}

// Statistikfilterraum hält die Wahl und reicht sie an beide Flächen.
export function Statistikfilterraum({ wechseln, children }: RaumEigenschaften) {
  const [wahl, setzeWahl] = useState<Gruppenwahl>()

  const filter = useMemo<Statistikfilter>(
    () => ({
      wahl,
      waehlen: setzeWahl,
      aufheben: () => {
        setzeWahl(undefined)
      },
      zurStatistik: () => {
        wechseln('statistik')
      },
      zurKartenliste: () => {
        wechseln('karten')
      },
    }),
    [wahl, wechseln],
  )

  return <Filterraum.Provider value={filter}>{children}</Filterraum.Provider>
}

// useStatistikfilter gibt den Filter dieses Sets — oder den, der nichts tut.
export function useStatistikfilter(): Statistikfilter {
  return useContext(Filterraum)
}
