// **Ein fremder Wert im Feld, in dem jemand tippt** — U01 bis U06
// (#2119).
//
// Steht unter der Eingabe, solange `useEntwurf` einen Konflikt sieht: der
// Satz nennt den neuen Wert, zwei Knöpfe entscheiden. **Kein Dialog** (§6):
// die Frage gehört an das Feld, das sie betrifft.
import type { Entwurfsgriff } from '../haken/useEntwurf'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Knopf } from './Knopf'

interface EntwurfskonfliktEigenschaften {
  griff: Entwurfsgriff

  // fremd ist der gespeicherte Wert, so wie er im Satz stehen soll.
  fremd: string

  angehalten: boolean
}

export function Entwurfskonflikt({
  griff,
  fremd,
  angehalten,
}: EntwurfskonfliktEigenschaften) {
  const texte = useTexte().bausteine.entwurf

  if (!griff.konflikt) {
    return null
  }

  return (
    <span role="status" className="flex flex-wrap items-center gap-2">
      <span className="text-sm">
        {einsetzen(texte.fremdGeaendert, { wert: fremd })}
      </span>
      <Knopf gewicht="neben" angehalten={angehalten} aufKlick={griff.behalten}>
        {texte.behalten}
      </Knopf>
      <Knopf gewicht="neben" angehalten={angehalten} aufKlick={griff.verwerfen}>
        {texte.nehmen}
      </Knopf>
    </span>
  )
}
