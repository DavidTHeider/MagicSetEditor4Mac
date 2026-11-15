// **Was dem Menschen gehört, nicht dem Profil** — Werkstück 8 von Phase 4.12
// (#1351),
// Punkte 9, 10 und 11.
//
// # Warum das auf der Nutzerseite steht und nicht auf der Einstellungsseite
//
// **Weil es eine Einstellung an einem Menschen ist, und Menschen stehen
// dort.** Auf der Einstellungsseite suchte zwar, wer eine Einstellung ändern
// will — aber dann lägen die Einstellungen einer Person an **zwei** Orten,
// unter *Nutzer und Profile › Mein Zugang* und unter *Nutzer*.
//
// **Zwei Orte für eine Sache heißt: an beiden suchen.** Die Frage „was gilt
// für mich?" hat eine Antwort und eine Lasche.
//
// # Und warum die Auskunft daruntersteht
//
// **Weil sie dieselbe Frage beantwortet**: was weiß dieses Programm über
// mich. Ein eigener Ort dafür wäre ein zweiter, an dem man suchen müsste.
import { Satz } from '../bausteine/Satz'
import { useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { meldungAus } from '../meldung'
import { useNutzereinstellungsstand } from '../nutzereinstellungsraum'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'
import { Wahl } from './Wahl'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

export function Persoenliches() {
  const texte = useTexte()
  const eigene = texte.seiten.nutzer
  const stand = useNutzereinstellungsstand()

  const [fehler, setzeFehler] = useState<string | undefined>(undefined)

  // **Ein Vorgang für die Auskunft** (§7, Regel 4): zweimal sichern zugleich
  // schriebe zweimal in dieselbe Datei.
  const vorgang = useVorgang()

  const sichern = () =>
    vorgang.starte('auskunft', async () => {
      setzeFehler(undefined)

      try {
        const antwort = await api().PersonendatenSichern(stand.nutzerID)

        auspacken(antwort, () => undefined)
      } catch (grund) {
        setzeFehler(meldungAus(grund, texte))
      }
    })

  return (
    <div className="flex flex-col gap-6">
      <Erklaerungsabsatz
        text={eigene.persoenlichErklaerung}
        klasse="max-w-prose text-leise"
      />

      {stand.fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {texte.bausteine.zustaende.fehler} {stand.fehler}
        </p>
      )}

      <div className="grid max-w-xl gap-5">
        {stand.eintraege.map((eintrag) => (
          <Wahl
            key={eintrag.schluessel}
            eintrag={eintrag}
            angehalten={stand.laeuft(eintrag.schluessel)}
            aendern={(wert) => {
              stand.setze(eintrag.schluessel, wert)
            }}
          />
        ))}
      </div>

      {/* **Wo nachzulesen ist, was hier erfasst wird** — Werkstück 13 von
          Phase 4.12 (#1351). **Vor der Auskunft und nicht danach**: wer
          wissen will, was in der Datei steht, liest es lieber, bevor er sie
          sich gibt. */}
      <div className="flex max-w-prose flex-col gap-2">
        <h3 className="text-base font-medium">{eigene.datenschutz}</h3>

        <p className="text-leise">
          <Satz text={eigene.datenschutzHinweis} />
        </p>
      </div>

      <div className="flex max-w-prose flex-col gap-2">
        <h3 className="text-base font-medium">{eigene.auskunft}</h3>

        <Erklaerungsabsatz
          text={eigene.auskunftErklaerung}
          klasse="text-leise"
        />

        <div>
          <Aktionsknopf
            aktion="sichern"
            angehalten={vorgang.laeuft('auskunft') || stand.nutzerID === ''}
            aufKlick={() => {
              void sichern()
            }}
          >
            {eigene.auskunftSichern}
          </Aktionsknopf>
        </div>

        {fehler !== undefined && (
          <p role="alert" className="text-fehler">
            {fehler}
          </p>
        )}
      </div>
    </div>
  )
}
