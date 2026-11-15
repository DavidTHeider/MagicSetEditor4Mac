// **Die Kommandozeile einrichten** — OPS-03
// (#1720).
//
// **Was es tut und was nie, steht in `internal/app/befehlsweg.go`**; hier
// steht, wie es aussieht: das Verzeichnis, je Befehl sein Stand, die Zeile
// fürs Shellprofil, wenn das Verzeichnis nicht im PATH steht, und eine
// ungefährliche Abfrage zum Prüfen.
import { useState } from 'react'

import { api, auspacken } from '../api'
import type { BefehlsvorschauAntwort } from '../antworten'
import { useLaden } from '../haken/useLaden'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { inDieZwischenablage } from '../zwischenablage'
import { Aktionsknopf } from './Aktionsknopf'
import { Knopf } from './Knopf'

async function befehlszeile(handlung: string): Promise<BefehlsvorschauAntwort> {
  const antwort = await api().Befehlszeile(handlung)

  return auspacken(antwort, () => antwort)
}

export function Befehlszeile({
  nachSchreiben,
}: {
  // nachSchreiben hört nach jedem Schreiben — die Schnittstellen lesen dann
  // nach (OPS03, #2349).
  nachSchreiben?: () => void
}) {
  const texte = useTexte()
  const eigene = texte.seiten.technik
  const vorgang = useVorgang()
  const gelesen = useLaden(() => befehlszeile(''), [])
  const [geschrieben, setzeGeschrieben] = useState<BefehlsvorschauAntwort>()
  const [fehler, setzeFehler] = useState<string>()
  const stand = geschrieben ?? gelesen.daten
  const laeuft = vorgang.laeuft('befehlszeile')

  // **Kopiert wird, nicht ausgeführt** — OPS-03.AK-04 (#1720), dieselbe
  // Zusage wie bei den Agenten: diese Anwendung führt keine Befehle für dich
  // aus. `kopiert` hält, **welcher** Knopf es war; das Nein steht darunter.
  const [kopiert, setzeKopiert] = useState('')
  const [kopierfehler, setzeKopierfehler] = useState(false)
  const kopiere = (marke: string, text: string) => {
    void inDieZwischenablage(text).then((geglueckt) => {
      setzeKopiert(geglueckt ? marke : '')
      setzeKopierfehler(!geglueckt)
    })
  }

  // **Geschrieben ist geschrieben** (§7, Regel 3): die Antwort ist der neue
  // Stand; ein Fehler steht daneben und nimmt nichts zurück.
  const handeln = (handlung: string) => {
    void vorgang.starte('befehlszeile', async () => {
      try {
        setzeGeschrieben(await befehlszeile(handlung))
        setzeFehler(undefined)
        nachSchreiben?.()
      } catch (grund) {
        setzeFehler(meldungAus(grund, texte))
      }
    })
  }

  return (
    <section className="mt-6 flex flex-col gap-2">
      <h3 className="text-base font-medium">{eigene.befehlszeileTitel}</h3>
      <p className="max-w-prose text-sm text-leise">
        {eigene.befehlszeileSatz}
      </p>

      {stand !== undefined && (
        <>
          <p className="font-mono text-sm break-all">
            {einsetzen(eigene.befehlszeileVerzeichnis, {
              verzeichnis: stand.verzeichnis,
            })}
          </p>
          <ul className="flex flex-col gap-1 text-sm">
            {stand.eintraege.map((eintrag) => (
              <li key={eintrag.name}>
                <span className="font-mono">{eintrag.name}</span>{' '}
                {eigene.befehlszeileStaende[eintrag.stand] ?? eintrag.stand}
                {eintrag.verdeckt !== '' && (
                  <span className="block text-hinweis">
                    {einsetzen(eigene.befehlszeileVerdeckt, {
                      pfad: eintrag.verdeckt,
                    })}
                  </span>
                )}
              </li>
            ))}
          </ul>
          {!stand.imPfad && (
            <p className="flex flex-wrap items-center gap-2 text-sm">
              {eigene.befehlszeileNichtImPfad}{' '}
              <code className="font-mono break-all">{stand.pfadzeile}</code>
              <Aktionsknopf
                aktion="kopieren"
                bezeichnung={eigene.befehlszeileKopierenPfad}
                aufKlick={() => {
                  kopiere('pfad', stand.pfadzeile)
                }}
              >
                {kopiert === 'pfad' ? eigene.kopiert : eigene.kopieren}
              </Aktionsknopf>
            </p>
          )}
          <p className="flex flex-wrap items-center gap-2 text-sm text-leise">
            {eigene.befehlszeilePruefen}{' '}
            <code className="font-mono">{stand.pruefbefehl}</code>
            <Aktionsknopf
              aktion="kopieren"
              bezeichnung={eigene.befehlszeileKopierenPruefung}
              aufKlick={() => {
                kopiere('pruefung', stand.pruefbefehl)
              }}
            >
              {kopiert === 'pruefung' ? eigene.kopiert : eigene.kopieren}
            </Aktionsknopf>
          </p>
          {kopierfehler && (
            <p role="alert" className="text-sm text-fehler">
              {eigene.kopierfehler}
            </p>
          )}
        </>
      )}

      <div className="flex gap-2">
        <Knopf
          angehalten={laeuft}
          aufKlick={() => {
            handeln('installieren')
          }}
        >
          {eigene.befehlszeileEinrichten}
        </Knopf>
        <Knopf
          angehalten={laeuft}
          aufKlick={() => {
            handeln('entfernen')
          }}
        >
          {eigene.befehlszeileHerausnehmen}
        </Knopf>
      </div>

      {(fehler ?? gelesen.fehler) !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler ?? gelesen.fehler}
        </p>
      )}
    </section>
  )
}
