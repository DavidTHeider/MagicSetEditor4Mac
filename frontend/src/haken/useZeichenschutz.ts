// **Der Wächter über die offene Zeichnung** — Befund 2 des Zweitblicks über
// Phase 4.11.
//
// # Warum es ihn gibt
//
// Jede Umformung im Symboleditor schickt den **vollständigen** Zeichenstand an
// eine Bindung und setzt die Antwort danach ganz ein: `Grundform`,
// `SymbolNachzeichnen`, die drei Wege der Teilliste, die Punktwege und die
// Symmetrie. Ohne Wächter laufen zwei davon parallel, und beide rechnen gegen
// den Stand von **vorher** — die zweite Antwort verliert, was die erste wollte,
// gleich in welcher Reihenfolge sie eintreffen. Antwortet die ältere zuletzt,
// überschreibt sie zusätzlich den neueren Stand und legt einen irreführenden
// Rücknahmeschritt an.
//
// # Die beiden Regeln aus §7, an einer Stelle
//
//   - **Regel 4**: ein Vorgang je Datensatz. Die Zeichnung ist **einer** — es
//     läuft höchstens eine Umformung.
//   - **Regel 2**: eine Antwort gehört zu dem Stand, aus dem sie stammt. Wer
//     das Bezugsobjekt wechselt, verwirft, was noch unterwegs ist.
//
// **Warum nicht `useVorgang` und `useLauf` einzeln?** Weil sie hier dasselbe
// Bezugsobjekt meinen und immer zusammen gebraucht werden. Zwei Haken an jeder
// der sechs Aufrufstellen wären sechs Gelegenheiten, einen davon zu vergessen.

import { useCallback, useMemo, useRef } from 'react'

export interface Zeichenschutz {
  // starte führt eine Umformung aus — oder gar nicht, wenn schon eine läuft.
  //
  // `nochAktuell` sagt der Arbeit, ob ihre Antwort noch zum heutigen Stand
  // gehört. **Sie muss danach fragen, bevor sie einträgt** — der Wächter kann
  // das nicht für sie tun, denn nur sie weiß, was ihr Ergebnis ist.
  starte: (
    arbeit: (nochAktuell: () => boolean) => Promise<void>,
  ) => Promise<void>

  // raeume erklärt jede laufende Antwort für überholt — **und gibt frei**.
  //
  // **Für den Wechsel des Bezugsobjekts** (§7, Regel 1): wer ein anderes Set
  // oder ein anderes Symbol aufschlägt, will nicht, dass die Antwort von
  // vorhin noch ankommt.
  //
  // **Das Freigeben gehört dazu.** Mit dem Hochzählen allein bliebe der
  // Vorgang bis zum Ende des alten Rufs belegt: hängt die Naht, wäre auch das
  // **neue** Symbol für jede Handlung gesperrt. Der alte Ruf räumt danach
  // nichts mehr weg — er sieht an seinem Stand, dass er nicht mehr gemeint
  // ist.
  raeume: () => void

  // beobachte nimmt den heutigen Stand und gibt die Frage, ob er noch gilt.
  //
  // **Für einen Weg, der sich nicht sperren lässt.** `GriffZiehen` schickt
  // eine Anfrage je Zeigerbewegung; unter `starte` käme nur die erste durch,
  // und der Griff bliebe am Rand kleben. Was er trotzdem braucht, ist §7
  // Regel 2: ein Wechsel des Bezugsobjekts verwirft, was noch unterwegs ist.
  beobachte: () => () => boolean
}

export function useZeichenschutz(): Zeichenschutz {
  const stand = useRef(0)
  const unterwegs = useRef(false)

  const raeume = useCallback(() => {
    stand.current += 1
    unterwegs.current = false
  }, [])

  const beobachte = useCallback(() => {
    const meiner = stand.current

    return () => meiner === stand.current
  }, [])

  const starte = useCallback(
    async (arbeit: (nochAktuell: () => boolean) => Promise<void>) => {
      if (unterwegs.current) {
        return
      }

      unterwegs.current = true

      const nochAktuell = beobachte()

      try {
        await arbeit(nochAktuell)
      } finally {
        // **Auch nach einem Fehler.** Sonst bliebe der Editor nach einem
        // einzigen gescheiterten Ruf für immer stumm — und §7 Regel 3 sagt,
        // dass ein Fehler die Arbeit nicht mitnimmt.
        //
        // **Aber nur den eigenen Vorgang.** Wer abgeräumt wurde, hat längst
        // freigegeben, und ein zweiter Ruf läuft vielleicht schon; ihn hier
        // freizugeben hieße, zwei zugleich zuzulassen.
        if (nochAktuell()) {
          unterwegs.current = false
        }
      }
    },
    [beobachte],
  )

  // **Ein Verbund, der derselbe bleibt** — dieselbe Begründung wie bei
  // `useLauf`: ohne `useMemo` liefe jeder Effekt mit ihm in den
  // Abhängigkeiten bei jedem Zeichnen erneut.
  return useMemo(
    () => ({ starte, raeume, beobachte }),
    [starte, raeume, beobachte],
  )
}
