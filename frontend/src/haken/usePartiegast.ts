// **An einem fremden Tisch sitzen** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum der Tisch gemeldet und nicht abgefragt wird
//
// **Weil der Gast nicht weiß, wann etwas geschieht.** Ein anderer Sitz zieht,
// und er erfährt es — über dieselbe Laufzeit, über die die Menüleiste meldet.
// Ein Fenster, das fragte, fragte die meiste Zeit umsonst und wäre trotzdem
// zu spät.
//
// # Und warum der Gast nichts rechnet
//
// **Weil der Wirt rechnet.** Was hier ankommt, ist die **Sicht**, die er für
// diesen Sitz gerechnet hat — fertig projiziert, ohne fremde Hand und ohne
// Bibliotheksreihenfolge. Ein zweiter Spielkern im Fenster hätte beim ersten
// Unterschied eine zweite Wahrheit.
//
// **Was das kostet, gehört dazu**: reißt die Leitung ab, steht kein Tisch mehr
// da — auch nicht der letzte. Ein alter Stand wäre ein Tisch, an dem inzwischen
// gespielt worden sein kann, und das ist schlimmer als keiner.
//
// # Die vier Regeln aus §7
//
//  1. und 2. **Ein `key` aus der Partiekennung** beim Aufrufer räumt beides:
//     wer den Tisch wechselt, baut diesen Haken neu auf.
//  3. **Geschrieben ist geschrieben**: ein Wunsch, der hinausging, kommt nicht
//     zurück — was er bewirkt, meldet der Wirt.
//  4. **Ein Vorgang je Datensatz**: `useVorgang` lässt den zweiten Ruf nicht
//     daneben laufen.
import { useCallback, useEffect, useRef, useState, type RefObject } from 'react'

import {
  api,
  auspacken,
  type PartieAntwort,
  type PartiebeitrittAntwort,
  type Spielereignis,
} from '../api'
import type { Partiemitbring } from '../apipartieweg'
import { horcheNachholend } from '../nachholen'
import { einsetzen } from '../sprache/einsetzen'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useSchlange } from './useSchlange'
import { useVorgang } from './useVorgang'

// Die Ereignisse, über die Go meldet, was am fremden Tisch geschieht.
//
// **Wörtlich dieselben wie in `internal/app/partiebeitritt.go`.**
// `internal/werkzeuge/fensterereignisse_test.go` hält beide Seiten zusammen —
// bis zum Zweitblick über Phase 6.11 stand hier, ein Vertragstest tue es, und
// es gab keinen.
export const ereignisTischbild = 'partie:tischbild'
export const ereignisZug = 'partie:zug'
export const ereignisBeitritt = 'partie:beitritt'

export interface Gastlauf {
  // beitreten setzt sich an den Tisch eines Gegenübers.
  // mit ist, womit sich ein eingeladener Gast setzt (P9c, #2377) — ohne:
  // nichts, und der Wirt stellt das Deck.
  beitreten: (gegenueber: string, partie: string, mit?: Partiemitbring) => void

  // aufstehen verlässt ihn — und gelingt immer.
  aufstehen: () => void

  // ansehen holt den Stand, ohne etwas zu ändern.
  ansehen: () => void

  // wuenschen schickt einen Zug an den Wirt — **der Reihe nach** (P5,
  // #2373): ein Wunsch, während der vorige unterwegs ist, wartet.
  wuenschen: (art: string, nutzlast: Record<string, unknown>) => void

  laeuft: boolean

  // wartend ist, wie viele Wünsche noch nicht beim Wirt angekommen sind.
  wartend: number

  stand: PartiebeitrittAntwort | undefined

  // tisch ist, was der Wirt zuletzt gerechnet hat — die Sicht **und** was sie
  // erlaubt. `undefined`, solange nichts kam.
  //
  // **Dieselbe Gestalt wie bei einer eigenen Partie**, damit dasselbe Bauteil
  // sie zeichnet: zwei Darstellungen derselben Sicht könnten auseinanderlaufen
  // (§16).
  tisch: PartieAntwort | undefined

  // zuege ist, was am Tisch geschehen ist, in der Reihenfolge des Wirts.
  zuege: Spielereignis[]

  // luecke sagt, dass vorn Züge fehlen (Z3, `nachholen.ts`).
  luecke: boolean

  fehler: string | undefined
}

// Gastwunsch ist ein Wunsch samt der Sitzung, in der er gestellt wurde.
interface Gastwunsch {
  sitzung: number
  art: string
  nutzlast: Record<string, unknown>
}

// nichtsMitgebracht setzt sich an einen Sitz, der schon ein Deck hat (P9b,
// #2377).
const nichtsMitgebracht: Partiemitbring = { deckID: '', leiheID: '' }

