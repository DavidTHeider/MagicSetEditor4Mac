// **Das Profilbild** — Werkstück 8 von Phase 4.12
// (#1351).
//
// # Warum ein Ersatz und kein leerer Kreis
//
// **Weil eine Liste von leeren Kreisen keine Zeile von der anderen
// unterscheidet.** Wer kein Bild hat, bekommt den ersten Buchstaben seines
// Namens — genug, um die Zeile zu finden, und keine Behauptung über den
// Menschen dahinter.
//
// # Und warum das `alt` den Namen nennt
//
// **Weil ein Bild ohne Text für eine Vorleseanwendung nicht da ist.** „Bild"
// wäre dasselbe wie nichts; der Name sagt, wessen Bild es ist.
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

interface Eigenschaften {
  name: string

  // datenURI ist das Bild — leer heißt: keines.
  datenURI: string

  // badgeURI ist das Abzeichen des gezeigten Profils — leer heißt: keines.
  //
  // **Es liegt rechts unten darüber** — Werkstück 9 (#1351), Punkt 13: es
  // zeigt die **Rolle**, in der jemand gerade arbeitet, und nicht eine
  // zweite Person. Deshalb überlappt es und steht nicht daneben.
  badgeURI?: string
}

export function Nutzerbild({ name, datenURI, badgeURI = '' }: Eigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.nutzer

  // **Das Abzeichen hängt an beiden Gestalten** — am Bild wie am
  // Anfangsbuchstaben. Wer keines hat, bekommt seinen Buchstaben; die Rolle
  // gilt trotzdem.
  const abzeichen = badgeURI !== '' && (
    <img
      src={badgeURI}
      alt={einsetzen(eigene.badgeVon, { name })}
      /* **`border-kante` und nicht `border-grund`**: der Ring trennt das
         Abzeichen vom Bild darunter, und dafür gibt es in diesem Programm
         **eine** Randfarbe — sie steht gemessen in `docs/KONTRASTE.md`. Eine
         zweite wäre eine Farbe, die niemand nachgemessen hat. */
      className="absolute -right-0.5 -bottom-0.5 size-4 rounded-full border border-kante bg-grund object-cover"
    />
  )

  if (datenURI === '') {
    return (
      <span className="relative inline-flex shrink-0">
        <span
          aria-label={einsetzen(eigene.ohneBild, { name })}
          className="flex size-10 items-center justify-center rounded-full border border-kante text-sm text-leise"
        >
          {/* **Der erste Buchstabe, und `Array.from` statt `[0]`** — ein
            zusammengesetztes Zeichen besteht aus mehreren Codeeinheiten, und
            `name[0]` risse es auseinander.

            **`slice(0, 1).join('')` und nicht `[0] ?? ''`**: das zweite hat
            einen Zweig für den leeren Namen, den kein Aufruf betreten kann —
            `domain.Nutzer` weist ihn ab. Ein Zweig ohne Fall ist eine
            Behauptung darüber, was geschieht (§3). */}
          {Array.from(name).slice(0, 1).join('')}
        </span>

        {abzeichen}
      </span>
    )
  }

  return (
    <span className="relative inline-flex shrink-0">
      <img
        src={datenURI}
        alt={einsetzen(eigene.bildVon, { name })}
        className="size-10 rounded-full border border-kante object-cover"
      />

      {abzeichen}
    </span>
  )
}
