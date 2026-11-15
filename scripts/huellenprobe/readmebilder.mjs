// **Die Bilder des README, am nativen Fenster** — der Abschluss von 0.6.0
// (#1660): die Oberfläche ist seit Phase 6.13 eine andere, und keiner der
// Abzüge aus dem Browsermodus zeigte sie noch.
//
// Ein eigenes Zuhause mit dem ganzen Pack (als APFS-Klon, ohne Platz zu
// kosten), einer Kopie eines kanonischen Sets, einer Sammlung, Decks, einem
// Turnier, einer Partie und einem Tausch — alles über die Kommandozeile, wie
// ein Mensch es vorher angelegt hätte. Dann je Erscheinungsbild jede Fläche
// aus `abzuegeliste.mjs`, bei 1280 × 800, und das Bild unter
// `docs/bilder/<name>.png` beziehungsweise `<name>-dunkel.png`. Zuletzt das
// Band einer neueren Fassung, mit einer Freigabe von der Schleife.
//
// **Was es nicht tut**: urteilen. Ob ein Bild zeigt, was der Text sagt, sieht
// ein Mensch. Und **die Rechte**: Rahmen und Kartenbilder gehören ihren
// Inhabern (`DRITTANBIETER.md`); gezeigt werden sie unter derselben
// Einwilligung wie die Mitlieferung (Vermerk 0025).
import { copyFileSync, existsSync, mkdirSync, readdirSync } from 'node:fs'
import { createServer } from 'node:http'
import { execFileSync } from 'node:child_process'
import { homedir } from 'node:os'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'

import { wege } from '../abzuegeliste.mjs'
import { oeffnen, zurSeite } from './uirevue.mjs'
import { kindumgebung } from '../huellenorte.mjs'

const wurzel = fileURLToPath(new URL('../..', import.meta.url))
// **Die Eingaben werden nur gelesen, und sie sind benannt** — R03 (#2095).
// Pack, Schriften und das kanonische Set liegen außerhalb des Zuhauses der
// Probe; kopiert wird hinein, geschrieben nur dort. Bis zum 2026-10-06 kamen
// beide Pfade unbenannt aus `homedir()`: unter einem umgelenkten `HOME` fehlte
// das Pack, und der Lauf scheiterte mitten in der Einrichtung an `ENOENT`.
const pack = process.env.MSE_PROBE_PACK ?? join(homedir(), 'full-magic-pack')
const kanon = process.env.MSE_PROBE_KANON ?? join(homedir(), 'Library/Application Support/MagicSetEditor4Mac/kanonische-sets')

// fehlendeEingaben nennt jede Eingabe, die nicht dasteht, mit ihrem Schalter.
export const fehlendeEingaben = ({ pack: p, kanon: k }) => [
  [join(p, 'data'), 'MSE_PROBE_PACK'],
  [join(p, 'Magic - Fonts'), 'MSE_PROBE_PACK'],
  [join(k, 'Blue.mse-set'), 'MSE_PROBE_KANON'],
].filter(([pfad]) => !existsSync(pfad)).map(([pfad, schalter]) => `${pfad} (${schalter})`)

// zeile liest einen Wert aus der Ausgabe von `mse` (`schluessel<TAB>wert`).
const zeile = (ausgabe, schluessel) => ausgabe.split('\n').find((z) => z.startsWith(`${schluessel}\t`))?.split('\t')[1] ?? ''

// erstesFeld gibt das erste Tabfeld der ersten Zeile — die Kennung, die `mse`
// nach dem Anlegen nennt.
const erstesFeld = (ausgabe) => ausgabe.trim().split('\n')[0].split('\t')[0]

