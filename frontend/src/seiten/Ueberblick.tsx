// **Der Überblick über die eigene Arbeit** — Werkstück 3 von Phase 3.5
// (#1166).
//
// # Warum er die ersten Schritte ablöst
//
// „Erste Schritte" stand **für immer** da. Bei einem leeren Bestand ist das
// genau richtig — §6 verlangt vom leeren Zustand, dass er den nächsten Schritt
// erklärt. Bei zwölf Sets und dreitausend Karten ist es eine Anleitung für
// jemanden, der längst weiter ist, und sie steht dort, wo die Antwort auf
// „woran habe ich gearbeitet?" hingehört.
//
// **Die ersten Schritte sind damit nicht weg.** Sie stehen, solange es nichts
// zu überblicken gibt — und das ist genau, wenn man sie braucht.
//
// # Eine Zeile und kein Raster — UXR-03
//
// #1911,
// gemessen am 2026-10-02 am nativen Fenster: hier standen vier Kacheln mit
// je einer Überschrift, einem Erklärsatz und einem Knopf „Zu …", darüber eine
// fünfte Überschrift — **fünf von dreizehn Überschriften der Startseite für
// vier Zahlen**. Das Artboard *Main* zeigt sie als eine Zeile: die Zahl groß,
// das Wort klein darunter.
//
// **Der Weg bleibt**: jede Kennzahl ist selbst der Knopf zu der Seite, auf der
// ihre Zahl entsteht. Ihr Name ist, was dasteht — „357 Karten" —, und das Wort
// zählt mit, denn unter einer 1 steht „Set" und nicht „Sets" (#1218).
//
// # Was sie zählt und was nicht
//
// Vier Zahlen, und jede beantwortet eine Frage, die jemand wirklich hat: wie
// viele Sets, wie viele Karten darin, wie viele Karten besitze ich, wie viele
// Decks. **Decks statt Vorlagenpakete**, wie im Artboard (UXR-03): die
// Startseite fragt „woran arbeite ich weiter?", und ein Deck ist Arbeit, ein
// Paket Material — dessen Zahl geht in einem echten Bestand in die Hunderte
// und sagt darüber nichts. Die Pakete erreicht man über die Leiste (⌘6) und
// über den leeren Zustand von *Weiterarbeiten*. **Keine Fortschrittsbalken
// und keine Ziele** — diese Anwendung weiß nicht, was jemand vorhat.
//
// # Ein Lesefehler hält nichts auf
//
// Die Zahlen sind eine Zugabe; wer die Anwendung öffnet, will damit arbeiten.
// Eine Startseite, die wegen einer unlesbaren Deckliste gar nichts zeigt,
// wäre der schlechtere Tausch (§8). Was nicht zu holen war, zählt null — und
// die übrigen stehen da.
import { useEffect } from 'react'

import { useAnschluss } from '../anschlussraum'
import type { Problemlage } from './problemlage'
import { problemkacheln, type Kachel } from './problemkacheln'
import { Problemzeile } from './Problemzeile'
import { sammlungsunterzeile } from './sammlungskachel'
import { deckkachel } from './deckkachel'
import type { AufmerksamkeitAntwort } from '../apiaufmerksamkeit'
import type { Portfolioantwort } from '../api'

import type { Seitenschluessel } from '../navigation'
import { useDarstellung, useTexte } from '../sprache/kontext'
import { zaehlform, type Zaehlform } from '../sprache/zaehlwort'

export interface Ueberblickszahlen {
  sets: number
  karten: number
  exemplare: number
  decks: number
}

// kennzahlmarke ist die Marke, die der Überblick bei seinem Erscheinen setzt
// — PERF02 (#2352).
// **Er erscheint erst mit seinen Zahlen** (`Start.tsx`), und daran misst
// `kennzahlen09`, wie lange Start und Startseite bis zur sichtbaren Kennzahl
// brauchen. Eine Marke ändert am Bild nichts.
export const kennzahlmarke = 'mse:kennzahlen'

