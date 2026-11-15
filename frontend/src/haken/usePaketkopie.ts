// **Was die eigene Kopie in der Werkstatt tut** — FMT-02.AK-04 und
// S10.NAT-01 (#1691).
//
// Ein mitgeliefertes oder fremdverwaltetes Paket wird über eine eigene Kopie
// angepasst. **Zwei Stellen bieten sie an**: der Baum unter einem Paket, das
// sich nur lesen lässt — die Saat liegt außerhalb der Werkstatt, ihre Dateien
// lassen sich nicht einmal aufzählen —, und der Hinweis über dem Editor eines
// bezogenen oder übernommenen Pakets. Der Haken legt die Kopie an und öffnet
// in ihr dieselbe Datei oder, vom Baum aus, ihre Hauptdatei (§2).
//
// # Die vier Regeln aus §7
//
// **Regel 1**: ein Fehler gehört zum Paket, an dem er entstand, und steht
// nur dort. **Regel 2**: eine späte Antwort wechselt nicht mehr, wenn
// inzwischen eine andere Datei gewählt ist — sie legt die Kopie an, aber sie
// entführt die Wahl nicht. **Regel 3**: nach einer geglückten Kopie wird die
// Liste nachgeladen, nach einer gescheiterten nicht. **Regel 4**: ein Vorgang
// je Paket.
import { useEffect, useRef, useState } from 'react'

import { api } from '../fassade'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { useVorgang } from './useVorgang'

import type { Paketherkunftsart } from '../pakete'
import type { Vorlagenwerkstatt } from './useVorlagenwerkstatt'

export interface Paketkopie {
  // herkunftsart ist die des gewählten Pakets — undefined ohne Wahl.
  herkunftsart: Paketherkunftsart | undefined

  // original nennt das Paket, von dem das gewählte eine Kopie ist — nur für
  // Kopien, die in dieser Sitzung entstanden.
  original: string | undefined

  // laeuft und fehler gelten je Paket, genannt über seinen Pfad.
  laeuft: (pfad: string) => boolean
  fehler: (pfad: string) => string | undefined

  kopieren: (pfad: string) => Promise<void>
}

type Werkstattanschluss = Pick<
  Vorlagenwerkstatt,
  'wahl' | 'eintraege' | 'waehlen' | 'paketeNeuLaden'
>

export function usePaketkopie(werkstatt: Werkstattanschluss): Paketkopie {
  const texte = useTexte()
  const vorgang = useVorgang()

  const { wahl } = werkstatt
  const paket = wahl?.paket ?? ''
  const datei = wahl?.datei ?? ''

  // kopien ordnet dem Pfad einer Kopie den Namen ihres Originals zu.
  const [kopien, setzeKopien] = useState<Record<string, string>>({})
  const [fehler, setzeFehler] = useState<Record<string, string>>({})

  // **Die Wahl wird mitgezählt** — über einen Effekt, denn eine Antwort
  // muss den Stand von jetzt sehen und nicht den ihres Aufrufs.
  const wahlstand = useRef(0)
  useEffect(() => {
    wahlstand.current += 1
  }, [paket, datei])

  const merkeFehler = (pfad: string, text: string | undefined) => {
    setzeFehler((bisher) =>
      text === undefined
        ? Object.fromEntries(
            Object.entries(bisher).filter(
              ([schluessel]) => schluessel !== pfad,
            ),
          )
        : { ...bisher, [pfad]: text },
    )
  }

  return {
    herkunftsart: werkstatt.eintraege.find((e) => e.pfad === paket)
      ?.herkunftsart,
    original: kopien[paket],
    laeuft: (pfad) => vorgang.laeuft(pfad),
    fehler: (pfad) => fehler[pfad],

    kopieren: async (pfad) => {
      const eintrag = werkstatt.eintraege.find((e) => e.pfad === pfad)

      // **Ein Pfad, den die Liste nicht kennt, kopiert nichts** — die Liste
      // kann zwischen Zeichnen und Klick nachgeladen worden sein.
      if (eintrag === undefined) {
        return
      }

      const stand = wahlstand.current
      const ziel = paket === pfad ? datei : undefined

      await vorgang.starte(pfad, async () => {
        try {
          const antwort = await api().PaketKopieren(
            eintrag.art,
            eintrag.name,
            '',
          )
          const neu = auspacken(antwort, () => antwort.paket)

          merkeFehler(pfad, undefined)
          werkstatt.paketeNeuLaden()

          if (neu === null) {
            return
          }

          setzeKopien((bisher) => ({ ...bisher, [neu.pfad]: eintrag.name }))

          if (wahlstand.current === stand) {
            werkstatt.waehlen(neu.pfad, ziel ?? neu.hauptdatei)
          }
        } catch (ursache: unknown) {
          merkeFehler(pfad, meldungAus(ursache, texte))
        }
      })
    },
  }
}