export function usePartiegast(): Gastlauf {
  const vorgang = useVorgang()
  const texte = useTexte()

  const [stand, setzeStand] = useState<PartiebeitrittAntwort>()
  const [tisch, setzeTisch] = useState<PartieAntwort>()
  const [zuege, setzeZuege] = useState<Spielereignis[]>([])
  const [luecke, setzeLuecke] = useState(false)
  const [fehler, setzeFehler] = useState<string>()

  // **Die Horcher hängen am Fenster und nicht am Beitritt.** Einer, der erst
  // beim Klick angemeldet würde, verpasste das erste Tischbild —
  // `EventsOn` meldet nicht nach. **Und sie holen nach** (M9 b): verborgen
  // baut `<Activity>` sie ab, und was in der Zeit kam, bekommen sie beim
  // Einblenden (`nachholen.ts`).
  //
  // **Und sie nehmen nur, was dem eigenen Platz gehört** (Z1 aus dem
  // Zweitblick über Phase 6.11): Go nennt die Partie bei jeder Meldung, und
  // das Fenster des Wirts bekommt dieselben Ereignisse für seine eigene
  // (`MitFenster`, #1627). Ohne Platz nimmt der Gast nichts; ohne genannte
  // Partie die des ersten Bildes.
  //
  // **Und nie die eines Tisches, den er gerade verlassen hat** (N2 aus dem
  // Nachlauf über 6.11): eine verspätete Meldung von dort käme sonst vor
  // dem ersten Bild des neuen an.
  const platz = useRef<{ partie: string; gegenueber: string }>(undefined)
  const verlassen = useRef(new Map<string, string>())
  useEffect(() => horcheAufTischbild(platz, verlassen, setzeTisch), [])
  useEffect(() => horcheAufZug(platz, verlassen, setzeZuege, setzeLuecke), [])

  // verlasse merkt die Partie, an der der Gast gerade saß.
  const verlasse = useCallback(() => {
    if (platz.current?.partie) {
      verlassen.current.set(platz.current.partie, platz.current.gegenueber)
    }
  }, [])

  // sitzung zählt Beitritt und Aufstehen mit — **eine Antwort gehört zu der
  // Sitzung, aus der sie stammt** (§7, Regel 2): ein Wunsch, dessen Antwort
  // nach dem Aufstehen kommt, setzte den Gast sonst an den alten Tisch.
  const sitzung = useRef(0)

  // hole wartet auf eine Antwort des Wirts und nimmt sie, solange die
  // Sitzung dieselbe ist.
  const hole = useCallback(
    async (arbeit: () => Promise<PartiebeitrittAntwort>, meine: number) => {
      setzeFehler(undefined)

      try {
        const antwort = await arbeit()

        auspacken(antwort, () => undefined)
        if (meine !== sitzung.current) {
          return
        }

        // **Ein Stand ohne Partie behält die gelernte** — der Wirt nennt sie
        // im ersten Bild, und das kann vor dieser Antwort kommen.
        // **Ein verlorener Platz nimmt seinen Tisch mit** (P9c, #2377, §7
        // Regel 1): der Wirt grüßt mit einem Bild, bevor er ablehnt, und
        // ein altes Bild wäre eines, an dem inzwischen gespielt sein kann.
        //
        // **Und er beendet die Sitzung** (Nachlauf über Phase 7, R6): der
        // Stand nach `partie:beitritt` kommt am Vorgang vorbei; eine ältere
        // Antwort, die danach noch „läuft“ sagt, setzte den Gast sonst wieder
        // an den Tisch.
        if (!antwort.laeuft) {
          sitzung.current += 1
          platz.current = undefined
          setzeTisch(undefined)
          setzeZuege([])
          setzeLuecke(false)
        } else if (antwort.partie !== '') {
          verlassen.current.delete(antwort.partie)
          platz.current = {
            partie: antwort.partie,
            gegenueber: antwort.gegenueber,
          }
        } else {
          platz.current ??= { partie: '', gegenueber: antwort.gegenueber }
        }
        setzeStand(antwort)
      } catch (ursache: unknown) {
        if (meine === sitzung.current) {
          setzeFehler(meldungAus(ursache, texte))
        }
      }
    },
    [texte],
  )

  // **Meldet Go, dass der Platz verloren ist** (P9c, #2377), fragt das
  // Fenster selbst nach — sonst stünde die Absage des Wirts erst nach
  // „Nachsehen“ da. **Am Vorgang vorbei**: es ist ein Lesen, und ein
  // Beitritt, der gerade noch antwortet, hielte sie sonst zurück.
  useEffect(
    () =>
      horcheNachholend(ereignisBeitritt, gaststelle, () => {
        void hole(() => api().PartieBeitrittStand(), sitzung.current)
      }),
    [hole],
  )

  const fahre = useCallback(
    (arbeit: () => Promise<PartiebeitrittAntwort>) => {
      void vorgang.starte('partie-gast', () => hole(arbeit, sitzung.current))
    },
    [vorgang, hole],
  )

  const beitreten = useCallback(
    (gegenueber: string, partie: string, mit?: Partiemitbring) => {
      // **Ein neuer Tisch fängt leer an** (§7, Regel 1): was der vorige
      // zeigte, gehört ihm und nicht diesem.
      setzeTisch(undefined)
      setzeZuege([])
      setzeLuecke(false)
      verlasse()
      verlassen.current.delete(partie)

      // **Beim selben Wirt gilt seine Partie wieder** (P9c, #2377): der
      // Schutz gilt späten Meldungen eines **anderen** Tisches; wer sich
      // dort wieder hinsetzt, wo er aufstand, sähe sonst nie ein Bild.
      for (const [verlassene, bei] of verlassen.current) {
        if (bei === gegenueber) verlassen.current.delete(verlassene)
      }
      platz.current = { partie, gegenueber }
      sitzung.current += 1

      fahre(() =>
        api().PartieBeitreten(gegenueber, partie, mit ?? nichtsMitgebracht),
      )
    },
    [fahre, verlasse],
  )

  const aufstehen = useCallback(() => {
    setzeTisch(undefined)
    setzeZuege([])
    setzeLuecke(false)
    verlasse()
    platz.current = undefined
    sitzung.current += 1

    fahre(() => api().PartieBeitrittBeenden())
  }, [fahre, verlasse])

  const ansehen = useCallback(() => {
    fahre(() => api().PartieBeitrittStand())
  }, [fahre])

  // **Wünsche der Reihe nach** (P5, #2373): ein zweiter, während der erste
  // zum Wirt unterwegs ist, wartet, statt an `useVorgang` abzuprallen — und
  // ein Wunsch hält das Aufstehen nicht auf. Gedeutet wird beim Wirt; der
  // Gast hält keine Regelkopie. **Was zu einer verlassenen Sitzung gehört,
  // geht nicht mehr hinaus** (§7, Regel 1).
  const { stelle, wartend } = useSchlange<Gastwunsch>(async (wunsch) => {
    if (wunsch.sitzung === sitzung.current) {
      await hole(
        () =>
          api().PartieWuenschen(
            neueWunschkennung(),
            wunsch.art,
            wunsch.nutzlast,
          ),
        wunsch.sitzung,
      )
    }
  })

  const wuenschen = useCallback(
    (art: string, nutzlast: Record<string, unknown>) => {
      stelle({ sitzung: sitzung.current, art, nutzlast })
    },
    [stelle],
  )

  // **Ein abgewiesener Wunsch steht am Tisch** (#2478): der Wirt schickt
  // sein Bild mit dem Grund, und das nächste Bild ohne nimmt ihn wieder weg.
  const absage = tisch?.absage

  return {
    beitreten,
    aufstehen,
    ansehen,
    wuenschen,
    laeuft: vorgang.laeuft('partie-gast'),
    wartend,
    stand,
    tisch,
    zuege,
    luecke,
    fehler:
      fehler ??
      (absage === undefined
        ? undefined
        : einsetzen(texte.seiten.spielen.gastWunschAbgewiesen, {
            grund: absage.grund,
          })),
  }
}

