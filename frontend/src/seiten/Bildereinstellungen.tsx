// Die Eingaben der Bildausgabe: Ordner, Vorlage, Format, Breite.
//
// **Eine eigene Datei** (AGENTS.md §2): *was der Lauf tut* steht in `Bilder`,
// *wie er eingestellt wird* hier. Der Schnitt folgt der Frage, nicht der
// Zeilenzahl.
import { useId } from 'react'

import { Zahlenfeld } from '../bausteine/Zahlenfeld'
import { useTexte } from '../sprache/kontext'

interface BildereinstellungenEigenschaften {
  ordner: string

  // gemerkt sagt, dass der Ordner der zuletzt gewählte ist und nicht einer
  // von eben; merkfehler, dass er sich nicht merken ließ (AP-10, #2367).
  gemerkt: boolean
  merkfehler?: string
  vorlage: string
  format: string
  breite: number
  laeuft: boolean

  // konflikt ist, was bei einem Namenskonflikt geschieht — die Wahl und die
  // möglichen Werte
  // (#491).
  //
  // **Sie steht hier und nicht auf der Einstellungsseite**, weil sie zu dem
  // Lauf gehört, den man gerade startet. MSE hält sie ebenso im
  // Ausgabefenster (`images_export_window.cpp`, Fassung f9d9356 vom
  // 2021-07-06).
  //
  // **`undefined`, solange die Einstellungen nicht da sind** — dann steht die
  // Auswahl nicht da, statt eine Wahl zu zeigen, die niemand getroffen hat.
  konflikt: { wert: string; werte: string[] } | undefined

  setzeVorlage: (vorlage: string) => void
  setzeFormat: (format: string) => void
  guete: string
  setzeGuete: (guete: string) => void
  setzeBreite: (breite: number) => void
  setzeKonflikt: (wert: string) => void
  waehleOrdner: () => void
}

