// **Die Vorrichtung der Tischläufe** — Werkstück G2 von Phase 5.5
// (#1459).
//
// **Eine eigene Datei**, weil zwei Läufe sie brauchen — `partietisch.test.tsx`
// über den Tisch und `partietischbilder.test.tsx` über die Bilder — und weil
// eine gemeinsame Datei über die Zeilengrenze aus §2 ging.
import { fireEvent, render, screen } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'

import { Partietisch } from './Partietisch'
import { Sprachraum } from '../sprache/kontext'
import { einePartie } from './partieprobe'
import { kataloge } from '../sprache/kataloge'

import type { Partie } from '../apipartie'
import type { Partielauf } from '../haken/usePartie'

// zeige zeichnet den Tisch und gibt die Hand.
//
// **`alsListe` sagt, welche Darstellung geprüft wird** — seit G3 gibt es zwei
// (#1460),
// und **eine zur Zeit**: beide zugleich legten jede Karte zweimal in die
// Tabulatorreihenfolge. Die Läufe über die Zonenliste sagen es ausdrücklich,
// statt sich auf eine Voreinstellung zu verlassen, die sich ändern kann.
export function zeige(
  lauf: Partielauf,
  partie: Partie = einePartie(),
  alsListe = true,
) {
  render(
    <Sprachraum sprache="de">
      <Partietisch lauf={lauf} partie={partie} />
    </Sprachraum>,
  )

  if (alsListe)
    fireEvent.click(
      screen.getByRole('button', {
        name: kataloge.de.seiten.spielen.tischListe,
      }),
    )

  return nutzerereignis.setup()
}
