// Das **Zufallspack** — einen Booster aus dem offenen Set ziehen
// (#1072).
//
// # Was MSE hier tut
//
// `src/gui/set/random_pack_panel.cpp` (Fassung f9d9356 vom 2021-07-06): links
// die Packarten mit einer Anzahl je Art, rechts die gezogenen Karten. Der Same
// steht dort als „seed" in einem Feld — dieselbe Zahl gibt dieselbe Ziehung.
//
// # Warum der Same sichtbar ist
//
// Weil er die einzige Möglichkeit ist, ein Ergebnis **wiederherzustellen**.
// Wer ein gutes Pack zieht und es zeigen will, gibt die Zahl weiter.
import { useCallback, useState } from 'react'

import { api } from '../fassade'
import type { KarteDTO } from '../api'
import { Faltliste } from '../bausteine/Faltliste'
import { Feld } from '../bausteine/Feld'
import { Hinweiskasten } from '../bausteine/Hinweiskasten'
import { Knopf } from '../bausteine/Knopf'
import { Ladefeld } from './Ladefeld'
import { useLaden } from '../haken/useLaden'
import { useVorgang } from '../haken/useVorgang'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { kartenname } from './kartenname'
import { meldungAus, verursachendesFeld } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

interface ZufallspackEigenschaften {
  setID: string
}

