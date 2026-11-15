// **Der Zustand der Deckseite** — Werkstück 7 von Phase 4.6
// (#1293).
//
// # Warum ein Haken und nicht die Seite
//
// §2: *was* eine Seite tut, steht im Haken; *wie sie aussieht*, in der
// Komponente. Und §7 lebt hier: ein Wechsel des Decks räumt das offene
// Formular (Regel 1), eine Antwort gehört zu dem Stand, aus dem sie stammt
// (Regel 2), ein geglücktes Schreiben wird nicht zurückgenommen (Regel 3),
// und ein zweiter Ruf während des ersten geht gegen einen Stand, den der
// Bestand noch nicht kennt (Regel 4).
import { useCallback, useRef, useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { istKonflikt, meldungAus } from '../meldung'
import { useFremdNachlesen } from './useFremdNachlesen'
import { useLaden } from './useLaden'
import { useLauf } from './useLauf'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from './useVorgang'

import type { DeckDTO, FormatDTO } from '../apidecks'

export interface Deckstand {
  decks: DeckDTO[]
  formate: FormatDTO[]

  // offen ist das aufgeschlagene Deck — samt seinen Plätzen.
  offen: DeckDTO | null

  // geladen sagt, dass die Liste einmal angekommen ist. **Vorher ist sie
  // nicht leer, sondern unbekannt** (UX-08.NAT-01) — ein Nachladen nach einem
  // Schreiben behält die vorige und ändert daran nichts.
  geladen: boolean
  ladefehler: string | undefined
  fehler: string | undefined

  // konflikt sagt, dass `fehler` ein Konflikt über den Stand ist —
  // UX-08.AK-02 (#1680). Dann bietet die Seite `nachKonflikt` an.
  konflikt: boolean

  // nachKonflikt holt den heutigen Stand: die Liste und das offene Deck.
  nachKonflikt: () => void

  vorgang: ReturnType<typeof useVorgang>

  waehle: (deckID: string) => Promise<void>

  // liesNach liest ein Deck neu — **nur, wenn es noch das offene ist**. Nach
  // einem Schreiben, während dessen jemand ein anderes aufschlug, schlüge
  // `waehle` das vorige wieder auf (§7, Regel 1 und 2).
  liesNach: (deckID: string) => Promise<void>

  // wahl gibt die Nummer der heutigen Wahl. Sie steigt mit jedem Wechsel
  // des aufgeschlagenen Decks — **auch zurück**: A → B → A ist eine neue
  // Wahl von A und nicht die alte (Rest R3 des Nachlaufs über Phase 6 von
  // 0.9.0).
  wahl: () => number

  // wechsleVon schlägt `nach` auf — **nur, wenn seit `wahl` niemand ein
  // anderes Deck aufgeschlagen hat**. Für einen Lauf, der nach seinem Ende
  // ein anderes Deck aufschlägt (die Kopie): wer inzwischen gewechselt hat,
  // behält seine Wahl (§7, Regel 2).
  wechsleVon: (wahl: number, nach: string) => Promise<void>

  // vergiss räumt ein Deck aus dem offenen Kontext — nach dem Löschen.
  vergiss: (deckID: string) => void
  neuLaden: () => void
  // amFeld nimmt einen Fehler, der zu einem Feld gehört, der Seite ab — dann
  // steht er dort und nicht über ihr (ENG-05.AK-03, §6).
  fuehre: (
    marke: string,
    arbeit: () => Promise<void>,
    amFeld?: (ursache: unknown) => boolean,
  ) => Promise<void>
  melde: (ursache: unknown) => void
}

// useDecks hält Liste, Formate und das aufgeschlagene Deck.
// **`fremdstand` zählt, was ein anderer Prozess oder die Leitung geschrieben hat**
// (#156) — Befund 3 des Gesamtreviews über 0.6.0: ohne ihn lud die Seite erst
// nach einem Wechsel nach. **Behalten**, damit die Liste dabei stehen bleibt.
export function useDecks(fremdstand = 0): Deckstand {
  const texte = useTexte()

  const [offen, setzeOffen] = useState<DeckDTO | null>(null)
  const [fehler, setzeFehler] = useState<string>()
  const [konflikt, setzeKonflikt] = useState(false)

  const vorgang = useVorgang()

  // **Ein Zähler über den Deckwechseln** (§7, Regel 2). Eine langsame Antwort
  // für Deck A kann sonst nach der Antwort für Deck B wieder A einsetzen —
  // und das Fenster zeigt einen Stand, den niemand aufgeschlagen hat.
  //
  // **Angestrichen von Befund 5 des Zweitblicks über Phase 4.6.**
  const lauf = useLauf()

  // **Welches Deck gerade gilt** — das aufgeschlagene **oder** das, dessen
  // Lesen läuft.
  //
  // **Eine Referenz und kein Zustand**: sie wird gelesen, nicht gezeichnet,
  // und ein Zustand ergäbe hier einen zweiten Durchlauf ohne Wirkung.
  //
  // **Warum sie nicht `offen` sein kann**: `waehle` setzt `offen` beim Klick
  // auf `null` und füllt es erst mit der Antwort. Zwischen beidem gibt es
  // kein aufgeschlagenes Deck — wohl aber ein Deck, um das es geht. Genau in
  // dieser Lücke stand der zweite Teil von Befund 4 des Nachlaufs: wer dort
  // löscht, findet über `offen` nichts zum Abräumen, und die Antwort von
  // vorhin schlägt das gelöschte Deck auf.
  const bezug = useRef<string | null>(null)

  // **Die Nummer der Wahl** — sie steigt, wenn `bezug` ein anderes Deck
  // nennt. Nicht die Laufnummer: die steigt auch beim leisen Nachlesen, und
  // das ist kein Wechsel.
  const wahlnummer = useRef(0)

  const stand = useLaden(
    async () => {
      // **Zwei Rufe in einem Zug — und das ist kein gemeinsamer Stand.**
      //
      // Berichtigt am 2026-09-07 (Zweitblick über die Phasen 4.0 bis 4.13):
      // `Promise.all` startet beide zugleich, aber jeder liest für sich. Wer
      // dazwischen ein Format anlegt oder löscht, sieht eine Formatliste, die
      // nicht zu den Deckverweisen passt. Der nächste Ladelauf richtet es.
      const [decks, formate] = await Promise.all([
        api().Decks(''),
        api().Formate(''),
      ])

      return {
        decks: auspacken(decks, () => decks.decks),
        formate: auspacken(formate, () => formate.formate),
      }
    },
    [fremdstand],
    true,
  )

  const neuLaden = stand.nachladen

  const melde = useCallback(
    (ursache: unknown) => {
      setzeFehler(meldungAus(ursache, texte))
      setzeKonflikt(istKonflikt(ursache))
    },
    [texte],
  )

  // lesen holt ein Deck und trägt es ein, solange `nummer` gilt (§7,
  // Regel 2) — für den Klick wie für das Nachlesen.
  const lesen = useCallback(
    async (deckID: string, nummer: number, vermerken: boolean) => {
      try {
        const antwort = await api().Deck(deckID)

        if (!lauf.istAktuell(nummer)) {
          return
        }

        setzeOffen(auspacken(antwort, () => antwort.deck))

        // **Vermerkt für die Übersicht** (#1422) — Bestenfalls-Arbeit wie
        // `SetBenutzt`: ein Fehler hier ändert nichts am geöffneten Deck.
        if (vermerken) {
          const vermerk = api().DeckBenutzt(deckID)
          void vermerk.catch(() => undefined)
        }
      } catch (ursache: unknown) {
        // **Auch der Fehler** — eine Meldung zu einem Deck, das niemand mehr
        // offen hat, sieht aus wie ein Fehler des aktuellen.
        if (lauf.istAktuell(nummer)) {
          melde(ursache)
        }
      }
    },
    [lauf, melde],
  )

  // **Ein Wechsel des Decks räumt den offenen Kontext mit** (§7, Regel 1):
  // den Fehler **und** das aufgeschlagene Deck. Ohne das Zweite stünde das
  // vorige Deck stehen, während das neue lädt — und ein Klick daneben ginge
  // gegen das falsche.
  //
  // **Und die Antwort gehört zu dem Stand, aus dem sie stammt** (Regel 2).
  const waehle = useCallback(
    async (deckID: string) => {
      const nummer = lauf.naechster()

      // **Nur ein *anderes* Deck räumt den offenen Kontext** (§7, Regel 1) —
      // Werkstück U10 von Phase 5.1
      // (#1421).
      //
      // Hier stand `setzeOffen(null)` **immer**, und das ist etwas anderes:
      // nach jedem Legen, Nehmen und jeder Mengenänderung liest dieser Haken
      // dasselbe Deck neu, und dazwischen war es für einen Augenblick zu.
      // Am Arbeitsplatz sieht man, was das kostet — der Inspector fällt auf
      // seine erste Lasche zurück, und wer die Prüfung offen hatte, steht
      // nach jedem `+` wieder bei der Auswertung.
      //
      // **Ein Wechsel räumt weiter** — das ist die Regel, und sie meint das
      // Bezugsobjekt, nicht den Ruf.
      const wechsel = bezug.current !== deckID

      bezug.current = deckID

      setzeFehler(undefined)

      if (wechsel) {
        wahlnummer.current += 1
        setzeOffen(null)
      }

      await lesen(deckID, nummer, true)
    },
    [lauf, lesen],
  )

  // **Und nach einem fremden Schreiben noch einmal** — Befund 2 des
  // Nachlaufs zum Gesamtreview über 0.6.0 (#1660). **Leise**: kein Wechsel,
  // keine weggeräumte Meldung (§7, Regel 3) und **kein Vermerk** — ein
  // Nachlesen ist kein Benutzen, und die Übersicht sortiert danach.
  const nachlesen = useCallback(() => {
    const deckID = bezug.current

    if (deckID !== null) {
      void lesen(deckID, lauf.naechster(), false)
    }
  }, [lauf, lesen])

  useFremdNachlesen(fremdstand, nachlesen)

  const liesNach = useCallback(
    async (deckID: string) => {
      if (bezug.current === deckID) {
        await waehle(deckID)
      }
    },
    [waehle],
  )

  const wahl = useCallback(() => wahlnummer.current, [])

  const wechsleVon = useCallback(
    async (gewaehlt: number, nach: string) => {
      if (wahlnummer.current === gewaehlt) {
        await waehle(nach)
      }
    },
    [waehle],
  )

  // **Löschen ist der stärkste Wechsel des Bezugsobjekts** (§7, Regel 1) —
  // und war bis zum 2026-09-01 der einzige, der den offenen Kontext stehen
  // ließ. Das Fenster zeigte und bediente danach einen Datensatz, den der
  // Bestand nicht mehr kennt; der nächste Schreibversuch scheiterte erst am
  // Backend (Befund 4 des Nachlaufs über Phase 4.6).
  //
  // **Und der Laufzähler geht mit** (Regel 2): ein vor dem Löschen
  // begonnenes Lesen trüge sonst weiterhin die aktuelle Nummer und schlüge
  // das gelöschte Deck **erneut** auf.
  //
  // **Nur das betroffene Deck.** Wer B löscht, schließt nicht A — und eine
  // Nummer, die unbedingt stiege, verwürfe die Antwort für A gleich mit.
  const vergiss = useCallback(
    (deckID: string) => {
      // **Der Laufzähler steigt außerhalb des Setzers.** Ein Seiteneffekt in
      // einem Zustandssetzer liefe unter React im Prüfmodus zweimal.
      if (bezug.current === deckID) {
        lauf.naechster()
        wahlnummer.current += 1

        bezug.current = null
      }

      setzeOffen((vorher) => (vorher?.id === deckID ? null : vorher))
    },
    [lauf],
  )

  // **Ein Vorgang je Bezugsobjekt** (§7, Regel 4): ein zweiter Ruf, während
  // der erste läuft, ginge gegen einen Stand, den der Bestand noch nicht
  // kennt.
  //
  // **Die Marke ist das Bezugsobjekt und nicht die Handlung** — angestrichen
  // von Befund 5 des Zweitblicks über Phase 4.6. Vorher hießen sie
  // `platz-legen:7`, `deck-format:7` und `deck-umbenennen:7`, und drei
  // verschiedene Schreibarten auf **dasselbe** Deck sperrten einander nicht.
  //
  // **Und nach dem Schreiben wird nachgeladen — aber nur nach einem
  // geglückten** (§7, Regel 3): sonst überschriebe ein Lesefehler die
  // Meldung, die sagt, was zu tun ist.
  // **Und der Fehler gehört zu dem Stand, aus dem er stammt** (§7, Regel 2).
  //
  // Ein Schreibvorgang für Deck A kann scheitern, während B aufgeschlagen
  // ist: `useVorgang` sperrt nur denselben Schlüssel, und ein Deckwechsel
  // dazwischen ist erlaubt. Ohne die Laufnummer stünde die Meldung über B.
  //
  // **Angestrichen am Turnierweg** (Befund 9 des Zweitblicks über Phase 4.7)
  // und hier nachgezogen: es ist dieselbe Gestalt, und behoben wird die
  // Klasse (§10).
  const fuehre = useCallback(
    async (
      marke: string,
      arbeit: () => Promise<void>,
      amFeld?: (ursache: unknown) => boolean,
    ) => {
      const nummer = lauf.nummer()

      await vorgang.starte(marke, async () => {
        try {
          await arbeit()

          // **Die Liste gehört keinem Deck** — sie wird nach jedem
          // geglückten Schreiben nachgeladen, auch wenn inzwischen ein
          // anderes offen ist (Regel 3). Hier stand sie hinter der
          // Laufnummer, und eine Kopie, während der jemand wechselte, stand
          // geschrieben im Bestand und nirgends in der Liste (Nachlauf über
          // Phase 6 von 0.9.0, R3).
          neuLaden()

          if (lauf.istAktuell(nummer)) {
            setzeFehler(undefined)
          }
        } catch (ursache: unknown) {
          if (!lauf.istAktuell(nummer)) {
            return
          }

          // **Am Feld oder über der Seite, nie beides** — und ein Befund am
          // Feld ist das neueste Wort, also geht ein älterer Hinweis weg.
          if (amFeld?.(ursache) === true) {
            setzeFehler(undefined)
          } else {
            melde(ursache)
          }
        }
      })
    },
    [lauf, melde, neuLaden, vorgang],
  )

  // **Ein Konflikt wird mit dem heutigen Stand beantwortet** (UX-08.AK-02,
  // #1680): die Liste und das offene Deck werden neu gelesen, und die
  // Meldung geht — sie gilt einem Stand, den es nicht mehr gibt.
  const nachKonflikt = useCallback(() => {
    setzeFehler(undefined)
    setzeKonflikt(false)
    neuLaden()

    if (bezug.current !== null) {
      void waehle(bezug.current)
    }
  }, [neuLaden, waehle])

  return {
    decks: stand.daten?.decks ?? [],
    formate: stand.daten?.formate ?? [],
    offen,
    geladen: stand.daten !== undefined,
    ladefehler: stand.fehler,
    fehler,
    konflikt,
    nachKonflikt,
    vorgang,
    waehle,
    liesNach,
    wahl,
    wechsleVon,
    vergiss,
    neuLaden,
    fuehre,
    melde,
  }
}
