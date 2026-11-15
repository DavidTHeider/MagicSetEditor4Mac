// **Rücknahme je Handgriff im Symboleditor** —
// #1340,
// Werkstück F von Phase 4.11.
//
// # Was hier portiert ist
//
// `ActionStack` (`src/util/action_stack.cpp` Zeilen 21–51, Fassung `f9d9356`
// vom 2021-07-06) und die Frage, wann zwei Handlungen **eine** sind.
//
// MSE beantwortet sie an drei Stellen, und alle drei stehen hier drin:
//
//  1. **`ExtendableAction`** (`symbol_part.hpp` Zeilen 45–70). Ein Zug legt
//     **eine** Handlung an und schreibt sie mit `move(delta)` fort, solange
//     der Zeiger unten ist.
//
//     **Und ein Tastendruck tut das nicht** (`point_editor.cpp` Zeilen
//     355–380): dort steht `resetActions()`, dann eine **neue** Handlung, und
//     `ControlPointMoveAction` kennt kein `merge`. MSE unterscheidet die
//     beiden am **Typ** — `ExtendableAction` oder nicht —, und deshalb steht
//     es hier als `fortschreibbar` am Handgriff und nicht an der Marke. Wer
//     es vergisst, bekommt einen Schritt zu viel; wer es an der Marke
//     festmachte, bekäme bei jedem zweiten Tastendruck einen zu wenig.
//  2. **`merge()`** (`action_stack.cpp` Zeile 32). Nur `last_was_add` erlaubt
//     es: nach einer Rücknahme wird **nicht** verschmolzen, sonst verschwände
//     der Schritt, auf den sie gerade gezeigt hat.
//  3. **`redo_actions.clear()`** (ebd. Zeile 27). Etwas Neues löscht, was
//     wiederherzustellen war.
//
// # Was hier **nicht** steht
//
// **Der Verlauf des Bestands.** `useVerlauf` daneben nimmt zurück, was
// **geschrieben** ist; dieser hier arbeitet vor dem Schreiben, in der offenen
// Zeichnung. Das ist auch der Grund, warum er an keiner anderen Fläche steht
// (§16): eine Kommandozeile hat keine offene Zeichnung. Was sie zurücknehmen
// kann, ist der abgelegte Stand, und das kann sie längst.

import { useCallback, useState } from 'react'

import type { SymbolteilDTO } from '../antworten'

// Verlaufsschritt hält den Stand **vor** einem Handgriff.
//
// **Vorher und nicht nachher**: eine Rücknahme stellt her, was war, und der
// heutige Stand steht ohnehin da. MSE hält es genauso — `oldValues` in
// `ControlPointMoveAction`.
interface Verlaufsschritt {
  // marke ist die Kennung des Handgriffs. Gleiche Marke, gleicher Handgriff.
  marke: string

  // name ist, was in der Beschriftung steht — MSEs `getName(to_undo)`.
  name: string | undefined

  stand: SymbolteilDTO[]
}

// Stapel ist der ganze Verlauf in **einem** Zustand — MSEs `ActionStack`.
//
// **Und das ist keine Bequemlichkeit, sondern eine Berichtigung.** Der erste
// Anlauf hielt vier Zustände nebeneinander und rief `setzeZurueck` aus dem
// Updater von `setzeTeile`. Ein Updater muss **rein** sein: React ruft ihn
// auf, wann es will, und in der Entwicklungsfassung zweimal.
//
// Teurer war die zweite Folge: `zuletztEingetragen` wurde synchron gesetzt
// und im aufgeschobenen Updater gelesen — beim Rendern stand dort längst
// `true`, und MSEs `last_was_add` war damit **wirkungslos**. Sichtbar wurde
// das erst, als die Gegenprobe dazu ihren Fall wirklich enthielt: sie blieb
// vorher grün, weil der Stapel im geprüften Ablauf leer war (§10).
interface Stapel {
  // teile ist `undefined`, solange niemand etwas geändert hat — dann gilt,
  // was geladen wurde.
  //
  // **Nicht `[]` als Anfangswert.** Dann ließe sich „noch nichts geändert"
  // nicht von „alles gelöscht" unterscheiden, und wer das letzte Teil einer
  // geladenen Zeichnung löscht, bekäme sie zurück.
  teile: SymbolteilDTO[] | undefined

  zurueck: Verlaufsschritt[]
  vorwaerts: Verlaufsschritt[]

  // zuletztEingetragen ist MSEs `last_was_add` (`action_stack.cpp` Zeile 30).
  zuletztEingetragen: boolean
}

const leererStapel: Stapel = {
  teile: undefined,
  zurueck: [],
  vorwaerts: [],
  zuletztEingetragen: false,
}