interface UeberblickEigenschaften {
  // zahlen fehlen, solange die Antwort aussteht — dann stehen Platzhalter da.
  zahlen: Ueberblickszahlen | undefined

  // probleme ist, was die Prüfstände der Sets sagen — `undefined`, solange sie
  // ausstehen oder nicht zu lesen waren (AP-8, #2365). **Sie kommen für sich**,
  // damit die Zahlen nicht auf sie warten.
  probleme: Problemlage | undefined

  // portfolio ist die Übersicht über die Sammlung — für die Unterzeile der
  // Kachel „Sammlung“, und **ebenso für sich** (AP-8, Teil 2b).
  portfolio?: Pick<Portfolioantwort, 'zeilen' | 'ohnePreis'> | undefined

  // aufmerksamkeit ist dieselbe Antwort, die die Liste darunter zeigt — die
  // Kachel „Decks“ zählt aus ihr die Decks mit fehlenden Karten.
  aufmerksamkeit?: AufmerksamkeitAntwort | undefined
  wechseln: (seite: Seitenschluessel) => void

  // waehleSet öffnet ein Set — vor der Kartenliste, in die eine Kachel führt.
  waehleSet: (setID: string, karteID?: string) => void

  // nachladen liest die Zahlen neu — nach einer Prüfung der Sets.
  nachladen: () => void
}

// kennzahl ist ein Knopf wie eine Fläche: Rand und Grund wie `Flaeche`, damit
// die Zeile zu den Bereichen darunter gehört, und ein Fokusring wie jeder
// Knopf (§6).
// **Der Platzhalter hat die Maße der Kachel** — AP-4
// (#2361),
// Befund B-12, Entwurf `Zustaende09` („Laden in Endhöhe, ohne Sprung“).
// Dieselben Ränder und Abstände, darin derselbe Text **unsichtbar**: so ist
// die Höhe die der fertigen Kachel und nicht geschätzt. Der Grund `betont`
// sagt, dass hier etwas kommt.
const platzhalter =
  'flex w-full flex-col items-start rounded-lg border border-kante bg-betont px-4 py-3'

// platzhalterzahl steht unsichtbar in einer Kachel, die noch lädt.
const platzhalterzahl = 0

const kennzahl =
  'flex w-full flex-col items-start rounded-lg border border-kante ' +
  'bg-flaeche px-4 py-3 text-left hover:bg-betont focus-visible:outline ' +
  'focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring'

// unterklasse ist die Unterzeile einer Kachel — klein und leise, immer eine
// Zeile hoch.
const unterklasse = 'block min-h-4 text-xs text-leise'

