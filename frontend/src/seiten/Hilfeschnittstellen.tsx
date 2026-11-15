// **Was aus dieser Anwendung hinausgeht** — die drei Wege ohne Fenster.
//
// Eigene Datei, weil `Hilfe.tsx` mit ihr über 400 Zeilen ginge (§2). Der
// Schnitt folgt der Frage: dort steht, **welche** Gebiete die Hilfe hat, hier
// der Inhalt des längsten.
import { api } from '../api'
import { auspacken } from '../umschlag'
import { Flaeche } from '../bausteine/Flaeche'
import { useLaden } from '../haken/useLaden'
import { einsetzen } from '../sprache/einsetzen'
import { Satz } from '../bausteine/Satz'
import { useTexte } from '../sprache/kontext'
import type { GrenzenAntwort } from '../grenzen'

async function grenzenHolen(): Promise<GrenzenAntwort> {
  const antwort = await api().Grenzen()

  return auspacken(antwort, () => antwort)
}

export function Hilfeschnittstellen() {
  const { hilfe } = useTexte()
  // **Die Grenzen dieses Prozesses** (#1429) — gelesen, nicht abgeschrieben:
  // eine Zahl in einem Text veraltet still, die Auskunft nicht.
  const grenzen = useLaden(grenzenHolen, [])

  return (
    <Flaeche
      ueberschrift={hilfe.schnittstellen.ueberschrift}
      sprungmarke="hilfe-schnittstellen"
    >
      <p className="mb-4 max-w-prose text-leise">
        <Satz text={hilfe.schnittstellen.einleitung} />
      </p>

      <dl className="grid gap-3 sm:grid-cols-[10rem_minmax(0,1fr)]">
        <div className="contents">
          <dt className="font-semibold">{hilfe.schnittstellen.bezugName}</dt>
          <dd className="max-w-prose text-leise">
            <Satz text={hilfe.schnittstellen.bezugSatz} />
          </dd>
        </div>

        {/* **Die beiden anderen Wege nach draußen** — sie fehlten hier,
            während die Einleitung „genau eine Stelle" sagte. Scryfall kam in
            Phase V dazu, die Aktualisierung in Phase Z; angestrichen vom
            Gesamtreview vor 0.2.0. Eine Anwendung, die erklärt, was sie
            hinausgibt, muss alles nennen. */}
        <div className="contents">
          <dt className="font-semibold">{hilfe.schnittstellen.scryfallName}</dt>
          <dd className="max-w-prose text-leise">
            <Satz text={hilfe.schnittstellen.scryfallSatz} />
          </dd>
        </div>

        <div className="contents">
          <dt className="font-semibold">{hilfe.schnittstellen.fassungName}</dt>
          <dd className="max-w-prose text-leise">
            <Satz text={hilfe.schnittstellen.fassungSatz} />
          </dd>
        </div>

        <div className="contents">
          <dt className="font-semibold">{hilfe.schnittstellen.mcpName}</dt>
          <dd className="max-w-prose text-leise">
            <Satz text={hilfe.schnittstellen.mcpSatz} />{' '}
            {/* **Die überraschendste Antwort steht hervorgehoben**: dass es
            keinen Schalter gibt, ist die häufigste Frage. */}
            <strong>
              <Satz text={hilfe.schnittstellen.mcpNichtAn} />
            </strong>
          </dd>
        </div>

        {/* **Der dritte Weg auf diesen Rechner** (#997): der Dienst auf der
            Rückschleife. Er steht neben MCP und nicht darunter — beide sind
            eigene Programme, die niemand von selbst startet. */}
        <div className="contents">
          <dt className="font-semibold">{hilfe.schnittstellen.dienstName}</dt>
          <dd className="max-w-prose text-leise">
            <Satz text={hilfe.schnittstellen.dienstSatz} />{' '}
            <strong>
              <Satz text={hilfe.schnittstellen.dienstNichtAn} />
            </strong>
          </dd>
        </div>

        {/* **Der letzte Weg, und der einzige selbsttätige** (#830). Er kam
            mit der Startmeldung dazu, und der Satz „keine davon läuft von
            selbst" stimmte damit nicht mehr. */}
        <div className="contents">
          <dt className="font-semibold">
            {hilfe.schnittstellen.startfensterName}
          </dt>
          <dd className="max-w-prose text-leise">
            <Satz text={hilfe.schnittstellen.startfensterSatz} />
          </dd>
        </div>

        <div className="contents">
          <dt className="font-semibold">{hilfe.schnittstellen.dateienName}</dt>
          <dd className="max-w-prose text-leise">
            <Satz text={hilfe.schnittstellen.dateienSatz} />
          </dd>
        </div>
        {/* **Die Grenzen** (#1429): was ein Maschinenzugang darf, steht neben
            den Wegen, die er nimmt — und kommt aus derselben Auskunft wie
            `lesen.grenzen`, `mse grenzen` und `GET /grenzen`. */}
        <div className="contents">
          <dt className="font-semibold">{hilfe.schnittstellen.grenzenName}</dt>
          <dd className="max-w-prose text-leise">
            <Satz text={hilfe.schnittstellen.grenzenSatz} />{' '}
            {grenzen.daten !== undefined && (
              <span data-grenzen>
                {einsetzen(hilfe.schnittstellen.grenzenZahlen, {
                  burst: String(grenzen.daten.grenzen.principal.burst),
                  nachfuellen: String(
                    grenzen.daten.grenzen.principal.nachfuellen,
                  ),
                  lesen: String(grenzen.daten.grenzen.kosten.lesen ?? 1),
                  schreiben: String(
                    grenzen.daten.grenzen.kosten.schreiben ?? 2,
                  ),
                  einfuhr: String(grenzen.daten.grenzen.kosten.einfuhr ?? 8),
                  rendern: String(grenzen.daten.grenzen.parallel.rendern ?? 2),
                  reserve: String(grenzen.daten.grenzen.reserve),
                })}
              </span>
            )}
            {/* **Abgeschaltete Kontingente stehen da** (QA-08.AK-03,
                #1739): wer kein Nein bekommt, soll wissen, warum — und was
                trotzdem gilt. */}
            {grenzen.daten?.grenzen.kontingente === 'aus' && (
              <span data-kontingente className="mt-2 block">
                <Satz text={hilfe.schnittstellen.grenzenOhneKontingente} />
              </span>
            )}
            {grenzen.fehler !== undefined && (
              <span role="alert">{grenzen.fehler}</span>
            )}
          </dd>
        </div>
      </dl>

      <p className="mt-4 max-w-prose text-leise">
        <Satz text={hilfe.schnittstellen.sonstSatz} />
      </p>
    </Flaeche>
  )
}
