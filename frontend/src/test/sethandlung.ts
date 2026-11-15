// **Eine Handlung an einer Setkarte wählen** — UXR-04
// (#1912).
//
// Sichern, Sichern unter …, Umschlag, Exportvorlage und Löschen standen bis
// dahin als Knöpfe an jeder Karte; jetzt stehen sie in ihrem Menü. **Eine
// Vorrichtung statt eines Griffs je Test**: der Weg ist derselbe, und wer ihn
// ändert, ändert ihn hier.
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import type { Katalog } from '../sprache/kataloge'
import { einsetzen } from '../sprache/einsetzen'

// waehleSethandlung öffnet das Menü des genannten Sets und wählt darin den
// Eintrag — beim Namen, wie ein Vorleseprogramm ihn hört.
export async function waehleSethandlung(
  texte: Katalog,
  setname: string,
  eintrag: string,
): Promise<void> {
  await userEvent.click(
    await screen.findByRole('button', {
      name: einsetzen(texte.seiten.sets.handlungenAn, { name: setname }),
    }),
  )
  await userEvent.click(screen.getByRole('menuitem', { name: eintrag }))
}
