// **Zurücksetzen benennt genau, was betroffen ist** — Werkstück O4 von Phase
// 5.2 (#1428).
//
// # Warum eine Auswahl und kein Knopf
//
// **Weil „Zurücksetzen" keine Handlung ist, sondern eine Familie.** Das
// Fenster wieder aufräumen und einen Menschen samt allem löschen haben nichts
// miteinander zu tun außer dem Wort; ein Knopf für beides ist die Gestalt, in
// der jemand das eine drückt und das andere bekommt.
//
// # Und warum die Vorschau zwischen Wahl und Tat steht
//
// **Weil eine Zahl eine Entscheidung ist und ein Satz keine.** „Dieses Profil
// und alles, was dazugehört" sagt niemandem, was er verliert.
import { useId, useState } from 'react'

import type { RuecksetzungsvorschauAntwort } from '../apiruecksetzung'
import type { Ruecksetzungslauf as Lauf } from '../haken/useRuecksetzung'
import { alsZeitpunkt } from '../zeitpunkt'
import { api } from '../api'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useLaden } from '../haken/useLaden'
import { useRuecksetzung } from '../haken/useRuecksetzung'
import { useDarstellung, useTexte, useZeitzone } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'
import { useSchreibsperre } from '../haken/schreibrecht'
import { kennungseingabe } from '../bausteine/kennungseingabe'

const knopf =
  'rounded-md border border-kante px-3 py-1.5 text-sm font-medium ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring disabled:opacity-60 ' +
  'aria-disabled:cursor-not-allowed aria-disabled:opacity-50'

const feld =
  'rounded-md border border-kante bg-grund px-2 py-1.5 text-sm ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring'

// Ruecksetzung holt sich ihren Stand selbst.
//
// **Ohne Durchreichen des Stands** — dieselbe Gestalt wie `Platz` und
// `Zugaenge` daneben: sie steht in **einem** Gebiet einer Seite, und ein
// Haken, der durch fünf Bauteile gereicht würde, machte vier davon zu
// Trägern einer Frage, die sie nicht stellen. **Nach oben meldet sie** über
// zwei Rückrufe aus der Bestandspflege, die das Gebiet ohnehin trägt (#2292).
export function Ruecksetzung({
  gewechselt,
  veraendert,
}: {
  // gewechselt meldet, dass das Bezugsobjekt fort ist — Neuanfang, Profil,
  // Mensch.
  gewechselt: () => void

  // veraendert meldet, dass sich geändert hat, was das Fenster gelesen hat.
  veraendert: () => void
}) {
  const texte = useTexte()
  const worte = texte.seiten.ruecksetzung
  const titelID = useId()
  const wahlID = useId()

  // **Die Reihe kommt von der Go-Seite** und steht nicht hier: sie ist dort
  // die Reihenfolge der Schwere, und eine zweite Aufzählung liefe an dem Tag
  // auseinander, an dem eine Art dazukommt.
  const reihe = useLaden(holenArten, [])
  const arten = reihe.daten ?? []

  const stand = useRuecksetzung(useVorgang(), gewechselt, veraendert)

  const [art, setzeArt] = useState('')

  // **Eine Stelle, an der „noch nichts gewählt" zur ersten Art wird.** Zwei
  // wären zwei Antworten auf dieselbe Frage — und die zweite fiele beim
  // nächsten Umbau auf.
  const gewaehlt = art === '' ? (arten[0] ?? '') : art

  return (
    <section aria-labelledby={titelID} className="max-w-xl">
      {/* **Derselbe Rang wie die Nachbarn**: das Gebiet „Bestand" führt vier
          Abschnitte — Sicherung, Rückkehr, Aufräumen, Zurücksetzen. Als `h3`
          läse sich der letzte wie ein **Teil** des Aufräumens — im Bild und
          in der Vorlesehilfe. */}
      <h2 id={titelID} className="text-lg font-semibold tracking-tight">
        {worte.titel}
      </h2>

      <p className="mt-1 max-w-prose text-sm text-leise">{worte.hinweis}</p>

      <div className="mt-3 flex flex-wrap items-end gap-2">
        <label className="flex flex-col gap-1 text-sm" htmlFor={wahlID}>
          {worte.artWaehlen}

          <select
            id={wahlID}
            className={feld}
            value={gewaehlt}
            onChange={(ereignis) => {
              setzeArt(ereignis.target.value)

              // **Ein Wechsel räumt die Vorschau ab** (§7, Regel 1): sie
              // beschreibt eine andere Rücksetzung, und eine
              // stehengebliebene wäre die gefährlichste Antwort.
              stand.raeumen()
            }}
          >
            {arten.map((eine) => (
              <option key={eine} value={eine}>
                {worte.arten[eine] ?? eine}
              </option>
            ))}
          </select>
        </label>

        <button
          type="button"
          className={knopf}
          disabled={stand.laeuft || arten.length === 0}
          onClick={() => {
            stand.nachsehen(gewaehlt)
          }}
        >
          {stand.laeuft ? worte.nachsehenLaeuft : worte.nachsehen}
        </button>
      </div>

      {stand.vorschau !== undefined && (
        <Vorschau
          key={stand.vorschau.art}
          stand={stand}
          vorschau={stand.vorschau}
        />
      )}

      {stand.bilanz !== undefined && (
        <div className="mt-4 rounded-md border border-kante p-3">
          <p className="text-sm">
            {einsetzen(worte.bilanz, {
              anzahl: String(stand.bilanz.zurueckgesetzt),
            })}
          </p>

          {stand.bilanz.beiseite !== '' && (
            <p className="mt-1 text-sm break-all text-leise">
              {einsetzen(worte.beiseite, { ort: stand.bilanz.beiseite })}
            </p>
          )}
        </div>
      )}

      {(stand.fehler ?? reihe.fehler) !== undefined && (
        <p role="alert" className="mt-4 text-fehler">
          {texte.bausteine.zustaende.fehler} {stand.fehler ?? reihe.fehler}
        </p>
      )}
    </section>
  )
}

