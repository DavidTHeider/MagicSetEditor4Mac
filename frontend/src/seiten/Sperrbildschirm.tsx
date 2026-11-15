// **Der Sperrbildschirm** — ID-02
// (#1701).
//
// **Er steht statt der Anwendung da**, nicht über ihr: gesperrt bekommt Go
// keine Frage an den Bestand eines Profils durch (`imProfilOhneSperre`), und
// eine Anwendung dahinter fragte trotzdem und zeigte Fehler.
import { useState } from 'react'

import { api } from '../fassade'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { Kennwortfeld } from '../bausteine/Kennwortfeld'
import { Knopf } from '../bausteine/Knopf'
import { Pflichtfeld } from '../bausteine/Pflichtfeld'
import { useSitzungsschritt } from '../haken/useSitzungsschritt'
import { useTexte } from '../sprache/kontext'

import type { SitzungsstandAntwort } from '../apisitzung'

interface Eigenschaften {
  stand: SitzungsstandAntwort

  // geoeffnet lässt die Sitzung neu lesen.
  geoeffnet: () => void
}

export function Sperrbildschirm({ stand, geoeffnet }: Eigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.nutzer.sitzung
  const schritt = useSitzungsschritt(geoeffnet)

  const [kennwort, setzeKennwort] = useState('')
  const [vergessen, setzeVergessen] = useState(false)
  const [code, setzeCode] = useState('')
  const [neu, setzeNeu] = useState('')

  return (
    <main
      data-sitzung="gesperrt"
      className="mx-auto flex max-w-md flex-col gap-4 p-8"
    >
      <h1 className="text-2xl font-semibold">{eigene.gesperrtTitel}</h1>
      <Erklaerungsabsatz text={eigene.gesperrtErklaerung} klasse="text-leise" />

      {stand.kennwort && (
        <div className="flex flex-col gap-3">
          <Kennwortfeld
            beschriftung={eigene.kennwort}
            wert={kennwort}
            befund={schritt.befund.kennwort}
            geaendert={setzeKennwort}
          />
          <div>
            <Knopf
              gewicht="haupt"
              angehalten={kennwort === '' || schritt.laeuft}
              aufKlick={() => {
                void schritt.fuehre(() => api().Entsperren(kennwort))
              }}
            >
              {eigene.entsperren}
            </Knopf>
          </div>
        </div>
      )}

      {stand.sensor && (
        <div className="flex flex-col gap-1">
          <div>
            <Knopf
              angehalten={schritt.laeuft}
              aufKlick={() => {
                void schritt.fuehre(() => api().EntsperrenMitSensor())
              }}
            >
              {eigene.lokalEntsperren}
            </Knopf>
          </div>
          <Erklaerungsabsatz
            text={eigene.lokalHinweis}
            klasse="text-sm text-leise"
          />
        </div>
      )}

      <div>
        <Knopf
          gewicht="leise"
          ausgeklappt={vergessen}
          aufKlick={() => {
            setzeVergessen(!vergessen)
          }}
        >
          {eigene.vergessen}
        </Knopf>
      </div>

      {vergessen && (
        <section aria-label={eigene.vergessen} className="flex flex-col gap-3">
          <Erklaerungsabsatz
            text={eigene.vergessenErklaerung}
            klasse="text-sm text-leise"
          />
          <Pflichtfeld
            beschriftung={eigene.code}
            wert={code}
            befund={schritt.befund.faktor}
            geaendert={setzeCode}
          />
          <Kennwortfeld
            beschriftung={eigene.kennwortNeu}
            wert={neu}
            befund={schritt.befund.neu}
            hinweis={eigene.kennwortNeuHinweis}
            geaendert={setzeNeu}
          />
          <div>
            <Knopf
              angehalten={code === '' || neu === '' || schritt.laeuft}
              aufKlick={() => {
                void schritt.fuehre(() =>
                  api().KennwortWiederherstellen(stand.nutzerID, code, neu),
                )
              }}
            >
              {eigene.neuSetzen}
            </Knopf>
          </div>
        </section>
      )}

      <div>
        <Knopf
          gewicht="leise"
          angehalten={schritt.laeuft}
          aufKlick={() => {
            void schritt.fuehre(() => api().Abmelden())
          }}
        >
          {eigene.abmelden}
        </Knopf>
      </div>

      {schritt.fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {texte.bausteine.zustaende.fehler} {schritt.fehler}
        </p>
      )}
    </main>
  )
}
