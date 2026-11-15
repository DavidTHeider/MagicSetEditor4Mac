// **»Zum Tausch« führt in einen neuen Entwurf, auch wenn mit demselben
// Gegenüber ein Handel offen ist** — H2 in zwei Lagen, #2391.
//
// Gefunden bei der Ausgangsmessung P0 (#2336): war der zuletzt
// aufgeschlagene Handel mit Bert **angeboten**, endete Sammlung → Exemplar →
// »Zum Tausch« → »Neuer Tausch …« → Bert in diesem Handel, und das
// vorbelegte Formular erschien nicht. **Festgelegt ist seither: jede Wahl in
// der Schublade beginnt einen neuen Entwurf** (`docs/PROTOTYP-0.9.0.md`,
// Abweichungen). Der angebotene Handel bleibt dabei, wie er ist.
//
// # Die zwei Lagen
//
// 1. **Angeboten** — Anna bietet Bert am Fenster einen Tausch an, und er
//    bleibt aufgeschlagen.
// 2. **Beendet** — derselbe Handel, über `mse handel abbrechen` beendet.
//
// **Die Gegenprobe ist dasselbe Drehbuch an einem Bündel ohne die
// Korrektur** in `Handeln.tsx` und `Handelsschublade.tsx`.
import { feld, gekoppelt, handel, knopf } from '../huellenhandel.mjs'

// **Seit #2378 steht die Kennung im Kasten »Kennung von Hand«** — eine Zeile
// entsteht über den Namen der Karte, und wer den Druck tippt, klappt ihn auf.
const kennungskasten = "[...document.querySelectorAll('button[aria-expanded]')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Kennung von Hand')"

