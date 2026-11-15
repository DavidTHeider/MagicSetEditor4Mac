// Die Übersicht der Tastenkürzel —
// #245.
//
// **Was sie tut, steht hier; wie sie aussieht, in den Klassen** (§2).
//
// # Der Abnahmekasten
//
// „⌘1…⌘8, ⌘K, ⌘S, ⌘Z, ⌘, stehen — **sichtbar** sind sie nirgends." Sie standen
// in der Navigationsleiste am jeweiligen Eintrag, aber nur die der Seiten; ⌘K,
// ⌘F, ⌘Z und ⌘S kannte, wer sie kannte.
//
// # Warum sie aus den Quellen kommt und nicht aus einer Aufzählung
//
// Die Seitenkürzel stehen in `navigation.ts`, die übrigen in `tastatur.ts`.
// Eine dritte Liste hier wäre eine dritte Wahrheit — und die, die niemand
// nachführt, weil sie nur eine Anzeige ist.
//
// **⌘0 bis ⌘9 lassen sich belegen**, seit die Ziffern den Bereichen gehören
// (#1753): UX-02.AK-03
// (#1674)
// verlangt zur Migration eine konfigurierbare Belegung. Die übrigen Kürzel
// bleiben, was sie sind.
import { useId } from 'react'

import { seiten } from '../navigation'
import { globaleKuerzel, kuerzelschluessel } from '../rahmen/tastatur'
import { beschriftung } from '../navigation'
import { useSprache, useTexte } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { Einstellungsherkunft } from '../bausteine/Einstellungsherkunft'
import { Knopf } from '../bausteine/Knopf'
import type { EinstellungDTO } from '../api'
import {
  alsWert,
  belegungVor06,
  mitZiffer,
  vorgabebelegung,
  ziffern,
  type Ziffernbelegung,
} from '../rahmen/ziffernbelegung'
import { vergleicher } from '../sprache/vergleich'

// zeichen macht aus einer Taste das Zeichen, das auf ihr steht.
//
// **`⌘` und `⇧` statt „Cmd" und „Shift".** So schreibt macOS sie in seinen
// Menüs, und daneben steht die Anwendung.
function zeichen(taste: string, umschalt: boolean, alt = false): string {
  return (alt ? '⌥⌘' : umschalt ? '⇧⌘' : '⌘') + taste.toUpperCase()
}

// zweitzeichen macht aus einer Ziffer der zweiten Reihe ihr Kürzel
// (#1291).
//
// **⌥ steht vor ⌘**, wie macOS es in seinen Menüs schreibt.
//
// # Warum die Ziffer trotzdem eine Ziffer bleibt
//
// Sie ist eine **Beschriftung**, und gedeutet wird über `event.code` — ⌥
// schreibt das Zeichen um (⌥1 gibt „¡"), also führt am Code kein Weg vorbei.
// `code` benennt aber die **Position**, nicht den Aufdruck: auf AZERTY liegt
// an `Digit1` ein `&`
// (#1345).
//
// **Chromium könnte es beschriften** — `navigator.keyboard.getLayoutMap()`.
// Dieses Fenster ist keines: Wails setzt auf macOS ein WKWebView, und die
// Keyboard-Map-API gehört nicht zu WebKit. Eine Beschriftung, die nur im
// Browservorschau-Modus stimmt, wäre die falsche Hälfte.
//
// **Also sagt die Tafel es aus** — der Satz darunter nennt die Reihe und die
// Position, und für QWERTZ und QWERTY stimmt die Ziffer weiterhin.
function zweitzeichen(ziffer: string): string {
  return '⌥⌘' + ziffer
}

interface Kuerzeltafeleigenschaften {
  // belegung sagt, welche Seite eine Ziffer öffnet — ohne sie die Vorgabe.
  belegung?: Ziffernbelegung | undefined

  // setze schreibt die Belegung; **ohne es bleibt die Tafel eine Anzeige**.
  setze?: ((wert: string) => void) | undefined

  // eintrag ist die Einstellung selbst — für Herkunft, Ebene und
  // Wirkzeitpunkt wie an jeder anderen (Befund M15, #1665).
  eintrag?: EinstellungDTO | undefined

  // angehalten sagt, dass gerade geschrieben wird — dann nimmt die Tafel
  // keinen zweiten Wunsch an (§7, Regel 4; Befund G13, #1665).
  angehalten?: boolean | undefined

  // eingebettet lässt Überschrift und Abstand weg — in den Einstellungen
  // steht die Tafel in einem Aufklapper, dessen Knopf schon „Tastenkürzel"
  // heißt (UXR-19, #1927). Zwei gleiche Namen übereinander sind einer zu viel.
  eingebettet?: boolean
}