// holenArten holt die sechs Arten in der Reihe ihrer Schwere.
async function holenArten(): Promise<string[]> {
  return api().Ruecksetzungsarten()
}

// Vorschau zeigt, was betroffen wäre, und nimmt die Bestätigung entgegen.
//
// **Sie bekommt die Vorschau als eigene Eigenschaft** — dieselbe Entscheidung
// wie bei der Datenpflege: als Eigenschaft ist die Zusage im Typ statt in
// einem `if`, den kein Aufruf betritt (§3).
function Vorschau({
  stand,
  vorschau,
}: {
  stand: Lauf
  vorschau: RuecksetzungsvorschauAntwort
}) {
  const worte = useTexte().seiten.ruecksetzung
  const nameID = useId()

  const [name, setzeName] = useState('')
  const [nachweis, setzeNachweis] = useState('')

  return (
    <div className="mt-4 flex flex-col gap-3">
      {vorschau.betroffen.length === 0 ? (
        <p className="text-sm text-leise">{worte.nichtsBetroffen}</p>
      ) : (
        <ul className="flex flex-col gap-1 text-sm">
          {vorschau.betroffen.map((gruppe) => (
            <li key={gruppe.was}>
              {einsetzen(worte.gruppe, {
                was: worte.gruppen[gruppe.was] ?? gruppe.was,
                anzahl: String(gruppe.anzahl),
              })}

              {gruppe.namen.length > 0 && (
                <span className="text-leise"> · {gruppe.namen.join(', ')}</span>
              )}
            </li>
          ))}
        </ul>
      )}

      {vorschau.sicherungEmpfohlen && (
        <p className="max-w-prose text-sm text-leise">{worte.sicherung}</p>
      )}

      {vorschau.hindernisse.length > 0 && (
        <ul className="flex flex-col gap-1 text-sm text-fehler">
          {vorschau.hindernisse.map((hindernis) => (
            <li key={hindernis.art + hindernis.stelle}>
              {worte.hindernisse[hindernis.art] ?? hindernis.art}
              {hindernis.stelle !== '' && <span> · {hindernis.stelle}</span>}
            </li>
          ))}
        </ul>
      )}

      {vorschau.bereit && (
        <>
          {/* **Das Namensfeld nur, wo ein Name die Bindung ist** — bei einer
              Wahlrücknahme gibt es nichts abzutippen, und ein leeres Feld
              darüber sähe aus wie eine Hürde, die es nicht gibt. */}
          {vorschau.nameNoetig && (
            <label className="flex flex-col gap-1 text-sm" htmlFor={nameID}>
              {worte.nameTippen}

              <input
                {...kennungseingabe}
                id={nameID}
                className={feld}
                value={name}
                onChange={(ereignis) => {
                  setzeName(ereignis.target.value)
                }}
              />
            </label>
          )}

          {/* **Die Freigabe nur, wo eine verlangt wird** — und als Knopf,
              nicht als Feld (G12, #2282): die Kennung stand nirgends, und
              ausgestellt hat sie bis dahin auch niemand. */}
          {vorschau.nachweisNoetig && (
            <Neuanfangsfreigabe laeuft={stand.laeuft} erteilt={setzeNachweis} />
          )}

          <div>
            <button
              type="button"
              className={knopf}
              disabled={
                stand.laeuft ||
                (vorschau.nameNoetig && name === '') ||
                (vorschau.nachweisNoetig && nachweis === '')
              }
              onClick={() => {
                stand.zuruecksetzen(
                  vorschau.art,
                  name,
                  nachweis,
                  vorschau.revision,
                )
              }}
            >
              {stand.laeuft ? worte.tunLaeuft : worte.tun}
            </button>
          </div>
        </>
      )}
    </div>
  )
}

