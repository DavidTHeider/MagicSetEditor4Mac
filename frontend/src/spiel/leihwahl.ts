// **Welche Leihe sich spielen lässt** — P1b
// (#2370),
// C-12: geliehen, nicht widerrufen, nicht eingelöst, mit dem Recht `spielen`
// und an keine andere Art gebunden. Abgelaufen oder fremd weist der Kern ab,
// mit Grund; das Fenster deutet keine Leihe. **Eine Regel an zwei Stellen** —
// am eigenen Tisch und beim Beitritt an einen fremden (P9c, #2377).
import type { Leihe } from '../apileihe'

export function spielbar(leihe: Leihe): boolean {
  return (
    !leihe.verliehen &&
    !leihe.widerrufen &&
    leihe.verbraucht === '' &&
    leihe.rechte.includes('spielen') &&
    (leihe.bindungsart === '' || leihe.bindungsart === 'partie')
  )
}
