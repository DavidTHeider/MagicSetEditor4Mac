// **Ein Eintrag aus dem Menü des offenen Turniers** — AP-6
// (#2363),
// B-04: Ausgeben, Bericht sichern und Löschen stehen hinter „…“.
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.turniere

// turniermenue ist der Knopf „…“ am offenen Turnier dieses Namens.
export async function turniermenue(turnier: string): Promise<HTMLElement> {
  return screen.findByRole('button', {
    name: einsetzen(eigene.handlungenAn, { name: turnier }),
  })
}

// turnierhandlung wählt einen Eintrag aus diesem Menü.
export async function turnierhandlung(
  turnier: string,
  eintrag: string,
): Promise<void> {
  await userEvent.click(await turniermenue(turnier))
  await userEvent.click(
    screen.getByRole('menuitem', { name: new RegExp(`^${eintrag}`) }),
  )
}
