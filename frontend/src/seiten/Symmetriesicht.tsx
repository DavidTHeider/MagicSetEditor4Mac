// **Was eine Symmetrie auf der Zeichenfläche zeigt**
// (#1092).
//
// # Warum die Kopien blass sind
//
// Weil sie nichts sind, was man anfassen kann. MSE zeichnet sie ebenso
// (`SymbolViewer::draw`, `src/render/symbol/viewer.cpp`, Zeilen 150–190,
// Fassung f9d9356 vom 2021-07-06): verändert wird immer nur das Original, und
// eine Kopie, die aussähe wie eines, lüde dazu ein, sie anzufassen.
//
// # Und warum Mitte und Griff Anfasser sind
//
// Weil MSE sie so hat (`SymmetryEditor`, `src/gui/symbol/symmetry_editor.cpp`).
// Die Mitte sagt, worum gedreht wird, der Griff, in welchem Winkel gespiegelt
// — beides ist Geometrie, und Geometrie zieht man.
//
// **Mit der Tastatur geht beides über die Zahlenfelder** (`Symmetriefeld.tsx`).
// Ein Anfasser allein wäre für jemanden ohne Maus nicht da (§6).

import type { KeyboardEvent } from 'react'

import type { SymbolteilDTO, SymbolzugDTO } from '../antworten'
import { useTexte } from '../sprache/kontext'

interface SymmetriesichtEigenschaften {
  teil: SymbolteilDTO
  kopien: SymbolzugDTO[][]
  groesse: number

  // schritt und grosserSchritt sind, wie weit eine Pfeiltaste die Anfasser
  // bewegt — dieselben Werte wie bei einem Punkt.
  schritt: number
  grosserSchritt: number

  mitteAuf: (x: number, y: number) => void
  griffAuf: (x: number, y: number) => void
}

export function Symmetriesicht({
  teil,
  kopien,
  groesse,
  schritt,
  grosserSchritt,
  mitteAuf,
  griffAuf,
}: SymmetriesichtEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.karten.symbol

  // **Die erste Kopie ist das Original, und sie wird kräftig gezeichnet.**
  //
  // Der erste Anlauf ließ sie ganz weg — „sie steht schon als Teil da". Das
  // stimmte nicht: ein `part:` vom Typ `symmetry` hat **keine eigenen
  // Punkte**, sein Umriss ist leer, und auf der Fläche blieb allein die blasse
  // Kopie stehen. Gefunden bei der Sichtprüfung im laufenden Fenster
  // (#1092),
  // nicht im Test — dort sagt kein Zusicherer, wie etwas **aussieht** (§3).
  const [original, ...blasse] = kopien

  return (
    <g>
      {(original ?? []).map((zug, wo) => (
        <path
          key={`original-${String(wo)}`}
          d={alsZug(zug, groesse)}
          className="fill-auswahl/60 stroke-auswahlkante"
          strokeWidth={2}
        />
      ))}

      {/* **Die Kopien sind für den Bildschirmleser nicht da.** Sie tragen
          keine Auskunft, die er nicht schon hätte: was sie zeigen, steht in
          `Symmetriefeld` als Art und Kopienzahl — und zehn Umrisse
          vorzulesen wäre Lärm statt Hilfe. */}
      {blasse.map((kopie, nummer) => (
        <g key={nummer} aria-hidden="true">
          {kopie.map((zug, wo) => (
            <path
              key={wo}
              d={alsZug(zug, groesse)}
              className="fill-auswahl/25 stroke-auswahlkante/50"
              strokeWidth={2}
            />
          ))}
        </g>
      ))}

      <Anfasser
        x={teil.mitteX}
        y={teil.mitteY}
        groesse={groesse}
        beschriftung={eigene.mitte}
        schritt={schritt}
        grosserSchritt={grosserSchritt}
        setzen={mitteAuf}
        gestalt="mitte"
      />

      {/* **Der Griff steht relativ zur Mitte** — `handle` ist bei MSE ein
          Vektor von ihr aus, und sein Winkel ist die Spiegelachse. */}
      <Anfasser
        x={teil.mitteX + teil.griffX}
        y={teil.mitteY + teil.griffY}
        groesse={groesse}
        beschriftung={eigene.griff}
        schritt={schritt}
        grosserSchritt={grosserSchritt}
        setzen={(x, y) => {
          griffAuf(x - teil.mitteX, y - teil.mitteY)
        }}
        gestalt="griff"
      />
    </g>
  )
}

