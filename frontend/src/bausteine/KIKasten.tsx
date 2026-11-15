// **Ein Kasten für die KI unter dem Kartenformular** — UXR-21
// (#1929).
//
// Das Artboard *Karten* zeigt im Inspector die Felder und sonst nichts. Was
// am Feld hilft — Name und Flavortext vorschlagen —, steht am Feld (AI-05);
// Vorschlag, Bild, Einschätzung, mehrere Karten, Herkunft und Ähnliche
// stehen hier, hinter **einem** Aufklapper statt sechs offen untereinander.
//
// # Warum ein Raum
//
// **Der Editor wird je Karte neu aufgebaut** (`Kartenarbeitsflaeche`, der
// Schlüssel trägt Karte und Verlaufsstand). Ein Zustand im Kasten fiele mit
// jeder Karte und jeder Rücknahme zurück auf zu, und wer mit der KI
// arbeitet, schlüge ihn an jeder Karte neu auf. Der Raum steht über dem
// Schlüssel; ohne ihn hält der Kasten seinen Zustand selbst.
import { createContext, useContext, useState, type ReactNode } from 'react'

import { useTexte } from '../sprache/kontext'

type Zustand = readonly [boolean, (offen: boolean) => void]

const Kontext = createContext<Zustand | undefined>(undefined)

// KIKastenraum hält, ob der Kasten offen ist, über einen Neuaufbau hinweg.
export function KIKastenraum({ children }: { children: ReactNode }) {
  const zustand = useState(false)

  return <Kontext.Provider value={zustand}>{children}</Kontext.Provider>
}

// KIKasten zeigt die KI-Teile einer Karte hinter einem Aufklapper.
export function KIKasten({ children }: { children: ReactNode }) {
  const titel = useTexte().bausteine.kikasten.titel
  const eigener = useState(false)
  const [offen, setzeOffen] = useContext(Kontext) ?? eigener

  return (
    <details
      open={offen}
      onToggle={(ereignis) => {
        setzeOffen(ereignis.currentTarget.open)
      }}
      className="rounded-sm border border-kante px-3 py-2"
    >
      <summary className="cursor-pointer text-sm font-semibold focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring">
        {titel}
      </summary>

      <div className="mt-3 flex flex-col gap-3">{children}</div>
    </details>
  )
}
