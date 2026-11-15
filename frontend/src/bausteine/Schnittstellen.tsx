// **Kommandozeile und Werkzeugserver per Knopf** — OPS-09, Schnitt 1
// (#1907),
// und seit Schnitt 2 Claude Desktop, Claude Code und der Dienst als
// LaunchAgent (#1961).
//
// **Was es tut und was nie, steht in `internal/app/schnittstellen.go`**;
// hier steht, wie es aussieht: je Ziel sein Stand, die Datei und die Zeilen,
// die hinzukommen (AK-01), und zwei Knöpfe, die nur dort gehen, wo sie etwas
// tun können.
import { useId, useState } from 'react'

import { api, auspacken } from '../api'
import type {
  SchnittstelleDTO,
  SchnittstellenAlleAntwort,
  SchnittstellenAntwort,
  SchnittstellenausgangDTO,
} from '../antworten'
import { useLaden } from '../haken/useLaden'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Knopf } from './Knopf'
import { SchnittstellenAlle } from './SchnittstellenAlle'

// zuhauseKurz schreibt das Zuhause als `~` — UXR-19
// (#1927).
// `/Users/…/Library/Application Support/Claude/claude_desktop_config.json`
// brach in der Hüllenprobe `ops09b` über zwei Zeilen, und der Teil, der in
// jeder Karte gleich ist, nahm die halbe Breite. Der ganze Pfad steht im
// Titel. **Liegt das Zuhause woanders, bleibt der Pfad, wie er ist** —
// das Fenster kennt es nicht und rät es nicht.
function zuhauseKurz(pfad: string): string {
  return pfad.replace(/^\/Users\/[^/]+(?=\/|$)/, '~')
}

// markenklassen färben den Stand: eingerichtet in der Marke, veraltet als
// Warnung, alles andere leise.
const markengrund = 'self-start rounded-full border px-2 py-0.5 text-xs '
const markenklassen: Record<string, string | undefined> = {
  eingerichtet: markengrund + 'border-marke text-marke',
  veraltet: markengrund + 'border-kante text-fehler',
}

async function gelesen(): Promise<SchnittstellenAntwort> {
  const antwort = await api().Schnittstellen()

  return auspacken(antwort, () => antwort)
}

