// **Die Kartenskripte eines Sets, im Fenster** — Werkstück B von Phase 4.11
// (#1337).
//
// **Ein eigener Haken**, weil er drei Dinge zugleich hält: das Angebot, den
// laufenden Vorgang und den Bericht des letzten Laufs. In `Karten.tsx` wären
// das drei weitere Zustände in einer Datei, die schon an der Zeilengrenze
// steht (§2).
import { useCallback, useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { berichttext } from '../seiten/Kartenskripte'
import { meldungAus } from '../meldung'
import { useLaden } from './useLaden'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from './useVorgang'
import type { KartenskriptDTO } from '../antwortenkartenskripte'

export interface Kartenskriptgriff {
  skripte: readonly KartenskriptDTO[]
  laeuft: boolean
  bericht: string | undefined
  befunde: readonly string[]
  ausfuehren: (name: string) => void
}

export function useKartenskripte(
  setID: string,
  angelegt: () => void,
): Kartenskriptgriff {
  const texte = useTexte()
  const eigene = texte.seiten.karten

  const [bericht, setzeBericht] = useState<string>()
  const [befunde, setzeBefunde] = useState<readonly string[]>([])

  const vorgang = useVorgang()

  const holen = useCallback(async () => {
    const antwort = await api().Kartenskripte(setID)

    return auspacken(antwort, () => antwort.skripte)
  }, [setID])

  const angebot = useLaden(holen, [setID])

  const ausfuehren = (name: string) => {
    void vorgang.starte('kartenskript', async () => {
      try {
        const antwort = await api().KartenskriptAusfuehren(setID, name)
        const karten = auspacken(antwort, () => antwort.karten)

        setzeBefunde(antwort.befunde)
        setzeBericht(
          berichttext(karten.length, {
            eine: eigene.kartenskriptEine,
            mehrere: eigene.kartenskriptMehrere,
          }),
        )

        // **Erst nach dem geglückten Schreiben** (§7, Regel 3).
        angelegt()
      } catch (ursache: unknown) {
        setzeBericht(undefined)
        setzeBefunde([meldungAus(ursache, texte)])
      }
    })
  }

  return {
    skripte: angebot.daten ?? [],
    laeuft: vorgang.laeuft('kartenskript'),
    bericht,
    befunde,
    ausfuehren,
  }
}
