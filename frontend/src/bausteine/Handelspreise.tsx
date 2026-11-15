// **Preise holen, wo getauscht wird** — P10
// (#2378),
// Befund C-17, Entwurf `Markttisch09`: „Preise holen" steht an der Bilanz.
//
// **Was es tut, steht im Haken** (`useHandelspreise`), hier nur, wie es
// aussieht: der Knopf, solange es läuft der Abbruch, darunter der Stand —
// und ein Fehler **am Knopf**, nicht als anonymer Hinweis (§6).
import { useHandelspreise, type Preislauf } from '../haken/useHandelspreise'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Knopf } from './Knopf'

import type { Handelsposition } from '../apihandelsleitung'

interface Eigenschaften {
  positionen: Handelsposition[]

  // geholt liest die Bilanz nach.
  geholt: () => void
}

export function Handelspreise({ positionen, geholt }: Eigenschaften) {
  const texte = useTexte().seiten.handeln.preise
  const { stand, laeuft, kannHolen, holen, abbrechen } = useHandelspreise(
    positionen,
    geholt,
  )

  return (
    <div className="flex flex-col gap-1 text-sm">
      <div className="flex flex-wrap items-center gap-2">
        <Knopf
          gewicht="leise"
          angehalten={laeuft || !kannHolen}
          aufKlick={() => {
            void holen()
          }}
        >
          {texte.holen}
        </Knopf>

        {laeuft && (
          <Knopf gewicht="leise" aufKlick={abbrechen}>
            {texte.abbrechen}
          </Knopf>
        )}
      </div>

      {stand !== undefined &&
        (stand.art === 'teils' ? (
          <p role="alert" className="text-fehler">
            {satzZu(stand, texte)}
          </p>
        ) : (
          <p aria-live="polite" className="text-leise">
            {satzZu(stand, texte)}
          </p>
        ))}
    </div>
  )
}

// satzZu sagt, wie weit das Holen ist.
function satzZu(
  stand: Preislauf,
  texte: {
    laeuft: string
    fertig: string
    teils: string
    abgebrochen: string
  },
): string {
  switch (stand.art) {
    case 'laeuft':
      return einsetzen(texte.laeuft, {
        fertig: String(stand.fertig),
        gesamt: String(stand.gesamt),
      })
    case 'fertig':
      return einsetzen(texte.fertig, { zahl: String(stand.zahl) })
    case 'teils':
      return einsetzen(texte.teils, {
        zahl: String(stand.zahl),
        gesamt: String(stand.gesamt),
        meldung: stand.meldung,
      })
    case 'abgebrochen':
      return einsetzen(texte.abgebrochen, {
        fertig: String(stand.fertig),
        gesamt: String(stand.gesamt),
      })
  }
}
