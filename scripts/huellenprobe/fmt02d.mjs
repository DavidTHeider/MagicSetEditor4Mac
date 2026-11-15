// **Include und Exportvorlage aus einem Gerüst — am nativen Fenster** —
// FMT-02.AK-01 (#1691):
//
// - das Anlegeformular der Paketseite nennt **vier** Arten, jede bei ihrem
//   Namen — vorher hieß alles außer einem Spiel „Stil",
// - eine Exportvorlage auf dem mitgelieferten Spiel `probe` entsteht und
//   liegt mit `game: probe` auf der Platte; ein Include ebenso,
// - die Werkstatt führt die Exportvorlage und öffnet ihre Hauptdatei,
// - **der Kurzname bleibt, wie er getippt ist**: im ersten Lauf hieß das
//   Paket `probe-Namen` — die Rechtschreibkorrektur von macOS hatte „namen“
//   beim Verlassen des Felds großgeschrieben. Getippt wird deshalb nativ und
//   ohne Nachhilfe am Feld.
//
// **Ausgeführt** wird die Vorlage im Flächenlauf
// (`internal/flaechen/paketgeruest_test.go`): die Kommandozeile dieser Probe
// kennt die Saat des Bündels nicht.
import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'

const bereich = "document.querySelector('section[aria-label=\"Neues Paket anlegen\"]')"
const knopf = (name) => `[...(${bereich}).querySelectorAll('button')].find((b) => b.innerText.trim().endsWith(${JSON.stringify(name)}))`

// waehle setzt eine Auswahl, wie ein Mensch sie trifft — React hört auf
// `change`.
async function waehle(p, stelle, wert) {
  await p.seite(`const s = (${bereich}).querySelectorAll('select')[${stelle}]; const setze = Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set; setze.call(s, ${JSON.stringify(wert)}); s.dispatchEvent(new Event('change', { bubbles: true })); return true`)
}

// waehleSpiel wählt das Spiel in der Auswahl mit Suche (UXR-F1, #2007) —
// vorher war es ein zweites `select`. Wie ein Mensch: das Feld anklicken,
// tippen, die Zeile anklicken. Die Zeilen stehen erst, wenn es offen ist.
async function waehleSpiel(p, spiel) {
  const feld = `[...(${bereich}).querySelectorAll('input[role=combobox]')].find((i) => [...i.labels].some((l) => l.innerText.startsWith('Spiel des neuen Pakets')))`
  await p.klickeAuf(feld, 'die Spielwahl')
  await new Promise((los) => setTimeout(los, 500))
  p.zeige(`nach dem Klick: ${JSON.stringify(await p.seite(`const f = ${feld}; return { offen: f.getAttribute('aria-expanded'), wert: f.value, platzhalter: f.placeholder, fokus: document.activeElement === f }`))}`)
  await p.tippe(spiel)
  const zeile = `[...document.querySelectorAll('[role=listbox] [role=option]')].find((o) => o.checkVisibility() && o.innerText.trim() === ${JSON.stringify(spiel)})`
  // **Was dasteht, wenn die Zeile fehlt** — sonst sagt ein roter Lauf nur
  // „nicht da".
  if (!(await p.seite(`return !!${zeile}`))) {
    await new Promise((los) => setTimeout(los, 1000))
    p.zeige(`im Feld: »${await p.seite(`return (${feld}).value`)}«, Zeilen: ${JSON.stringify(await p.seite("return [...document.querySelectorAll('[role=option]')].map((o) => o.innerText.trim() + (o.checkVisibility() ? '' : ' (unsichtbar)')).slice(0, 8)"))}`)
  }
  await p.klickeAuf(zeile, `das Spiel »${spiel}«`)
  await p.warte('das Spiel steht im Feld', () => p.seite(`return (${feld}).value === ${JSON.stringify(spiel)}`))
}

