// **Die Massenänderung mit Vorschau** — Werkstück U6 von Phase 5.1
// (#1417).
//
// # Zwei Schritte statt einem
//
// Bis U6 schrieb »Anwenden« sofort. Jetzt zeigt »Vorschau« je Karte, was
// sich ändern würde — Bisher, Danach, und ob eine Karte schon so steht —,
// und »Anwenden« trägt die gezeigten Altwerte als **Erwartung** mit: der
// Bestand vergleicht, bevor er schreibt, und wer zwischen Vorschau und
// Anwenden geschrieben hat, macht daraus einen Konflikt statt einer stillen
// Überschreibung.
//
// # Die vier Regeln aus AGENTS.md §7
//
//  1. **Ein Wechsel räumt den Kontext**: eine andere Auswahl, ein anderes
//     Feld, ein anderer Wert — die Vorschau ist dann keine mehr und geht weg.
//  2. **Eine Antwort gehört zu ihrem Stand**: `stand` zählt die Wechsel, und
//     eine Antwort aus einem alten Stand wird verworfen.
//  3. **Geschrieben ist geschrieben**: nach dem Anwenden wird nachgeladen;
//     ein Konflikt lädt nicht nach — die Vorschau bleibt stehen und zeigt,
//     welche Karten inzwischen anders sind.
//  4. **Ein Vorgang je Auswahl**: `useVorgang` hält den zweiten Anlauf an.
import { useCallback, useEffect, useRef, useState } from 'react'

import {
  api,
  auspacken,
  type FeldbeschreibungDTO,
  type KartenzeileDTO,
  type MassenvorschauAntwort,
} from '../api'
import { useLaden } from '../haken/useLaden'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Massenfelder } from './Massenfelder'
import { Massenvorschau, faltschritt, faltvorn } from './Massenvorschau'
import { zeilenname } from './kartenname'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { Faltfuss } from '../bausteine/Faltliste'

interface MassenaenderungEigenschaften {
  setID: string
  // karten sind die angehakten Karten — die Auswahl, um die es geht.
  karten: readonly KartenzeileDTO[]
  // nachladen holt die Liste neu, wenn geschrieben wurde.
  nachladen: () => void
}

