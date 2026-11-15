// Ein **Feld der Schlüsselwortzeile** — mit dem Menü, das etwas hineinschreibt.
//
// **Eine eigene Datei**, seit die Parameterhilfen dazukamen
// (#1342):
// `Schluesselwortzeile.tsx` beantwortet „wie sieht ein Schlüsselwort aus",
// diese Datei „wie schreibt man in eines seiner Felder" (§2).
import type { ReactNode, RefObject } from 'react'

import { Entwurfskonflikt } from '../bausteine/Entwurfskonflikt'
import { Feld } from '../bausteine/Feld'
import type { Entwurfsgriff } from '../haken/useEntwurf'
import { useLesegrund } from '../haken/schreibrecht'

export interface SchluesselwortfeldEigenschaften {
  aufschrift: string
  wert: string
  angehalten: boolean
  breit?: boolean

  // entwurf hält, was getippt und noch nicht geschrieben ist (U06, #2122).
  //
  // **Der Aufrufer hält ihn und nicht dieses Bauteil** — aus demselben Grund
  // wie `feld` darunter: der Einschub aus dem Menü schreibt an diesem Feld
  // vorbei und muss es dem Entwurf sagen (`uebernehmen`), sonst erschiene
  // sein eigener Text danach als fremder Wert.
  entwurf: Entwurfsgriff

  // feld ist der Griff auf den Knoten — für das Einfügen an der Textmarke.
  //
  // **Der Aufrufer hält ihn und nicht dieses Bauteil.** Der erste Anlauf gab
  // die Einfügefunktion als Rückruf nach außen, und `react-hooks/refs` wies
  // es ab: eine Funktion, die `ref.current` liest, darf nicht beim Zeichnen
  // weitergereicht werden. Wer den Griff hält, ruft sie aus einem Ereignis.
  feld?: RefObject<HTMLInputElement | null>

  // befund ist der Satz, der zu **diesem** Feld gehört — Werkstück U3 von
  // Phase 5.1 (#1414).
  //
  // **Er hängt am Muster und nicht an der Zeile.** Ob ein Schlüsselwort
  // greift, entscheidet `service.musterGreift` (portiert aus
  // `Keyword::prepare`) — und zwar am **Muster**; §6 verlangt den Satz am
  // verursachenden Feld.
  befund?: string | undefined

  // menue steht unter dem Feld.
  menue?: ReactNode
}

// eingabe sind die gemeinsamen Klassen — dieselben wie an einem Kartenfeld.
export const eingabe =
  'rounded-sm border border-kante bg-transparent px-2 py-1 text-sm ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring disabled:opacity-60 '

// Schluesselwortfeld ist eine beschriftete Eingabe — abgeschaltet bei einem
// des Spiels.
//
// **Geschrieben wird beim Verlassen**, nicht bei jedem Tastendruck: dieselbe
// Regel wie an einem Kartenfeld, und aus demselben Grund — ein Schreibvorgang
// je Buchstabe wären hundert Verlaufsschritte für ein Wort.
export function Schluesselwortfeld({
  aufschrift,
  wert,
  angehalten,
  breit,
  entwurf,
  feld,
  befund,
  menue,
}: SchluesselwortfeldEigenschaften) {
  // **Ein Leser schreibt nicht** (#2290) — das Feld bleibt lesbar und sagt,
  // warum (H01, #2255).
  const lesegrund = useLesegrund()
  const lesend = lesegrund.gesperrt

  return (
    <div
      className={
        'flex flex-col gap-1' + (breit === true ? '' : ' flex-1 basis-40')
      }
    >
      {/* **Beschriftung und Befund kommen aus [Feld]** — dieselbe Naht wie
          bei `Pflichtfeld` und `Zahlenfeld` (Werkstück U2, #1413). */}
      <Feld
        beschriftung={aufschrift}
        {...(befund === undefined ? {} : { befund })}
      >
        {(verbindung) => (
          <input
            {...verbindung}
            key={entwurf.schluessel}
            ref={feld}
            type="text"
            className={eingabe}
            defaultValue={wert}
            disabled={angehalten}
            readOnly={lesend}
            aria-describedby={lesegrund.beschreibe(
              verbindung['aria-describedby'],
            )}
            onChange={(ereignis) => {
              entwurf.aendern(ereignis.target.value)
            }}
            onBlur={(ereignis) => {
              entwurf.verlassen(ereignis.target.value)
            }}
          />
        )}
      </Feld>

      <Entwurfskonflikt griff={entwurf} fremd={wert} angehalten={angehalten} />

      {menue}
      {lesegrund.satz}
    </div>
  )
}

// einfuegenAnDerMarke schreibt einen Text an die Textmarke eines Feldes.
//
// **An der Textmarke** — so tut es MSE (`TextCtrl::insert`, gerufen aus
// `KeywordsPanel::onCommand`, `src/gui/set/keywords_panel.cpp` Zeilen 228 und
// 231, Fassung f9d9356 vom 2021-07-06). Ein Anhängen ans Ende wäre die halbe
// Hilfe: `<atom-param>` steht in der Mitte eines Musters, nicht dahinter.
//
// **Und geschrieben wird sofort.** Das Feld ist ungesteuert und schreibt sonst
// erst beim Verlassen; wer nach dem Einfügen gleich das zweite Menü benutzt,
// verlöre den ersten Einschub.
export function einfuegenAnDerMarke(
  feld: RefObject<HTMLInputElement | null>,
  text: string,
  schreiben: (wert: string) => void,
): void {
  const eingang = feld.current

  /* v8 ignore next 3 -- unerreichbar: gerufen wird aus einem Ereignis des
     Menüs unter dem Feld, und dann steht der Knoten. Der Typ verlangt die
     Prüfung trotzdem. */
  if (eingang === null) {
    return
  }

  /* v8 ignore next -- `selectionStart` ist nur bei Eingabearten ohne Auswahl
     null (`email`, `number`, `url`); dieses Feld ist `text`. Der Typ verlangt
     den Rückfall trotzdem. */
  const anfang = eingang.selectionStart ?? eingang.value.length

  /* v8 ignore next -- dasselbe für das Ende der Auswahl. */
  const ende = eingang.selectionEnd ?? anfang

  const neu = eingang.value.slice(0, anfang) + text + eingang.value.slice(ende)

  eingang.value = neu
  eingang.setSelectionRange(anfang + text.length, anfang + text.length)
  eingang.focus()

  schreiben(neu)
}
