// **Welches Ziel der Ausgabeassistent je Set zuletzt zeigte** — AP-10
// (#2367),
// Befund B-20. Getragen von der Einstellung `ausgabe.ziele`: eine Liste in
// JSON, die jüngste Wahl hinten, höchstens zweihundert — die Regel steht in
// `internal/domain/ausgabeziele.go`.
//
// # Warum ein Abgleich und kein Schreiben je Wahl
//
// **Weil `setze` je Einstellung einen Vorgang führt** (§7, Regel 4): eine
// zweite Wahl, während die erste geschrieben wird, wiese er still ab — und
// nach dem Neustart stünde die erste da. Gemerkt wird deshalb, was gewünscht
// ist; geschrieben wird es, sobald kein Schreiben läuft und der Bestand es
// noch nicht hat. **Ein gescheitertes Schreiben wird nicht wiederholt**,
// bis jemand neu wählt: sonst liefe gegen einen klemmenden Bestand eine
// Schleife.

import { useEffect, useRef, useState } from 'react'

import { ausgabeziele, type Ausgabeziel } from '../ausgabe'
import { useEinstellungen } from './useEinstellungen'

export const ausgabezielschluessel = 'ausgabe.ziele'

// hoechstens ist `domain.HoechstensGemerkteZiele`.
const hoechstens = 200

interface Gemerkt {
  set: string
  ziel: string
}

// gelesen gibt die Liste — leer, wenn sie fehlt oder nicht zu lesen ist.
// **Ein unlesbarer Wert wählt nichts vor**: der Bestand prüft beim Setzen,
// und was trotzdem dasteht, ist kein Grund, ein falsches Ziel zu zeigen.
function gelesen(wert: string): Gemerkt[] {
  try {
    const roh: unknown = JSON.parse(wert)

    return Array.isArray(roh)
      ? roh.filter(
          (eintrag): eintrag is Gemerkt =>
            typeof (eintrag as Partial<Gemerkt> | null)?.set === 'string' &&
            typeof (eintrag as Partial<Gemerkt>).ziel === 'string',
        )
      : []
  } catch {
    return []
  }
}

function istZiel(wert: string | undefined): wert is Ausgabeziel {
  return (ausgabeziele as readonly (string | undefined)[]).includes(wert)
}

export interface Ausgabezielgedaechtnis {
  // ziel ist das gemerkte Ziel dieses Sets — `undefined`, solange keines da
  // ist oder die Einstellungen noch laden.
  ziel: Ausgabeziel | undefined
  merke: (ziel: Ausgabeziel) => void

  // fehlschlag ist die Ursache, wenn das Merken scheiterte.
  fehlschlag: unknown
}

export function useAusgabeziel(setID: string): Ausgabezielgedaechtnis {
  const einstellungen = useEinstellungen()
  const gespeichert = einstellungen.wert(ausgabezielschluessel)

  // **Der Wunsch trägt sein Set** — ein Wechsel des Sets machte aus der Wahl
  // für A sonst eine für B (§7, Regel 1).
  const [wunsch, setzeWunsch] = useState<Gemerkt>()
  const versucht = useRef<string | undefined>(undefined)

  const { geladen, setze } = einstellungen
  const laeuft = einstellungen.laeuft(ausgabezielschluessel)

  useEffect(() => {
    // **Erst nach dem Laden**: eine Liste aus dem leeren Stand davor ließe
    // die Ziele aller anderen Sets fallen.
    if (!geladen || wunsch === undefined || laeuft) {
      return
    }

    const liste = gelesen(gespeichert)
    const kennung = `${wunsch.set}\u0000${wunsch.ziel}`
    if (
      liste.some(
        (eintrag) => eintrag.set === wunsch.set && eintrag.ziel === wunsch.ziel,
      ) ||
      versucht.current === kennung
    ) {
      return
    }

    versucht.current = kennung
    setze(
      ausgabezielschluessel,
      JSON.stringify(
        [
          ...liste.filter((eintrag) => eintrag.set !== wunsch.set),
          wunsch,
        ].slice(-hoechstens),
      ),
    )
  }, [geladen, wunsch, gespeichert, laeuft, setze])

  const eigenes =
    wunsch?.set === setID
      ? wunsch.ziel
      : gelesen(gespeichert).find((eintrag) => eintrag.set === setID)?.ziel

  // **Ein gleicher Wunsch schreibt nichts**: der Abgleich oben findet ihn
  // im Bestand oder unter `versucht`.
  const merke = (ziel: Ausgabeziel) => {
    setzeWunsch({ set: setID, ziel })
  }

  const fehlschlag =
    einstellungen.fehlschlag?.schluessel === ausgabezielschluessel
      ? einstellungen.fehlschlag.ursache
      : undefined

  return { ziel: istZiel(eigenes) ? eigenes : undefined, merke, fehlschlag }
}
