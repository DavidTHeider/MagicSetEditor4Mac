// **Das Fenster spricht eine eigene Übersetzung** — LOC-02.AK-03
// (#1688).
//
// # Die Rückfallkette
//
// **Eigene Übersetzung, dann ihre Basis, dann Deutsch** — eine leere
// Beschriftung gibt es damit nicht. Die Basis ist ein vollständiger Katalog
// (das hält `zusatzsprachen.test.ts`), und die Sätze der Übersetzung liegen
// darüber. Gibt es das Paket nicht mehr oder lässt es sich nicht laden,
// spricht das Fenster die gewählte Sprache — zurückgefallen, nicht leer.
import { useCallback, useEffect, useMemo, useState } from 'react'

import { api } from '../fassade'
import { aufAenderung } from '../sprache/eigenesprache'
import { istSprache } from '../sprache/einstellung'
import { kataloge, type Katalog, type Sprache } from '../sprache/kataloge'
import { ueberlagere } from '../sprache/flach'
import { auspacken } from '../umschlag'
import { useLaden } from './useLaden'

export interface EigeneSprache {
  kennung: string
  basis: Sprache
  katalog: Katalog
}

// useEigeneSprache lädt die Übersetzung, die `sprache.eigene` nennt.
export function useEigeneSprache(kennung: string): EigeneSprache | undefined {
  const [runde, setzeRunde] = useState(0)

  // **Nachladen, wenn der Editor diese geändert hat** — die Sätze sollen im
  // Fenster stehen, ohne dass jemand die Sprache neu wählt.
  useEffect(
    () =>
      aufAenderung((geaendert) => {
        if (geaendert === kennung) {
          setzeRunde((zahl) => zahl + 1)
        }
      }),
    [kennung],
  )

  const geladen = useLaden(
    useCallback(async () => {
      if (kennung === '') {
        return undefined
      }

      const antwort = await api().Sprachpaket(kennung)

      return auspacken(antwort, () => antwort)
    }, [kennung]),
    [kennung, runde],
    true,
  )

  const paket = geladen.fehler === undefined ? geladen.daten : undefined

  return useMemo(() => {
    if (
      kennung === '' ||
      paket?.paket.kennung !== kennung ||
      !istSprache(paket.paket.basis)
    ) {
      return undefined
    }

    const saetze = Object.fromEntries(
      Object.entries(paket.texte).map(([pfad, text]) => [pfad, text.text]),
    )

    return {
      kennung,
      basis: paket.paket.basis,
      katalog: ueberlagere(kataloge[paket.paket.basis], saetze, kennung),
    }
  }, [kennung, paket])
}