export default async function readmebilder(p) {
  const fehlt = fehlendeEingaben({ pack, kanon })
  p.pruefe('die Eingaben liegen bereit, bevor etwas angelegt wird', fehlt.length === 0, fehlt.length ? `fehlt: ${fehlt.join(', ')}` : `${pack}, ${kanon}`)
  if (fehlt.length > 0) return
  const ort = p.ort()
  einrichten(p, ort)

  // **Beide Erscheinungsbilder, oder eines** (`MSE_README_BILD`), und **alle
  // Wege, oder die genannten** (`MSE_README_WEGE`, durch Kommas getrennt) —
  // damit ein gescheiterter Weg sich allein nachfahren lässt.
  const bilder = process.env.MSE_README_BILD ? [process.env.MSE_README_BILD] : ['hell', 'dunkel']
  const genannt = (process.env.MSE_README_WEGE ?? '').split(',').filter(Boolean)
  const auswahl = genannt.length === 0 ? wege : wege.filter((weg) => genannt.includes(weg.name))
  for (const bild of bilder) {
    p.notiz(`Erscheinungsbild ${bild}`)
    p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
    await oeffnen(p, ort, { name: bild, breite: 1440, hoehe: 900 })
    for (const weg of auswahl) {
      try {
        await zumWeg(p, weg)
        const pfad = await p.bild(`${weg.name}-${bild}`)
        if (pfad) copyFileSync(pfad, join(wurzel, 'docs/bilder', `${weg.name}${bild === 'dunkel' ? '-dunkel' : ''}.png`))
        p.pruefe(`${weg.name} (${bild})`, Boolean(pfad))
      } catch (grund) {
        p.pruefe(`${weg.name} (${bild})`, false, String(grund.message ?? grund))
        // **Was stattdessen dastand** — sonst sagt ein gescheiterter Klick
        // nur, dass er scheiterte.
        p.zeige(String(await p.seite("return (document.querySelector('main')?.innerText ?? '').replace(/\\s+/g, ' ').slice(0, 500)")))
        // **Und in einer offenen Schublade** — sie steht nicht in `main`.
        p.zeige(String(await p.seite("return [...document.querySelectorAll('[role=dialog]')].filter((d) => d.checkVisibility()).map((d) => d.innerText.replace(/\\s+/g, ' ').slice(0, 500)).join(' | ')")))
        await p.bild(`${weg.name}-${bild}-fehlschlag`)
      }
      await aufraeumen(p)
    }
    p.pruefe(`⌘Q beendet die Anwendung (${bild})`, await p.beende())
  }

  if (genannt.length === 0 || genannt.includes('oberflaeche-neuefassung')) await neueFassung(p, ort, bilder)
}

