// Das Zurückspielen einer Sicherung — der Gegenweg zu `useBestandSichern`
// (#940).
//
// # Warum zwei Schritte und nicht einer
//
// Sichern schreibt eine neue Datei; zurückspielen **ersetzt** den ganzen
// Bestand. Dazwischen gehört eine Entscheidung, und eine Entscheidung braucht
// etwas zu entscheiden: was in der Datei steht, wann sie entstanden ist, wie
// viel darin liegt. Deshalb `waehlen` und danach `zurueckspielen`.
//
// # Die Regeln aus §7, die er hält
//
//  1. **Ein Vorgang je Datensatz** (Regel 4): der Datensatz ist der Bestand,
//     und es gibt genau einen. Zwei Läufe zugleich gingen gegen eine Datei,
//     die halb ausgetauscht ist.
//  2. **Geschrieben ist geschrieben** (Regel 3): nach einem gescheiterten
//     Zurückspielen bleibt die Meldung stehen, und die Auswahl bleibt es auch
//     — wer es noch einmal versuchen will, muss die Datei nicht neu suchen.
//  3. **Regel 1 und 2 greifen nicht**: es gibt kein Bezugsobjekt, das wechseln
//     könnte. Der Bestand ist der Bestand.
import { useCallback, useState } from 'react'
import { epocheGesehen } from '../bestandsepoche'

import type { RuecksprunglageAntwort } from '../apirueckkehr'
import type { Sicherungsauswahl } from '../api'
import { api, auspacken } from '../api'
import { Bindungsfehler } from '../bindungsfehler'
import { meldungAus } from '../meldung'
import { useLaden } from './useLaden'
import { useTexte } from '../sprache/kontext'
import type { Vorgang } from './useVorgang'

export interface Rueckspiellauf {
  // waehlen öffnet den Dateidialog und prüft, was gewählt wurde.
  waehlen: () => void

  // gesperrt ist der Pfad einer gewählten Sicherung, die ihr Kennwort
  // braucht, bevor sie etwas zeigt (OPS-06, #1723) — undefined sonst.
  gesperrt: string | undefined

  // entsperren liest die gesperrte Sicherung mit diesem Kennwort. **Es bleibt
  // für den Vergleich und das Zurückspielen**, bis die Auswahl geräumt ist —
  // wer es zweimal tippen müsste, tippte es beim zweiten Mal falsch.
  entsperren: (pfad: string, kennwort: string) => void

  // zurueckspielen übernimmt die Sicherung an diesem Pfad.
  //
  // **Der Pfad kommt vom Aufrufer und nicht aus dem Zustand hier.** Er steht
  // dort ohnehin — die Rückfrage zeigt ihn —, und ein Griff in den eigenen
  // Zustand brauchte eine Wache gegen „noch nichts gewählt", die kein Aufruf
  // betreten kann (§3).
  zurueckspielen: (pfad: string) => void

  laeuft: boolean

  // auswahl ist die geprüfte Sicherung — undefined, solange keine dasteht.
  auswahl: Sicherungsauswahl | undefined

  // lage ist der Vergleich zur gewählten Sicherung (O7, #1431) — undefined,
  // solange keine gewählt ist oder er noch unterwegs ist.
  //
  // **Neben der Auswahl und nicht darin**: jene sagt, was **in** der Datei
  // liegt, diese, was das Zurückspielen **kostet**. Zwei Fragen, und die
  // zweite kommt von einer anderen Bindung.
  lage: RuecksprunglageAntwort | undefined

  // beiseite ist der Ort des ersetzten Bestands nach einem geglückten Lauf.
  beiseite: string | undefined

  fehler: string | undefined

  raeumen: () => void
}

