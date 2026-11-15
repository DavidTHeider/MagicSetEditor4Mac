// **Der Raum, in dem sich das gewählte Objekt meldet** — GESTE
// (#2358).
// Was gemeldet wird und warum, steht in `haken/useHandlungsliste.ts`; hier
// steht, wie es nach Go kommt und wie ein Klick zurückfindet.
//
// # Das jüngste Objekt hat die Leiste
//
// **Angemeldet wird, wer gezeichnet ist** — und eine verborgene Seite ist es
// nicht: `Activity` beendet ihre Effekte, und damit meldet sie sich ab. Bleiben
// zwei, gilt das zuletzt angemeldete; geht es, bekommt das vorige die Leiste
// zurück. Eine Aktualisierung behält ihren Platz, sonst rückte jede Zeile, die
// neu zeichnet, nach vorn.
//
// # Wann gesendet wird
//
// Wie im `Darstellungsstandraum`: **beim Einhängen immer**, auch „nichts
// gewählt" — der Raum entsteht neu, wenn die Sitzung aufgeht oder das Profil
// wechselt, und Go hat dann noch die Leiste von vorher —, **danach nur bei
// einer Änderung**, und **der Reihe nach**: zwei schnelle Wechsel dürfen in Go
// nicht in umgekehrter Folge ankommen. Ein gescheiterter Ruf wird nicht
// gemeldet; die Leiste ist Bestenfalls-Arbeit (§7, Regel 3).
//
// # Ein Klick zurück
//
// **Nur für das Objekt, das die Leiste hat, und nur in seiner Epoche.** Ein
// Klick, der nach einem Wechsel ankommt — eine andere Karte, ein anderer
// Eintrag —, bewirkt nichts (§7, Regel 2).
import {
  createContext,
  useCallback,
  useEffect,
  useRef,
  useState,
  type ReactNode,
} from 'react'

import type { Menueeintrag } from './bausteine/Menue'
import { api } from './fassade'
import { ereignisObjekthandlung, horche } from './fenster'
import {
  keinObjekt,
  objektkuerzel,
  type Objektausloesung,
  type Objekthandlungen,
} from './objekthandlungen'
import { auspacken } from './umschlag'

// Anmeldung ist, was ein Objekt dem Raum gibt: die Meldung für Go, die
// Einträge, wie sie zuletzt gezeichnet waren — mit Sperre und Grund für das
// Kontextmenü —, und den Weg, einen Eintrag auszulösen.
export interface Anmeldung {
  meldung: Objekthandlungen
  eintraege: readonly Menueeintrag[]
  ausloesen: (schluessel: string) => void
}

type Anmelden = (platz: string, anmeldung: Anmeldung | undefined) => void

// GewaehltesObjekt ist die Anmeldung, die die Leiste hat — für das
// `Objektkontextmenue`, das dieselbe Liste an eine andere Stelle legt.
export const GewaehltesObjekt = createContext<Anmeldung | undefined>(undefined)

// ObjekthandlungsAnmeldung ist die Naht zu `useHandlungsliste` — außerhalb
// eines Raums `undefined`, und dann meldet niemand etwas.
export const ObjekthandlungsAnmeldung = createContext<Anmelden | undefined>(
  undefined,
)

// ueberlagert nennt, was über dem Objekt steht: ein Dialog oder ein offenes
// Menü.
const ueberlagert = '[role="dialog"], [role="menu"]'

// ruhigeStelle sagt, wo eine Taste nicht dem Objekt gehört — in einem Feld,
// einem Dialog und einem offenen Menü hat ⇟ eine eigene Bedeutung.
function ruhigeStelle(ziel: EventTarget | null): boolean {
  return (
    ziel instanceof Element &&
    ziel.closest(
      `input, textarea, select, [contenteditable]:not([contenteditable="false"]), ${ueberlagert}`,
    ) !== null
  )
}

// verdeckt sagt, ob der Fokus in einem Dialog steht — dann gehört die
// Leiste nicht dem Objekt dahinter. **Ein offenes Menü verdeckt nicht**: es
// zeigt die Handlungen desselben Objekts, und ein Wechsel über die Leiste
// schließt es (AP-5, `ap5kontext09`) — gesperrt, wäre „Nächste Karte“ grau,
// solange es offen ist.
function verdeckt(): boolean {
  return document.activeElement?.closest('[role="dialog"]') != null
}

// gesperrtWeilVerdeckt meldet jeden Eintrag gesperrt.
function gesperrtWeilVerdeckt(meldung: Objekthandlungen): Objekthandlungen {
  return {
    ...meldung,
    eintraege: meldung.eintraege.map((eintrag) => ({
      ...eintrag,
      gesperrt: true,
    })),
  }
}

// tastenschrift schreibt einen Tastendruck mit ⌘ wie `aria-keyshortcuts` —
// `Meta+D`, `Meta+Backspace`.
function tastenschrift(ereignis: KeyboardEvent): string {
  const taste =
    ereignis.key.length === 1 ? ereignis.key.toUpperCase() : ereignis.key

  return `Meta+${taste}`
}

// istAusloesung sagt, ob eine Nutzlast die Gestalt eines Klicks hat.
function istAusloesung(wert: unknown): wert is Objektausloesung {
  const klick = wert as Partial<Objektausloesung> | null
  return (
    typeof klick === 'object' &&
    klick !== null &&
    typeof klick.art === 'string' &&
    typeof klick.kennung === 'string' &&
    typeof klick.epoche === 'number' &&
    typeof klick.schluessel === 'string'
  )
}

