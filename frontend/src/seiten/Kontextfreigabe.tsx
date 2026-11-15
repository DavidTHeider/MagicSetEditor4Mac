// **Was an ein Modell ginge** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Zusammenstellen ist nicht senden
//
// **Dieser Knopf ruft kein Modell.** Er stellt zusammen und zeigt: welche
// Ausschnitte, wie viele Tokens, welche Datenkategorien, und **wo gerechnet
// würde**. Das Issue verlangt genau das vor einer Generierung: „konkreten
// Provider, Datenkategorien und Umfang des **zusammengestellten** Kontexts
// samt Kostenlimit zeigen".
//
// **Nach dem Senden ist diese Frage nicht mehr zu stellen.** Deshalb steht
// sie davor und nicht daneben.
import { useEffect, useRef, useState } from 'react'

import { Knopf } from '../bausteine/Knopf'
import { Zahlenfeld } from '../bausteine/Zahlenfeld'
import { api } from '../fassade'
import type { KontextpaketAntwort } from '../apirecherche'
import { useLauf } from '../haken/useLauf'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { feldklassen } from './recherchefeld'
import { auspacken } from '../umschlag'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

interface FreigabeEigenschaften {
  fall: string
  frage: string

  // freigegeben sind die privaten Korpora, die ausdrücklich gewählt wurden —
  // A13.4 (#1452).
  //
  // **Sie reisen mit**, damit das gezeigte Paket **dasselbe** ist, das
  // hinausginge. Ein Paket, das hier mit eigenen Karten stünde und ohne sie
  // gesendet würde, zeigte etwas anderes, als es schickt — und umgekehrt.
  freigegeben: string[]
}

// kianbieter sind die vier, an die ein Paket gehen kann — dieselben, die
// `internal/app/kiwerkzeug.go` kennt: einer rechnet hier, drei anderswo.
const kianbieter = ['ollama', 'anthropic', 'openai', 'xai'] as const

