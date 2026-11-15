// **Was der Editor einer eigenen Übersetzung tut** — LOC-02
// (#1688).
//
// # Ein Vorgang je Paket, nicht je Satz
//
// **Die Sätze stehen in einer Zeile.** Zwei Änderungen zugleich hält die
// Ablage aus (sie ändert in einem Zug); im Fenster aber gilt §7, Regel 4 —
// solange eine läuft, wartet die nächste, und die Antwort trägt den ganzen
// neuen Stand. Nachgeladen wird deshalb nicht: die Antwort **ist** der
// Stand.
//
// **Den Wechsel des Pakets räumt der `key`** in `Eigenuebersetzungen.tsx`:
// ein anderes Paket ist ein anderes Bauteil (Regeln 1 und 2).
import { useCallback, useState } from 'react'

import { api } from '../fassade'
import { Bindungsfehler } from '../bindungsfehler'
import { eigeneSpracheGeaendert } from '../sprache/eigenesprache'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { useLaden } from './useLaden'
import { useVorgang } from './useVorgang'

import type { SprachpaketAntwort, SprachpaketangabenDTO } from '../apisprache'
import type { Sprachbefund } from './useSprachpakete'

// ohneAngaben lässt alle Angaben, wie sie sind.
export const ohneAngaben: SprachpaketangabenDTO = {
  name: null,
  stand: null,
  lizenz: null,
  herkunft: null,
}

export function useUebersetzung(kennung: string) {
  const texte = useTexte()
  const vorgang = useVorgang()
  const [geschrieben, setzeGeschrieben] = useState<SprachpaketAntwort>()
  const [befund, setzeBefund] = useState<Sprachbefund>()

  const geladen = useLaden(
    useCallback(async () => {
      const antwort = await api().Sprachpaket(kennung)

      return auspacken(antwort, () => antwort)
    }, [kennung]),
    [kennung],
  )

  // aendere setzt Angaben und Sätze — und meldet dem Fenster, dass seine
  // Sprache neue Sätze hat.
  const aendere = async (
    angaben: SprachpaketangabenDTO,
    saetze: Record<string, string>,
  ) => {
    setzeBefund(undefined)

    await vorgang.starte(kennung, async () => {
      try {
        const antwort = await api().SprachpaketAendern(kennung, angaben, saetze)

        setzeGeschrieben(auspacken(antwort, () => antwort))
        eigeneSpracheGeaendert(kennung)
      } catch (ursache) {
        setzeBefund({
          feld: ursache instanceof Bindungsfehler ? (ursache.feld ?? '') : '',
          text: meldungAus(ursache, texte),
        })
      }
    })
  }

  return {
    stand: geschrieben ?? geladen.daten,
    ladefehler: geladen.fehler,
    befund,
    laeuft: vorgang.laeuft(kennung),
    aendere,
  }
}
