// **Wann zuletzt gesichert wurde, und der freiwillige Plan** — OPS01
// (#2347),
// nach dem Zielartboard `Betrieb09`, Fläche „Sicherung".
//
// **Unbekannt heißt unbekannt**: ohne Vermerk steht hier kein Datum, und nie
// „heute", ohne dass eine Datei zurückgelesen wurde. Über 14 Tage wird es
// hervorgehoben — die Erinnerung.
import { useId } from 'react'

import type { Sicherungsstand } from '../antwortenbestand'
import type { Sicherungsstandlauf } from '../haken/useSicherungsstand'
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useTexte, useZeitzone } from '../sprache/kontext'
import { alsZeitpunkt } from '../zeitpunkt'

type Worte = ReturnType<typeof useTexte>['seiten']['bestandspflege']['alter']

// alterstext ist der kurze Stand im Kopf — `heute`, `vor 3 Tagen`, `unbekannt`.
function alterstext(stand: Sicherungsstand, worte: Worte): string {
  switch (stand.stufe) {
    case 'unbekannt':
      return worte.unbekannt
    case 'unstimmig':
      return worte.unstimmig
    default:
      if (stand.tage === 0) {
        return worte.heute
      }

      return stand.tage === 1
        ? worte.gestern
        : einsetzen(worte.vorTagen, { tage: String(stand.tage) })
  }
}

// flaechenname nennt die Fläche, über die gesichert wurde.
// **Eine Fläche, die der Katalog nicht kennt, heißt, wie sie heißt** — ein
// älterer Vermerk oder eine neuere Fläche.
function flaechenname(flaeche: string, worte: Worte): string {
  const namen: Partial<Record<string, string>> = worte.flaechen

  return namen[flaeche] ?? flaeche
}

export function Sicherungsalter({ lauf }: { lauf: Sicherungsstandlauf }) {
  const texte = useTexte()
  const sprache = useDarstellung()
  const zeitzone = useZeitzone()
  const titelID = useId()
  const hilfeID = useId()
  const worte = texte.seiten.bestandspflege.alter
  const { stand } = lauf

  if (stand === undefined) {
    return lauf.fehler ? (
      <p role="alert" className="mt-4 text-sm text-fehler">
        {lauf.fehler}
      </p>
    ) : null
  }

  const zeitpunkt = alsZeitpunkt(stand.gesichert, sprache, zeitzone)
  const hervor = stand.stufe === 'alt' || stand.stufe === 'unstimmig'

  let satz = worte.unbekanntSatz
  if (stand.stufe === 'unstimmig') {
    satz = einsetzen(worte.unstimmigSatz, { zeitpunkt })
  } else if (stand.stufe !== 'unbekannt') {
    satz = einsetzen(stand.anlass === 'plan' ? worte.satzPlan : worte.satz, {
      zeitpunkt,
      flaeche: flaechenname(stand.flaeche, worte),
      datei: stand.datei,
    })
  }

  return (
    <section aria-labelledby={titelID} className="mt-4 flex flex-col gap-2">
      <div className="flex flex-wrap items-center gap-2">
        <h3 id={titelID} className="text-base font-medium">
          {worte.ueberschrift}
        </h3>
        <span
          className={
            'rounded-full border border-kante px-2 py-0.5 text-xs ' +
            (hervor ? 'font-medium text-fehler' : 'text-leise')
          }
        >
          {stand.stufe === 'alt'
            ? einsetzen(worte.alt, { alter: alterstext(stand, worte) })
            : alterstext(stand, worte)}
        </span>
      </div>

      <p className="max-w-prose text-sm">{satz}</p>

      <label className="flex items-center gap-2 text-sm">
        <input
          type="checkbox"
          checked={stand.plan.ordner !== ''}
          disabled={lauf.laeuft}
          aria-describedby={hilfeID}
          onChange={(ereignis) => {
            if (ereignis.currentTarget.checked) {
              lauf.einrichten()
            } else {
              lauf.beenden()
            }
          }}
          className="focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        />
        {worte.planSchalter}
      </label>

      <p id={hilfeID} className="max-w-prose text-xs text-leise">
        {worte.planHilfe}
      </p>

      {stand.plan.ordner !== '' && (
        <p className="flex max-w-prose flex-wrap gap-x-1 text-sm">
          <span>
            {einsetzen(worte.planOrdner, { ordner: stand.plan.ordner })}
          </span>
          <span>
            {stand.plan.letzter === ''
              ? worte.planNie
              : einsetzen(worte.planZuletzt, {
                  zeitpunkt: alsZeitpunkt(
                    stand.plan.letzter,
                    sprache,
                    zeitzone,
                  ),
                })}
          </span>
        </p>
      )}

      {lauf.planfehler !== undefined && (
        <p role="alert" className="max-w-prose text-sm text-fehler">
          {einsetzen(worte.planFehler, { grund: lauf.planfehler })}
        </p>
      )}

      {lauf.fehler !== undefined && (
        <p role="alert" className="max-w-prose text-sm text-fehler">
          {lauf.fehler}
        </p>
      )}
    </section>
  )
}