// alsZug macht aus einer Kopie einen SVG-Pfad.
//
// **Ohne Kurven.** Die Kopien sind ein Hinweis darauf, wohin die Symmetrie
// zeichnet; die Griffe drehen sich mit, und sie hier mitzuführen hieße, die
// Abbildung ein zweites Mal zu rechnen — dieses Mal in TypeScript.
function alsZug(zug: SymbolzugDTO, groesse: number): string {
  const erster = zug.punkte[0]
  if (erster === undefined) {
    return ''
  }

  const stelle = (wert: number) => (wert * groesse).toFixed(2)

  let pfad = `M ${stelle(erster.x)} ${stelle(erster.y)}`

  for (const punkt of zug.punkte.slice(1)) {
    pfad += ` L ${stelle(punkt.x)} ${stelle(punkt.y)}`
  }

  return `${pfad} Z`
}

interface AnfasserEigenschaften {
  x: number
  y: number
  groesse: number
  beschriftung: string
  schritt: number
  grosserSchritt: number
  setzen: (x: number, y: number) => void
  gestalt: 'mitte' | 'griff'
}

// Anfasser ist ein bedienbarer Punkt — mit Maus und mit Tastatur.
function Anfasser({
  x,
  y,
  groesse,
  beschriftung,
  schritt,
  grosserSchritt,
  setzen,
  gestalt,
}: AnfasserEigenschaften) {
  const taste = (ereignis: KeyboardEvent<SVGElement>) => {
    const weite = ereignis.shiftKey ? grosserSchritt : schritt

    const richtungen: Record<string, [number, number]> = {
      ArrowLeft: [-weite, 0],
      ArrowRight: [weite, 0],
      ArrowUp: [0, -weite],
      ArrowDown: [0, weite],
    }

    const richtung = richtungen[ereignis.key]
    if (richtung === undefined) {
      return
    }

    ereignis.preventDefault()
    setzen(begrenzt(x + richtung[0]), begrenzt(y + richtung[1]))
  }

  const gemeinsam = {
    tabIndex: 0,
    role: 'button',
    'aria-label': beschriftung,
    onKeyDown: taste,
    strokeWidth: 2,
  }

  if (gestalt === 'mitte') {
    return (
      <circle
        {...gemeinsam}
        cx={x * groesse}
        cy={y * groesse}
        r={7}
        className="cursor-pointer fill-markierung stroke-markierungskante focus:outline-2 focus:outline-offset-2 focus:outline-ring"
      />
    )
  }

  // **Ein Quadrat für den Griff.** Zwei Anfasser, die gleich aussehen, sind
  // einer zu viel — MSE unterscheidet sie ebenso.
  const kante = 12

  return (
    <rect
      {...gemeinsam}
      x={x * groesse - kante / 2}
      y={y * groesse - kante / 2}
      width={kante}
      height={kante}
      className="cursor-pointer fill-flaeche stroke-markierungskante focus:outline-2 focus:outline-offset-2 focus:outline-ring"
    />
  )
}

// begrenzt hält eine Stelle im Einheitsquadrat.
//
// **Wie ein Punkt**: eine `.mse-symbol` rechnet in `[0…1]`, und eine Mitte
// außerhalb wäre eine Drehachse, die niemand sieht.
function begrenzt(wert: number): number {
  return Math.min(1, Math.max(0, wert))
}
