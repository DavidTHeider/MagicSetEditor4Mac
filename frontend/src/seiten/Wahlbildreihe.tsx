// Die Wahlen einer Stiloption als Bilder
// (#1225).
//
// # Warum Kacheln und keine Auswahlliste
//
// **Ein `<option>` trägt kein Bild.** HTML lässt in einer Auswahlliste nur
// Text zu; wer dort ein Bild zeigen will, baut eine Liste nach — und eine
// nachgebaute Auswahlliste ist mit Tastatur, Vorlesehilfe und
// Bildschirmtastatur eine dauerhafte Baustelle.
//
// **Eine Reihe von Optionsfeldern ist dieselbe Sache mit fertiger Bedienung**:
// eine Wahl aus mehreren, Pfeiltasten dazwischen, ein Fokusring, und der Name
// steht neben dem Bild statt darunter zu verschwinden.
//
// # Warum das nicht ausufert
//
// Weil das Material es nicht hergibt. Gezählt über `full-magic-pack/data` am
// 2026-08-28 führt **kein** `choice images:`-Block mehr als **fünf** Wahlen:
// 591 führen zwei, 500 drei, fünf führen fünf. Eine Reihe aus höchstens fünf
// Kacheln passt neben die Beschriftung.
//
// # Wie das Bild zur Wahl findet
//
// **Über die Beschriftung.** MSE schlägt das Bild unter `choiceName(i)` nach
// (`src/gui/value/choice.cpp`, Zeile 175, Fassung f9d9356 vom 2021-07-06), und
// das ist bei einer Auswahl und einem Schalter der Wert, bei einer Farbe aber
// der **Name**: gespeichert wird `rgb(0,0,0)`, erklärt ist `schwarz`.
import { type StiloptionDTO, type WahlbildDTO } from '../api'
import { useTexte } from '../sprache/kontext'

interface WahlbildreiheEigenschaften {
  option: StiloptionDTO
  bilder: WahlbildDTO[]
  laeuft: boolean

  // beschriebenVon nennt den Grund, wenn die Wahl für einen Leser gesperrt
  // ist — **an jedem Optionsfeld**, nicht nur an der Gruppe (H01, #2255).
  beschriebenVon?: string | undefined
  setzen: (wert: string) => void

  // zuruecksetzen entfernt den gespeicherten Wert — dieselbe Naht wie an der
  // Auswahlliste (#764).
  zuruecksetzen: (() => void) | undefined
}

export function Wahlbildreihe({
  option,
  bilder,
  laeuft,
  beschriebenVon,
  setzen,
  zuruecksetzen,
}: WahlbildreiheEigenschaften) {
  const texte = useTexte()

  // **Dieselbe Regel wie an der Auswahlliste**: der Vorgabeeintrag steht nur
  // da, wenn die Vorlage ein `default:` nennt, und er ist gewählt, solange
  // nichts gesetzt ist (`src/gui/value/color.cpp`, Zeilen 38 und 104–108,
  // Fassung f9d9356 vom 2021-07-06).
  const vorgabe = option.hatVorgabe && zuruecksetzen !== undefined

  return (
    <div className="flex flex-wrap gap-2">
      {/* **Zuoberst, wie bei MSE.**
          
          **Und im Pack kommt er hier nie vor**: von den 1.096 Optionen mit
          `choice images:` nennt **keine einzige** ein `default:` (gezählt am
          2026-08-28). Er steht trotzdem da, weil die Kacheln sonst können,
          was die Auswahlliste kann, minus das Zurücksetzen — und diese
          Ungleichheit fiele erst der ersten Vorlage auf, die beides führt. */}
      {vorgabe && (
        <Kachel
          name={option.name}
          beschriftung={texte.seiten.stil.vorgabe}
          bild={undefined}
          gewaehlt={!option.gesetzt}
          laeuft={laeuft}
          beschriebenVon={beschriebenVon}
          waehle={zuruecksetzen}
        />
      )}

      {option.moeglichkeiten.map((moeglich) => {
        const bild = bilder.find(
          (eines) => eines.wahl === moeglich.beschriftung,
        )

        return (
          <Kachel
            key={moeglich.wert}
            name={option.name}
            beschriftung={moeglich.beschriftung}
            bild={bild?.bild}
            gewaehlt={
              !vorgabe || option.gesetzt ? option.wert === moeglich.wert : false
            }
            laeuft={laeuft}
            beschriebenVon={beschriebenVon}
            waehle={() => {
              setzen(moeglich.wert)
            }}
          />
        )
      })}
    </div>
  )
}

interface KachelEigenschaften {
  name: string
  beschriftung: string
  bild: string | undefined
  gewaehlt: boolean
  laeuft: boolean
  beschriebenVon: string | undefined
  waehle: () => void
}

// Kachel ist eine Wahl: ein Optionsfeld, ein Bild und der Name.
function Kachel({
  name,
  beschriftung,
  bild,
  gewaehlt,
  laeuft,
  beschriebenVon,
  waehle,
}: KachelEigenschaften) {
  return (
    <label className="flex cursor-pointer flex-col items-center gap-1 rounded-sm border border-kante px-2 py-1 text-sm has-checked:border-marke has-checked:bg-betont focus-within:outline focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-ring">
      <input
        type="radio"
        name={name}
        value={beschriftung}
        checked={gewaehlt}
        disabled={laeuft}
        aria-describedby={beschriebenVon}
        onChange={waehle}
        // **Nicht `hidden` und nicht `display: none`.** Ein so entferntes
        // Optionsfeld ist auch für die Tastatur weg; die Kachel bekäme dann
        // gar keinen Fokus. `sr-only` nimmt es aus dem Bild und lässt es im
        // Baum.
        className="sr-only"
      />

      {/* **Ohne Bild bleibt der Name, und der Platz bleibt frei.**
          
          Ein `<img>` ohne Quelle zeigte ein zerbrochenes Symbol und sähe nach
          einem Fehler der Anwendung aus (§8); eine Kachel ohne den Platz
          stünde niedriger als ihre Nachbarn. MSE hält es ebenso — seine
          Listenzeilen haben eine feste Höhe (`min_item_size = thumbnail_size`,
          `src/gui/value/choice.cpp`, Zeile 18, Fassung f9d9356 vom
          2021-07-06), und der Platz bleibt leer.
          
          **Der Fall ist keiner aus dem Lehrbuch**: `magic-m15-altered` erklärt
          an `center text` vier Wahlen und drei Bilder. */}
      {bild === undefined ? (
        <span className="block size-6" />
      ) : (
        <img
          src={bild}
          alt=""
          width={24}
          height={24}
          className="block size-6 object-contain"
        />
      )}

      <span>{beschriftung}</span>
    </label>
  )
}
