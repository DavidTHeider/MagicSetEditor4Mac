// **Womit die Kartenlasche arbeitet** — Zwischenablage und Listengestalt.
//
// Werkstück 3 von Phase 3.7
// (#1168).
//
// **Eine eigene Datei**, weil `Karten.tsx` mit ihr über 400 Zeilen ginge (§2).
// Der Schnitt folgt einer Frage: dort steht, **wie** die Lasche angeordnet
// ist, hier, **womit** sie arbeitet — zwei Griffe, die beide am Set hängen und
// beide denselben Vorgang und dieselbe Meldung teilen.
import { useKartenablage, type Kartenablagegriff } from './useKartenablage'
import { useListengestalt } from './useListengestalt'
import type { Listengestalt } from '../seiten/listengestalt'
import type { Vorgang } from './useVorgang'

interface Werkzeuglage {
  setID: string

  // angehakt und aufgeschlagen sagen, welche Karten die Ablage meint.
  angehakt: readonly string[]
  aufgeschlagen: string | undefined

  // verlaufsstand steigt bei jeder Änderung — die **Werte** der Listenspalten
  // hängen daran.
  verlaufsstand: number

  vorgang: Vorgang

  // nurAktuell verwirft, was nach einem Setwechsel eintrifft (§7, Regel 2).
  nurAktuell: <T extends unknown[]>(
    setID: string,
    tun: (...werte: T) => void,
  ) => (...werte: T) => void

  setzeAnlegefehler: (meldung: string | undefined) => void
  nachladen: () => void
  bestandGeaendert: () => void
}

export interface Kartenwerkzeug {
  ablage: Kartenablagegriff
  listengestalt: Listengestalt
}

export function useKartenwerkzeug({
  setID,
  angehakt,
  aufgeschlagen,
  verlaufsstand,
  vorgang,
  nurAktuell,
  setzeAnlegefehler,
  nachladen,
  bestandGeaendert,
}: Werkzeuglage): Kartenwerkzeug {
  // **Der Griff auf die Zwischenablage** — **einer**, geteilt von den Knöpfen
  // der Handlungszeile und den Kürzeln in der Kartenspalte.
  const ablage = useKartenablage({
    setID,
    angehakt,
    aufgeschlagen,
    vorgang,
    aufFehler: nurAktuell(setID, setzeAnlegefehler),
    geaendert: nurAktuell(setID, () => {
      nachladen()
      bestandGeaendert()
    }),
  })

  // **Die Gestalt der Liste** — Spalten, Ordnung und der Weg, sie zu ändern.
  // Sie hängt am **Spiel** und nicht am Set; nachgeladen wird sie trotzdem
  // beim Setwechsel, denn welches Spiel gilt, sagt das Set.
  const listengestalt = useListengestalt({
    setID,
    stand: verlaufsstand,
    vorgang,
    aufFehler: nurAktuell(setID, setzeAnlegefehler),
  })

  return { ablage, listengestalt }
}
