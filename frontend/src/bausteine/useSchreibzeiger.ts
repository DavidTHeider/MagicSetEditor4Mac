// **Wo der Schreibzeiger steht, und was daraus folgt** — abgetrennt am
// 2026-09-02, weil `Quelltextfeld.tsx` mit den Klammerpaaren über die
// Zeilengrenze aus §2 ging.
//
// # Der Schnitt folgt der Frage
//
// Zwei Dinge im Feld hängen an **einer** Zahl: das angefangene Wort (für die
// Vervollständigung) und das Klammerpaar unter dem Zeiger. Beide beantworten
// „wo steht der Zeiger, und was folgt daraus"; daneben in `Quelltextfeld.tsx`
// steht, **wie** das Feld aussieht und sich bedienen lässt.
//
// **`-1` heißt: der Zeiger steht nicht im Feld.** Dann gibt es weder
// vorzuschlagen noch hervorzuheben — und der Wert ist keine Erfindung dieser
// Datei, sondern der Anfangszustand, den es geben muss: vor dem ersten Klick
// hat niemand irgendwo hingezeigt.
import { useState } from 'react'

import { klammerpaar } from './klammerpaare'
import { passende, wortstelle } from './wortstelle'

import type { Klammerpaar } from './klammerpaare'
import type { Merkmalstelle } from '../api'
import type { Wortstelle } from './wortstelle'

// hoechstensVorschlaege ist, wie viele Namen zugleich dastehen.
//
// **Acht.** Am laufenden Fenster gezählt gibt es an einer Karte mit Vorlage
// 1.644 Namen; eine Liste ohne Deckel wäre keine Hilfe, sondern eine Wand.
const hoechstensVorschlaege = 8

// Schreibzeiger ist, was an der Zeigerstelle hängt.
export interface Schreibzeiger {
  stelle: number
  setzeStelle: (stelle: number) => void

  wort: Wortstelle
  vorschlaege: string[]
  paar: Klammerpaar | null
}

// useSchreibzeiger führt die Zeigerstelle und was daran hängt.
export function useSchreibzeiger(
  wert: string,
  merkmale: Merkmalstelle[],
  namen: readonly string[],
): Schreibzeiger {
  const [stelle, setzeStelle] = useState(-1)

  const wort =
    stelle < 0 ? { wort: '', von: 0, bis: 0 } : wortstelle(wert, stelle)

  return {
    stelle,
    setzeStelle,
    wort,
    vorschlaege: passende(namen, wort.wort, hoechstensVorschlaege),
    paar: stelle < 0 ? null : klammerpaar(merkmale, wert, stelle),
  }
}
