// **Wen der Vorrang fragt, für das Formular** — #2054, aus AI-02 (#1743).
//
// # Erst auf Zuruf, und nicht ewig
//
// **Geladen wird erst, wenn jemand die Feldzeile berührt** (`beruehrt`): der
// Katalog fragt auch eingeschaltete Cloudanbieter nach ihren Modellen, und
// das soll nicht bei jeder Karte geschehen. **Eine Minute lang für alle
// Felder** — Name und Flavortext fragten sonst zweimal dasselbe.
//
// **Und ein Feld, das schon steht, folgt** (Z5 aus dem Zweitblick über Phase
// 6.11): eine geänderte KI-Einstellung benachrichtigt jedes Feld, das die
// Reihe zeigt, und wer nach der Minute wieder berührt, bekommt eine frische.
// Vorher wurde nur der Speicher geleert — das Feld zeigte weiter die alte.
//
// **Mit den Befunden der Reihe** (Z4): ein gewähltes Modell, das fehlt, stand
// an Kommandozeile, MCP und REST, am Formular nicht.
//
// **Ein Fehler zeigt nichts**: die Reihe ist ein Hinweis; ob der Vorschlag
// selbst geht, sagt der Vorschlag.
import { useCallback, useEffect, useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'

// frist ist, wie lange eine gelesene Reihe gilt.
const frist = 60_000

interface Gelesen {
  reihe: string[]
  befunde: string[]
}

export interface Vorrangreihe {
  // reihe ist `undefined`, solange nichts gelesen ist.
  reihe: string[] | undefined
  befunde: string[]

  // beruehrt sagt, dass jemand die Zeile berührt hat — beim ersten Mal und
  // nach der Frist wird gefragt.
  beruehrt: () => void
}

let gemerkt: { seit: number; antwort: Promise<Gelesen> } | undefined
const lauscher = new Set<() => void>()

// vergissVorrangreihe räumt den Speicher und sagt es jedem Feld, das die
// Reihe zeigt — nach einer Änderung der Einstellung und in Tests.
export function vergissVorrangreihe(): void {
  gemerkt = undefined

  for (const neu of lauscher) neu()
}

function frisch(): boolean {
  return gemerkt !== undefined && Date.now() - gemerkt.seit <= frist
}

function holen(): Promise<Gelesen> {
  if (gemerkt !== undefined && frisch()) return gemerkt.antwort

  const frage = api().KIModelle()
  const antwort = frage.then((gelesen) => {
    const katalog = auspacken(gelesen, () => gelesen)

    return { reihe: katalog.vorrang, befunde: katalog.vorrangbefunde }
  })

  gemerkt = { seit: Date.now(), antwort }
  antwort.catch(() => {
    gemerkt = undefined
  })

  return antwort
}

export function useVorrangreihe(): Vorrangreihe {
  // anlass zählt, wie oft gefragt werden soll — 0 heißt: noch nie berührt.
  const [anlass, setzeAnlass] = useState(0)
  const [gelesen, setzeGelesen] = useState<Gelesen>()

  useEffect(() => {
    const neu = () => {
      setzeAnlass((bisher) => (bisher === 0 ? 0 : bisher + 1))
    }

    lauscher.add(neu)

    return () => {
      lauscher.delete(neu)
    }
  }, [])

  useEffect(() => {
    if (anlass === 0) return undefined

    let aktuell = true

    // **Ein Fehler zeigt nichts** — auch nicht die Reihe von vorhin, die
    // nach einer Änderung nicht mehr gilt (N5 aus dem Nachlauf über 6.11).
    holen().then(
      (antwort) => {
        if (aktuell) setzeGelesen(antwort)
      },
      () => {
        if (aktuell) setzeGelesen(undefined)
      },
    )

    return () => {
      aktuell = false
    }
  }, [anlass])

  const beruehrt = useCallback(() => {
    setzeAnlass((bisher) => (bisher === 0 || !frisch() ? bisher + 1 : bisher))
  }, [])

  return { reihe: gelesen?.reihe, befunde: gelesen?.befunde ?? [], beruehrt }
}
