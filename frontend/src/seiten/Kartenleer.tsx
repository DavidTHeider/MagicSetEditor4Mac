// **Ein leeres Set** — der Leerzustand der Kartenlasche.
//
// **Einfügen steht hier, weil es sonst nirgends steht**: die Handlungszeile
// und ⌘V hängen an der Kartenspalte, und die gibt es erst ab der ersten
// Karte. Gefunden hat es die Hüllenprobe (`kopieren`, #1727): in ein neues
// Set ließ sich nichts einfügen. Derselbe Griff wie in der Handlungszeile —
// ein zweiter wäre eine zweite Antwort auf dieselbe Frage.
//
// **Und für einen Leser gesperrt** — neuer Befund des Nachlaufs über Phase G:
// Einfügen schreibt ins Set, und bis dahin scheiterte es erst am Tor.
import { Knopf } from '../bausteine/Knopf'
import { Leerzustand } from '../bausteine/Leerzustand'
import type { Kartenablagegriff } from '../haken/useKartenablage'
import { useTexte } from '../sprache/kontext'

export function Kartenleer({ ablage }: { ablage: Kartenablagegriff }) {
  const texte = useTexte()

  return (
    <Leerzustand
      titel={texte.seiten.karten.leerTitel}
      erklaerung={texte.seiten.karten.leer}
    >
      <Knopf schreibt angehalten={ablage.laeuft} aufKlick={ablage.einfuegen}>
        {texte.seiten.karten.ablage.einfuegen}
      </Knopf>
    </Leerzustand>
  )
}
