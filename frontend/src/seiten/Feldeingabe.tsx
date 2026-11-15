// Das Eingabefeld eines Kartenfeldes — **in der Form, die seine Art verlangt**.
//
// # Warum es das gibt
//
// Bis #560 war
// jedes Feld ein einzeiliges Textfeld: der Regeltext, die Seltenheit, die
// Manakosten, die Kennung. Was ein Feld ist, steht im Spielpaket — `type:`,
// `multi line:`, die `choice:`-Liste — und kam nie im Frontend an.
//
// # Der Rückfall ist der Normalfall
//
// Ein Set kann Felder führen, die sein Spiel nicht erklärt — ein fremdes Set,
// ein älteres Paket, ein Feld aus einem Modul, das fehlt. Ohne Beschreibung
// bleibt es ein Textfeld; das ist die Form, in der jeder Wert dasteht.
import { useId } from 'react'

import type { FeldbeschreibungDTO } from '../api'
import { Suchauswahl } from '../bausteine/Suchauswahl'
import { suchschwelle } from '../bausteine/suchwahl'
import { mitAuszeichnung } from './mitauszeichnung'
import { ohneAuszeichnung } from './ohneauszeichnung'
import { tippen } from './feldtippen'
import { useLesegrund } from '../haken/schreibrecht'

interface FeldeingabeEigenschaften {
  feld: string
  wert: string

  // form ist die Beschreibung aus dem Spielpaket — `undefined`, wenn das Spiel
  // dieses Feld nicht erklärt.
  form: FeldbeschreibungDTO | undefined

  angehalten: boolean
  fehlerhaft: boolean
  beschriebenVon: string | undefined

  // hergeleitet ist der gerechnete Wert — als **Platzhalter** (#563).
  //
  // **Nicht `vorgabe`**: so heißt bei [FeldbeschreibungDTO] der Anfangswert
  // des **Feldes** (#562), und das ist etwas anderes — er gilt für jede Karte,
  // dieser hier für diese eine.
  //
  // **Nur für die tippbaren Arten.** Eine Auswahlliste hat keinen Platzhalter,
  // und eine Ankreuzfläche auch nicht; dort sagt es die Zeile darunter.
  hergeleitet?: string | undefined

  aendern: (wert: string) => void
  verlassen: (wert: string) => void

  // autoErsetzen wendet die `auto replace:`-Regeln des Spiels an der
  // Schreibmarke an — `undefined`, wo es keine gibt.
  //
  // **Nur beim Tippen**, und das ist portiert: `TextValueEditor::onChar` gibt
  // als einziger Aufruf `allow_auto_replace = true` mit
  // (`src/gui/value/text.cpp`, Zeilen 497 und 499, Fassung f9d9356 vom
  // 2021-07-06). Einfügen, Löschen, Rückschritt und Zeilenumbruch ersetzen
  // nicht — wer eingefügten Text ersetzte, änderte, was jemand bewusst
  // hineingelegt hat.
  autoErsetzen?:
    | ((
        text: string,
        stelle: number,
      ) => Promise<{ text: string; stelle: number }>)
    | undefined
}

// eingabeklassen sind die gemeinsamen Klassen jeder Eingabe dieser Seite.
//
// **Ausgeliehen und nicht abgeschrieben** ([Wortlistenwahl] braucht dieselben):
// eine zweite Zeichenkette hätte eine zweite Randfarbe, und `docs/KONTRASTE.md`
// führte sie als zwei Fälle, von denen einer irgendwann auseinanderliefe.
export const eingabeklassen =
  'rounded-sm border border-kante bg-transparent px-2 py-1 ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring aria-[invalid=true]:border-fehler ' +
  'disabled:opacity-60'