// **Derselbe Vorgang wie beim Sichern** — der Grund steht dort.
//
// `danach` meldet dem Aufrufer, dass der Bestand ein anderer ist.
export function useSicherungZurueckspielen(
  vorgang: Vorgang,
  danach: () => void,
): Rueckspiellauf {
  const texte = useTexte()

  const [auswahl, setzeAuswahl] = useState<Sicherungsauswahl>()
  const [gesperrt, setzeGesperrt] = useState<string>()
  const [kennwort, setzeKennwort] = useState('')
  const [beiseite, setzeBeiseite] = useState<string>()
  const [fehler, setzeFehler] = useState<string>()

  // **Der Vergleich hängt an der gewählten Datei** — `useLaden` verwirft die
  // Antwort eines überholten Laufs (§7, Regel 2), und beim Wechsel der
  // Auswahl beginnt er von vorn.
  const vergleich = useLaden(async () => {
    if (auswahl === undefined) {
      return undefined
    }

    const antwort = await api().Ruecksprunglage(auswahl.pfad, kennwort)

    return auspacken(antwort, () => antwort)
  }, [auswahl?.pfad, kennwort])

  const raeumen = useCallback(() => {
    setzeAuswahl(undefined)
    setzeGesperrt(undefined)
    setzeKennwort('')
    setzeBeiseite(undefined)
    setzeFehler(undefined)
  }, [])

  const waehlen = useCallback(() => {
    void vorgang.starte('bestand', async () => {
      raeumen()

      try {
        const antwort = await api().SicherungWaehlen()

        // **Eine verschlüsselte Sicherung fragt nach ihrem Kennwort**, statt
        // als Fehler dazustehen: gewählt ist sie ja.
        if (antwort.fehler?.schluessel === 'sicherungskennwortFehlt') {
          setzeGesperrt(antwort.pfad)

          return
        }

        auspacken(antwort, () => undefined)

        // **Ein Abbruch hinterlässt nichts.** Wer den Dialog wegklickt, will
        // keine Rückfrage über eine Datei, die er nicht gewählt hat.
        if (!antwort.abgebrochen) {
          setzeAuswahl(antwort)
        }
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }, [raeumen, texte, vorgang])

  const entsperren = useCallback(
    (pfad: string, eingabe: string) => {
      void vorgang.starte('bestand', async () => {
        setzeFehler(undefined)

        try {
          // **Der Pfad kommt vom Aufrufer** — dieselbe Überlegung wie bei
          // `zurueckspielen`: die Seite zeigt das Feld nur, solange er da ist.
          const antwort = await api().SicherungBeschauen(pfad, eingabe)

          auspacken(antwort, () => undefined)

          setzeKennwort(eingabe)
          setzeGesperrt(undefined)
          setzeAuswahl(antwort)
        } catch (ursache: unknown) {
          setzeFehler(meldungAus(ursache, texte))
        }
      })
    },
    [texte, vorgang],
  )

  const zurueckspielen = useCallback(
    (pfad: string) => {
      void vorgang.starte('bestand', async () => {
        setzeFehler(undefined)

        try {
          const antwort = await api().SicherungZurueckspielen(pfad, kennwort)

          // **Die Epoche zuerst, auch vor einem Fehler** (J05, #2269): ein
          // Rückweg, der scheiterte, kann einen anderen Bestand hinterlassen.
          epocheGesehen(antwort.bestandsepoche)

          // **Ein Fehler sagt, was danach dasteht** — die ausdrückliche
          // Rückkehrmeldung aus A05 (#2219): der bisherige Bestand ist
          // unverändert, oder auch der Weg zurück scheiterte.
          if (antwort.fehler) {
            const worte = texte.seiten.bestandspflege
            const danach = antwort.unveraendert
              ? worte.rueckkehrUnveraendert
              : worte.rueckkehrUnvollstaendig

            setzeFehler(
              `${meldungAus(new Bindungsfehler(antwort.fehler), texte)} ${danach}`,
            )

            return
          }

          // **Die Auswahl verschwindet erst jetzt.** Bis hierher konnte der
          // Lauf scheitern, und dann soll sie stehen bleiben (Regel 3).
          setzeAuswahl(undefined)
          setzeBeiseite(antwort.beiseite)

          // **Erst jetzt, und nur hier.** Ein Aufruf vor der Antwort räumte
          // den Kontext für einen Tausch, der noch scheitern kann.
          danach()
        } catch (ursache: unknown) {
          setzeFehler(meldungAus(ursache, texte))
        }
      })
    },
    [danach, kennwort, texte, vorgang],
  )

  return {
    waehlen,
    gesperrt,
    entsperren,
    zurueckspielen,
    laeuft: vorgang.laeuft('bestand'),
    auswahl,
    lage: vergleich.daten,
    beiseite,
    fehler,
    raeumen,
  }
}
