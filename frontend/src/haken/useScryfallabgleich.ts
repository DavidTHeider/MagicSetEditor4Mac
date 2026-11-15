// **Der Feldabgleich mit Scryfall**
// (#1013).
//
// Ein eigener Haken, weil `Karteneditor.tsx` mit ihm über 400 Zeilen ginge
// (§2) — und weil der Schnitt einer Frage folgt: dort steht, **wie** der
// Editor angeordnet ist, hier, **was** dieser Vorgang tut.
import { useState } from 'react'

import type { Scryfallvorschlag } from '../antworten'
import type { Katalog } from '../sprache/kataloge'
import { api } from '../api'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import type { Vorgang } from './useVorgang'

interface Abgleichswege {
  setID: string
  karteID: string
  vorgang: Vorgang
  texte: Katalog

  // geglueckt läuft nach einer Übernahme — der Editor lädt dann nach.
  geglueckt: () => void
}

// Abgleichsstand ist, was der Kasten zum Zeichnen braucht.
//
// **Exportiert**, damit die Vorschau ihn benennen kann, ohne den Haken zu
// kennen — sie zeigt ihn nur an.
export interface Abgleichsstand {
  vorschlag: Scryfallvorschlag | undefined
  meldung: string | undefined
  holen: (name: string) => void
  // **`stand` bindet den Vorschlag an die Kartenwerte** (#1241) — der
  // Aufrufer hat ihn in der Hand, er zeigt ihn ja gerade.
  uebernehmen: (name: string, felder: string[], stand: string) => void
}

// useScryfallabgleich gibt die beiden Schritte samt ihrem Stand.
//
// # Die vier Regeln aus §7
//
//   - **Regel 1**: ein Kartenwechsel räumt den Vorschlag — er gehört zu der
//     Karte, für die er geholt wurde. Der Aufrufer setzt dafür einen `key`.
//   - **Regel 2**: der Vorgangsschlüssel trägt die Kartenkennung.
//   - **Regel 3**: nach einer geglückten Übernahme wird nachgeladen; nach
//     einer gescheiterten **nicht** — sonst überschriebe ein Lesefehler die
//     Meldung, die sagt, was zu tun ist.
//   - **Regel 4**: ein Vorgang je Karte.
//
// # Warum der Vorschlag nach der Übernahme fällt
//
// **Weil er nicht mehr stimmt.** Die übernommenen Felder sind jetzt gleich;
// ein stehengebliebener Vorschlag böte an, sie noch einmal zu schreiben.
export function useScryfallabgleich({
  setID,
  karteID,
  vorgang,
  texte,
  geglueckt,
}: Abgleichswege): Abgleichsstand {
  const [vorschlag, setzeVorschlag] = useState<Scryfallvorschlag | undefined>(
    undefined,
  )
  const [meldung, setzeMeldung] = useState<string | undefined>(undefined)

  const holen = (name: string) => {
    void vorgang.starte(`${karteID}/abgleich`, async () => {
      try {
        const antwort = await api().ScryfallVorschlag(setID, karteID, name)

        setzeVorschlag(auspacken(antwort, () => antwort.vorschlag ?? undefined))
        setzeMeldung(undefined)
      } catch (ursache: unknown) {
        setzeVorschlag(undefined)
        setzeMeldung(meldungAus(ursache, texte))
      }
    })
  }

  // uebernehmen schreibt die angehakten Felder des **gezeigten** Vorschlags.
  //
  // **Der Stand geht mit** (#1241):
  // Go liest die Karte neu und rechnet den Vorschlag ein zweites Mal, und
  // zwischen Zeigen und Übernehmen kann sie sich geändert haben.
  //
  // **`stand` wird als Parameter genommen und nicht aus dem Zustand gelesen.**
  // Der Aufrufer hat den Vorschlag in der Hand — er zeigt ihn ja gerade; ein
  // `vorschlag?.stand ?? ''` hier wäre ein Zweig, den kein Aufruf betritt (§3),
  // denn der Knopf steht nur da, wo ein Vorschlag steht.
  const uebernehmen = (name: string, felder: string[], stand: string) => {
    void vorgang.starte(`${karteID}/abgleich`, async () => {
      try {
        auspacken(
          // **Der Stand des gezeigten Vorschlags geht mit** (#1241).
          await api().ScryfallUebernehmen(setID, karteID, name, felder, stand),
          () => undefined,
        )

        setzeVorschlag(undefined)
        setzeMeldung(undefined)
        geglueckt()
      } catch (ursache: unknown) {
        setzeMeldung(meldungAus(ursache, texte))
      }
    })
  }

  return { vorschlag, meldung, holen, uebernehmen }
}
