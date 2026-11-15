// **Ein Fenster der Sammlung nachholen** — aus `useSammlung.ts`
// herausgelöst, als U09
// (#2157)
// die Datei über die Zeilengrenze aus §2 brachte. Dort steht, **welche** Sicht
// gilt; hier, wie ein Ausschnitt dazukommt, ohne doppelt zu fragen und ohne
// in eine fremde Sicht zu fallen.
import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type Dispatch,
  type SetStateAction,
} from 'react'

import { api } from '../fassade'
import { meldungAus } from '../meldung'
import type { Sicht } from '../seiten/sammlungssicht'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { erstesFenster, type Nachschlag } from './sammlungszeilen'

interface Bereich {
  anfang: number
  ende: number
}

interface Register {
  schluessel: string
  bereiche: Bereich[]
}

// ersteLuecke gibt die erste Stelle in `[anfang, ende)`, die kein Bereich
// deckt — oder `undefined`, wenn alles gedeckt ist.
export function ersteLuecke(
  bereiche: readonly Bereich[],
  anfang: number,
  ende: number,
): number | undefined {
  let stelle = anfang

  for (;;) {
    const deckend = bereiche.find(
      (bereich) => bereich.anfang <= stelle && stelle < bereich.ende,
    )
    if (deckend === undefined) {
      return stelle < ende ? stelle : undefined
    }

    stelle = deckend.ende
  }
}

interface Nachschlaggriffe {
  sicht: Sicht
  holtAlles: boolean
  schluessel: string
  nachgeholt: Nachschlag
  setzeNachgeholt: Dispatch<SetStateAction<Nachschlag>>
  setzeFehler: (fehler: string | undefined) => void
}

