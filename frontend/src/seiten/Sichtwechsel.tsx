// **Was zwischen zwei Sichten dasteht** — Befund 1 des Nachlaufs über Phase
// 6.9 (PLAY-05.AK-02,
// #1713).
//
// **Nichts vom vorigen Sitz.** Die alte Sicht zeigte weiter seine Hand und
// ließ an ihr handeln, bis die neue ankam — und nach einem Fehlschlag für
// immer. Hier steht, dass geladen wird, oder was schiefging; gewählt wird
// die Sicht im Kopf darüber, der bleibt.
import { useTexte } from '../sprache/kontext'

export function Sichtwechsel({ fehler }: { fehler: string | undefined }) {
  const eigene = useTexte().seiten.spielen

  return fehler === undefined ? (
    <p role="status" className="p-3 text-sm text-leise">
      {eigene.sichtLaedt}
    </p>
  ) : (
    <p role="alert" className="p-3 text-sm text-fehler">
      {fehler}
    </p>
  )
}
