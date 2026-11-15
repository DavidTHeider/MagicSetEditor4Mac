// **Das Feld »Druck« nimmt, was getippt wurde, am nativen Fenster** — #2389.
//
// Gefunden bei der Ausgangsmessung P0 (#2336): getippt `p0-blitz`, im Handel
// `P0-blitz`. **Gemessen wird am selben Feld, auf demselben Weg wie H1**:
// Anna und Bert sind einander über die Kommandozeile bekannt, Anna öffnet
// »Neuer Tausch …«, wählt Bert, legt je Seite eine Zeile an, tippt den Druck
// mit Tasten und bietet an. Was danach im Feld steht und was `mse handel
// liste --json` trägt, sagt, ob macOS etwas berichtigt hat.
//
// **Die Gegenprobe ist dasselbe Drehbuch an einem Bündel ohne die
// Eigenschaften** — sie steht in `docs/FELDPRUEFUNG.md`, »Kennungsfelder«:
// dort trug der Handel `P0-blitz`. **Das Feld zeigte auch dann den getippten
// Wert**; macOS schreibt erst beim Verlassen groß. Deshalb zählt, was die
// Kommandozeile sagt, und nicht, was im Feld steht.
import { feld, gekoppelt, handel, knopf } from '../huellenhandel.mjs'

// **Seit #2378 steht die Kennung im Kasten »Kennung von Hand«** — eine Zeile
// entsteht über den Namen der Karte, und wer den Druck tippt, klappt ihn auf.
const kennungskasten = "[...document.querySelectorAll('button[aria-expanded]')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Kennung von Hand')"

export default async function druckfeld09(p) {
  const bert = new p.constructor('druckfeld09-bert')
  bert.zeilen = p.zeilen

  try {
    const annaOrt = p.ort()
    gekoppelt(p, { Anna: annaOrt, Bert: bert.ort() })

    p.mse(annaOrt, 'exemplar', 'p0-blitz', '1', 'nearMint', 'de', 'nonfoil')

    await p.starte({ ort: annaOrt })
    await p.taste('5', 'befehl')
    await p.wartetext('Neuer Tausch', 20000)
    await p.klickeAuf(knopf('Neuer Tausch …'), '»Neuer Tausch …«')
    const bertWahl = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && /^Bert aufg?e?schlagen$/i.test((b.getAttribute('aria-label') ?? '').trim()))"
    await p.warte('Bert steht zur Wahl', () => p.seite(`return !!${bertWahl}`), 10000)
    await p.klickeAuf(bertWahl, 'Bert in der Schublade')

    // **Wie in P0, Zeichen für Zeichen**: der Druck beginnt mit einem
    // Buchstaben, und die Zeile wird erst mit dem Knopf übernommen — die
    // Großschreibung kann beim Tippen kommen oder beim Verlassen des Felds.
    for (const [seite, druck] of [['Von dir', 'p0-blitz'], ['Vom Gegenüber', 'p0-schock']]) {
      await p.warte(`»Zeile hinzufügen: ${seite}« steht`, () => p.seite(`return !!${knopf(`Zeile hinzufügen: ${seite}`)}`), 10000)
      await p.klickeAuf(knopf(`Zeile hinzufügen: ${seite}`), `»Zeile hinzufügen: ${seite}«`)
      await p.warte('der Kasten »Kennung von Hand« steht', () => p.seite(`return !!${kennungskasten}`), 10000)
      if ((await p.seite(`return ${kennungskasten}.getAttribute('aria-expanded')`)) !== 'true') {
        await p.klickeAuf(kennungskasten, 'der Kasten »Kennung von Hand«')
      }
      await p.warte('das Feld »Druck« steht', () => p.seite(`return !!${feld('Druck')}`), 10000)
      await p.klickeAuf(feld('Druck'), 'das Feld »Druck«')
      await p.seite(`(${feld('Druck')}).focus(); return true`)
      await p.tippe(druck)
      await p.ruhe(1500)
      const wert = await p.seite(`return (${feld('Druck')}).value`)
      p.zeige(`Eigenschaften: ${JSON.stringify(await p.seite(`const f = ${feld('Druck')}; return [f.getAttribute('autocapitalize'), f.getAttribute('autocorrect'), f.getAttribute('spellcheck')]`))}`)
      p.pruefe(`im Feld »Druck« steht, was getippt wurde (${seite})`, wert === druck, `»${wert}«`)
      await p.klickeAuf(knopf('+ Zeile hinzufügen'), '»+ Zeile hinzufügen«')
    }
    await p.klickeAuf(knopf('Eröffnen'), '»Eröffnen«')
    // **Seit P13 ein Zug** (#2381): „Anbieten und zustimmen" — angeboten ist
    // der Handel danach wie vorher, wie in `p10handel09`.
    await p.warte('»Anbieten und zustimmen« steht', () => p.seite(`return !!${knopf('Anbieten und zustimmen')}`), 10000)
    await p.klickeAuf(knopf('Anbieten und zustimmen'), '»Anbieten und zustimmen«')
    await p.warte('angeboten', () => handel(p, annaOrt).zeilen.some((x) => x.vorgang.zustand === 'angeboten'), 10000).catch(() => false)

    // **Was im Handel steht, sagt die Kommandozeile** — nicht die Seite, die
    // es zeigt.
    const roh = JSON.stringify(handel(p, annaOrt))
    p.pruefe('im Handel steht der Druck, wie er getippt wurde', roh.includes('"p0-blitz"') && roh.includes('"p0-schock"'), roh.match(/"[pP]0-[a-z]+"/gu)?.join(', ') ?? 'kein Druck')
    p.pruefe('kein Druck ist großgeschrieben', !/"P0-/u.test(roh))
    await p.bild('druckfeld')
    p.pruefe('⌘Q beendet die Anwendung', await p.beende())
  } finally {
    p.fehler += bert.fehler
    bert.aufraeumen()
  }
}