export function useFensternachschlag({
  sicht,
  holtAlles,
  schluessel,
  nachgeholt,
  setzeNachgeholt,
  setzeFehler,
}: Nachschlaggriffe): (anfang: number, ende: number) => void {
  const texte = useTexte()

  // gemerkterSchluessel ist die Sicht, die **jetzt** gilt.
  //
  // **Als Referenz und nicht als Zustand** — Befund 6 des Nachlaufs zum
  // Zweitblick über Phase 5.2: ein Ruf, der Sekunden später zurückkommt, trägt
  // den Schlüssel von damals in seiner Hülle. Verglichen wird gegen den von
  // jetzt, und eine Referenz ist das Einzige, was beim Zurückkommen schon
  // weitergewandert ist.
  //
  // **Gesetzt im Effekt und nicht beim Zeichnen**: eine Referenz während des
  // Zeichnens zu ändern ist ein Fund von `react-hooks/refs`, und zu Recht —
  // React darf ein Zeichnen verwerfen. Der Effekt läuft nach dem Festschreiben,
  // und dazwischen liegt keine Antwort: jeder Ruf hier geht über die Naht und
  // braucht mindestens eine Makroaufgabe.
  const gemerkterSchluessel = useRef(schluessel)

  useEffect(() => {
    gemerkterSchluessel.current = schluessel
  }, [schluessel])

  // gedeckt sind die Bereiche, die geholt sind oder unterwegs — **je Sicht
  // ein Register**. Befund 2 des Zweitblicks über Phase G: gemerkt war der
  // Anfang, und `[400,440)` galt als geholt, sobald `[400,420)` lief; und eine
  // Sicht, die nach A → B → A wiederkam, fand den Ruf des ersten A noch vor.
  // Ein neues Register beim ersten Ruf einer anderen Sicht ist die
  // Wechselgeneration — erkannt an der Identität und nicht am Text.
  //
  // **Als Zustand, der beim Zeichnen nachzieht** — dieselbe Gestalt wie
  // `nachgeholt` in `useSammlung`: ein Register, das erst beim nächsten Ruf
  // wechselte, sähe A → B → A nicht, wenn in B niemand fragt. Die Referenz
  // daneben sagt der Antwort, ob ihr Register noch gilt.
  const [register, setzeRegister] = useState<Register>(() => ({
    schluessel,
    bereiche: [],
  }))

  if (register.schluessel !== schluessel) {
    setzeRegister({ schluessel, bereiche: [] })
  }

  const geltendes = useRef(register)

  useEffect(() => {
    geltendes.current = register
  }, [register])

  return useCallback(
    (anfang: number, ende: number) => {
      if (
        holtAlles ||
        ende <= erstesFenster ||
        nachgeholt.schluessel !== schluessel
      ) {
        return
      }

      // **Auch was unterwegs ist, gilt als geholt** (U09, #2157): gemerkt
      // war nur, was angekommen ist, und wer scrollte, während ein Fenster
      // lief, stieß denselben Ruf ein zweites Mal an. Nach einem Fehler ist
      // es wieder frei — der Platzhalter soll sich holen lassen. **Geholt
      // wird ab der ersten Lücke**, nicht der ganze Bereich noch einmal.
      const ab = ersteLuecke(register.bereiche, anfang, ende)
      if (ab === undefined) {
        return
      }

      const bereich = { anfang: ab, ende }
      register.bereiche.push(bereich)

      // gilt sagt, ob die Antwort noch zur Sicht **und** ihrer Generation
      // gehört — Befund 2 des Nachlaufs über Phase G: nach A → B → A heißt
      // die Sicht wie vorher, und ein Vergleich über den Namen allein ließ
      // einen alten Fehler stehen oder einen neuen verschwinden.
      const gilt = () =>
        schluessel === gemerkterSchluessel.current &&
        register === geltendes.current

      void (async () => {
        // **Auch ein nachgeholtes Fenster kann scheitern.** Ohne `auspacken`
        // würde ein Fehler aus Go zu `undefined`, landete als Fenster im
        // Speicher und ließe die Zeilen leer stehen, ohne dass jemand etwas
        // erführe (§8).
        try {
          const antwort = await api().Sammlungsfenster(
            sicht.suche,
            sicht.ausfuehrung,
            sicht.sortierung,
            ab,
            ende - ab,
          )

          const fenster = auspacken(antwort, () => antwort.exemplare)

          setzeNachgeholt((vorher) => {
            if (
              vorher.schluessel !== schluessel ||
              register !== geltendes.current
            ) {
              return vorher
            }

            const zeilen = new Map(vorher.zeilen)
            zeilen.set(ab, fenster)

            return { schluessel: vorher.schluessel, zeilen }
          })

          // **Auch das Löschen der Meldung gehört zur Sicht** (§7, Regel 2):
          // ein später Erfolg der **alten** Sicht nähme sonst einen Fehler
          // weg, der zur neuen gehört.
          if (gilt()) {
            setzeFehler(undefined)
          }
        } catch (ursache: unknown) {
          // **Nichts eintragen** (§7, Regel 3): ein leeres Fenster im Speicher
          // wäre die Behauptung, dort stünde nichts — und es würde nie wieder
          // geholt. So bleiben die Platzhalter, und die Meldung sagt, warum.
          //
          // **Und nur in der Sicht, aus der der Ruf stammt** — Befund 6 des
          // Nachlaufs zum Zweitblick über Phase 5.2. Der Erfolgsweg prüfte
          // den Schlüssel, der Fehlerweg nicht: ein abgelehnter Ruf von
          // vorhin hinterließ seine Meldung in einer Sicht, die ihn nie
          // gestellt hat.
          if (gilt()) {
            setzeFehler(meldungAus(ursache, texte))
          }

          register.bereiche.splice(register.bereiche.indexOf(bereich), 1)
        }
      })()
    },
    [
      holtAlles,
      nachgeholt,
      register,
      schluessel,
      setzeFehler,
      setzeNachgeholt,
      sicht.ausfuehrung,
      sicht.sortierung,
      sicht.suche,
      texte,
    ],
  )
}
