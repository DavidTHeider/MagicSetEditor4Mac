// **Wie ein Handel dorthin kam** — TRADE-03.AK-03
// (#1717).
//
// **Geladen wird erst, wenn jemand aufklappt** — der Kasten zeichnet seinen
// Inhalt beim ersten Öffnen, und erst dann fragt dieses Bauteil.
import { api } from '../fassade'
import { alsZeitpunkt } from '../zeitpunkt'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { standsatz } from './handelsstaende'
import { useDarstellung, useTexte, useZeitzone } from '../sprache/kontext'
import { useLaden } from '../haken/useLaden'

export function Handelsverlauf({ vorgang }: { vorgang: string }) {
  const texte = useTexte()
  const handeln = texte.seiten.handeln
  const eigene = handeln.assistent
  const darstellung = useDarstellung()
  const zeitzone = useZeitzone()

  const protokoll = useLaden(async () => {
    if (vorgang === '') {
      return []
    }

    const antwort = await api().HandelProtokoll(vorgang)

    return auspacken(antwort, () => antwort.zeilen)
  }, [vorgang])

  if (vorgang === '') {
    return <p className="text-sm text-leise">{eigene.verlaufOhneHandel}</p>
  }

  if (protokoll.fehler !== undefined) {
    return (
      <p role="alert" className="text-sm text-fehler">
        {protokoll.fehler}
      </p>
    )
  }

  const zeilen = protokoll.daten ?? []

  if (zeilen.length === 0) {
    return <p className="text-sm text-leise">{eigene.verlaufLeer}</p>
  }

  return (
    <ol aria-label={eigene.verlauf} className="flex flex-col gap-1 text-sm">
      {zeilen.map((zeile, index) => (
        <li key={`${String(index)}/${zeile.zeit}`}>
          {einsetzen(
            zeile.von === '' ? eigene.verlaufNeu : eigene.verlaufWechsel,
            {
              revision: String(zeile.revision),
              von: standsatz(handeln, zeile.von),
              nach: standsatz(handeln, zeile.nach),
              zeit: alsZeitpunkt(zeile.zeit, darstellung, zeitzone),
            },
          )}
        </li>
      ))}
    </ol>
  )
}
