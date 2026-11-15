// **Bildmaße zwischen den Programmen vergleichen** — Z2 des Zweitblicks über
// Phase 6.11.
//
// `docs/LEISTUNGSVERGLEICH.md` §2 nennt die Auflösung gleich, „wenn je Karte
// Breite und Höhe gleich sind; das prüft der Lauf". Er speicherte beides und
// verglich nichts: ein Bild mit anderer Größe galt als intakt, solange es ein
// PNG war. Hier steht, **wie gepaart wird** und **was als Abweichung zählt**.
//
// # Gepaart wird nach der Karte
//
//   - **Im Batch** (`bilder`) schreibt MSE2 nach `{card.name}.png`, MSE4Mac
//     nach `<Stelle>.png`. Der Dateiname von MSE2 wird aus dem Kartennamen
//     abgeleitet, **wie MSE2 ihn bildet** — sonst ließe sich keine Datei der
//     einen einer der anderen zuordnen.
//   - **In den Sitzungen** (`navigation`) tragen die Schritte Satz und Stelle.
//
// **Was nicht gepaart werden kann, ist benannt** und kein stilles Überspringen:
// ein Bild, das auf einer Seite fehlt, steht als `ungepaart` da.

/**
 * untag nimmt die Auszeichnung aus einem Text: alles zwischen `<` und `>`
 * (`src/util/tagged_string.cpp`, `untag`, Zeilen 33–47, Fassung f9d9356 vom
 * 2021-07-06). Die Maskierung einzelner Zeichen (`untag_char`) betrifft
 * Namen nicht, die hier vorkommen, und ist nicht nachgebaut.
 */
function untag(text) {
  let ausgabe = ''
  let imTag = false
  for (const zeichen of text) {
    if (zeichen === '<') imTag = true
    else if (!imTag) ausgabe += zeichen
    else if (zeichen === '>') imTag = false
  }
  return ausgabe
}

const dateizeichen = /[\p{L}\p{N} _.-]/u

/**
 * bereinigt macht aus einer Zeichenkette einen Dateinamen
 * (`src/util/file_utils.cpp`, `clean_filename`, Zeilen 54–72, Fassung f9d9356
 * vom 2021-07-06): erlaubt sind Buchstaben, Ziffern, Leerzeichen, `_`, `-` und
 * `.`; vorn fallen Leerzeichen und Punkte, hinten ebenso; was leer bleibt,
 * heißt `no-name`.
 *
 * **Abweichung, und sie ist beabsichtigt**: MSE2 fragt `iswalnum`, und was das
 * für Zeichen außerhalb von ASCII ergibt, hängt vom Gebietsschema des
 * Prozesses ab. Hier gelten Unicode-Buchstaben und -Ziffern. Wo beides
 * auseinanderläuft, findet sich eine Datei nicht — und das steht als
 * `ungepaart` im Beleg und nicht als Maßabweichung.
 */
function bereinigt(text) {
  let sauber = ''
  let anfang = true
  for (const zeichen of text) {
    const kappen = anfang && (zeichen === ' ' || zeichen === '.')
    if (dateizeichen.test(zeichen) && !kappen) {
      anfang = false
      sauber += zeichen
    }
  }
  while (sauber.endsWith(' ') || sauber.endsWith('.')) sauber = sauber.slice(0, -1)
  return sauber === '' ? 'no-name' : sauber
}

/**
 * mse2Dateiname gibt den Dateinamen, unter dem `--export-images` mit der
 * Vorlage `{card.name}.png` eine Karte ablegt
 * (`src/data/format/image.cpp`, `export_images`, Zeilen 69–95).
 */
export function mse2Dateiname(kartenname) {
  return bereinigt(untag(`${kartenname}.png`))
}

/**
 * mse2Dateinamen gibt die Dateinamen für Karten in der Reihenfolge des Satzes.
 * **Gleiche Namen werden nummeriert**: das zweite `Name.1.png`, das dritte
 * `Name.2.png` (`resolve_filename_conflicts`, `CONFLICT_NUMBER_OVERWRITE`,
 * `src/util/file_utils.cpp`, Zeilen 88–95, Fassung f9d9356 vom 2021-07-06 —
 * so ruft `main.cpp` Zeile 254 `export_images` auf).
 */
