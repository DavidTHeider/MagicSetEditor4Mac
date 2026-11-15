// **Was das Bearbeiten eines Symbols der Bibliothek tut** — DOM-02.AK-02
// (#1697).
//
// Der Haken hält, **welches** Symbol der Editor bearbeitet, welche Sets nach
// dem Sichern die vorige Fassung tragen, und das Nachziehen. Die Leiste
// zeichnet es (`seiten/Symbolbearbeitung.tsx`).
//
// # Die vier Regeln aus §7
//
// **Regel 1**: ein anderes Symbol räumt Ergebnis, Hinweis und Fehler — und ein
// anderes Set baut das ganze Fach neu. **Regel 2**: eine Antwort trägt sich
// nur ein, solange ihr Symbol offen ist — wer mitten im Nachziehen ein
// anderes öffnet, sähe sonst dessen Sets bei diesem. **Regel 3**: nach einem geglückten
// Nachziehen liest die Bibliothek nach, nach einem gescheiterten nicht.
// **Regel 4**: ein Vorgang je Symbol.
import { useRef, useState } from 'react'

import { api } from '../api'
import type { BibliothekssymbolDTO } from '../apisymbolbibliothek'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { useVorgang } from './useVorgang'

export interface Bearbeitetes {
  id: string
  name: string
}

export interface Symbolbearbeitungsstand {
  bearbeitet: Bearbeitetes | undefined

  // veraltet sind die Sets mit der vorigen Fassung — undefined, solange
  // nicht gesichert wurde.
  veraltet: string[] | undefined
  hinweis: string | undefined
  fehler: string | undefined
  laeuft: boolean

  oeffnen: (symbol: BibliothekssymbolDTO) => void
  schliessen: () => void
  gesichert: (id: string, veraltet: string[]) => void
  nachziehen: (symbol: Bearbeitetes) => Promise<void>
}

// useSymbolbearbeitung führt das Bearbeiten; `geaendert` sagt der
// Bibliothek, dass sie nachlesen soll.
export function useSymbolbearbeitung(
  geaendert: () => void,
): Symbolbearbeitungsstand {
  const texte = useTexte()
  const eigene = texte.seiten.karten.symbolbibliothek
  const vorgang = useVorgang()

  const [bearbeitet, setzeBearbeitet] = useState<Bearbeitetes>()
  const [veraltet, setzeVeraltet] = useState<string[]>()
  const [hinweis, setzeHinweis] = useState<string>()
  const [fehler, setzeFehler] = useState<string>()

  // offen ist die Kennung des bearbeiteten Symbols — über eine Referenz, weil
  // eine späte Antwort den Stand von **jetzt** sehen muss.
  const offen = useRef<string>(undefined)

  const raeumen = () => {
    setzeVeraltet(undefined)
    setzeHinweis(undefined)
    setzeFehler(undefined)
  }

  return {
    bearbeitet,
    veraltet,
    hinweis,
    fehler,
    laeuft: bearbeitet !== undefined && vorgang.laeuft(bearbeitet.id),

    oeffnen: (symbol) => {
      raeumen()
      offen.current = symbol.id
      setzeBearbeitet({ id: symbol.id, name: symbol.name })
    },

    schliessen: () => {
      raeumen()
      offen.current = undefined
      setzeBearbeitet(undefined)
    },

    gesichert: (id, neu) => {
      geaendert()

      if (offen.current === id) {
        raeumen()
        setzeVeraltet(neu)
      }
    },

    nachziehen: async (symbol) => {
      await vorgang.starte(symbol.id, async () => {
        try {
          const antwort = await api().BibliothekssymbolNachziehen(symbol.id)
          const sets = auspacken(antwort, () => antwort.sets)

          geaendert()

          if (offen.current === symbol.id) {
            raeumen()
            setzeHinweis(
              einsetzen(eigene.nachgezogen, { sets: sets.join(', ') }),
            )
          }
        } catch (ursache: unknown) {
          if (offen.current === symbol.id) {
            setzeFehler(meldungAus(ursache, texte))
          }
        }
      })
    },
  }
}
