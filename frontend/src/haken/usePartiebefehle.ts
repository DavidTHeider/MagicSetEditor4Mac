// **Die Befehle einer Partie, der Reihe nach** — P5
// (#2373),
// Befunde C-21 und C-24. Herausgelöst aus `usePartie.ts`, als die
// Warteschlange es über die Zeilengrenze aus §2 hob; der Schnitt folgt der
// Frage: dort, welche Partie offen ist; hier, wie ein Befehl hinausgeht.
//
// # Kein Wunsch geht still verloren
//
// **Bis P5 verwarf ein zweiter Befehl**, solange der erste unterwegs war
// (`useVorgang`), und drei schnell getappte Länder hinterließen eines. Jetzt
// wartet er (`useSchlange`), und der Tisch zeigt, wie viele warten
// (`wartend`). **Ein wartender Wunsch ist kein Spielstand** (Vermerk 0126):
// gezeichnet wird weiter nur, was der Kern bestätigt hat.
import { useCallback, useRef, type Dispatch, type SetStateAction } from 'react'

import {
  api,
  auspacken,
  type BefehlAntwort,
  type Partie,
  type Spielereignis,
} from '../api'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { neueBefehlskennung, partievorgang } from './partiewege'
import type { Lauf } from './useLauf'
import { useSchlange } from './useSchlange'
import type { Vorgang } from './useVorgang'

// Befehlswunsch ist ein Befehl, der wartet — mit der Partie und dem Stand,
// aus dem er kam.
interface Befehlswunsch {
  offen: Partie
  stand: number
  art: string
  nutzlast: Record<string, unknown>
}

export function usePartiebefehle({
  partie,
  vorgang,
  lauf,
  nachladen,
  nurAktuell,
  setzeMeldung,
  setzeEreignisse,
}: {
  partie: Partie | undefined
  vorgang: Vorgang
  lauf: Lauf
  nachladen: (kennung: string, fuer: string) => Promise<void>
  nurAktuell: (meiner: number, arbeit: () => void) => void
  setzeMeldung: Dispatch<
    SetStateAction<{ satz: string; vorgang: string } | undefined>
  >
  setzeEreignisse: Dispatch<SetStateAction<Spielereignis[]>>
}): {
  befehlen: (art: string, nutzlast: Record<string, unknown>) => void
  wartend: number
} {
  const texte = useTexte()

  // **Der zuletzt bestätigte Stand je Partie** — P5 (#2373): ein Wunsch,
  // der wartete, rechnet gegen die Fassung, die der vorige hinterlassen hat,
  // und nicht gegen die, die beim Einreihen zu sehen war.
  const bestaetigt = useRef<{ id: string; fassung: number }>(undefined)

  // stellen schickt einen Wunsch hinaus — **der Reihe nach** (`useSchlange`).
  const stellen = useCallback(
    async (wunsch: Befehlswunsch, ersterDerFolge: boolean) => {
      const { offen, stand, art, nutzlast } = wunsch

      // **Was ein Wechsel überholt hat, geht nicht mehr hinaus** (Regel 1).
      if (!lauf.istAktuell(stand)) {
        return
      }

      const lief = await vorgang.starte(partievorgang(offen.id), async () => {
        // **Ein Befehl wechselt das Bezugsobjekt nicht** — er ändert, was
        // darin steht. Deshalb merkt er sich seinen Stand, statt
        // weiterzuzählen (Regel 2).
        const meiner = lauf.nummer()

        // **Eine neue Folge räumt die alte Meldung**, ein Wunsch mitten in
        // der Folge nicht — sonst wäre eine Ablehnung still verloren.
        if (ersterDerFolge) {
          setzeMeldung(undefined)
        }

        const zuletzt = bestaetigt.current
        const fassung = Math.max(
          Number(offen.fassung),
          zuletzt?.id === offen.id ? zuletzt.fassung : 0,
        )

        try {
          const antwort: BefehlAntwort = await api().PartieBefehl(
            offen.id,
            neueBefehlskennung(),
            offen.fuer,
            art,
            nutzlast,
            // **Gegen den Stand, den der Mensch zuletzt bestätigt bekam.**
            // Ist dazwischen etwas geschehen, das er nicht kennt, kommt ein
            // Konflikt zurück statt einer stillen Wirkung.
            String(fassung),
          )

          auspacken(antwort, () => undefined)
          bestaetigt.current = {
            id: offen.id,
            fassung: Number(antwort.fassung),
          }

          if (!lauf.istAktuell(meiner)) {
            return true
          }

          setzeEreignisse(antwort.ereignisse)
        } catch (ursache: unknown) {
          nurAktuell(meiner, () => {
            setzeMeldung({
              satz: meldungAus(ursache, texte),
              vorgang: partievorgang(offen.id),
            })
          })

          // **Nach einem gescheiterten Schreiben wird nicht nachgeladen**
          // (Regel 3) — ein Lesefehler überschriebe sonst die Meldung, die
          // sagt, was zu tun ist.
          return true
        }

        await nachladen(offen.id, offen.fuer)

        return true
      })

      // **Lief gerade ein anderer Vorgang am Tisch** — ein Halt oder ein
      // Weiter —, ging der Wunsch nicht hinaus. Das wird gesagt und nicht
      // verschwiegen.
      if (lief === undefined) {
        setzeMeldung({
          satz: texte.seiten.spielen.wunschNichtGestellt,
          vorgang: partievorgang(offen.id),
        })
      }
    },
    [
      vorgang,
      lauf,
      texte,
      nachladen,
      nurAktuell,
      setzeEreignisse,
      setzeMeldung,
    ],
  )

  const schlange = useSchlange(stellen)

  const befehlen = useCallback(
    (art: string, nutzlast: Record<string, unknown>) => {
      if (!partie) {
        return
      }

      schlange.stelle({ offen: partie, stand: lauf.nummer(), art, nutzlast })
    },
    [partie, schlange, lauf],
  )

  return { befehlen, wartend: schlange.wartend }
}
