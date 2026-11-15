// **Das Diagnosepaket im Fenster** — Werkstück O10 von Phase 5.2
// (#1434).
//
// # Drei Schritte, und jeder ist eine Handlung
//
// **Zusammenstellen, ansehen, sichern.** Dieselbe Ordnung wie bei der
// Diagnose darüber und aus demselben Grund: eine Anwendung, die von selbst
// etwas hinausschickte, wäre etwas anderes. Der vierte Knopf nimmt es wieder
// weg — wer es sich anders überlegt, soll das Paket loswerden, ohne den
// Finder zu suchen.
//
// # Was hier zu sehen sein muss
//
// **Jeder Teil mit seiner Größe, der Zeitraum und die Grenze** — vor der
// Zustimmung. „Ein Paket" ohne Zahl daneben ist keine Auskunft, und wer
// zustimmt, soll wissen, über welche Zeit er Auskunft gibt.
//
// **Und ein ausgelassener Teil steht da, nicht weggelassen**: sonst vermisst
// ihn niemand und keiner weiß, warum er fehlt.
import { useState } from 'react'

import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Knopf } from '../bausteine/Knopf'
import { api } from '../fassade'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { useDarstellung, useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

import type { DiagnosepaketAntwort } from '../apidiagnosepaket'

export function Diagnosepaket() {
  const texte = useTexte()
  const darstellung = useDarstellung()
  const eigene = texte.seiten.technik.paket
  const vorgang = useVorgang()

  const [vorschau, setzeVorschau] = useState<DiagnosepaketAntwort>()
  const [geschrieben, setzeGeschrieben] = useState(false)

  // ohne sind die Teile, die jemand vor dem Sichern herausgenommen hat —
  // OPS-04.AK-04 (#1721). Ein neues Zusammenstellen nimmt wieder alle.
  const [ohne, setzeOhne] = useState<ReadonlySet<string>>(new Set())
  const [fehler, setzeFehler] = useState<string>()

  // fuehre kapselt den immer gleichen Rahmen: **ein Vorgang je Paket** (§7,
  // Regel 4), der Fehler am Ende sichtbar.
  const fuehre = (arbeit: () => Promise<void>) =>
    vorgang.starte('diagnosepaket', async () => {
      try {
        await arbeit()
        setzeFehler(undefined)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  const zusammenstellen = () =>
    fuehre(async () => {
      const antwort = await api().Diagnosepaket()

      setzeVorschau(auspacken(antwort, () => antwort))
      setzeGeschrieben(false)
      setzeOhne(new Set())
    })

  const sichern = () =>
    fuehre(async () => {
      auspacken(await api().DiagnosepaketSichern([...ohne]), () => null)
      setzeGeschrieben(true)
    })

  const loeschen = () =>
    fuehre(async () => {
      auspacken(await api().DiagnosepaketLoeschen(), () => null)
      setzeGeschrieben(false)
    })

  return (
    <section className="flex flex-col gap-3 border-t border-kante pt-4">
      <h3 className="text-lg font-semibold">{eigene.ueberschrift}</h3>

      <Erklaerungsabsatz text={eigene.erklaerung} klasse="text-sm text-leise" />

      {/* **Was drinsteht, steht vor dem Knopf** — dieselbe Ordnung wie bei
          der Diagnose darüber. */}
      <p className="text-sm text-leise">{eigene.inhalt}</p>

      <div className="flex flex-wrap gap-2">
        <Knopf
          bezeichnung={eigene.zusammenstellen}
          aufKlick={() => {
            void zusammenstellen()
          }}
        >
          {eigene.zusammenstellen}
        </Knopf>

        {vorschau !== undefined && (
          <Aktionsknopf
            aktion="sichern"
            bezeichnung={eigene.sichern}
            aufKlick={() => {
              void sichern()
            }}
          >
            {eigene.sichern}
          </Aktionsknopf>
        )}

        {geschrieben && (
          <Aktionsknopf
            aktion="loeschen"
            bezeichnung={eigene.loeschen}
            aufKlick={() => {
              void loeschen()
            }}
          >
            {eigene.loeschen}
          </Aktionsknopf>
        )}
      </div>

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      {vorschau !== undefined && (
        <Paketvorschau
          vorschau={vorschau}
          texte={eigene}
          darstellung={darstellung}
          ohne={ohne}
          umschalten={(teil) => {
            setzeOhne((bisher) => {
              const neu = new Set(bisher)
              if (!neu.delete(teil)) {
                neu.add(teil)
              }

              return neu
            })
          }}
        />
      )}
    </section>
  )
}

// Paketvorschau zeigt die Teile, den Zeitraum und die Größe.
function Paketvorschau({
  vorschau,
  texte,
  darstellung,
  ohne,
  umschalten,
}: {
  vorschau: DiagnosepaketAntwort
  texte: ReturnType<typeof useTexte>['seiten']['technik']['paket']
  darstellung: string
  ohne: ReadonlySet<string>
  umschalten: (teil: string) => void
}) {
  return (
    <>
      <p className="text-sm">
        {texte.groesse
          .replace('{bytes}', alsGroesse(vorschau.bytes, darstellung))
          .replace('{grenze}', alsGroesse(vorschau.grenze, darstellung))}
      </p>

      {/* **Der Zeitraum steht nur da, wenn es einen gibt** — ohne
          Protokollzeilen wäre „von leer bis leer" eine Zeile ohne Aussage. */}
      {vorschau.von !== '' && (
        <p className="text-sm">
          {texte.zeitraum
            .replace('{von}', vorschau.von)
            .replace('{bis}', vorschau.bis)}
        </p>
      )}

      <table className="w-full max-w-2xl text-sm">
        <caption className="sr-only">{texte.ueberschrift}</caption>

        <thead>
          <tr className="text-left">
            <th scope="col">{texte.spalteMitnehmen}</th>
            <th scope="col">{texte.spalteTeil}</th>
            <th scope="col">{texte.spalteDatei}</th>
            <th scope="col" className="text-right">
              {texte.spalteZeilen}
            </th>
            <th scope="col" className="text-right">
              {texte.spalteBytes}
            </th>
          </tr>
        </thead>

        <tbody>
          {vorschau.teile.map((teil) => (
            <tr key={teil.teil}>
              <td>
                <input
                  type="checkbox"
                  checked={!ohne.has(teil.teil)}
                  aria-label={texte.mitnehmen.replace(
                    '{teil}',
                    namenVon(texte.teile, teil.teil),
                  )}
                  onChange={() => {
                    umschalten(teil.teil)
                  }}
                  className="size-4 accent-marke focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
                />
              </td>
              <td>{namenVon(texte.teile, teil.teil)}</td>
              <td className="font-mono text-xs">{teil.datei}</td>
              <td className="text-right tabular-nums">{teil.zeilen}</td>
              <td className="text-right tabular-nums">
                {alsGroesse(teil.bytes, darstellung)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {/* **Ausgelassenes wird genannt.** Ein Teil, der fehlt und den niemand
          vermisst, ist eine stille Lücke. */}
      {vorschau.teile
        .filter((teil) => teil.grund !== '')
        .map((teil) => (
          <p key={teil.teil} className="text-sm text-leise">
            {texte.ausgelassen.replace(
              '{teil}',
              namenVon(texte.teile, teil.teil),
            )}{' '}
            {teil.grund}
          </p>
        ))}
    </>
  )
}

// namenVon sucht den lesbaren Namen eines Teils.
//
// **Übersetzt und nicht roh** — dieselbe Entscheidung wie beim
// Wirkungszähler eine Überschrift weiter oben, und aus demselben Grund: die
// Spalte daneben nennt schon die Datei; der Bezeichner darüber wäre ihre
// Wiederholung. **Der Rückfall ist der Bezeichner selbst**: ein Teil, den
// dieses Fenster noch nicht kennt, ist immer noch eine Auskunft.
function namenVon(namen: Record<string, string>, bezeichner: string): string {
  return namen[bezeichner] ?? bezeichner
}

// alsGroesse schreibt eine Bytezahl so, wie jemand sie liest.
//
// **Weil „8388608" keine Auskunft ist.** Wer zustimmen soll, muss die Zahl
// einordnen können, und dafür ist ein Vielfaches von 1024 der falsche
// Maßstab. Unter einem Kilobyte bleibt die rohe Zahl stehen: „0,0 KB" wäre
// weniger genau, nicht mehr.
export function alsGroesse(bytes: number, darstellung: string): string {
  if (bytes < 1024) {
    return `${String(bytes)} B`
  }

  // **Das Komma gehört zur Darstellung**, nicht zur Zahl: `toFixed`
  // schreibt immer einen Punkt, und „23.6 KB" ist in einer deutschen
  // Oberfläche falsch. Die Darstellung ist eine **Einstellung** dieser
  // Anwendung (LOC-03.AK-03, #1689) und kann von der Sprache der Texte
  // abweichen — `de-CH` schreibt »23.6 KB«.
  const zahl = (wert: number) =>
    wert.toLocaleString(darstellung, {
      minimumFractionDigits: 1,
      maximumFractionDigits: 1,
    })

  if (bytes < 1024 * 1024) {
    return `${zahl(bytes / 1024)} KB`
  }

  return `${zahl(bytes / (1024 * 1024))} MB`
}