export default async function tauschneu09(p) {
  const bert = new p.constructor('tauschneu09-bert')
  bert.zeilen = p.zeilen

  try {
    const annaOrt = p.ort()
    gekoppelt(p, { Anna: annaOrt, Bert: bert.ort() })
    p.mse(annaOrt, 'exemplar', '09-blitz', '3', 'nearMint', 'de', 'nonfoil')

    const bertWahl = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && /^Bert aufg?e?schlagen$/i.test((b.getAttribute('aria-label') ?? '').trim()))"
    const assistent = async () => (await p.seite("return document.querySelector('[data-handelsassistent]')?.innerText ?? ''")) ?? ''

    await p.starte({ ort: annaOrt })
    await p.taste('5', 'befehl')
    await p.wartetext('Neuer Tausch', 20000)
    await p.klickeAuf(knopf('Neuer Tausch …'), '»Neuer Tausch …«')
    await p.warte('Bert steht zur Wahl', () => p.seite(`return !!${bertWahl}`), 10000)
    await p.klickeAuf(bertWahl, 'Bert in der Schublade')
    await p.warte('»Zeile hinzufügen: Vom Gegenüber« steht', () => p.seite(`return !!${knopf('Zeile hinzufügen: Vom Gegenüber')}`), 10000)
    await p.klickeAuf(knopf('Zeile hinzufügen: Vom Gegenüber'), '»Zeile hinzufügen: Vom Gegenüber«')
    await p.warte('der Kasten »Kennung von Hand« steht', () => p.seite(`return !!${kennungskasten}`), 10000)
    if ((await p.seite(`return ${kennungskasten}.getAttribute('aria-expanded')`)) !== 'true') {
      await p.klickeAuf(kennungskasten, 'der Kasten »Kennung von Hand«')
    }
    await p.warte('das Feld »Druck« steht', () => p.seite(`return !!${feld('Druck')}`), 10000)
    await p.klickeAuf(feld('Druck'), 'das Feld »Druck«')
    await p.seite(`(${feld('Druck')}).focus(); return true`)
    await p.tippeInsFeld('09-schock')
    await p.klickeAuf(knopf('+ Zeile hinzufügen'), '»+ Zeile hinzufügen«')
    await p.klickeAuf(knopf('Eröffnen'), '»Eröffnen«')
    // **Seit P13 ein Zug** (#2381): „Anbieten und zustimmen" — angeboten ist
    // der Handel danach wie vorher, wie in `p10handel09`.
    await p.warte('»Anbieten und zustimmen« steht', () => p.seite(`return !!${knopf('Anbieten und zustimmen')}`), 10000)
    await p.klickeAuf(knopf('Anbieten und zustimmen'), '»Anbieten und zustimmen«')
    await p.warte('angeboten', () => handel(p, annaOrt).zeilen.some((x) => x.vorgang.zustand === 'angeboten'), 10000)
    const vorgang = handel(p, annaOrt).zeilen.find((x) => x.vorgang.zustand === 'angeboten')?.vorgang.vorgang ?? ''
    p.pruefe('ein Handel mit Bert ist angeboten und aufgeschlagen', vorgang !== '' && (await assistent()).includes('09-schock'))

    const zumTausch = async (lage) => {
      p.notiz(`H2 — ${lage}`)
      await p.taste('3', 'befehl')
      const reihe = "[...document.querySelectorAll('[role=row], [role=option]')].find((e) => e.checkVisibility() && e.innerText.includes('09-blitz') && !e.querySelector('[role=row], [role=option]'))"
      await p.warte('das Exemplar steht', () => p.seite(`return !!${reihe}`), 20000)
      await p.klickeAuf(reihe, 'die Zeile »09-blitz«')
      await p.warte('»Zum Tausch« steht', () => p.seite(`return !!${knopf('Zum Tausch')}`), 10000)
      await p.klickeAuf(knopf('Zum Tausch'), '»Zum Tausch«')
      await p.warte('»Neuer Tausch …« steht', () => p.seite(`return !!${knopf('Neuer Tausch …')}`), 10000)
      await p.klickeAuf(knopf('Neuer Tausch …'), '»Neuer Tausch …«')
      await p.warte('Bert steht zur Wahl', () => p.seite(`return !!${bertWahl}`), 10000)
      const beschriftung = await p.seite(`return (${bertWahl}).getAttribute('aria-label')`)
      p.pruefe(`${lage}: der Knopf sagt, was er tut`, beschriftung === 'Bert aufschlagen', `»${beschriftung}«`)
      await p.klickeAuf(bertWahl, 'Bert in der Schublade')
      const vorbelegt = await p.warte('das vorbelegte Formular steht', () => p.seite(`return !!${knopf('+ Zeile hinzufügen')}`), 10000).then(() => true, () => false)
      p.pruefe(`${lage}: das vorbelegte Formular steht`, vorbelegt)
      if (vorbelegt) await p.klickeAuf(knopf('+ Zeile hinzufügen'), '»+ Zeile hinzufügen«')
      await p.ruhe()
      const text = await assistent()
      p.pruefe(`${lage}: ein neuer Entwurf trägt das Exemplar`, text.includes('09-blitz') && !text.includes('09-schock') && Boolean(await p.seite(`return !!${knopf('Eröffnen')}`)), text.replace(/\s+/gu, ' ').slice(0, 120))
      await p.bild(`h2-${lage}`)
    }

    await zumTausch('angeboten')
    p.pruefe('der angebotene Handel blieb, wie er war', handel(p, annaOrt).zeilen.find((x) => x.vorgang.vorgang === vorgang)?.vorgang.zustand === 'angeboten')

    p.mse(annaOrt, 'handel', 'abbrechen', vorgang)
    p.pruefe('der Handel ist beendet', handel(p, annaOrt).zeilen.find((x) => x.vorgang.vorgang === vorgang)?.vorgang.zustand === 'abgebrochen')
    await zumTausch('beendet')

    p.pruefe('⌘Q beendet die Anwendung', await p.beende())
  } finally {
    p.fehler += bert.fehler
    bert.aufraeumen()
  }
}