// gaststelle ist der Name, unter dem der Gast nachholt.
const gaststelle = 'partiegast'

type Platz = RefObject<{ partie: string; gegenueber: string } | undefined>
type Verlassen = RefObject<Map<string, string>>

// gehoertDemPlatz sagt, ob eine Meldung der Partie gilt, an der der Gast
// sitzt — und lernt sie aus dem ersten **Bild** (`lernt`), wenn er beim
// Beitreten keine genannt hat. Ein Zug lehrt nichts: der Wirt grüßt mit
// einem Bild.
function gehoertDemPlatz(
  platz: Platz,
  verlassen: Verlassen,
  partie: unknown,
  lernt: boolean,
): boolean {
  const jetzt = platz.current
  if (jetzt === undefined) return false

  if (jetzt.partie === '') {
    if (!lernt || verlassen.current.has(String(partie))) return false

    jetzt.partie = String(partie)
  }

  return jetzt.partie === partie
}

// horcheAufTischbild meldet jedes neue Bild.
function horcheAufTischbild(
  platz: Platz,
  verlassen: Verlassen,
  setze: (tisch: PartieAntwort) => void,
): () => void {
  return horcheNachholend(ereignisTischbild, gaststelle, (daten) => {
    const [partie, bild] = daten

    if (gehoertDemPlatz(platz, verlassen, partie, true)) {
      setze(bild as PartieAntwort)
    }
  })
}

// horcheAufZug hängt jeden Zug an — und sagt, wenn vorn welche fehlen.
function horcheAufZug(
  platz: Platz,
  verlassen: Verlassen,
  setze: (fort: (bisher: Spielereignis[]) => Spielereignis[]) => void,
  luecke: (ja: boolean) => void,
): () => void {
  return horcheNachholend(
    ereignisZug,
    gaststelle,
    (daten) => {
      const [partie, zug] = daten

      if (gehoertDemPlatz(platz, verlassen, partie, false)) {
        setze((bisher) => [...bisher, zug as Spielereignis])
      }
    },
    () => {
      luecke(platz.current !== undefined)
    },
  )
}

// neueWunschkennung gibt eine Kennung für **einen** Wunsch.
//
// **`crypto.randomUUID` und kein Zähler** — derselbe Grund wie bei
// `usePartie`: ein Zähler finge nach einem Neustart wieder bei eins an, und
// der Wirt hielte den ersten Wunsch für eine Wiederholung.
function neueWunschkennung(): string {
  return crypto.randomUUID()
}
