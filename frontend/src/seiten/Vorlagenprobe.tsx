// **Wie es wird, bevor die Vorlage schreibt** — UXR-F7
// (#2015).
//
// Eine eigene Komponente, weil die Frage an eine **gewählte** Vorlage hängt:
// die Fläche darüber hat sie erst, wenn die Liste geladen ist, und ein Haken
// darf nicht hinter einem Zweig stehen. Ein Wechsel der Vorlage oder der
// Wahl macht aus den Zahlen darunter die einer früheren Frage — sie bleiben
// stehen, mit dem Vermerk, dass gerechnet wird, bis die neuen da sind.
import { Wieeswird } from '../bausteine/Wieeswird'
import { useVorlagenprobelauf } from '../haken/useVorlagenprobelauf'
import { Vorlagenzahlen } from './Vorlagenzahlen'

interface VorlagenprobeEigenschaften {
  setID: string
  paket: string

  // wahl und auswahl sind **dieselben**, mit denen der Knopf danach ausführt.
  wahl: Readonly<Record<string, string>>
  auswahl: readonly string[]
}

export function Vorlagenprobe({
  setID,
  paket,
  wahl,
  auswahl,
}: VorlagenprobeEigenschaften) {
  const probe = useVorlagenprobelauf(setID, paket, wahl, auswahl)

  return (
    <Wieeswird
      wartet={probe.daten === undefined && probe.fehler === undefined}
      veraltet={probe.veraltet}
      fehler={probe.fehler}
    >
      {probe.daten !== undefined && <Vorlagenzahlen probe={probe.daten} />}
    </Wieeswird>
  )
}
