// Die Lasche **Setangaben** —
// #567.
//
// # Wofür
//
// Ein Set hat Angaben: Titel, Beschreibung, Künstler, Copyright, Setcode,
// Sprache, Symbol. Über `magic.mse-game` sind es **89 Felder**. Sie standen im
// Bestand und waren in der Oberfläche nirgends erreichbar — wer den Setcode
// ändern wollte, konnte es nicht.
//
// # Was MSE tut
//
//	void SetInfoEditor::onChangeSet() {
//	  setStyles(set->stylesheet, set->stylesheet->set_info_style);
//	  setData(set->data);
//	}
//
// `src/gui/control/native_look_editor.cpp`, Zeilen 187–190, Fassung f9d9356
// vom 2021-07-06 — eine eigene Lasche neben „Cards", mit denselben
// Steuerelementen wie der Karteneditor.
//
// **Ohne Sichtbarkeitsfilter.** `NativeLookEditor::resizeViewers` legt jede
// Ansicht untereinander und setzt ihren Kasten selbst (Zeilen 48–93); anders
// als beim Karteneditor entscheidet der Stil hier nicht, was zu sehen ist.
//
// **Aber in Gruppen, und nur die erste offen** — UXR-07
// (#1915).
// Die Gruppen setzt das Spiel mit seinen Auskünften; warum und wie, steht in
// `setangabengruppen.ts`.
import { useCallback, useState } from 'react'

import { api, auspacken, type FeldbeschreibungDTO } from '../api'
import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import { Setkopfangaben } from '../bausteine/Setkopfangaben'
import { Paketbedarf } from './Paketbedarf'
import { Vorbehalte } from './Vorbehalte'
import { Leerzustand } from '../bausteine/Leerzustand'
import { useLaden } from '../haken/useLaden'
import { useAutoersetzung } from '../haken/useAutoersetzung'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus, verursachendesFeld } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { Ladefeld } from './Ladefeld'
import { gliedere } from './setangabengruppen'
import { Setangabezeile } from './Setangabezeile'
import { Teilen } from '../bausteine/Teilen'

interface SetangabenEigenschaften {
  setID: string

  // verlaufsstand zählt ⌘Z und ⌘⇧Z mit — er liest die Angaben nach.
  //
  // **Er baut die Seite nicht mehr neu auf** (U03, #2120). Bis dahin stand
  // er im `key` des Aufrufers, damit nach ⌘Z kein zurückgenommener Text im
  // Feld stand — und jede Änderung irgendwo nahm dem, der tippte, Entwurf
  // und Fokus. Jetzt trägt jede Zeile den gespeicherten Wert im Schlüssel
  // ihrer Eingabe, solange niemand tippt (`Setangabezeile`). Der Messbefund
  // von damals — ein neuer Knoten mit leerem `defaultValue` — kam vom
  // Nachlesen, das die Daten räumte; seit U07 bleiben sie stehen.
  verlaufsstand: number
}

// Feldfehler ordnet eine Meldung dem Feld zu, das sie verursacht hat.
interface Feldfehler {
  feld: string
  meldung: string
}

