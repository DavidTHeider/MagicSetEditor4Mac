// **Ein Schritt an der Sitzung** — ID-02
// (#1701).
//
// **Was die drei Flächen der Sitzung gemeinsam tun**: einen Ruf fahren, einen
// Vorgang zur Zeit (§7, Regel 4), den Fehler **am Feld** zeigen, das Go nennt
// (§6), und danach die Sitzung neu lesen lassen.
import { createContext, useContext, useState } from 'react'

import { Bindungsfehler } from '../bindungsfehler'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { useVorgang } from './useVorgang'

import type { SchreibAntwort } from '../api'

// Sitzungsraum reicht den Weg herunter, die Sitzung neu zu lesen — nach
// Sperren, Abmelden und Anmelden.
export const Sitzungsraum = createContext<() => void>(() => undefined)

export function useSitzungNeuLesen() {
  return useContext(Sitzungsraum)
}

// Die Felder, an denen Go einen Fehler der Sitzung nennt.
const felder = new Set(['kennwort', 'neu', 'faktor', 'sensor'])

export function useSitzungsschritt(danach: () => void) {
  const texte = useTexte()
  const vorgang = useVorgang()
  const [befund, setzeBefund] = useState<Record<string, string>>({})
  const [fehler, setzeFehler] = useState<string>()
  const [faktorGefragt, setzeFaktorGefragt] = useState(false)

  const fuehre = (ruf: () => Promise<SchreibAntwort>) =>
    vorgang.starte('sitzung', async () => {
      setzeBefund({})
      setzeFehler(undefined)

      try {
        const antwort = await ruf()
        auspacken(antwort, () => antwort)
        danach()
      } catch (grund) {
        const meldung = meldungAus(grund, texte)

        // **Ein fehlender zweiter Faktor ist ein fehlender Schritt**, keine
        // Abweisung — wie beim Konto (`Kontoanlage.tsx`).
        if (
          grund instanceof Bindungsfehler &&
          grund.schluessel === 'zweiterFaktor'
        ) {
          setzeFaktorGefragt(true)

          return
        }

        if (
          grund instanceof Bindungsfehler &&
          grund.feld !== undefined &&
          felder.has(grund.feld)
        ) {
          setzeBefund({ [grund.feld]: meldung })

          return
        }

        setzeFehler(meldung)
      }
    })

  return {
    fuehre,
    befund,
    fehler,
    faktorGefragt,
    laeuft: vorgang.laeuft('sitzung'),
  }
}
