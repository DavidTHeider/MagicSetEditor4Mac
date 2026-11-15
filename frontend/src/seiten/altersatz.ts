// **Wie lange es her ist** — so grob, wie ein Stand altert. Herausgelöst
// aus `Problemzeile.tsx`, weil die Kachel „Sammlung“ das Alter ihres ältesten
// Preises in denselben Worten sagt (AP-8, Teil 2b,
// #2365).
import { einsetzen } from '../sprache/einsetzen'

// alterSatz sagt, wie lange es her ist — so grob, wie eine Prüfung altert.
export function alterSatz(
  sekunden: number,
  woerter: {
    gerade: string
    minuten: string
    stunden: string
    tag: string
    tage: string
  },
): string {
  if (sekunden < 60) {
    return woerter.gerade
  }

  if (sekunden < 3600) {
    return einsetzen(woerter.minuten, {
      minuten: String(Math.floor(sekunden / 60)),
    })
  }

  if (sekunden < 86400) {
    return einsetzen(woerter.stunden, {
      stunden: String(Math.floor(sekunden / 3600)),
    })
  }

  const tage = Math.floor(sekunden / 86400)

  // **Ein Tag ist einer** — hier stand „vor 1 Tagen“ und „1 days ago“.
  return tage === 1
    ? woerter.tag
    : einsetzen(woerter.tage, { tage: String(tage) })
}
