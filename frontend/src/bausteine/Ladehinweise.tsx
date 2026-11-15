// **Laden und Fehlschlag, ohne einen leeren Bestand zu behaupten** — UX-08.NAT-01
// und S15.NAT-01/02 (#1680).
//
// # Wozu es das neben `Ladefeld` gibt
//
// **`Ladefeld` rahmt eine Liste, die für sich lädt.** Die Listen der Decks und
// Turniere laden in einem Haken, der Seite und Liste zugleich bedient: der
// Fehlschlag steht über der Seite, die Liste darunter darf dann nichts
// behaupten. Beides braucht dieselben zwei Sätze, und die stehen hier.
import { Knopf } from './Knopf'
import { useTexte } from '../sprache/kontext'

// Wirdgeladen sagt, dass eine Antwort unterwegs ist — `role="status"`, damit
// es auch ein Vorleser erfährt.
export function Wirdgeladen() {
  const zustaende = useTexte().bausteine.zustaende

  return (
    <p role="status" className="text-leise">
      {zustaende.laedt}
    </p>
  )
}

// Ladefehler nennt, was schiefging, und bietet den nächsten Versuch an. **Ein
// Fehlschlag ist kein leerer Bestand** — ohne den Knopf bliebe nur, die Seite
// zu verlassen und wiederzukommen.
export function Ladefehler({
  fehler,
  wiederholen,
}: {
  fehler: string
  wiederholen: () => void
}) {
  const zustaende = useTexte().bausteine.zustaende

  return (
    <div className="flex flex-wrap items-center gap-3">
      <p role="alert" className="text-fehler">
        {fehler}
      </p>
      <Knopf aufKlick={wiederholen}>{zustaende.wiederholen}</Knopf>
    </div>
  )
}
