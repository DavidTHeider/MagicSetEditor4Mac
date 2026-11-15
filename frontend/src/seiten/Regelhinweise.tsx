// **Was vor dem Anfangen sichtbar bleiben muss** — die Hindernisse einer
// Absage und die Widersprüche einer Abwahl (G4,
// #1461;
// PLAY-02, #1710).
//
// # Warum nicht in der Regelschau
//
// **Weil die seit dem 2026-09-24 eingeklappt steht** — das Artboard *Eine
// Partie anfangen*, Befund F13: ihre Aufzählungen standen zwischen den
// Sitzen und dem Startknopf. Was den Anfang anhält, und was an einer Abwahl
// noch hängt, darf nicht mit ihnen verschwinden.
import { einsetzen } from '../sprache/einsetzen'
import { regelbereichwort } from './regelwoerter'
import { useTexte } from '../sprache/kontext'

import type { RegelfreigabeDTO } from '../apiregeln'

export function Regelhinweise({ freigabe }: { freigabe: RegelfreigabeDTO }) {
  return (
    <>
      {/* **Die Rolle am Rahmen, nicht an der Liste** (P4, #2372, axe in
          WebKit): eine `ul` mit fremder Rolle ist keine Liste mehr. */}
      {freigabe.hindernisse.length > 0 && (
        <div role="alert">
          <ul className="flex flex-col gap-1 text-sm text-fehler">
            {freigabe.hindernisse.map((hindernis) => (
              <li key={hindernis}>{hindernis}</li>
            ))}
          </ul>
        </div>
      )}

      <Widersprueche freigabe={freigabe} />
    </>
  )
}

// Widersprueche erklärt, was an einer Abwahl noch hängt.
//
// **Eine Erklärung und keine Sperre** (AK-02): die Partie darf trotzdem
// anfangen — wer die Priorität selbst übernimmt und den Stapel dem Programm
// lässt, soll wissen, dass der Stapel dann auf ihn wartet.
function Widersprueche({ freigabe }: { freigabe: RegelfreigabeDTO }) {
  const eigene = useTexte().seiten.spielen

  if (freigabe.widersprueche.length === 0) {
    return null
  }

  return (
    <section className="flex flex-col gap-1">
      <h4 className="text-sm font-medium">{eigene.regelWidersprueche}</h4>

      <ul className="flex flex-col gap-1 text-sm text-leise">
        {freigabe.widersprueche.map((widerspruch) => (
          <li key={`${widerspruch.bereich}-${widerspruch.haengt}`}>
            {einsetzen(eigene.regelWiderspruch, {
              bereich: regelbereichwort(eigene, widerspruch.bereich),
              haengt: regelbereichwort(eigene, widerspruch.haengt),
              quelle: widerspruch.quelle,
            })}
          </li>
        ))}
      </ul>
    </section>
  )
}
