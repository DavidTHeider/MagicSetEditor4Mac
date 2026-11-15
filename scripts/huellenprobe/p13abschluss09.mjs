// **Ein Tausch bis zum Abschluss, ein Zug je Entscheidung** — P13
// (#2381),
// Befund C-16, Vermerk 0200.
//
// # Was es prüft
//
// **Zwei Fenster**: Anna und Bert als zwei Programme auf diesem Mac, je mit
// eigenem Bestand, beide horchen und kennen die Adresse des anderen. Anna
// hat den Glutläufer, Bert den Sturmrufer und gibt Anna seine Sammlung frei.
//
// - **H1** am Markttisch, mit der Maus — „Anbieten und zustimmen" bringt das
//   Angebot samt Annas Zustimmung hinüber (Schwelle 5).
// - **H4**: Bert nimmt an — ⌘5, und der eine Handel, an dem er dran ist,
//   steht offen (Schwelle 2; Ausgang 3).
// - **H5** je Seite bis „beidseitig verbucht": Anna „Vorbereiten und
//   übergeben", „Verbuchen", „Gegenseite bestätigt"; Bert „Übergabe
//   bestätigen", „Verbuchen", „Gegenseite bestätigt" (Schwelle 3 je Seite;
//   Ausgang 35 und 30). **Annas Stimme zur Übergabe zählt bei Bert nicht** —
//   sie reist mit dem Schritt nach „vorbereitet" (`NimmtStand`); die Übergabe
//   schließt bei Anna, wenn Berts Stimme kommt, und ihr „Verbuchen" bringt
//   sie hinüber. Das fand der erste Lauf dieser Probe.
// - **Ein Teilfehler am echten Fenster**: Anna horcht nicht mehr, als Bert
//   die Gegenseite bestätigt — bei ihm ist es eingetragen, das Fenster sagt,
//   dass Anna es noch nicht hat und gerade nicht zu erreichen ist (#2491),
//   und „Jetzt übertragen" bringt es hinaus, sobald sie wieder horcht.
// - **Am Ende beider Seiten dasselbe**: der Bestand getauscht, ein Beleg mit
//   beiden Unterschriften, im Journal je zwei Bewegungen und keine doppelt.
//
// # Was es nicht belegt
//
// **Zwei Geräte in zwei Netzen** (#1998), **VoiceOver**, und dass ein Mensch
// „Gegenseite bestätigt" erst drückt, wenn er es vom Gegenüber erfahren hat
// — hier drückt die Probe, sobald drüben verbucht ist.
import { writeFileSync } from 'node:fs'
import { join, resolve } from 'node:path'

import { bindung, bisZu, gekoppelt, handel, knopf, vorgangsstand } from '../huellenhandel.mjs'
import { ziehe } from '../huellenrufbuch.mjs'
import { mitteVon } from '../huellenspieltisch.mjs'
import { wegzaehler } from '../huellenwege.mjs'

const wurzel = resolve(import.meta.dirname, '..', '..')
const ziel = join(wurzel, 'docs/belege/p13abschluss09-wege.json')
const glut = '00000000-0000-4000-8000-0000000a1301'
const sturm = '00000000-0000-4000-8000-0000000b1303'

const region = (name) => `[...document.querySelectorAll('section')].find((s) => s.checkVisibility() && (s.getAttribute('aria-label') === ${JSON.stringify(name)} || s.querySelector('p')?.innerText.trim() === ${JSON.stringify(name)}))`
const fachkarte = (fach, name) => `[...((${region(fach)})?.querySelectorAll('button[aria-pressed]') ?? [])].find((b) => b.checkVisibility() && b.getAttribute('aria-label') === ${JSON.stringify(name)})`
const spaltentext = (titel) => `return (${region(titel)})?.innerText.replace(/\\s+/g, ' ').trim() ?? ''`
const meldungen = (p) => p.seite("return [...document.querySelectorAll('[role=alert]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.trim()).filter(Boolean)")
const steht = (text) => `return [...document.querySelectorAll('[role=status]')].some((e) => e.checkVisibility() && e.innerText.includes(${JSON.stringify(text)}))`

