// **Welchen Befehl ein losgelassener Zug stellt** — herausgelöst am
// 2026-09-22, als `Partietisch.tsx` mit der Autoanordnung (PLAY-04.AK-03,
// #1712)
// an die Zeilengrenze aus §2 kam.
//
// **Der Schnitt folgt der Frage**: nebenan steht, **was die Seite zeigt**;
// in `tischzug.ts`, **was ein Zug bedeutet**; hier, **wie er als Befehl
// heißt**.

import type { Zugziel } from '../bausteine/tischziehen'

// befehlFuer übersetzt einen losgelassenen Zug in seinen Befehl.
//
// **Eine Liste und kein einzelner Rückgabewert**: so bleibt sie eine reine
// Funktion, und der Aufrufer stellt, was darin steht. Ein `switch`, der
// selbst rief, wäre ohne einen Baum nicht zu prüfen.
//
// **Drei Arten, drei Befehle** — und alle drei gehen denselben Weg wie die
// Handlungsleiste (§16).
export function befehlFuer(ziel: Zugziel): [string, Record<string, unknown>][] {
  switch (ziel.art) {
    // **Eine Geste deutet der Kern** — P5 (#2373) über P3 (#2371): ein Land
    // aus der Hand wird gespielt, ein Zauber gewirkt, im Sandkasten
    // geschoben. **Der Ziel-Sitz geht mit**; bis dahin fiel er weg (C-03).
    case 'zone':
      return [
        [
          'objekt.ziehen',
          {
            objekt: ziel.kennung,
            zone: ziel.zone,
            sitz: ziel.sitz,
            ...(ziel.frei === true ? { frei: true } : {}),
          },
        ],
      ]

    case 'haufen':
      return [['objekt.haeufen', { objekt: ziel.kennung, worauf: ziel.worauf }]]
  }

  return [['objekt.enthaeufen', { objekt: ziel.kennung }]]
}
