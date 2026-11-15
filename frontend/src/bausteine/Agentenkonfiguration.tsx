// **Fertige Konfiguration für einen Agenten** — Werkstück A14 von Phase 5.3
// (#1453).
//
// # Wofür
//
// **Weil `mse-mcp` in keinem Suchpfad liegt.** Es steckt im Programmbündel,
// und eine Anleitung, die `mse-mcp` schreibt, führt bei niemandem zum Ziel.
// Was hier steht, ist der **volle Pfad dieses Rechners** in der Gestalt, die
// der jeweilige Agent liest.
//
// # Und was hier nicht geschieht
//
// **Es wird nichts geschrieben und nichts gestartet.** Eine Anwendung, die in
// `~/.claude.json` schriebe, überschriebe die Arbeit eines Menschen an einer
// Datei, die ihm gehört. Kopiert wird, nicht ausgeführt — derselbe Satz wie
// bei den Befehlszeilen darüber.
import { useState } from 'react'

import { Aktionsknopf } from './Aktionsknopf'
import { einsetzen } from '../sprache/einsetzen'
import { inDieZwischenablage } from '../zwischenablage'
import { useTexte } from '../sprache/kontext'

import type { AgentenbruchstueckDTO } from '../antworten'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

interface AgentenEigenschaften {
  agenten: AgentenbruchstueckDTO[]
}

export function Agentenkonfiguration({ agenten }: AgentenEigenschaften) {
  const eigene = useTexte().seiten.technik

  // kopiert hält, **was** zuletzt kopiert wurde — Agent und Art zusammen,
  // denn je Agent gibt es zwei Knöpfe.
  const [kopiert, setzeKopiert] = useState('')
  const [kopierfehler, setzeKopierfehler] = useState(false)

  // **Kein Kasten ohne Inhalt** (§6): ohne Bündel gibt es kein `mse-mcp`, und
  // eine Überschrift über nichts verspräche etwas, was nicht kommt.
  if (agenten.length === 0) {
    return null
  }

  function kopiere(marke: string, text: string) {
    void inDieZwischenablage(text).then((geglueckt) => {
      setzeKopiert(geglueckt ? marke : '')
      setzeKopierfehler(!geglueckt)
    })
  }

  return (
    <section aria-label={eigene.agenten} className="mt-8 flex flex-col gap-3">
      <h2 className="text-lg font-semibold">{eigene.agenten}</h2>

      <Erklaerungsabsatz
        text={eigene.agentenErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      {/* **Der Satz über das Geheimnis steht hier und nicht im Code**: er ist
          die Zusage, die den Kasten überhaupt erlaubt. */}
      <p className="max-w-prose text-sm text-leise">
        {eigene.agentenOhneGeheimnis}
      </p>

      <ul className="flex flex-col gap-5">
        {agenten.map((stueck) => (
          <li key={stueck.agent} className="flex flex-col gap-2">
            <h3 className="text-base font-medium">
              {eigene.agentenNamen[stueck.agent] ?? stueck.agent}
            </h3>

            <p className="font-mono text-sm break-all text-leise">
              {einsetzen(eigene.agentenDatei, { datei: stueck.datei })}
            </p>

            <div className="flex flex-wrap items-start gap-2">
              <pre
                role="region"
                aria-label={einsetzen(eigene.agentenInhalt, {
                  datei: stueck.datei,
                })}
                tabIndex={0}
                className="overflow-x-auto rounded-sm bg-betont/60 px-2 py-1 font-mono text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
              >
                {stueck.inhalt}
              </pre>

              <Aktionsknopf
                aktion="kopieren"
                bezeichnung={eigene.agentenKopierenDatei}
                aufKlick={() => {
                  kopiere(stueck.agent + '/datei', stueck.inhalt)
                }}
              >
                {kopiert === stueck.agent + '/datei'
                  ? eigene.kopiert
                  : eigene.kopieren}
              </Aktionsknopf>
            </div>

            <span className="text-sm font-medium">{eigene.agentenBefehl}</span>

            <div className="flex flex-wrap items-center gap-2">
              <code className="overflow-x-auto rounded-sm bg-betont/60 px-2 py-1 font-mono text-sm">
                {stueck.befehl}
              </code>

              <Aktionsknopf
                aktion="kopieren"
                bezeichnung={eigene.agentenKopierenBefehl}
                aufKlick={() => {
                  kopiere(stueck.agent + '/befehl', stueck.befehl)
                }}
              >
                {kopiert === stueck.agent + '/befehl'
                  ? eigene.kopiert
                  : eigene.kopieren}
              </Aktionsknopf>
            </div>

            {/* **Woher die Gestalt stammt, steht dabei** (§10): eine
                Konvention aus dem Gedächtnis ist kein Prüfergebnis, und wer
                sie morgen nachsieht, soll wissen, wo und wann. */}
            <p className="text-sm text-leise">
              {einsetzen(eigene.agentenNachgelesen, {
                quelle: stueck.nachgelesen,
              })}
            </p>
          </li>
        ))}
      </ul>

      {/* **Auch das Nein wird gezeigt** — dieselbe Zusage wie bei den
          Befehlszeilen darüber. */}
      {kopierfehler && (
        <p role="alert" className="text-sm text-fehler">
          {eigene.kopierfehler}
        </p>
      )}
    </section>
  )
}
