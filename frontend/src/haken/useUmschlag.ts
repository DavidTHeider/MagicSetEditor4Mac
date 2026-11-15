// Der portable Umschlag im Fenster — Werkstück N4 von Phase 5.4
// (#1457).
//
// # Die Regeln aus §7, die er hält
//
//  1. **Ein Vorgang je Datensatz** (Regel 4): der Datensatz ist beim Schnüren
//     das Set und beim Annehmen die Sendung. Zweimal denselben Umschlag
//     hereinzunehmen legte zwei Sets an, und das zweite wäre eine Kopie, die
//     niemand wollte.
//  2. **Geschrieben ist geschrieben** (Regel 3): nach einem Fehler bleibt die
//     Meldung stehen, und der angesehene Umschlag bleibt liegen — wer den
//     Fehler gelesen hat, kann es erneut versuchen, ohne noch einmal zu
//     wählen.
//  3. **Ein Wechsel räumt den offenen Kontext** (Regel 1): wer einen zweiten
//     Umschlag ansieht, sieht den ersten nicht mehr. `raeumen` ist der Weg
//     dorthin, und das Wählen räumt von selbst.
//  4. **Eine Antwort gehört zu dem Stand, aus dem sie stammt** (Regel 2) —
//     nachgetragen mit Befund 4 des Zweitblicks über Phase 5.4. Wählen und
//     Übernehmen hatten **zwei** Vorgangsschlüssel und hielten einander
//     deshalb nicht auf: wer während einer Übernahme einen zweiten Umschlag
//     wählte, sah ihn kurz — und dann löschte die verspätete Antwort der
//     Übernahme seine Vorschau mit `setzeAngesehen(undefined)`. Jetzt trägt
//     jeder Ruf seinen Stand, und eine Antwort aus einem überholten trägt
//     sich nicht mehr ein.
//
// # Ein Zähler für beide Richtungen, und warum
//
// **Weil es eine sichtbare Fläche ist und zwei Richtungen darauf schreiben.**
// `schnueren` hat einen Vorgangsschlüssel **je Set** — zwei Sets dürfen
// nebeneinander hinausgehen —, schreibt aber in denselben `fehler`, dasselbe
// `geschnuert` und dasselbe `fuer`. Der Nachlauf zum Zweitblick hat es
// angestrichen (Befund 2): die späte Antwort von A stand über B.
//
// **Die Regel lautet deshalb: der jüngste Ruf besitzt die Fläche.** Wählen
// und Schnüren zählen weiter, Übernehmen merkt sich nur; `raeumen` zählt
// ebenfalls weiter, denn wer aufräumt, will nichts zurückbekommen.
//
// **Und `fuer` sagt, wem die Meldung gehört** (UXR-04, #1912): was das
// Schnüren meldet, steht an seiner Setkarte; was Wählen und Übernehmen
// melden, an der Annahme. Beide räumen deshalb `fuer` — sonst stünde ihr
// Fehler an der Karte des zuletzt abgelegten Sets.
//
// **Was das kostet, gehört ausgesprochen**: wer während eines Schnürens
// einen Umschlag wählt, sieht die Bestätigung des Schnürens nicht mehr. Die
// **Datei ist trotzdem geschrieben** — verloren geht die Zeile, nicht die
// Arbeit. Die Gegenrechnung wäre eine Bestätigung für Set A über der
// Vorschau von B, und die ist schlechter.
import { useCallback, useState } from 'react'

import { api, auspacken, type UmschlagAntwort } from '../api'
import { meldungAus } from '../meldung'
import { useLauf } from './useLauf'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from './useVorgang'

export interface Umschlaglauf {
  // schnueren legt ein Set ab — der Ort kommt aus dem nativen Dialog.
  schnueren: (setID: string) => void

  // waehlen fragt nach einem Umschlag und sagt, was drin steht.
  waehlen: () => void

  // uebernehmen schreibt den angesehenen Umschlag in den Bestand.
  uebernehmen: (pfad: string) => void

  raeumen: () => void

  laeuft: (was: string) => boolean

  // geschnuert ist, was zuletzt hinausging — samt der Kennung seines Sets.
  geschnuert: UmschlagAntwort | undefined
  fuer: string | undefined

  // angesehen ist, was hereinkommen will — **noch nicht** hereingekommen.
  angesehen: UmschlagAntwort | undefined

  // uebernommen ist die Kennung des Sets, das daraus entstand.
  uebernommen: string | undefined

  fehler: string | undefined
}