// Neuanfangsfreigabe holt die Freigabe eines Menschen für **einen** Neuanfang.
//
// **Die Kennung bleibt hier und reist nur zum Knopf darunter** — sie ist kein
// Wissen, das jemand braucht; gezeigt wird, bis wann sie gilt. Ein Wechsel
// der Art räumt sie mit der Vorschau ab (`key` an der Vorschau, §7).
function Neuanfangsfreigabe({
  laeuft,
  erteilt,
}: {
  laeuft: boolean
  erteilt: (kennung: string) => void
}) {
  const texte = useTexte()
  const worte = texte.seiten.ruecksetzung
  const darstellung = useDarstellung()
  const zeitzone = useZeitzone()

  // **Ein Leser erteilt keine Freigabe** — Befund 4 des Zweitblicks über
  // Phase G: der Dienst verlangt »Mitarbeiter«, und bis dahin ruhte der Knopf
  // nur während des Ladens.
  const schreibt = useSchreibsperre(true)

  const [erteilte, setzeErteilte] = useState<{ kennung: string; bis: string }>()
  const [holt, setzeHolt] = useState(false)
  const [fehler, setzeFehler] = useState<string>()

  async function erteilen() {
    // **Eine neue Freigabe verwirft die alte zuerst** — sonst stünde
    // „Zurücksetzen" während des Holens mit der vorigen Kennung offen.
    erteilt('')
    setzeErteilte(undefined)
    setzeFehler(undefined)
    setzeHolt(true)

    try {
      const antwort = await api().RuecksetzungFreigabeErteilen()

      const kennung = auspacken(antwort, () => antwort.kennung)

      erteilt(kennung)
      setzeErteilte({ kennung, bis: antwort.laeuftab })
    } catch (ursache) {
      setzeFehler(meldungAus(ursache, texte))
    } finally {
      setzeHolt(false)
    }
  }

  return (
    <div className="flex flex-col gap-2 text-sm">
      <p className="max-w-prose text-leise">{worte.freigabeSatz}</p>

      <div>
        <button
          type="button"
          {...schreibt.knopf}
          title={schreibt.titel}
          className={knopf}
          disabled={laeuft || holt}
          onClick={() => {
            if (!schreibt.gesperrt) {
              void erteilen()
            }
          }}
        >
          {holt ? worte.freigebenLaeuft : worte.freigeben}
          {schreibt.satz}
        </button>
      </div>

      {erteilte !== undefined && (
        <div role="status" className="flex flex-col gap-1">
          <p>
            {einsetzen(worte.freigegeben, {
              zeit: alsZeitpunkt(erteilte.bis, darstellung, zeitzone),
            })}
          </p>

          {/* **Die Kennung steht da, und das ist der Weg zur Kommandozeile**:
              ausgestellt wird nur hier, eingelöst auch dort — ohne sie zu
              zeigen, gäbe es den zweiten Weg nur dem Namen nach. */}
          <p className="text-leise">{worte.kennungFuerKommandozeile}</p>
          <code className="break-all select-all">{erteilte.kennung}</code>
        </div>
      )}

      {fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {fehler}
        </p>
      )}
    </div>
  )
}
