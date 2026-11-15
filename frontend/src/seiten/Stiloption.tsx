// Eine Stiloption samt ihrem Bedienelement.
//
// **Eine eigene Datei**, weil `Stiloptionen.tsx` sonst zwei Fragen beantwortete
// (§2): dort steht, wie die Liste geladen und geschrieben wird, hier, wie eine
// einzelne Option aussieht.
//
// # Warum eine Auskunft gar kein Feld ist
//
// **`InfoField() { editable = false; }`** (`src/data/field/information.hpp`,
// Zeilen 26–28, Fassung f9d9356 vom 2021-07-06). Sie trägt keinen Wert; über
// das Pack sind alle 13 Vorkommen **Zwischenüberschriften** in langen
// Optionslisten — „Advanced customization", „Other stuff". Ein beschreibbares
// Feld dafür ist nicht zu wenig, sondern zu viel: es schriebe einen Wert in
// den Bestand, den kein Skript liest
// (#639).
//
// # Warum eine Mehrfachwahl anders eingefasst ist
//
// **Ein `<label>` beschriftet genau ein Bedienelement.** Bei einer
// Mehrfachwahl sind es mehrere, und geschachtelte `<label>` sind weder gültig
// noch für eine Vorlesehilfe auflösbar — der Name eines Kästchens käme dort
// gar nicht an. Sie bekommt deshalb ein `<fieldset>` mit `<legend>`, jede
// andere Art ein `<label>`.
import { type StiloptionDTO } from '../api'
import { useTexte } from '../sprache/kontext'
import { useWahlbilder } from '../haken/useWahlbilder'
import { Farbwahl } from './Farbwahl'
import { Mehrfachwahl } from './Mehrfachwahl'
import { Wahlbildreihe } from './Wahlbildreihe'
import { useLesegrund } from '../haken/schreibrecht'

interface StiloptionEigenschaften {
  // setID steht hier, weil die Wahlbilder gerastert werden müssen und der
  // Rasterlauf das Set braucht — es liefert Stil und Setangaben (#1225).
  setID: string

  option: StiloptionDTO
  laeuft: boolean
  setzen: (wert: string) => void

  // tippt meldet jeden Tastendruck im Textfeld — **ohne zu schreiben**
  // (#590).
  //
  // **Nur das Textfeld hat es nötig.** Eine Auswahl, ein Kästchen und ein
  // Farbwähler schreiben beim Umstellen; ein Textfeld erst beim Verlassen, und
  // Verlassen ist eine Bestätigung (§6). `undefined` heißt „getippt wird
  // nichts gemeldet" — die Vorschau bleibt dann beim geschriebenen Stand.
  tippt?: (wert: string) => void

  // zuruecksetzen stellt die Option auf ihre Vorgabe zurück — es **entfernt**
  // den gespeicherten Wert
  // (#764).
  //
  // **Nicht `setzen('')`**: leer ist bei einer Textoption ein gültiger Wert,
  // und zwei Bedeutungen an einem Aufruf wären später nicht mehr
  // auseinanderzuhalten.
  zuruecksetzen?: () => void
}

// vorgabewahl ist der Wert des Vorgabeeintrags in der Auswahlliste.
//
// **Er verlässt das Bauteil nie** — `onChange` fängt ihn ab und ruft
// `zuruecksetzen` statt `setzen`. Kollidieren kann er nicht: der Eintrag steht
// nur an einer Farboption, und deren Wahlen sind `rgb(…)`-Angaben aus der
// Vorlage.
const vorgabewahl = '\u0000vorgabe'

