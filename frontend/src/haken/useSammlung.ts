// **Was die Sammlungsseite tut** — Werkstück U8 von Phase 5.1 (#1419): laden,
// sehen, wählen, schreiben. Wie sie aussieht, steht in `seiten/Sammlung.tsx`
// (§2). Bis U8 stand beides in der Seite; mit Arbeitsplatz, Inspector und
// drei Ansichten trug das die Datei über die Zeilengrenze.
import { useCallback, useMemo, useState } from 'react'

import type { ExemplarDTO } from '../api'
import { api } from '../fassade'
import { istKonflikt, meldungAus } from '../meldung'
import { leereSicht, type Sicht } from '../seiten/sammlungssicht'
import {
  erstesFenster,
  leererNachschlag,
  zeilenAus,
  type Nachschlag,
} from './sammlungszeilen'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { useLaden } from './useLaden'
import { useFensternachschlag } from './useFensternachschlag'
import { useVorgang, type Vorgang } from './useVorgang'

export interface Sammlungsstand {
  laedt: boolean
  // geladen sagt, ob je ein Stand ankam — **ein Lesefehler belegt keine
  // leere Sammlung** (Befund 5 des Zweitblicks über Phase G).
  geladen: boolean
  ladefehler: string | undefined
  fehler: string | undefined

  // konflikt sagt, dass `fehler` ein Konflikt über den Stand ist —
  // UX-08.AK-02 (#1680). Dann bietet die Seite `nachKonflikt` an.
  konflikt: boolean
  nachKonflikt: () => void
  exemplare: readonly ExemplarDTO[]
  sichtbare: readonly ExemplarDTO[]

  // gesamt sind die Zeilen dieser Sicht — `aria-rowcount` nennt sie.
  gesamt: number

  // aufFenster holt nach, was die Liste zeichnen will (#1500).
  aufFenster: (anfang: number, ende: number) => void

  // insgesamt sind die Exemplare ohne Sicht — „N von M". stueck ist die
  // Summe ihrer Mengen, ebenfalls ohne Sicht (die Leiste, AP-9).
  insgesamt: number
  stueck: number
  sicht: Sicht
  setzeSicht: (teil: Partial<Sicht>) => void
  // gewaehlt sind die Haken für die Massenvorgänge — an der Liste, über ids.
  gewaehlt: ReadonlySet<number>
  anhaken: (id: number, an: boolean) => void
  gewaehlteZeilen: { id: number; fassung: number }[]
  // gehalten ist die Wahl für den Inspector — **eine Kennung**, kein Index:
  // ein Filterwechsel trifft danach denselben Datensatz.
  gehalten: number | undefined
  // bekannt ist die Zeile, wenn die Liste sie nicht zeigt (DOM-03).
  waehlen: (id: number, bekannt?: ExemplarDTO) => void
  gehaltenes: ExemplarDTO | undefined
  inSicht: boolean
  vorgang: Vorgang
  anlegen: (exemplar: ExemplarDTO) => Promise<boolean>
  // dublette ist die offene Frage nach der Menge (DOM-03, #1698).
  dublette: Dublette | undefined
  dubletteSchliessen: () => void
  aendern: (exemplar: ExemplarDTO) => Promise<boolean>
  loeschen: (exemplar: ExemplarDTO) => Promise<boolean>
  nachladen: () => void
  // nachSchreiben lädt nach einem eigenen Schreiben nach **und meldet es**
  // (R3) — für die Wege, die außerhalb dieses Hakens schreiben: Einfuhr,
  // Preise, Mengenentscheidung.
  nachSchreiben: () => void
  nachMasse: () => void
  runde: number
}

// Dublette ist eine Aufnahme, die an einer vorhandenen Zeile scheiterte.
export interface Dublette {
  neu: ExemplarDTO
  vorhanden: ExemplarDTO
}

export function exemplarschluessel(id: number): string {
  return `exemplar:${String(id)}`
}

