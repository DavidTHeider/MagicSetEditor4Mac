// **Das Feld der Druckkennung** — seit AP-10 (#2367) unter „Weitere
// Angaben“, der Name führt (`bausteine/Drucksuche.tsx`). Tests, die eine
// Kennung tippen, holen das Feld hier — aufgeklappt, wenn es nötig ist.
import { within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { kataloge } from '../sprache/kataloge'

// **In der Sprache des Tests** — die Vorgabe ist Deutsch.
export async function druckfeld(
  wurzel: HTMLElement = document.body,
  texte: { druck: string; weitereAngaben: string } = kataloge.de.seiten
    .sammlung,
): Promise<HTMLElement> {
  const bereich = within(wurzel)
  const offen = bereich.queryByLabelText(texte.druck)
  if (offen !== null) {
    return offen
  }

  await userEvent.click(
    await bereich.findByRole('button', { name: texte.weitereAngaben }),
  )

  return bereich.getByLabelText(texte.druck)
}
