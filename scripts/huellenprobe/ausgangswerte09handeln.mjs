// **Ausgangswerte Handeln am nativen Fenster** — P0
// (#2336),
// Befund C-22 aus dem Audit zur Produktionsreife 0.9.0.
//
// Die Wege H1 bis H6 aus Audit C (Abschnitt 3.2), gezählt nach
// `huellenwege.mjs` und abgelegt als `docs/belege/ausgangswerte09-handeln.json`
// (Vertrag: `scripts/ausgangswege.mjs`).
//
// # Der Bestand
//
// **Selbst geschrieben und künstlich**: Anna und Bert als zwei Bestände auf
// diesem Mac, einander bekannt über die Kommandozeile. Die Drucke heißen
// `09-blitz` und `09-schock` und gibt es nirgends; Preise kommen von einer
// Gegenstelle in diesem Prozess (`MSE_SCRYFALL_ADRESSE`), nie aus dem Netz.
//
// **Mit einer Ziffer vorn**: das Feld »Druck« schreibt einen Buchstaben am
// Anfang groß (Befund dieses Laufs, `p0-blitz` wurde `P0-blitz`), und der
// Handel kennte danach sein eigenes Exemplar nicht. Gemessen wird der Weg,
// nicht die Berichtigung. **Behoben mit #2389** (`druckfeld09`); die Ziffer
// bleibt, damit eine Wiederholung dieselben Eingaben hat.
//
// # Was gezählt wird
//
// Wie in `ausgangswerte09spielen.mjs`, und dazu **jedes Übertragen**: ohne
// »In den Ausgang legen« und »Jetzt übertragen« erfährt das Gegenüber nichts.
// H5 zählt je Seite vom Angebot bis »beidseitig verbucht«; beim Gegenüber
// steckt H4 darin. Die Maschinenzeit von H5 enthält das Warten auf die
// andere Seite.
//
// # Was es nicht belegt
//
// **Menschenzeit, Fehlgriffe, Lesen und Suchen** — die Probe kennt den Weg.
// Zwei Geräte in zwei Netzen (#1998): beide Fenster laufen auf diesem Mac.
import { writeFileSync } from 'node:fs'
import { createServer } from 'node:http'
import { join, resolve } from 'node:path'

import { bestaetigeGegenseite, bindung, bisZu, einfuegen, feld, gehe, gekoppelt, halte, handel, knopf, preisstelle, uebertrage, vorgangsstand, waehle } from '../huellenhandel.mjs'
import { tastenweg } from '../huellenspieltisch.mjs'
import { wegzaehler } from '../huellenwege.mjs'

const wurzel = resolve(import.meta.dirname, '..', '..')
const ziel = join(wurzel, 'docs/belege/ausgangswerte09-handeln.json')
const meldungen = "return [...document.querySelectorAll('[role=alert]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.trim()).filter(Boolean)"

export default async function ausgangswerte09handeln(p) {
  const bert = new p.constructor('ausgangswerte09handeln-bert')
  bert.zeilen = p.zeilen
  const { stelle, adresse } = await preisstelle(createServer)

  try {
    await zweiFenster(p, bert, adresse)
  } finally {
    stelle.close()
    p.fehler += bert.fehler
    bert.aufraeumen()
  }
}

