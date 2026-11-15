// **Der Scryfall-Vorschlag der Kartenlasche** — der Treffer, den die Suche
// gefunden hat (#1167).
//
// # Warum er einen eigenen Haken hat
//
// **Weil ihn zwei Seiten der Lasche brauchen, seit die Handlungen getrennt
// sind** (#1240):
// der Anlegekasten liest seinen **Namen**, der Verweiskasten der
// aufgeschlagenen Karte seine **Kennung**. Ein Zustand in einem von beiden
// wäre eine Kante, die es nicht gibt.

import { useState } from 'react'

import type { SuchtrefferDTO } from '../apiscryfall'

export interface Scryfallvorschlag {
  // treffer ist der Vorschlag für die **aufgeschlagene** Karte — oder keiner.
  treffer: SuchtrefferDTO | undefined

  waehlen: (treffer: SuchtrefferDTO | undefined) => void
}

// useScryfallvorschlag hält den Treffer und räumt ihn beim Kartenwechsel.
export function useScryfallvorschlag(
  karteID: string | undefined,
): Scryfallvorschlag {
  // **Der Treffer, den die Suche gefunden hat** (#1167).
  //
  // Er steht hier und nicht in `Scryfallsuche`, weil ihn **drei** Bauteile
  // brauchen: die Suche setzt ihn, der Anlegekasten liest seinen Namen, und
  // der Verweis liest seine **Kennung**. Ein Zustand in einem von ihnen wäre
  // eine Kante, die es nicht gibt.
  //
  // **Der ganze Treffer und nicht nur der Name**: mit dem Namen allein wäre
  // die Kennung des Drucks weggeworfen, bevor irgendjemand sie brauchen
  // könnte.
  //
  // **Und er merkt sich, für welche Karte er gewählt wurde** — Befund 5 des
  // Zweitblicks über Phase 3.6. Ohne diesen Merker bot der Verweiskasten nach
  // einem Kartenwechsel weiter den Treffer an, den jemand für die **vorige**
  // Karte angeklickt hatte: §7, Regel 1 — ein Wechsel des Bezugsobjekts räumt
  // den offenen Kontext.
  //
  // **Die Trefferliste bleibt stehen.** Geräumt wird die *Wahl*, nicht die
  // Suche; ein Klick nimmt sie wieder auf, und die Liste steht noch da.
  const [gewaehlterTreffer, setzeGewaehltenTreffer] = useState<{
    fuer: string | undefined
    treffer: SuchtrefferDTO
  }>()

  // **Und beim Wechsel wird geräumt, nicht gefiltert.** Ein Filter ließe den
  // Treffer im Zustand und blendete ihn nur aus: Karte A → B → **A** brächte
  // denselben Vorschlag wieder — die Wahl wäre nie geräumt, nur unsichtbar.
  // §7, Regel 1 meint räumen.
  if (gewaehlterTreffer !== undefined && gewaehlterTreffer.fuer !== karteID) {
    setzeGewaehltenTreffer(undefined)
  }

  return {
    treffer: gewaehlterTreffer?.treffer,
    waehlen: (treffer) => {
      setzeGewaehltenTreffer(
        treffer === undefined ? undefined : { fuer: karteID, treffer },
      )
    },
  }
}
