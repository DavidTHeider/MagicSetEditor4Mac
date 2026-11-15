// **Besitze ich diese Karte?** — Werkstück 5 von Phase 4.5
// (#1292,
// Vermerk [0057](../../../docs/entscheidungen/0057-wie-eine-sammlung-geordnet-wird.md)).
//
// # Warum das an der Karte steht und nicht in der Sammlung
//
// **Weil die Frage dort gestellt wird.** Wer eine Karte ansieht, will wissen,
// ob er sie hat — nicht andersherum. Auf der Sammlungsseite wäre es ein
// Suchfeld für eine Kennung, die dort ohnehin in jeder Zeile steht.
//
// **Es ist die Naht zum Deckbau** aus Phase 4.6
// (#1293):
// dieselbe Frage, dann für jede Zeile eines Decks.
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'

export interface BesitzstandEigenschaften {
  // printing ist der Druck, nach dem gefragt wird.
  printing: string
}

export function Besitzstand({ printing }: BesitzstandEigenschaften) {
  const eigene = useTexte().seiten.sammlung

  // **Am Druck aufgehängt** (§7, Regel 1): ein anderer Druck lädt neu, und
  // die Antwort zum vorigen verschwindet mit ihm.
  const stand = useLaden(async () => {
    const antwort = await api().SammlungBesitz(printing)

    return auspacken(antwort, () => antwort.besitz)
  }, [printing])

  const besitz = stand.daten ?? undefined

  if (besitz === undefined) {
    return null
  }

  return (
    <p className="text-sm">
      {besitz.besitzt
        ? einsetzen(eigene.besitzJa, { stueck: String(besitz.stueck) })
        : eigene.besitzNein}
      {besitz.gewuenscht > 0 && (
        <span className="text-leise">
          {' '}
          {einsetzen(eigene.besitzGewuenscht, {
            gewuenscht: String(besitz.gewuenscht),
          })}
        </span>
      )}
    </p>
  )
}
