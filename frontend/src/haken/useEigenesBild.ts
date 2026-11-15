// **Ein eigenes Bild im Fenster** — das Tischbild aus Werkstück G3 von Phase
// 5.5 (#1460)
// und seit UX-09 (#1681)
// das Fensterbild. **Ein Haken für beide**, weil beide dieselben drei Wege
// haben — lesen, wählen, entfernen — und dieselben Regeln aus §7; nur die
// Bindungen sind andere ([Bildwege]).
//
// # Warum er den Nutzer selbst sucht
//
// **Weil das Fenster in einem Profil arbeitet und nicht in einem Nutzer** —
// dieselbe Lage wie bei `useNutzereinstellungen`, und derselbe Weg: wem das
// gezeigte Profil gehört, sagt `GezeigtesProfil`, **in demselben Zug** (§7,
// Regel 2). Zwei Rufe wären zwei Stände, und im Zwischenraum gehörte das Bild
// jemand anderem.
import { useCallback, useState } from 'react'

import { api, auspacken } from '../api'
import type { TischbildAntwort } from '../apitischbild'
import { meldungAus } from '../meldung'
import { useLaden } from './useLaden'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from './useVorgang'

export interface Eigenbildstand {
  // adresse ist leer, solange kein eigenes Bild abgelegt ist.
  adresse: string

  // waehlen öffnet den Dateidialog.
  waehlen: () => void

  // entfernen nimmt das Bild weg — danach gilt wieder, was ohne Bild gilt.
  entfernen: () => void

  // stand zählt die geglückten Wechsel — **er gehört an die Adresse**, wo
  // ein Stilblatt das Bild zeigt: dieselbe Adresse trüge sonst im Speicher
  // des WebKit das alte Bild weiter.
  stand: number

  laeuft: boolean

  // bereit heißt: der Mensch ist bekannt. **Vorher ginge ein Wählen ins
  // Leere** — wer den Knopf zeigt, hält ihn bis dahin an, statt ihn nichts
  // tun zu lassen.
  bereit: boolean
  fehler: string | undefined
}

// Bildwege sind die drei Bindungen eines Bildes.
export interface Bildwege {
  lesen: (wem: string) => Promise<TischbildAntwort>
  waehlen: (wem: string) => Promise<TischbildAntwort>
  entfernen: (wem: string) => Promise<TischbildAntwort>
}

// **`bekannt` ist der Mensch, wenn der Aufrufer ihn schon kennt** — der
// `Gestaltungsraum` hat ihn aus dem Nutzereinstellungsraum und fragt
// `GezeigtesProfil` deshalb nicht ein zweites Mal.
export function useEigenesBild(
  wege: Bildwege,
  bekannt?: string,
): Eigenbildstand {
  // gewaehlt ist, was seit dem Laden gewählt wurde, **mit dem Menschen, der
  // es gewählt hat** — `undefined` heißt: es gilt noch, was der Bestand sagt.
  const [gewaehlt, setzeGewaehlt] = useState<
    { wem: string; adresse: string } | undefined
  >(undefined)
  const [stand, setzeStand] = useState(0)
  const [schreibfehler, setzeSchreibfehler] = useState<string | undefined>(
    undefined,
  )
  const vorgang = useVorgang()
  const texte = useTexte()

  const geladen = useLaden(
    useCallback(
      async () => bildLesen(wege, bekannt ?? (await gezeigterNutzer())),
      [wege, bekannt],
    ),
    [wege, bekannt],
  )

  const nutzerID = geladen.daten?.nutzerID ?? ''

  // **Ein Vorgang je Datensatz** (§7, Regel 4): wer zweimal hintereinander
  // wählt, während der Dialog noch offen ist, stellte zwei Fragen an einen
  // Stand, den der Wirt noch nicht kennt.
  const laufenLassen = useCallback(
    (was: (wem: string) => Promise<string>) => {
      if (nutzerID === '') return

      void vorgang.starte(nutzerID, async () => {
        try {
          setzeGewaehlt({ wem: nutzerID, adresse: await was(nutzerID) })
          setzeStand((bisher) => bisher + 1)
          setzeSchreibfehler(undefined)
        } catch (grund) {
          // **Ein gescheitertes Schreiben lädt nicht nach** (§7, Regel 3):
          // ein Lesefehler überschriebe sonst die Meldung, die sagt, was zu
          // tun ist.
          //
          // **Und übersetzt wird über den Katalog** (§9) — `meldungAus` kennt
          // die vier Gestalten, in denen ein Grund ankommt; ein eigenes
          // `instanceof Error` hier wäre die fünfte und die unübersetzte.
          setzeSchreibfehler(meldungAus(grund, texte))
        }
      })
    },
    [nutzerID, texte, vorgang],
  )

  // **Was gilt, gehört dem Menschen, um den es geht** (Nachlauf über Phase
  // 6.3, N6): ein Bild eines anderen zählt nicht — nach einem Wechsel stand
  // sonst sein geladenes da, bis das neue gelesen war, und ein in dieser
  // Sitzung gewähltes für immer. Bis dahin hielt das nur der `key` am
  // `Gestaltungsraum`.
  const wer = bekannt ?? nutzerID
  const geladenesBild =
    geladen.daten?.nutzerID === wer ? geladen.daten.adresse : ''

  return {
    adresse: gewaehlt?.wem === wer ? gewaehlt.adresse : geladenesBild,
    laeuft: vorgang.laeuft(nutzerID),
    bereit: nutzerID !== '',
    fehler: schreibfehler ?? geladen.fehler,
    stand,

    waehlen: useCallback(() => {
      laufenLassen(async (wem) => {
        const antwort = await wege.waehlen(wem)

        return auspacken(antwort, () => antwort.adresse)
      })
    }, [laufenLassen, wege]),

    entfernen: useCallback(() => {
      laufenLassen(async (wem) => {
        const antwort = await wege.entfernen(wem)

        return auspacken(antwort, () => antwort.adresse)
      })
    }, [laufenLassen, wege]),
  }
}

// bildLesen liest das Bild eines Menschen — **ohne Menschen keines**, und
// dann wird gar nicht erst gefragt.
async function bildLesen(
  wege: Bildwege,
  wem: string,
): Promise<{ nutzerID: string; adresse: string }> {
  if (wem === '') return { nutzerID: '', adresse: '' }

  const bild = await wege.lesen(wem)

  return { nutzerID: wem, adresse: auspacken(bild, () => bild.adresse) }
}

// gezeigterNutzer fragt, wem das gezeigte Profil gehört.
async function gezeigterNutzer(): Promise<string> {
  const profil = await api().GezeigtesProfil()

  return auspacken(profil, () => profil.profil?.nutzerID ?? '')
}
