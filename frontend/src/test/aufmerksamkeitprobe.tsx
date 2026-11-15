// **Die Aufmerksamkeit, wie `Start` sie zeichnet** — samt ihrem Laden.
//
// Seit AP-8, Teil 2b (#2365) fragt `Start` die Übersicht und reicht sie
// hinein, weil die Kachel „Decks“ dieselbe Antwort liest. Die Tests der
// Liste prüfen die Liste; dieses Bauteil setzt das Laden davor, wie `Start`
// es tut.
import type { ComponentProps } from 'react'

import { Aufmerksamkeit, useAufmerksamkeit } from '../seiten/Aufmerksamkeit'

export function AufmerksamkeitProbe({
  stand,
  ...rest
}: Omit<ComponentProps<typeof Aufmerksamkeit>, 'uebersicht'> & {
  stand: number
}) {
  const uebersicht = useAufmerksamkeit(stand)

  return <Aufmerksamkeit uebersicht={uebersicht} {...rest} />
}
