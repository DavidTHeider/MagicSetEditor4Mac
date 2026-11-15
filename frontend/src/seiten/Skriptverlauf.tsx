// **Der Verlauf unter der Skriptkonsole** — Werkstück U13 von Phase 5.1
// (#1424). Ein Klick lädt den Ausdruck zurück; was redigiert ist, lässt sich
// nicht zurückladen, denn der Text ist fort.
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { redigiertesZeichen, leere, useSkriptverlauf } from '../skriptverlauf'
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useZeitzone, useTexte } from '../sprache/kontext'
import { alsZeitpunkt } from '../zeitpunkt'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

interface VerlaufEigenschaften {
  laden: (ausdruck: string) => void
}

export function Skriptverlauf({ laden }: VerlaufEigenschaften) {
  const eigene = useTexte().seiten.karten.konsole
  const darstellung = useDarstellung()
  const zeitzone = useZeitzone()
  const eintraege = useSkriptverlauf()

  const ausgangstexte = {
    gelungen: eigene.ausgangGelungen,
    gescheitert: eigene.ausgangGescheitert,
    abgebrochen: eigene.ausgangAbgebrochen,
  }

  return (
    <section aria-label={eigene.verlauf} className="flex flex-col gap-2">
      <h3 className="text-sm font-medium">{eigene.verlauf}</h3>
      <Erklaerungsabsatz
        text={eigene.verlaufErklaerung}
        klasse="text-sm text-leise"
      />

      {eintraege.length === 0 ? (
        <p className="text-sm text-leise">{eigene.verlaufLeer}</p>
      ) : (
        <>
          <ul className="flex max-h-48 flex-col gap-1 overflow-y-auto text-sm">
            {eintraege.map((eintrag) => (
              <li
                key={eintrag.nummer}
                className="flex flex-wrap items-baseline gap-x-2"
              >
                <button
                  type="button"
                  disabled={eintrag.ausdruck === redigiertesZeichen}
                  onClick={() => {
                    laden(eintrag.ausdruck)
                  }}
                  className="rounded-sm px-1 text-left font-mono hover:bg-betont focus-visible:bg-betont focus-visible:outline focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-ring disabled:cursor-not-allowed disabled:opacity-60"
                >
                  <code>{eintrag.ausdruck}</code>
                </button>
                <span className="text-leise">
                  {ausgangstexte[eintrag.ausgang]}
                </span>
                <span className="text-leise">
                  {einsetzen(eigene.dauer, { ms: String(eintrag.dauerMs) })}
                  {' · '}
                  {alsZeitpunkt(eintrag.zeitpunkt, darstellung, zeitzone)}
                </span>
              </li>
            ))}
          </ul>
          <div>
            <Aktionsknopf
              aktion="loeschen"
              gewicht="neben"
              bezeichnung={eigene.verlaufLoeschen}
              aufKlick={leere}
            >
              {eigene.verlaufLoeschen}
            </Aktionsknopf>
          </div>
        </>
      )}
    </section>
  )
}
