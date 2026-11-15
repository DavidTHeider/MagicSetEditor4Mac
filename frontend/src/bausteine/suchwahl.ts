// **Was eine Auswahl mit Suche weiß, ohne etwas zu zeichnen** — UXR-F1
// (#2007).
//
// Die Zeile, die Suche und die zwei Zahlen, an denen R7 aus
// `docs/UIREVIEW.md` hängt. Sie stehen hier und nicht in `Suchauswahl.tsx`:
// diese Datei zeichnet, und die Zahlen brauchen auch die, die nur
// **entscheiden**, ob eine Suche dazugehört (`Auswahlfeld`, die Kartenfelder).

import { useRef } from 'react'

import { useFokusUeberAnhalten } from '../haken/useFokusUeberAnhalten'

// Suchwahl ist eine Zeile der Auswahl.
export interface Suchwahl {
  // wert ist, was die Auswahl meldet. Leer ist erlaubt: „nichts" als Wahl.
  wert: string

  // name steht in der Liste und, wenn die Zeile gilt, im Feld.
  name: string

  // zusatz steht blass hinter dem Namen und wird mitgesucht — die Kennung
  // eines Stils, sein fremdes Spiel.
  zusatz?: string
}

// suchschwelle ist, ab wie vielen Werten eine Auswahl eine Suche trägt: R7,
// „Auswahl mit Suche ab 15 Werten". Darunter bleibt es ein natives Feld, das
// Menü von macOS.
export const suchschwelle = 15

// zeilengrenze ist, wie viele Zeilen die Liste höchstens zeichnet. Ein Set
// trägt tausende Karten, und jede Zeile ist ein Knoten; wer mehr sehen will,
// grenzt die Suche ein.
export const zeilengrenze = 200

// vereinfacht bringt einen Text auf das, womit gesucht wird: klein, mit einem
// Leerzeichen statt `_`, `-` und `/`, ohne Rand.
function vereinfacht(text: string): string {
  return text
    .toLowerCase()
    .replace(/[\s_/-]+/g, ' ')
    .trim()
}

// passtZu sagt, ob eine Zeile zu dem passt, was jemand tippt — über Name und
// Zusatz, ohne Rücksicht auf Groß und klein. Leer passt alles.
export function passtZu(eintrag: Suchwahl, suche: string): boolean {
  return vereinfacht(`${eintrag.name} ${eintrag.zusatz ?? ''}`).includes(
    vereinfacht(suche),
  )
}

// suchtreffer sucht in **allen** Zeilen und gibt davon höchstens die
// `zeilengrenze` zum Zeichnen zurück — mit der Zahl aller Treffer, damit die
// Auswahl sagen kann, dass es mehr gibt.
//
// **Ohne Suche kein Filtern**: vierhundert Zeilen je Zeichnen zu vereinfachen,
// solange niemand tippt, wäre Arbeit für nichts.
export function suchtreffer(
  eintraege: readonly Suchwahl[],
  suche: string,
): { gefunden: number; gezeigt: readonly Suchwahl[] } {
  const passend =
    suche.trim() === ''
      ? eintraege
      : eintraege.filter((eintrag) => passtZu(eintrag, suche))

  return {
    gefunden: passend.length,
    gezeigt: passend.slice(0, zeilengrenze),
  }
}

// zeigen rollt die hervorgehobene Zeile in den sichtbaren Bereich der Liste.
//
// **Als Rückruf an der Zeile und nicht als Effekt**: er läuft, sobald die
// Zeile im Baum steht — auch beim Öffnen, wenn die Liste im selben Zug
// entsteht.
export function zeigen(zeile: HTMLElement | null) {
  zeile?.scrollIntoView({ block: 'nearest' })
}

// Zeigerstelle ist, wo ein Mausereignis stand.
interface Zeigerstelle {
  clientX: number
  clientY: number
}

// useZeigerbewegung unterscheidet einen Zeiger, der sich bewegt, von einem,
// der stillsteht — **gemessen am Fenster** (`uxrf1`, #2007): das Öffnen rollt
// die Liste, WebKit schickt danach ein `mousemove` an **derselben** Stelle,
// und die Zeile unter dem ruhenden Zeiger galt als hervorgehoben statt der
// gewählten. `merke` hält die Stelle eines Klicks fest, `bewegt` sagt, ob ein
// Ereignis woanders steht als das vorige.
export function useZeigerbewegung() {
  const zuletzt = useRef<Zeigerstelle | null>(null)

  return {
    merke(stelle: Zeigerstelle) {
      zuletzt.current = { clientX: stelle.clientX, clientY: stelle.clientY }
    },
    bewegt(stelle: Zeigerstelle): boolean {
      const vorher = zuletzt.current
      zuletzt.current = { clientX: stelle.clientX, clientY: stelle.clientY }

      return (
        vorher?.clientX !== stelle.clientX || vorher.clientY !== stelle.clientY
      )
    },
  }
}

// useSuchfeld ist, was das Feld einer Auswahl mit Suche am Fenster braucht
// (`uxrf1`, #2007): **den Fokus über das Schreiben hinweg** — gesperrt nimmt
// WebKit ihn, und danach ging ein Pfeil ins Leere (#1990) — und **einen
// Zeiger, der stillstehen darf** ([useZeigerbewegung]).
export function useSuchfeld(angehalten: boolean) {
  const feld = useRef<HTMLInputElement>(null)

  return {
    feld,
    merkeFokus: useFokusUeberAnhalten(feld, angehalten),
    zeiger: useZeigerbewegung(),
  }
}
