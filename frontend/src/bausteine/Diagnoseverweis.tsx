// **„Diagnose öffnen" an einem Fehler mit Vorgangskennung** — NET-02,
// S13.NAT-04 (#1706).
//
// Der Satz eines unerwarteten Fehlers nennt die Kennung, unter der im
// Protokoll alle Zeilen des Vorgangs stehen. **Dieser Knopf führt dorthin**
// — zur Gruppe *Diagnose* in den Einstellungen, die die Kennung als Filter
// übernimmt —, statt dass jemand die Stelle sucht und die Kennung abtippt.
import { Knopf } from './Knopf'
import { useAnschluss } from '../anschlussraum'
import { vorgangAus } from '../meldung'
import { useTexte } from '../sprache/kontext'

export function Diagnoseverweis({ ursache }: { ursache: unknown }) {
  const texte = useTexte()
  const anschluss = useAnschluss()
  const vorgang = vorgangAus(ursache)

  if (vorgang === undefined) {
    return null
  }

  return (
    <Knopf
      gewicht="leise"
      aufKlick={() => {
        anschluss.uebergib({
          art: 'einstellungsgruppe',
          name: texte.bausteine.diagnoseverweis.oeffnen,
          gruppe: 'diagnose',
          vorgang,
        })
      }}
    >
      {texte.bausteine.diagnoseverweis.oeffnen}
    </Knopf>
  )
}
