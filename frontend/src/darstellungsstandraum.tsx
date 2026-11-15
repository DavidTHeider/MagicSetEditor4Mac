// **Der Raum, in dem die Teile des Fensters sich melden** — UXR-F14
// (#2029).
// Was gemeldet wird und warum, steht in `darstellungsstand.ts`; hier steht, wie
// es nach Go kommt.
//
// # Warum ein Raum und kein Aufruf aus jedem Teil
//
// Die Seitenleiste lebt in `Anwendung`, die Spalten im `Arbeitsplatz` einer
// Seite, und **Go braucht den ganzen Stand auf einmal** — die Leiste wird aus
// allen dreien gebaut. Riefe jedes Teil für sich, wüsste Go nach jedem Ruf nur
// ein Drittel und müsste den Rest raten. Der Raum sammelt und sendet das
// Ganze, dieselbe Gestalt wie `ablage.tsx` und `ruecknahme.tsx`.
//
// # Wann gesendet wird
//
// - **Beim Einhängen immer**, auch den Anfangsstand. Der Raum entsteht neu,
//   wenn die Sitzung aufgeht (`Sitzungstor`) oder das Profil wechselt, und Go
//   hat dann noch den Stand von vorher; eine Seite ohne Arbeitsplatz meldet
//   nichts und müsste sonst bei einem Menü stehen bleiben, das einen
//   Navigator anbietet, den es nicht gibt.
// - **Danach nur bei einer Änderung.** Gleiches Wort an Go zu schicken wäre
//   Arbeit für nichts: ein Wechsel zwischen zwei Arbeitsplätzen ändert den
//   Stand nicht.
//
// # Der Reihe nach, und ohne Anspruch auf Erfolg
//
// **Jeder Ruf wartet auf den vorigen**, aus demselben Grund wie
// `sprache/menue.ts`: zwei schnelle Wechsel dürfen in Go nicht in umgekehrter
// Reihenfolge ankommen, sonst steht die Leiste im Stand von vorhin.
//
// **Ein gescheiterter Ruf wird nicht gemeldet.** Die Leiste ist
// Bestenfalls-Arbeit (§7, Regel 3): es gibt kein Feld, an dem ein Fehler
// stünde, und der nächste Stand setzt sie ohnehin richtig. Der häufige Grund
// ist keiner, der jemanden etwas angeht — gesperrt hält Go jede Bindung an,
// und hinter der Sperre gibt es diesen Raum nicht.
import { useEffect, useReducer, useRef, type ReactNode } from 'react'

import {
  Darstellungsmeldung,
  anfangsstand,
  gleicherStand,
  zusammenfuehren,
  type Darstellungsstand,
} from './darstellungsstand'
import { api } from './fassade'
import { auspacken } from './umschlag'

export function Darstellungsstandraum({ children }: { children: ReactNode }) {
  const [stand, melde] = useReducer(zusammenfuehren, anfangsstand)

  // gesendet ist der letzte Stand, der abgeschickt wurde — `undefined`, bis
  // der erste gegangen ist.
  const gesendet = useRef<Darstellungsstand>(undefined)

  // kette ist der letzte Ruf, der noch unterwegs sein kann.
  const kette = useRef<Promise<unknown>>(Promise.resolve())

  useEffect(() => {
    const vorher = gesendet.current
    if (vorher !== undefined && gleicherStand(vorher, stand)) {
      return
    }

    gesendet.current = stand

    kette.current = kette.current.then(async () => {
      try {
        auspacken(await api().DarstellungsstandMelden(stand), () => undefined)
      } catch {
        // Bestenfalls-Arbeit, siehe oben: der nächste Stand setzt die Leiste
        // richtig, und es gibt kein Feld, an dem ein Fehler stünde.
      }
    })
  }, [stand])

  return <Darstellungsmeldung value={melde}>{children}</Darstellungsmeldung>
}
