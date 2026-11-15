// **Was zu einem Gegenüber noch unterwegs ist** — Werkstück T5 von Phase 5.6
// (#1470).
//
// # Warum das dasteht, auch wenn nichts aussteht
//
// **Weil „nichts steht aus" eine Auskunft ist.** Ein Kasten, der nur bei
// einem Stau erscheint, lässt im ruhigen Fall offen, ob gerade alles
// angekommen ist oder ob niemand nachgesehen hat.
//
// # Und warum hier Zahlen stehen und keine Nachrichten
//
// **Weil eine Diagnose sagen soll, ob etwas klemmt, und nicht, worum es
// geht** — der Inhalt steht nebenan, in der Sicht und im Abgleich.
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'

interface Kontaktposteigenschaften {
  kennung: string
}

export function Kontaktpost({ kennung }: Kontaktposteigenschaften) {
  const eigene = useTexte().seiten.handeln

  const stand = useLaden(async () => {
    const antwort = await api().Handelspoststand(kennung)

    return auspacken(antwort, () => antwort.stand)
  }, [kennung])

  return (
    <>
      {stand.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {stand.fehler}
        </p>
      )}

      {stand.daten !== undefined && (
        <p className="text-sm text-leise tabular-nums">
          {stand.daten.offen === 0
            ? eigene.postLeer
            : `${eigene.postOffen}: ${String(stand.daten.offen)} · ${eigene.postVersuche}: ${String(stand.daten.hoechsteVersuche)}`}
        </p>
      )}
    </>
  )
}
