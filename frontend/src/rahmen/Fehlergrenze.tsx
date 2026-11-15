// **Ein Fehler beim Zeichnen nimmt nicht das ganze Fenster mit** — F01
// (#2340),
// Befund A-01.
//
// # Warum eine Klasse
//
// **Weil React Fehler beim Zeichnen nur in einer Klasse fängt**
// (`getDerivedStateFromError`, `componentDidCatch`; nachgelesen am
// 2026-10-13 in der Referenz zu `Component`, React 19). Sie ist die einzige
// im Fenster und bleibt klein: was sie zeigt, entscheidet ihr Aufrufer.
//
// # Zwei Ebenen
//
// Die Wurzel (`Wurzelgrenze`) fängt, was sonst das ganze Fenster leer ließe;
// jeder Bereich (`Inhalt`) fängt seinen eigenen Wurf, und Seitenleiste,
// Kopf und die übrigen Bereiche bleiben bedienbar.
//
// # Die Naht der Hüllenprobe
//
// **`globalThis.__mseProbewurf`** nennt einen Bereich, der beim nächsten
// Zeichnen wirft — sonst ist die Naht wirkungslos. Ohne sie ließe sich am
// nativen Fenster kein Renderfehler auslösen, ohne einen echten Fehler in
// den Baum zu legen. **Sie gilt, bis die Grenze den Wurf gefangen hat**: ein
// einmaliger Wurf verschwände im zweiten Versuch, den React nach einem Fehler
// beim nebenläufigen Zeichnen unternimmt.
//
// **Für die Wurzel auch `sessionStorage`** (`mse.probewurf`): die Wurzel
// zeichnet nur beim Laden, und ein Neuladen nimmt `globalThis` mit.
import { Component, type ReactNode } from 'react'

import { api } from '../fassade'
import { neueKennung } from '../kennung'

interface Probenaht {
  __mseProbewurf?: string
}

const nahtschluessel = 'mse.probewurf'

// probewurf nennt den Bereich, den die Hüllenprobe werfen lässt — oder keinen.
//
// **Ein gesperrter Speicher ist keine Probe**: wirft sein Zugriff, gilt die
// Naht als leer.
function probewurf(): string | undefined {
  const naht = (globalThis as Probenaht).__mseProbewurf
  if (naht !== undefined) {
    return naht
  }

  try {
    return sessionStorage.getItem(nahtschluessel) ?? undefined
  } catch {
    return undefined
  }
}

// probewurfGefangen räumt die Naht, wenn sie diesen Bereich nannte.
function probewurfGefangen(bereich: string) {
  if (probewurf() !== bereich) {
    return
  }

  delete (globalThis as Probenaht).__mseProbewurf

  try {
    sessionStorage.removeItem(nahtschluessel)
  } catch {
    // Ein gesperrter Speicher hielt auch nichts.
  }
}

// Fehlerstand ist, was eine Fehlerfläche zum Zeigen bekommt.
export interface Fehlerstand {
  kennung: string
  neu: () => void
}

interface Eigenschaften {
  children: ReactNode
  ebene: 'bereich' | 'wurzel'
  bereich: string
  zeige: (stand: Fehlerstand) => ReactNode
}

interface Zustand {
  kennung: string | null

  // bereich ist der, in dem der Fehler steht — ein anderer räumt ihn.
  bereich: string
}

// Probewurf wirft, solange die Hüllenprobe diesen Bereich nennt.
function Probewurf({
  bereich,
  children,
}: {
  bereich: string
  children: ReactNode
}) {
  if (probewurf() === bereich) {
    throw new Probewurfsfehler()
  }

  return children
}

// Probewurfsfehler ist der Wurf der Naht — **ohne Satz**: er erreicht die
// Oberfläche nie, nur das Protokoll, und dort genügt sein Name
// (`keinfehlertext.test.ts`).
class Probewurfsfehler extends Error {
  override name = 'Probewurf'
}

export class Fehlergrenze extends Component<Eigenschaften, Zustand> {
  override state: Zustand = { kennung: null, bereich: this.props.bereich }

  // **Ein Wechsel des Bereichs räumt den Fehler** (§7, Regel 1) — in der
  // Zeichenphase, also **vor** einem Wurf der neuen Seite. Ein Räumen nach
  // dem Fangen (`componentDidUpdate`) nahm die Fläche einer Seite, die beim
  // Hinwechseln warf, sofort wieder weg (Hüllenprobe, 2026-10-14). Ohne
  // `key` am Aufrufer, der den ganzen Inhalt neu zeichnen ließe.
  static getDerivedStateFromProps(
    eigenschaften: Eigenschaften,
    zustand: Zustand,
  ): Partial<Zustand> | null {
    return eigenschaften.bereich === zustand.bereich
      ? null
      : { kennung: null, bereich: eigenschaften.bereich }
  }

  static getDerivedStateFromError(): Partial<Zustand> {
    return { kennung: neueKennung() }
  }

  // **Gemeldet wird, was zu haben ist** — die Meldung ist Bestenfalls-Arbeit
  // (§7, Regel 3): scheitert sie, steht die Fläche trotzdem.
  override componentDidCatch(fehler: Error) {
    probewurfGefangen(this.props.bereich)

    // **Die Kennung steht schon**: `getDerivedStateFromError` lief vorher.
    // Ein Rückfall auf `''` wäre ein Zweig, den kein Fang erreicht.
    const gemeldet = api().FensterfehlerMelden({
      kennung: String(this.state.kennung),
      ebene: this.props.ebene,
      bereich: this.props.bereich,
      // **Mit dem Namen**: `TypeError: …` sagt mehr als der Satz allein.
      meldung: String(fehler),
    })
    gemeldet.catch(() => undefined)
  }

  neu = () => {
    this.setState({ kennung: null })
  }

  override render() {
    if (this.state.kennung !== null) {
      return this.props.zeige({ kennung: this.state.kennung, neu: this.neu })
    }

    return (
      <Probewurf bereich={this.props.bereich}>{this.props.children}</Probewurf>
    )
  }
}
