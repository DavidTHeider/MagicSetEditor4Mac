// Die Liste der Karten eines Sets.
//
// **Aus `Karten.tsx` herausgelöst** (#288),
// als die Kartenlasche drei Spalten bekam: Liste, Felder, Vorschau. Der
// Schnitt folgt der Frage — hier steht, *welche* Karte gewählt ist, dort, *was*
// mit ihr geschieht (§2).
import { useMemo } from 'react'

import { type KartenzeileDTO } from '../api'
import { useListenbilder } from '../haken/useListenbilder'
import { useSichtfenster } from '../haken/useSichtfenster'
import { useCallback, useRef, useState, type KeyboardEvent } from 'react'

import { Objektkontextmenue } from '../bausteine/Objektkontextmenue'
import { useAnstoss } from '../haken/useAnstoss'

import { useTexte } from '../sprache/kontext'
import { useZeilenfokus } from '../haken/useZeilenfokus'
import { blaettertaste, zielstelleVon, zurZeile } from './kartenlistetasten'
import { Kartenzeile } from './Kartenzeile'
export type { Listenzusatz } from './listenzusatz'
import type { Listenzusatz } from './listenzusatz'
import { zeilenname } from './kartenname'

interface KartenlisteEigenschaften {
  // setID sagt, wessen Bilder zu holen sind (#1166).
  setID: string

  // stand steigt bei jeder Änderung am Bestand — die Bilder hängen daran
  // (#1215,
  // Befund 5): wer ein Kartenbild einsetzt, ändert die Datei, auf die ein
  // gemerktes Listenbild zeigt.
  stand: number

  karten: readonly KartenzeileDTO[]
  gewaehlt: string | undefined
  waehlen: (karteID: string) => void

  // angehakt sind die Karten für die Massenänderung.
  angehakt: readonly string[]
  anhaken: (karteID: string, an: boolean) => void

  // mitKaestchen zeichnet die Kästchen dafür — nur im Auswahlmodus
  // (UXR-05, #1913).
  mitKaestchen: boolean

  // spalten sind die **zusätzlichen** Spalten neben dem Namen — Werkstück 3
  // von Phase 3.7
  // (#1168).
  //
  // **Leer ist der Normalfall.** `magic.mse-game` führt 243 Felder mit einer
  // Stelle und elf sichtbare; wer nichts wählt, sieht die Liste wie zuvor.
  spalten: readonly Listenzusatz[]

  // marken ist Kartenkennung → ihre Marken
  // (#1208).
  //
  // **Sie stehen in der Zeile und nicht nur im Editor.** Wer ein Set
  // durchsieht, will sehen, was schon abgehakt ist — dafür sind sie da.
  marken: Readonly<Record<string, readonly string[]>>
}

// Listenzusatz ist eine gewählte Spalte samt den Werten, die sie zeigt.
//
// **Die Werte kommen mit und werden nicht nachgeschlagen**: die Liste zeichnet
// je Rollschritt zwanzig Zeilen, und ein Nachschlagen je Zelle ginge über