export function Objekthandlungsraum({ children }: { children: ReactNode }) {
  const [plaetze, setzePlaetze] = useState<
    readonly (readonly [string, Anmeldung])[]
  >([])

  const anmelden = useCallback<Anmelden>((platz, anmeldung) => {
    setzePlaetze((vorher) => {
      const stelle = vorher.findIndex(([belegt]) => belegt === platz)
      if (anmeldung === undefined) {
        return vorher.filter(([belegt]) => belegt !== platz)
      }
      if (stelle < 0) {
        return [...vorher, [platz, anmeldung]]
      }

      return vorher.map((eintrag, bei) =>
        bei === stelle ? ([platz, anmeldung] as const) : eintrag,
      )
    })
  }, [])

  const oben = plaetze.at(-1)?.[1]

  // **Solange ein Dialog den Fokus hat, ist die Leiste gesperrt**
  // — Rest R1 des Nachlaufs über Phase 6 von 0.9.0: ⌘D aus einem Feld im
  // Dialog ging am Raum vorbei an die Leiste, und sie verdoppelte die Karte
  // dahinter. Wie ein Blatt unter macOS. Gelesen wird bei jedem Fokuswechsel
  // und, solange etwas verdeckt, bei jeder Entfernung: ein Dialog, der samt
  // Fokus verschwindet, meldet kein `focusout`.
  const [zugedeckt, setzeZugedeckt] = useState(false)
  useEffect(() => {
    let beobachter: MutationObserver | undefined

    const lesen = () => {
      const jetzt = verdeckt()
      setzeZugedeckt(jetzt)

      if (jetzt && beobachter === undefined) {
        beobachter = new MutationObserver(lesen)
        beobachter.observe(document.body, { childList: true, subtree: true })
      } else if (!jetzt) {
        beobachter?.disconnect()
        beobachter = undefined
      }
    }

    document.addEventListener('focusin', lesen)
    document.addEventListener('focusout', lesen)

    return () => {
      document.removeEventListener('focusin', lesen)
      document.removeEventListener('focusout', lesen)
      beobachter?.disconnect()
    }
  }, [])

  const eigen = oben?.meldung ?? keinObjekt
  const meldung = zugedeckt ? gesperrtWeilVerdeckt(eigen) : eigen

  // **Der Klick findet das heutige Objekt**, ohne dass der Horcher bei jedem
  // Wechsel neu angemeldet wird.
  const heute = useRef(oben)
  useEffect(() => {
    heute.current = oben
  })

  useEffect(
    () =>
      horche(ereignisObjekthandlung, ([klick]) => {
        const objekt = heute.current
        if (
          !istAusloesung(klick) ||
          verdeckt() ||
          objekt?.meldung.art !== klick.art ||
          objekt.meldung.kennung !== klick.kennung ||
          objekt.meldung.epoche !== klick.epoche
        ) {
          return
        }

        objekt.ausloesen(klick.schluessel)
      }),
    [],
  )

  // **Die Kürzel mit ⌘ hört der Raum selbst** — AP-7 (#2364), B-02. Die
  // Leiste zeigt sie, kann aber ⌘⌫ nicht zustellen; warum, steht in
  // `objekthandlungstasten.test.tsx`. Ausgelöst wird wie aus der Leiste, mit
  // denselben Sperren — nur außerhalb eines Felds, und nur eine Taste, die
  // noch niemand verbraucht hat. Verbraucht der Raum sie, fragt WebKit die
  // Leiste nicht mehr: ausgelöst wird einmal. **⇞ und ⇟ nicht**: sie blättern
  // über `useKartenblaettern`, mit der Regel aus MSE2.
  useEffect(() => {
    const aufTaste = (ereignis: KeyboardEvent) => {
      const objekt = heute.current
      if (
        objekt === undefined ||
        ereignis.defaultPrevented ||
        !ereignis.metaKey ||
        ereignis.ctrlKey ||
        ereignis.altKey ||
        ereignis.shiftKey ||
        ruhigeStelle(ereignis.target)
      ) {
        return
      }

      const gedrueckt = tastenschrift(ereignis)
      const eintrag = objekt.eintraege.find(
        (kandidat) =>
          kandidat.kuerzel !== undefined &&
          objektkuerzel[kandidat.kuerzel].tasten === gedrueckt,
      )
      if (eintrag === undefined) {
        return
      }

      ereignis.preventDefault()
      objekt.ausloesen(eintrag.schluessel)
    }

    window.addEventListener('keydown', aufTaste)

    return () => {
      window.removeEventListener('keydown', aufTaste)
    }
  }, [])

  const kette = useRef<Promise<unknown>>(Promise.resolve())
  const gestalt = JSON.stringify(meldung)

  // **Nur bei neuer Gestalt** — das sagt die Abhängigkeit; ein Vergleich mit
  // dem zuletzt Gesendeten wäre derselbe zweimal.
  useEffect(() => {
    const wert = JSON.parse(gestalt) as Objekthandlungen
    kette.current = kette.current.then(async () => {
      try {
        auspacken(await api().ObjekthandlungenMelden(wert), () => undefined)
      } catch {
        // Bestenfalls-Arbeit, siehe oben: die nächste Meldung setzt die
        // Leiste richtig, und es gibt kein Feld, an dem ein Fehler stünde.
      }
    })
  }, [gestalt])

  return (
    <ObjekthandlungsAnmeldung value={anmelden}>
      <GewaehltesObjekt value={oben}>{children}</GewaehltesObjekt>
    </ObjekthandlungsAnmeldung>
  )
}