export function Zufallspack({ setID }: ZufallspackEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.karten.pack

  const [art, setzeArt] = useState('')
  const [same, setzeSame] = useState('1')
  const [gezogen, setzeGezogen] = useState<KarteDTO[]>()

  // **Der Fehler trägt sein Feld mit** — `app.samenAus` gibt für einen
  // unlesbaren Samen `NewFieldError("same", …)` zurück, und §6 verlangt den
  // Satz am verursachenden Feld. Ohne den Feldbezug stand er als anonymer
  // Hinweis über der Lasche.
  const [fehler, setzeFehler] = useState<{ feld: string; meldung: string }>()

  const vorgang = useVorgang()

  const holen = useCallback(async () => {
    const antwort = await api().Packarten(setID)

    return auspacken(antwort, () => ({
      arten: antwort.arten,
      befunde: antwort.befunde,
    }))
  }, [setID])

  const angebot = useLaden(holen, [setID])

  const arten = angebot.daten?.arten ?? []

  // **Einmal ausgepackt und nicht zweimal.** Der zweite `?? []` stünde in
  // einem Zweig, den kein Aufruf betreten kann — geprüft wird oben derselbe
  // Ausdruck (§3).
  const befunde = angebot.daten?.befunde ?? []

  // **Nur die wählbaren stehen zur Wahl.** In `magic.mse-game` aus dem
  // Full-Magic-Pack sind es zehn von dreißig (nachgezählt am 2026-10-02 nach
  // der Regel aus `msepaket.packs`: wählbar, wer Posten hat und nicht
  // `selectable: false` sagt). **Hier stand „vier"**, gezählt an einer
  // Fassung des Pakets, die noch `fat pack` und `custom pack` führte — so
  // steht es bei `service.Packangebot`; die Zahl hängt am Paket. Die übrigen
  // sind Bausteine, die von jenen aus gezogen werden. **Zehn, und deshalb ein
  // Auswahlfeld** und keine Segmente: R7 aus `docs/UIREVIEW.md` lässt
  // Segmente bis vier Werte zu.
  const waehlbare = arten.filter((eine) => eine.waehlbar)

  // gewaehlt ist die Packart, die dasteht — die erste wählbare, bis jemand
  // eine andere nimmt.
  const gewaehlt = art === '' ? (waehlbare[0]?.name ?? '') : art

  const ziehen = async () => {
    const karten = await vorgang.starte(setID, async () => {
      try {
        const antwort = await api().PackZiehen(setID, gewaehlt, same)

        return auspacken(antwort, () => antwort.karten)
      } catch (ursache: unknown) {
        setzeFehler({
          feld: verursachendesFeld(ursache, ''),
          meldung: meldungAus(ursache, texte),
        })

        return undefined
      }
    })

    if (karten !== undefined) {
      setzeFehler(undefined)
      setzeGezogen(karten)
    }
  }

  // **Die Vorprüfung sagt genau das, was Go sagt** — nicht mehr: Ziffern,
  // höchstens 2⁶⁴−1, und leer ist gültig. Der Befund aus Go schlägt sie: er
  // weiß, woran es lag.
  const samenbefund =
    fehler?.feld === 'same'
      ? fehler.meldung
      : samenPasst(same)
        ? undefined
        : eigene.sameZahl

  return (
    <section className="flex flex-col gap-4">
      <header className="flex flex-col gap-1">
        <h2 className="text-xl font-semibold">{eigene.ueberschrift}</h2>
        <Erklaerungsabsatz
          text={eigene.erklaerung}
          klasse="text-sm text-leise"
        />
      </header>

      {/* **Der Grund steht beim Laden** — Werkstück C von Phase 4.10. Der
          erste Aufruf braucht mit *Blue* (3.693 Karten) 16,6 Sekunden
          (2026-10-19, Mac17,3, `zufallspack09`; vorher rund 100, #2466), jeder
          weitere Millisekunden; Sekunden „Wird geladen …" ohne Erklärung
          sehen aus wie ein Fehler (§8). */}
      <Ladefeld
        laedt={angebot.laedt}
        fehler={angebot.fehler}
        leer={waehlbare.length === 0}
        grund={eigene.laedtGrund}
      >
        <p className="text-sm text-leise">{eigene.keineArten}</p>
      </Ladefeld>

      {waehlbare.length > 0 && (
        <div className="flex flex-wrap items-end gap-3">
          <label className="flex flex-col gap-1 text-sm">
            {eigene.art}
            <select
              className={eingabe}
              value={gewaehlt}
              disabled={vorgang.laeuft(setID)}
              onChange={(ereignis) => {
                setzeArt(ereignis.target.value)
              }}
            >
              {waehlbare.map((eine) => (
                <option key={eine.name} value={eine.name}>
                  {einsetzen(eigene.artMitZahl, {
                    name: eine.name,
                    zahl: String(eine.karten),
                  })}
                </option>
              ))}
            </select>
          </label>

          {/* **Der Same ist eine ganze Zahl ab null** — `app.samenAus` liest
              ihn mit `strconv.ParseUint(text, 10, 64)`, und ein leerer Text
              heißt null. Beides steht am Feld, statt erst nach dem Klick als
              Satz über der Lasche zu erscheinen (§6). */}
          <Feld
            beschriftung={eigene.same}
            {...(samenbefund === undefined ? {} : { befund: samenbefund })}
          >
            {(verbindung) => (
              <input
                {...verbindung}
                type="text"
                inputMode="numeric"
                className={eingabe}
                value={same}
                disabled={vorgang.laeuft(setID)}
                onChange={(ereignis) => {
                  setzeSame(ereignis.target.value)
                }}
              />
            )}
          </Feld>

          <Knopf
            gewicht="haupt"
            bezeichnung={eigene.ziehen}
            angehalten={vorgang.laeuft(setID) || samenbefund !== undefined}
            aufKlick={() => {
              void ziehen()
            }}
          >
            {eigene.ziehen}
          </Knopf>

          {/* **Ein neuer Same ist ein Knopf und kein Zufall im Hintergrund.**
              Wer zieht, soll die Zahl sehen, mit der gezogen wurde — sonst ist
              das Ergebnis nicht wiederherstellbar. */}
          <Knopf
            bezeichnung={eigene.neuerSame}
            angehalten={vorgang.laeuft(setID) || samenbefund !== undefined}
            aufKlick={() => {
              // **Über `BigInt` und nicht über `Number`.** Oberhalb von 2⁵³
              // verliert eine JavaScript-Zahl die letzten Stellen — genau der
              // Grund, aus dem der Same als **Text** über die Bindung reist
              // ([App.PackZiehen]). Ein `+ 1`, das die Zahl verändert, nähme
              // dem Samen seine einzige Eigenschaft.
              setzeSame(String(samenzahl(same) + 1n))
            }}
          >
            {eigene.neuerSame}
          </Knopf>
        </div>
      )}

      {/* **Was zu keinem Feld gehört, steht über der Lasche** — ein Paket,
          das fehlt, nennt kein Feld. */}
      {fehler !== undefined && fehler.feld !== 'same' && (
        <Hinweiskasten art="fehler" meldung={fehler.meldung} />
      )}

      {gezogen !== undefined && (
        <section className="flex flex-col gap-2">
          <h3 className="text-sm font-medium">{eigene.gezogen}</h3>

          {gezogen.length === 0 ? (
            <p className="text-sm text-leise">{eigene.keineKarten}</p>
          ) : (
            /* **Vorn steht ein Booster, der Rest auf Verlangen** — UXR-08
               (#1916).
               Die erste wählbare Art von `magic.mse-game` ist das
               *tournament pack* mit 76 Karten, und sie steht ungefragt da;
               untereinander wären das gut 1.800 Punkte. Sechzehn zieht das
               *booster pack* desselben Spiels. */
            <Faltliste
              eintraege={gezogen}
              vorn={einBooster}
              schritt={100}
              geordnet
              klasse="flex flex-col gap-1 text-sm"
              zeile={(karte, stelle) => (
                <li key={`${karte.id}-${String(stelle)}`}>
                  {kartenname(karte, texte.seiten.karten.ohneName)}
                </li>
              )}
            />
          )}
        </section>
      )}

      {befunde.length > 0 && (
        <section className="flex flex-col gap-1">
          <h3 className="text-sm font-medium">{eigene.vorbehalte}</h3>
          <ul className="flex flex-col gap-1 text-sm text-leise">
            {befunde.map((befund) => (
              <li key={befund}>{befund}</li>
            ))}
          </ul>
        </section>
      )}
    </section>
  )
}

// einBooster ist, wie viele gezogene Karten zu Anfang dastehen: so viele, wie
// das *booster pack* von `magic.mse-game` zieht — 1 + 3 + 10 + 1 + 1
// (`packs`, nachgezählt am 2026-10-02).
const einBooster = 16

// hoechsterSame ist 2⁶⁴−1 — die Obergrenze von `strconv.ParseUint(…, 10, 64)`.
const hoechsterSame = 18446744073709551615n

// samenPasst sagt, ob Go diesen Text als Samen lesen kann.
//
// **Leer ist gültig**: `samenAus` gibt dafür null zurück, und wer keinen Samen
// angibt, bekommt eine Ziehung, die sich wiederholen lässt.
function samenPasst(text: string): boolean {
  return text === '' || (/^\d+$/.test(text) && BigInt(text) <= hoechsterSame)
}

// samenzahl liest den Samen als ganze Zahl — leer ist null.
function samenzahl(text: string): bigint {
  return text === '' ? 0n : BigInt(text)
}

// eingabe sind die gemeinsamen Klassen — dieselben wie an einem Kartenfeld.
const eingabe =
  'rounded-sm border border-kante bg-transparent px-2 py-1 text-sm ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring disabled:opacity-60 '