const exemplar = (p, ort, druck, name, nummer) =>
  p.mse(ort, 'exemplar', druck, '1', 'nearMint', 'de', 'nonfoil', '', '', '', '', name, 'p13', nummer, '', 'besitz')
const stueck = (p, ort, druck) => JSON.parse(p.mse(ort, '--json', 'besitz', druck)).stueck ?? 0

async function horche(hier, dort, kennung) {
  const horch = await bindung(hier, "b.HorchStarten('127.0.0.1:0')")
  const port = horch?.stand?.adresse?.split(':').pop()
  hier.pruefe('das Fenster horcht', Boolean(horch?.stand?.laeuft && port), horch?.stand?.adresse ?? JSON.stringify(horch?.fehler))
  await bindung(dort, `b.PeerkontaktAdresse(${JSON.stringify(kennung)}, '127.0.0.1:${port}')`)
}

// ziehen zieht eine Karte über AppKit auf eine Spalte — ein Zug ist eine
// Handlung (wie `p12markttisch09`).
async function ziehen(p, z, von, nach) {
  const a = await mitteVon(p, von, false)
  const b = await mitteVon(p, nach, false)
  if (!a || !b) return
  await ziehe(p, [a, { x: (a.x + b.x) / 2, y: (a.y + b.y) / 2 }, b])
  await z.wahl('return true')
}

// drueckt drückt einen Knopf, sobald er dasteht, gezählt.
async function drueckt(p, z, text) {
  await p.warte(`»${text}« steht`, () => p.seite(`return !!${knopf(text)}`), 20000)
  await z.klickAuf(knopf(text), `»${text}«`)
}

export default async function p13abschluss09(p) {
  const bert = new p.constructor('p13abschluss09-bert')
  bert.zeilen = p.zeilen

  try {
    await zweiFenster(p, bert)
  } finally {
    p.fehler += bert.fehler
    bert.aufraeumen()
  }
}

