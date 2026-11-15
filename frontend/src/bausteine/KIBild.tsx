// **Ein Modell erzeugt ein Kartenbild** — Werkstück A12 von Phase 5.3
// (#1451).
//
// # Vier Zustände, und nur der letzte schreibt
//
// **Das ist derselbe Satz wie beim Textvorschlag** (A7), eine Stufe strenger:
// erzeugen legt in den Vorrat, die Vorschau zeigt, was dort liegt, der
// Zuschnitt geht durch **denselben** Rahmen wie eine Datei von der Platte —
// und erst das Einsetzen berührt die Karte.
//
// # Und die Freigabe gilt nicht doppelt
//
// **Ein Bildauftrag wird gesondert freigegeben.** Das Issue verlangt es
// („Kosten-/Datenfreigabe für Text- und Bildschritt getrennt"), und es ist
// keine Regel dieses Bauteils: ein Nachweis gilt einem Auftrag und ist
// danach eingelöst. Der Satz darüber sagt es, damit niemand überrascht ist.
//
// # Die vier Regeln aus §7
//
//   - **Regel 1 und 2** löst der `key` am [KIBereich]: ein Kartenwechsel
//     räumt den offenen Vorschlag mit ab.
//   - **Regel 3**: nach einer geglückten Übernahme wird gemeldet, nicht
//     zurückgenommen — und der Vorschlag fällt erst danach.
//   - **Regel 4** hält [useVorgang].
import { useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useKIModelle } from '../haken/useKIModelle'
import { useVorgang } from '../haken/useVorgang'

import { Knopf } from './Knopf'
import { KIBildnachweis } from './KIBildnachweis'
import { KIBildvorschau } from './KIBildvorschau'
import { KIKostenfreigabe } from './KIKostenfreigabe'
import { KIModellwahl } from './KIModellwahl'
import { aliasAus, anbieterAus, kostetGeld } from './kiwahl'

import type { KIBildvorschlagAntwort } from '../antworten'

interface KIBildEigenschaften {
  karteID: string

  // bildfelder sind die Bildfelder dieser Karte — leer heißt: es gibt keines.
  bildfelder: string[]

  // zuschneiden öffnet das vorhandene Zuschnittfenster für einen wartenden
  // Vorschlag ([useBildfeld.vonKI]).
  //
  // **Als Aufruf und nicht als eigenes Fenster**: das Issue verlangt
  // „Übernahme nutzt vorhandene Crop-/Ablage-/Renderfunktion".
  zuschneiden: (feld: string, kennung: string) => Promise<boolean>

  // stand zählt, was über den Bildweg geschrieben wurde — danach liest der
  // Bildnachweis neu.
  //
  // **Er kommt von außen**, weil nur der Editor weiß, wann eingesetzt wurde:
  // dieses Bauteil öffnet den Zuschnitt, und was danach geschieht, entscheidet
  // sich dort ([useBildfeld]).
  stand: number
}

// dieGroessen und dieGueten stehen hier und nicht im Katalog.
//
// **Weil es Werte sind und keine Texte** (§9): `1024x1536` und `medium` gehen
// so an den Anbieter. Übersetzt wird die Beschriftung daneben, nicht das
// Wort.
const dieGroessen = ['1024x1536', '1024x1024', '1536x1024'] as const
const dieGueten = ['low', 'medium', 'high'] as const

