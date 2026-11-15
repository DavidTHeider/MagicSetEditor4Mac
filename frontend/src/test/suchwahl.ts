// **Eine Zeile aus einer Auswahl mit Suche wählen** — UXR-F1
// (#2007).
//
// Ein natives Auswahlfeld wählte `userEvent.selectOptions`; die `Suchauswahl`
// hat keine `<option>`, solange sie zu ist. Der Weg ist der der Hand: das Feld
// anklicken, die Zeile anklicken — beides über die rohe Eingabe (§3).
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

// waehleZeile öffnet das Kombinationsfeld und klickt die Zeile mit diesem
// Namen. **Der Name ist der der Zeile, wie ein Vorleseprogramm ihn nennt** —
// Name und Zusatz, durch ein Leerzeichen getrennt.
export async function waehleZeile(
  feld: HTMLElement,
  zeile: string | RegExp,
): Promise<void> {
  await userEvent.click(feld)
  await userEvent.click(await screen.findByRole('option', { name: zeile }))
}
