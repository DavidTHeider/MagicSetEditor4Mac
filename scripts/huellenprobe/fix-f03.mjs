// **Das CSV der Sammlung zurücklesen, am nativen Fenster** — FIX-F03
// (#1747)
// und die Spezifikation S03.
//
// Der Befund war `received 1 arguments … expected 2` — die Oberfläche rief
// die Bindung anders, als Go sie erklärte, und das zeigte sich nur über die
// echte Naht von Wails. Hier geht jeder Schritt über sie:
//
// - S03.NAT-01: ein künstliches CSV mit neuer Kennung und Menge 3 legt
//   genau drei Stück an,
// - S03.NAT-02 und NAT-03: was die Anwendung ausgibt, liest sie über
//   dieselbe Oberfläche wieder ein — in einem frischen Bestand steht danach
//   dieselbe Sammlung,
// - NAT-02: eine Zeile für einen Druck, der schon da ist, erhöht ihn — die
//   fachliche Antwort auf eine Dublette,
// - S03.NAT-03: eine kaputte Zeile gibt einen fachlichen Bericht, ohne
//   technischen Wortlaut, und keine unerklärte Teiländerung.
import { randomUUID } from 'node:crypto'

const kopf = 'printing,menge,zustand,sprache,finish,foil,name,setcode,nummer,erwerbsdatum,erwerbspreis,erwerbswaehrung,notiz,lagerort,wunsch,wunschrang,wunschart,verweisfassung,kennungssystem,kennung'
const zeile = (kennung, menge) => `${kennung},${String(menge)},nearMint,en,nonfoil,false,,,,,,,,,false,normal,genau,1,scryfall,${kennung}`

// feld ist das Textfeld der Ausfuhr — eingefügt wird wie mit ⌘V, über den
// Wert und ein `input`, damit React es sieht.
const feld = "[...document.querySelectorAll('textarea')].find((t) => t.checkVisibility() && (t.labels?.[0]?.innerText ?? '').includes('Ausgabe'))"

async function einfuegen(p, text) {
  await p.seite(`const t = ${feld}; const setze = Object.getOwnPropertyDescriptor(HTMLTextAreaElement.prototype, 'value').set; setze.call(t, ${JSON.stringify(text)}); t.dispatchEvent(new Event('input', { bubbles: true })); return true`)
}

// knopf drückt einen Knopf, **nativ** — mit der Maus der Leitung.
const knopf = (p, text) => p.klickeText(text)

const berichte = "return [...document.querySelectorAll('[role=status],[role=alert],p')].map((e) => e.innerText.trim()).filter((t) => /aufgenommen|Zeile \\d/.test(t)).join(' | ')"

// zuruecklesen drückt „CSV zurücklesen" und wartet auf einen **neuen**
// Bericht — der alte steht noch da, bis der neue ihn ersetzt.
async function zuruecklesen(p) {
  const vorher = await p.seite(berichte)
  await knopf(p, 'CSV zurücklesen')

  // **Trifft eine Zeile eine vorhandene, fragt die Einfuhr erst** (DOM-03,
  // #1698): „Vor dem Einlesen" mit „Mengen addieren" als Vorgabe. Dann wird
  // bestätigt, wie ein Mensch es täte.
  const vorschau = await p.warte('Bericht oder Vorschau', async () => {
    const jetzt = await p.seite(berichte)
    if (jetzt !== '' && jetzt !== vorher) return 'bericht'
    return (await p.seite("return document.body.innerText.includes('Vor dem Einlesen')")) ? 'vorschau' : false
  })
  if (vorschau === 'vorschau') {
    // **Über die Schublade und das Ende des Texts** — der Knopf trägt sein
    // Zeichen „+" vor dem Wort, und `klickeText` nähme deshalb „Einlesen und
    // ausgeben …" im Kopf der Seite.
    const einlesen = "[...document.querySelectorAll('[role=dialog] button')].find((b) => b.innerText.replace(/\\s+/g, ' ').trim().endsWith('Einlesen'))"
    await p.seite(`(${einlesen})?.scrollIntoView({ block: 'center' }); return true`)
    await p.klickeAuf(einlesen, '»Einlesen« in der Vorschau')
  }

  return p.warte('ein neuer Bericht steht da', async () => {
    const jetzt = await p.seite(berichte)
    return jetzt !== '' && jetzt !== vorher ? jetzt : false
  }).catch(async (grund) => {
    p.zeige(`Meldungen: ${JSON.stringify(await p.seite("return [...document.querySelectorAll('[role=status],[role=alert]')].map((e) => e.innerText.trim()).filter(Boolean).slice(0, 6)"))}`)
    throw grund
  })
}

const csvVon = (p, ort) => p.mse(ort, 'sammlung-ausgeben', 'csv').trim().split('\n')

