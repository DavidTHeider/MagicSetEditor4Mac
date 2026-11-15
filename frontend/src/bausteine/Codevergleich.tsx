// **Den Vergleichscode an einem aufgenommenen Gegenüber bestätigen** —
// NET-01 (#1705).
//
// **Derselbe Code wie beim Annehmen**, denn er gehört dem Paar und nicht der
// Stelle, an der man ihn sieht. Wer „später vergleichen" gewählt hat,
// vergleicht hier; der Fingerabdruck steht dahinter für Fachleute.
import { Abdruckvergleich } from './Abdruckvergleich'
import { Aufklappkasten } from './Aufklappkasten'
import { Knopf } from './Knopf'
import { useTexte } from '../sprache/kontext'

import type { Peerkontakt } from '../apipeerkontakt'

interface Codevergleichseigenschaften {
  kontakt: Peerkontakt
  laeuft: boolean
  bestaetigen: (kennung: string, verglichen: string) => void
}

export function Codevergleich({
  kontakt,
  laeuft,
  bestaetigen,
}: Codevergleichseigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.technik.einladung

  return (
    <div className="flex flex-col gap-2">
      <p className="text-sm">{eigene.zeileVergleichen}</p>
      <p data-vergleichscode="" className="font-mono text-base tracking-wider">
        {kontakt.vergleichscode}
      </p>

      <div>
        <Knopf
          gewicht="haupt"
          angehalten={laeuft}
          aufKlick={() => {
            bestaetigen(kontakt.kennung, kontakt.vergleichscode)
          }}
        >
          {eigene.zeileStimmt}
        </Knopf>
      </div>

      <Aufklappkasten titel={eigene.zeileVonHand}>
        <Abdruckvergleich
          kennung={kontakt.kennung}
          laeuft={laeuft}
          bestaetigen={bestaetigen}
        />
      </Aufklappkasten>
    </div>
  )
}
