// **Was eine andere Fläche an dieser Partie schreibt, kommt an** — #2388.
//
// Ein Zug über `mse`, MCP oder REST hob den fremden Stand (`useBestandsstand`,
// #156), und der Tisch zeigte weiter seine Fassung, bis der nächste Befehl an
// ihr abprallte (§7, Regel 2). Ein eigener Haken, weil `usePartie` sonst über
// die Zeilengrenze aus §2 wüchse; **was** nachgelesen wird, bleibt dort.
//
// **Das offene Formular bleibt stehen**: dieselbe Partie ist dasselbe
// Bezugsobjekt — Regel 1 räumt beim Wechsel, nicht bei einer neuen Fassung —,
// und ein Befehl daraus rechnet danach gegen die neue. **Nicht, während ein
// eigener Befehl oder Wechsel unterwegs ist** (Regel 4): dessen Nachladen
// kommt ohnehin, und ein zweites ginge gegen einen Stand, den der Kern noch
// nicht kennt.
import { useCallback } from 'react'

import type { Partie } from '../api'
import { partievorgang } from './partiewege'
import { useFremdNachlesen } from './useFremdNachlesen'
import type { Vorgang } from './useVorgang'

export function usePartiefremd(
  fremdstand: number,
  partie: Partie | undefined,
  wechselt: boolean,
  vorgang: Vorgang,
  nachladen: (kennung: string, fuer: string) => Promise<void>,
): void {
  const nachlesen = useCallback(() => {
    if (
      partie === undefined ||
      wechselt ||
      vorgang.laeuft(partievorgang(partie.id))
    ) {
      return
    }

    void nachladen(partie.id, partie.fuer)
  }, [partie, wechselt, vorgang, nachladen])

  useFremdNachlesen(fremdstand, nachlesen)
}
