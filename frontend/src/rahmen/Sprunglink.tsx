// **Der Sprung zum Inhalt ist die erste Station der Tastaturbedienung**: wer
// mit Tab beginnt, will nicht durch fünf Navigationseinträge.
//
// **Aus `Rahmen.tsx` herausgezogen**, weil die Datei an der Zeilengrenze
// stand (§2) — derselbe Grund wie bei der `Seitenleiste`.
import { useTexte } from '../sprache/kontext'

const klassen =
  'sr-only focus:not-sr-only focus:absolute focus:left-2 focus:top-2 ' +
  'focus:rounded-sm focus:bg-flaeche focus:px-3 focus:py-2 focus:shadow ' +
  'dark:focus:bg-betont'

export function Sprunglink() {
  const texte = useTexte()

  return (
    <a href="#inhalt" className={klassen}>
      {texte.allgemein.navigation.sprungZumInhalt}
    </a>
  )
}