export function Ueberblick({
  zahlen,
  probleme,
  portfolio,
  aufmerksamkeit,
  wechseln,
  waehleSet,
  nachladen,
}: UeberblickEigenschaften) {
  const texte = useTexte()
  const darstellung = useDarstellung()
  const woerter = texte.allgemein.zaehlwoerter
  const anschluss = useAnschluss()
  const { unterzeile, auswahl } = problemkacheln(probleme, texte)
  const decks = deckkachel(aufmerksamkeit, texte)

  // **Die Marke fällt, wenn die Zahlen dastehen** — nicht, wenn der
  // Platzhalter erscheint: PERF misst an ihr, wann die Kennzahlen sichtbar
  // sind.
  const geladen = zahlen !== undefined
  useEffect(() => {
    if (geladen) {
      performance.mark(kennzahlmarke)
    }
  }, [geladen])

  const kacheln: {
    schluessel: Kachel
    form: Zaehlform
    ziel: Seitenschluessel
  }[] = [
    { schluessel: 'sets', form: woerter.sets, ziel: 'sets' },

    // **Die Kartenzahl führt zu den Sets und nicht zu den Karten.**
    //
    // Sie zählt über **alle** Sets; die Kartenlasche zeigt die eines
    // **einen**. Ohne offenes Set endete der Knopf im Leerzustand „kein Set
    // gewählt", mit einem erklärte das Ziel die genannte Zahl nicht — eine
    // Kachel, deren Zahl und Handlung verschiedene Bezugsobjekte haben.
    // Angestrichen vom Zweitblick über Phase 3.5 als Befund 3.
    //
    // **Die Setliste ist der Ort, an dem diese Zahl zerfällt**: dort steht je
    // Set, wie viele Karten es hat.
    { schluessel: 'karten', form: woerter.karten, ziel: 'sets' },
    { schluessel: 'exemplare', form: woerter.exemplare, ziel: 'sammlung' },
    { schluessel: 'decks', form: woerter.decks, ziel: 'decks' },
  ]

  return (
    <>
      <ul
        aria-label={texte.seiten.start.ueberblick.titel}
        aria-busy={geladen ? undefined : true}
        className="mb-6 grid grid-cols-2 gap-3 sm:grid-cols-4"
      >
        {kacheln.map(({ schluessel, form, ziel }) => {
          if (zahlen === undefined) {
            return (
              <li key={schluessel}>
                <div aria-hidden="true" className={platzhalter}>
                  <span className="invisible text-2xl font-semibold tabular-nums">
                    {platzhalterzahl}
                  </span>
                  <span className="invisible text-sm text-leise">
                    {zaehlform(platzhalterzahl, form)}
                  </span>
                  <span className={unterklasse} />
                </div>
              </li>
            )
          }

          const zahl = zahlen[schluessel]
          const unter =
            schluessel === 'exemplare'
              ? sammlungsunterzeile(portfolio, texte, darstellung)
              : schluessel === 'decks'
                ? decks.unterzeile
                : unterzeile(schluessel)
          const gefiltert =
            schluessel === 'decks' ? decks.ziel : auswahl(schluessel)

          return (
            <li key={schluessel}>
              <button
                type="button"
                className={kennzahl}
                onClick={() => {
                  // **Mit gesetztem Filter, wo die Zahl einen trägt** (AP-8,
                  // B-13); ohne bekannte Probleme die ganze Liste.
                  if (gefiltert === undefined) {
                    wechseln(ziel)
                    return
                  }

                  // **Erst das Set mit seiner ersten Karte mit Befund, dann
                  // die Liste** — die Kartenliste zeigt das offene Set, und
                  // die Auswahl gilt nur für ihres.
                  if (gefiltert.art === 'kartenauswahl') {
                    waehleSet(gefiltert.setID, gefiltert.karteIDs[0])
                  }
                  anschluss.uebergib(gefiltert)
                }}
              >
                {/* **Die Zahl zuerst und groß.** Sie ist die Antwort; das Wort
                  sagt nur, worauf. `tabular-nums`, damit vier Zahlen
                  nebeneinander nicht unterschiedlich breit wirken, wenn sich
                  eine ändert. **Das Leerzeichen dazwischen ist für den
                  Namen** — „357 Karten" und nicht „357Karten"; im Bild
                  stehen beide untereinander. */}
                <span className="text-2xl font-semibold tabular-nums">
                  {zahl}
                </span>{' '}
                <span className="text-sm text-leise">
                  {zaehlform(zahl, form)}
                </span>
                {unter !== undefined && ' '}
                {/* **Die Zeile steht immer**, leer, wo nichts zu sagen ist —
                  sonst wüchse die Kachel beim Eintreffen der Prüfstände
                  (AP-4, Laden in Endhöhe). */}
                <span className={unterklasse}>{unter}</span>
              </button>
            </li>
          )
        })}
      </ul>
      <Problemzeile
        lage={probleme}
        geladen={probleme !== undefined}
        nachladen={nachladen}
      />
    </>
  )
}
