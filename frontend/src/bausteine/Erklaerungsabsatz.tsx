// **Ein Erläuterungsabsatz, der sich kurz fasst** — UX-15.AK-01 bis AK-03
// (#1771).
//
// # Wofür es ihn gibt
//
// **Weil dreißig Absätze über den Baum verteilt stehen.** `Seitenkopf`
// erledigt die Kopferklärung jeder Seite und `Wahl` den Hinweis unter jedem
// Feld; dazwischen liegen die **Abschnittserklärungen** — gemessen am
// 2026-09-22 dreißig Stück von 200 bis 478 Zeichen, jede als roher `<p>` in
// ihrem Bauteil.
//
// **Derselbe Schnitt an einer dritten Stelle**: erster Satz sichtbar, Rest
// hinter dem Fragezeichen. Wer ihn an dreißig Stellen einzeln schriebe,
// hätte dreißig Fassungen davon (§10).
import { Feldhilfe } from './Feldhilfe'
import { Satz } from './Satz'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { kurzform, langform, ohneAuszeichnung } from '../sprache/kurzform'

interface Absatzeigenschaften {
  text: string

  // klasse reicht die Stilklassen durch — die Absätze stehen in sehr
  // verschiedenen Umgebungen, und eine feste Größe wäre für die meisten die
  // falsche.
  klasse?: string
}

export function Erklaerungsabsatz({
  text,
  klasse = 'max-w-prose text-sm text-leise',
}: Absatzeigenschaften) {
  const eigene = useTexte().bausteine

  const kurz = kurzform(text)

  // **Außerhalb des JSX zusammengesetzt** — `keintextimjsx` weist jede
  // Zeichenkette mit Platzhalter darin ab, und das zu Recht: so sieht ein
  // übersehener Oberflächentext aus. Hier sind es Stilklassen.
  const klassen = `flex items-start gap-2 ${klasse}`
  const langer = langform(ohneAuszeichnung(text))

  return (
    <p className={klassen}>
      <span>
        <Satz text={kurz} />
      </span>

      {/* **Der Knopf heißt nach dem Satz, den er fortsetzt** — der Grund
          steht bei `bausteine.mehrZu`. */}
      {langer !== undefined && (
        <Feldhilfe
          bezeichnung={einsetzen(eigene.mehrZu, {
            satz: ohneAuszeichnung(kurz),
          })}
          text={langer}
        />
      )}
    </p>
  )
}
