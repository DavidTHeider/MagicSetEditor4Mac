// **Der Stand der Turnierseite** — Werkstück 5 von Phase 4.7
// (#1294),
// nach Vermerk 0062.
//
// # Die vier Regeln aus §7, und wo sie hier stehen
//
//  1. **Ein Wechsel räumt den offenen Kontext** — `waehle` leert das
//     aufgeschlagene Turnier, wenn es ein anderes ist, `vergiss` räumt es
//     nach dem Löschen. Dasselbe nachzuladen ist kein Wechsel.
//  2. **Eine Antwort gehört zu dem Stand, aus dem sie stammt** — `useLauf`
//     zählt die Wechsel mit.
//  3. **Geschrieben ist geschrieben** — nachgeladen wird nur nach einem
//     geglückten Schreiben.
//  4. **Ein Vorgang je Datensatz** — `useVorgang`, und die Marke ist das
//     **Bezugsobjekt**, nicht die Handlung.
//
// **Dieselbe Gestalt wie `useDecks`**, und das ist Absicht: zwei Seiten mit
// derselben Aufgabe sollen nicht zwei Auffassungen davon haben, was ein
// offener Kontext ist.
import { useCallback, useRef, useState } from 'react'

import { api } from '../api'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useFremdNachlesen } from './useFremdNachlesen'
import { useLaden } from './useLaden'
import { useLauf } from './useLauf'
import { useVorgang } from './useVorgang'
import { useTexte } from '../sprache/kontext'
import type { RangDTO, TurnierDTO } from '../apiturniere'

export interface Turnierstand {
  turniere: TurnierDTO[]

  // offen ist das aufgeschlagene Turnier — samt Teilnehmern und Verlauf.
  offen: TurnierDTO | null

  // rangliste gehört zum aufgeschlagenen Turnier.
  //
  // **Neben dem Turnier und nicht darin**: sie wird gerechnet, nicht
  // gespeichert (Vermerk 0062), und sie kommt über einen eigenen Ruf.
  rangliste: RangDTO[]

  // geladen sagt, dass die Liste einmal angekommen ist — vorher ist sie
  // unbekannt und nicht leer (UX-08.NAT-01).
  geladen: boolean
  ladefehler: string | undefined
  fehler: string | undefined

  vorgang: ReturnType<typeof useVorgang>

  waehle: (turnierID: string) => Promise<void>

  // lies führt ein **Lesen** aus — mit Vorgangssperre und Fehlermeldung,
  // aber **ohne** Nachladen.
  lies: (marke: string, arbeit: () => Promise<void>) => Promise<void>
  vergiss: (turnierID: string) => void
  neuLaden: () => void
  // amFeld nimmt einen Fehler, der zu einem Feld gehört, der Seite ab — dann
  // steht er dort und nicht über ihr (ENG-05.AK-03, §6).
  fuehre: (
    marke: string,
    arbeit: () => Promise<void>,
    amFeld?: (ursache: unknown) => boolean,
  ) => Promise<void>
}