export function Kartenliste({
  setID,
  stand,
  karten,
  gewaehlt,
  waehlen,
  angehakt,
  anhaken,
  mitKaestchen,
  spalten,
  marken,
}: KartenlisteEigenschaften) {
  const texte = useTexte()

  // **Gezeichnet wird, was zu sehen ist** (#891). Bei 3.788 Karten kostete
  // jeder Klick in die Liste eine Langaufgabe von 73 bis 132 ms — gemessen am
  // 2026-08-25 in der laufenden Anwendung.
  const {
    behaelter,
    zeilenmass,
    erste,
    letzte,
    obenHoch,
    untenHoch,
    zeilenhoehe,
  } = useSichtfenster(karten.length)

  // **Das Gitter noch einmal als Griff** (#1290): `useSichtfenster` gibt eine
  // Rückrufreferenz und behält den Knoten für sich. Zum Rollen braucht dieser
  // Bauteil ihn selbst.
  const gitter = useRef<HTMLElement | null>(null)

  // tastenziel ist die Karte, auf die eine **Taste** gezeigt hat — nur dann
  // wandert der Fokus mit (siehe [useZeilenfokus]).
  const [tastenziel, setzeTastenziel] = useState<string>()

  useZeilenfokus(
    gitter,
    tastenziel,
    useCallback(() => {
      setzeTastenziel(undefined)
    }, []),
    gewaehlt,
    [erste, letzte],
  )

  // **Und die Bilder dazu** (#1166). MSE2 zeigt sie seit jeher
  // (`FilteredImageCardList`, `src/gui/set/cards_panel.cpp`, Zeile 37); für
  // ein Werkzeug, mit dem man Karten **gestaltet**, ist eine Liste aus
  // Textzeilen die falsche Darstellung.
  // **Die Tabstelle liegt auf der gewählten Karte — wenn sie gezeichnet ist.**
  //
  // Befund 5 des Zweitblicks über Phase 4.3: die Liste zeichnet nur ihr
  // Sichtfenster (#891). Wer von der gewählten Karte wegrollt, hat sie nicht
  // mehr im Baum — und **keine** gezeichnete Zeile nahm dann den Tabulator an.
  // Nachgestellt mit 200 Karten, Auswahl 0, gerollt zu Karte 92: **null**
  // Elemente mit `tabindex="0"`, und das Gitter war per Tastatur nicht mehr zu
  // erreichen.
  //
  // **Sonst die erste gezeichnete.** Der Tabulator führt in die Liste; wohin
  // genau, entscheidet sie.
  const gezeichnete = karten.slice(erste, letzte)
  const tabkarte = gezeichnete.some((karte) => karte.id === gewaehlt)
    ? gewaehlt
    : gezeichnete[0]?.id

  const sichtbare = useMemo(
    () => karten.slice(erste, letzte).map((karte) => karte.id),
    [karten, erste, letzte],
  )

  const bilder = useListenbilder(setID, sichtbare, stand)

  // **Ein Rechtsklick oder ⇧F10 auf einer anderen Karte wählt sie zuerst**
  // — AP-5 (#2362), Befund B-01; wie, steht am Haken.
  const { anstoss, verbraucht, anstossen } = useAnstoss<string>(
    gewaehlt,
    waehlen,
  )

  return (
    // **Das Menü der gewählten Karte liegt um die ganze Liste** — ein
    // Rahmen statt einer je Zeile, denn gezeichnet wird nur das Sichtfenster.
    <Objektkontextmenue
      art="karte"
      kennung={gewaehlt ?? ''}
      klasse="flex min-h-0 flex-col"
      anstoss={anstoss}
      anstossVerbraucht={verbraucht}
      nurAngestossen
    >
      {/* **Die Liste scrollt für sich.** Bei 3.000 Karten scrollte sonst die
          ganze Seite, und die Felder daneben wären weg, sobald man weit unten
          sucht.
          **Ein Gitter und keine Liste** (#1290) — eine Tabstelle statt 54.
          Warum `grid` und nicht `listbox`, steht in `kartenlistetasten.ts`. */}
      <div
        ref={(knoten) => {
          // **Die Aufräumfunktion wird durchgereicht, nicht verschluckt.**
          //
          // `useSichtfenster` hängt an den Knoten einen `scroll`-Horcher und
          // einen `ResizeObserver` und gibt dafür eine Aufräumfunktion zurück.
          // Der erste Anlauf rief `behaelter(knoten)` und gab **nichts** zurück
          // — React nimmt dann seinen alten Weg und ruft die Referenz beim
          // Abräumen mit `null`, und dort steigt sie sofort aus. Horcher und
          // Beobachter blieben hängen, an einem Knoten, den es nicht mehr gibt.
          gitter.current = knoten

          const aufraeumen = behaelter(knoten)

          return () => {
            gitter.current = null
            aufraeumen?.()
          }
        }}
        role="grid"
        aria-label={texte.seiten.karten.titel}
        aria-rowcount={karten.length}
        className="divide-y divide-kante overflow-y-auto"
      >
        {/* **Die Abstandhalter tragen den Rollbalken.** Ohne sie wäre die Liste
          so hoch wie ihr Fenster, und niemand käme über die zwanzigste Karte
          hinaus. `role="presentation"`, weil ein Kind eines Gitters sonst eine
          Zeile sein müsste — und sie sind keine. */}
        {obenHoch > 0 && (
          <div role="presentation" style={{ height: obenHoch }} />
        )}

        {karten.slice(erste, letzte).map((karte, imFenster) => {
          const stelle = erste + imFenster
          const name = zeilenname(karte, texte.seiten.karten.ohneName)
          const ist = karte.id === gewaehlt

          // **Einmal aufgelöst und nicht zweimal** (§3): ein zweites `?? []`
          // wäre ein Zweig, den kein Aufruf betreten kann — die Liste ist an
          // dieser Stelle schon nichtleer.
          const eigene = marken[karte.id] ?? []

          return (
            <Kartenzeile
              key={karte.id}
              karte={karte}
              stelle={stelle}
              ist={ist}
              name={name}
              eigene={eigene}
              spalten={spalten}
              waehlen={waehlen}
              angehakt={angehakt}
              anhaken={anhaken}
              mitKaestchen={mitKaestchen}
              bild={bilder[karte.id]}
              tabstelle={karte.id === tabkarte}
              zeilenmass={imFenster === 0 ? zeilenmass : undefined}
              aufRechtsklick={(ereignis) => {
                anstossen(ereignis, karte.id, karte.id)
              }}
              aufTaste={(ereignis: KeyboardEvent<HTMLElement>) => {
                anstossen(ereignis, karte.id, karte.id)
                if (ereignis.defaultPrevented) {
                  return
                }

                // **Erst rollen, dann wählen.** Die Zeile am Ende der Liste ist
                // nicht gezeichnet; ohne diesen Zug fände der Fokus sie nie.
                const ziel = zielstelleVon(ereignis.key, stelle, karten.length)
                if (ziel !== undefined) {
                  setzeTastenziel(karten[ziel]?.id)
                  zurZeile(gitter.current, ziel, zeilenhoehe)
                }

                blaettertaste(ereignis, karten, stelle, waehlen)
              }}
            />
          )
        })}

        {untenHoch > 0 && (
          <div role="presentation" style={{ height: untenHoch }} />
        )}
      </div>
    </Objektkontextmenue>
  )
}
