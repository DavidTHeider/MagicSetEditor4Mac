// **Laden und Scheitern einer Seite, an einer Stelle** — G11
// (#2277).
//
// B03 verzögerte die Bindungen, die eine Seite beim Öffnen ruft, und ließ sie
// scheitern: sechs Seiten sagten nicht, dass sie laden, drei nicht, dass
// etwas scheiterte — die Sammlung stand als leere Seite da, als fehlten die
// Daten (`docs/belege/huellenprobe/zustaende.md`).
//
// **Das Scheitern geht vor dem Laden**, und nachgeladen wird nur, was
// scheiterte: eine Bindung, die gerade glückte, noch einmal zu fragen, hieße
// ihr Ergebnis wegzuwerfen.
import type { Ladezustand } from '../haken/useLaden'
import { Ladefehler, Wirdgeladen } from './Ladehinweise'

export function Seitenladung({
  zustaende,
}: {
  zustaende: readonly Ladezustand<unknown>[]
}) {
  const gescheitert = zustaende.filter(
    (zustand) => zustand.fehler !== undefined,
  )
  const erster = gescheitert[0]?.fehler

  if (erster !== undefined) {
    return (
      <Ladefehler
        fehler={erster}
        wiederholen={() => {
          for (const zustand of gescheitert) zustand.nachladen()
        }}
      />
    )
  }

  if (zustaende.some((zustand) => zustand.laedt)) {
    return (
      <div aria-busy="true">
        <Wirdgeladen />
      </div>
    )
  }

  return null
}