export function Schnittstellen({
  nachSchreiben,
}: {
  // nachSchreiben hört nach jedem Schreiben — die Kommandozeile liest dann
  // nach (OPS03, #2349).
  nachSchreiben?: () => void
}) {
  const texte = useTexte()
  const eigene = texte.seiten.technik.schnittstellen
  const vorgang = useVorgang()
  const ueberschrift = useId()
  const geladen = useLaden(gelesen, [])
  const [geschrieben, setzeGeschrieben] = useState<SchnittstellenAntwort>()
  const [fehler, setzeFehler] = useState<string>()
  const [ausgaenge, setzeAusgaenge] = useState<SchnittstellenausgangDTO[]>()
  const stand = geschrieben ?? geladen.daten
  const laeuft = vorgang.laeuft('schnittstellen')

  const namen: Record<string, string | undefined> = eigene.namen
  const staende: Record<string, string | undefined> = eigene.staende
  const saetze: Record<string, string | undefined> = eigene.staendeSaetze
  const vorschau: Record<string, string | undefined> = eigene.vorschau

  // **Geschrieben ist geschrieben** (§7, Regel 3): die Antwort trägt den
  // neuen Stand auch dann, wenn sie einen Fehler meldet.
  const handeln = (weg: () => Promise<SchnittstellenAntwort>) => {
    //
    // **Und eine abgelehnte Zusage kommt als Meldung an** — seit OPS-10
    // (#1908) weist eine Bindung so ab, wenn ihr Bereich abgeschaltet ist.
    void vorgang.starte('schnittstellen', async () => {
      try {
        const antwort = await weg()
        setzeGeschrieben(antwort)
        nachSchreiben?.()
        auspacken(antwort, () => undefined)
        setzeFehler(undefined)
      } catch (grund) {
        setzeFehler(meldungAus(grund, texte))
      }
    })
  }

  // **Alle mit einem Ruf** (G08, #2272): die Antwort trägt den Stand aller
  // Ziele **und** je Ziel seinen Ausgang — auch wenn eines scheiterte, dann
  // steht sein Grund am Ziel und der erste darunter.
  const handelnAlle = (weg: () => Promise<SchnittstellenAlleAntwort>) => {
    void vorgang.starte('schnittstellen', async () => {
      try {
        const antwort = await weg()
        setzeGeschrieben({
          schnittstellen: antwort.schnittstellen,
          nurEintrag: '',
          fehler: null,
        })
        setzeAusgaenge(antwort.ausgaenge)
        nachSchreiben?.()
        auspacken(antwort, () => undefined)
        setzeFehler(undefined)
      } catch (grund) {
        setzeFehler(meldungAus(grund, texte))
      }
    })
  }

  const zeile = (eine: SchnittstelleDTO) => {
    const kennung = `${ueberschrift}-${eine.name}`
    const einrichtbar = eine.stand === 'offen' || eine.stand === 'veraltet'
    const widerrufbar =
      eine.stand === 'eingerichtet' || eine.stand === 'veraltet'

    return (
      <li
        key={eine.name}
        aria-labelledby={kennung}
        className="flex flex-col gap-1 rounded-sm border border-kante p-2"
      >
        <span className="flex flex-wrap items-center gap-2">
          <span id={kennung} className="font-medium">
            {namen[eine.name] ?? eine.name}
          </span>

          {/* **Eine Marke und kein Satz** (UXR-19, #1927): „eingerichtet"
              stand als Fließtext und las sich wie jede andere Zeile. */}
          <span
            data-stand={eine.stand}
            className={
              markenklassen[eine.stand] ??
              markengrund + 'border-kante text-leise'
            }
          >
            {staende[eine.stand] ?? eine.stand}
          </span>
        </span>

        {/* **Der Satz nur, wo er mehr sagt als die Marke** — warum es
            veraltet ist, dass ein fremder Eintrag bleibt. */}
        {saetze[eine.stand] !== undefined && (
          <span className="text-sm text-leise">{saetze[eine.stand]}</span>
        )}

        <span
          title={eine.datei}
          className="font-mono text-sm break-all text-leise"
        >
          {zuhauseKurz(eine.datei)}
        </span>
        <details className="text-sm">
          <summary>{vorschau[eine.art] ?? eigene.zeilen}</summary>
          <pre className="font-mono text-xs whitespace-pre-wrap break-all">
            {eine.zeilen.join('\n')}
          </pre>
        </details>

        {/* **Nur, was etwas tut** (UXR-19, #1927): bis dahin standen in
            jeder der fünf Karten beide Knöpfe, und in den meisten war einer
            gesperrt. Ein fremder Eintrag hat keinen — er bleibt, wie er ist,
            und der Satz darüber sagt es. */}
        {(einrichtbar || widerrufbar) && (
          <div className="flex gap-2">
            {einrichtbar && (
              <Knopf
                angehalten={laeuft}
                aufKlick={() => {
                  handeln(() => api().SchnittstelleEinrichten(eine.name))
                }}
              >
                {eigene.einrichten}
              </Knopf>
            )}
            {widerrufbar && (
              <Knopf
                angehalten={laeuft}
                aufKlick={() => {
                  handeln(() => api().SchnittstelleWiderrufen(eine.name))
                }}
              >
                {eigene.widerrufen}
              </Knopf>
            )}
          </div>
        )}
      </li>
    )
  }

  return (
    <section
      aria-labelledby={ueberschrift}
      className="mt-6 flex flex-col gap-2"
    >
      <h3 id={ueberschrift} className="text-base font-medium">
        {eigene.titel}
      </h3>
      <p className="max-w-prose text-sm text-leise">{eigene.satz}</p>

      {/* **Ohne Ziele kein Block** — es gäbe nichts einzurichten. */}
      {stand !== undefined && stand.schnittstellen.length > 0 && (
        <SchnittstellenAlle
          ziele={stand.schnittstellen}
          ausgaenge={ausgaenge}
          laeuft={laeuft}
          einrichten={() => {
            handelnAlle(() => api().SchnittstellenAlleEinrichten())
          }}
          widerrufen={() => {
            handelnAlle(() => api().SchnittstellenAlleWiderrufen())
          }}
        />
      )}

      {stand !== undefined && (
        <ul className="flex flex-col gap-2">
          {stand.schnittstellen.map(zeile)}
        </ul>
      )}

      {/* **Nur der eigene Eintrag ging** (OPS-09b, #1961): die Datei hatte
          sich seither geändert, und ihr Stand davor kam nicht zurück. */}
      {stand !== undefined && stand.nurEintrag !== '' && (
        <p role="status" className="text-sm">
          {einsetzen(eigene.nurEintrag, {
            name: namen[stand.nurEintrag] ?? stand.nurEintrag,
          })}
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
