// **Ob die Seitenleiste steht** — Werkstück U14 von Phase 5.1 (#1490).
//
// Eine Einstellung auf **Geräteebene** (`seitenleiste`, 0079): eine
// Fensterfrage, die nicht mit dem Profil reist. Gesetzt wird optimistisch und
// bei einem Fehler zurückgenommen (§7, Regel 3); ein Vorgang zur Zeit (Regel
// 4). Dasselbe Muster wie `useArbeitsplatzgedaechtnis`.
import { useCallback, useState } from 'react'

import { useFensterteilmeldung } from '../darstellungsstand'
import { api } from '../fassade'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import type { Einstellungsstand } from './useEinstellungen'
import { useVorgang } from './useVorgang'

export const seitenleistenschluessel = 'seitenleiste'

// leistenlageschluessel ist, ob die Leiste oben steht — UX-14.AK-02 (#1686).
export const leistenlageschluessel = 'leistenlage'

const sichtbar = 'sichtbar'
const verborgen = 'verborgen'

export interface Seitenleistenstand {
  // offen sagt, ob die Leiste steht — **offen, solange nichts geladen ist**:
  // eine Leiste, die beim Start kurz fehlt und dann erscheint, springt.
  offen: boolean
  // bereit sagt, ob die Einstellung da ist und ein Wunsch geschrieben wird.
  bereit: boolean
  laeuft: boolean
  fehler: string | undefined
  umschalten: () => void

  // waagerecht sagt, ob die Leiste oben steht statt links — UX-14.AK-02
  // (#1686). **Senkrecht, solange nichts geladen ist**, aus demselben Grund
  // wie `offen`: eine Leiste, die beim Start umspringt, springt.
  waagerecht: boolean
}

// **Der Stand kommt herein**, statt hier geladen zu werden: `App` hält die
// eine Instanz von `useEinstellungen`, und eine zweite lüde alles doppelt —
// gefunden an `einstellungslaeufe.test.tsx`, der vier Läufe statt zwei sah.
export function useSeitenleiste(stand: Einstellungsstand): Seitenleistenstand {
  const texte = useTexte()
  const vorgang = useVorgang()

  const [gewuenscht, setzeGewuenscht] = useState<boolean>()
  const [fehler, setzeFehler] = useState<string>()

  const bereit =
    stand.geladen &&
    stand.eintraege.some(
      (eintrag) => eintrag.schluessel === seitenleistenschluessel,
    )

  const offen = gewuenscht ?? stand.wert(seitenleistenschluessel) !== verborgen
  const waagerecht = stand.wert(leistenlageschluessel) === 'waagerecht'

  // **Die Menüleiste erfährt, wie die Leiste dasteht** — UXR-F14 (#2029):
  // *Darstellung* nennt den Eintrag nach dem Stand, und **oben gibt es nichts
  // einzuklappen**, der Eintrag ist dort gegraut (derselbe Fall wie in
  // `umschalten`). Gemeldet wird hier und nicht im `Rahmen`: dieser Haken
  // hält die Wahrheit, die der Menüeintrag auch schaltet.
  useFensterteilmeldung('seitenleiste', !waagerecht, offen)

  const umschalten = useCallback(() => {
    // **Oben gibt es nichts einzuklappen** — Befund G6 des Zweitblicks über
    // Phase 6.3. ⌥⌘S und der Menüeintrag enden beide hier; eine Sperre nur
    // am Knopf ließe sie einen Zustand umlegen, den niemand sieht.
    if (waagerecht || vorgang.laeuft(seitenleistenschluessel)) {
      return
    }

    const soll = !offen
    setzeGewuenscht(soll)

    if (!bereit) {
      return
    }

    void vorgang.starte(seitenleistenschluessel, async () => {
      try {
        const antwort = await api().EinstellungSetzen(
          seitenleistenschluessel,
          soll ? sichtbar : verborgen,
        )
        auspacken(antwort, () => undefined)
        setzeFehler(undefined)
      } catch (ursache: unknown) {
        // **Zurück auf den Stand des Bestands** (§7, Regel 3).
        setzeGewuenscht(undefined)
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }, [bereit, offen, texte, vorgang, waagerecht])

  return {
    offen,
    bereit,
    laeuft: vorgang.laeuft(seitenleistenschluessel),
    fehler,
    umschalten,
    waagerecht,
  }
}