export function Stiloption({
  setID,
  option,
  laeuft,
  zuruecksetzen,
  setzen,
  tippt,
}: StiloptionEigenschaften) {
  const texte = useTexte()
  // **Ein Leser stellt nichts ein** (#2290): jede Option schreibt sofort.
  // **Und sie sagt, warum** (H01, #2255) — jede Eingabe an sich selbst, auch
  // in einer Gruppe: die Beschreibung eines `fieldset` spricht ein Vorleser
  // am einzelnen Feld nicht verlässlich (Nachlauf über Phase I, Befund 4).
  const lesegrund = useLesegrund()
  const gesperrt = laeuft || lesegrund.gesperrt
  const grund = lesegrund.beschreibe()

  // **Gefragt wird nur, wo es etwas gibt** — sonst holte diese Seite für jede
  // Option Bilder und bekäme bei `magic-m15` neunmal von dreizehn nichts.
  const wahlbilder = useWahlbilder(setID, option.name, option.hatWahlbilder)

  // **Nur wenn die Vorlage eine Vorgabe nennt** — und das ist etwas anderes,
  // als dass sie sich ausrechnet.
  //
  // Ein Vergleich wie `option.vorgabe !== ''` verlöre einen Fall:
  // `magic-m15-bigtext` schreibt `default: black` ohne Anführungszeichen. Das
  // ist ein Skriptfehler, der Wert fällt auf `initial:` zurück — und der
  // Eintrag verschwände. **MSE zeigt ihn dort**, denn `hasDefault()` fragt
  // nach dem Skript, nicht nach seinem Ergebnis.
  const hatVorgabe = option.hatVorgabe && zuruecksetzen !== undefined

  if (option.art === 'info') {
    return (
      <div className="mt-2 flex flex-col gap-1 border-t border-kante pt-3">
        <h4 className="font-semibold">{option.name}</h4>

        <Beschreibung option={option} />
      </div>
    )
  }

  // **Bilder statt Wörter, wo die Vorlage welche erklärt** (#1225). MSE zeigt
  // an dieser Stelle drei Tapsymbole.
  //
  // **Ein `<fieldset>` und kein `<label>`** — aus demselben Grund wie bei der
  // Mehrfachwahl: eine Reihe von Optionsfeldern sind mehrere Bedienelemente,
  // und ein `<label>` beschriftet genau eines.
  if (wahlbilder.length > 0) {
    return (
      <fieldset className="flex flex-col gap-1">
        <legend className="text-sm">{option.name}</legend>

        <Beschreibung option={option} />
        {lesegrund.satz}

        <Wahlbildreihe
          option={option}
          bilder={wahlbilder}
          laeuft={gesperrt}
          beschriebenVon={grund}
          setzen={setzen}
          zuruecksetzen={zuruecksetzen}
        />
      </fieldset>
    )
  }

  if (option.art === 'multiple choice') {
    return (
      <fieldset className="flex flex-col gap-1">
        <legend className="text-sm">{option.name}</legend>

        <Beschreibung option={option} />
        {lesegrund.satz}

        <Mehrfachwahl
          option={option}
          laeuft={gesperrt}
          beschriebenVon={grund}
          setzen={setzen}
        />
      </fieldset>
    )
  }

  return (
    <label className="flex flex-col gap-1">
      <span className="text-sm">{option.name}</span>

      <Beschreibung option={option} />
      {lesegrund.satz}

      {option.moeglichkeiten.length > 0 && (
        <select
          // **Der Vorgabeeintrag ist gewählt, solange nichts gesetzt ist** —
          // und nicht die Wahl, die zufällig dieselbe Farbe trägt. MSE hält es
          // ebenso: `selection()` gibt 0, sobald `value.isDefault()`
          // (`src/gui/value/color.cpp`, Zeilen 104–108, Fassung f9d9356 vom
          // 2021-07-06).
          value={hatVorgabe && !option.gesetzt ? vorgabewahl : option.wert}
          disabled={gesperrt}
          aria-describedby={grund}
          onChange={(ereignis) => {
            if (ereignis.target.value === vorgabewahl) {
              zuruecksetzen?.()

              return
            }

            setzen(ereignis.target.value)
          }}
          className="rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
        >
          {/* **Zuoberst und nur, wenn die Vorlage eine Vorgabe nennt** — MSE
              blendet ihn ohne `default_script` ganz aus (`color.cpp`, Zeile
              38, Fassung f9d9356 vom 2021-07-06). */}
          {hatVorgabe && (
            <option value={vorgabewahl}>{texte.seiten.stil.vorgabe}</option>
          )}

          {option.moeglichkeiten.map((moeglich) => (
            <option key={moeglich.wert} value={moeglich.wert}>
              {moeglich.beschriftung}
            </option>
          ))}
        </select>
      )}

      {option.art === 'boolean' && (
        <input
          type="checkbox"
          checked={option.wert === 'yes'}
          disabled={gesperrt}
          aria-describedby={grund}
          onChange={(ereignis) => {
            // **`yes` und `no`, wie die Pakete es schreiben** — nicht
            // `true`/`false`. Ein Feldskript liest `styling.…` und vergleicht
            // mit dem, was in der Vorlage steht.
            setzen(ereignis.target.checked ? 'yes' : 'no')
          }}
          className="size-4 self-start focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
        />
      )}

      {option.eigeneFarbe && (
        <Farbwahl
          option={option}
          laeuft={gesperrt}
          beschriebenVon={grund}
          setzen={setzen}
        />
      )}

      {option.moeglichkeiten.length === 0 &&
        option.art !== 'boolean' &&
        option.art !== 'color' && (
          <input
            type="text"
            defaultValue={option.wert}
            disabled={gesperrt}
            aria-describedby={grund}
            onChange={(ereignis) => {
              tippt?.(ereignis.target.value)
            }}
            onBlur={(ereignis) => {
              if (ereignis.target.value !== option.wert) {
                setzen(ereignis.target.value)
              }
            }}
            className="rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
          />
        )}
    </label>
  )
}

// Beschreibung ist der Satz, den das Paket mitbringt — **unübersetzt** (§9).
function Beschreibung({ option }: { option: StiloptionDTO }) {
  if (option.beschreibung === '') {
    return null
  }

  return <span className="text-sm text-leise">{option.beschreibung}</span>
}
