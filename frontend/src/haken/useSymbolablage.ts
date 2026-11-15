// **Die Zeichnung neben das Set legen**
// (#890).
//
// **Ein eigener Haken**, seit `Symbol.tsx` mit der Symmetrie über die
// Zeilengrenze aus §2 wuchs
// (#1092).
// Der Schnitt folgt der Frage: dort, **was** der Editor zeigt; hier, **wie**
// das Ergebnis ins Set kommt.

import { useCallback } from 'react'

import type { SymbolteilDTO } from '../antworten'
import { meldungAus } from '../meldung'
import type {
  Bibliotheksbearbeitung,
  Symbolziel,
} from '../seiten/symbolablegen'
import {
  abgelegtesSymbol,
  inBibliothekGesichert,
} from '../seiten/symbolablegen'
import { useTexte } from '../sprache/kontext'
import type { Vorgang } from './useVorgang'

interface AblageEigenschaften {
  setID: string

  // ziel sagt, welches Feld die Zeichnung trägt — ohne Ziel das Setsymbol
  // (#1339).
  ziel?: Symbolziel | undefined

  // bibliothek sagt, dass eine neue Fassung in der Bibliothek gesichert wird
  // und nichts ins Set geht (DOM-02.AK-02).
  bibliothek?: Bibliotheksbearbeitung | undefined

  teile: SymbolteilDTO[]

  // geladenerName ist die Datei, die im Feldwert steht.
  geladenerName: string

  // name ist die Datei aus **dieser** Sitzung — leer, solange nichts abgelegt
  // wurde.
  name: string

  setzeName: (name: string) => void
  setzeFehler: (meldung: string | undefined) => void
  vorgang: Vorgang
}

export interface Symbolablage {
  // abgelegterName ist die Datei, die dieses Set **heute** als Symbol führt.
  //
  // **Erst der Name aus dieser Sitzung, dann der geladene.** Wer eben abgelegt
  // hat, hat den frischen; wer das Set gerade geöffnet hat, den aus dem
  // Feldwert. Beides ist derselbe Satz: *diese* Datei ist das Symbol.
  abgelegterName: string

  ablegen: () => Promise<void>
}

export function useSymbolablage({
  setID,
  ziel,
  bibliothek,
  teile,
  geladenerName,
  name,
  setzeName,
  setzeFehler,
  vorgang,
}: AblageEigenschaften): Symbolablage {
  const texte = useTexte()

  const abgelegterName = name !== '' ? name : geladenerName

  // ablegen legt die Zeichnung neben das Set **und macht sie zum Setsymbol**.
  //
  // # Warum beides zusammengehört
  //
  // Die Datei allein tut nichts: gezeichnet wird, was in `set.symbol` steht.
  // Der erste Anlauf legte nur ab — die Zeichnung lag neben dem Set, und auf
  // der Karte blieb MSEs Standardquadrat. Angestrichen vom Zweitblick über
  // Phase X; §10 nennt es zuerst: **woran ist zu sehen, dass es geht?**
  //
  // **Nacheinander und nicht nebeneinander.** Der Feldwert zeigt auf die
  // Datei; wer ihn setzt, bevor sie liegt, hat einen Verweis ins Leere. Und
  // scheitert das Ablegen, wird gar nicht erst gesetzt (§7, Regel 3).
  //
  // **Ein Vorgang je Set** (§7, Regel 4): zwei gleichzeitige Ablagen gäben
  // zwei Dateien, und die zweite wäre die, auf die niemand zeigt.
  const ablegen = useCallback(async () => {
    const abgelegt = await vorgang.starte(setID, async () => {
      try {
        if (bibliothek !== undefined) {
          setzeName(await inBibliothekGesichert(bibliothek, teile))

          return bibliothek.name
        }

        return await abgelegtesSymbol(
          setID,
          teile,
          abgelegterName,
          setzeName,
          ziel,
        )
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))

        return undefined
      }
    })

    if (abgelegt !== undefined) {
      setzeFehler(undefined)
    }
  }, [
    setID,
    ziel,
    bibliothek,
    teile,
    abgelegterName,
    setzeName,
    setzeFehler,
    vorgang,
    texte,
  ])

  return { abgelegterName, ablegen }
}
