// Eine Fläche: der Kasten, in dem Inhalt steht.
//
// # Warum es sie gibt
//
// Bis #288
// stand das Aussehen jeder Gruppe auf jeder Seite neu — mal ein Rahmen, mal
// keiner, mal `p-4`, mal gar nichts. Eine Anwendung, deren Abstände von der
// Seite abhängen, sieht nicht schlicht aus, sondern unfertig.
//
// # Der Rahmen ist `slate-500`, und das ist Absicht
//
// Ein Ton wie 300 läge hell bei 1,49 : 1 und 400 bei 2,63 —
// `slate-500` ist der einzige Ton, der in **beiden** Erscheinungsbildern über
// 3 : 1 liegt (`docs/KONTRASTE.md`). AGENTS.md §3 verlangt das für
// Trennlinien; hier gilt dieselbe Zahl, damit nicht zwei Linien derselben
// Anwendung verschiedenen Regeln folgen.
import type { ReactNode } from 'react'

interface FlaecheEigenschaften {
  // ueberschrift macht aus der Fläche einen **benannten Bereich**.
  //
  // **Ohne sie bleibt es ein `div`.** Ein `section` ohne Namen zählt ein
  // Vorleseprogramm mit, ohne sagen zu können, was darin steht — er kostet
  // eine Station und trägt nichts.
  ueberschrift?: string

  // sprungmarke ist die Kennung, auf die ein Verweis zeigt — Werkstück C von
  // Phase 4.10 ([Sprungmarken]).
  //
  // **Ohne Angabe steht keine `id` da.** Eine leere Kennung im Baum wäre ein
  // Ziel, das nichts benennt.
  sprungmarke?: string

  // marke steht neben der Überschrift — „2 Sitze" im Artboard *Eine Partie
  // anfangen*. **Neben und nicht in ihr**: der Name des Bereichs bleibt, was
  // er ist, und eine Zahl darin änderte ihn bei jedem neuen Sitz.
  marke?: string

  children: ReactNode
}

export function Flaeche({
  ueberschrift,
  sprungmarke,
  marke,
  children,
}: FlaecheEigenschaften) {
  const rahmen = 'rounded-lg border border-kante bg-flaeche p-5'

  if (ueberschrift === undefined) {
    return <div className={rahmen}>{children}</div>
  }

  return (
    <section id={sprungmarke} aria-label={ueberschrift} className={rahmen}>
      <div className="mb-3 flex flex-wrap items-baseline gap-2">
        <h2 className="text-lg font-semibold">{ueberschrift}</h2>

        {marke !== undefined && (
          <span className="rounded-full border border-kante px-2 text-xs text-leise">
            {marke}
          </span>
        )}
      </div>

      {children}
    </section>
  )
}
