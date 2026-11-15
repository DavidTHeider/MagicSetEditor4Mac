// **Eine Identität anlegen, wo sie fehlt** — UXR-23
// (#1931),
// Befund #2027.
//
// Ein Profil ohne Identität sah unter *Handeln* einen roten Fehler und
// „Erneut laden" — das konnte nichts ändern, und der Kasten, der die
// Identität anlegt, stand unter *Einstellungen › Verbindungen*. **Hier ist der
// Anfang**: ein Leerzustand, der sagt, wozu die Identität dient, und sie auf
// einen Klick anlegt (`PeeridentitaetSicherstellen`, derselbe Weg wie dort).
import { useState } from 'react'

import { Aktionsknopf } from './Aktionsknopf'
import { Leerzustand } from './Leerzustand'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'

const dieMarke = 'identitaet'

export function Identitaetanlegen({ angelegt }: { angelegt: () => void }) {
  const texte = useTexte()
  const eigene = texte.seiten.handeln.liste
  const vorgang = useVorgang()

  const [fehler, setzeFehler] = useState<string>()

  return (
    <Leerzustand
      titel={eigene.ohneIdentitaetTitel}
      erklaerung={eigene.ohneIdentitaetErklaerung}
    >
      <Aktionsknopf
        aktion="anlegen"
        gewicht="haupt"
        schreibt
        angehalten={vorgang.laeuft(dieMarke)}
        aufKlick={() => {
          void vorgang.starte(dieMarke, async () => {
            setzeFehler(undefined)

            try {
              const antwort = await api().PeeridentitaetSicherstellen()
              auspacken(antwort, () => undefined)
            } catch (ursache: unknown) {
              setzeFehler(meldungAus(ursache, texte))

              return
            }

            // **Geschrieben ist geschrieben** (§7, Regel 3): die Liste lädt
            // nach und zeigt dann, was es zu handeln gibt.
            angelegt()
          })
        }}
      >
        {eigene.ohneIdentitaetAnlegen}
      </Aktionsknopf>

      {fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {fehler}
        </p>
      )}
    </Leerzustand>
  )
}