async function zurPalette(p, ziel) {
  await p.taste('k', 'befehl')
  await p.warte('die Palette steht', () => p.seite("return !!document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]')"))
  await p.seite("document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]').focus(); return true")
  await p.tippeInsFeld(ziel)
  await p.klickeAuf(`[...document.querySelectorAll('[data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(ziel)})`, `»${ziel}« in der Palette`)
}

async function lege(p, art, name, spiel) {
  await waehle(p, 0, art)
  await p.seite(`const f = [...(${bereich}).querySelectorAll('input')].find((i) => [...i.labels].some((l) => l.innerText.startsWith('Kurzname'))); f.focus(); return true`)
  await p.tippeInsFeld(name)
  if (spiel) await waehleSpiel(p, spiel)
  await p.klickeAuf(knopf('Anlegen'), `Anlegen (${art})`)
}

export default async function fmt02d(p) {
  const ort = p.ort({ zuhause: true })

  await p.starte({ ort })
  await p.groesse(1280, 820)

  p.notiz('Das Anlegeformular nennt vier Arten')
  await zurPalette(p, 'Pakete')
  // **Seit UXR-17 eingeklappt** (#1925): der Kasten unter dem Arbeitsplatz.
  // **Die Falte muss aufgehen, nicht nur geklickt sein** — R04 (#2094).
  await p.klappeAuf("[...document.querySelectorAll('main button')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Neues Paket anlegen')", 'den Kasten »Neues Paket anlegen«')
  await p.warte('das Formular steht', () => p.seite(`return !!(${bereich})`), 20000)
  const arten = await p.seite(`return [...(${bereich}).querySelectorAll('select')[0].options].map((o) => o.textContent)`)
  p.pruefe('jede Art bei ihrem Namen', JSON.stringify(arten) === JSON.stringify([
    'Spiel (.mse-game)', 'Stil (.mse-style)', 'Include (.mse-include)', 'Exportvorlage (.mse-export-template)',
  ]), JSON.stringify(arten))
  await p.seite(`(${bereich}).scrollIntoView(); return true`)
  await p.bild('arten')

  p.notiz('Eine Exportvorlage und ein Include anlegen')
  await lege(p, 'mse-export-template', 'namen', 'probe')
  await p.wartetext('Angelegt: probe-namen', 15000)
  await lege(p, 'mse-include', 'bausteine')
  await p.wartetext('Angelegt: bausteine', 15000)
  const pakete = JSON.parse(p.mse(ort, 'pakete', '--json')).pakete
  const vorlage = pakete.find((eintrag) => eintrag.name === 'probe-namen')
  const baustein = pakete.find((eintrag) => eintrag.name === 'bausteine')
  p.pruefe('der Kurzname blieb klein', pakete.every((eintrag) => eintrag.name !== 'probe-Namen' && eintrag.name !== 'Bausteine'), pakete.map((eintrag) => eintrag.name).join(', '))
  p.pruefe('die Exportvorlage liegt auf der Platte', vorlage !== undefined && existsSync(join(vorlage.pfad, 'export-template')))
  p.pruefe('sie trägt ihr Spiel', vorlage !== undefined && readFileSync(join(vorlage.pfad, 'export-template'), 'utf8').includes('game: probe'))
  p.pruefe('das Include liegt auf der Platte', baustein !== undefined && existsSync(join(baustein.pfad, 'include')))
  await p.bild('angelegt')

  p.notiz('Die Werkstatt führt die Exportvorlage')
  await zurPalette(p, 'Vorlagen')
  await p.wartetext('probe-namen.mse-export-template', 20000)
  await p.klickeAuf("[...document.querySelectorAll('nav button')].find((b) => b.checkVisibility() && b.innerText.trim() === 'export-template')", 'export-template')
  await p.warte('der Editor zeigt das Gerüst', () => p.seite("return document.querySelector('textarea')?.value.includes('script:') ?? false"), 15000)
  await p.wartetext('Kein Befund in dieser Datei.', 15000)
  p.pruefe('das Gerüst hat keinen Befund', true)
  await p.bild('werkstatt')

  await p.beende()
}
