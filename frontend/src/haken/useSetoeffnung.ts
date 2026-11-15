// **Ein Set öffnen — aus welcher Quelle auch immer**: eine Datei, ein
// Beispielset, MTG JSON, ein Satz von Scryfall.
//
// **Herausgelöst aus `Sets.tsx`** (FMT-03, #1692), als die Seite an der
// Zeilengrenze stand (§2): was sie tut, steht hier, wie sie aussieht, dort.
import { useState } from 'react'
import { auspacken, type OeffnenAntwort } from '../api'
import { istAbgebrochen, meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from './useVorgang'

export function useSetoeffnung(nachladen: () => void) {
  const texte = useTexte()
  const vorgang = useVorgang()

  const [ausgang, setzeAusgang] = useState<OeffnenAntwort>()
  const [oeffnungsfehler, setzeOeffnungsfehler] = useState<string>()

  // teilerfolg zählt die Sets, die ein **abgebrochener** Import zurückgelassen
  // hat (#155); `undefined` heißt „kein Abbruch".
  const [teilerfolg, setzeTeilerfolg] = useState<number>()

  // **Ein Vorgang** (§7, Regel 4): der Datensatz ist hier das Öffnen selbst —
  // zwei Dateidialoge übereinander sind unter macOS ohnehin nichts, was jemand
  // will. Deshalb teilen sich alle Wege **einen** Vorgangsnamen.
  const oeffnen = async (waehlen: () => Promise<OeffnenAntwort>) => {
    await vorgang.starte('oeffnen', async () => {
      setzeOeffnungsfehler(undefined)
      setzeAusgang(undefined)
      setzeTeilerfolg(undefined)

      try {
        const antwort = await waehlen()

        // **Vor dem Auspacken** (#155), denn genau im Fehlerfall zählt es:
        // ein Abbruch mitten in einer AllSets-Datei lässt Sets zurück. Sie
        // stehen im Bestand — jedes lag in einer eigenen Transaktion —, und wer
        // das nicht erfährt, findet sie später in der Liste und hält sie für
        // vollständig. `auspacken` wirft; was danach steht, liefe nie.
        if (antwort.unvollstaendig && antwort.angelegt.length > 0) {
          setzeTeilerfolg(antwort.angelegt.length)
        }

        auspacken(antwort, () => undefined)

        if (antwort.abgebrochen) {
          return
        }

        setzeAusgang(antwort)
      } catch (ursache: unknown) {
        // **Wer anhält, bekommt keine Fehlermeldung dafür** (FMT-03) — was
        // geblieben ist, sagt der Laufstand.
        if (!istAbgebrochen(ursache)) {
          setzeOeffnungsfehler(meldungAus(ursache, texte))
        }
      } finally {
        // **Geschrieben ist geschrieben** (§7, Regel 3) — und deshalb im
        // `finally`: bei einem halb geglückten Import liegen Sets im Bestand,
        // obwohl der Aufruf mit einem Fehler endete.
        nachladen()
      }
    })
  }

  return {
    oeffnen,
    laeuft: vorgang.laeuft('oeffnen'),
    ausgang,
    oeffnungsfehler,
    teilerfolg,
  }
}
