// **Eine Geste ausführen, nur in dem Stand, aus dem sie kommt** — GESTE
// (#2358).
//
// Der Vertrag steht in `gestenvertrag.ts`; hier wird er bedient: die Tasten
// gedeutet, die Ansage gesetzt, die Absicht ausgeführt. **Und bewacht** —
// AGENTS.md §7:
//
// - **Regel 1**: wechselt Profil, Epoche oder Besetzung, gilt ein offener
//   Griff nicht mehr. Er wird nicht weggeräumt, sondern **gilt nicht** — der
//   Zustand wird aus dem Stand abgeleitet, statt in einem Effekt
//   nachgezogen.
// - **Regel 2**: eine Antwort gehört zu ihrem Stand. Kommt sie, nachdem
//   Profil oder Epoche gewechselt haben oder die Geste abgebrochen wurde,
//   ruft sie nichts zurück — weder Erfolg noch Fehler.
// - **Vor dem Ausführen**: eine Absicht aus einem vorigen Stand geht gar
//   nicht erst hinaus, und eine auf einen Sitz, den es nicht mehr gibt, auch
//   nicht. Was der Kern dazu sagt, entscheidet er (Vermerk 0126); hier geht
//   es nur darum, dass nichts Fremdes geschieht.
import { useCallback, useEffect, useRef, useState } from 'react'

import {
  gestentaste,
  greifen as greifenIm,
  type Gestenabsicht,
  type Gestenansage,
  type Gestenbezug,
  type Gestenzustand,
  type Griff,
} from '../bausteine/gestenvertrag'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { useLauf } from './useLauf'

// Gestenstand ist der Bezug samt den Sitzen, die es gerade gibt — am
// Markttisch keine.
export interface Gestenstand extends Gestenbezug {
  sitze?: readonly string[]
}

interface Gestenoptionen {
  stand: Gestenstand
  ausfuehren: (absicht: Gestenabsicht) => Promise<void>
  nachher?: (absicht: Gestenabsicht) => void
  fehler?: (fehler: unknown, absicht: Gestenabsicht) => void
}

export interface Geste {
  zustand: Gestenzustand
  // ansage ist der Satz für die Live-Region.
  ansage: string
  greifen: (griff: Griff) => void
  // taste deutet eine Taste und sagt, ob sie die Geste betraf.
  taste: (name: string) => boolean
  // ablegen nimmt eine Absicht von Maus oder Finger (`absichtAusZug`).
  ablegen: (absicht: Gestenabsicht) => void
  // abbrechen lässt die Antwort einer laufenden Geste ins Leere gehen.
  abbrechen: () => void
}

const ruhend: Gestenzustand = { art: 'ruhend' }

// standSchluessel fasst, woran ein Griff hängt.
function standSchluessel(stand: Gestenstand): string {
  return `${stand.profil}\n${String(stand.epoche)}\n${(stand.sitze ?? []).join(',')}`
}

export function useGeste({
  stand,
  ausfuehren,
  nachher,
  fehler,
}: Gestenoptionen): Geste {
  const eigene = useTexte().bausteine.geste
  const lauf = useLauf()
  const [griff, setzeGriff] = useState<{
    zustand: Gestenzustand
    schluessel: string
  }>({
    zustand: ruhend,
    schluessel: '',
  })
  const [ansage, setzeAnsage] = useState('')

  // **Der heutige Stand für die Antwort**, die erst später kommt.
  // **Nachgezogen in einem Effekt**, nicht im Zeichnen (`react-hooks/refs`):
  // gelesen wird er erst in einem Rückruf, und der kommt nach dem Effekt.
  const heute = useRef(stand)
  useEffect(() => {
    heute.current = stand
  })

  const zustand =
    griff.schluessel === standSchluessel(stand) ? griff.zustand : ruhend

  const sagen = useCallback(
    (inhalt: Gestenansage) => {
      const { schluessel, ...werte } = inhalt
      setzeAnsage(
        einsetzen(
          eigene[schluessel],
          Object.fromEntries(
            Object.entries(werte).map(([name, wert]) => [name, String(wert)]),
          ),
        ),
      )
    },
    [eigene],
  )

  const ablegen = useCallback(
    (absicht: Gestenabsicht) => {
      const jetzt = heute.current
      if (absicht.profil !== jetzt.profil || absicht.epoche !== jetzt.epoche) {
        sagen({ schluessel: 'veraltet' })

        return
      }
      if (
        absicht.zielSitz !== undefined &&
        jetzt.sitze !== undefined &&
        !jetzt.sitze.includes(absicht.zielSitz)
      ) {
        sagen({ schluessel: 'zielFehlt' })

        return
      }

      const nummer = lauf.naechster()
      const gilt = () => {
        const dann = heute.current

        return (
          lauf.istAktuell(nummer) &&
          dann.profil === absicht.profil &&
          dann.epoche === absicht.epoche
        )
      }
      ausfuehren(absicht).then(
        () => {
          if (gilt()) nachher?.(absicht)
        },
        (grund: unknown) => {
          if (gilt()) fehler?.(grund, absicht)
        },
      )
    },
    [ausfuehren, nachher, fehler, lauf, sagen],
  )

  const greifen = useCallback(
    (neu: Griff) => {
      const schritt = greifenIm(neu, heute.current)
      setzeGriff({
        zustand: schritt.zustand,
        schluessel: standSchluessel(heute.current),
      })
      sagen(schritt.ansage)
    },
    [sagen],
  )

  const taste = useCallback(
    (name: string) => {
      const schritt = gestentaste(zustand, name)
      if (schritt === undefined) {
        return false
      }
      setzeGriff({
        zustand: schritt.zustand,
        schluessel: standSchluessel(heute.current),
      })
      sagen(schritt.ansage)
      if (schritt.absicht !== undefined) {
        ablegen(schritt.absicht)
      }

      return true
    },
    [zustand, sagen, ablegen],
  )

  const abbrechen = useCallback(() => {
    lauf.naechster()
  }, [lauf])

  return { zustand, ansage, greifen, taste, ablegen, abbrechen }
}
