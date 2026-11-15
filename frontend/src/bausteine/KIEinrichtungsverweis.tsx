// **„KI einrichten" an einer KI-Meldung** — AI-01.AK-04
// (#1742):
// ein Fachweg, dem ein Anbieter oder ein Modell fehlt, öffnet den KI-Bereich
// der Einstellungen — und nicht die Einstellungen im Ganzen, in denen jemand
// suchen müsste.
//
// **Derselbe Anschluss wie `Diagnoseverweis`**: die Gruppe reist mit, und die
// Einstellungen schlagen sie auf.
import { Knopf } from './Knopf'
import { useAnschluss } from '../anschlussraum'
import { useTexte } from '../sprache/kontext'

export function KIEinrichtungsverweis() {
  const texte = useTexte()
  const anschluss = useAnschluss()

  return (
    <Knopf
      gewicht="leise"
      aufKlick={() => {
        anschluss.uebergib({
          art: 'einstellungsgruppe',
          name: texte.bausteine.kieinrichtung.oeffnen,
          gruppe: 'ki',
        })
      }}
    >
      {texte.bausteine.kieinrichtung.oeffnen}
    </Knopf>
  )
}