export function Setangaben({ setID, verlaufsstand }: SetangabenEigenschaften) {
  // **Auch die Setangaben sind Textfelder** (#897). MSE2 ersetzt in jedem
  // Werteditor, nicht nur im Karteneditor.
  const autoErsetzen = useAutoersetzung(setID)
  const texte = useTexte()
  const vorgang = useVorgang()
  const [fehler, setzeFehler] = useState<Feldfehler>()

  // **`verlaufsstand` steht nicht in den Abhängigkeiten von `holen`.** Er
  // ändert die Frage nicht — `Setangaben(setID)` fragt dasselbe —, sondern nur
  // den **Zeitpunkt**; das steht im Schlüssel von [useLaden] darunter.
  const holen = useCallback(async () => {
    const antwort = await api().Setangaben(setID)

    return auspacken(antwort, () => antwort)
  }, [setID])

  const angaben = useLaden(holen, [setID], false, verlaufsstand)

  // **Einmal ausgepackt.** `angaben.daten` ist erst `undefined` und dann da;
  // drei Rückfälle daneben wären drei Zweige für dieselbe Frage.
  const felder = angaben.daten?.felder ?? []
  const werte = new Map(
    (angaben.daten?.werte ?? []).map((wert) => [wert.feld, wert.wert]),
  )

  const schreiben = async (
    feld: string,
    wert: string,
    bekannt: readonly string[],
  ) => {
    // Regel 4: ein Vorgang je Datensatz. Der Schlüssel ist Set **und** Feld —
    // zwei Angaben desselben Sets dürfen gleichzeitig gehen.
    const ergebnis = await vorgang.starte(`${setID}/${feld}`, async () => {
      try {
        auspacken(
          await api().SetwertSetzen(setID, feld, wert, bekannt),
          () => undefined,
        )

        return true
      } catch (ursache: unknown) {
        setzeFehler({
          feld: verursachendesFeld(ursache, feld),
          meldung: meldungAus(ursache, texte),
        })

        return false
      }
    })

    // Regel 3: nachgeladen wird **nur** nach einem geglückten Schreiben.
    if (ergebnis === true) {
      setzeFehler(undefined)
      angaben.nachladen()
    }
  }

  // feldzeile ist ein Setfeld samt Aufschrift und Befund — in jeder Gruppe
  // dasselbe.
  const feldzeile = (form: FeldbeschreibungDTO) => (
    <Setangabezeile
      key={form.name}
      form={form}
      wert={werte.get(form.name) ?? ''}
      fehler={fehler?.feld === form.name ? fehler.meldung : undefined}
      laeuft={vorgang.laeuft(`${setID}/${form.name}`)}
      autoErsetzen={autoErsetzen}
      schreiben={(neu, bekannt) => {
        void schreiben(form.name, neu, bekannt)
      }}
    />
  )

  // raster legt die Felder einer Gruppe aus.
  //
  // **Zwei Spalten** — Werkstück C von Phase 4.10
  // (#1325).
  //
  // # Was vorher war
  //
  // Fünfundfünfzig Felder in **einer** Spalte, `max-w-2xl` breit:
  // gemessen am 2026-09-03 ist die Lasche **7,8 Bildschirme** hoch und
  // trägt 107 Tabstellen — die größte Seite der Anwendung, und die mit
  // dem geringsten Aufbau. Das Feld für `Set code` — drei Zeichen —
  // war 673 Punkte breit.
  //
  // # Und warum nicht mehr als zwei
  //
  // **Weil die Aufschriften aus dem Paket kommen** und beliebig lang
  // sein können (§9: was aus einem MSE-Paket kommt, wird nicht
  // übersetzt und nicht gekürzt). Drei Spalten bei 1280 Punkten
  // ließen für jede 280 — dort bricht „Secondary copyright" um, und
  // eine umgebrochene Aufschrift über einem einzeiligen Feld sieht aus
  // wie zwei Felder.
  //
  // **Zwei Spalten fragen den Behälter, nicht das Fenster** (UX-08,
  // #1680): die Seite steht in der Schublade *Eigenschaften*, und
  // `sm:` hätte dort bei jedem Fenster über 640 Punkten zwei Spalten
  // in 430 gesetzt.
  const raster = (gruppe: readonly FeldbeschreibungDTO[]) => (
    <form
      className="grid max-w-4xl gap-x-6 gap-y-3 @2xl:grid-cols-2"
      onSubmit={(ereignis) => {
        ereignis.preventDefault()
      }}
    >
      {gruppe.map(feldzeile)}
    </form>
  )

  const { erste, weitere } = gliedere(felder)

  return (
    // **Kein Kasten im Kasten** — AP-12 (#2369), B-28: die Seite steht in
    // der Schublade *Eigenschaften*, und Stil, Symbol und Schlüsselwörter
    // daneben tragen keinen eigenen Rahmen.
    <div className="flex flex-col gap-4">
      {/* **`Ladefeld` zeigt Laden, Fehler und Leere** — der Inhalt steht
          daneben, denn seine `children` **sind** der leere Zustand. */}
      <Ladefeld
        laedt={angaben.laedt && angaben.daten === undefined}
        fehler={angaben.fehler}
        leer={felder.length === 0}
      >
        <Leerzustand
          titel={texte.seiten.setangaben.leerTitel}
          erklaerung={texte.seiten.setangaben.leer}
        />
      </Ladefeld>

      {/* **Die fünf Formatangaben** — Werkstück I von Phase 4.10
          (#1325).
          Sie standen im Kasten „Nicht gedeutet" auf der Kartenlasche; sie
          handeln vom Set und gehören hierher. */}
      {angaben.daten && <Setkopfangaben kopf={angaben.daten.kopf} />}

      {/* **Was dem Set an Paketen fehlt** — DOM-01.AK-04 (#1696): hier, wo
          Spiel und Stil des Sets stehen, und mit dem Weg, es zu beheben. */}
      <Paketbedarf setID={setID} />

      {/* **Die Vorbehalte des Imports** (Werkstück I von Phase 4.10). Sie
          standen auf der Kartenlasche und handeln vom Set. */}
      <Vorbehalte setID={setID} />

      {felder.length > 0 && (
        <div className="flex flex-col gap-4">
          <section className="flex flex-col gap-2">
            {erste.kopf !== undefined && (
              <h3 className="font-semibold">{erste.kopf.aufschrift}</h3>
            )}
            {raster(erste.felder)}
          </section>

          {/* **Eine Gruppe mit dem Feld, das ein Fehler nennt, steht offen**
              (§6) — er erschiene sonst hinter einem Knopf. */}
          {weitere.map((gruppe) => (
            <Aufklappkasten
              key={gruppe.kopf.name}
              titel={gruppe.kopf.aufschrift}
              offenGehalten={gruppe.felder.some(
                (form) => form.name === fehler?.feld,
              )}
            >
              {raster(gruppe.felder)}
            </Aufklappkasten>
          ))}
        </div>
      )}

      {/* **Das Set als Archiv teilen** — NET-04 (#1708). */}
      <Teilen key={setID} art="set" bezug={setID} />
    </div>
  )
}
