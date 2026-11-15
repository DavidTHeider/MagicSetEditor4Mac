// Eine Zuschnittantwort für Tests
// (#576).
//
// **An einer Stelle**, weil drei Testdateien sie brauchen — und weil eine
// Vorrichtung, die jede Datei für sich baut, beim ersten neuen Feld
// auseinanderläuft.
import type { ZuschnittAntwort } from '../api'

// Ein Bildpunkt als PNG — kürzer geht ein Data-URI nicht.
const punkt =
  'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAC0lEQVR4nGNgYAAAAAMAASsJTYQAAAAASUVORK5CYII='

export function zuschnittantwort(
  teile: Partial<ZuschnittAntwort> = {},
): ZuschnittAntwort {
  return {
    pfad: '/pfad/urlaub.jpg',
    quelle: punkt,
    quelleBreite: 400,
    quelleHoehe: 100,
    zielBreite: 100,
    zielHoehe: 80,
    auswahl: { x: 0, y: 0, breite: 125, hoehe: 100 },
    ...teile,
  }
}
