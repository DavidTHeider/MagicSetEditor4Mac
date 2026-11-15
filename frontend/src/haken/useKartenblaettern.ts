// **⇞ und ⇟ blättern die Karten** —
// #1290.
//
// # ported (§15)
//
// MSE2 hängt *Select Previous Card* und *Select Next Card* mit PgUp und PgDn
// ins Kartenmenü (`data/en.mse-locale/locale`, Zeilen 48–49;
// `src/gui/set/cards_panel.cpp`, Zeilen 68–69, Fassung f9d9356 vom
// 2021-07-06).
//
// **Sie wirken auch während des Tippens**, und das ist nachgelesen und nicht
// angenommen: der Karteneditor reicht jede Taste weiter, die er nicht selbst
// braucht (`src/gui/control/card_editor.cpp`, Zeilen 458–461), und der
// Texteditor beansprucht ←, →, ↑, ↓, Pos1, Ende, Rück, Entf und Eingabe —
// nicht ⇞ und ⇟ (`src/gui/value/text.cpp`, Zeilen 412–475).
//
// # Eine Abweichung, und sie gehört genannt (§15, Regel 7)
//
// **Mit geöffneter Auswahlliste blättert MSE2 nicht.** `text.cpp`, Zeilen
// 405–407, reicht dann an `DropDownList::onCharInParent` weiter — und das gibt
// im geöffneten Zustand für **jede** Taste `true`
// (`src/gui/drop_down_list.cpp`, Zeilen 449–498, Fassung f9d9356 vom
// 2021-07-06). Hier blättert es weiter.
//
// **Nicht nachgebaut**, und der Grund gehört dazu: dafür bräuchte dieser Haken
// ein Zeichen, dass irgendwo im Fenster eine Liste offen steht — ein globaler
// Zustand für einen seltenen Fall. Angestrichen vom Zweitblick über Phase 4.3,
// Befund 7.
//
// Deshalb steht hier **kein** `inEingabe`-Wächter, obwohl er neben jedem
// anderen Kürzel dieser Anwendung steht. Er wäre bequem und falsch: der Weg,
// für den es dieses Werkstück gibt, heißt „anlegen, füllen, weiter", und wer
// zum Weiterblättern erst das Feld verlassen muss, hat ihn nicht.
//
// # Warum am Fenster und nicht an einem Kasten
//
// Weil der Fokus überall stehen darf — in einem Feld, auf einem Knopf, in der
// Liste. Ein Horcher an einem `div` bekäme die Taste nur, wenn der Fokus
// darin liegt, und genau dann bräuchte man ihn nicht.
//
// # Was er **nicht** tut
//
// **Er hält keine Zusatztaste ab, er hört sie einfach nicht.** ⌘⇟ ist bei
// macOS die nächste Lasche, ⌥⇟ blättert in manchen Feldern; wer sie hier
// abfinge, nähme dem Betriebssystem eine Geste ab.

import { useEffect, useState } from 'react'

import type { KartenzeileDTO } from '../api'
import { nachbarkarte, type Richtung } from '../seiten/kartennachbar'
import { useFeldfokus } from './useFeldfokus'
import { offenesFeld, wegschreiben } from '../feldwegschreiben'
import { useVerdeckt } from '../verdeckung'

// imUeberlagerten sagt, ob eine Taste in einem Dialog oder offenen Menü fällt.
function imUeberlagerten(ziel: EventTarget | null): boolean {
  return (
    ziel instanceof Element &&
    ziel.closest('[role="dialog"], [role="menu"]') !== null
  )
}

// richtungen ordnet den beiden Tasten ihre Richtung zu.
const richtungen: Record<string, Richtung> = {
  PageUp: 'zurueck',
  PageDown: 'vor',
}

export function useKartenblaettern(
  karten: readonly KartenzeileDTO[],
  gewaehlt: string | undefined,
  waehlen: (karteID: string) => void,
): void {
  // **Der Fokus reist mit** (siehe [useFeldfokus]). Wer im Feld `name` steht
  // und ⇟ drückt, steht danach im Feld `name` der nächsten Karte — sonst
  // endete „anlegen, füllen, weiter" nach dem ersten *weiter* an der Maus.
  const [fokuswunsch, wuenschen] = useState<string>()

  useFeldfokus(fokuswunsch, wuenschen)

  // **Unter einer Schublade hört er nicht** (UX-08, #1680) — warum, steht in
  // `verdeckung.ts`.
  const verdeckt = useVerdeckt()

  useEffect(() => {
    if (verdeckt) {
      return undefined
    }

    function aufTaste(ereignis: KeyboardEvent) {
      // **Eine verbrauchte Taste, ein Dialog und ein offenes Menü** gehören
      // nicht der Karte — die Palette ist ein Dialog, und die Verdeckung
      // kennt nur Schubladen (Zweitblick über Phase 6 von 0.9.0, Befund 1).
      if (
        ereignis.defaultPrevented ||
        imUeberlagerten(ereignis.target) ||
        ereignis.metaKey ||
        ereignis.altKey ||
        ereignis.ctrlKey ||
        ereignis.shiftKey
      ) {
        return
      }

      const richtung = richtungen[ereignis.key]
      if (richtung === undefined) {
        return
      }

      const ziel = nachbarkarte(karten, gewaehlt, richtung)
      if (ziel === undefined) {
        // **Am Rand geschieht nichts, und die Taste bleibt trotzdem
        // verbraucht** — sonst scrollte die Seite, und das sähe aus wie ein
        // halber Sprung.
        ereignis.preventDefault()

        return
      }

      ereignis.preventDefault()

      // **Gemerkt wird vor dem Wechsel**, denn danach gibt es das Feld nicht
      // mehr: der Editor trägt die Kartenkennung in seinem `key`. Und
      // **verlassen** wird es auch — warum, steht bei [wegschreiben].
      wuenschen(offenesFeld())
      wegschreiben()

      waehlen(ziel)
    }

    window.addEventListener('keydown', aufTaste)

    return () => {
      window.removeEventListener('keydown', aufTaste)
    }
  }, [karten, gewaehlt, waehlen, verdeckt])
}
