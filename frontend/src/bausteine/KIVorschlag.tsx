// **Ein Modell schlägt Felder vor** — Werkstück A7 von Phase 5.3
// (#1446),
// seit A8 (#1447)
// mit **Anbieter und Modell aus dem Katalog** statt einem Textfeld.
//
// # Warum eine Wahl und kein getipptes Wort
//
// **Weil der Anbieter dazugehört.** Mit einem war der Alias die ganze
// Auskunft; mit zweien ist er die halbe — beide dürfen denselben Namen
// führen. Und ein getippter Name, den kein Katalog kennt, wird erst beim
// Ruf zu einer Meldung: die Wahl kennt nur, was wirklich bereitsteht.
//
// # Zwei Schritte, und der erste ändert nichts
//
// **Das ist der Satz des Issues**: „Generieren verändert Zielkarte nicht."
// Wer vorschlägt, sieht die Felder; wer übernimmt, schreibt sie. Dazwischen
// liegt ein Blick — und der ist der Zweck der Trennung.
//
// # Die vier Regeln aus §7
//
//   - **Regel 1** löst der `key` am Bauteil: ein Kartenwechsel räumt den
//     offenen Vorschlag mit ab. Wer ihn entfernt, nimmt auch Regel 2 weg.
//   - **Regel 2** ebenso: ein neues Bauteil trägt keine Antwort von vorhin.
//   - **Regel 3**: nach einer geglückten Übernahme wird gemeldet, nicht
//     zurückgenommen.
//   - **Regel 4** hält [useVorgang]: ein zweiter Ruf, während der erste
//     läuft, geht gegen einen Stand, den niemand kennt.
import { useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useKIModelle } from '../haken/useKIModelle'
import { useVorgang } from '../haken/useVorgang'

import { Bestandsfreigabe } from './Bestandsfreigabe'
import { KIKostenfreigabe } from './KIKostenfreigabe'
import { KIModellwahl } from './KIModellwahl'
import { Vorschlagsliste } from './KIVorschlagsliste'
import { aliasAus, anbieterAus, kostetGeld } from './kiwahl'
import { KIEinrichtungsverweis } from './KIEinrichtungsverweis'

import type { KIVorschlagAntwort } from '../antworten'

interface KIVorschlagEigenschaften {
  karteID: string

  // geaendert meldet dem Editor, dass die Karte neu zu lesen ist.
  geaendert: () => void
}

