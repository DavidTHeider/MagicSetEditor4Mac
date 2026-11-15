// **Zu einem Gebiet einer Seite gehen** — Werkstück I von Phase 4.10
// (#1325).
//
// Seit Einstellungen, Hilfe und die Deckseite ihre Themen in Laschen führen,
// steht nicht mehr alles zugleich da. Ein Test, der etwas aus einem Gebiet
// sucht, geht denselben Weg wie ein Anwender — und diese Hilfe hält den Weg an
// **einer** Stelle.
//
// **Sie wartet auf die Lasche, nicht auf ihren Inhalt.** Wer danach etwas
// zusichert, wartet dafür selbst (§3).
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

export async function zuGebiet(name: string): Promise<void> {
  await userEvent.click(await screen.findByRole('tab', { name }))
}

// zuAbschnitt geht in ein Gebiet und klappt dort einen Abschnitt auf — seit
// #2465 stehen in *Daten und Sicherung* Zurückgehen, Aufräumen und
// Zurücksetzen hinter einem.
export async function zuAbschnitt(
  gebiet: string,
  abschnitt: string,
): Promise<void> {
  await zuGebiet(gebiet)
  await userEvent.click(await screen.findByRole('button', { name: abschnitt }))
}
