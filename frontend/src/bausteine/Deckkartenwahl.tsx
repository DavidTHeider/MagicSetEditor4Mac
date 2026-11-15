// **Eine Karte über ihren Namen ins Deck legen** — Werkstück I von Phase 4.10
// (#1325).
//
// # Der Befund, für den es diese Datei gibt
//
// Die Werkbank hatte zwei Textfelder: „Druck (Scryfall-Kennung)" und „Eigene
// Karte (Kennung)". Wer eine eigene Karte legen wollte, musste ihre
// Datenbankkennung **tippen** — eine Zahl, die nirgends in der Oberfläche
// steht. Am 2026-09-03 im Fenster nachgesehen: es gab keinen Weg, sie zu
// erfahren.
//
// **Das Scryfall-Feld bleibt.** Ein Druck wird über seine Scryfall-Kennung
// benannt, und die steht auf jeder Kartenseite dort; sie hier zu ersetzen
// hieße, eine Suche gegen einen fremden Dienst zu bauen — die gibt es auf der
// Kartenseite schon, und §16 Regel 5 sagt, dass die Anwendung ohne Netz
// weiterarbeitet. Der Weg über eigene Karten braucht kein Netz.
//
// # Zwei Listen, und die zweite mit Suche
//
// Ein Set hat bei `Black.mse-set` 3.788 Karten, und `Kartenzeilen` gibt sie
// ohne ihre Feldwerte (245.641 statt 4.761.182 Bytes). Ein Auswahlfeld mit
// 3.788 Einträgen ist unhandlich — aber es ist **bedienbar**, und ein Suchfeld
// wäre eine zweite Suche neben der auf der Kartenseite.
//
// **Seit UXR-11 grenzt eine Suche das Auswahlfeld ein**, sobald es fünfzehn
// Karten hat — R7 in `docs/UIREVIEW.md`: eine Auswahl mit Suche ab fünfzehn
// Werten
// (#1919).
// **Keine zweite Suche**: sie fragt niemanden, sie lässt nur Einträge weg,
// die schon da sind, über den Namen. Gezeichnet wird ohne Suchwort dieselbe
// Liste wie vorher, mit einem Suchwort weniger; die Zeichenzeit, die der Satz
// oben zu messen verlangt, wächst damit nicht.
import { useId, useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { Knopf } from './Knopf'
import { Zahlenfeld } from './Zahlenfeld'
import { useAnschluss, useAufgenommen } from '../anschlussraum'
import { useLaden } from '../haken/useLaden'
import { useSetImFenster } from '../haken/offenesset'
import { useSprache, useTexte } from '../sprache/kontext'

import type { KartenzeileDTO, SetDTO } from '../api'
import { falte } from '../sprache/vergleich'

interface Wahleigenschaften {
  // spiel ist das Spiel des offenen Decks — vorbelegt wird nur ein Set
  // dieses Spiels.
  spiel: string

  // angehalten hält jedes Bedienelement an, solange ein Vorgang am Deck läuft
  // (§7, Regel 4).
  angehalten: boolean

  legen: (
    karteID: string,
    printing: string,
    menge: number,
    bereich: string,
  ) => void
}

export function Deckkartenwahl({
  spiel,
  angehalten,
  legen,
}: Wahleigenschaften) {
  const texte = useTexte().seiten.decks
  const sprache = useSprache()

  // gewaehltesSet ist die Wahl des Menschen — `null`, solange er keine
  // getroffen hat; bis dahin gilt die Vorgabe darunter.
  const [gewaehltesSet, setzeSetID] = useState<string | null>(null)
  // karte merkt sich die gewählte Karte **mit dem Set, aus dem sie kam** —
  // Zweitblick über Phase 6, zweiter Teil, Befund 2 (§7, Regel 1 und 2):
  // folgt die Vorbelegung des Sets dem offenen Set, gilt eine Karte des
  // vorigen nicht mehr, auch ohne dass jemand das Setfeld berührt hat.
  const [karte, setzeKarte] = useState({ setID: '', id: '' })
  const [printing, setzePrinting] = useState('')
  const [menge, setzeMenge] = useState(4)

  // mengeGilt sagt, ob das Feld eine Menge trägt, die gelegt werden darf (F03,
  // #2245): bis dahin wurde ein leeres null und „2.5" ging an ein `int`.
  const [mengeGilt, setzeMengeGilt] = useState(true)
  const [bereich, setzeBereich] = useState('hauptdeck')
  const [suche, setzeSuche] = useState('')

  // **Was aus der Sammlung mitgereist ist, steht schon im Feld** —
  // UX-03.AK-01
  // (#1675).
  //
  // **Der Druck und nicht die Kartenkennung**: ein Exemplar der Sammlung ist
  // ein Stück Papier mit einer Scryfall-Kennung; eine Karte **dieses**
  // Bestands ist es nicht. Deshalb wird das Scryfall-Feld gefüllt und die
  // Kartenwahl geräumt — dieselbe Entweder-oder-Regel wie beim Tippen.
  const anschluss = useAnschluss()
  const mitgereist =
    anschluss.offen?.art === 'deck' ? anschluss.offen : undefined

  useAufgenommen(mitgereist, (angekommen) => {
    setzePrinting(angekommen.printing)
    setzeKarte({ setID: '', id: '' })
  })

  const sets = useLaden<SetDTO[]>(async () => {
    const antwort = await api().Sets()

    return auspacken(antwort, () => antwort.sets)
  }, [])

  // **Vorbelegt ist das offene Set, wenn es zum Spiel des Decks gehört** —
  // sonst das einzige Set dieses Spiels, sonst keines. AP-6 (#2363): die
  // Wahl des Sets war eine der Handlungen über dem Ziel des Kernwegs `deck`,
  // und `docs/UX-ARCHITEKTUR.md` nannte schon, was fehlte — das Fenster
  // kannte das offene Set nicht.
  const offenesSet = useSetImFenster()
  const passende = (sets.daten ?? []).filter((satz) => satz.spiel === spiel)
  const [einziges, ...weitere] = passende
  const vorgabe =
    offenesSet !== undefined &&
    passende.some((satz) => satz.id === offenesSet.id)
      ? offenesSet.id
      : einziges !== undefined && weitere.length === 0
        ? einziges.id
        : ''
  const setID = gewaehltesSet ?? vorgabe

  // **Ein Setwechsel verwirft die Karte, statt sie auszublenden** — Nachlauf
  // zum Zweitblick, R2: ausgeblendet stand sie nach A → B → A sofort wieder
  // da, während As Karten noch luden. Geräumt wird beim Zeichnen und nicht
  // im Effekt, damit kein Bild mit der alten Wahl entsteht.
  if (karte.id !== '' && karte.setID !== setID) {
    setzeKarte({ setID: '', id: '' })
  }

  const karteID = karte.setID === setID ? karte.id : ''
  const setzeKarteID = (id: string) => {
    setzeKarte({ setID, id })
  }

  // **Die Karten hängen am gewählten Set.**
  const karten = useLaden<KartenzeileDTO[]>(async () => {
    if (setID === '') {
      return []
    }

    const antwort = await api().Kartenzeilen(setID)

    return auspacken(antwort, () => antwort.zeilen)
  }, [setID])

  const bereit = karteID !== '' || printing !== ''

  // kannLegen gilt für den Knopf **und** für ⏎ und Doppelklick — drei Wege,
  // eine Bedingung.
  const kannLegen = !angehalten && bereit && mengeGilt

  // **Ab fünfzehn Karten eine Suche** (R7, UXR-11). **Die gewählte bleibt
  // angeboten**, auch wenn das Suchwort sie nicht trifft: sonst fiele das
  // Auswahlfeld still auf den ersten Eintrag, und ⏎ legte eine andere Karte.
  const alle = karten.daten ?? []
  const frage = falte(suche.trim(), sprache)
  const angeboten =
    frage === ''
      ? alle
      : alle.filter(
          (zeile) =>
            zeile.id === karteID || falte(zeile.name, sprache).includes(frage),
        )
  const bereichID = useId()

  // **⏎ und Doppelklick legen ebenfalls** — Befund 3 des Zweitblicks über
  // Phase 5.1: #1421
  // verlangt beides, und produktiv gab es nur den Knopf. Der Kernweg `deck`
  // rechnet mit ⏎ (`docs/UX-ARCHITEKTUR.md`): ohne ihn steht die Anlage bei
  // elf Handlungen statt sechs.
  //
  // **An der Kartenwahl und nicht am ganzen Kasten**: ein ⏎ im Mengenfeld
  // oder in der Druckkennung soll die Eingabe abschließen, nicht heimlich
  // etwas ins Deck legen.
  const jetztLegen = () => {
    if (kannLegen) {
      legen(karteID, printing, menge, bereich)

      // **Verbraucht ist verbraucht** (§7, Regel 3): die Karte liegt, und der
      // Satz darüber spräche sonst weiter über eine Handlung von vorhin.
      anschluss.verbrauche()
    }
  }

  return (
    <div className="flex flex-col gap-3">
      {/* **Oben bündig, nicht unten** (P02, #2166): Felder verschiedener
          Höhe standen am unteren Rand ausgerichtet, und ihre Beschriftungen
          sprangen — gemessen am nativen Fenster 449, 449, 437 und 448 px
          (`politur`). Dieselbe Regel wie in der Sitzliste. */}
      <div className="flex flex-wrap items-start gap-3">
        <label className="flex flex-col gap-1">
          <span className="text-sm">{texte.setWaehlen}</span>

          <select
            disabled={angehalten}
            value={setID}
            onChange={(ereignis) => {
              setzeSetID(ereignis.target.value)

              // **Der Wechsel räumt die gewählte Karte mit** (§7, Regel 1):
              // eine Kennung aus dem vorigen Set wäre sonst noch gewählt,
              // während die Liste daneben andere Namen zeigt.
              setzeKarteID('')
              setzeSuche('')
            }}
            className="rounded-md border border-kante px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          >
            <option value="">{texte.ohneSet}</option>

            {(sets.daten ?? []).map((satz) => (
              <option key={satz.id} value={satz.id}>
                {satz.name}
              </option>
            ))}
          </select>
        </label>

        {alle.length >= 15 && (
          <label className="flex flex-col gap-1">
            <span className="text-sm">{texte.karteSuchen}</span>

            <input
              type="search"
              disabled={angehalten}
              value={suche}
              onChange={(ereignis) => {
                setzeSuche(ereignis.target.value)
              }}
              className="w-40 rounded-md border border-kante px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            />
          </label>
        )}

        <label className="flex flex-col gap-1">
          <span className="text-sm">{texte.karteWaehlen}</span>

          <select
            disabled={angehalten || setID === ''}
            value={karteID}
            onChange={(ereignis) => {
              setzeKarteID(ereignis.target.value)
              setzePrinting('')
            }}
            onKeyDown={(ereignis) => {
              if (ereignis.key === 'Enter') {
                ereignis.preventDefault()
                jetztLegen()
              }
            }}
            onDoubleClick={jetztLegen}
            className="rounded-md border border-kante px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          >
            <option value="">{texte.ohneKarte}</option>

            {angeboten.map((zeile) => (
              <option key={zeile.id} value={zeile.id}>
                {zeile.name}
              </option>
            ))}
          </select>
        </label>

        <Zahlenfeld
          beschriftung={texte.menge}
          wert={menge}
          kleinstens={1}
          ganzzahlig
          angehalten={angehalten}
          geaendert={setzeMenge}
          gilt={setzeMengeGilt}
        />

        {/* **Zwei Werte sind zwei Optionsfelder** — R7 (UXR-11): Segmente
            bis vier Werte; ein Auswahlfeld verbarg den zweiten hinter einem
            Klick. **Kein Knopfpaar**: ein gedrückter `Knopf` trägt das
            Gewicht des Hauptknopfs, und neben „Ins Deck legen" stünden zwei
            gefüllte Flächen in einer Region (R1). */}
        <fieldset className="flex flex-col gap-1">
          <legend className="text-sm">{texte.bereich}</legend>

          {/* **`mt-1` wie die `gap-1` der Beschriftungen daneben** — WebKit
              gibt der `legend` keinen Abstand, und die Optionsfelder standen
              4 px höher als die Felder daneben (gemessen). */}
          <div className="mt-1 flex gap-3 py-1">
            {(['hauptdeck', 'sideboard'] as const).map((wert) => (
              <label key={wert} className="flex items-center gap-1 text-sm">
                <input
                  type="radio"
                  name={bereichID}
                  value={wert}
                  checked={bereich === wert}
                  disabled={angehalten}
                  onChange={() => {
                    setzeBereich(wert)
                  }}
                  className="size-4 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
                />
                {wert === 'hauptdeck' ? texte.hauptdeck : texte.sideboard}
              </label>
            ))}
          </div>
        </fieldset>

        {/* **Unten bündig mit den Feldern seiner Zeile** — und bricht er in
            eine eigene Zeile um, steht er dort ohne Versatz. */}
        <span className="self-end">
          <Knopf gewicht="haupt" angehalten={!kannLegen} aufKlick={jetztLegen}>
            {texte.legen}
          </Knopf>
        </span>
      </div>

      {/* **Das Scryfall-Feld steht daneben und nicht darüber.** Es ist der
          seltenere Weg: eigene Karten sind das, wofür es dieses Programm
          gibt. */}
      <label className="flex max-w-md flex-col gap-1">
        <span className="text-sm">{texte.druck}</span>

        <input
          disabled={angehalten}
          value={printing}
          onChange={(ereignis) => {
            setzePrinting(ereignis.target.value)
            setzeKarteID('')
          }}
          className="rounded-md border border-kante px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        />
      </label>

      {mitgereist !== undefined && (
        <p className="text-sm text-leise">
          {einsetzen(texte.anschlussWartet, { name: mitgereist.name })}
        </p>
      )}

      {sets.daten?.length === 0 && (
        <p className="text-sm text-leise">{texte.setLeer}</p>
      )}

      {/* **Gefiltert leer ist nicht leer** (§6): das Set hat Karten, nur
          passt keine zum Suchwort — und der nächste Schritt ist ein anderes
          Wort, nicht ein anderes Set. */}
      {frage !== '' && angeboten.length === 0 && (
        <p className="text-sm text-leise">
          {einsetzen(texte.karteKeine, { suche: suche.trim() })}
        </p>
      )}

      {setID !== '' && karten.daten?.length === 0 && (
        <p className="text-sm text-leise">{texte.setOhneKarten}</p>
      )}
    </div>
  )
}