export function Kontextfreigabe({
  fall,
  frage,
  freigegeben,
}: FreigabeEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.recherche
  const vorgang = useVorgang()

  // **Eine Antwort gehört zu dem Stand, aus dem sie stammt** (§7, Regel 2;
  // F03, #2269). Die Felder bleiben änderbar, während der Ruf läuft; ohne
  // diesen Zähler trüge seine Antwort das Paket danach wieder ein — auch zu
  // einem Budget, das gar nicht mehr gilt, und der Knopf daneben ist dann
  // angehalten.
  const lauf = useLauf()

  const [aufgabe, setzeAufgabe] = useState('')
  const [anbieter, setzeAnbieter] = useState('ollama')
  const [modell, setzeModell] = useState('')

  // **Fenster und Ausgabereserve stehen als Felder da und nicht im Code.**
  //
  // **Weil sie zum Zielmodell gehören und nicht zu dieser Anwendung**: eine
  // Zahl, die hier fest stünde, gälte für kein Modell — und ein Paket, das
  // das Fenster sprengt, schneidet das Modell still ab. Die Vorbelegung ist
  // ein **sichtbarer** Anfangswert, den jeder ändern kann; eine unsichtbare
  // wäre die Behauptung, die hier nicht hingehört.
  const [fenster, setzeFenster] = useState(8192)
  const [reserve, setzeReserve] = useState(1024)

  // **Und sie sind ganze Zahlen** (F03, #2245): `Number(...)` machte aus
  // einem leeren Feld eine Null und gab „8192.5" an ein `int`. Was ein
  // zulässiges Budget ist, sagt weiter `service.pruefeKontextwunsch`.
  const [fensterGilt, setzeFensterGilt] = useState(true)
  const [reserveGilt, setzeReserveGilt] = useState(true)
  // **Paket und Fehler tragen den Auftrag, zu dem sie gehören** (Befund 5
  // des Nachlaufs über J05, #2269): Fall, Frage und Freigabe kommen von oben,
  // und gezeigt wird nur, was zum heutigen passt. **Verglichen wird der
  // Inhalt**, nicht die Liste — die Seite darüber baut sie bei jedem
  // Zeichnen neu.
  const auftrag = JSON.stringify([fall, frage, freigegeben])
  const [ergebnis, setzeErgebnis] = useState<{
    auftrag: string
    paket?: KontextpaketAntwort
    fehler?: string
  }>()
  const paket = ergebnis?.auftrag === auftrag ? ergebnis.paket : undefined
  const fehler = ergebnis?.auftrag === auftrag ? ergebnis.fehler : undefined

  // **Der Auftrag, den es jetzt gibt**, für eine Antwort, die später kommt.
  const auftragJetzt = useRef(auftrag)
  useEffect(() => {
    auftragJetzt.current = auftrag
  }, [auftrag])

  // eingabeGeaendert räumt das Paket **und zählt den Stand weiter**.
  //
  // **Beides zugleich** (§7, Regel 1 und 2): wer nur räumt, lässt die
  // Antwort eines Rufs, der noch unterwegs ist, das Paket zurückbringen.
  const eingabeGeaendert = () => {
    lauf.naechster()
    setzeErgebnis(undefined)
  }

  const zusammenstellen = () => {
    // **`nummer` und nicht `naechster`**: ein Ruf wechselt nichts, er merkt
    // sich nur, für welchen Stand er fragt.
    const gefragt = lauf.nummer()
    const gefragterAuftrag = auftrag

    // gilt sagt, ob die Antwort noch zu dem gehört, was jetzt dasteht —
    // dieselben Eingaben **und** derselbe Auftrag von oben.
    const gilt = () =>
      lauf.istAktuell(gefragt) && auftragJetzt.current === gefragterAuftrag

    return vorgang.starte('kontextpaket', async () => {
      try {
        const antwort = await api().KontextZusammenstellen(
          aufgabe,
          '',
          fall,
          frage,
          20,
          anbieter,
          modell,
          fenster,
          reserve,
          0,
          freigegeben,
        )

        const gepackt = auspacken(antwort, () => antwort)

        if (!gilt()) {
          return
        }

        setzeErgebnis({ auftrag: gefragterAuftrag, paket: gepackt })
      } catch (ursache: unknown) {
        // **Auch der Fehler gehört zu seinem Stand**: eine Meldung zu einer
        // Eingabe, die niemand mehr hat, sieht aus wie die zu der heutigen.
        if (gilt()) {
          setzeErgebnis({
            auftrag: gefragterAuftrag,
            fehler: meldungAus(ursache, texte),
          })
        }
      }
    })
  }

  return (
    // **Ohne eigene Überschrift** — seit UXR-05 (#1913) steht es in einem
    // Aufklapper der Recherche, und der trägt den Titel als Knopf. Zwei
    // Namen für dieselbe Sache untereinander wären einer zu viel.
    <div className="flex flex-col gap-3">
      <Erklaerungsabsatz
        text={eigene.freigabeErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      <div className="flex flex-wrap items-end gap-3">
        <label className="flex max-w-md flex-1 flex-col gap-1 text-sm">
          {eigene.freigabeAufgabe}

          <input
            className={feldklassen}
            value={aufgabe}
            onChange={(ereignis) => {
              setzeAufgabe(ereignis.target.value)
              eingabeGeaendert()
            }}
          />
        </label>

        <label className="flex flex-col gap-1 text-sm">
          {eigene.freigabeAnbieter}

          {/* **Gewählt und nicht getippt** — UX-06.AK-01 (#1678): es gibt
              vier Anbieter, und ein Tippfehler ergab ein Paket für einen,
              den es nicht gibt. Das Modell bleibt frei: seine Namen ändern
              sich schneller als diese Anwendung. */}
          <select
            className={feldklassen}
            value={anbieter}
            onChange={(ereignis) => {
              // **Ein Wechsel räumt das Paket mit** (§7, Regel 1): ein
              // anderes Zielmodell ist ein anderes Budget, und eine
              // stehengebliebene Zahl gälte für einen anderen Anbieter.
              setzeAnbieter(ereignis.target.value)
              eingabeGeaendert()
            }}
          >
            {kianbieter.map((einer) => (
              <option key={einer} value={einer}>
                {einer}
              </option>
            ))}
          </select>
        </label>

        <label className="flex flex-col gap-1 text-sm">
          {eigene.freigabeModell}

          <input
            className={feldklassen}
            value={modell}
            onChange={(ereignis) => {
              setzeModell(ereignis.target.value)
              eingabeGeaendert()
            }}
          />
        </label>

        {/* **Ein anderes Budget ist ein anderes Paket** (§7, Regel 1) —
            auch eines, das noch keine Zahl ist: deshalb räumt `gilt`, das
            bei jeder Eingabe kommt, und nicht `geaendert`. */}
        <Zahlenfeld
          beschriftung={eigene.freigabeFenster}
          wert={fenster}
          kleinstens={0}
          ganzzahlig
          geaendert={setzeFenster}
          gilt={(gilt) => {
            setzeFensterGilt(gilt)
            eingabeGeaendert()
          }}
        />

        <Zahlenfeld
          beschriftung={eigene.freigabeReserve}
          wert={reserve}
          kleinstens={0}
          ganzzahlig
          geaendert={setzeReserve}
          gilt={(gilt) => {
            setzeReserveGilt(gilt)
            eingabeGeaendert()
          }}
        />

        <Knopf
          gewicht="leise"
          aufKlick={() => void zusammenstellen()}
          angehalten={
            vorgang.laeuft('kontextpaket') || !fensterGilt || !reserveGilt
          }
        >
          {eigene.freigabeZeigen}
        </Knopf>
      </div>

      {paket && <Freigabeschau paket={paket} />}

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}
    </div>
  )
}

