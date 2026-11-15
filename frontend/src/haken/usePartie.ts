// **Die Partie im Fenster** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Was er hält und was er ausdrücklich nicht ist
//
// **Er hält eine Sicht und nicht den Zustand.** Was hier liegt, ist die
// Projektion eines Sitzes; es gibt keinen Weg von der Wahrheit zur Oberfläche,
// der an ihr vorbeiführt. Wer den Sitz wechselt, bekommt eine **neue** Sicht
// vom Kern — er darf nicht aus der alten gerechnet werden, denn sonst stünde
// das Geheimnis des einen im Speicher des anderen.
//
// **Und er ist nicht der Tisch.** Der GPU-gerenderte Spieltisch mit Kartenflug,
// Mulligan und Perspektivwechsel ist Werkstück G3
// (#1460);
// hier steht der Weg, über den das Fenster den Kern überhaupt erreicht.
//
// # Die vier Regeln aus §7
//
//  1. **Ein Wechsel räumt den offenen Kontext**: `raeumen` legt Partie, Sicht,
//     Ereignisse und Meldung zugleich weg, und `anfangen` räumt von selbst.
//     Ein Sitzwechsel räumt die **Ereignisse** mit — sie gehörten dem Sitz von
//     vorhin, und private Ergänzungen sind empfängergebunden.
//  2. **Eine Antwort gehört zu dem Stand, aus dem sie stammt**: ein Zähler für
//     alle Richtungen, denn sie schreiben in **dieselbe** Fläche. `anfangen`,
//     `ansehen` und `raeumen` zählen weiter, Befehl und Standwechsel merken
//     sich nur — sie wechseln das Bezugsobjekt nicht.
//  3. **Geschrieben ist geschrieben**: nach einem geglückten Befehl wird
//     nachgeladen, weil die Befehlsantwort nur Fassung und Ereignisse trägt.
//     Scheitert das Nachladen, bleibt die Meldung stehen — der **Befehl** ist
//     trotzdem gewirkt, und die nächste Sicht holt ihn. Nach einem
//     gescheiterten Befehl wird gar nicht nachgeladen.
//  4. **Ein Vorgang je Datensatz**: der Datensatz ist die Partie. Ein zweiter
//     Befehl, während der erste läuft, ginge gegen eine Fassung, die der Kern
//     schon nicht mehr führt.
//
// # Woher die Befehlskennung kommt
//
// **Aus dem Fenster, je Handlung eine neue.** Sie schützt vor der doppelten
// Wirkung *derselben* Absicht; ein erneuter Versuch nach einem Fehler ist eine
// neue Absicht, denn der erste hat nichts bewirkt. Was sie hier **nicht**
// leisten muss, ist die verlorene Antwort: zwischen Fenster und Kern liegt
// kein Netz, sondern ein Aufruf im selben Prozess.
import { useCallback, useState } from 'react'

import {
  api,
  auspacken,
  type Partie,
  type PartieAntwort,
  type Partieentwurf,
  type Spielereignis,
} from '../api'
import { meldungAus } from '../meldung'
import { useEigenerTischfunk } from './useEigenerTischfunk'
import {
  partievorgang,
  starthaendeZiehen,
  tischFuerEingeladeneOeffnen,
  standwegFuer,
  type Standwechsel,
} from './partiewege'
import { useLauf } from './useLauf'
import { usePartiebefehle } from './usePartiebefehle'
import { usePartiefremd } from './usePartiefremd'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from './useVorgang'

import { keineRegeln, type Partielauf, type Regelstand } from './partielauf'

export type { Partielauf }

