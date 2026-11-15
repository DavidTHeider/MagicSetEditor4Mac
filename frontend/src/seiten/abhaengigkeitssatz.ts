// **Der Satz zu einem Vorbehalt gegen eine Abhängigkeit** (#57) — an einer
// Stelle, seit ihn zwei Flächen zeigen: der Abschnitt „Abhängigkeiten" und
// der Plan vor einer Paketänderung (DOM-01, #1696). Zwei Fassungen desselben
// Satzes liefen auseinander.
import type { AbhaengigkeitDTO } from '../api'
import { einsetzen } from '../sprache/einsetzen'
import type { Katalog } from '../sprache/kataloge'

export function abhaengigkeitssatz(
  befund: AbhaengigkeitDTO,
  texte: Katalog,
): string {
  const werte = {
    paket: befund.paket,
    braucht: befund.braucht,
    verlangt: befund.verlangt,
    da: befund.da,
  }

  switch (befund.art) {
    case 'fehlt':
      return einsetzen(texte.seiten.pakete.abhaengigFehlt, werte)

    case 'zualt':
      return einsetzen(texte.seiten.pakete.abhaengigZuAlt, werte)

    case 'zyklus':
      return einsetzen(texte.seiten.pakete.abhaengigZyklus, werte)

    default:
      return einsetzen(texte.seiten.pakete.abhaengigUnlesbar, werte)
  }
}
