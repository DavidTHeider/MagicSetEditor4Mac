// **Passt die neue Fassung zu diesem Bestand?** — Werkstück O6 von Phase 5.2
// (#1430).
//
// # Ein Bauteil der Darstellung
//
// **Geholt wird im Band, nicht hier.** Es braucht das Urteil selbst: eine
// Fassung, die diesen Bestand **nicht** öffnen kann, soll auch keinen
// Knopf zum Herunterladen daneben haben — das fand die Sichtprüfung, und
// ein Rückruf von hier nach oben wäre derselbe Zustand an zwei Orten.
//
// # Warum das im Band steht und nicht in den Einstellungen
//
// **Weil es hier gefragt wird.** Das Band meldet eine neuere Fassung und
// bietet an, sie zu laden; die Frage „kann sie meinen Bestand öffnen?"
// gehört genau dorthin — eine Seite weiter wäre sie eine Auskunft, die
// niemand sucht.
//
// # Drei Ausgänge, und der dritte ist kein Nein
//
// **`ungeklaert` heißt: die Freigabe sagt nichts über sich.** Ältere tun das
// nicht. Der Weg von Hand steht dann daneben — das ist die zweite Abnahme aus
// dem Issue, und sie verlangt einen **nachvollziehbaren** Weg, keinen
// erfundenen Wert.
import { useTexte } from '../sprache/kontext'

import type { VertraeglichkeitAntwort } from '../apivertraeglichkeit'

export function Vertraeglichkeit({
  urteil,
  aufVorbereiten,
  fehler,
}: {
  urteil: VertraeglichkeitAntwort | undefined
  aufVorbereiten: () => void
  fehler: string | undefined
}) {
  const eigene = useTexte().allgemein.vertraeglichkeit

  if (fehler !== undefined) {
    return <span className="text-fehler">{fehler}</span>
  }

  if (urteil === undefined || urteil.urteil === '') {
    return null
  }

  return (
    <>
      <span>{satzZumUrteil(eigene, urteil)}</span>

      {urteil.hindernisse.map((hindernis) => (
        <span key={`${hindernis.art}/${hindernis.stelle}`}>
          {satzZumHindernis(eigene, hindernis.art)}
          {hindernis.stelle === '' ? '' : ` (${hindernis.stelle})`}
        </span>
      ))}

      {/* **Vorbereiten steht nur da, wenn die Fassung passt** — sonst wäre
          es ein Knopf, der sicher nein sagt. */}
      {urteil.urteil === 'vertraeglich' && !urteil.bereit ? (
        <button
          type="button"
          onClick={aufVorbereiten}
          className="rounded-sm border border-marke px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        >
          {eigene.vorbereiten}
        </button>
      ) : null}
    </>
  )
}

// satzZumUrteil und satzZumHindernis übersetzen die Marken aus Go.
//
// **Marken und keine Sätze** — dieselbe Entscheidung wie beim
// Betriebszustand (O9): vier Flächen bauen aus derselben Marke ihren eigenen
// Satz. Der Rückfall ist die Marke selbst; sie ist immer noch eine Auskunft.
function satzZumUrteil(
  eigene: ReturnType<typeof useTexte>['allgemein']['vertraeglichkeit'],
  urteil: VertraeglichkeitAntwort,
): string {
  if (urteil.bereit) {
    return eigene.bereit
  }

  const saetze: Record<string, string> = {
    vertraeglich: eigene.vertraeglich,
    unvertraeglich: eigene.unvertraeglich,
    ungeklaert: eigene.ungeklaert,
  }

  return saetze[urteil.urteil] ?? urteil.urteil
}

function satzZumHindernis(
  eigene: ReturnType<typeof useTexte>['allgemein']['vertraeglichkeit'],
  art: string,
): string {
  const saetze: Record<string, string> = eigene.hindernisse

  return saetze[art] ?? art
}