export function Bildereinstellungen({
  ordner,
  gemerkt,
  merkfehler,
  vorlage,
  format,
  breite,
  laeuft,
  konflikt,
  setzeVorlage,
  setzeFormat,
  guete,
  setzeGuete,
  setzeBreite,
  setzeKonflikt,
  waehleOrdner,
}: BildereinstellungenEigenschaften) {
  const texte = useTexte()
  const vorlagehilfe = useId()
  const breitehilfe = useId()
  const konflikthilfe = useId()

  const feld =
    'mt-1 w-full rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60'

  return (
    <div className="mt-4 space-y-4">
      <div className="flex flex-wrap items-center gap-3">
        <button
          type="button"
          disabled={laeuft}
          onClick={waehleOrdner}
          className="rounded-sm border border-kante px-3 py-1 whitespace-nowrap focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
        >
          {texte.seiten.bilder.ordnerWaehlen}
        </button>

        <span className="font-mono text-sm text-leise">
          {ordner === '' ? texte.seiten.bilder.keinOrdner : ordner}
        </span>
        {gemerkt && (
          <span className="text-sm text-leise">
            {texte.seiten.bilder.gemerkt}
          </span>
        )}
      </div>

      {merkfehler !== undefined && (
        <p role="alert" className="max-w-prose text-sm text-fehler">
          {merkfehler}
        </p>
      )}

      <div className="grid max-w-2xl gap-4 sm:grid-cols-3">
        <label className="block">
          <span className="text-sm">{texte.seiten.bilder.vorlage}</span>
          <input
            type="text"
            value={vorlage}
            disabled={laeuft}
            aria-describedby={vorlagehilfe}
            onChange={(ereignis) => {
              setzeVorlage(ereignis.target.value)
            }}
            className={feld}
          />
        </label>

        <label className="block">
          <span className="text-sm">{texte.seiten.bilder.format}</span>
          <select
            value={format}
            disabled={laeuft}
            onChange={(ereignis) => {
              setzeFormat(ereignis.target.value)
            }}
            className={feld}
          >
            <option value="png">{texte.seiten.bilder.formatPNG}</option>
            <option value="jpg">{texte.seiten.bilder.formatJPG}</option>
            {/* **BMP und TIFF, weil MSE2 sie anbietet** (#826). BMP legt ein
                durchsichtiges Bild auf Weiß, TIFF behält die
                Durchsichtigkeit — beides gemessen in
                `bildformate_test.go`. */}
            <option value="bmp">{texte.seiten.bilder.formatBMP}</option>
            <option value="tiff">{texte.seiten.bilder.formatTIFF}</option>
            {/* **SVG entsteht aus dem Zeichenplan, nicht aus Bildpunkten**
                (#852). Flächen und Text bleiben Geometrie und skalieren mit;
                Kartenrahmen und Symbole sind eingebettete Rasterbilder, weil
                die Vorlagen sie so mitbringen. */}
            <option value="svg">{texte.seiten.bilder.formatSVG}</option>
          </select>
        </label>

        {/* **Die Güte ist eine Wahl und kein Schluss aus der Breite**
            (Vermerk 0049). Sie steht neben dem Format, weil beide sagen, wie
            das Bild aussieht — und nicht neben der Breite, die sagt, wie groß
            es wird.

            **Über zwei Spalten** (Werkstück C von Phase 4.10): in einer
            stand „Nach der Regel von Mag…" — ein Auswahlfeld, das seine
            längste Wahl nicht zeigen kann, verlangt vom Lesenden, sie zu
            erraten. */}
        <label className="block sm:col-span-2">
          <span className="text-sm">{texte.seiten.bilder.guete}</span>
          <select
            value={guete}
            disabled={laeuft}
            onChange={(ereignis) => {
              setzeGuete(ereignis.target.value)
            }}
            className={feld}
          >
            <option value="fast">{texte.seiten.bilder.gueteFast}</option>
            <option value="medium">{texte.seiten.bilder.gueteMedium}</option>
            <option value="high">{texte.seiten.bilder.gueteHigh}</option>
            <option value="ultra">{texte.seiten.bilder.gueteUltra}</option>
          </select>
        </label>

        {/* **Ein Zahlenfeld, das seinen Bereich am Feld sagt** — Werkstück F
            von Phase 4.10
            (#1325).

            Ein bloßes `<input type="number">` mit `min` und `max` **meldet
            nichts**: im laufenden Fenster gemessen gab `checkValidity()` nach
            `99999` zwar `false` und `validationMessage` einen fertigen Satz,
            im Fenster stand aber kein `aria-invalid`, kein `role=alert` und
            kein Hinweis. Der Browser meldet das beim Absenden eines
            Formulars, und hier wird nichts abgesendet. */}
        <Zahlenfeld
          beschriftung={texte.seiten.bilder.breite}
          wert={breite}
          kleinstens={1}
          groesstens={10000}
          angehalten={laeuft}
          beschreibung={breitehilfe}
          geaendert={setzeBreite}
        />
      </div>

      {/* **Was bei einem Namenskonflikt geschieht** (#491). Bis dahin wurde
          überschrieben, ohne zu fragen — und das Muster `{card.name}` erzeugt
          für zwei Länder desselben Namens denselben Dateinamen. */}
      {konflikt !== undefined && (
        <label className="block max-w-md">
          <span className="text-sm">{texte.seiten.bilder.konflikt}</span>
          <select
            value={konflikt.wert}
            disabled={laeuft}
            aria-describedby={konflikthilfe}
            onChange={(ereignis) => {
              setzeKonflikt(ereignis.target.value)
            }}
            className={feld}
          >
            {konflikt.werte.map((wert) => (
              <option key={wert} value={wert}>
                {
                  texte.seiten.bilder.konflikte[
                    wert as keyof typeof texte.seiten.bilder.konflikte
                  ]
                }
              </option>
            ))}
          </select>
        </label>
      )}

      {/* **Die Erklärung steht neben dem Label, nicht darin.** Ein `<label>`,
          das Beschriftung **und** Erklärung umschließt, macht beides zum Namen
          des Feldes: ein Vorleseprogramm sagte dann drei Sätze, wo einer
          gemeint ist. Die anderen Felder hier machen es genauso. */}
      {konflikt !== undefined && (
        <p id={konflikthilfe} className="max-w-prose text-sm text-leise">
          {texte.seiten.bilder.konfliktHilfe}
        </p>
      )}

      <p id={vorlagehilfe} className="max-w-prose text-sm text-leise">
        {texte.seiten.bilder.vorlageHilfe}
      </p>

      {/* **Warum keine dpi-Angabe steht**, gehört neben das Feld und nicht in
          einen Vermerk allein (0012): wer 300 sucht und 750 tippen soll, will
          den Grund sehen. */}
      <p id={breitehilfe} className="max-w-prose text-sm text-leise">
        {texte.seiten.bilder.breiteHilfe}
      </p>
    </div>
  )
}
