// **Die Parameterhilfen eines Schlüsselworts** —
// #1342.
//
// **Ein eigener Haken**, weil die Hilfen an **zwei** Dingen hängen: am Set und
// am Muster des Schlüsselworts. Wer den Parameter einfügt, ändert das Muster —
// und die Verweise darunter müssen dann andere sein.
import { useCallback } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useLaden } from './useLaden'
import type { Wahleintrag } from '../seiten/Parameterwahl'

export interface Parameterhilfengriff {
  arten: readonly Wahleintrag[]
  verweise: readonly Wahleintrag[]
  befunde: readonly string[]
}

export function useParameterhilfen(
  setID: string,
  wortID: string,
  // stand steigt bei jeder Änderung am Schlüsselwort — dann werden die
  // Verweise neu gerechnet (§7, Regel 2).
  stand: number,
): Parameterhilfengriff {
  // **Nur für ein eigenes Schlüsselwort**: die des Spiels stehen seit UXR-07
  // (#1915) als Text und rufen diesen Haken nicht. Hier stand ein Wächter
  // gegen eine leere Wortkennung — er schützte vor einem Ruf, den es nicht
  // mehr gibt.
  const holen = useCallback(async () => {
    const antwort = await api().Parameterhilfen(setID, wortID)

    return auspacken(antwort, () => antwort)
  }, [setID, wortID])

  const hilfen = useLaden(holen, [setID, wortID], false, stand)

  return {
    arten: (hilfen.daten?.arten ?? []).map((art) => ({
      schluessel: art.name,
      aufschrift:
        art.beschreibung === ''
          ? art.name
          : `${art.name} — ${art.beschreibung}`,
      einfuegen: art.einfuegen,
    })),

    verweise: (hilfen.daten?.verweise ?? []).map((verweis, stelle) => ({
      // **Der Schlüssel ist die Stelle**, nicht der Name: dieselbe
      // Parameterart darf im Muster zweimal stehen, und ein Verweisskript
      // heißt dann zweimal gleich.
      schluessel: String(stelle),
      aufschrift: verweisaufschrift(verweis),
      einfuegen: verweis.einfuegen,
    })),

    befunde: hilfen.daten?.befunde ?? [],
  }
}

// verweisaufschrift baut die Zeile eines Verweises — MSEs `1. <name>`.
//
//	String item = String() << ++param << ". " << LEFT_ANGLE_BRACKET
//	                       << p->name << RIGHT_ANGLE_BRACKET;
//
// `src/gui/set/keywords_panel.cpp`, Zeile 204, Fassung f9d9356 vom
// 2021-07-06. **Die spitzen Klammern sind MSEs eigene** — sie zeigen an, dass
// es sich um einen Platzhalter handelt und nicht um Text.
function verweisaufschrift(verweis: {
  nummer: number
  art: string
  name: string
}): string {
  const kopf = `${String(verweis.nummer)}. ⟨${verweis.art}⟩`

  return verweis.name === '' ? kopf : `${kopf} — ${verweis.name}`
}
