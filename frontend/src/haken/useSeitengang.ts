// **Zurück und Vorwärts zwischen den Seiten** — UX-03.AK-02
// (#1675):
// *„Zurück/Vorwärts, Direktaufruf und Wiederaufnahme funktionieren ohne
// Verlust ungespeicherter Eingaben."*
//
// # Der Gang
//
// **Die Seiten dahinter, die aktuelle und die davor** — die Form eines
// Browserverlaufs: wer zurückgeht und dann eine andere Seite aufruft,
// verwirft, was vor ihm lag. Die Kürzel sind die des Finders für den vorigen
// und den nächsten Ordner, ⌘[ und ⌘].
//
// # Und das offene Feld
//
// **Ein Schritt verlässt es zuerst** (`feldwegschreiben.ts`, #1290). Ein
// Kartenfeld schreibt beim Verlassen; ein Kürzel wechselt die Seite, ohne
// dass ein `blur` käme — und der zuletzt getippte Wert wäre fort. Bis hierher
// ging nur der Wechsel des **Sets** durch diese Stelle, der Wechsel der
// **Seite** nicht.
import {
  useCallback,
  useLayoutEffect,
  useRef,
  useState,
  type SetStateAction,
} from 'react'

import { wegschreiben } from '../feldwegschreiben'
import { useFokusNachSeitenwechsel } from './useFokusNachSeitenwechsel'
import { istSetseite, type Seitenschluessel } from '../navigation'

export interface Gangslage {
  // zurueck sind die Seiten dahinter, die älteste zuerst.
  zurueck: readonly Seitenschluessel[]
  aktuell: Seitenschluessel

  // vor sind die Seiten davor, die nächste zuerst.
  vor: readonly Seitenschluessel[]
}

// hoechstensGemerkt begrenzt den Gang. **Fünfzig reichen für einen Tag**, und
// eine Liste ohne Grenze wüchse in einer Sitzung, die niemand beendet.
export const hoechstensGemerkt = 50

// gehe ruft eine Seite auf und verwirft, was vor der aktuellen lag.
export function gehe(lage: Gangslage, seite: Seitenschluessel): Gangslage {
  if (lage.aktuell === seite) {
    return lage
  }

  return {
    zurueck: [...lage.zurueck, lage.aktuell].slice(-(hoechstensGemerkt - 1)),
    aktuell: seite,
    vor: [],
  }
}

// schreite geht einen Schritt zurück oder vor — und bleibt am Rand stehen.
export function schreite(lage: Gangslage, schritt: -1 | 1): Gangslage {
  if (schritt === -1) {
    const vorige = lage.zurueck.at(-1)

    return vorige === undefined
      ? lage
      : {
          zurueck: lage.zurueck.slice(0, -1),
          aktuell: vorige,
          vor: [lage.aktuell, ...lage.vor],
        }
  }

  const naechste = lage.vor.at(0)

  return naechste === undefined
    ? lage
    : {
        zurueck: [...lage.zurueck, lage.aktuell],
        aktuell: naechste,
        vor: lage.vor.slice(1),
      }
}

// ohneSetseiten räumt die Seiten eines Sets aus dem Gang — **bis auf die, auf
// der man steht**.
//
// **Ein Wechsel des Bezugsobjekts räumt den offenen Kontext** (AGENTS.md §7,
// Regel 1), und der Gang gehört dazu: *Karten* hieß vorhin „die Karten von
// Set A". Wer nach einem Wechsel zu Set B zurückginge, landete sonst auf den
// Karten von B — auf einer Seite, auf der er nie war.
//
// **Zwei gleiche Nachbarn fallen zusammen** — auch über die aktuelle hinweg;
// sonst führte ein Schritt auf die Seite, auf der man schon steht.
export function ohneSetseiten(lage: Gangslage): Gangslage {
  const zurueck = ohneWiederholung([
    ...lage.zurueck.filter((seite) => !istSetseite(seite)),
    lage.aktuell,
  ]).slice(0, -1)
  const vor = ohneWiederholung([
    lage.aktuell,
    ...lage.vor.filter((seite) => !istSetseite(seite)),
  ]).slice(1)

  return { zurueck, aktuell: lage.aktuell, vor }
}