async function zweiFenster(p, bert, adresse) {
  const annaOrt = p.ort()
  const bertOrt = bert.ort()
  const wer = gekoppelt(p, { Anna: annaOrt, Bert: bertOrt })
  p.mse(annaOrt, 'exemplar', '09-blitz', '3', 'nearMint', 'de', 'nonfoil')
  p.mse(bertOrt, 'exemplar', '09-schock', '1', 'nearMint', 'de', 'nonfoil')

  // **»Bert aufgeschlagen« statt »Bert aufschlagen«**, solange ein Handel mit
  // Bert offen steht — ein Mensch drückt, was dasteht. Seit #2391 heißt der
  // Knopf immer »Bert aufschlagen«; der Ausdruck nimmt beide, damit die
  // Messung am alten Bündel wiederholbar bleibt.
  const bertWahl = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && /^Bert aufg?e?schlagen$/i.test((b.getAttribute('aria-label') ?? '').trim()))"
  const datensatz = { bauabdruck: p.lauf?.bauabdruck ?? '', erhoben: new Date().toISOString(), fenster: '1440 × 900', wege: [] }
  const z = wegzaehler(p, datensatz)
  const zb = wegzaehler(bert, datensatz)

  await p.starte({ ort: annaOrt, umgebung: { MSE_SCRYFALL_ADRESSE: adresse } })
  await p.groesse(1440, 900)

  p.notiz('H1 — einen Tausch mit je einer Karte pro Seite anbieten, mit der Maus')
  await z.beginne('H1', 'bereich', 'maus')
  await z.taste('5', 'befehl')
  await p.wartetext('Neuer Tausch', 20000)
  await z.klickAuf(knopf('Neuer Tausch …'), '»Neuer Tausch …«')
  await p.warte('Bert steht zur Wahl', () => p.seite(`return !!${bertWahl}`), 10000)
  await z.klickAuf(bertWahl, 'Bert in der Schublade')
  for (const [seite, druck] of [['Von dir', '09-blitz'], ['Vom Gegenüber', '09-schock']]) {
    await p.warte(`»Zeile hinzufügen: ${seite}« steht`, () => p.seite(`return !!${knopf(`Zeile hinzufügen: ${seite}`)}`), 10000)
    await z.klickAuf(knopf(`Zeile hinzufügen: ${seite}`), `»Zeile hinzufügen: ${seite}«`)
    await p.warte('das Feld »Druck« steht', () => p.seite(`return !!${feld('Druck')}`), 10000)
    await z.klickAuf(feld('Druck'), 'das Feld »Druck«')
    await p.seite(`(${feld('Druck')}).focus(); return true`)
    await z.tippe(druck)
    await z.klickAuf(knopf('+ Zeile hinzufügen'), '»+ Zeile hinzufügen«')
  }
  await z.klickAuf(knopf('Eröffnen'), '»Eröffnen«')
  await p.warte('»Anbieten« steht', () => p.seite(`return !!${knopf('Anbieten')}`), 10000)
  await z.klickAuf(knopf('Anbieten'), '»Anbieten«')
  await p.warte('angeboten', () => handel(p, annaOrt).zeilen.some((x) => x.vorgang.zustand === 'angeboten'), 10000).catch(() => false)
  const vorgang = handel(p, annaOrt).zeilen.find((x) => x.vorgang.zustand === 'angeboten')?.vorgang.vorgang ?? ''
  await z.ende({ wirkung: vorgang !== '' })
  await p.bild('h1-angeboten')

  p.notiz('H6 — die Wertbilanz, wenn noch kein Preis vorliegt')
  const bilanz = async () => p.seite("return document.querySelector('[data-handelsassistent]')?.innerText ?? ''")
  p.zeige(`vorher: ${(await bilanz()).includes('noch keine Preise vor') ? 'der Hinweis auf die Sammlung steht' : 'kein Hinweis'}`)
  await z.beginne('H6', 'handel', 'maus')
  await z.taste('3', 'befehl')
  await p.warte('»Preise für alle holen« steht', () => p.seite(`return !!${knopf('Preise für alle holen')}`), 20000)
  await z.klickAuf(knopf('Preise für alle holen'), '»Preise für alle holen«')
  await p.wartetext('Alle Preise sind da.', 20000).catch(() => false)
  await z.taste('5', 'befehl')
  await waehle({ name: 'Anna', p, z, gegenueber: 'Bert' })
  await p.warte('die Bilanz nennt Preise', async () => (await bilanz()).includes('Preise vom'), 10000).catch(() => false)
  await z.ende({ wirkung: (await bilanz()).includes('Preise vom') })
  await p.bild('h6-bilanz')

  p.notiz('H4 und H5 — vom Angebot bis »beidseitig verbucht«, in zwei Fenstern')
  await bert.starte({ ort: bertOrt })
  await bert.groesse(1440, 900)
  for (const [hier, dort, kennung] of [[p, bert, wer.Anna.kennung], [bert, p, wer.Bert.kennung]]) {
    const horch = await bindung(hier, "b.HorchStarten('127.0.0.1:0')")
    const port = horch?.stand?.adresse?.split(':').pop()
    hier.pruefe('das Fenster horcht', Boolean(horch?.stand?.laeuft && port), horch?.stand?.adresse ?? JSON.stringify(horch?.fehler))
    await bindung(dort, `b.PeerkontaktAdresse(${JSON.stringify(kennung)}, '127.0.0.1:${port}')`)
  }
  const parteien = [
    { name: 'Anna', p, z, ort: annaOrt, gegenueber: 'Bert' },
    { name: 'Bert', p: bert, z: zb, ort: bertOrt, gegenueber: 'Anna' },
  ]
  await z.beginne('H5', 'handel', 'maus')
  let h4 = null
  let runden = 0
  let still = 0
  while (runden < 24 && still < 3) {
    runden += 1
    let bewegt = false
    for (const partei of parteien) {
      const st = vorgangsstand(partei.p, partei.ort, vorgang)
      if (st.lage !== 'dran' || !st.naechster) continue
      bewegt = true
      p.notiz(`${partei.name}: ${st.naechster} (Stand ${st.zustand})`)
      await partei.p.aktiviere()
      // **H4 ist Berts erster Schritt**: ⌘5, den Handel aufschlagen,
      // zustimmen — danach beginnt seine Zeile für H5.
      if (partei.name === 'Bert' && h4 === null) {
        await zb.beginne('H4', 'bereich', 'maus')
        await zb.taste('5', 'befehl')
        await waehle(partei)
        await gehe(partei, st.naechster)
        await partei.p.warte('Bert hat zugestimmt', () => vorgangsstand(bert, bertOrt, vorgang).zustand !== st.zustand, 15000).catch(() => false)
        h4 = await zb.ende({ wirkung: vorgangsstand(bert, bertOrt, vorgang).zustand !== st.zustand })
        await zb.beginne('H5', 'handel-gegenueber', 'maus')
      } else {
        await waehle(partei)
        // **Festgehalten wird vor »Vorbereiten«** — gemessen am 2026-10-13:
        // ohne Exemplar weist es ab (»festgehalten sind 0 von 1 Stück«).
        // Beide Seiten geben hier eine Karte, also halten beide fest.
        if (['vorbereiten', 'uebergeben'].includes(st.naechster) && !partei.gehalten) {
          await halte(partei, vorgang)
          partei.gehalten = true
        }
        await gehe(partei, st.naechster)
        const gestellt = await bisZu(15000, () => {
          const neu = vorgangsstand(partei.p, partei.ort, vorgang)

          return neu.zustand !== st.zustand || neu.naechster !== st.naechster || neu.lage !== st.lage
        })
        if (!gestellt) {
          partei.p.pruefe(`${partei.name} hat »${st.naechster}« gestellt`, false, JSON.stringify(await partei.p.seite(meldungen)))
          partei.p.zeige(`Assistent: ${JSON.stringify(((await partei.p.seite("return document.querySelector('[data-handelsassistent]')?.innerText ?? ''")) ?? '').slice(0, 1500))}`)
          await partei.p.bild(`${partei.name.toLowerCase()}-${st.naechster}-klemmt`)
          runden = 99
          break
        }
      }
      await uebertrage(partei)
    }
    if (!bewegt) {
      // **Wer die Abschlussstimme der anderen Seite sieht, sagt es**
      // (T10, #2026) — erst danach geht der Beleg hinüber.
      for (const partei of parteien) {
        if (vorgangsstand(partei.p, partei.ort, vorgang).zustand !== 'abgleichOffen') continue
        bewegt = true
        await partei.p.aktiviere()
        await waehle(partei)
        await bestaetigeGegenseite(partei)
        await bisZu(15000, () => vorgangsstand(partei.p, partei.ort, vorgang).zustand !== 'abgleichOffen')
        await uebertrage(partei)
      }
    }
    still = bewegt ? 0 : still + 1
    if (!bewegt) await new Promise((los) => setTimeout(los, 3000))
  }
  const fertig = (partei) => vorgangsstand(partei.p, partei.ort, vorgang).zustand === 'beidseitigVerbucht'
  p.zeige(`Stand: Anna ${vorgangsstand(p, annaOrt, vorgang).zustand} · Bert ${vorgangsstand(bert, bertOrt, vorgang).zustand}`)
  await z.ende({ wirkung: parteien.every(fertig) })
  const gegenueber = await zb.ende({ wirkung: parteien.every(fertig) })
  if (h4) {
    for (const spalte of ['handlungen', 'sichtwechsel', 'tabstellen', 'maschinenzeitMs']) gegenueber[spalte] += h4[spalte]
  }
  await p.bild('h5-anna')
  await bert.bild('h5-bert')
  bert.pruefe('⌘Q beendet Bert', await bert.beende())
  await p.aktiviere()

  p.notiz('H2 — ein Exemplar aus der Sammlung in den Tausch geben')
  const angezeigt = async () => (await p.seite("return document.querySelector('[data-handelsassistent]')?.innerText ?? ''")) ?? ''
  const zeigeKnoepfe = async (wo) => p.zeige(`${wo}: ${JSON.stringify(await p.seite("return [...document.querySelectorAll('main button, [role=dialog] button')].filter((b) => b.checkVisibility()).map((b) => (b.getAttribute('aria-label') || b.innerText).trim().replace(/\\s+/g, ' ')).slice(0, 40)"))}`)
  await z.beginne('H2', 'sammlung', 'maus')
  await z.taste('3', 'befehl')
  // **Die Zeile und nicht ihr Kästchen** — das Kästchen hakt für die
  // Massenhandlungen an, ein Klick auf die Zeile hält das Exemplar im
  // Inspector, und dort stehen die Anschlussknöpfe.
  const reihe = "[...document.querySelectorAll('[role=row], [role=option]')].find((e) => e.checkVisibility() && e.innerText.includes('09-blitz') && !e.querySelector('[role=row], [role=option]'))"
  await p.warte('das Exemplar steht', () => p.seite(`return !!${reihe}`), 20000)
  await z.klickAuf(reihe, 'die Zeile »09-blitz«')
  await p.warte('»Zum Tausch« steht', () => p.seite(`return !!${knopf('Zum Tausch')}`), 10000).catch(() => false)
  await z.klickAuf(knopf('Zum Tausch'), '»Zum Tausch«')
  await zeigeKnoepfe('nach »Zum Tausch«')
  // **Kommt »Neuer Tausch …« nicht, scheitert der Weg** — hier stand ein
  // bedingtes ⌘5, und eine Messung mit ihm sah aus wie ein längerer Weg
  // (Befund 3 des Nachlaufs über Phase 0).
  await p.warte('»Neuer Tausch …« steht', () => p.seite(`return !!${knopf('Neuer Tausch …')}`), 10000)
  await z.klickAuf(knopf('Neuer Tausch …'), '»Neuer Tausch …«')
  await p.warte('Bert steht zur Wahl', () => p.seite(`return !!${bertWahl}`), 10000)
  p.zeige(`Bert steht als »${await p.seite(`return (${bertWahl}).getAttribute('aria-label')`)}«`)
  await z.klickAuf(bertWahl, 'Bert in der Schublade')
  await zeigeKnoepfe('nach Bert')
  await p.warte('das vorbelegte Formular steht', () => p.seite(`return !!${knopf('+ Zeile hinzufügen')}`), 10000).catch(() => false)
  await z.klickAuf(knopf('+ Zeile hinzufügen'), '»+ Zeile hinzufügen«')
  await z.ende({ wirkung: (await angezeigt()).includes('09-blitz') })
  await p.bild('h2-vorbelegt')

  p.notiz('H3 — sehen, was die Gegenseite hat und was zusammenpasst')
  const stand = JSON.stringify({
    angebot: [{ referenz: 'b1', printing: '09-stein', menge: 1, zustand: 'nearMint', sprache: 'de', finish: 'nonfoil', bezeichnung: 'Probestein' }],
    wuensche: [{ referenz: 'w1', printing: '09-blitz', menge: 1, zustand: 'nearMint', sprache: 'de', finish: 'nonfoil', bezeichnung: 'Probeblitz' }],
  })
  const abgleich = "[...document.querySelectorAll('section, details, div')].find((e) => e.checkVisibility() && e.querySelector('h3, h4, summary')?.innerText.trim() === 'Was ihr tauschen könntet')"
  await z.beginne('H3', 'bereich', 'maus')
  await z.klickAuf(knopf('Was ihr tauschen könntet'), '»Was ihr tauschen könntet«')
  const standfeld = "[...document.querySelectorAll('textarea')].find((e) => e.checkVisibility() && !e.readOnly)"
  await p.warte('das Feld für den fremden Stand steht', () => p.seite(`return !!${standfeld}`), 10000)
  await z.wahl(einfuegen(standfeld, stand))
  await z.klickAuf(knopf('Abgleichen'), '»Abgleichen«')
  await p.warte('der Abgleich nennt die Karte', async () => ((await p.seite(`return (${abgleich})?.innerText ?? document.body.innerText`)) ?? '').includes('Probeblitz'), 10000).catch(() => false)
  p.zeige(`Abgleich: ${JSON.stringify(((await p.seite(`return (${abgleich})?.innerText ?? ''`)) ?? '').slice(0, 600))}`)
  await z.ende({ wirkung: ((await p.seite('return document.body.innerText')) ?? '').includes('Probeblitz') })
  await p.bild('h3-abgleich')



  // **H1 mit der Tastatur zuletzt**: davor ginge H2 in den angebotenen
  // Handel statt ins vorbelegte Formular — »Zum Tausch« landet in dem
  // Handel, der mit Bert gerade offen steht (Befund dieses Laufs, behoben mit
  // #2391 — `tauschneu09`).
  p.notiz('H1 — einen Tausch anbieten, mit der Tastatur')
  const offenVorher = handel(p, annaOrt).zeilen.filter((x) => x.vorgang.zustand === 'angeboten').length
  await tastenweg(p, z, 'H1', [
    { taste: ['5', 'befehl'] },
    ['»Neuer Tausch …«', knopf('Neuer Tausch …')], ['Bert in der Schublade', bertWahl],
    ['»Zeile hinzufügen: Von dir«', knopf('Zeile hinzufügen: Von dir')], ['das Feld »Druck«', feld('Druck'), '09-blitz'], ['»+ Zeile hinzufügen«', knopf('+ Zeile hinzufügen')],
    ['»Zeile hinzufügen: Vom Gegenüber«', knopf('Zeile hinzufügen: Vom Gegenüber')], ['das Feld »Druck«', feld('Druck'), '09-schock'], ['»+ Zeile hinzufügen«', knopf('+ Zeile hinzufügen')],
    ['»Eröffnen«', knopf('Eröffnen')], ['»Anbieten«', knopf('Anbieten')],
  ], () => handel(p, annaOrt).zeilen.filter((x) => x.vorgang.zustand === 'angeboten').length === offenVorher + 1, 'bereich')
  await p.bild('h1-tastatur')

  writeFileSync(ziel, `${JSON.stringify(datensatz, null, 2)}\n`)
  p.zeige(`Datensatz: ${ziel}`)
  p.pruefe('⌘Q beendet', await p.beende())
}
