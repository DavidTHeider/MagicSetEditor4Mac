// **Der Vorhang zwischen zwei Sitzen** — P7
// (#2375),
// Artboard `Vorhang09`: der Tisch ruht, **beide Hände sind verdeckt**, bis
// der neue Sitz bestätigt. Wann er zugeht, steht in `useVorhang`.
//
// # Tastatur
//
// **Der Fokus steht auf „Ich bin …“** — ↩ bestätigt; Esc lässt den
// bisherigen Sitz weiterhandeln. Kein Dialog über dem Tisch, sondern an
// seiner Stelle (§6: keine verschachtelten Dialoge) — der Kopf mit der
// Sichtwahl bleibt erreichbar.
import { useEffect, useId, useRef } from 'react'

import { Abschnittstitel } from './Abschnittstitel'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { Knopf } from './Knopf'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

export function Vorhang({
  wer,
  bisher,
  bestaetigen,
  bleiben,
}: {
  // wer und bisher sind die Namen des neuen und des bisherigen Sitzes.
  wer: string
  bisher: string
  bestaetigen: () => void
  bleiben: () => void
}) {
  const eigene = useTexte().seiten.spielen
  const titel = useId()
  const knopf = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    // **Ohne Rollen** — der Vorhang steht an der Stelle des Tisches, und
    // der Kopf mit der Sichtwahl soll im Blick bleiben.
    knopf.current?.focus({ preventScroll: true })
  }, [])

  return (
    // eslint-disable-next-line jsx-a11y/no-noninteractive-element-interactions -- Esc für den ganzen Vorhang, die Knöpfe bleiben die Bedienelemente
    <section
      aria-labelledby={titel}
      onKeyDown={(ereignis) => {
        if (ereignis.key === 'Escape') {
          ereignis.preventDefault()
          bleiben()
        }
      }}
      className="flex min-h-0 flex-1 flex-col items-center justify-center gap-4 rounded-md border border-kante bg-flaeche p-6 text-center"
    >
      <Abschnittstitel id={titel} className="text-lg font-semibold">
        {einsetzen(eigene.vorhangTitel, { wer })}
      </Abschnittstitel>

      <p className="text-sm">{einsetzen(eigene.vorhangSatz, { wer })}</p>

      <div className="flex flex-wrap justify-center gap-2">
        <Knopf griff={knopf} gewicht="haupt" aufKlick={bestaetigen}>
          {einsetzen(eigene.vorhangBestaetigen, { wer })}
        </Knopf>

        <Knopf gewicht="leise" kuerzel="Escape" aufKlick={bleiben}>
          {einsetzen(eigene.vorhangBleiben, { wer: bisher })}
        </Knopf>
      </div>

      <Erklaerungsabsatz text={eigene.vorhangHinweis} />
    </section>
  )
}
