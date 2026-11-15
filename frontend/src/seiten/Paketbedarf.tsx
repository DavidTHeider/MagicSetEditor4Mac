// **Was dem Set an Paketen fehlt, und wie es dazukommt** — DOM-01.AK-04
// (#1696):
// „Fehlende Abhängigkeiten lassen sich aus dem betroffenen Fachworkflow
// diagnostizieren und gezielt beheben."
//
// **Der Fachworkflow ist das Set**: hier steht, welches Spiel, welcher Stil
// und welche Abhängigkeit darunter fehlt oder zu alt ist — und **zwei Wege**,
// es zu beheben, ohne die Lasche zu verlassen oder zu raten, wohin:
//
//  1. *Paket einbinden …* wählt ein Verzeichnis und zeigt **hier** den Plan,
//     derselbe wie auf der Paketseite (`usePaketplan`).
//  2. *Bei den Paketen suchen* springt zur Paketseite, die Suche schon mit
//     dem Namen gefüllt — für den Bezug oder die Übernahme aus MSE2.
//
// Ohne Befund steht nichts da: eine Fläche, die sagt, dass alles da ist,
// wäre eine Zeile mehr ohne Handlung.
import { useCallback } from 'react'

import { api, auspacken, type AbhaengigkeitDTO } from '../api'
import { useAnschluss } from '../anschlussraum'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { Flaeche } from '../bausteine/Flaeche'
import { Knopf } from '../bausteine/Knopf'
import { useLaden } from '../haken/useLaden'
import { usePaketplan } from '../haken/usePaketplan'
import { useTexte } from '../sprache/kontext'
import { abhaengigkeitssatz } from './abhaengigkeitssatz'
import { Paketplanansicht } from './Paketplanansicht'

interface Eigenschaften {
  setID: string
}

export function Paketbedarf({ setID }: Eigenschaften) {
  const texte = useTexte()
  const t = texte.seiten.pakete.plan.bedarf
  const anschluss = useAnschluss()

  const holen = useCallback(async () => {
    const antwort = await api().Paketbedarf(setID)

    return auspacken(antwort, () => antwort.befunde)
  }, [setID])
  const bedarf = useLaden<AbhaengigkeitDTO[]>(holen, [setID])
  const planung = usePaketplan(bedarf.nachladen)

  const befunde = bedarf.daten ?? []

  // **Ein Fehler steht am Ort** (§6) — ein stiller Kasten hieße „nichts
  // fehlt", und das wüsste hier niemand.
  if (bedarf.fehler !== undefined) {
    return (
      <p role="alert" className="text-fehler">
        {texte.bausteine.zustaende.fehler} {bedarf.fehler}
      </p>
    )
  }

  if (befunde.length === 0 && planung.plan === undefined) {
    return null
  }

  return (
    <Flaeche ueberschrift={t.ueberschrift}>
      <Erklaerungsabsatz
        text={t.erklaerung}
        klasse="mb-2 max-w-prose text-leise"
      />

      <ul className="flex flex-col gap-2">
        {befunde.map((befund) => (
          <li
            key={`${befund.paket}/${befund.braucht}/${befund.art}`}
            className="flex flex-wrap items-center gap-2"
          >
            <span className="max-w-prose">
              {abhaengigkeitssatz(befund, texte)}
            </span>
            <Knopf
              klein
              aufKlick={() => {
                anschluss.uebergib({
                  art: 'paketsuche',
                  name: befund.braucht,
                  paket: befund.braucht,
                })
              }}
            >
              {t.suchen}
            </Knopf>
          </li>
        ))}
      </ul>

      <div className="mt-3 flex flex-col gap-3">
        <div>
          <Knopf
            angehalten={planung.laeuft}
            aufKlick={() => {
              void planung.waehlen()
            }}
          >
            {texte.seiten.pakete.plan.einbinden}
          </Knopf>
        </div>

        <Paketplanansicht planung={planung} />
      </div>
    </Flaeche>
  )
}
