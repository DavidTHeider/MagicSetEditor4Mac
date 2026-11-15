// **Wer jetzt entscheidet** — P7
// (#2375),
// Befund C-05, Artboard `Vorhang09`.
//
// # Wozu
//
// Im Hot-Seat gibt der Vorhang das Gerät an **diesen** Sitz weiter. Bis P7
// wählte man ihn in der Sichtwahl selbst — nach jedem Wechsel der Priorität.
//
// # In welcher Folge
//
// **Im Beginn** der gezeigte Sitz, solange er nicht behalten hat, sonst der
// nächste in der Reihe, der noch entscheidet (CR 103.5). **Danach** zuerst,
// wer eine Frage beantworten muss, bevor jemand Priorität bekommt: Auslöser
// ordnen (CR 603.3b), eine Legende wählen (CR 704.5j), abwerfen (CR 514.1) —
// dann wer Priorität hat (CR 117.1). Alles Fassung 2026-09-25.
//
// # Was sie nicht sagt
//
// **Wer handeln darf.** Das prüft der Kern; hier steht nur, vor wessen Augen
// die nächste Frage gehört. Ein Sandkasten ohne Zug hat niemanden.
import type { Partie, Sitz } from '../apipartie'

const entscheidetImBeginn = (sitz: Sitz) =>
  sitz.eroeffnung === 'ausstehend' || sitz.eroeffnung === 'offen'

export function entscheider(partie: Partie): string | undefined {
  if (partie.stand === 'beendet') {
    return undefined
  }

  if (partie.sitze.some(entscheidetImBeginn)) {
    const stelle = Math.max(
      0,
      partie.sitze.findIndex((sitz) => sitz.kennung === partie.fuer),
    )
    const reihe = [
      ...partie.sitze.slice(stelle),
      ...partie.sitze.slice(0, stelle),
    ]

    return reihe.find(entscheidetImBeginn)?.kennung
  }

  const { zug } = partie

  return [
    partie.ausloeserreihung,
    partie.legendenwahlen[0]?.sitz ?? '',
    zug.abwurf > 0 ? zug.aktiv : '',
    zug.prioritaet,
  ].find((kennung) => kennung !== '')
}