export function mse2Dateinamen(kartennamen) {
  const benutzt = new Set()
  return kartennamen.map((name) => {
    const stamm = mse2Dateiname(name)
    let datei = stamm
    for (let i = 1; benutzt.has(datei); i++) datei = `${stamm.replace(/\.png$/, '')}.${i}.png`
    benutzt.add(datei)
    return datei
  })
}

/**
 * vergleicheBilderpaare hält Breite und Höhe je Karte gegen die Referenz.
 *
 * `referenz` und `andere` bilden Dateinamen auf `{ breite, hoehe }` ab
 * (`bilderImVerzeichnis`); `namen` sind die Kartennamen in Satzreihenfolge,
 * `andererName` gibt den Dateinamen der anderen Seite je Stelle.
 *
 * **Was auf einer Seite fehlt, steht getrennt nach der Seite**: fehlt ein Bild
 * bei den anderen, steht das schon in der Zahl ihrer Bilder; fehlt es bei
 * MSE2, kann sich die Ableitung des Dateinamens geirrt haben, und das ist ein
 * Befund über den Vergleich.
 */
export function vergleicheBilderpaare({ namen, referenz, andere, andererName }) {
  const dateien = mse2Dateinamen(namen)
  const ergebnis = { gepaart: 0, abweichungen: [], ohneReferenz: [], ohneAndere: [] }
  namen.forEach((name, stelle) => {
    const r = referenz[dateien[stelle]]
    const a = andere[andererName(stelle)]
    const wer = `Karte ${stelle} (${name})`
    if (!r) return ergebnis.ohneReferenz.push(`${wer}: kein Bild bei MSE2 (${dateien[stelle]})`)
    if (!a) return ergebnis.ohneAndere.push(`${wer}: kein Bild bei den anderen (${andererName(stelle)})`)
    ergebnis.gepaart++
    if (r.breite !== a.breite || r.hoehe !== a.hoehe) {
      ergebnis.abweichungen.push(`${wer}: ${a.breite}x${a.hoehe}, MSE2 ${r.breite}x${r.hoehe}`)
    }
  })
  return ergebnis
}

/**
 * schrittmasseVergleichen hält die Maße der Bildschritte aus den Sitzungen
 * gegen MSE2 — je Satz und Stelle, über alle Zustände und Wiederholungen.
 * Gibt je abweichendem Programm und je Karte eine Abweichung mit Text.
 *
 * **Ein Schritt ohne Maße zählt nicht**: ein Bild, das nicht entstand oder
 * nicht heil ist, steht schon als Fehler in der Zahl der Schritte. Eine
 * Abweichung braucht zwei Seiten, die ein Maß haben.
 */
export function schrittmasseVergleichen(schritte, referenz = 'mse2') {
  const je = new Map()
  for (const s of schritte) {
    if (s.breite === null || s.breite === undefined || s.satz === undefined) continue
    const karte = je.get(`${s.satz}\u0000${s.stelle}`) ?? { satz: s.satz, stelle: s.stelle, programme: new Map() }
    const masse = karte.programme.get(s.programm) ?? new Set()
    masse.add(`${s.breite}x${s.hoehe}`)
    karte.programme.set(s.programm, masse)
    je.set(`${s.satz}\u0000${s.stelle}`, karte)
  }
  const abweichungen = []
  for (const { satz, stelle, programme } of je.values()) {
    const sollwerte = programme.get(referenz)
    for (const [programm, masse] of programme) {
      const andersAlsSelbst = masse.size > 1
      const andersAlsReferenz = programm !== referenz && sollwerte !== undefined && (masse.size !== sollwerte.size || [...masse].some((m) => !sollwerte.has(m)))
      if (!andersAlsSelbst && !andersAlsReferenz) continue
      const soll = sollwerte ? `, MSE2 ${[...sollwerte].join(' und ')}` : ''
      abweichungen.push({ programm, satz, stelle, text: `${satz}[${stelle}]: ${programm} ${[...masse].join(' und ')}${soll}` })
    }
  }
  return abweichungen
}
