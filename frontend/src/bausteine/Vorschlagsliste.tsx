// **Die Vorschläge unter dem Schreibzeiger** — Werkstück 1 von Phase 3.3
// (#1164).
//
// # Warum eine Leiste über dem Feld und kein schwebender Kasten
//
// Ein Kasten am Zeiger müsste die Pixelbreite jedes Zeichens kennen, sich am
// Rand umklappen und beim Rollen mitwandern — drei Wege, auf denen er
// danebenliegt. Eine Leiste steht immer richtig, ist mit der Tastatur
// erreichbar und verdeckt nichts.
//
// **Sie steht oben und nicht unten**: unter dem Feld liegen die Knöpfe, und
// eine Liste, die sie beim Tippen verschöbe, ließe jemanden danebengreifen.
import { Knopf } from './Knopf'
import { useTexte } from '../sprache/kontext'

interface VorschlagslisteEigenschaften {
  vorschlaege: readonly string[]
  uebernehmen: (name: string) => void
}

export function Vorschlagsliste({
  vorschlaege,
  uebernehmen,
}: VorschlagslisteEigenschaften) {
  const texte = useTexte()
  const eigene = texte.bausteine.vorschlaege

  return (
    <div className="order-3 flex flex-wrap items-center gap-1">
      <span className="text-sm text-leise">{eigene.bezeichnung}</span>

      {vorschlaege.map((name) => (
        <Knopf
          key={name}
          bezeichnung={name}
          // **`onMouseDown` und nicht `onClick`.** Ein Klick nimmt dem Feld
          // zuerst den Fokus, und der `onBlur` dort räumt die Vorschläge — der
          // Knopf wäre weg, bevor er gedrückt ist.
          //
          // **Und das gilt weiter, seit das Räumen an der Gruppe hängt.** Wer
          // dort `relatedTarget` prüft, verlässt sich darauf, dass der Knopf
          // den Fokus bekommt — unter WebKit auf macOS bekommt ein `<button>`
          // ihn beim Klicken **nicht**. Unter jsdom bekommt er ihn, und eine
          // Vorrichtung, die großzügiger ist als der Betrieb, prüft den
          // falschen Weg (§3). Deshalb bleibt der Griff beim Drücken.
          aufDruck={() => {
            uebernehmen(name)
          }}
          // **Und für die Tastatur derselbe Weg.** Ein `mousedown` gibt es
          // dort nicht; ohne `aufKlick` wäre die Liste nur mit der Maus zu
          // bedienen (§6).
          aufKlick={() => {
            uebernehmen(name)
          }}
        >
          <span className="font-mono text-xs">{name}</span>
        </Knopf>
      ))}
    </div>
  )
}