// **`fremdstand` zählt, was ein anderer Prozess an diesem Bestand schrieb**
// (#156) — steigt er, liest der offene Tisch seine Partie nach (#2388).
export function usePartie(fremdstand = 0): Partielauf {
  const vorgang = useVorgang()
  const lauf = useLauf()
  const texte = useTexte()

  const [partie, setzePartie] = useState<Partie>()
  const [wechselt, setzeWechselt] = useState(false)
  const [regeln, setzeRegeln] = useState<Regelstand>(keineRegeln)
  const [befunde, setzeBefunde] = useState<string[]>([])
  const [ereignisse, setzeEreignisse] = useState<Spielereignis[]>([])
  // **Satz und Vorgang in einem Zustand** und nicht in zweien: getrennt
  // zeigte irgendwann ein Bauteil einen Satz, der zum Vorgang eines
  // anderen gehört — genau das, was der Vorgang verhindern soll.
  const [meldung, setzeMeldung] = useState<{ satz: string; vorgang: string }>()

  // **Was ein Gast am eigenen Tisch tut, kommt hier an** (zu Z1, #1672).
  const ereignisluecke = useEigenerTischfunk(
    partie,
    wechselt,
    setzePartie,
    setzeRegeln,
    setzeEreignisse,
  )

  // nurAktuell führt etwas aus, solange der Stand noch der eigene ist.
  //
  // **Einer statt achtmal derselben Zeile** (§2): `if (lauf.istAktuell(…))`
  // stand an jedem Erfolgs- und jedem Fehlerweg, und jede Kopie war ein
  // eigener Zweig, den ein Lauf betreten musste. Die Frage ist überall
  // dieselbe — also steht sie einmal da.
  const nurAktuell = useCallback(
    (meiner: number, arbeit: () => void) => {
      if (lauf.istAktuell(meiner)) {
        arbeit()
      }
    },
    [lauf],
  )

  const raeumen = useCallback(() => {
    // **Aufräumen zählt weiter** — sonst trüge eine Antwort, die noch
    // unterwegs ist, gleich wieder ein, was gerade weggeräumt wurde.
    lauf.naechster()

    setzePartie(undefined)
    setzeWechselt(false)
    setzeRegeln(keineRegeln)
    setzeEreignisse([])
    setzeMeldung(undefined)
  }, [lauf])

  const eintragen = useCallback((antwort: PartieAntwort) => {
    setzePartie(antwort.partie)
    setzeWechselt(false)
    setzeRegeln({ erlaubt: antwort.erlaubt, handlungen: antwort.handlungen })
    setzeBefunde(antwort.befunde)
  }, [])

  const nachladen = useCallback(
    async (kennung: string, fuer: string) => {
      // **Bestenfalls-Arbeit** (Regel 3): der Befehl ist gewirkt, gleich wie
      // das hier ausgeht. Ein Fehler wird gemeldet und nimmt nichts zurück.
      const meiner = lauf.nummer()

      try {
        const antwort = await api().Partie(kennung, fuer)

        auspacken(antwort, () => undefined)

        nurAktuell(meiner, () => {
          eintragen(antwort)
        })
      } catch (ursache: unknown) {
        nurAktuell(meiner, () => {
          setzeMeldung({
            satz: meldungAus(ursache, texte),
            vorgang: `partie-ansehen-${kennung}`,
          })
        })
      }
    },
    [lauf, texte, eintragen, nurAktuell],
  )

  const anfangen = useCallback(
    (entwurf: Partieentwurf) => {
      void vorgang.starte('partie-anfangen', async () => {
        const meiner = lauf.naechster()

        setzeMeldung(undefined)
        setzeEreignisse([])

        let antwort: PartieAntwort

        try {
          antwort = await api().PartieAnlegen(entwurf)

          auspacken(antwort, () => undefined)

          // **Angelegt ist angelegt** (§7, Regel 3; #1970): die leere Sicht
          // daneben sagt nichts. Gelesen wird über den gewöhnlichen Weg —
          // mit Kennung und Sicht, die Go trotzdem nennt.
          if (antwort.nachladefehler) {
            await nachladen(antwort.partie.id, antwort.partie.fuer)

            return
          }

          nurAktuell(meiner, () => {
            eintragen(antwort)
          })
        } catch (ursache: unknown) {
          nurAktuell(meiner, () => {
            setzeMeldung({
              satz: meldungAus(ursache, texte),
              vorgang: 'partie-anfangen',
            })
          })

          return
        }

        // **Die Starthand wird beim Anfangen gezogen** (P6, #2374) — die
        // Partie steht schon; scheitert das Ziehen, bleibt sie, und die
        // Meldung sagt es. Nachgelesen wird nur nach geglücktem Ziehen
        // (Regel 3).
        let gezogen: boolean

        //
        // **Und nach „Partie schließen“ geht nichts weiter** (Nachlauf über
        // Phase 7, R7): wer räumt, zählt den Lauf weiter, und ein Tisch, der
        // nicht mehr dasteht, wird weder freigegeben noch nachgelesen.
        try {
          gezogen = await starthaendeZiehen(antwort.partie)
          if (!lauf.istAktuell(meiner)) return

          await tischFuerEingeladeneOeffnen(antwort.partie)
        } catch (ursache: unknown) {
          nurAktuell(meiner, () => {
            setzeMeldung({
              satz: meldungAus(ursache, texte),
              vorgang: partievorgang(antwort.partie.id),
            })
          })

          return
        }

        if (gezogen && lauf.istAktuell(meiner)) {
          await nachladen(antwort.partie.id, antwort.partie.fuer)
        }
      })
    },
    [vorgang, lauf, texte, eintragen, nurAktuell, nachladen],
  )

  usePartiefremd(fremdstand, partie, wechselt, vorgang, nachladen)

  const ansehen = useCallback(
    (kennung: string, fuer: string) => {
      void vorgang.starte(`partie-ansehen-${kennung}`, async () => {
        // **Ein Sitzwechsel ist ein Wechsel des Bezugsobjekts** (Regel 1):
        // die Sicht gehört dem Sitz, und jede Antwort für den vorigen ist
        // damit überholt.
        const meiner = lauf.naechster()

        setzeMeldung(undefined)
        setzeWechselt(true)

        // **Und die Ereignisse gehen mit.** Sie tragen die private
        // Ergänzung **eines** Empfängers; am neuen Sitz stehen zu lassen,
        // was der alte bekommen hat, wäre genau die Vermischung, die es
        // hier nicht geben darf.
        setzeEreignisse([])

        try {
          const antwort = await api().Partie(kennung, fuer)

          auspacken(antwort, () => undefined)

          nurAktuell(meiner, () => {
            eintragen(antwort)
          })
        } catch (ursache: unknown) {
          nurAktuell(meiner, () => {
            setzeMeldung({
              satz: meldungAus(ursache, texte),
              vorgang: `partie-ansehen-${kennung}`,
            })
          })
        }
      })
    },
    [vorgang, lauf, texte, eintragen, nurAktuell],
  )

  // **Befehle der Reihe nach** — P5 (#2373), in `usePartiebefehle`.
  const { befehlen, wartend } = usePartiebefehle({
    partie,
    vorgang,
    lauf,
    nachladen,
    nurAktuell,
    setzeMeldung,
    setzeEreignisse,
  })

  const wechseln = useCallback(
    (wohin: Standwechsel) => {
      const offen = partie
      if (!offen) {
        return
      }

      void vorgang.starte(partievorgang(offen.id), async () => {
        const meiner = lauf.nummer()

        setzeMeldung(undefined)

        try {
          const antwort = await standwegFuer(wohin)(offen.id, offen.fuer)

          auspacken(antwort, () => undefined)

          // **Gewechselt ist gewechselt** (#1970) — dieselbe Regel wie
          // beim Anlegen.
          if (antwort.nachladefehler) {
            await nachladen(offen.id, offen.fuer)

            return
          }

          nurAktuell(meiner, () => {
            // **Der Standwechsel gibt die neue Sicht selbst zurück** — hier
            // gibt es kein Nachladen und damit auch keinen zweiten Weg, an
            // dem etwas schiefgehen könnte.
            eintragen(antwort)
          })
        } catch (ursache: unknown) {
          nurAktuell(meiner, () => {
            setzeMeldung({
              satz: meldungAus(ursache, texte),
              vorgang: partievorgang(offen.id),
            })
          })
        }
      })
    },
    [partie, vorgang, lauf, texte, eintragen, nurAktuell, nachladen],
  )

  const laeuft = useCallback((was: string) => vorgang.laeuft(was), [vorgang])

  return {
    anfangen,
    ansehen,
    befehlen,
    wechseln,
    raeumen,
    laeuft,
    partie,
    wechselt,
    erlaubt: regeln.erlaubt,
    handlungen: regeln.handlungen,
    wartend,
    befunde,
    ereignisse,
    ereignisluecke,
    fehler: meldung?.satz,
    fehlerVorgang: meldung?.vorgang,
  }
}
