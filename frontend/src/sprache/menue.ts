// Die Naht zur nativen Menüleiste.
//
// Sie liegt in Go und weiß von React nichts. Ohne diesen Aufruf bliebe die
// Leiste bei der Sprache, mit der die Anwendung gestartet ist — eine Anwendung
// in zwei Sprachen.
import { api, auspacken } from '../api'
import { type Sprache } from './kataloge'

// vorige ist der letzte Ruf, der noch unterwegs sein kann.
let vorige: Promise<unknown> = Promise.resolve()

// spracheMelden meldet Go die Sprache — **der Reihe nach**.
//
// **Jeder Ruf wartet auf den vorigen** (Befund 3 des Nachlaufs über Phase
// 6.2). `useLauf` verwirft im Fenster die verspätete Antwort, aber Go setzt
// die Leiste bei jedem Ruf, der ankommt: zwei schnelle Wechsel konnten dort
// in umgekehrter Reihenfolge ankommen, und die Leiste stand zuletzt in der
// älteren Sprache. **Ein gescheiterter Ruf hält den nächsten nicht auf** —
// sein Fehler gehört dem, der ihn gerufen hat.
export function spracheMelden(sprache: Sprache): Promise<void> {
  const dieser = vorige.then(() => api().SpracheSetzen(sprache))

  vorige = dieser.catch(() => undefined)

  return dieser.then((antwort) => {
    auspacken(antwort, () => undefined)
  })
}
