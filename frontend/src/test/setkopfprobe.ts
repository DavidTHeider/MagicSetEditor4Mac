// **Ein Setkopf ohne eine einzige Angabe** —
// #1325.
//
// **Für die Läufe, die ihn nicht meinen.** Die fünf Formatangaben haben ihren
// eigenen Lauf (`seiten/setkopf.test.tsx`); überall sonst stünde eine gefüllte
// Gestalt nur im Weg — und `SetangabenAntwort` verlangt sie.
import type { SetkopfDTO } from '../api'

export const leererSetkopf: SetkopfDTO = {
  mseFassung: '',
  spielfassung: '',
  stilfassung: '',
  versionskontrolle: '',
  apprenticecode: '',
}
