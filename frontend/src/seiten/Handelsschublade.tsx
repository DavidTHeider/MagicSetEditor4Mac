// **Neuer Tausch — mit wem?** — UXR-15
// (#1923),
// nach dem Artboard *Handeln*: „+ Neuer Tausch" führt die Spalte links und
// geht zum Gegenüber.
//
// # Was vorher war
//
// **Vier Kästen untereinander in einer Spalte von 320 Punkten**: „Was gerade
// läuft", darunter jedes bestätigte Gegenüber mit seinem Knopf, dann
// „Gegenüber aufnehmen oder bestätigen" und die Kontakte. Mit zehn
// Gegenübern und zwölf offenen Handeln waren das 31 Bedienelemente (R2 in
// `docs/UIREVIEW.md`) — und die Liste der Handel, um die es meist geht, stand
// über einer Liste von Menschen, die man selten neu wählt.
//
// # Was jetzt ist
//
// **Die drei Fragen nach dem Gegenüber stehen in einer Schublade**: wen
// wählen, wen aufnehmen, wen aus den Kontakten holen. Wer eines wählt, sieht
// den Raum dahinter — die Schublade schließt sich, und ein neuer Handel steht
// zum Zusammenstellen da. **Nichts geht verloren**: dasselbe Bauteil, derselbe
// Weg zur Bestätigung (UX-03.AK-03), nur einen Klick weiter.
//
// # Gezeichnet beim ersten Öffnen, danach nur verborgen
//
// **Wie die Schubladen der Sammlung** (`Sammlungswerkzeuge`): ein halb
// getippter Einladungscode und die Vorschau der Kontakte überstehen das
// Schließen.
import { Activity, useState } from 'react'

import { Partnerwahl } from '../bausteine/Partnerwahl'
import { Schublade } from '../bausteine/Schublade'
import { useTexte } from '../sprache/kontext'
import { Adressbuch } from './Adressbuch'
import { Gegenueberkasten } from './Gegenueberkasten'

interface Handelsschubladeneigenschaften {
  offen: boolean
  schliessen: () => void

  // waehlen schlägt ein Gegenüber mit einem neuen Handel auf.
  waehlen: (kennung: string, name: string) => void
}

export function Handelsschublade({
  offen,
  schliessen,
  waehlen,
}: Handelsschubladeneigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.handeln

  // **Nach einer Bestätigung lädt die Wahl nach** (UX-03.AK-03): das eben
  // bestätigte Gegenüber steht sofort zur Wahl.
  const [kontaktstand, setzeKontaktstand] = useState(0)

  // **Beim Zeichnen gesetzt und nicht in einem Effekt**, wie im
  // `Aufklappkasten`: sonst stünde die Schublade erst einen Zug später da.
  const [gezeichnet, setzeGezeichnet] = useState(offen)
  if (offen && !gezeichnet) {
    setzeGezeichnet(true)
  }

  if (!gezeichnet) {
    return null
  }

  return (
    <Activity mode={offen ? 'visible' : 'hidden'}>
      <Schublade
        titel={eigene.neuerTauschTitel}
        schliessenBezeichnung={
          texte.allgemein.arbeitsplatz.schubladen.schliessen
        }
        schliessen={schliessen}
      >
        <div className="flex flex-col gap-4 px-4 pb-5 pt-3">
          {/* **Kein Gegenüber steht hier »aufgeschlagen«** (#2391): eine
              Wahl beginnt immer einen neuen Entwurf, und der Knopf sagt, was
              er tut — nicht, was schon ist. */}
          <Partnerwahl
            offen={undefined}
            waehlen={(kennung, name) => {
              waehlen(kennung, name)
              schliessen()
            }}
            stand={kontaktstand}
          />

          {/* **Aufnehmen und Bestätigen hier und nicht in den
              Einstellungen** (UX-03.AK-03, #1675) — wer handeln will,
              lernt sein Gegenüber dort kennen, wo er handelt. */}
          <Gegenueberkasten
            aufGeaendert={() => {
              setzeKontaktstand((bisher) => bisher + 1)
            }}
          />

          {/* **Kontakte aus einem Adressbuch daneben, nicht darin** (ID-04,
              #1703): ein Kontakt ist kein Gegenüber. */}
          <Adressbuch />
        </div>
      </Schublade>
    </Activity>
  )
}
