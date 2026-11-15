// **Die beiden Wege, die Teile hinzufügen** — eine Grundform und ein
// nachgezeichnetes Bild.
//
// # Warum eine eigene Datei
//
// `Symbol.tsx` stand an der Zeilengrenze aus §2, als der Wächter über die
// gleichzeitigen Handlungen dazukam (Befund 2 des Zweitblicks über Phase
// 4.11). Der Schnitt folgt der Frage: *was der Editor zeigt*, steht dort;
// *woher neue Teile kommen*, hier.

import { useCallback } from 'react'

import type { SymbolteilDTO } from '../antworten'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import type { Handgriff } from './useZeichenverlauf'
import type { Zeichenschutz } from './useZeichenschutz'

interface ZugaengeEigenschaften {
  gezeigte: SymbolteilDTO[]
  eintragen: (neue: SymbolteilDTO[], handgriff: Handgriff) => void
  schutz: Zeichenschutz
  waehlen: (stelle: number, hinzu: boolean) => void
  klagen: (ursache: unknown) => void
}

export interface Symbolzugaenge {
  hinzufuegen: (art: string, ecken: number, eng: boolean) => Promise<void>
  nachzeichnen: () => Promise<void>
}

export function useSymbolzugaenge({
  gezeigte,
  eintragen,
  schutz,
  waehlen,
  klagen,
}: ZugaengeEigenschaften): Symbolzugaenge {
  // hinzufuegen holt eine Grundform von der Go-Seite und hängt sie an.
  //
  // **Sie wird nicht hier gerechnet** — der Kreis benutzt eine Konstante, die
  // MSE selbst „magic" nennt, und eine zweite Rechnung liefe mit jener
  // auseinander (§2).
  const hinzufuegen = useCallback(
    (art: string, ecken: number, eng: boolean) =>
      schutz.starte(async (nochAktuell) => {
        try {
          const antwort = await api().Grundform({ art, ecken, eng })
          const neue = auspacken(antwort, () => antwort.teile)

          if (!nochAktuell()) {
            return
          }

          eintragen([...gezeigte, ...neue], {
            marke: `grundform:${art}:${String(gezeigte.length)}`,
          })
          waehlen(gezeigte.length, false)
        } catch (ursache: unknown) {
          klagen(ursache)
        }
      }),
    [eintragen, gezeigte, klagen, schutz, waehlen],
  )

  // nachzeichnen holt eine Bilddatei und hängt ihre Formen an.
  //
  // **Ein Handgriff wie jeder andere**: er landet im Verlauf, und ⌘Z nimmt ihn
  // zurück. Ohne das wäre ein versehentlich gewähltes Bild eine Zeichnung, die
  // man von Hand wieder auseinandernehmen müsste.
  const nachzeichnen = useCallback(
    () =>
      schutz.starte(async (nochAktuell) => {
        try {
          const antwort = await api().SymbolNachzeichnen()
          const neue = auspacken(antwort, () => antwort.teile)

          if (!nochAktuell()) {
            return
          }

          // **Ein Abbruch im Dateidialog gibt eine leere Liste ohne
          // Fehler.** Ihn als Handgriff einzutragen hieße, einen leeren
          // Schritt in den Verlauf zu legen — und ⌘Z täte dann nichts
          // Sichtbares.
          if (neue.length === 0) {
            return
          }

          eintragen([...gezeigte, ...neue], {
            marke: `nachzeichnen:${String(gezeigte.length)}`,
          })
          waehlen(gezeigte.length, false)
        } catch (ursache: unknown) {
          klagen(ursache)
        }
      }),
    [eintragen, gezeigte, klagen, schutz, waehlen],
  )

  return { hinzufuegen, nachzeichnen }
}