export function Massenaenderung({
  setID,
  karten,
  nachladen,
}: MassenaenderungEigenschaften) {
  const texte = useTexte()
  const worte = texte.seiten.karten.massen
  const vorgang = useVorgang()
  const [feld, setzeFeld] = useState('')
  const [wert, setzeWert] = useState('')
  const [fehler, setzeFehler] = useState<string>()
  const [vorschau, setzeVorschau] = useState<MassenvorschauAntwort>()
  const [konflikt, setzeKonflikt] = useState<readonly string[]>([])
  const [angewendet, setzeAngewendet] = useState<number>()

  // **Wie viele der Auswahl dastehen** (UXR-06, #1914) — zehn vorn, der Rest
  // stückweise; gerufen wird trotzdem für alle.
  const [gezeigt, setzeGezeigt] = useState(faltvorn)

  const felder = useLaden<FeldbeschreibungDTO[]>(
    useCallback(async () => {
      const antwort = await api().Kartenfelder(setID)

      return auspacken(antwort, () => antwort.felder)
    }, [setID]),
    [setID],
  )
  const angeboten = (felder.daten ?? []).filter(
    (eines) => eines.editierbar && eines.gespeichert,
  )
  const gewaehltesFeld = angeboten.find((eines) => eines.name === feld)

  const kennungen = karten.map((karte) => karte.id)
  const schluessel = kennungen.join(',')

  // **Ein Wechsel räumt den Kontext** (Regel 1) — und zählt den Stand hoch
  // (Regel 2). Die Vorlage hält fest, wofür die Vorschau galt.
  const [vorlage, setzeVorlage] = useState(schluessel)
  const [stand, setzeStand] = useState(0)

  if (vorlage !== schluessel) {
    setzeVorlage(schluessel)
    setzeStand((zahl) => zahl + 1)
    setzeVorschau(undefined)
    setzeKonflikt([])
    setzeFehler(undefined)
    setzeAngewendet(undefined)
  }

  const raeumen = () => {
    setzeStand((zahl) => zahl + 1)
    setzeVorschau(undefined)
    setzeKonflikt([])
    setzeFehler(undefined)
    setzeAngewendet(undefined)
  }

  const laeuft = vorgang.laeuft(schluessel)

  // **Die gespeicherten Werte des gewählten Feldes** — vor der Vorschau steht
  // neben jeder Karte, was sie heute trägt; die Vorschau zeigt es dann als
  // „Bisher" aus dem Bestand.
  const holenWerte = useCallback(async () => {
    if (feld === '') {
      return new Map<string, string>()
    }

    const antwort = await api().Kartenfeldwerte(setID, feld)

    return auspacken(
      antwort,
      () => new Map(antwort.werte.map((wert) => [wert.id, wert.wert])),
    )
  }, [setID, feld])
  const gespeicherte = useLaden<Map<string, string>>(holenWerte, [setID, feld])

  const jetzigerStand = useRef(stand)
  useEffect(() => {
    jetzigerStand.current = stand
  }, [stand])

  const vorschauen = async () => {
    const gestartetMit = stand

    await vorgang.starte(schluessel, async () => {
      setzeFehler(undefined)
      setzeKonflikt([])
      setzeAngewendet(undefined)

      try {
        const antwort = await api().FelderVorschau(setID, kennungen, feld, wert)
        auspacken(antwort, () => undefined)

        if (jetzigerStand.current === gestartetMit) {
          setzeVorschau(antwort)
        }
      } catch (ursache: unknown) {
        if (jetzigerStand.current === gestartetMit) {
          setzeFehler(meldungAus(ursache, texte))
        }
      }
    })
  }

  const anwenden = async () => {
    /* v8 ignore next 3 -- der Knopf »Anwenden« ist ohne Vorschau angehalten;
       der Wächter ist für TypeScript da, das den Knopf nicht sieht. */
    if (vorschau === undefined) {
      return
    }

    const gestartetMit = stand
    // **Die Erwartung sind die Zeilen, die geschrieben würden** — die
    // ausgeschlossenen bleiben weg, für sie wird nichts geschrieben.
    const erwartet = vorschau.zeilen
      .filter((zeile) => zeile.ausschluss === '')
      .map((zeile) => ({
        karteID: zeile.karteID,
        wert: zeile.alt,
        gabEs: zeile.gabAlt,
      }))

    await vorgang.starte(schluessel, async () => {
      setzeFehler(undefined)
      setzeKonflikt([])

      let betroffene: string[] = []

      try {
        const antwort = await api().FelderSetzenMitErwartung(
          setID,
          erwartet,
          feld,
          wert,
        )
        betroffene = antwort.betroffen
        auspacken(antwort, () => undefined)
      } catch (ursache: unknown) {
        if (jetzigerStand.current !== gestartetMit) {
          return
        }

        // **Ein Konflikt lässt die Vorschau stehen** und nennt, welche
        // Karten inzwischen anders sind (Regel 3: nichts nachladen).
        setzeKonflikt(betroffene)
        setzeFehler(meldungAus(ursache, texte))

        return
      }

      if (jetzigerStand.current === gestartetMit) {
        setzeAngewendet(erwartet.length)
        setzeVorschau(undefined)
        nachladen()
      }
    })
  }

  if (karten.length === 0) {
    return <p className="mt-3 max-w-prose text-leise">{worte.keineAuswahl}</p>
  }

  const bereit = feld !== '' && wert !== '' && !laeuft
  const aenderbar =
    vorschau?.zeilen.filter((zeile) => zeile.ausschluss === '') ?? []

  return (
    <>
      <Erklaerungsabsatz
        text={worte.erklaerung}
        klasse="mt-1 max-w-prose text-leise"
      />
      <Massenfelder
        angeboten={angeboten}
        feld={feld}
        gewaehltesFeld={gewaehltesFeld}
        wert={wert}
        feldWaehlen={(neues) => {
          setzeFeld(neues)
          setzeWert('')
          raeumen()
        }}
        wertSetzen={(neuer) => {
          setzeWert(neuer)
          raeumen()
        }}
      />
      <div className="mt-3 flex flex-wrap items-center gap-3">
        <button
          type="button"
          disabled={!bereit}
          onClick={() => {
            void vorschauen()
          }}
          className="rounded-sm border border-kante px-3 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
        >
          {worte.vorschau}
        </button>
        <button
          type="button"
          disabled={laeuft || aenderbar.length === 0 || konflikt.length > 0}
          onClick={() => {
            void anwenden()
          }}
          className="rounded-sm border border-kante bg-marke px-3 py-1 text-marketext focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
        >
          {worte.anwenden}
        </button>
        <p className="text-leise">
          {karten.length} {worte.ausgewaehlt}
        </p>
      </div>
      {fehler !== undefined && (
        <p role="alert" className="mt-3 text-fehler">
          {texte.bausteine.zustaende.fehler} {fehler}
        </p>
      )}
      {konflikt.length > 0 && (
        <p className="mt-1 max-w-prose text-fehler">{worte.konflikt}</p>
      )}
      {angewendet !== undefined && (
        <p role="status" className="mt-3 text-leise">
          {einsetzen(worte.angewendet, { anzahl: String(angewendet) })}
        </p>
      )}
      {vorschau === undefined ? (
        <>
          <ul className="mt-3 divide-y divide-kante">
            {karten.slice(0, gezeigt).map((karte) => {
              const heute = gespeicherte.daten?.get(karte.id)

              return (
                <li key={karte.id} className="flex items-baseline gap-3 py-1">
                  <span>{zeilenname(karte, texte.seiten.karten.ohneName)}</span>
                  {feld !== '' && (
                    <span className="text-leise">
                      {heute === undefined || heute === ''
                        ? worte.ohneWert
                        : heute}
                    </span>
                  )}
                </li>
              )
            })}
          </ul>
          {karten.length > faltvorn && (
            <Faltfuss
              gezeigt={Math.min(gezeigt, karten.length)}
              gesamt={karten.length}
              schritt={faltschritt}
              weiter={() => {
                setzeGezeigt(gezeigt + faltschritt)
              }}
            />
          )}
        </>
      ) : (
        <Massenvorschau vorschau={vorschau} konflikt={konflikt} />
      )}
    </>
  )
}
