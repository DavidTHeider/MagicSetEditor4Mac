// **Wo die Dienste lauschen** — G08
// (#2272).
//
// Drei Möglichkeiten: **nur dieser Mac** (die Vorgabe), **ausgewählte
// Adressen** aus der Liste der Schnittstellen dieses Rechners — mehrere
// zugleich — oder **alle**, und die nur ausdrücklich. Gespeichert wird in
// `netz.lauschen`, über denselben Weg wie jede Einstellung; die Regeln,
// was eine Wahl sein darf, stehen in `internal/domain/lauschwahl.go`.
//
// **Eine Wahl, die den Mac verlässt, sagt das** — dann gelten TLS und ein
// Zugang, und macOS fragt womöglich nach der Erlaubnis für das lokale Netz.
// Still geändert wird keine Systemeinstellung.
import { useId, useState } from 'react'

import { api, auspacken } from '../api'
import type { NetzschnittstelleDTO } from '../antworten'
import { useLaden } from '../haken/useLaden'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { Aktionsknopf } from './Aktionsknopf'

const schluessel = 'netz.lauschen'
const alleAdressen = '0.0.0.0'

type Art = 'schleife' | 'auswahl' | 'alle'

interface Gelesen {
  schnittstellen: NetzschnittstelleDTO[]
  wert: string
}

async function lesen(): Promise<Gelesen> {
  const [liste, einstellungen] = await Promise.all([
    api().Netzschnittstellen(),
    api().Einstellungen(),
  ])

  return {
    schnittstellen: auspacken(liste, () => liste.schnittstellen),
    wert: auspacken(
      einstellungen,
      () =>
        einstellungen.einstellungen.find(
          (eintrag) => eintrag.schluessel === schluessel,
        )?.wert ?? '',
    ),
  }
}

// artAus deutet eine gespeicherte Wahl.
function artAus(wert: string): Art {
  const adressen = wert.split(',').map((teil) => teil.trim())

  if (
    wert.trim() === '' ||
    (adressen.length === 1 && adressen[0] === '127.0.0.1')
  ) {
    return 'schleife'
  }

  return adressen.includes(alleAdressen) ? 'alle' : 'auswahl'
}

// aufDerSchleife sagt, ob eine Adresse diesen Mac nicht verlässt.
function aufDerSchleife(adresse: string): boolean {
  return adresse.startsWith('127.') || adresse === '::1'
}

export function Lauschwahl() {
  const texte = useTexte()
  const eigene = texte.seiten.technik.schnittstellen.lauschwahl
  const arten: Record<string, string | undefined> = eigene.arten
  const kennung = useId()
  const vorgang = useVorgang()
  const geladen = useLaden(lesen, [])

  const [art, setzeArt] = useState<Art>()
  const [gewaehlt, setzeGewaehlt] = useState<string[]>()
  const [meldung, setzeMeldung] = useState<string>()
  const [fehler, setzeFehler] = useState<string>()

  const gespeichert = geladen.daten?.wert ?? ''
  const wirksameArt = art ?? artAus(gespeichert)
  const wirksameWahl =
    gewaehlt ??
    (artAus(gespeichert) === 'auswahl'
      ? gespeichert.split(',').map((teil) => teil.trim())
      : [])

  const verlaesst =
    wirksameArt === 'alle' ||
    (wirksameArt === 'auswahl' &&
      wirksameWahl.some((adresse) => !aufDerSchleife(adresse)))

  const umschalten = (adresse: string) => {
    setzeMeldung(undefined)
    setzeGewaehlt(
      wirksameWahl.includes(adresse)
        ? wirksameWahl.filter((eine) => eine !== adresse)
        : [...wirksameWahl, adresse],
    )
  }

  const speichern = () => {
    if (wirksameArt === 'auswahl' && wirksameWahl.length === 0) {
      setzeFehler(eigene.keineGewaehlt)

      return
    }

    const wert =
      wirksameArt === 'schleife'
        ? ''
        : wirksameArt === 'alle'
          ? alleAdressen
          : wirksameWahl.join(',')

    void vorgang.starte('lauschwahl', async () => {
      try {
        auspacken(
          await api().EinstellungSetzen(schluessel, wert),
          () => undefined,
        )
        setzeFehler(undefined)
        setzeMeldung(eigene.gespeichert)
      } catch (grund) {
        setzeMeldung(undefined)
        setzeFehler(meldungAus(grund, texte))
      }
    })
  }

  const wahl = (wert: Art, beschriftung: string) => (
    <label className="flex items-center gap-2">
      <input
        type="radio"
        name={`${kennung}-art`}
        value={wert}
        checked={wirksameArt === wert}
        onChange={() => {
          setzeArt(wert)
          setzeMeldung(undefined)
        }}
      />
      {beschriftung}
    </label>
  )

  return (
    <section
      aria-labelledby={kennung}
      className="flex flex-col gap-2 rounded-sm border border-kante p-2"
    >
      <h4 id={kennung} className="font-medium">
        {eigene.titel}
      </h4>
      <p className="max-w-prose text-sm text-leise">{eigene.satz}</p>

      {/* **Erst, wenn gelesen ist, was gilt** — vorher stand „Nur dieser
          Mac" angehakt, auch über einer gespeicherten Auswahl. */}
      {geladen.daten !== undefined && (
        <fieldset className="flex flex-col gap-1 text-sm">
          <legend className="font-medium">{eigene.wahl}</legend>
          {wahl('schleife', eigene.schleife)}
          {wahl('auswahl', eigene.auswahl)}
          {wahl('alle', eigene.alle)}
        </fieldset>
      )}

      {wirksameArt === 'alle' && (
        <p className="text-sm text-leise">{eigene.alleSatz}</p>
      )}

      {wirksameArt === 'auswahl' && geladen.daten !== undefined && (
        <fieldset className="flex flex-col gap-2 text-sm">
          <legend className="font-medium">{eigene.liste}</legend>
          {geladen.daten.schnittstellen.map((eine) => (
            <div key={eine.name} className="flex flex-col gap-1">
              <span className="flex flex-wrap gap-x-2">
                <span className="font-mono">{eine.name}</span>
                {arten[eine.art] !== undefined && (
                  <span className="text-leise">{arten[eine.art]}</span>
                )}
                {!eine.aktiv && (
                  <span className="text-leise">{eigene.inaktiv}</span>
                )}
              </span>
              {eine.adressen.length === 0 ? (
                <span className="text-leise">{eigene.ohneAdresse}</span>
              ) : (
                eine.adressen.map((adresse) => (
                  <label key={adresse} className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      checked={wirksameWahl.includes(adresse)}
                      onChange={() => {
                        umschalten(adresse)
                      }}
                    />
                    <span className="font-mono">{adresse}</span>
                  </label>
                ))
              )}
            </div>
          ))}
        </fieldset>
      )}

      {verlaesst && (
        <p className="max-w-prose text-sm text-fehler">{eigene.aussen}</p>
      )}

      <div>
        <Aktionsknopf
          aktion="sichern"
          angehalten={vorgang.laeuft('lauschwahl')}
          aufKlick={speichern}
        >
          {eigene.speichern}
        </Aktionsknopf>
      </div>

      {meldung !== undefined && (
        <p role="status" className="text-sm">
          {meldung}
        </p>
      )}
      {(fehler ?? geladen.fehler) !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler ?? geladen.fehler}
        </p>
      )}
    </section>
  )
}