// **`nur` sind die Exemplare, die ein Markenfilter nennt** (UX-12, #1684);
// welche, sagt Go. **`fremdstand` zählt fremde Schreibvorgänge** (#156,
// Befund 3 des Gesamtreviews über 0.6.0) und geht mit der Runde in `zug`.
//
// **`geschrieben` meldet jeden eigenen Schreibvorgang dem Fenster** —
// Nachlauf zum Zweitblick über Phase 6 von 0.9.0, zweiter Teil, R3: sonst
// hob er nur die eigene Runde, und die Deckseite unter `Activity` las ihren
// Abgleich beim Zurückwechseln nicht neu.
export function useSammlung(
  nur?: ReadonlySet<string>,
  fremdstand = 0,
  geschrieben?: () => void,
): Sammlungsstand {
  const texte = useTexte()
  const [runde, setzeRunde] = useState(0)
  const [fehler, setzeFehler] = useState<string>()
  const [konflikt, setzeKonflikt] = useState(false)
  const [sicht, setzeSichtRoh] = useState<Sicht>(leereSicht)
  const [gewaehlt, setzeGewaehlt] = useState<ReadonlySet<number>>(new Set())
  const [dublette, setzeDublette] = useState<Dublette>()
  const [gehalten, setzeGehalten] = useState<number>()

  // **Das gewählte Exemplar wird gemerkt und nicht nur nachgeschlagen** (#1500).
  // Seit die Liste ein **Ausschnitt** ist, kann die gewählte Zeile außerhalb
  // liegen: wer eine wählt und dann sucht, hätte sonst einen leeren Inspector
  // statt der Zeile, die er bearbeiten wollte. Die Abnahme verlangt es —
  // „Auswahl … stimmt über nicht geholte Zeilen hinweg".
  //
  // **Der frische Stand gewinnt**: liegt die Zeile im Fenster, kommt sie von
  // dort; das Gemerkte ist der Rückfall und nicht die Quelle.
  const [gemerkt, setzeGemerkt] = useState<ExemplarDTO>()
  const vorgang = useVorgang()

  // **Nur holen, was gezeichnet wird** — Werkstück
  // #1500.
  //
  // Gefiltert und geordnet wird in Go; über die Naht kommt, was das Fenster
  // gerade braucht. Bei 10.000 Exemplaren waren es vorher **alle** — rund
  // 3 MB JSON, und das Gitter stand nach 209 bis 371 ms.
  //
  // **Die Liste behält ihre volle Länge.** Was noch nicht geholt ist, steht
  // als Platzhalter darin: `aria-rowcount` nennt weiter alle Zeilen, die
  // Tastatur erreicht jede, und das Springen ans Ende holt nach — genau die
  // Zusagen aus U8, die ein fester Ausschnitt gebrochen hätte.
  //
  // **Die Gruppenansicht und ein Markenfilter holen alles**: die eine fasst
  // über die ganze Liste zusammen, der andere nennt Kennungen — ein
  // Ausschnitt gäbe Gruppen, die es nicht gibt, und enthielte sie nicht.
  const holtAlles = sicht.ansicht === 'gruppen' || nur !== undefined

  const zug = runde + fremdstand
  const stand = useLaden(
    async () => {
      const antwort = await api().Sammlungsfenster(
        sicht.suche,
        sicht.ausfuehrung,
        sicht.sortierung,
        0,
        holtAlles ? 0 : erstesFenster,
      )

      return auspacken(antwort, () => antwort)
    },
    [zug, sicht.suche, sicht.ausfuehrung, sicht.sortierung, holtAlles],
    true,
  )

  const gesamt = stand.daten?.gesamt ?? 0
  const insgesamt = stand.daten?.insgesamt ?? 0
  const stueck = stand.daten?.stueck ?? 0

  // nachgeholt sind die Zeilen, die ein späteres Fenster gebracht hat.
  const [nachgeholt, setzeNachgeholt] = useState<Nachschlag>(leererNachschlag)

  // **Ein neuer Stand räumt das Nachgeholte** (§7, Regel 1): es gehört zu
  // einer Sicht, und unter einer anderen zeigte es die falschen Zeilen.
  const schluessel = `${String(zug)}|${sicht.suche}|${sicht.ausfuehrung}|${sicht.sortierung}|${String(holtAlles)}`

  if (nachgeholt.schluessel !== schluessel) {
    setzeNachgeholt({ schluessel, zeilen: new Map() })
  }

  // **Gebaut wird nur, wenn sich etwas geändert hat**: bei zehntausend
  // Exemplaren entstünden sonst zehntausend Platzhalter bei **jedem** Zeichnen
  // — an einer Seite, die bei jedem Tastendruck im Suchfeld zeichnet.
  const erste = stand.daten?.exemplare
  const exemplare = useMemo(
    () => zeilenAus(erste ?? [], gesamt, nachgeholt),
    [erste, gesamt, nachgeholt],
  )
  const sichtbare = useMemo(
    () =>
      nur === undefined
        ? exemplare
        : exemplare.filter((exemplar) => nur.has(String(exemplar.id))),
    [exemplare, nur],
  )

  // aufFenster holt nach, was die Liste zeichnen will und noch nicht hat —
  // eigener Haken, eigene Datei ([useFensternachschlag]).
  const aufFenster = useFensternachschlag({
    sicht,
    holtAlles,
    schluessel,
    nachgeholt,
    setzeNachgeholt,
    setzeFehler,
  })

  const nachladen = useCallback(() => {
    setzeRunde((zahl) => zahl + 1)
  }, [])

  const nachSchreiben = useCallback(() => {
    nachladen()
    geschrieben?.()
  }, [nachladen, geschrieben])

  // **Eine neue Sicht räumt die Haken** (§7, Regel 1) — aber nicht die Wahl
  // des Inspectors: die hängt am Exemplar, nicht an der Liste.
  // **Nur ein Filter räumt die Haken** (#2016) — die übrigen Teile zeigen
  // dieselben Zeilen anders.
  const setzeSicht = useCallback((teil: Partial<Sicht>) => {
    setzeSichtRoh((vorher) => ({ ...vorher, ...teil }))
    if (teil.suche !== undefined || teil.ausfuehrung !== undefined) {
      setzeGewaehlt(new Set())
    }
  }, [])

  const anhaken = useCallback((id: number, an: boolean) => {
    setzeGewaehlt((vorher) => {
      const neu = new Set(vorher)
      if (an) {
        neu.add(id)
      } else {
        neu.delete(id)
      }

      return neu
    })
  }, [])

  // schreibe führt einen Schreibweg aus. **`false` heißt: nichts
  // geschrieben, und das ist kein Fehler** — eine Dublette ist eine Frage
  // (DOM-03), nach ihr wird nicht nachgeladen.
  const schreibe = async (
    kennung: string,
    weg: () => Promise<unknown>,
  ): Promise<boolean> => {
    let geglueckt = false
    await vorgang.starte(kennung, async () => {
      try {
        const ergebnis = await weg()
        setzeFehler(undefined)
        setzeKonflikt(false)

        if (ergebnis === false) {
          return
        }

        nachSchreiben()
        geglueckt = true
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
        setzeKonflikt(istKonflikt(ursache))
      }
    })

    return geglueckt
  }

  const gehaltenes =
    exemplare.find((eines) => eines.id === gehalten) ??
    (gemerkt?.id === gehalten ? gemerkt : undefined)

  const waehlen = (id: number, bekannt?: ExemplarDTO) => {
    setzeGehalten(id)
    setzeGemerkt(bekannt ?? exemplare.find((eines) => eines.id === id))
  }

  return {
    laedt: stand.laedt,
    geladen: stand.daten !== undefined,
    ladefehler: stand.fehler,
    fehler,
    konflikt,

    // **Ein Konflikt wird mit dem heutigen Stand beantwortet** (UX-08.AK-02):
    // die Meldung geht, denn sie gilt einem Stand, den es nicht mehr gibt.
    nachKonflikt: () => {
      setzeFehler(undefined)
      setzeKonflikt(false)
      nachladen()
    },
    exemplare,
    sichtbare,
    gesamt,
    insgesamt,
    stueck,
    aufFenster,
    sicht,
    setzeSicht,
    gewaehlt,
    anhaken,
    gewaehlteZeilen: exemplare
      .filter((eines) => gewaehlt.has(eines.id))
      .map((eines) => ({ id: eines.id, fassung: eines.fassung })),
    gehalten,
    waehlen,
    gehaltenes,
    inSicht: sichtbare.some((eines) => eines.id === gehalten),
    vorgang,
    anlegen: (exemplar) =>
      schreibe('neu', async () => {
        setzeDublette(undefined)
        const antwort = await api().ExemplarAnlegen(exemplar)
        const vorhanden = antwort.vorhanden ?? null

        // **Eine Dublette mit ihrer Zeile ist eine Frage** (DOM-03,
        // S04.NAT-01): erhöhen, zur Zeile, oder nichts tun.
        if (antwort.fehler !== null && vorhanden !== null) {
          setzeDublette({ neu: exemplar, vorhanden })

          return false
        }

        return auspacken(antwort, () => antwort.id)
      }),
    dublette,
    dubletteSchliessen: () => {
      setzeDublette(undefined)
    },
    aendern: (exemplar) =>
      schreibe(exemplarschluessel(exemplar.id), async () => {
        const antwort = await api().ExemplarAendern(exemplar)

        return auspacken({ fehler: antwort }, () => null)
      }),
    loeschen: (exemplar) =>
      schreibe(exemplarschluessel(exemplar.id), async () => {
        const antwort = await api().ExemplarLoeschen(
          exemplar.id,
          exemplar.fassung,
        )

        return auspacken({ fehler: antwort }, () => null)
      }),
    nachladen,
    nachSchreiben,
    nachMasse: () => {
      setzeGewaehlt(new Set())
      nachSchreiben()
    },
    runde,
  }
}
