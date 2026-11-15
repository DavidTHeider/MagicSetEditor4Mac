// **Ein Kasten für Hinweis, Erfolg und Befund** — Werkstück U3 von Phase 5.1
// (#1414).
//
// # Wozu
//
// Über hundert Stellen im Baum schreiben denselben Absatz von Hand:
// `<p role="alert" className="text-sm text-red-800 dark:text-red-300">`. Jede
// Kopie trägt ihre eigene Entscheidung darüber, ob sie `alert` oder `status`
// ist — und **das ist keine Geschmacksfrage**: `alert` unterbricht ein
// Vorleseprogramm, `status` wartet. Wer eine gute Nachricht als Alarm ausgibt,
// stört; wer einen Fehler als Zustand ausgibt, wird überhört.
//
// # Warum die Farben hier stehen und nicht neu gewählt werden
//
// Die drei Paare sind in [`docs/KONTRASTE.md`](../../../docs/KONTRASTE.md)
// **gemessen** — `text-red-800`/`text-red-300` mit 8,36 und 9,29,
// `text-emerald-800`/`text-emerald-200` mit 7,61 und 13,94, `text-leise` mit
// 5,85. Eine vierte Farbe braucht eine Messung, keine Meinung (§3).
//
// # Was er **nicht** tut
//
// Er räumt nichts weg und merkt sich nichts. Wann eine Meldung verschwindet,
// weiß die Seite — `schliessen` ist ihr Rückruf, und ohne ihn steht kein
// Knopf da: einer, der nichts schließt, wäre eine Zusage ohne Rumpf (§10).
//
// # Und warum er eine Handlung tragen kann
//
// **Weil eine Meldung sagen soll, was jetzt zu tun ist** (UX-08.AK-02,
// #1680).
// „Der Eintrag wurde zwischenzeitlich geändert. Bitte neu laden." ohne einen
// Knopf dazu schickte den Menschen suchen, **wie** man neu lädt. Dieselbe
// Regel wie beim Schließen: ohne Rückruf kein Knopf.
import { Knopf } from './Knopf'
import { useTexte } from '../sprache/kontext'

// Handlung ist der nächste Schritt, den eine Meldung anbietet.
export interface Handlung {
  beschriftung: string
  aufKlick: () => void
}

// Hinweisart entscheidet über Rolle **und** Farbe — beide folgen derselben
// Frage: wie dringend ist das hier?
export type Hinweisart = 'fehler' | 'erfolg' | 'hinweis'

interface HinweiskastenEigenschaften {
  art: Hinweisart

  meldung: string

  // kennung ist der halbe Zweck: ein Feld zeigt mit `aria-describedby`
  // darauf, und damit steht der Befund **am** Feld (§6).
  kennung?: string

  // schliessen macht den Knopf sichtbar — und nur dann.
  schliessen?: () => void

  // handlung ist der nächste Schritt — beim Konflikt „Neu laden".
  handlung?: Handlung | undefined
}

// farben ist das gemessene Paar je Art.
const farben: Record<Hinweisart, string> = {
  fehler: 'text-fehler',
  erfolg: 'text-erfolg',
  hinweis: 'text-leise',
}

const rahmen = 'flex items-start gap-2 text-sm '

const schliesser =
  'shrink-0 rounded-sm px-1 underline focus-visible:outline ' +
  'focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring'

export function Hinweiskasten({
  art,
  meldung,
  kennung,
  schliessen,
  handlung,
}: HinweiskastenEigenschaften) {
  const texte = useTexte()

  return (
    <div
      id={kennung}
      // **`alert` nur beim Fehler.** Die beiden anderen warten, bis das
      // Vorleseprogramm eine Pause hat.
      role={art === 'fehler' ? 'alert' : 'status'}
      className={rahmen + farben[art]}
    >
      <p className="flex-1">{meldung}</p>

      {handlung !== undefined && (
        <Knopf gewicht="leise" aufKlick={handlung.aufKlick}>
          {handlung.beschriftung}
        </Knopf>
      )}

      {schliessen !== undefined && (
        <button type="button" className={schliesser} onClick={schliessen}>
          {texte.bausteine.aktionen.schliessen}
        </button>
      )}
    </div>
  )
}