// useTurniere hält Liste, das aufgeschlagene Turnier und seine Rangliste.
// **`fremdstand` zählt, was ein anderer Prozess oder die Leitung geschrieben hat**
// (#156) — Befund 3 des Gesamtreviews über 0.6.0: ohne ihn lud die Seite erst
// nach einem Wechsel nach. **Behalten**, damit die Liste dabei stehen bleibt.
export function useTurniere(fremdstand = 0): Turnierstand {
  const texte = useTexte()

  const [offen, setzeOffen] = useState<TurnierDTO | null>(null)
  const [rangliste, setzeRangliste] = useState<RangDTO[]>([])
  const [fehler, setzeFehler] = useState<string>()

  const vorgang = useVorgang()
  const lauf = useLauf()

  // **Welches Turnier gerade gilt** — das aufgeschlagene oder das, dessen
  // Lesen läuft. Der Grund steht in `useDecks`: zwischen Klick und Antwort
  // gibt es kein aufgeschlagenes, wohl aber eines, um das es geht.
  const bezug = useRef<string | null>(null)

  const stand = useLaden(
    async () => {
      const antwort = await api().Turniere('')

      return auspacken(antwort, () => antwort.turniere)
    },
    [fremdstand],
    true,
  )

  const neuLaden = stand.nachladen

  const melde = useCallback(
    (ursache: unknown) => {
      setzeFehler(meldungAus(ursache, texte))
    },
    [texte],
  )

  // **Turnier und Rangliste kommen in einem Zug.** Eine Rangliste ohne ihr
  // Turnier gehörte zu nichts, und zwei getrennte Ladewege ließen sie
  // auseinanderlaufen.
  //
  // **Ein Zug ist aber kein gemeinsamer Stand** — berichtigt am 2026-09-07,
  // Zweitblick über die Phasen 4.0 bis 4.13. Beide gehen als eigene Anfrage
  // über die Naht; wer dazwischen ein Ergebnis setzt, sieht ein Turnier von
  // vorhin neben einer Rangliste von jetzt. Der nächste Ladelauf richtet es,
  // und ein Verbund im Backend wäre eine neue Fähigkeit an vier Flächen (§16).
  //
  // **Für den Klick wie für das Nachlesen** — ob die Antwort noch gilt,
  // entscheidet `nummer` (§7, Regel 2).
  const lesen = useCallback(
    async (turnierID: string, nummer: number) => {
      try {
        const [turnier, stand] = await Promise.all([
          api().Turnier(turnierID),
          api().TurnierRangliste(turnierID),
        ])

        if (!lauf.istAktuell(nummer)) {
          return
        }

        setzeOffen(auspacken(turnier, () => turnier.turnier))
        setzeRangliste(auspacken(stand, () => stand.rangliste))
      } catch (ursache: unknown) {
        if (lauf.istAktuell(nummer)) {
          melde(ursache)
        }
      }
    },
    [lauf, melde],
  )

  const waehle = useCallback(
    async (turnierID: string) => {
      const nummer = lauf.naechster()

      // **Nur ein Wechsel räumt** (§7, Regel 1). Nach jedem Schreiben lädt
      // die Seite dasselbe Turnier nach; leerte das den offenen Kontext, baute
      // sich die Werkbank ab und wieder auf, und die gewählte Lasche sprang
      // auf ihre Vorgabe zurück (Befund aus UXR-12, #1920). Dass eine späte
      // Antwort nicht über eine neuere schreibt, hält weiter `lauf`.
      const wechsel = bezug.current !== turnierID

      bezug.current = turnierID

      setzeFehler(undefined)

      if (wechsel) {
        setzeOffen(null)
        setzeRangliste([])
      }

      await lesen(turnierID, nummer)
    },
    [lauf, lesen],
  )

  // **Und nach einem fremden Schreiben noch einmal** — Befund 2 des
  // Nachlaufs zum Gesamtreview über 0.6.0 (#1660). **Leise**: kein Wechsel,
  // also bleibt der Kontext, und eine stehende Meldung räumt ein Nachlesen
  // nicht weg (§7, Regel 3).
  const nachlesen = useCallback(() => {
    const turnierID = bezug.current

    if (turnierID !== null) {
      void lesen(turnierID, lauf.naechster())
    }
  }, [lauf, lesen])

  useFremdNachlesen(fremdstand, nachlesen)

  // vergiss räumt ein Turnier aus dem offenen Kontext — nach dem Löschen.
  const vergiss = useCallback(
    (turnierID: string) => {
      // **Nur das betroffene Turnier.** Wer B löscht, schließt nicht A — und
      // eine Nummer, die unbedingt stiege, verwürfe die Antwort für A gleich
      // mit.
      //
      // **Der Bezug entscheidet und nicht `offen`**: zwischen Klick und
      // Antwort gibt es kein aufgeschlagenes Turnier, wohl aber eines, um
      // das es geht (die Lehre aus Befund 4 des Nachlaufs über Phase 4.6).
      if (bezug.current !== turnierID) {
        return
      }

      lauf.naechster()

      bezug.current = null

      setzeOffen(null)
      setzeRangliste([])
    },
    [lauf],
  )

  // **Ein Lesen lädt nicht nach.**
  //
  // Der Paarungsvorschlag lief zuerst durch [fuehre] — und der räumt nach
  // getaner Arbeit den offenen Kontext und lädt neu. Dabei wurde das Bauteil
  // abgebaut und der Vorschlag mit ihm: er blitzte auf und war weg (Befund 1
  // des Nachlaufs über Phase 4.7).
  //
  // **Die Vorgangssperre bleibt**: ein Vorschlag, der neben einem laufenden
  // Schreiben entsteht, rechnete gegen einen Stand, den der Bestand noch
  // nicht kennt (§7, Regel 4).
  const lies = useCallback(
    async (marke: string, arbeit: () => Promise<void>) => {
      const nummer = lauf.nummer()

      await vorgang.starte(marke, async () => {
        try {
          await arbeit()

          if (lauf.istAktuell(nummer)) {
            setzeFehler(undefined)
          }
        } catch (ursache: unknown) {
          if (lauf.istAktuell(nummer)) {
            melde(ursache)
          }
        }
      })
    },
    [lauf, melde, vorgang],
  )

  // **Nach einem geglückten Schreiben wird nachgeladen** (§7, Regel 3) — und
  // nach einem gescheiterten nicht: sonst überschriebe ein Lesefehler die
  // Meldung, die sagt, was zu tun ist.
  //
  // **Beides: die Liste und das offene Turnier.** Ein Ergebnis ändert die
  // Rangliste, und eine Seite, die nur die Liste nachlädt, zeigte die alte.
  // **Und der Fehler gehört zu dem Stand, aus dem er stammt** (§7, Regel 2).
  //
  // Ein Schreibvorgang für A kann scheitern, während B aufgeschlagen ist:
  // `useVorgang` sperrt nur denselben Schlüssel, und ein Auswahlwechsel
  // dazwischen ist erlaubt. Ohne die Laufnummer stünde die Meldung über B —
  // und jemand versuchte es am falschen Bezugsobjekt noch einmal (Befund 9
  // des Zweitblicks über Phase 4.7).
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

          // **Die Liste lädt immer nach** (§7, Regel 3) — sie gehört keinem
          // offenen Turnier. Wer A löschte und vor der Antwort B öffnete,
          // sah A sonst weiter in der Liste stehen.
          neuLaden()

          if (!lauf.istAktuell(nummer)) {
            return
          }

          setzeFehler(undefined)

          if (bezug.current !== null) {
            await waehle(bezug.current)
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
    [lauf, melde, neuLaden, vorgang, waehle],
  )

  return {
    turniere: stand.daten ?? [],
    offen,
    rangliste,
    geladen: stand.daten !== undefined,
    ladefehler: stand.fehler,
    fehler,
    vorgang,
    waehle,
    lies,
    vergiss,
    neuLaden,
    fuehre,
  }
}
