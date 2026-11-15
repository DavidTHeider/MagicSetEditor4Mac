// **Der Anstoß für ein anderes Objekt** — AP-5
// (#2362),
// Befund B-01.
//
// Eine Liste trägt **ein** Kontextmenü, das des gewählten Objekts: gezeichnet
// wird nur ein Sichtfenster, und je Zeile ein Menü wären Hunderte. Ein
// Rechtsklick oder ⇧F10 auf einer anderen Zeile wählt sie deshalb zuerst —
// wie der Finder — und das Menü geht auf, sobald **sie** die gewählte ist.
// Nie mit den Einträgen der vorigen.
//
// **Kommt stattdessen ein drittes Objekt**, verfällt der Anstoß: über das
// Menü oder die Tasten gewählt, ginge er sonst später an einer Stelle auf, an
// der niemand mehr geklickt hat. **Und ebenso mit der nächsten Eingabe**: hält
// die Seite die Wahl an — die Werkstatt fragt bei Ungesichertem nach —, ginge
// das Menü sonst nach der Antwort auf.
import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type KeyboardEvent,
  type MouseEvent,
} from 'react'

import {
  imFeld,
  oeffnetKontextmenue,
  type Anstoss,
} from '../bausteine/Kontextmenue'

type Ereignis = MouseEvent<HTMLElement> | KeyboardEvent<HTMLElement>

// anstossstelle sagt, wo ein Ereignis das Kontextmenü haben will — beim
// Rechtsklick am Zeiger, bei ⇧F10 und der Menütaste unter der Zeile. Ein
// Rechtsklick in einem Feld gehört macOS, wie in `Kontextmenue`.
export function anstossstelle(
  ereignis: Ereignis,
): { x: number; y: number } | undefined {
  if (ereignis.type === 'contextmenu') {
    const zeiger = ereignis as MouseEvent<HTMLElement>

    return imFeld(zeiger.target)
      ? undefined
      : { x: zeiger.clientX, y: zeiger.clientY }
  }

  const taste = ereignis as KeyboardEvent<HTMLElement>
  if (!oeffnetKontextmenue(taste.key, taste.shiftKey)) {
    return undefined
  }

  const zeile = taste.currentTarget.getBoundingClientRect()

  return { x: zeile.left, y: zeile.bottom }
}

// Ziel ist, was `waehlen` braucht — eine Kennung, ein Paket, eine Datei. Es
// kommt mit dem Ereignis und wird nicht über die Kennung nachgeschlagen.
export function useAnstoss<Ziel>(
  gewaehlt: string | undefined,
  waehlen: (ziel: Ziel) => void,
) {
  const [anstoss, setzeAnstoss] = useState<
    Anstoss & { ziel: string; von: string | undefined }
  >()
  // **Nichts gewählt ist kein drittes Objekt**, sondern der Übergang: ein
  // anderes Deck räumt zuerst den offenen Kontext (`useDecks`).
  if (
    anstoss !== undefined &&
    gewaehlt !== undefined &&
    gewaehlt !== anstoss.ziel &&
    gewaehlt !== anstoss.von
  ) {
    setzeAnstoss(undefined)
  }

  const nummer = useRef(0)
  const verbraucht = useCallback(() => {
    setzeAnstoss(undefined)
  }, [])

  const wartet = anstoss !== undefined
  useEffect(() => {
    if (!wartet) {
      return
    }

    document.addEventListener('pointerdown', verbraucht, true)
    document.addEventListener('keydown', verbraucht, true)

    return () => {
      document.removeEventListener('pointerdown', verbraucht, true)
      document.removeEventListener('keydown', verbraucht, true)
    }
  }, [wartet, verbraucht])

  // anstossen nimmt ein Ereignis an einer Zeile entgegen — **auch an der
  // gewählten**: dann geht das Menü gleich auf. Die Liste öffnet ihr Menü nur
  // so (`nurAngestossen` an `Kontextmenue`), damit ein Rechtsklick neben die
  // Zeilen nicht das Menü des gewählten Objekts zeigt.
  const anstossen = (ereignis: Ereignis, kennung: string, ziel: Ziel) => {
    const stelle = anstossstelle(ereignis)
    if (stelle === undefined) {
      return
    }

    ereignis.preventDefault()
    ereignis.stopPropagation()
    nummer.current += 1
    // **Die Zeile und die Stelle in ihr** — geht das Menü erst auf, wenn
    // sie woanders steht, folgt es ihr (Befund 5 des Zweitblicks über
    // Phase 6 von 0.9.0).
    const zeile = ereignis.currentTarget.getBoundingClientRect()
    setzeAnstoss({
      ...stelle,
      nummer: nummer.current,
      anker: ereignis.currentTarget,
      versatz: { x: stelle.x - zeile.left, y: stelle.y - zeile.top },
      ziel: kennung,
      von: gewaehlt,
    })
    if (kennung !== gewaehlt) {
      waehlen(ziel)
    }
  }

  return {
    anstoss: anstoss?.ziel === gewaehlt ? anstoss : undefined,
    verbraucht,
    anstossen,
  }
}