export function KIVorschlag({ karteID, geaendert }: KIVorschlagEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen

  const vorgang = useVorgang()

  // **Ein Schlüssel und nicht zwei Zustände**: `anbieter/alias` ist das,
  // was eine Wahl liefert, und zwei Felder daneben liefen auseinander.
  const [gewaehlt, setzeGewaehlt] = useState('')
  const [auftrag, setzeAuftrag] = useState('')
  const [vorschlag, setzeVorschlag] = useState<KIVorschlagAntwort>()
  const [fehler, setzeFehler] = useState<string>()

  // **Leer, und das ist die Vorgabe** (A13.4, #1452): ohne Frage wird der
  // Index nicht gefragt, und ohne Kreuz bleibt der eigene Bestand außen.
  const [belegfrage, setzeBelegfrage] = useState('')
  const [freigegeben, setzeFreigegeben] = useState<string[]>([])
  const [geschrieben, setzeGeschrieben] = useState<number>()

  // **Der Katalog wird erst gelesen, wenn jemand die Wahl anfasst.**
  //
  // # Warum nicht beim Aufschlagen
  //
  // **Weil `KIModelle` fragen geht.** Ist Anthropic eingeschaltet, führt der
  // Ruf zu `api.anthropic.com` — und das Aufschlagen einer Karte wäre damit
  // ein Ruf nach draußen, den niemand ausgelöst hat. Das widerspricht dem,
  // was die Hilfe zusagt, und §16, Regel 6: nichts geht still hinaus.
  //
  // Der erste Anlauf las beim Aufschlagen, und sechzehn Läufe der
  // Kartenseiten wurden dabei rot — jede von ihnen hätte plötzlich eine
  // Bindung belegen müssen, die sie nie gerufen hat. Das war der Hinweis,
  // nicht der Schaden.
  //
  // # Warum am Aufklappen und nicht am Fokus
  //
  // **Weil Durchtabben keine Absicht ist.** Der zweite Anlauf hing am
  // `onFocus` der Wahl, und die Läufe der Kartenseiten blieben rot: wer mit
  // der Tastatur am Feld vorbeigeht, hat nichts gewollt. Aufklappen hat
  // jemand gewollt — und es ist über die Tastatur genauso erreichbar (§6).
  const [gefragt, setzeGefragt] = useState(false)

  const katalog = useKIModelle(gefragt)

  // offen sagt, ob der Kasten aufgeklappt ist — **und ob sein Inhalt im DOM
  // steht**. `gefragt` bleibt davon unberührt: wer einmal gefragt hat, soll
  // beim zweiten Aufklappen nicht noch einmal fragen.
  const [offen, setzeOffen] = useState(false)

  // freigabe sagt, ob gerade nach einer Freigabe gefragt wird (#1539).
  //
  // **Ein Zwischenschritt und kein zweiter Knopf**: wer »Vorschlagen« drückt,
  // hat einen Vorschlag gewollt — was er noch nicht gesehen hat, ist der
  // Preis. Der Kasten zeigt ihn und fragt; danach geht derselbe Auftrag
  // hinaus, den er ausgelöst hat.
  const [freigabe, setzeFreigabe] = useState(false)

  // ausloesen entscheidet, ob vorher zu fragen ist.
  //
  // **Aus dem Katalog und nicht aus dem Namen** ([kostetGeld]): wo ein
  // Anbieter rechnet, sagt er selbst.
  function ausloesen() {
    if (kostetGeld(katalog.modelle, gewaehlt)) {
      setzeFreigabe(true)

      return
    }

    void vorschlagen('')
  }

  async function vorschlagen(nachweis: string) {
    setzeFehler(undefined)
    setzeGeschrieben(undefined)

    await vorgang.starte(`vorschlag/${karteID}`, async () => {
      try {
        const antwort = await api().KIVorschlagen(karteID, {
          anbieter: anbieterAus(gewaehlt),
          alias: aliasAus(gewaehlt),
          auftrag,
          nachweis,

          // **Die Belegfrage ist wahlfrei** (A13.4, #1452): leer heißt, dass
          // der Suchindex gar nicht gefragt wird. Wo sie steht, sieht das
          // Modell dieselben Belege wie eine Einschätzung.
          belegfrage,
          freigegeben,
        })

        setzeVorschlag(auspacken(antwort, () => antwort))
      } catch (ursache) {
        setzeVorschlag(undefined)
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  async function uebernehmen(kennung: string, felder: string[]) {
    setzeFehler(undefined)

    await vorgang.starte(`uebernahme/${karteID}`, async () => {
      try {
        const antwort = await api().KIVorschlagUebernehmen(kennung, felder)
        const geschriebene = auspacken(antwort, () => antwort.felder)

        setzeGeschrieben(geschriebene)
        setzeVorschlag(undefined)

        // **Erst melden, dann nachladen** (§7, Regel 3): geschrieben ist
        // geschrieben, und ein Lesefehler danach nimmt es nicht zurück.
        geaendert()
      } catch (ursache) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  const laeuft = vorgang.laeuft(`vorschlag/${karteID}`)

  return (
    <details
      className="rounded-sm border border-kante px-3 py-2"
      onToggle={(ereignis) => {
        const auf = ereignis.currentTarget.open

        setzeOffen(auf)

        if (auf) {
          setzeGefragt(true)
        }
      }}
    >
      {/* **Zugeklappt, bis jemand ihn aufschlägt.** Erst dann wird der
          Modellkatalog gelesen — und erst dann geht überhaupt etwas hinaus. */}
      <summary className="cursor-pointer text-sm font-medium focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring">
        {eigene.kiVorschlagTitel}
      </summary>

      {/* **Was zugeklappt ist, steht nicht da.** Ein Browser zeigt den Inhalt
          eines geschlossenen `details` nicht — er hängt aber im DOM, und
          damit fände ihn jede Abfrage über die Rolle. Drei Läufe anderer
          Seiten haben genau das angestrichen: »hier gibt es kein
          Auswahlfeld« war plötzlich falsch, obwohl niemand eines sah. */}
      {offen && (
        <div className="mt-3 flex flex-col gap-3">
          <p className="max-w-prose text-sm text-leise">
            {eigene.kiVorschlagAendertNichts}
          </p>

          <KIModellwahl
            gefragt={gefragt}
            modelle={katalog.modelle}
            laedt={katalog.laedt}
            fehler={katalog.fehler}
            gewaehlt={gewaehlt}
            waehlen={setzeGewaehlt}
          />

          <label className="flex flex-col gap-1 text-sm">
            {eigene.kiVorschlagAuftrag}
            <textarea
              className="rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
              rows={2}
              placeholder={eigene.kiVorschlagAuftragHinweis}
              value={auftrag}
              onChange={(ereignis) => {
                setzeAuftrag(ereignis.target.value)
              }}
            />
          </label>

          {/* **Die Belegfrage ist wahlfrei** (A13.4, #1452): leer heißt,
              dass der Suchindex gar nicht gefragt wird. Wo sie steht, sieht
              das Modell dieselben Stellen wie eine Einschätzung — über
              denselben Dienst und mit demselben Manifest. */}
          <label className="flex flex-col gap-1 text-sm">
            {eigene.kiVorschlagBelegfrage}
            <input
              className="rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
              placeholder={eigene.kiVorschlagBelegfrageHinweis}
              value={belegfrage}
              onChange={(ereignis) => {
                setzeBelegfrage(ereignis.target.value)
              }}
            />
          </label>

          {/* **Die Freigabe steht nur da, wo gesucht wird.** Ohne Belegfrage
              gibt es keine Suche, und ein Kreuz dafür bedeutete nichts. */}
          {belegfrage !== '' && (
            <Bestandsfreigabe
              freigegeben={freigegeben}
              waehlen={setzeFreigegeben}
            />
          )}

          <button
            type="button"
            className="self-start rounded-sm border border-kante px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
            disabled={laeuft || gewaehlt === '' || auftrag === ''}
            onClick={ausloesen}
          >
            {eigene.kiVorschlagen}
          </button>

          {/* **Was Geld kostet, wird vorher freigegeben** (#1539). Der Kasten
          steht zwischen Knopf und Auftrag — und ohne ihn ging am Fenster kein
          einziger kostenpflichtiger Auftrag hinaus. */}
          {freigabe && (
            <KIKostenfreigabe
              anbieter={anbieterAus(gewaehlt)}
              freigegeben={(nachweis) => {
                setzeFreigabe(false)
                void vorschlagen(nachweis)
              }}
              verzichtet={() => {
                setzeFreigabe(false)
              }}
            />
          )}

          {/* **Die Ladezeit wird erklärt, nicht verschwiegen** — beim ersten Ruf
          legt Ollama mehrere Gigabyte in den Speicher. */}
          {laeuft && (
            <p role="status" className="text-sm text-leise">
              {eigene.kiVorschlagLaeuft}
            </p>
          )}

          {/* **Mit dem Weg zur Einrichtung** (AI-01.AK-04, #1742): eine

              KI-Meldung heißt meist, dass ein Anbieter oder Modell fehlt. */}

          {fehler !== undefined && (
            <div className="flex flex-col items-start gap-1">
              <p role="alert" className="text-sm text-fehler">
                {fehler}
              </p>
              <KIEinrichtungsverweis />
            </div>
          )}

          {geschrieben !== undefined && (
            <p role="status" className="text-sm text-leise">
              {einsetzen(eigene.kiVorschlagUebernommen, {
                felder: String(geschrieben),
              })}
            </p>
          )}

          {vorschlag !== undefined && (
            <Vorschlagsliste
              vorschlag={vorschlag}
              eigene={eigene}
              laeuft={vorgang.laeuft(`uebernahme/${karteID}`)}
              uebernehmen={(felder) => {
                void uebernehmen(vorschlag.kennung, felder)
              }}
              verwerfen={() => {
                setzeVorschlag(undefined)
              }}
            />
          )}
        </div>
      )}
    </details>
  )
}
