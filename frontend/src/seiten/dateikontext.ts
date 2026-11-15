// **Das Kontextmenü einer Vorlagendatei** — AP-5
// (#2362),
// Befund B-01.
//
// Die Handgriffe wirken auf die **gewählte** Datei, wie im Menü »Datei«
// über dem Editor. Ein Rechtsklick auf eine andere wählt sie deshalb zuerst
// (`useAnstoss`) — samt der Rückfrage der Werkstatt bei Ungesichertem —, und
// das Menü geht erst auf, wenn sie **offen** ist: davor hätte ein Handgriff
// kein Feld, in dem er erscheint.
import type { Dateiaktionen } from '../haken/useDateiaktionen'
import { useAnstoss } from '../haken/useAnstoss'
import type {
  Dateiwahl,
  Vorlagenwerkstatt,
} from '../haken/useVorlagenwerkstatt'
import { useTexte } from '../sprache/kontext'
import { dateieintraege } from './dateieintraege'
import { dateischluessel, type Dateianstoss } from './Vorlagenzweig'

export function useDateikontext(
  werkstatt: Vorlagenwerkstatt,
  aktionen: Dateiaktionen,
  offen: boolean,
) {
  const texte = useTexte().seiten.vorlagen
  const { wahl, puffer } = werkstatt
  const beiwerk = wahl?.beiwerk === true

  const { anstoss, verbraucht, anstossen } = useAnstoss<Dateiwahl>(
    wahl === null ? undefined : dateischluessel(wahl.paket, wahl.datei),
    (ziel) => {
      werkstatt.waehlen(ziel.paket, ziel.datei, ziel.beiwerk)
    },
  )

  const dateianstoss: Dateianstoss = (ereignis, paket, datei, istBeiwerk) => {
    anstossen(ereignis, dateischluessel(paket, datei), {
      paket,
      datei,
      beiwerk: istBeiwerk,
    })
  }

  const bereit = wahl !== null && (offen || beiwerk)

  return {
    eintraege: bereit
      ? dateieintraege(
          aktionen,
          !beiwerk && wahl.datei === werkstatt.hauptdatei,
          !beiwerk && puffer.geaendert,
          texte,
        )
      : [],
    objekt: bereit ? `datei/${wahl.paket}/${wahl.datei}` : undefined,
    anstoss,
    verbraucht,
    anstossen: dateianstoss,
  }
}
