// **Die Turniere in einer Liste** — Werkstück 5 von Phase 4.7
// (#1294).
//
// **Eine Zeile öffnet, die Marken stehen daran** — seit UXR-12
// (#1920).
// Gelöscht wird am offenen Turnier (`useTurnierhandlungen`), wie am offenen Deck.
import { Knopf } from './Knopf'
import { Markenfilterknopf } from './Markenfilterkasten'
import { Markenzeile } from './Markenzeile'
import { nachMarken, useMarkenfilter } from '../haken/useMarkenfilter'
import { Wirdgeladen } from './Ladehinweise'
import { useTexte } from '../sprache/kontext'
import type { Turnierstand } from '../haken/useTurniere'

export function Turnierliste({
  stand,
  oeffnen,
}: {
  stand: Turnierstand

  // oeffnen schlägt ein Turnier auf — die Seite merkt sich dabei, dass der
  // Fokus ihm folgen soll (UXR-12, #1920).
  oeffnen: (turnierID: string) => void
}) {
  const texte = useTexte()
  const eigene = texte.seiten.turniere

  // **Marken an Turnieren** — UX-12 (#1684).
  const marken = useMarkenfilter('turnier')

  // **Leer heißt: angekommen und leer** (UX-08.NAT-01) — wie bei den Decks.
  if (!stand.geladen) {
    return stand.ladefehler === undefined ? <Wirdgeladen /> : null
  }

  if (stand.turniere.length === 0) {
    return <p>{eigene.leer}</p>
  }

  return (
    <section aria-label={eigene.titel} className="flex flex-col gap-3">
      <h2 className="text-lg font-semibold">{eigene.titel}</h2>

      <Markenfilterknopf art="turnier" stand={marken} />

      <ul className="flex flex-col gap-2">
        {nachMarken(
          stand.turniere,
          (turnier) => turnier.id,
          marken.getroffen,
        ).map((turnier) => (
          <li
            key={turnier.id}
            className="flex flex-wrap items-center gap-3 rounded-md border border-kante p-3"
          >
            <span className="font-semibold">{turnier.name}</span>

            <span className="text-sm text-leise">{turnier.spiel}</span>

            {turnier.notiz !== '' && (
              <span className="text-sm text-leise">{turnier.notiz}</span>
            )}

            <Knopf
              bezeichnung={`${eigene.waehlen}: ${turnier.name}`}
              angehalten={stand.vorgang.laeuft(`turnier:${turnier.id}`)}
              aufKlick={() => {
                oeffnen(turnier.id)
              }}
            >
              {eigene.waehlen}
            </Knopf>

            <div className="basis-full">
              <Markenzeile
                art="turnier"
                kennung={turnier.id}
                name={turnier.name}
                marken={marken.je[turnier.id] ?? []}
                nachladen={marken.nachladen}
              />
            </div>
          </li>
        ))}
      </ul>
    </section>
  )
}
