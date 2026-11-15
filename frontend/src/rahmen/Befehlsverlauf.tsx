// **Der Verlauf dieser Sitzung** — Werkstück U12 von Phase 5.1 (#1423).
import { Knopf } from '../bausteine/Knopf'
import { useBefehlsverlauf } from '../befehlsverlauf'
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useZeitzone, useTexte } from '../sprache/kontext'
import { alsZeitpunkt } from '../zeitpunkt'

interface VerlaufEigenschaften {
  uebernehmen: (name: string, argumente: Record<string, unknown>) => void
}

export function Befehlsverlauf({ uebernehmen }: VerlaufEigenschaften) {
  const texte = useTexte().bausteine.kommandozentrale
  const darstellung = useDarstellung()
  const zeitzone = useZeitzone()
  const eintraege = useBefehlsverlauf()

  return (
    <section aria-label={texte.verlauf} className="mt-4">
      <h3 className="mb-1 text-sm font-semibold">{texte.verlauf}</h3>
      {eintraege.length === 0 ? (
        <p className="text-sm text-leise">{texte.verlaufLeer}</p>
      ) : (
        <ul className="flex max-h-40 flex-col gap-1 overflow-y-auto text-sm">
          {eintraege.map((eintrag) => (
            <li
              key={eintrag.nummer}
              className="flex flex-wrap items-baseline gap-x-2"
            >
              <code>{eintrag.name}</code>
              <span className="text-leise">{texte[eintrag.ausgang]}</span>
              <span className="text-leise">
                {einsetzen(texte.dauer, { ms: String(eintrag.dauerMs) })}
                {' · '}
                {alsZeitpunkt(eintrag.zeitpunkt, darstellung, zeitzone)}
              </span>
              <code className="text-xs text-leise">
                {JSON.stringify(eintrag.argumente)}
              </code>
              <Knopf
                gewicht="neben"
                bezeichnung={einsetzen(texte.wiederholen, {
                  name: eintrag.name,
                })}
                aufKlick={() => {
                  uebernehmen(eintrag.name, eintrag.argumente)
                }}
              >
                ↺
              </Knopf>
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
