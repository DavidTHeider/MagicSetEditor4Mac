// **Ein Knopf aus dem Aktionskatalog** — UX-10.AK-01
// (#1682).
//
// Er trägt, was jede Aktion überall trägt: das **Symbol** vor der Aufschrift,
// die **Bedeutung** als Tooltip und — wo sein Aufrufer es verdrahtet hat — das
// **Kürzel**, zu sehen im Tooltip und zu hören über `aria-keyshortcuts`.
//
// # Die Aufschrift kommt vom Aufrufer
//
// „Deck löschen", „Karte duplizieren", „Zugang kopieren": die Sache gehört in
// die Aufschrift, wo sie mehrdeutig wäre, und nur der Aufrufer kennt sie. Ohne
// Aufschrift steht das Kurzlabel des Katalogs da. **Dass das Wort zur Aktion
// passt, hält `aktionskatalog.test.ts`** — nicht dieser Baustein.
//
// # Die Aufschrift bricht nicht um
//
// „Sichern unter …" brach in der Settabelle um, und der Knopf wurde zwei
// Zeilen hoch — gesehen in der Sichtprüfung (§3), bis dahin in `SetSichern`
// einzeln gehalten. Eine Aufschrift aus dem Katalog ist kurz; wo der Platz
// nicht reicht, bricht die Reihe um und nicht der Knopf.
//
// # Das Symbol ist stumm
//
// `aria-hidden`: ein Vorleseprogramm läse sonst „Plus, Anlegen" oder
// „Multiplikationszeichen, Löschen". Die Aufschrift trägt den Namen allein.
import type { ReactNode, RefObject } from 'react'

import { aktionsarten, type Aktion } from '../aktionen'
import { useTexte } from '../sprache/kontext'
import { Knopf, type Gewicht } from './Knopf'
import { Tooltip } from './Tooltip'

interface AktionsknopfEigenschaften {
  aktion: Aktion

  // mitKuerzel sagt, dass der Aufrufer das Kürzel des Katalogs **verdrahtet**
  // hat. Ohne das kündigt der Knopf keines an — ein angesagtes Kürzel, das
  // nichts tut, ist schlimmer als keines (§10).
  mitKuerzel?: boolean

  gewicht?: Gewicht
  klein?: boolean
  angehalten?: boolean
  bezeichnung?: string
  ohneTabstelle?: boolean
  griff?: RefObject<HTMLButtonElement | null>
  aufKlick?: () => void

  // schreibt heißt: die Handlung ändert etwas im gezeigten Profil — G10
  // (#2276). **Eine Angabe am Knopf und keine des Katalogs**: „Entfernen"
  // nimmt am Paket etwas vom Rechner und an einem Deck etwas aus dem Profil,
  // und nur das zweite verbietet die Rolle (`docs/AUTORISIERUNG.md`).
  schreibt?: boolean

  children?: ReactNode
}

export function Aktionsknopf({
  aktion,
  mitKuerzel = false,
  children,
  ...knopf
}: AktionsknopfEigenschaften) {
  const texte = useTexte().bausteine.aktionskatalog
  const art = aktionsarten[aktion]
  const eintrag = texte.aktionen[aktion]

  const hinweis = mitKuerzel
    ? `${eintrag.heisst} ${art.zeichen}`
    : eintrag.heisst

  return (
    <Tooltip text={hinweis}>
      {(griffe) => (
        <Knopf
          {...knopf}
          beschreibung={griffe}
          aktion={aktion}
          kuerzel={mitKuerzel ? art.tasten : undefined}
        >
          <span aria-hidden="true">{art.symbol}</span>
          <span className="whitespace-nowrap">{children ?? eintrag.kurz}</span>
        </Knopf>
      )}
    </Tooltip>
  )
}