// Handgriff beschreibt **eine** Handlung — MSEs `Action`.
export interface Handgriff {
  // marke ist die Identität. Gleiche Marke, dieselbe Handlung.
  //
  // Für einen Zug ist sie über den ganzen Zug dieselbe und wechselt beim
  // nächsten Aufsetzen; das ist die Lebensdauer der `ExtendableAction`.
  marke: string

  // name ist, was in der Beschriftung steht — MSEs `getName(to_undo)`.
  name?: string

  // fortschreibbar sagt, ob dieser Handgriff den obersten Schritt
  // **fortschreibt**, statt einen neuen anzulegen — `ExtendableAction`.
  //
  // **Ohne Angabe: nein.** Das ist die sichere Richtung: ein Schritt zu viel
  // ist eine Unbequemlichkeit, ein Schritt zu wenig ist verlorene Arbeit.
  fortschreibbar?: boolean
}

export interface Zeichenverlauf {
  // teile ist der Stand, der gezeigt wird.
  teile: SymbolteilDTO[]

  // eintragen übernimmt einen Handgriff.
  eintragen: (neue: SymbolteilDTO[], handgriff: Handgriff) => void

  zuruecknehmen: () => void
  wiederherstellen: () => void

  kannZuruecknehmen: boolean
  kannWiederherstellen: boolean

  // naechsteRuecknahme nennt den Handgriff, den ⌘Z träfe.
  naechsteRuecknahme: string | undefined
  naechsteWiederherstellung: string | undefined
}

export function useZeichenverlauf(geladen: SymbolteilDTO[]): Zeichenverlauf {
  const [stapel, setzeStapel] = useState<Stapel>(leererStapel)

  const gezeigte = stapel.teile ?? geladen

  const eintragen = useCallback(
    (neue: SymbolteilDTO[], handgriff: Handgriff) => {
      const { fortschreibbar = false, marke, name } = handgriff

      setzeStapel((bisher) => {
        const vorher = bisher.teile ?? geladen
        const oberster = bisher.zurueck.at(-1)

        // **Fortschreiben statt anlegen** — `ExtendableAction`. Der Stand
        // *vor* dem Zug bleibt der, der beim Aufsetzen dastand.
        const fortschreiben =
          fortschreibbar &&
          bisher.zuletztEingetragen &&
          oberster?.marke === marke

        return {
          teile: neue,
          zurueck: fortschreiben
            ? bisher.zurueck
            : [...bisher.zurueck, { marke, name, stand: vorher }],

          // **Etwas Neues löscht die Wiederherstellung** —
          // `redo_actions.clear()`, `action_stack.cpp` Zeile 27.
          vorwaerts: [],
          zuletztEingetragen: true,
        }
      })
    },
    [geladen],
  )

  // bewegen ist Rücknahme und Wiederherstellung in einer Form.
  //
  // **Weil sie in MSE dieselbe sind**: `undo()` und `redo()`
  // (`action_stack.cpp` Zeilen 41–60) unterscheiden sich in nichts als der
  // Richtung, aus der genommen und in die gelegt wird. Zweimal geschrieben
  // liefen sie auseinander — und die zweite Fassung wäre die, die niemand
  // prüft.
  const bewegen = useCallback(
    (rueckwaerts: boolean) => {
      setzeStapel((bisher) => {
        const quelle = rueckwaerts ? bisher.zurueck : bisher.vorwaerts
        const oberster = quelle.at(-1)

        if (oberster === undefined) {
          return bisher
        }

        /* v8 ignore next 4 -- unerreichbar: `teile` ist genau dann
           `undefined`, wenn noch nichts eingetragen wurde — dann sind beide
           Stapel leer, und die Prüfung darüber ist schon zurückgekehrt. Der
           Typ verlangt den Rückfall trotzdem. */
        const gegen = [
          ...(rueckwaerts ? bisher.vorwaerts : bisher.zurueck),
          { ...oberster, stand: bisher.teile ?? geladen },
        ]

        return {
          teile: oberster.stand,
          zurueck: rueckwaerts ? quelle.slice(0, -1) : gegen,
          vorwaerts: rueckwaerts ? gegen : quelle.slice(0, -1),

          // **Nach einer Bewegung wird nicht verschmolzen** — `last_was_add`
          // steht in MSE nach `undo()` wie nach `redo()` auf `false`.
          zuletztEingetragen: false,
        }
      })
    },
    [geladen],
  )

  const zuruecknehmen = useCallback(() => {
    bewegen(true)
  }, [bewegen])

  const wiederherstellen = useCallback(() => {
    bewegen(false)
  }, [bewegen])

  return {
    teile: gezeigte,
    eintragen,
    zuruecknehmen,
    wiederherstellen,
    kannZuruecknehmen: stapel.zurueck.length > 0,
    kannWiederherstellen: stapel.vorwaerts.length > 0,
    naechsteRuecknahme: stapel.zurueck.at(-1)?.name,
    naechsteWiederherstellung: stapel.vorwaerts.at(-1)?.name,
  }
}
