// **Die Sicht eines Gegenübers über die Leitung, am nativen Fenster** — P11
// (#2379),
// Befunde C-15 und Z-02 aus dem Audit zur Produktionsreife 0.9.0.
//
// # Was es prüft
//
// **Zwei getrennte Peers**: Anna im Fenster, Bert als eigener Prozess
// `mse horchen` mit eigenem Bestand, eigener Identität und eigenem Bund. Bert
// besitzt zwei Sturmrufer und sucht einen Glutläufer, Anna hat zwei
// Glutläufer und sucht Sturmrufer; Bert
// gibt Anna Sammlung und Wünsche zum Handeln frei. Anna öffnet einen Tausch
// mit Bert, fragt „Freigegebene Sicht anfragen“ — **ohne Datei und ohne
// Text** — und übernimmt den Vorschlag: zwei Zeilen, keine doppelt. Nach
// Berts Widerruf kommt eine Absage und keine Zeile. Der Weg ohne Leitung
// (Datei oder Text) steht weiter im Kasten „Was ihr tauschen könntet“.
//
// # Was es nicht belegt
//
// **Zwei Rechner** (#1998), **VoiceOver**. Frist, fremde Fassung,
// Doppelzustellung und eine fremde Antwort prüfen die Läufe in
// `internal/service/handelssicht_test.go` über eine echte Schleife; der Weg
// über TCP mit TLS zwischen zwei Prozessen steht in
// `internal/flaechen/handelssicht_test.go`.
import { spawn } from 'node:child_process'
import { join } from 'node:path'

import { kindumgebung } from '../huellenorte.mjs'
import { gekoppelt, handel, knopf } from '../huellenhandel.mjs'

const sturm = '00000000-0000-4000-8000-0000000b1002'
const glut = '00000000-0000-4000-8000-0000000a1001'

const bertWahl = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && /^Bert aufg?e?schlagen$/i.test((b.getAttribute('aria-label') ?? '').trim()))"
const gruppe = "[...document.querySelectorAll('[role=group]')].find((g) => g.checkVisibility() && g.getAttribute('aria-label') === 'Was dein Gegenüber freigibt')"
const gruppentext = `return (${gruppe})?.innerText.replace(/\\s+/g, ' ').trim() ?? ''`
const knopfIn = (name) => `[...((${gruppe})?.querySelectorAll('button') ?? [])].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(name)})`
const spalte = (titel) => `return [...document.querySelectorAll('section')].find((s) => s.checkVisibility() && s.querySelector('p')?.innerText.trim() === ${JSON.stringify(titel)})?.innerText.replace(/\\s+/g, ' ').trim() ?? ''`
const meldung = "return [...document.querySelectorAll('[role=alert]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.trim()).join(' | ')"

// exemplar legt ein Stück mit lesbarem Bezug an — Besitz oder Wunsch.
const exemplar = (p, ort, druck, menge, name, nummer, absicht) =>
  p.mse(ort, 'exemplar', druck, String(menge), 'nearMint', 'de', 'nonfoil', '', '', '', '', name, 'p11', nummer, '', absicht)

// horche startet Bert als eigenen Prozess — wie `net01b`.
function horche(p, ort) {
  const prozess = spawn(join(process.cwd(), '.go/bin/mse'), ['horchen', '127.0.0.1:0'], {
    env: kindumgebung(ort),
    stdio: ['ignore', 'pipe', 'ignore'],
  })
  p.pids.push(prozess.pid)

  return new Promise((erfuellt, verworfen) => {
    let gelesen = ''
    prozess.stdout.on('data', (stueck) => {
      gelesen += stueck.toString()
      if (gelesen.includes('\n')) erfuellt(gelesen.split('\n')[0].trim())
    })
    prozess.on('exit', (code) => verworfen(new Error(`mse horchen endete mit ${String(code)}`)))
  })
}

async function neuerTauschMitBert(p) {
  await p.taste('5', 'befehl')
  await p.wartetext('Neuer Tausch', 20000)
  await p.klickeAuf(knopf('Neuer Tausch …'), '»Neuer Tausch …«')
  await p.warte('Bert steht zur Wahl', () => p.seite(`return !!${bertWahl}`), 10000)
  await p.klickeAuf(bertWahl, 'Bert in der Schublade')
  // **Seit #2494 holt der Markttisch die Sicht selbst**, und der Knopf steht
  // nur noch im Listenweg — dort prüft dieses Drehbuch das ausdrückliche
  // Anfragen, Übernehmen und die Absage; den Tisch hält `p12markttisch09`.
  await p.warte('»Als Liste« steht', () => p.seite(`return !!${knopf('Als Liste')}`), 10000)
  if ((await p.seite(`return ${knopf('Als Liste')}.getAttribute('aria-pressed')`)) !== 'true') {
    await p.klickeAuf(knopf('Als Liste'), '»Als Liste«')
  }
  await p.warte('»Freigegebene Sicht anfragen« steht', () => p.seite(`return !!${knopf('Freigegebene Sicht anfragen')}`), 10000)
}

