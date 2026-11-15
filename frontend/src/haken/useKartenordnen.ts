// **Eine Karte nach oben oder unten schieben** — der Haken hinter den zwei
// Menüeinträgen.
//
// Bis Werkstück U2 von Phase 5.1
// (#1413) war
// das ein Bauteil mit zwei Knöpfen (`seiten/Kartenordnen.tsx`); der Tausch
// selbst ist derselbe wie seit #1168: `KartenTauschen` mit dem Nachbarn in
// der Sortierung, die die Liste zeigt.
//
// **Am Rand gibt es keinen Nachbarn** — dann ist der Eintrag da und ruht
// (`nachOben === undefined`): was fehlt, kann niemand vermissen.
import { useState } from 'react'

import { api, auspacken, type KartenzeileDTO } from '../api'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { type Vorgang } from './useVorgang'

interface Ordnungsgriff {
  nachOben: (() => void) | undefined
  nachUnten: (() => void) | undefined
  laeuft: boolean
  fehler: string | undefined
}

export function useKartenordnen(
  setID: string,
  karteID: string,
  karten: readonly KartenzeileDTO[],
  vorgang: Vorgang,
  aufGetauscht: () => void,
): Ordnungsgriff {
  const texte = useTexte()
  const [fehler, setzeFehler] = useState<string>()

  const [zuletzt, setzeZuletzt] = useState(karteID)
  if (zuletzt !== karteID) {
    setzeZuletzt(karteID)
    setzeFehler(undefined)
  }

  const stelle = karten.findIndex((karte) => karte.id === karteID)

  const tausch = (nachbar: KartenzeileDTO | undefined) =>
    nachbar === undefined
      ? undefined
      : () => {
          void vorgang.starte(`ordnen:${karteID}`, async () => {
            try {
              const antwort = await api().KartenTauschen(
                setID,
                karteID,
                nachbar.id,
              )
              auspacken(antwort, () => undefined)

              setzeFehler(undefined)
              aufGetauscht()
            } catch (ursache: unknown) {
              setzeFehler(meldungAus(ursache, texte))
            }
          })
        }

  return {
    nachOben: tausch(karten[stelle - 1]),
    nachUnten: tausch(karten[stelle + 1]),
    laeuft: vorgang.laeuft(`ordnen:${karteID}`),
    fehler,
  }
}