export function Kuerzeltafel({
  belegung = vorgabebelegung,
  setze,
  eintrag,
  angehalten = false,
  eingebettet = false,
}: Kuerzeltafeleigenschaften) {
  const texte = useTexte()

  // **Die Ziffern aus der Belegung, die übrigen aus `navigation.ts`** —
  // ⌘, ⌘' und ⌘? sind nicht einstellbar.
  const seitenkuerzel = [
    ...ziffern.flatMap((ziffer) => {
      const seite = belegung[ziffer]

      return seite === undefined ? [] : [{ kuerzel: ziffer, schluessel: seite }]
    }),
    ...seiten.filter(
      (seite) => seite.kuerzel !== '' && !ziffern.includes(seite.kuerzel),
    ),
  ]

  // **Die zweite Reihe steht darunter und nicht dazwischen**: erst die zehn
  // Ziffern, dann die acht mit ⌥. So liest sich die Tafel wie die Navigation.
  const zweitkuerzel = seiten.filter((seite) => seite.zweitkuerzel !== '')

  return (
    <section className={eingebettet ? undefined : 'mt-10'}>
      {!eingebettet && (
        <h2 className="text-lg font-medium">
          {texte.seiten.einstellungen.kuerzelUeberschrift}
        </h2>
      )}

      <Erklaerungsabsatz
        text={texte.seiten.einstellungen.kuerzelHinweis}
        klasse="mt-1 max-w-prose text-sm text-leise"
      />

      <table className="mt-3 w-full max-w-xl border-collapse text-left">
        <caption className="sr-only">
          {texte.seiten.einstellungen.kuerzelUeberschrift}
        </caption>
        <thead>
          <tr className="border-b border-kante">
            <th scope="col" className="w-28 py-1 pr-4">
              {texte.seiten.einstellungen.kuerzelTaste}
            </th>
            <th scope="col" className="py-1">
              {texte.seiten.einstellungen.kuerzelHandlung}
            </th>
          </tr>
        </thead>
        <tbody>
          {globaleKuerzel.map((angabe) => (
            <tr
              key={kuerzelschluessel(angabe)}
              className="border-b border-kante"
            >
              <td className="py-1 pr-4 font-mono">
                {zeichen(angabe.taste, angabe.umschalt, angabe.alt)}
              </td>
              <td className="py-1">
                {
                  texte.seiten.einstellungen.kuerzel[
                    angabe.art as keyof typeof texte.seiten.einstellungen.kuerzel
                  ]
                }
              </td>
            </tr>
          ))}

          {seitenkuerzel.map((seite) => (
            <tr
              key={`${seite.kuerzel}/${seite.schluessel}`}
              className="border-b border-kante"
            >
              <td className="py-1 pr-4 font-mono">
                {zeichen(seite.kuerzel, false)}
              </td>
              <td className="py-1">{beschriftung(texte, seite.schluessel)}</td>
            </tr>
          ))}

          {zweitkuerzel.map((seite) => (
            <tr key={seite.schluessel} className="border-b border-kante">
              <td className="py-1 pr-4 font-mono">
                {zweitzeichen(seite.zweitkuerzel)}
              </td>
              <td className="py-1">{beschriftung(texte, seite.schluessel)}</td>
            </tr>
          ))}
        </tbody>
      </table>

      {/* **Was ⌥⌘1 wirklich heißt** (#1345): die Beschriftung stimmt nicht auf
          jeder Belegung, die Taste schon. Der Grund steht bei
          [zweitzeichen]. */}
      <Erklaerungsabsatz
        text={texte.seiten.einstellungen.kuerzelZweiteReihe}
        klasse="max-w-3xl text-sm text-leise"
      />

      {setze !== undefined && (
        <Ziffernwahl
          belegung={belegung}
          setze={setze}
          eintrag={eintrag}
          angehalten={angehalten}
        />
      )}
    </section>
  )
}

// Ziffernwahl legt je Ziffer eine Seite fest — oder keine.
function Ziffernwahl({
  belegung,
  setze,
  eintrag,
  angehalten,
}: {
  belegung: Ziffernbelegung
  setze: (wert: string) => void
  eintrag: EinstellungDTO | undefined
  angehalten: boolean
}) {
  const texte = useTexte()
  const sprache = useSprache()
  const eigene = texte.seiten.einstellungen
  const titel = useId()

  // **Nach dem Namen geordnet**, wie ein Mensch eine Seite sucht.
  const auswahl = seiten
    .map((seite) => ({
      schluessel: seite.schluessel,
      name: beschriftung(texte, seite.schluessel),
    }))
    .sort((a, b) => vergleicher(sprache)(a.name, b.name))

  const waehle = (ziffer: string, seite: string) => {
    setze(alsWert(mitZiffer(belegung, ziffer, seite)))
  }

  return (
    <div
      role="group"
      aria-labelledby={titel}
      data-einstellung="kuerzel.ziffern"
      className="mt-6 flex max-w-xl flex-col gap-3"
    >
      <h3 id={titel} className="text-base font-medium">
        {eigene.belegungUeberschrift}
      </h3>

      <Erklaerungsabsatz text={eigene.belegungHinweis} />

      {eintrag !== undefined && (
        <Einstellungsherkunft eintrag={eintrag} ohneWert />
      )}

      <div className="grid grid-cols-2 gap-x-6 gap-y-2">
        {ziffern.map((ziffer) => (
          <label key={ziffer} className="flex items-center gap-2 text-sm">
            <span className="w-8 font-mono">{zeichen(ziffer, false)}</span>
            <select
              aria-label={einsetzen(eigene.belegungFuer, {
                kuerzel: zeichen(ziffer, false),
              })}
              value={belegung[ziffer] ?? ''}
              disabled={angehalten}
              onChange={(ereignis) => {
                waehle(ziffer, ereignis.target.value)
              }}
              className="min-w-0 flex-1 rounded-md border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            >
              <option value="">{eigene.belegungFrei}</option>
              {auswahl.map((seite) => (
                <option key={seite.schluessel} value={seite.schluessel}>
                  {seite.name}
                </option>
              ))}
            </select>
          </label>
        ))}
      </div>

      <div className="flex flex-wrap gap-2">
        <Knopf
          gewicht="leise"
          angehalten={angehalten}
          aufKlick={() => {
            setze(alsWert(belegungVor06))
          }}
        >
          {eigene.belegungVor06}
        </Knopf>
        <Knopf
          gewicht="leise"
          angehalten={angehalten}
          aufKlick={() => {
            setze('')
          }}
        >
          {eigene.belegungVorgabe}
        </Knopf>
      </div>
    </div>
  )
}
