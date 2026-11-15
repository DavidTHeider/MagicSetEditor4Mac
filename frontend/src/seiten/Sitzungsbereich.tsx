// **Die Sitzung auf der Nutzerseite** — ID-02
// (#1701):
// sperren, abmelden und das Kennwort ändern.
import { useState } from 'react'

import { api } from '../fassade'
import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { Kennwortfeld } from '../bausteine/Kennwortfeld'
import { Knopf } from '../bausteine/Knopf'
import {
  useSitzungNeuLesen,
  useSitzungsschritt,
} from '../haken/useSitzungsschritt'
import { useTexte } from '../sprache/kontext'

export function Sitzungsbereich() {
  const texte = useTexte()
  const eigene = texte.seiten.nutzer.sitzung
  const neuLesen = useSitzungNeuLesen()
  const sitzung = useSitzungsschritt(neuLesen)

  const [alt, setzeAlt] = useState('')
  const [neu, setzeNeu] = useState('')
  const [gesetzt, setzeGesetzt] = useState(false)
  const kennwort = useSitzungsschritt(() => {
    setzeAlt('')
    setzeNeu('')
    setzeGesetzt(true)
  })

  return (
    <section aria-label={eigene.ueberschrift} className="flex flex-col gap-3">
      <h2 className="text-base font-medium">{eigene.ueberschrift}</h2>
      <Erklaerungsabsatz
        text={eigene.erklaerung}
        klasse="max-w-prose text-leise"
      />

      <div className="flex flex-wrap gap-3">
        <Knopf
          angehalten={sitzung.laeuft}
          aufKlick={() => {
            void sitzung.fuehre(() => api().Sperren())
          }}
        >
          {eigene.sperren}
        </Knopf>
        <Knopf
          gewicht="leise"
          angehalten={sitzung.laeuft}
          aufKlick={() => {
            void sitzung.fuehre(() => api().Abmelden())
          }}
        >
          {eigene.abmelden}
        </Knopf>
      </div>

      {sitzung.fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {texte.bausteine.zustaende.fehler} {sitzung.fehler}
        </p>
      )}

      {/* **Das Kennwort ändern ist kein Alltag** — UXR-20 (#1928): es steht
          eingeklappt unter Sperren und Abmelden, und der Kasten heißt, wie
          die Überschrift hieß. */}
      <Aufklappkasten titel={eigene.kennwortAendern}>
        <div className="flex max-w-xl flex-col gap-3">
          <Kennwortfeld
            beschriftung={eigene.kennwortAlt}
            wert={alt}
            befund={kennwort.befund.kennwort}
            hinweis={eigene.kennwortAltHinweis}
            geaendert={setzeAlt}
          />
          <Kennwortfeld
            beschriftung={eigene.kennwortNeu}
            wert={neu}
            befund={kennwort.befund.neu}
            hinweis={eigene.kennwortNeuHinweis}
            geaendert={(wert) => {
              setzeNeu(wert)
              setzeGesetzt(false)
            }}
          />
          <div>
            <Knopf
              gewicht="haupt"
              angehalten={neu === '' || kennwort.laeuft}
              aufKlick={() => {
                void kennwort.fuehre(() => api().KennwortWechseln(alt, neu))
              }}
            >
              {eigene.kennwortSetzen}
            </Knopf>
          </div>
        </div>

        {kennwort.fehler !== undefined && (
          <p role="alert" className="text-fehler">
            {texte.bausteine.zustaende.fehler} {kennwort.fehler}
          </p>
        )}
        {gesetzt && (
          <p role="status" className="text-leise">
            {eigene.kennwortGeaendert}
          </p>
        )}
      </Aufklappkasten>
    </section>
  )
}