// Freigabeschau zeigt, was hinausginge.
function Freigabeschau({ paket }: { paket: KontextpaketAntwort }) {
  const eigene = useTexte().seiten.recherche

  return (
    <div className="flex flex-col gap-2 text-sm">
      {/* **Der Ort steht zuerst**, weil er die Frage ist, an der alles
          hängt: bleibt es hier, oder verlässt es diesen Rechner? */}
      <p
        className={
          paket.freigabe.ort === 'lokal' ? 'text-leise' : 'font-semibold'
        }
      >
        {paket.freigabe.ort === 'lokal'
          ? eigene.freigabeOrtLokal
          : eigene.freigabeOrtCloud}
      </p>

      {/* **Die Adresse steht daneben, wo eine eingestellt ist** (A18,
          #1533): hinter „openai" kann ein Gateway stehen, und das ist ein
          Datenexport an einen Dritten — auch wenn dahinter derselbe Anbieter
          rechnet. Leer heißt: die Vorgabe des Anbieters, und dann sagt der
          Name alles. */}
      {paket.freigabe.zieladresse !== '' && (
        <p className="text-leise">
          {eigene.freigabeZieladresse}: {paket.freigabe.zieladresse}
        </p>
      )}

      <p className="text-leise">
        {einsetzen(eigene.freigabeBudget, {
          verbraucht: String(paket.verbraucht),
          grenze: String(paket.kontextgrenze),
          reserve: String(paket.ausgabereserve),
        })}{' '}
        ·{' '}
        {einsetzen(eigene.freigabeZaehlweise, {
          zaehlweise: paket.zaehlweise,
        })}
      </p>

      <p className="text-leise">
        {eigene.freigabeKategorien}: {paket.freigabe.kategorien.join(' · ')} ·{' '}
        {paket.freigabe.privat
          ? eigene.freigabePrivat
          : eigene.freigabeOeffentlich}
      </p>

      <p className="text-xs text-leise">
        {einsetzen(eigene.freigabeLaeuftAb, { zeit: paket.laeuftab })}
      </p>

      {paket.belege.length === 0 ? (
        <p className="max-w-prose text-leise">{eigene.freigabeKeineBelege}</p>
      ) : (
        <ul className="divide-y divide-kante">
          {paket.belege.map((beleg) => (
            <li key={beleg.kennung} className="flex flex-col gap-1 py-2">
              <span className="font-semibold">
                [{beleg.kennung}] {beleg.titel}
              </span>

              {/* **Der Ausschnitt und nicht der ganze Abschnitt**: was hier
                  steht, ist genau das, was ein Modell bekäme. */}
              <span className="text-leise">{beleg.ausschnitt}</span>

              <span className="text-xs text-leise">
                {beleg.quelle} · {beleg.quellrevision}{' '}
                {beleg.veraltet && `· ${eigene.freigabeVeraltet}`}
              </span>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
