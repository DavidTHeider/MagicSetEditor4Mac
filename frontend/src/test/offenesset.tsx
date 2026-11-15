// **Ein Fenster in klein**: liest das offene Set, wie `Anwendung` es tut, und
// reicht es hinunter — für Läufe, die eine Fläche ohne das ganze Fenster
// zeichnen.
import type { ReactNode } from 'react'

import { OffenesSet, useOffenesSet } from '../haken/offenesset'

export function MitOffenemSet({
  setID,
  stand = 0,
  children,
}: {
  setID: string | undefined
  stand?: number
  children: ReactNode
}) {
  return <OffenesSet wert={useOffenesSet(setID, stand)}>{children}</OffenesSet>
}
