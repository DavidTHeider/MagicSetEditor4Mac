// **Die Verwaltung einer Partie als Ganzem** — PLAY-03.AK-05
// (#1711).
//
// # Was hier steht
//
// *Was die Seite tut*, steht hier; *wie sie aussieht*, in
// `Partienverwaltung.tsx` (§2).
//
// # Die vier Regeln aus §7
//
// **Regel 4 trägt `useVorgang`**: ein Vorgang je Partie. Wer zweimal auf
// „Löschen" klickt, löscht nicht zweimal — der zweite Ruf ginge gegen einen
// Stand, den der Bestand noch nicht kennt.
//
// **Regel 3 steht in der Reihenfolge**: nachgeladen wird **nach** einer
// geglückten Änderung, und ein Fehler dabei macht sie nicht rückgängig.
// Nach einer **gescheiterten** wird gar nicht nachgeladen — sonst
// überschriebe ein Lesefehler die Meldung, die sagt, was zu tun ist.
//
// **Regel 1 und 2** treffen hier nicht: diese Seite hat kein Bezugsobjekt,
// das wechseln könnte — sie steht nur da, solange **kein** Tisch offen ist.
// Wer einsteigt, sieht diese Liste nicht mehr.
import { useState } from 'react'

import { api, auspacken } from '../api'
import { einsetzen } from '../sprache/einsetzen'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from './useVorgang'

import type { PartieverwaltungAntwort } from '../apipartieverwaltung'

// Einlesekennung ist der Schlüssel des Weges, der zu keiner Zeile gehört.
export const Einlesekennung = 'partie-einlesen'

// Verwaltungsmeldung gehört zu **einer** Zeile — sonst stünde die Antwort
// auf das Löschen von A an der Zeile von B.
export interface Verwaltungsmeldung {
  kennung: string
  satz: string
  istFehler: boolean
}

export interface Verwaltungsstand {
  laeuft: (kennung: string) => boolean
  meldung: Verwaltungsmeldung | undefined

  loeschen: (kennung: string) => void
  kopieren: (kennung: string) => void
  ausfuehren: (kennung: string) => void
  einlesen: () => void
}

export function usePartieverwaltung(nachladen: () => void): Verwaltungsstand {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  const lauf = useVorgang()
  const [meldung, setzeMeldung] = useState<Verwaltungsmeldung>()

  // fahre führt einen der vier Wege und meldet, was daraus wurde.
  const fahre = (
    kennung: string,
    arbeit: () => Promise<PartieverwaltungAntwort>,
    satzAus: (antwort: PartieverwaltungAntwort) => string,
  ): void => {
    void lauf.starte(kennung, async () => {
      setzeMeldung(undefined)

      let antwort: PartieverwaltungAntwort

      try {
        const roh = await arbeit()

        antwort = auspacken(roh, () => roh)
      } catch (ursache) {
        setzeMeldung({
          kennung,
          satz: meldungAus(ursache, texte),
          istFehler: true,
        })

        return
      }

      // **Ein weggeklickter Dialog ist kein Ergebnis** (§6): wer abbricht,
      // hat sich entschieden, und eine Meldung darüber wäre Lärm.
      if (antwort.abgebrochen) {
        return
      }

      setzeMeldung({
        kennung,
        satz: [satzAus(antwort), ...antwort.befunde].join(' '),
        istFehler: false,
      })

      nachladen()
    })
  }

  return {
    laeuft: lauf.laeuft,
    meldung,

    loeschen: (kennung) => {
      fahre(
        kennung,
        () => api().PartieLoeschen(kennung),
        () => eigene.partienGeloescht,
      )
    },

    kopieren: (kennung) => {
      fahre(
        kennung,
        () => api().PartieKopieren(kennung),
        () => eigene.partienDupliziert,
      )
    },

    ausfuehren: (kennung) => {
      fahre(
        kennung,
        () => api().PartieAusfuhrSpeichern(kennung),
        (antwort) => einsetzen(eigene.partienAbgelegt, { pfad: antwort.pfad }),
      )
    },

    einlesen: () => {
      fahre(
        Einlesekennung,
        () => api().PartieEinfuhrOeffnen(),
        () => eigene.partienEingelesen,
      )
    },
  }
}
