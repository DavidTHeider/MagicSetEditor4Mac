// **Was unter dem Beleg steht: wer gegengezeichnet hat — oder warum nicht** —
// #1650 und
// #2500.
//
// **Aus `Handelsbuchung.tsx` gelöst**, als der Befund dazukam: die Frage ist
// eine eigene — was sagt die zweite Unterschrift? —, und die Datei hätte sie
// nur länger gemacht.
//
// # Drei Lagen
//
// 1. **Ein geprüfter Beleg liegt da** — von wem und wann. Er verdrängt einen
//    Befund, der vor ihm kam.
// 2. **Einer kam und passte nicht** — von wem, wann, woran er scheiterte und
//    was als Nächstes zu tun ist. Bis dahin stand hier dauerhaft „noch
//    keiner", als wäre nichts gekommen.
// 3. **Keiner kam** — und der Satz sagt es ausdrücklich, statt es
//    offenzulassen.
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useTexte, useZeitzone } from '../sprache/kontext'
import { alsZeitpunkt } from '../zeitpunkt'

import type { Handelsbeleg } from '../apihandelsverbuchung'

export function Gegenbelegzeile({
  beleg,
  kennung,
  name,
}: {
  beleg: Handelsbeleg

  // kennung und name sind die des Gegenübers in diesem Kasten.
  kennung: string
  name: string
}) {
  const eigene = useTexte().seiten.handeln
  const darstellung = useDarstellung()
  const zeitzone = useZeitzone()

  // **Das Gegenüber beim Namen** (UXR-15, #1923) — eine fremde Kennung
  // bleibt, was sie ist.
  const wer = (peer: string) => (peer === kennung && name !== '' ? name : peer)

  // **Lesbar** wie „zuletzt" in der Liste — hier stand die rohe Zeichenkette
  // (Hüllenprobe `nat04handel`, #1672).
  const wann = (zeit: string) => alsZeitpunkt(zeit, darstellung, zeitzone)

  if (!beleg.beidseitig && beleg.befund !== null) {
    const { befund } = beleg

    return (
      <p className="text-xs text-fehler">
        {`${einsetzen(eigene.buchungGegenbelegAbgewiesen, {
          wer: wer(befund.von),
          wann: wann(befund.empfangen),
          // **Ein Mangel, den dieses Fenster nicht kennt, wird genannt**
          // (§16) — eine neuere Gegenseite darf einen sechsten haben.
          mangel:
            (eigene.gegenbelegMaengel as Record<string, string>)[
              befund.mangel
            ] ?? befund.mangel,
        })} ${einsetzen(eigene.buchungGegenbelegRat, { wer: wer(befund.von) })}`}
      </p>
    )
  }

  return (
    <p className="text-xs text-leise">
      {beleg.gegenzeichner === ''
        ? eigene.buchungOhneGegenbeleg
        : einsetzen(eigene.buchungGegenbeleg, {
            wer: wer(beleg.gegenzeichner),
            wann: wann(beleg.gegengezeichnet),
          })}
    </p>
  )
}
