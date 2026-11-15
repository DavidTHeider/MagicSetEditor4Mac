// **Ob die ersten Schritte stehen** — Werkstück U11 von Phase 5.1
// (#1422).
//
// Bis dahin kannte die Anwendung nur „in dieser Sitzung schon geschlossen".
// Die Hilfe führt jetzt die Führungen — und wer die ersten Schritte dort
// **erneut** verlangt, bekommt sie auch dann, wenn die Einstellung sagt, sie
// sollen beim Start nicht mehr kommen: das ist eine Bitte, keine Vorgabe.
import { useCallback, useState } from 'react'

export interface Einfuehrungslage {
  // zugeklappt heißt: in dieser Sitzung geschlossen — bis jemand sie erneut
  // verlangt.
  zugeklappt: boolean

  // erneut heißt: ausdrücklich verlangt, gleich was die Einstellung sagt.
  erneut: boolean

  schliessen: () => void
  zeigen: () => void
}

// zuSchluessel merkt im Speicher der Sitzung, dass sie geschlossen wurde.
//
// **Nicht im Zustand der Seite allein** (#2428): endet der Webprozess, lädt
// Wails die Seite neu, und die Einführung stand wieder da. Der Speicher der
// Sitzung gehört dem Fenster und überlebt das; ein neuer Start der Anwendung
// beginnt ihn leer. **Ist er nicht zu lesen**, gilt der Zustand der Seite.
const zuSchluessel = 'mse:einfuehrung-zu'

function gemerkt(): boolean {
  try {
    return sessionStorage.getItem(zuSchluessel) === '1'
  } catch {
    return false
  }
}

function merken(zu: boolean) {
  try {
    if (zu) {
      sessionStorage.setItem(zuSchluessel, '1')
    } else {
      sessionStorage.removeItem(zuSchluessel)
    }
  } catch {
    // Ohne Speicher bleibt es beim Zustand der Seite.
  }
}

export function useEinfuehrung(): Einfuehrungslage {
  const [lage, setzeLage] = useState(() => ({
    zugeklappt: gemerkt(),
    erneut: false,
  }))

  const schliessen = useCallback(() => {
    merken(true)
    setzeLage({ zugeklappt: true, erneut: false })
  }, [])

  const zeigen = useCallback(() => {
    merken(false)
    setzeLage({ zugeklappt: false, erneut: true })
  }, [])

  return { ...lage, schliessen, zeigen }
}
