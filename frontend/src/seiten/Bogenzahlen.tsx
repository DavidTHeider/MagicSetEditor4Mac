// **Wie viele Karten auf wie viele Seiten** — der Inhalt von „Wie es wird" für
// den Druckbogen, UXR-F7
// (#2015).
//
// Links ein Raster mit den Plätzen einer Seite — gefüllt, wo eine Karte stünde —,
// rechts die Zahlen aus derselben Rechnung, die der Knopf danach legt. Das
// Raster ist ein Bild und nur das: es steht nicht im Zugänglichkeitsbaum, und
// was es sagt, steht daneben als Satz.
import type { BogenprobelaufAntwort } from '../antwortenprobelauf'
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useTexte } from '../sprache/kontext'
import { zaehlwort } from '../sprache/zaehlwort'

interface BogenzahlenEigenschaften {
  probe: BogenprobelaufAntwort
}

export function Bogenzahlen({ probe }: BogenzahlenEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.ausgeben.probelauf.bogen
  const darstellung = useDarstellung()

  const zahl = (wert: number) =>
    wert.toLocaleString(darstellung, { maximumFractionDigits: 1 })

  // Belegt ist, was auf der ersten Seite steht: alle Plätze, wenn es mehr
  // Karten gibt als Platz.
  const belegt = Math.min(probe.karten, probe.jeSeite)

  return (
    <div className="flex items-start gap-4">
      <div
        aria-hidden="true"
        className="grid w-24 shrink-0 gap-1 rounded-sm border border-kante p-1.5"
        style={{
          gridTemplateColumns: `repeat(${String(Math.max(probe.spalten, 1))}, 1fr)`,
        }}
      >
        {Array.from({ length: probe.jeSeite }, (_, platz) => (
          <span
            key={platz}
            className={
              platz < belegt
                ? 'aspect-[5/7] rounded-sm border border-kante bg-marke opacity-60'
                : 'aspect-[5/7] rounded-sm border border-kante'
            }
          />
        ))}
      </div>

      <div className="flex flex-col gap-1">
        <p className="font-medium">
          <Bogenzeile probe={probe} />
        </p>

        <p className="text-sm text-leise">
          {einsetzen(eigene.masse, {
            karten: zaehlwort(
              probe.karten,
              texte.allgemein.zaehlwoerter.karten,
            ),
            breite: zahl(probe.seitenbreite),
            hoehe: zahl(probe.seitenhoehe),
            dpi: String(probe.dpi),
          })}
        </p>

        {/* **Eine halbe letzte Seite sieht man sonst erst am Papier.** */}
        {probe.seiten > 1 && probe.aufLetzterSeite < probe.jeSeite && (
          <p className="text-sm text-leise">
            {einsetzen(eigene.letzte, {
              karten: zaehlwort(
                probe.aufLetzterSeite,
                texte.allgemein.zaehlwoerter.karten,
              ),
            })}
          </p>
        )}
      </div>
    </div>
  )
}

// Bogenzeile ist „9 je Seite · 2 Seiten" — **ein Satz an zwei Orten**: in
// „Wie es wird" und neben dem Knopf in der Fußleiste (AP-10, #2367).
export function Bogenzeile({ probe }: BogenzahlenEigenschaften) {
  const texte = useTexte()

  return (
    <>
      {einsetzen(texte.seiten.ausgeben.probelauf.bogen.jeSeite, {
        zahl: String(probe.jeSeite),
        seiten: zaehlwort(probe.seiten, texte.allgemein.zaehlwoerter.seiten),
      })}
    </>
  )
}