// neueFassung nimmt das Band einer neueren Fassung auf — B-32 (AP-12,
// #2369).
// Das Bild stand seit dem 2026-08-26 ohne Lauf da und zeigte eine Leiste
// mit elf Einträgen.
//
// **Die Freigabe kommt von der Schleife, nicht von GitHub**: ein Dienst hier
// antwortet auf `GET /repos/…/releases/latest` (`internal/aktualisierung`,
// `NeuesteFreigabe`), und `MSE_FREIGABE_ADRESSE` lenkt die Anwendung dorthin
// (`internal/app/aktualisierungverdrahtung.go`). Der Verweis zeigt auf
// `beispiel.invalid` — geklickt wird er nicht, und hinaus geht nichts.
async function neueFassung(p, ort, bilder) {
  const dienst = createServer((anfrage, antwort) => {
    if (!(anfrage.url ?? '').endsWith('/releases/latest')) {
      antwort.writeHead(404).end()

      return
    }
    antwort.writeHead(200, { 'content-type': 'application/json' })
    antwort.end(JSON.stringify({ tag_name: 'v0.9.0', html_url: 'https://beispiel.invalid/fassung/0.9.0', published_at: '2026-10-20T10:00:00Z', assets: [] }))
  })
  await new Promise((bereit) => { dienst.listen(0, '127.0.0.1', bereit) })
  const adresse = `http://127.0.0.1:${String(dienst.address().port)}`
  p.mse(ort, 'einstellungen', 'setzen', 'aktualisierung', 'melden')
  try {
    for (const bild of bilder) {
      p.notiz(`Das Band einer neueren Fassung (${bild})`)
      p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
      p.mse(ort, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')
      p.mse(ort, 'einstellungen', 'setzen', 'fensterbreite', '1440')
      p.mse(ort, 'einstellungen', 'setzen', 'fensterhoehe', '900')
      await p.starte({ ort, umgebung: { MSE_FREIGABE_ADRESSE: adresse } })
      await p.taste('1', 'befehl')
      await p.wartetext('Fassung 0.9.0 ist da.', 30000)
      // **Erst, wenn die Startseite fertig ist** — der erste Lauf nahm sie
      // auf, als „Was Aufmerksamkeit braucht“ noch „Wird geladen …“ sagte.
      await p.warte('die Startseite ist geladen', () => p.seite("return !document.querySelector('main')?.innerText.includes('Wird geladen')"), 60000)
      await p.ruhe()
      // **Unter „melden“ kein Ladeknopf** — das Bild vom 2026-10-17 zeigte
      // ihn, und ein Druck kam als Fehler zurück (Commit 3371526f5).
      p.pruefe(`unter „melden“ steht kein Ladeknopf (${bild})`, !(await p.seite("return [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Herunterladen')")))
      const pfad = await p.bild(`oberflaeche-neuefassung-${bild}`)
      if (pfad) copyFileSync(pfad, join(wurzel, 'docs/bilder', `oberflaeche-neuefassung${bild === 'dunkel' ? '-dunkel' : ''}.png`))
      p.pruefe(`oberflaeche-neuefassung (${bild})`, Boolean(pfad))
      p.pruefe(`⌘Q beendet die Anwendung (${bild}, Band)`, await p.beende())
    }
  } finally {
    p.mse(ort, 'einstellungen', 'setzen', 'aktualisierung', 'aus')
    dienst.close()
  }
}

// einrichten legt an, was die Bilder zeigen.
function einrichten(p, ort) {
  p.notiz('Der Bestand, über die Kommandozeile')
  const anwendungsort = join(ort.heim, 'Library/Application Support/MagicSetEditor4Mac')
  mkdirSync(join(anwendungsort, 'pakete'), { recursive: true })
  mkdirSync(join(ort.heim, 'Library/Fonts'), { recursive: true })
  // **Als Klon** (`cp -c`): auf APFS kostet er keinen Platz, und das Original
  // bleibt unberührt.
  for (const name of readdirSync(join(pack, 'data'))) {
    execFileSync('cp', ['-cR', join(pack, 'data', name), join(anwendungsort, 'pakete', name)])
  }
  for (const name of readdirSync(join(pack, 'Magic - Fonts')).filter((n) => /\.(ttf|otf)$/i.test(n))) {
    execFileSync('cp', ['-c', join(pack, 'Magic - Fonts', name), join(ort.heim, 'Library/Fonts', name)])
  }
  const kopie = join(ort.ort, 'Blue.mse-set')
  execFileSync('cp', ['-c', join(kanon, 'Blue.mse-set'), kopie])
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  // **Beide Spalten stehen**, wie das README den Arbeitsplatz beschreibt.
  p.mse(ort, 'einstellungen', 'setzen', 'arbeitsplatz.navigator', 'sichtbar')
  p.mse(ort, 'einstellungen', 'setzen', 'arbeitsplatz.inspektor', 'sichtbar')
  p.mse(ort, 'oeffnen', kopie)
  const setID = erstesFeld(p.mse(ort, 'sets'))
  p.zeige(`Set ${setID}`)

  // **Eine Sammlung mit Namen** — ein Druck ohne Namen hieße nach seiner Kennung.
  const exemplare = [
    ['counterspell-7ed', '4', 'nearMint', 'de', 'nonfoil', '2026-08-01', '1.20', 'EUR', 'getauscht', 'Counterspell', '7ed', '67', 'Kiste 1'],
    ['serra-angel-m10', '2', 'excellent', 'en', 'foil', '2026-07-12', '3.50', 'EUR', 'gekauft', 'Serra Angel', 'm10', '39', 'Kiste 1'],
    ['lightning-bolt-m10', '3', 'nearMint', 'de', 'nonfoil', '2026-06-03', '2.10', 'EUR', 'gekauft', 'Lightning Bolt', 'm10', '146', 'Kiste 2'],
    ['llanowar-elves-m19', '4', 'played', 'en', 'nonfoil', '2026-05-20', '0.30', 'EUR', 'Booster', 'Llanowar Elves', 'm19', '314', 'Kiste 2'],
    ['dark-ritual-a25', '2', 'nearMint', 'en', 'etched', '2026-09-01', '0.80', 'EUR', 'getauscht', 'Dark Ritual', 'a25', '88', 'Ordner'],
    ['opt-dom', '6', 'mint', 'de', 'nonfoil', '2026-09-15', '0.10', 'EUR', 'Booster', 'Opt', 'dom', '60', 'Ordner'],
  ]
  for (const zeilen of exemplare) p.mse(ort, 'exemplar', ...zeilen)

  // **Zwei Decks aus dem Set**, je fünfzehn Karten zu vier — sie füllen
  // Decks, Turnier und Tisch. **Nicht über `deck aus-set`**: dessen Menge
  // gilt je Karte, und 40 machten aus den 3.693 Karten von _Blue_ ein Deck
  // mit 147.720 — eine Partie daraus sprengte die Partienliste (#2065).
  const kartenIDs = p.mse(ort, 'karten', setID).trim().split('\n')
    .map((z) => z.split('\t')[0].trim()).filter((k) => /^\d+$/.test(k))
  const deckA = erstesFeld(p.mse(ort, 'deck', 'neu', 'Blaue Flut', 'magic'))
  const deckB = erstesFeld(p.mse(ort, 'deck', 'neu', 'Tiefsee', 'magic'))
  for (const [deck, von] of [[deckA, 0], [deckB, 15]]) {
    for (const karte of kartenIDs.slice(von, von + 15)) p.mse(ort, 'deck', 'legen', deck, `--karte=${karte}`, '4')
  }
  p.zeige(`Decks ${deckA}, ${deckB}`)

  // **Ein Turnier mit einer gespielten Runde.**
  const turnier = erstesFeld(p.mse(ort, 'turnier', 'neu', 'Freitagsrunde', 'magic'))
  for (const name of ['Anna', 'Bert', 'Cleo', 'Dana']) p.mse(ort, 'turnier', 'anmelden', turnier, name)
  p.zeige(p.mse(ort, 'turnier', 'runde', turnier).trim().split('\n').slice(0, 3).join(' · '))

  // **Eine Partie am Tisch.**
  p.mse(ort, 'partie-anlegen', 'magic', `anna:Anna:${deckA}`, `bert:Bert:${deckB}`)

  // **Ein angebotener Tausch** mit einem bestätigten Gegenüber.
  const ich = zeile(p.mse(ort, 'peer-anlegen'), 'kennung')
  const fremd = p.ort()
  const gegenueber = p.mse(fremd, 'peer-anlegen')
  const kennung = zeile(gegenueber, 'kennung')
  p.mse(ort, 'kontakt-aufnehmen', kennung, 'Turfer')
  p.mse(ort, 'kontakt-bestaetigen', kennung, zeile(gegenueber, 'fingerabdruck'))
  p.mse(ort, 'zweck-erlauben', kennung, 'handel')
  const positionen = JSON.stringify([
    { von: ich, printing: 'counterspell-7ed', menge: 1, guete: 'nearMint', sprache: 'de', finish: 'nonfoil' },
    { von: kennung, printing: 'serra-angel-m10', menge: 1, guete: 'nearMint', sprache: 'en', finish: 'nonfoil' },
  ])
  execFileSync(join(wurzel, '.go/bin/mse'), ['handel', 'eroeffnen', 'v-readme', 'tausch', kennung], {
    cwd: wurzel, input: positionen,
    env: kindumgebung(ort),
  })
  p.mse(ort, 'handel', 'anbieten', 'v-readme')
}

// zumWeg führt zu einem Bild: die Seite über die Palette, dann was der Weg
// außerdem verlangt.
async function zumWeg(p, weg) {
  if (weg.palette) {
    await p.taste('k', 'befehl')
    await p.warte('die Palette steht', () => p.seite("return !!document.querySelector('[role=dialog] input')"))
    if (weg.tippen) await p.tippeInsFeld(weg.tippen)
    await new Promise((los) => setTimeout(los, 600))

    return
  }
  await zurSeite(p, weg.seite)
  if (weg.lasche) await p.klickeAuf(`document.querySelector('[role=tab][data-lasche="${weg.lasche}"]')`, `die Lasche ${weg.lasche}`)
  if (weg.klick) await p.klickeAuf(weg.klick, weg.klickName ?? weg.klick)
  for (const text of weg.texte ?? []) await p.klickeText(text)
  for (const text of weg.vielleicht ?? []) {
    const da = () => p.seite(`return [...document.querySelectorAll('button')].some((e) => e.checkVisibility() && e.innerText.trim() === ${JSON.stringify(text)})`)
    let gesehen = false
    for (let versuch = 0; versuch < 10 && !gesehen; versuch++) {
      gesehen = await da()
      if (!gesehen) await new Promise((los) => setTimeout(los, 300))
    }
    if (gesehen) await p.klickeText(text)
  }
  if (weg.fertig) {
    const beginn = Date.now()
    await p.warte(`${weg.name} ist fertig`, () => p.seite(`return ${weg.fertig}`), weg.fertigFrist ?? 20000)
    p.zeige(`${weg.name}: fertig nach ${String(Date.now() - beginn)} ms`)
  }
  // **Oben anfangen**: ein Klick tief in der Seite hat sie gerollt — und
  // nicht nur sie: der Klick auf eine Lasche rollt die Laschenspalte, bis
  // die Lasche mittig steht, und „Einstellung suchen“ stand angeschnitten da.
  await p.seite("for (const e of [document.scrollingElement, ...document.querySelectorAll('main, main *')]) if (e && e.scrollTop > 0) e.scrollTo(0, 0); return true")
  // **Oder bei dem Abschnitt, den das Bild zeigen soll** — er steht sonst
  // am unteren Rand, angeschnitten.
  if (weg.zeige) {
    await p.seite(`const e = [...document.querySelectorAll('button, h2, h3')].find((e) => e.checkVisibility() && e.innerText.trim() === ${JSON.stringify(weg.zeige)}); e?.scrollIntoView({ block: 'start' }); return !!e`)
  }
  await new Promise((los) => setTimeout(los, weg.ruhe ?? 800))
}

// aufraeumen schließt, was ein Weg offen gelassen hat — Palette, Schublade.
async function aufraeumen(p) {
  for (let i = 0; i < 2; i++) {
    await p.taste('Escape')
    await new Promise((los) => setTimeout(los, 150))
  }
}