// **Ausgabe und Rücklesen stehen in einer Schublade** — seit UXR-10 (#1918)
// hinter *Einlesen und ausgeben …* im Kopf der Seite, eingeklappt unter der
// Listeneinfuhr. Vorher in einem Kasten im Navigator, der bei leerer
// Sammlung offen stand und mit dem ersten Exemplar zuklappte.
async function zurSammlung(p) {
  await p.taste('3', 'befehl')
  await p.wartetext('Einlesen und ausgeben …', 20000)
  await p.klickeText('Einlesen und ausgeben …')
  await p.warte('die Schublade steht', () => p.seite("return [...document.querySelectorAll('[role=dialog]')].some((d) => d.checkVisibility() && d.getAttribute('aria-label') === 'Einlesen und ausgeben')"), 20000)
  await p.klickeText('Ausgeben und sichern')
  await p.wartetext('Als CSV', 20000)
}

export default async function fixF03(p) {
  const ort = p.ort()
  p.mse(ort, 'exemplar', '4c1a8f6e-2b3d-4e5f-8a9b-0c1d2e3f4a5b', '2')
  await p.starte({ ort })
  await zurSammlung(p)

  p.notiz('S03.NAT-01 — ein künstliches CSV mit neuer Kennung und Menge 3')
  const neu = randomUUID()
  await einfuegen(p, `${kopf}\n${zeile(neu, 3)}\n`)
  const bericht = await zuruecklesen(p)
  p.zeige(`Bericht: ${bericht}`)
  const nachher = csvVon(p, ort).find((z) => z.startsWith(neu)) ?? ''
  p.pruefe('genau drei Stück des neuen Drucks', nachher.split(',')[1] === '3', nachher.slice(0, 80))

  p.notiz('NAT-02 — eine Dublette erhöht, statt einen zweiten Eintrag anzulegen')
  await einfuegen(p, `${kopf}\n${zeile(neu, 2)}\n`)
  p.zeige(`Bericht: ${await zuruecklesen(p)}`)
  const zeilen = csvVon(p, ort).filter((z) => z.startsWith(neu))
  p.pruefe('ein Eintrag mit fünf Stück', zeilen.length === 1 && zeilen[0].split(',')[1] === '5', zeilen.join(' / ').slice(0, 120))

  p.notiz('S03.NAT-03 — eine kaputte Zeile gibt einen fachlichen Bericht')
  const vorKaputt = csvVon(p, ort).join('\n')
  await einfuegen(p, `${kopf}\n${zeile(randomUUID(), 'drei')}\n`)
  const kaputt = await zuruecklesen(p)
  p.pruefe('der Bericht nennt die Zeile', /Zeile\s*2/.test(kaputt), kaputt.slice(0, 200))
  p.pruefe('ohne technischen Wortlaut', !/strconv|parse|invalid syntax|sql/i.test(kaputt))
  p.pruefe('und ohne unerklärte Teiländerung', csvVon(p, ort).join('\n') === vorKaputt)

  p.notiz('S03.NAT-02 — was die Anwendung ausgibt, liest sie wieder ein')
  await einfuegen(p, '')
  await knopf(p, 'Als CSV')
  const ausgabe = await p.warte('die Ausgabe steht im Feld', async () => {
    const text = await p.seite(`return ${feld}?.value ?? ''`)
    return text.startsWith('printing,') && text.includes(neu) ? text : false
  })
  p.pruefe('⌘Q beendet', await p.beende())

  const frisch = p.ort()
  await p.starte({ ort: frisch })
  await zurSammlung(p)
  await einfuegen(p, ausgabe)
  // **Gewartet wird auf die Sammlung selbst** — bis UXR-10 (#1918)
  // verschwand der Bericht hier mit dem Kasten, der bei leerer Sammlung
  // offen stand; die Schublade steht seither weiter da.
  await knopf(p, 'CSV zurücklesen')
  await p.warte('die Sammlung ist eingelesen', () => csvVon(p, frisch).length === ausgabe.trim().split('\n').length, 20000).catch((grund) => {
    p.zeige(`ausgegeben: ${JSON.stringify(ausgabe.trim().split('\n'))}`)
    p.zeige(`eingelesen: ${JSON.stringify(csvVon(p, frisch))}`)
    throw grund
  })
  const vergleich = (liste) => liste.map((z) => z.split(',').filter((_, i) => i !== 17).join(',')).sort().join('\n')
  p.pruefe('im frischen Bestand steht dieselbe Sammlung', vergleich(csvVon(p, frisch)) === vergleich(ausgabe.trim().split('\n')),
    `${String(csvVon(p, frisch).length - 1)} Zeilen`)
  p.pruefe('⌘Q beendet', await p.beende())
}