export default async function p11sicht09(p) {
  const annaOrt = p.ort()
  const bertOrt = p.ort()
  const wer = gekoppelt(p, { Anna: annaOrt, Bert: bertOrt })
  p.mse(annaOrt, 'einstellungen', 'setzen', 'ersteschritte', 'aus')

  exemplar(p, bertOrt, sturm, 2, 'Sturmrufer', '3', 'besitz')
  exemplar(p, bertOrt, glut, 1, 'Glutläufer', '4', 'wunsch')
  // **Zwei, denn der Abgleich gibt nur her, was über eines hinausgeht**
  // (`behalten` 1, wie im Kasten des Abgleichs) — der erste Lauf hatte
  // einen, und „Du gibst“ blieb zu Recht leer.
  exemplar(p, annaOrt, glut, 2, 'Glutläufer', '4', 'besitz')
  exemplar(p, annaOrt, sturm, 2, 'Sturmrufer', '3', 'wunsch')
  for (const umfang of ['sammlung', 'wuensche']) p.mse(bertOrt, 'kontakt-freigeben', wer.Anna.kennung, 'handel', umfang)

  p.notiz('Bert horcht in einem eigenen Prozess')
  const adresse = await horche(p, bertOrt)
  p.mse(annaOrt, 'kontakt-adresse', wer.Bert.kennung, adresse)
  p.zeige(`Bert unter ${adresse}`)

  await p.starte({ ort: annaOrt })
  await p.groesse(1440, 900)

  p.notiz('Anna fragt Berts Sicht an — ohne Datei und ohne Text')
  await neuerTauschMitBert(p)
  await p.klickeAuf(knopf('Freigegebene Sicht anfragen'), '»Freigegebene Sicht anfragen«')
  const gesehen = await p.warte('Berts Sicht steht', async () => { const t = await p.seite(gruppentext); return t.includes('freigegebene Zeilen') ? t : null }, 20000).catch(() => '')
  p.pruefe('zwei freigegebene Zeilen, davon ein Wunsch', gesehen.includes('2 freigegebene Zeilen, davon 1 gesucht.'), gesehen)
  p.pruefe('der Vorschlag nennt beide Richtungen', gesehen.includes('Du gibst') && gesehen.includes('1 × Glutläufer') && gesehen.includes('Du bekommst') && gesehen.includes('Sturmrufer'), gesehen)
  await p.bild('sicht')

  p.notiz('Den Vorschlag übernehmen')
  await p.klickeAuf(knopfIn('Vorschlag übernehmen'), '»Vorschlag übernehmen«')
  const meine = await p.warte('die Zeile von Anna steht', async () => { const t = await p.seite(spalte('Von dir')); return t.includes('Glutläufer') ? t : null }, 10000).catch(() => '')
  const seine = await p.seite(spalte('Vom Gegenüber'))
  p.pruefe('zwei Zeilen mit Namen, in ihren Spalten', meine.includes('1 × Glutläufer') && seine.includes('× Sturmrufer'), `${meine} || ${seine}`)
  p.pruefe('ein zweites Übernehmen geht nicht', await p.seite(`return (${knopfIn('Vorschlag übernehmen')})?.disabled === true`))
  await p.bild('uebernommen')

  await p.klickeAuf(knopf('Eröffnen'), '»Eröffnen«')
  const eroeffnet = await p.warte('der Entwurf steht im Bestand', () => handel(p, annaOrt).zeilen.find((x) => x.vorgang.zustand === 'entwurf')?.vorgang ?? null, 10000).catch(() => null)
  const positionen = eroeffnet === null ? [] : JSON.parse(p.mse(annaOrt, 'handel', 'stand', eroeffnet.vorgang, '--json')).positionen
  p.pruefe('im Entwurf genau diese zwei Zeilen', positionen.length === 2
    && positionen.some((x) => x.von === wer.Anna.kennung && x.printing === glut && x.menge === 1)
    && positionen.some((x) => x.von === wer.Bert.kennung && x.printing === sturm), JSON.stringify(positionen))

  p.notiz('Nach Berts Widerruf: eine Absage und keine Zeile')
  for (const umfang of ['sammlung', 'wuensche']) p.mse(bertOrt, 'kontakt-widerrufen', wer.Anna.kennung, 'handel', umfang)
  await neuerTauschMitBert(p)
  await p.klickeAuf(knopf('Freigegebene Sicht anfragen'), '»Freigegebene Sicht anfragen«')
  const absage = await p.warte('die Absage steht', async () => { const m = await p.seite(meldung); return m.includes('nichts frei') ? m : null }, 20000).catch(() => '')
  p.pruefe('nach dem Widerruf eine Absage', absage.includes('Das Gegenüber gibt dir zum Handeln nichts frei.'), absage)
  p.pruefe('und keine fremde Sicht', !(await p.seite(`return !!${gruppe}`)))
  await p.bild('absage')

  p.notiz('Der Weg ohne Leitung bleibt')
  await p.klickeAuf(knopf('Was ihr tauschen könntet'), '»Was ihr tauschen könntet«')
  p.pruefe('Datei und Text stehen im Kasten', await p.warte('der Kasten ist offen', () => p.seite("return [...document.querySelectorAll('textarea')].some((t) => t.checkVisibility())"), 5000).catch(() => false))
  p.pruefe('⌘Q beendet', await p.beende())
}
