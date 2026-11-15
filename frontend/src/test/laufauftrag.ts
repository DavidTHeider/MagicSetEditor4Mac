// **Ein Auftrag des Fensters, wie Go ihn meldet** — für die Läufe der Seiten,
// die einen Datei-Lauf anstoßen und anhalten (FMT-03, #1692).
import type { AuftragDTO } from '../apiaufmerksamkeit'
import type { FehlerDTO } from '../formen'

export function laufauftrag(teile: Partial<AuftragDTO>): AuftragDTO {
  return {
    id: '4',
    art: 'bogen',
    was: '',
    fortschritt: '',
    stand: 'laeuft',
    grund: '',
    begonnen: '2026-09-29T10:00:00Z',
    sekunden: 1,
    erledigt: 1,
    gesamt: 3,
    einheit: 'karte',
    abbrechbar: true,
    ...teile,
  }
}

// abgebrochen ist der Fehler, mit dem Go einen angehaltenen Lauf beantwortet.
export const abgebrochen: FehlerDTO = {
  code: 'abgebrochen',
  schluessel: '',
  meldung: 'abgebrochen',
  feld: '',
}