// Feldeingabe ist ein Kartenfeld oder eine Setangabe in der Form, die das
// Spielpaket nennt.
//
// **Ein Leser schreibt nicht** (#2290): ein Textfeld bleibt lesbar und
// fokussierbar, aber schreibgeschützt; was sofort schreibt — Auswahl und
// Ankreuzfeld —, ist gesperrt. **Und jede Form sagt, warum** (H01, #2255):
// der Satz steht neben ihr und in ihrer Beschreibung, wie an einem Knopf.
export function Feldeingabe(eigenschaften: FeldeingabeEigenschaften) {
  const lesegrund = useLesegrund()

  return (
    <>
      <Form
        {...eigenschaften}
        lesend={lesegrund.gesperrt}
        beschriebenVon={lesegrund.beschreibe(eigenschaften.beschriebenVon)}
      />
      {lesegrund.satz}
    </>
  )
}

function Form({
  feld,
  wert,
  form,
  angehalten,
  fehlerhaft,
  beschriebenVon,
  hergeleitet,
  aendern,
  verlassen,
  autoErsetzen,
  lesend,
}: FeldeingabeEigenschaften & { lesend: boolean }) {
  const kennung = useId()
  const gesperrt = angehalten || lesend

  // getippt reicht eine Eingabe weiter und ersetzt, wenn getippt wurde
  // (`feldtippen.ts`).
  const getippt = (
    ereignis: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
    nachher: (wert: string) => void,
  ) => {
    tippen(ereignis, nachher, autoErsetzen)
  }

  // gemeinsameAngaben sind die, die jede Form trägt.
  const gemeinsameAngaben = {
    'aria-label': feld,
    'aria-invalid': fehlerhaft,
    'aria-describedby': beschriebenVon,
    disabled: angehalten,
    className: eingabeklassen,
  }

  // **Der gerechnete Wert steht auch im Steuerelement, nicht nur daneben.**
  //
  // Der erste Anlauf reichte ihn nur als Platzhalter durch — und einen
  // Platzhalter haben nur Textfelder. Bei einer Auswahl und einer
  // Ankreuzfläche stand er im Hinweis darunter, während das Steuerelement
  // seinen ersten Eintrag zeigte und die Karte mit dem gerechneten gezeichnet
  // wurde: **drei Antworten auf dieselbe Frage** (Zweitblick über Phase G).
  //
  // **Ein eigener Wert schlägt ihn**: `wert` steht vorn. Wer `no` schreibt,
  // meint `no`, auch wenn die Herleitung `yes` sagt — und [Feldzeile] gibt
  // `hergeleitet` ohnehin nur weiter, solange das Feld leer ist.
  const gezeigt = wert === '' ? (hergeleitet ?? '') : wert

  // **Getippt wird der Klartext, geschrieben der ausgezeichnete Wert**
  // (#687).
  //
  // Gesehen an der gebauten Anwendung: das Feld **Type** zeigte
  // `<word-list-type-en></word-list-type-en><sep-soft> — </sep-soft>` — als
  // das Erste, was jemand an diesem Feld sieht. Wer die Marken löscht,
  // verliert die Wortliste, und er sieht nirgends, warum.
  //
  // **Nur für die getippten Arten.** Eine Auswahl schreibt einen ihrer
  // Einträge zurück, eine Ankreuzfläche `yes` oder `no`; beide tragen keine
  // Auszeichnung, und ein Rückweg dorthin wäre ein Zweig ohne Fall.
  // **Der eigene Wert, nicht die Herleitung.** `gezeigt` setzt die Herleitung
  // an die Stelle eines fehlenden Wertes — für die Auswahl und die
  // Ankreuzfläche ist das richtig, denn sie haben keinen Platzhalter. Ein
  // Textfeld hat einen, und wer die Herleitung dort als **Wert** einsetzte,
  // machte aus einer Rechnung eine Eingabe: beim nächsten Speichern stünde sie
  // in der Datei.
  const klartext = ohneAuszeichnung(wert)
  const zurueck = (getippt: string) => mitAuszeichnung(wert, getippt)

  // **Der gerechnete Wert als Platzhalter** (#563) — an der Stelle, an der der
  // Wert stünde. `placeholder={undefined}` schreibt kein Attribut, also
  // braucht es keinen zweiten Zweig.
  //
  // **Mit eigener Farbe, und die ist gemessen.** Tailwinds Vorgabe ist
  // `currentColor` zu 50 % — im hellen Erscheinungsbild ergibt das
  // **3,34 : 1** gegen den Anwendungsgrund und **3,41 : 1** gegen eine Fläche,
  // beides unter der Schwelle 4,5 aus WCAG 1.4.3. Gefunden hat es die
  // Sichtprüfung; axe meldet `color-contrast` unter jsdom als *unentschieden*,
  // und ein Platzhalter ist Text.
  //
  // `slate-600` / `slate-400` sind die Nebentextfarben dieses Projekts und
  // stehen gemessen in `docs/KONTRASTE.md` — 6,92 und 7,66 gegen den
  // Anwendungsgrund.
  const tippbar = {
    ...gemeinsameAngaben,
    readOnly: lesend,
    // **Geprüft, aber nicht ersetzt** — G03
    // (#2250).
    // MSE2 unterringelt und ersetzt nie von selbst (`check_spelling`,
    // `src/script/functions/spelling.cpp`, Fassung f9d9356 vom 2021-07-06);
    // ein Feld in WebKit korrigiert nach der Systemeinstellung, und ein
    // Kartenname, den das System nicht kennt, würde still „berichtigt".
    // Die Unterringelung bleibt: sie ändert nichts am Text.
    autoCorrect: 'off',
    autoCapitalize: 'off',
    // **Auch der Platzhalter ohne seine Marken** (#687). Das war der gemeldete
    // Fall: `type` einer frischen Karte trägt keinen Wert, und der gerechnete
    // `<word-list-type-en>…</word-list-type-en>` stand roh im Feld.
    placeholder:
      hergeleitet === undefined ? undefined : ohneAuszeichnung(hergeleitet),
    className:
      eingabeklassen + ' placeholder:text-leise dark:placeholder:text-leise',
  }

  const art = form?.art ?? 'text'

  // **Ein `info`-Feld ist eine Auskunft, kein Eingabefeld.**
  //
  //	class InfoField : public Field {
  //	  … // Only for showing a caption
  //
  // `src/data/field/information.hpp`, Fassung f9d9356 vom 2021-07-06.
  if (art === 'info') {
    return (
      <input
        type="text"
        readOnly
        defaultValue={klartext}
        {...gemeinsameAngaben}
        // **Und es sieht aus, als ließe es sich nicht ändern**
        // (#674).
        //
        // `readOnly` allein ist unsichtbar: das Feld sah aus wie jedes andere,
        // und wer hineinschrieb, verlor seine Eingabe ohne ein Wort. Der Grund
        // in gedämpfter Fläche sagt es, bevor jemand es versucht.
        className={eingabeklassen + ' bg-betont/60/60'}
      />
    )
  }

  // **Ein Wahrheitswert wird angekreuzt.** Seine Wahlen heißen bei MSE `yes`
  // und `no` (`src/data/field/boolean.cpp`, Zeilen 14–18, Fassung f9d9356 vom
  // 2021-07-06); geschrieben wird derselbe Text zurück.
  if (art === 'boolean') {
    return (
      <input
        type="checkbox"
        checked={gezeigt === 'yes'}
        onChange={(ereignis) => {
          const gesetzt = ereignis.target.checked ? 'yes' : 'no'

          aendern(gesetzt)
          verlassen(gesetzt)
        }}
        {...gemeinsameAngaben}
        disabled={gesperrt}
        className="size-4 accent-marke focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
      />
    )
  }

  // **Eine Auswahl wird gewählt, nicht getippt.**
  //
  // **Und sie schreibt sofort.** Ein `select` hat kein „fertig getippt": wer
  // wählt, hat gewählt.
  //
  // **Gesteuert und nicht vorbelegt.** Der erste Anlauf setzte `defaultValue`
  // — und damit galt die Herleitung nur beim **ersten** Einhängen. Die
  // Probevorschau rechnet sie bei jedem Tastendruck neu; wer die Manakosten
  // ändert, sah danach die Karte in der neuen Farbe, den Hinweis unter dem
  // Feld mit dem neuen Wert und in der Liste den **alten**. Der Nachlauf über
  // Phase G hat es gefunden — derselbe Befund wie oben, eine Ebene tiefer.
  //
  // **Ein Wert, der nicht in der Liste steht, steht trotzdem da**, und das ist
  // nachgelesen:
  //
  //	if (value.empty()) return;
  //	…
  //	String text = tr(viewer.getStylePackage(), value, capitalize_sentence);
  //
  // `src/render/value/choice.cpp`, Zeilen 63 und 97, Fassung f9d9356 vom
  // 2021-07-06 — MSE zeichnet den Wert selbst, gleich ob die Liste ihn führt,
  // und bei einem leeren gar nichts. Ohne den zusätzlichen Eintrag zeigte ein
  // gesteuertes `select` den **ersten** Eintrag, während das Feld leer ist:
  // eine Wahl, die niemand getroffen hat.
  // **Ab fünfzehn Wahlen eine Auswahl mit Suche** — UXR-F1
  // (#2007),
  // R7: das Wasserzeichen von `magic` hat 229. **Sie schreibt wie das native
  // Feld sofort**, trägt den Namen des Feldes als `aria-label` (eine
  // Kartenzeile hat kein `<label>`) und zeigt auch einen Wert, den die Liste
  // nicht führt, wie er ist.
  if (
    art === 'choice' &&
    form !== undefined &&
    form.wahlen.length >= suchschwelle
  ) {
    return (
      <Suchauswahl
        id={kennung}
        bezeichnung={feld}
        benennen
        eintraege={form.wahlen.map((wahl) => ({ wert: wahl, name: wahl }))}
        wert={gezeigt}
        aendern={(neu) => {
          aendern(neu)
          verlassen(neu)
        }}
        angehalten={gesperrt}
        ungueltig={fehlerhaft}
        beschriebenVon={beschriebenVon}
      />
    )
  }

  if (art === 'choice' && form !== undefined && form.wahlen.length > 0) {
    return (
      <select
        value={gezeigt}
        onChange={(ereignis) => {
          aendern(ereignis.target.value)
          verlassen(ereignis.target.value)
        }}
        {...gemeinsameAngaben}
        disabled={gesperrt}
      >
        {!form.wahlen.includes(gezeigt) && (
          <option value={gezeigt}>{gezeigt}</option>
        )}
        {form.wahlen.map((wahl) => (
          <option key={wahl} value={wahl}>
            {wahl}
          </option>
        ))}
      </select>
    )
  }

  // **Mehrzeilig bekommt einen Kasten.** Über `magic.mse-game` steht
  // `multi line: true` **68**-mal — darunter Regeltext und Flavortext.
  if (form?.mehrzeilig === true) {
    return (
      <textarea
        rows={4}
        defaultValue={klartext}
        onChange={(ereignis) => {
          getippt(ereignis, (wert) => {
            aendern(zurueck(wert))
          })
        }}
        onBlur={(ereignis) => {
          verlassen(zurueck(ereignis.target.value))
        }}
        {...tippbar}
      />
    )
  }

  return (
    <input
      type="text"
      defaultValue={klartext}
      onChange={(ereignis) => {
        getippt(ereignis, (wert) => {
          aendern(zurueck(wert))
        })
      }}
      onBlur={(ereignis) => {
        verlassen(zurueck(ereignis.target.value))
      }}
      {...tippbar}
    />
  )
}