export function KIBild({
  karteID,
  bildfelder,
  zuschneiden,
  stand,
}: KIBildEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen

  const vorgang = useVorgang()

  const [gewaehlt, setzeGewaehlt] = useState('')
  const [feld, setzeFeld] = useState(bildfelder[0] ?? '')
  const [aufgabe, setzeAufgabe] = useState('')
  const [groesse, setzeGroesse] = useState<string>(dieGroessen[0])
  const [guete, setzeGuete] = useState<string>(dieGueten[1])
  const [bearbeiten, setzeBearbeiten] = useState(false)
  const [vorschlag, setzeVorschlag] = useState<KIBildvorschlagAntwort>()
  const [fehler, setzeFehler] = useState<string>()
  const [verworfen, setzeVerworfen] = useState(false)

  // **Erst beim Aufklappen wird der Katalog gelesen** — sonst wäre das
  // Aufschlagen einer Karte ein Ruf nach draußen, den niemand ausgelöst hat
  // (§16, Regel 6). Dieselbe Überlegung wie in [KIVorschlag].
  const [gefragt, setzeGefragt] = useState(false)

  // **Abgebrochen ist kein Fehler** (A17, #1532): wer selbst abbricht, hat
  // keinen Fehlschlag erlebt — er hat es sich anders überlegt. Ein roter
  // Hinweis daneben behauptete etwas anderes.
  const [abgebrochen, setzeAbgebrochen] = useState(false)

  const katalog = useKIModelle(gefragt)

  // freigabe sagt, ob gerade nach einer Freigabe gefragt wird (#1539).
  //
  // **Ein Bildauftrag braucht eine eigene** — das stand als Satz schon da
  // (`kiBildKosten`), und es gab keinen Weg, sie zu geben.
  const [freigabe, setzeFreigabe] = useState(false)

  // ausloesen entscheidet, ob vorher zu fragen ist.
  function ausloesen() {
    if (kostetGeld(katalog.modelle, gewaehlt)) {
      setzeFreigabe(true)

      return
    }

    void erzeugen('')
  }

  async function erzeugen(nachweis: string) {
    setzeFehler(undefined)
    setzeVerworfen(false)
    setzeAbgebrochen(false)

    await vorgang.starte(`kibild/${karteID}`, async () => {
      try {
        const antwort = await api().KIBildVorschlagen(
          karteID,
          feld,
          anbieterAus(gewaehlt),
          aliasAus(gewaehlt),
          aufgabe,
          groesse,
          guete,
          bearbeiten,
          nachweis,
        )

        setzeVorschlag(auspacken(antwort, () => antwort))
      } catch (ursache) {
        setzeVorschlag(undefined)
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  // abbrechen hält den laufenden Auftrag an.
  //
  // **Ohne `useVorgang`**: er bricht einen **anderen** Lauf ab und hat selbst
  // nichts, worauf zu warten wäre — ein zweiter Vorgang daneben hielte sich
  // gegenseitig auf.
  //
  // **Der Satz steht erst nach der Antwort** (Zweitblick über Phase 5.3,
  // Befund 3): die Bindung sagt, **ob** ein Auftrag lief. Wer das `false`
  // wegwirft, schreibt „Abgebrochen. Es wurde nichts übernommen …" auch dann
  // hin, wenn der Auftrag längst durch war und sein Bild gleich danach
  // ankommt — zwei Sätze, die einander widersprechen.
  async function abbrechen() {
    if (await api().KIBildAbbrechen(Number(karteID))) {
      setzeAbgebrochen(true)
    }
  }

  async function verwerfen(kennung: string) {
    setzeFehler(undefined)

    await vorgang.starte(`kibildweg/${karteID}`, async () => {
      try {
        auspacken(await api().KIBildVerwerfen(kennung), () => undefined)

        setzeVorschlag(undefined)
        setzeVerworfen(true)
      } catch (ursache) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  const laeuft = vorgang.laeuft(`kibild/${karteID}`)

  // **Ohne Bildfeld gibt es nichts zu erzeugen.** Eine Vorlage ohne Bild ist
  // kein Fehler — sie hat nur keine Illustration.
  if (bildfelder.length === 0) {
    return null
  }

  return (
    <details
      className="rounded-sm border border-kante px-3 py-2"
      onToggle={(ereignis) => {
        if (ereignis.currentTarget.open) {
          setzeGefragt(true)
        }
      }}
    >
      <summary className="cursor-pointer text-sm font-medium focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring">
        {eigene.kiBildTitel}
      </summary>

      <div className="mt-3 flex flex-col gap-3">
        <p className="max-w-prose text-sm text-leise">
          {eigene.kiBildAendertNichts}
        </p>

        <p className="max-w-prose text-sm text-leise">{eigene.kiBildKosten}</p>

        <KIModellwahl
          gefragt={gefragt}
          modelle={katalog.modelle}
          laedt={katalog.laedt}
          fehler={katalog.fehler}
          gewaehlt={gewaehlt}
          waehlen={setzeGewaehlt}
        />

        <label className="flex flex-col gap-1 text-sm">
          {eigene.kiBildFeld}
          <select
            className="rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            value={feld}
            onChange={(ereignis) => {
              setzeFeld(ereignis.target.value)
            }}
          >
            {bildfelder.map((name) => (
              <option key={name} value={name}>
                {name}
              </option>
            ))}
          </select>
        </label>

        {/* **Was im gewählten Feld schon steht** — der Bildnachweis nach
            einer früheren Übernahme. Er hängt am Feld und nicht an der
            Karte: zwei Bildfelder haben zwei Herkünfte. */}
        <KIBildnachweis
          karteID={karteID}
          feld={feld}
          gefragt={gefragt}
          stand={stand}
        />

        <label className="flex flex-col gap-1 text-sm">
          {eigene.kiBildAufgabe}
          <textarea
            className="rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            rows={2}
            placeholder={eigene.kiBildAufgabeHinweis}
            value={aufgabe}
            onChange={(ereignis) => {
              setzeAufgabe(ereignis.target.value)
            }}
          />
        </label>

        <div className="flex flex-wrap gap-3">
          <label className="flex flex-col gap-1 text-sm">
            {eigene.kiBildGroesse}
            <select
              className="rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
              value={groesse}
              onChange={(ereignis) => {
                setzeGroesse(ereignis.target.value)
              }}
            >
              {dieGroessen.map((wert) => (
                <option key={wert} value={wert}>
                  {wert}
                </option>
              ))}
            </select>
          </label>

          <label className="flex flex-col gap-1 text-sm">
            {eigene.kiBildGuete}
            <select
              className="rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
              value={guete}
              onChange={(ereignis) => {
                setzeGuete(ereignis.target.value)
              }}
            >
              {dieGueten.map((wert) => (
                <option key={wert} value={wert}>
                  {wert}
                </option>
              ))}
            </select>
          </label>
        </div>

        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            className="focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            checked={bearbeiten}
            onChange={(ereignis) => {
              setzeBearbeiten(ereignis.target.checked)
            }}
          />
          {eigene.kiBildBearbeiten}
        </label>

        <button
          type="button"
          className="self-start rounded-sm border border-kante px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
          disabled={laeuft || gewaehlt === '' || aufgabe === '' || feld === ''}
          onClick={ausloesen}
        >
          {eigene.kiBildErzeugen}
        </button>

        {/* **Und hier wird sie gegeben** (#1539). Der Satz darüber sagte
            schon, dass ein Bildauftrag gesondert freigegeben wird — nur gab
            es dafür keinen Weg. */}
        {freigabe && (
          <KIKostenfreigabe
            anbieter={anbieterAus(gewaehlt)}
            freigegeben={(nachweis) => {
              setzeFreigabe(false)
              void erzeugen(nachweis)
            }}
            verzichtet={() => {
              setzeFreigabe(false)
            }}
          />
        )}

        {/* **Die Wartemeldung sagt, worauf gewartet wird** (A17, #1532) —
            Anbieter und Modell. „Es läuft" allein sagte nicht, wen man
            gerade bezahlt. Und **daneben steht der Abbruch**: vorher gab es
            keinen, und ein Auftrag dauert Sekunden bis Minuten. */}
        {laeuft && (
          <p
            role="status"
            className="flex items-center gap-2 text-sm text-leise"
          >
            <span>
              {einsetzen(eigene.kiBildLaeuft, {
                anbieter: anbieterAus(gewaehlt),
                modell: aliasAus(gewaehlt),
              })}
            </span>

            <Knopf gewicht="leise" aufKlick={() => void abbrechen()}>
              {eigene.kiBildAbbrechen}
            </Knopf>
          </p>
        )}

        {abgebrochen && !laeuft && (
          <p role="status" className="text-sm text-leise">
            {eigene.kiBildAbgebrochen}
          </p>
        )}

        {fehler !== undefined && (
          <p role="alert" className="text-sm text-fehler">
            {fehler}
          </p>
        )}

        {verworfen && (
          <p role="status" className="text-sm text-leise">
            {eigene.kiBildVerworfen}
          </p>
        )}

        {vorschlag !== undefined && (
          <KIBildvorschau
            vorschlag={vorschlag}
            eigene={eigene}
            titel={einsetzen(eigene.kiBildVorschau, { feld })}
            laeuft={vorgang.laeuft(`kibildweg/${karteID}`)}
            zuschneiden={() => {
              // **Der Vorschlag bleibt stehen, bis eingesetzt ist** (§7,
              // Regel 3): das Zuschnittfenster lässt sich abbrechen, und
              // dann soll er noch da sein.
              void zuschneiden(feld, vorschlag.kennung)
            }}
            verwerfen={() => {
              void verwerfen(vorschlag.kennung)
            }}
          />
        )}
      </div>
    </details>
  )
}
