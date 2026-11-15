// **Was die bezogenen Stände tun** — G04
// (#2251).
// Wie sie aussehen, steht in `seiten/Bezugsstaende.tsx` (§2).
//
// # Die vier Regeln aus §7
//
//  1. **Ein Wechsel räumt den offenen Kontext**: wer einen anderen Baum
//     ansieht, verwirft die Vorschau des vorigen — samt ihrer Bestätigung.
//  2. **Eine Antwort gehört zu ihrem Stand** (`useLauf`): eine späte
//     Vorschau landet nicht über einer neueren.
//  3. **Geschrieben ist geschrieben**: nach einer geglückten Aktualisierung
//     wird nachgeladen, nach einer gescheiterten nicht — die Meldung bleibt.
//  4. **Ein Vorgang je Baum** (`useVorgang`).
//
// **Gefragt wird nur auf Knopfdruck**: ohne ihn geht nichts ins Netz (§16,
// Regel 5).

import { useCallback, useState } from 'react'

import { api, auspacken } from '../api'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useLaden } from './useLaden'
import { useLauf } from './useLauf'
import { useVorgang } from './useVorgang'

import type { BezugsvorschauAntwort, PaketbezugDTO } from '../paketbezuege'

export interface Bezugsstaende {
  bezuege: PaketbezugDTO[] | undefined
  laedt: boolean
  fehler: string | undefined

  // gefragt sagt, ob die Quellen schon gefragt wurden.
  gefragt: boolean
  fragen: () => void

  // vorschau ist die offene Vorschau — und zu welchem Baum sie gehört.
  vorschau: BezugsvorschauAntwort | undefined
  vorschauFuer: string | undefined
  vorschauFehler: string | undefined

  ansehen: (bezug: PaketbezugDTO) => Promise<void>
  // aktualisieren führt die Vorschau aus, die dasteht — **ihren** Commit.
  aktualisieren: (
    vorschau: BezugsvorschauAntwort,
    bestaetigt: boolean,
  ) => Promise<void>
  zurueck: (bezug: PaketbezugDTO) => Promise<void>
  verwerfen: () => void

  // laeuft sagt, ob für diesen Baum gerade etwas unterwegs ist.
  laeuft: (name: string) => boolean
}

export function useBezugsstaende(geaendert: () => void): Bezugsstaende {
  const texte = useTexte()
  const lauf = useLauf()
  const vorgang = useVorgang()

  const [gefragt, setzeGefragt] = useState(false)
  const [vorschau, setzeVorschau] = useState<BezugsvorschauAntwort>()
  const [vorschauFuer, setzeVorschauFuer] = useState<string>()
  const [vorschauFehler, setzeVorschauFehler] = useState<string>()

  const holen = useCallback(async () => {
    const antwort = await api().PaketBezuege(gefragt)

    return auspacken(antwort, () => antwort.bezuege)
  }, [gefragt])

  // **Die Liste bleibt stehen, während gefragt wird** — es ist derselbe
  // Bezug, nur genauer.
  const liste = useLaden<PaketbezugDTO[]>(holen, [gefragt], true)

  const verwerfen = useCallback(() => {
    lauf.naechster()
    setzeVorschau(undefined)
    setzeVorschauFuer(undefined)
    setzeVorschauFehler(undefined)
  }, [lauf])

  const ansehen = useCallback(
    async (bezug: PaketbezugDTO) => {
      const stand = lauf.naechster()
      setzeVorschau(undefined)
      setzeVorschauFuer(bezug.name)
      setzeVorschauFehler(undefined)

      await vorgang.starte(bezug.name, async () => {
        try {
          const antwort = await api().PaketBezugVorschau(
            bezug.name,
            bezug.spitze,
          )
          const gelesen = auspacken(antwort, () => antwort)

          if (lauf.istAktuell(stand)) {
            setzeVorschau(gelesen)
          }
        } catch (ursache: unknown) {
          if (lauf.istAktuell(stand)) {
            setzeVorschauFehler(meldungAus(ursache, texte))
          }
        }
      })
    },
    [lauf, vorgang, texte],
  )

  const aktualisieren = useCallback(
    async (gesehen: BezugsvorschauAntwort, bestaetigt: boolean) => {
      const { name, zu } = gesehen
      const stand = lauf.nummer()

      await vorgang.starte(name, async () => {
        setzeVorschauFehler(undefined)

        try {
          const antwort = await api().PaketBezugAktualisieren(
            name,
            zu,
            bestaetigt,
          )
          auspacken(antwort, () => undefined)

          // **Nur, was noch zu ihm gehört**, wird verworfen (§7, Regel 2;
          // Befund 7 des Zweitblicks über Phase G, zweiter Teil): wer
          // inzwischen einen anderen Baum ansieht, behält dessen Vorschau.
          // Nachgeladen wird trotzdem — geschrieben ist geschrieben.
          if (lauf.istAktuell(stand)) {
            verwerfen()
          }

          liste.nachladen()
          geaendert()
        } catch (ursache: unknown) {
          if (lauf.istAktuell(stand)) {
            setzeVorschauFehler(meldungAus(ursache, texte))
          }
        }
      })
    },
    [lauf, vorgang, texte, verwerfen, liste, geaendert],
  )

  const zurueck = useCallback(
    async (bezug: PaketbezugDTO) => {
      verwerfen()
      const stand = lauf.nummer()

      await vorgang.starte(bezug.name, async () => {
        try {
          const antwort = await api().PaketBezugZurueck(bezug.name)
          auspacken(antwort, () => undefined)
          liste.nachladen()
          geaendert()
        } catch (ursache: unknown) {
          if (lauf.istAktuell(stand)) {
            setzeVorschauFuer(bezug.name)
            setzeVorschauFehler(meldungAus(ursache, texte))
          }
        }
      })
    },
    [verwerfen, lauf, vorgang, texte, liste, geaendert],
  )

  return {
    bezuege: liste.daten,
    laedt: liste.laedt,
    fehler: liste.fehler,
    gefragt,
    // **Ein zweites Fragen fragt neu**; das erste ändert die Frage, und
    // `useLaden` lädt dann ohnehin.
    fragen: () => {
      if (gefragt) {
        liste.nachladen()
      } else {
        setzeGefragt(true)
      }
    },
    vorschau,
    vorschauFuer,
    vorschauFehler,
    ansehen,
    aktualisieren,
    zurueck,
    verwerfen,
    laeuft: vorgang.laeuft,
  }
}