function ohneWiederholung(
  seiten: readonly Seitenschluessel[],
): Seitenschluessel[] {
  return seiten.filter((seite, index) => seite !== seiten[index - 1])
}

export interface Seitengang {
  aktuell: Seitenschluessel

  // wechseln ruft eine Seite auf — der Direktaufruf, gleich woher. **Auch
  // mit einer Rechnung** über die aktuelle Seite, wie `setState`: eine Karte
  // aufzuschlagen führt nur dann zu den Karten, wenn keine Schublade offen ist.
  wechseln: (ziel: SetStateAction<Seitenschluessel>) => void

  // zurueck und vor liegen hinter ⌘[ und ⌘] und unter *Darstellung*.
  zurueck: () => void
  vor: () => void
}

// useSeitengang führt die aktuelle Seite samt ihrem Gang.
//
// **Er ersetzt den bloßen Zustand der Seite**, und das ist Absicht: gäbe es
// neben ihm einen zweiten Weg, die Seite zu setzen, fehlte dessen Schritt im
// Gang.
export function useSeitengang(setID: string | undefined): Seitengang {
  const [lage, setzeLage] = useState<Gangslage>({
    zurueck: [],
    aktuell: 'start',
    vor: [],
  })
  const [setVorher, setzeSetVorher] = useState(setID)

  // **Der Vergleich steht neben dem Zustand** — dieselbe Form wie im
  // `Anschlussraum`.
  if (setVorher !== setID) {
    setzeSetVorher(setID)
    setzeLage(ohneSetseiten(lage))
  }

  // **Die Lage für die Rückrufe** — nachgeführt nach jedem Zeichnen, damit
  // die drei Rückrufe **dieselben** bleiben. Wechselten sie mit jeder Seite,
  // liefe jeder Effekt, der von ihnen abhängt, bei jedem Wechsel neu.
  const lageJetzt = useRef(lage)
  useLayoutEffect(() => {
    lageJetzt.current = lage
  })

  // **Das Feld wird nur verlassen, wenn die Seite wirklich wechselt.** Eine
  // Rechnung kommt aus der Anwendung und nicht von einer Taste — die
  // Kartenliste meldet ihre Wahl auch, während jemand in einer Schublade
  // tippt, und ein `blur` dort nähme ihm das Feld unter den Fingern weg.
  const wechseln = useCallback((ziel: SetStateAction<Seitenschluessel>) => {
    if (typeof ziel === 'function') {
      setzeLage((vorher) => gehe(vorher, ziel(vorher.aktuell)))

      return
    }

    if (ziel !== lageJetzt.current.aktuell) {
      wegschreiben()
    }
    setzeLage((vorher) => gehe(vorher, ziel))
  }, [])

  // **Ein Schritt ins Leere ist keiner** — er verlässt auch kein Feld.
  const schreiten = useCallback((schritt: -1 | 1) => {
    if (schreite(lageJetzt.current, schritt) !== lageJetzt.current) {
      wegschreiben()
      setzeLage((vorher) => schreite(vorher, schritt))
    }
  }, [])

  const zurueck = useCallback(() => {
    schreiten(-1)
  }, [schreiten])

  const vor = useCallback(() => {
    schreiten(1)
  }, [schreiten])

  // **Und der Fokus, der mit der Seite ging, kommt im Inhalt an** (UX-10.AK-02,
  // #1682) — sonst fand Tab in WebKit keinen Weg mehr.
  useFokusNachSeitenwechsel(lage.aktuell)

  return { aktuell: lage.aktuell, wechseln, zurueck, vor }
}
