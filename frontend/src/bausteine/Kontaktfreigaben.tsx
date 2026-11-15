// **Die erteilten Sichten eines Gegenübers laden und ändern** — Werkstück T3
// von Phase 5.6
// (#1468).
//
// # Warum das Laden hier steht und nicht auf der Seite
//
// **Weil es je Gegenüber eine eigene Frage ist.** Die Seite kennt die Liste
// der Kontakte; welche Sichten einem davon gelten, ist eine Frage an **diesen
// einen**. Sie in der Seite zu halten hieße, eine Tabelle über alle Kontakte
// zu führen — und die wäre nach jeder Änderung an einem von ihnen zur Hälfte
// alt.
//
// # Und warum die Vorschau geräumt wird, sobald sich etwas ändert
//
// **§7, Regel 1**: wer eine Freigabe erteilt oder widerruft, hat einen anderen
// Stand vor sich. Eine stehengebliebene Vorschau zeigte, was **vorher** galt
// — und das ist die gefährlichste Auskunft, die dieser Kasten geben kann.
//
// # Und warum nachgeladen und nicht neu gefragt wird
//
// **Weil beides in [useLaden] nicht dasselbe ist.** Ein Zähler in den
// Abhängigkeiten ändert die **Frage** — der Haken hält das für einen Wechsel
// des Bezugsobjekts, räumt den Stand und nimmt den Kasten mit; er kommt mit
// der ersten Wahl zurück. Gemessen am Fenster am 2026-09-18: wer eine Sicht
// für `teilen` erteilte und danach „Ansehen, was es sieht" drückte, sah die
// Sicht für `handel` — zwei Zeilen, wo die eben erteilte **null** zeigt.
//
// # Und warum ein Befund zu einem Feld nicht über dem Kasten steht
//
// **Weil §6 es verlangt** (UXR-F18,
// #2033):
// die Domäne nennt das Feld, an dem eine Freigabe scheitert (`auswahl`,
// `behalten`), und die Meldung gehört dorthin. Was kein Feld nennt, bleibt
// ein Satz über dem Kasten.
import { useState } from 'react'

import { Sichtfreigaben, type Freigabebefunde } from './Sichtfreigaben'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { feldbefund, meldungAus } from '../meldung'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'

import type {
  Sichtfreigabe,
  Sichtfreigabenantwort,
  Sichtkarte,
} from '../apisichtfreigabe'

interface Kontaktfreigabeneigenschaften {
  kennung: string
  zwecke: string[]
}

export function Kontaktfreigaben({
  kennung,
  zwecke,
}: Kontaktfreigabeneigenschaften) {
  const texte = useTexte()
  const vorgang = useVorgang()

  const [fehler, setzeFehler] = useState<string>()
  const [befunde, setzeBefunde] = useState<Freigabebefunde>({})
  const [vorschau, setzeVorschau] = useState<Sichtkarte[]>()

  const liste = useLaden(async () => {
    const antwort = await api().Sammlungsfreigaben(kennung)

    return auspacken(antwort, () => antwort)
  }, [kennung])

  const kennzeichen = `sichtfreigaben/${kennung}`
  const laeuft = vorgang.laeuft(kennzeichen)

  // handeln fährt einen der drei Wege.
  //
  // **Ein Vorgang für alle drei** (§7, Regel 4): sie gehen an dieselbe Liste.
  const handeln = (weg: () => Promise<unknown>): Promise<void> =>
    vorgang.starte(kennzeichen, async () => {
      setzeFehler(undefined)
      setzeBefunde({})

      try {
        await weg()
      } catch (ursache) {
        const amFeld: Freigabebefunde = {
          auswahl: feldbefund(ursache, 'auswahl', texte),
          behalten: feldbefund(ursache, 'behalten', texte),
        }

        if (amFeld.auswahl !== undefined || amFeld.behalten !== undefined) {
          setzeBefunde(amFeld)
        } else {
          setzeFehler(meldungAus(ursache, texte))
        }
      }
    })

  const schreiben = (
    weg: () => Promise<Sichtfreigabenantwort>,
  ): Promise<void> =>
    handeln(async () => {
      auspacken(await weg(), () => undefined)

      // **Die Vorschau gehört zum vorigen Stand** — §7, Regel 1.
      setzeVorschau(undefined)

      // **Nachladen und nicht neu fragen.** [useLaden] unterscheidet beides
      // an den Abhängigkeiten: ein Zähler darin machte aus jeder Erteilung
      // einen **Wechsel des Bezugsobjekts**, und der räumt — der Kasten
      // verschwände und käme mit der ersten Wahl zurück.
      liste.nachladen()
    })

  return (
    <>
      {liste.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {liste.fehler}
        </p>
      )}

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      {liste.daten !== undefined && (
        <Sichtfreigaben
          kennung={kennung}
          freigaben={liste.daten.freigaben}
          umfaenge={liste.daten.umfaenge}
          zwecke={zwecke}
          vorschau={vorschau}
          laeuft={laeuft}
          befunde={befunde}
          befundGeraeumt={() => {
            setzeBefunde({})
          }}
          erteilen={(freigabe: Sichtfreigabe) => {
            void schreiben(() => api().SammlungFreigeben(freigabe))
          }}
          widerrufen={(zweck, umfang) => {
            void schreiben(() =>
              api().SammlungsfreigabeWiderrufen(kennung, zweck, umfang),
            )
          }}
          zeigen={(zweck) => {
            void handeln(async () => {
              const antwort = await api().Freigabesicht(kennung, zweck)

              setzeVorschau(auspacken(antwort, () => antwort.zeilen))
            })
          }}
        />
      )}
    </>
  )
}