async function zweiFenster(p, bert) {
  const annaOrt = p.ort()
  const bertOrt = bert.ort()
  const wer = gekoppelt(p, { Anna: annaOrt, Bert: bertOrt })
  for (const ort of [annaOrt, bertOrt]) p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  exemplar(p, annaOrt, glut, 'Glutläufer', '1')
  exemplar(bert, bertOrt, sturm, 'Sturmrufer', '3')
  bert.mse(bertOrt, 'kontakt-freigeben', wer.Anna.kennung, 'handel', 'sammlung')

  const datensatz = { bauabdruck: p.lauf?.bauabdruck ?? '', erhoben: new Date().toISOString(), fenster: '1440 × 900', wege: [] }
  const z = wegzaehler(p, datensatz)
  const zb = wegzaehler(bert, datensatz)

  await p.starte({ ort: annaOrt })
  await p.groesse(1440, 900)
  await bert.starte({ ort: bertOrt })
  await bert.groesse(1440, 900)
  await horche(p, bert, wer.Anna.kennung)
  await horche(bert, p, wer.Bert.kennung)

  p.notiz('H1 — am Tisch anbieten; das Angebot trägt Annas Zustimmung und geht hinüber')
  await p.aktiviere()
  await z.beginne('H1', 'bereich', 'maus')
  await z.taste('5', 'befehl')
  await drueckt(p, z, 'Neuer Tausch mit Bert')
  await p.warte('beide Fächer stehen', async () => Boolean(await p.seite(`return !!${fachkarte('Mein Fach', 'Glutläufer · P13 #1')} && !!${fachkarte('Fach von Bert', 'Sturmrufer · P13 #3')}`)), 20000)
  await ziehen(p, z, fachkarte('Mein Fach', 'Glutläufer · P13 #1'), region('Von dir'))
  await p.warte('Glutläufer liegt', async () => (await p.seite(spaltentext('Von dir'))).includes('Glutläufer'), 10000).catch(() => false)
  await ziehen(p, z, fachkarte('Fach von Bert', 'Sturmrufer · P13 #3'), region('Vom Gegenüber'))
  await p.warte('Sturmrufer liegt', async () => (await p.seite(spaltentext('Vom Gegenüber'))).includes('Sturmrufer'), 10000).catch(() => false)
  await z.klickAuf(knopf('Anbieten und zustimmen'), '»Anbieten und zustimmen«')
  const vorgang = await p.warte('bei Bert angekommen', () => handel(bert, bertOrt).zeilen.find((x) => x.vorgang.zustand === 'angeboten')?.vorgang.vorgang ?? null, 20000).catch(() => null)
  const h1 = await z.ende({ wirkung: vorgang !== null })
  p.zeige(`H1: ${String(h1.handlungen)} Handlungen (P12 5)`)
  p.pruefe('H1 ≤ 5', h1.handlungen <= 5, String(h1.handlungen))
  p.pruefe('drüben trägt das Angebot Annas Zustimmung', (vorgangsstand(bert, bertOrt, vorgang).zustimmungen ?? []).includes(wer.Anna.kennung), JSON.stringify(vorgangsstand(bert, bertOrt, vorgang)))
  p.pruefe('Anna sieht, dass es hinausging', await p.warte('»An Bert hinausgegangen.«', () => p.seite(steht('An Bert hinausgegangen.')), 10000).catch(() => false))
  await p.bild('anna-angeboten')

  p.notiz('H4 — Bert nimmt an')
  await bert.aktiviere()
  await zb.beginne('H4', 'bereich', 'maus')
  await zb.taste('5', 'befehl')
  await drueckt(bert, zb, 'Zustimmen')
  const angenommen = await bert.warte('bei Anna angenommen', () => vorgangsstand(p, annaOrt, vorgang).zustand === 'angenommen', 20000).catch(() => false)
  const h4 = await zb.ende({ wirkung: angenommen })
  bert.zeige(`H4: ${String(h4.handlungen)} Handlungen (Ausgang 3)`)
  bert.pruefe('H4 ≤ 2', h4.handlungen <= 2, String(h4.handlungen))

  p.notiz('H5 — bis beidseitig verbucht, je Seite')
  await z.beginne('H5', 'handel', 'maus')
  await zb.beginne('H5', 'handel-gegenueber', 'maus')
  await p.aktiviere()
  await drueckt(p, z, 'Vorbereiten und übergeben')
  await p.warte('bei Bert vorbereitet', () => vorgangsstand(bert, bertOrt, vorgang).zustand === 'vorbereitet', 20000)
  await bert.aktiviere()
  await drueckt(bert, zb, 'Übergabe bestätigen')
  await p.warte('bei Anna ist die Übergabe offen', () => vorgangsstand(p, annaOrt, vorgang).zustand === 'uebergabeOffen', 20000)
  await p.aktiviere()
  await drueckt(p, z, 'Verbuchen')
  await p.warte('bei Anna gebucht und bestätigt', () => vorgangsstand(p, annaOrt, vorgang).zustand === 'abgleichOffen', 20000)
  await bert.warte('bei Bert ist die Übergabe offen', () => vorgangsstand(bert, bertOrt, vorgang).zustand === 'uebergabeOffen', 20000)
  await bert.aktiviere()
  await drueckt(bert, zb, 'Verbuchen')
  await bert.warte('bei Bert gebucht und bestätigt', () => vorgangsstand(bert, bertOrt, vorgang).zustand === 'abgleichOffen', 20000)
  await p.aktiviere()
  await drueckt(p, z, 'Gegenseite bestätigt')
  const annaFertig = await p.warte('Anna ist beidseitig verbucht', () => vorgangsstand(p, annaOrt, vorgang).zustand === 'beidseitigVerbucht', 20000).catch(() => false)
  const h5a = await z.ende({ wirkung: annaFertig })

  p.notiz('Ein Teilfehler: Anna horcht nicht, als Bert die Gegenseite bestätigt')
  await bindung(p, 'b.HorchBeenden()')
  await bert.aktiviere()
  await drueckt(bert, zb, 'Gegenseite bestätigt')
  const bertFertig = await bert.warte('Bert ist beidseitig verbucht', () => vorgangsstand(bert, bertOrt, vorgang).zustand === 'beidseitigVerbucht', 20000).catch(() => false)
  const h5b = await zb.ende({ wirkung: bertFertig })
  p.zeige(`H5: Anna ${String(h5a.handlungen)}, Bert ${String(h5b.handlungen)} (Ausgang 35 und 30)`)
  p.pruefe('H5 Anna ≤ 3', h5a.handlungen <= 3, String(h5a.handlungen))
  bert.pruefe('H5 Bert ≤ 3', h5b.handlungen <= 3, String(h5b.handlungen))

  const unterwegs = await bisZu(15000, async () => (await meldungen(bert)).some((m) => m.startsWith('Bei dir ist es eingetragen, Anna hat es noch nicht')))
  bert.pruefe('bei Bert eingetragen, und das Fenster sagt, dass Anna es noch nicht hat', unterwegs, JSON.stringify(await meldungen(bert)))
  // **Und warum** (#2491): bis dahin stand hier der Satz zu „nicht
  // gefunden" — „lade die Liste neu", obwohl nichts gelöscht war.
  const warum = (await meldungen(bert)).find((m) => m.startsWith('Bei dir ist es eingetragen')) ?? ''
  bert.pruefe('und dass Anna gerade nicht zu erreichen ist, ohne „lade die Liste neu"', warum.includes('Das Gegenüber ist gerade nicht zu erreichen') && !warum.includes('lade die Liste neu'), warum)
  await bert.bild('bert-unterwegs')
  await horche(p, bert, wer.Anna.kennung)
  await bert.aktiviere()
  await bert.klickeAuf(knopf('Jetzt übertragen'), '»Jetzt übertragen«')
  bert.pruefe('und „Jetzt übertragen" bringt es hinaus', await bert.warte('»An Anna hinausgegangen.«', () => bert.seite(steht('An Anna hinausgegangen.')), 15000).catch(() => false))

  p.notiz('Am Ende beider Seiten dasselbe')
  for (const [partei, ort, name] of [[p, annaOrt, 'Anna'], [bert, bertOrt, 'Bert']]) {
    const beleg = await bisZu(15000, async () => (await bindung(partei, `b.Handelsbeleg(${JSON.stringify(vorgang)})`))?.beleg?.beidseitig === true)
    partei.pruefe(`${name}: der Beleg trägt beide Unterschriften`, beleg, JSON.stringify((await bindung(partei, `b.Handelsbeleg(${JSON.stringify(vorgang)})`))?.beleg))
    const journal = (await bindung(partei, `b.Handelsjournal(${JSON.stringify(vorgang)})`))?.zeilen ?? []
    partei.pruefe(`${name}: im Journal zwei Bewegungen, keine doppelt`, journal.length === 2 && new Set(journal.map((zeile) => zeile.art)).size === 2, JSON.stringify(journal.map((zeile) => `${zeile.art} ${zeile.printing}`)))
    await partei.aktiviere()
    await partei.bild(`${name.toLowerCase()}-abschluss`)
  }
  p.pruefe('Anna hat den Sturmrufer und nicht mehr den Glutläufer', stueck(p, annaOrt, sturm) === 1 && stueck(p, annaOrt, glut) === 0)
  bert.pruefe('Bert hat den Glutläufer und nicht mehr den Sturmrufer', stueck(bert, bertOrt, glut) === 1 && stueck(bert, bertOrt, sturm) === 0)
  p.pruefe('⌘Q beendet Anna', await p.beende())
  bert.pruefe('⌘Q beendet Bert', await bert.beende())

  writeFileSync(ziel, `${JSON.stringify(datensatz, null, 2)}\n`)
}
