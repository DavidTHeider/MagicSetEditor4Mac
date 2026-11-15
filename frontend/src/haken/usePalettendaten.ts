// **Was die Befehlspalette über den Bestand weiß** — UX-02.AK-02
// (#1674).
//
// # Wann geladen wird
//
// **Sets, Decks und Gegenüber einmal, wenn die Palette aufgeht** — sie ist
// abgebaut, solange sie zu ist (`Palette.tsx`), und ein neuer Stand kommt mit
// dem nächsten Öffnen. **Karten erst ab zwei Zeichen**: ein Buchstabe träfe
// jede zweite Karte, und die Volltextsuche ginge über alle Sets.
//
// # Was ein Fehler bedeutet
//
// **Nichts, was die Palette anhält.** Scheitert eine Quelle, fehlen ihre
// Zeilen; Seiten und Befehle stehen weiter da — dieselbe Regel wie beim
// Befehlskatalog seit U12.
import { useCallback } from 'react'

import { api } from '../api'
import { auspacken } from '../umschlag'
import { useLaden } from './useLaden'

import type { Palettendaten } from '../rahmen/palettenziele'
import type { TrefferDTO } from '../formen'

// Mindestens ist, ab wie vielen Zeichen Karten gesucht werden.
export const Mindestens = 2

// derBestaetigteStand ist der Stand eines Gegenübers, mit dem gehandelt wird.
const derBestaetigteStand = 'bestaetigt'

async function sets() {
  const antwort = await api().Sets()

  return auspacken(antwort, () => antwort.sets)
}

async function decks() {
  const antwort = await api().Decks('')

  return auspacken(antwort, () => antwort.decks)
}

async function gegenueber() {
  const antwort = await api().Peerkontakte()

  return auspacken(antwort, () => antwort.kontakte).filter(
    (kontakt) => kontakt.stand === derBestaetigteStand,
  )
}

// nurNamen hält von den Treffern der Volltextsuche die, deren **Name** passt
// — und jede Karte einmal: ein Treffer im Regeltext ist kein Treffer beim
// Namen, und dieselbe Karte träfe sonst je Feld.
export function nurNamen(
  treffer: readonly TrefferDTO[],
  eingabe: string,
): TrefferDTO[] {
  const gesucht = eingabe.trim().toLowerCase()
  const gesehen = new Set<string>()

  return treffer.filter((einer) => {
    if (gesehen.has(einer.karteID)) return false
    if (!einer.kartenname.toLowerCase().includes(gesucht)) return false

    gesehen.add(einer.karteID)

    return true
  })
}

// leer ist, was eine Palette ohne Wege lädt: nichts.
const leer = () => Promise.resolve([])

// usePalettendaten lädt, **wenn die Palette Wege hat** — ohne sie gäbe es
// keine Zeile, die etwas täte, und kein Grund, den Bestand zu fragen.
export function usePalettendaten(
  eingabe: string,
  aktiv: boolean,
): Palettendaten {
  const alleSets = useLaden(aktiv ? sets : leer, [aktiv])
  const alleDecks = useLaden(aktiv ? decks : leer, [aktiv])
  const alleGegenueber = useLaden(aktiv ? gegenueber : leer, [aktiv])

  const frage = eingabe.trim()
  const karten = useLaden(
    useCallback(async () => {
      if (!aktiv || frage.length < Mindestens) return []

      const antwort = await api().Suchen(frage, '', '')

      return nurNamen(
        auspacken(antwort, () => antwort.treffer),
        frage,
      )
    }, [aktiv, frage]),
    [aktiv, frage],
  )

  return {
    sets: alleSets.daten ?? [],
    decks: alleDecks.daten ?? [],
    gegenueber: alleGegenueber.daten ?? [],
    karten: karten.daten ?? [],
  }
}