export function useUmschlag(): Umschlaglauf {
  const vorgang = useVorgang()
  const lauf = useLauf()
  const texte = useTexte()

  const [geschnuert, setzeGeschnuert] = useState<UmschlagAntwort>()
  const [fuer, setzeFuer] = useState<string>()
  const [angesehen, setzeAngesehen] = useState<UmschlagAntwort>()
  const [uebernommen, setzeUebernommen] = useState<string>()
  const [fehler, setzeFehler] = useState<string>()

  const raeumen = useCallback(() => {
    // **Aufräumen zählt weiter** — sonst trüge eine Antwort, die noch
    // unterwegs ist, gleich wieder ein, was gerade weggeräumt wurde.
    lauf.naechster()

    setzeGeschnuert(undefined)
    setzeFuer(undefined)
    setzeAngesehen(undefined)
    setzeUebernommen(undefined)
    setzeFehler(undefined)
  }, [lauf])

  // **Die Art kommt dazu** (T7, #1472) — `set` bleibt die Vorgabe, damit
  // jeder vorhandene Aufrufer weiter ein Set schnürt.
  const schnueren = useCallback(
    (setID: string, art = 'set') => {
      void vorgang.starte(`schnueren-${setID}`, async () => {
        // **Auch das Schnüren zählt weiter** — es wechselt, worüber die
        // Fläche spricht, und der Schlüssel je Set lässt zwei davon
        // nebeneinander laufen. Der Grund steht im Kopf dieser Datei.
        const meiner = lauf.naechster()

        setzeFehler(undefined)
        setzeGeschnuert(undefined)
        setzeFuer(setID)

        try {
          const antwort = await api().UmschlagSchnueren(art, setID)

          auspacken(antwort, () => undefined)

          // **Ein abgebrochener Dialog ist keine Meldung wert** — derselbe
          // Satz wie in `useSetSpeichern`.
          if (!lauf.istAktuell(meiner) || antwort.abgebrochen) {
            return
          }

          setzeGeschnuert(antwort)
        } catch (ursache: unknown) {
          if (lauf.istAktuell(meiner)) {
            setzeFehler(meldungAus(ursache, texte))
          }
        }
      })
    },
    [vorgang, lauf, texte],
  )

  const waehlen = useCallback(() => {
    void vorgang.starte('umschlag-waehlen', async () => {
      // **Das Wählen zählt weiter** — es wechselt das Bezugsobjekt, und
      // jede Antwort, die noch zum vorigen gehört, ist damit überholt.
      //
      // **Und es prüft seinen eigenen Stand nicht.** Es ist das Einzige, was
      // weiterzählt, und [useVorgang] lässt keinen zweiten Wählvorgang
      // daneben laufen — ein `istAktuell` hier wäre ein Zweig, den kein Lauf
      // betreten kann (§3). Gebraucht wird er in `uebernehmen`, und dort
      // steht er.
      lauf.naechster()

      setzeFehler(undefined)
      setzeAngesehen(undefined)
      setzeUebernommen(undefined)
      setzeGeschnuert(undefined)
      setzeFuer(undefined)

      try {
        const antwort = await api().UmschlagWaehlen()

        auspacken(antwort, () => undefined)

        if (antwort.abgebrochen) {
          return
        }

        setzeAngesehen(antwort)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }, [vorgang, lauf, texte])

  const uebernehmen = useCallback(
    (pfad: string) => {
      void vorgang.starte('umschlag-uebernehmen', async () => {
        // **Die Übernahme merkt sich den Stand, zählt ihn aber nicht
        // weiter** — sie wechselt nichts, sie schreibt. Genau das
        // unterscheidet [Lauf.nummer] von [Lauf.naechster].
        const meiner = lauf.nummer()

        setzeFehler(undefined)
        setzeUebernommen(undefined)
        setzeFuer(undefined)

        try {
          const antwort = await api().UmschlagUebernehmen(pfad)

          auspacken(antwort, () => undefined)

          if (!lauf.istAktuell(meiner)) {
            return
          }

          setzeUebernommen(antwort.setID)

          // **Erst nach der Übernahme geht die Vorschau weg.** Vorher wäre
          // sie fort, während der Lauf noch liefe — und nach einem Fehler
          // stünde nichts mehr da, womit es ein zweites Mal ginge.
          setzeAngesehen(undefined)
        } catch (ursache: unknown) {
          if (lauf.istAktuell(meiner)) {
            setzeFehler(meldungAus(ursache, texte))
          }
        }
      })
    },
    [vorgang, lauf, texte],
  )

  const laeuft = useCallback((was: string) => vorgang.laeuft(was), [vorgang])

  return {
    schnueren,
    waehlen,
    uebernehmen,
    raeumen,
    laeuft,
    geschnuert,
    fuer,
    angesehen,
    uebernommen,
    fehler,
  }
}
