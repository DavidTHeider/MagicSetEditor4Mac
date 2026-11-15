// **Die Autoersetzungsregeln eines Spiels** —
// #1341,
// Vermerk [0060](../../docs/entscheidungen/0060-eigene-autoersetzungen-liegen-im-bestand.md).
//
// **Ein eigener Haken**, weil er fünf Wege hält: laden und vier Arten zu
// schreiben. In `Karten.tsx` wären das fünf weitere Zustände in einer Datei,
// die schon an der Zeilengrenze steht (§2).
import { useCallback, useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useLaden } from './useLaden'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from './useVorgang'
import type { AutoersetzungsregelDTO } from '../antwortenautoersetzung'

export interface Autoersetzungsgriff {
  regeln: readonly AutoersetzungsregelDTO[]
  an: boolean
  laeuft: boolean
  fehler: string | undefined

  // setzen nimmt die Regel **ohne** `eigen`: ob sie dem Paket gehört,
  // entscheidet der Dienst — eine Fläche, die es mitschickte, könnte es
  // falsch behaupten.
  setzen: (regel: Omit<AutoersetzungsregelDTO, 'eigen'>) => void
  loeschen: (muster: string) => void
  zuruecksetzen: () => void
  schalten: (an: boolean) => void
}

export function useAutoersetzungen(setID: string): Autoersetzungsgriff {
  const texte = useTexte()

  const [fehler, setzeFehler] = useState<string>()

  const vorgang = useVorgang()

  const holen = useCallback(async () => {
    const antwort = await api().Autoersetzungen(setID)

    return auspacken(antwort, () => ({
      regeln: antwort.regeln,
      an: antwort.an,
    }))
  }, [setID])

  const liste = useLaden(holen, [setID])

  // handeln fährt einen Schreibvorgang und lädt danach nach.
  //
  // **Ein Vorgang je Datensatz** (§7, Regel 4) — der Schlüssel ist das Set:
  // zwei gleichzeitige Änderungen an derselben Liste gingen gegen einen
  // Stand, den das Backend noch nicht kennt.
  //
  // **Nachgeladen wird nur nach einem geglückten Schreiben** (Regel 3): sonst
  // überschriebe ein Lesefehler die Meldung, die sagt, was zu tun ist.
  const handeln = (tun: () => Promise<void>) => {
    void vorgang.starte(setID, async () => {
      try {
        await tun()
        setzeFehler(undefined)
        liste.nachladen()
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  return {
    regeln: liste.daten?.regeln ?? [],

    // **Ohne Antwort gilt „an"** — dieselbe Vorgabe wie im Bestand, und die
    // Seite soll nicht behaupten, die Autoersetzung sei aus, solange sie
    // nichts weiß.
    an: liste.daten?.an ?? true,

    laeuft: vorgang.laeuft(setID),
    fehler: fehler ?? liste.fehler,

    setzen: (regel) => {
      handeln(async () => {
        auspacken(
          await api().AutoersetzungSetzen(
            setID,
            regel.muster,
            regel.ersatz,
            regel.ganzesWort,
            regel.aktiv,
          ),
          () => undefined,
        )
      })
    },

    loeschen: (muster) => {
      handeln(async () => {
        auspacken(
          await api().AutoersetzungLoeschen(setID, muster),
          () => undefined,
        )
      })
    },

    zuruecksetzen: () => {
      handeln(async () => {
        auspacken(
          await api().AutoersetzungenZuruecksetzen(setID),
          () => undefined,
        )
      })
    },

    schalten: (an) => {
      handeln(async () => {
        auspacken(await api().AutoersetzungSchalten(setID, an), () => undefined)
      })
    },
  }
}
