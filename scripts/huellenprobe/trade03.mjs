// **Art, Vereinbarung und Ausweg am nativen Fenster** — TRADE-03.AK-02 und
// AK-04 (#1717):
//
// - ein Tausch, dem beide zugestimmt haben, steht als **Vereinbarung** da —
//   ohne „Zugestimmt haben: noch niemand" daneben —, und er lässt sich
//   abbrechen,
// - eine Deckleihe sagt, dass sie keinen Bestand bewegt,
// - und der Verlauf in *Verbuchen* nennt jeden Wechsel des Tauschs (AK-03).
//
// **Die Positionen kommen von der Standardeingabe**, wie in `trade01`.
import { execFileSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'
import { kindumgebung } from '../huellenorte.mjs'

const wurzel = fileURLToPath(new URL('../..', import.meta.url))

// eroeffne legt einen Handel einer Art an und gibt seine Kennung.
function eroeffne(ort, art, gegenueber) {
  const positionen = JSON.stringify([
    { von: gegenueber, printing: 'bolt-m10', menge: 1, guete: 'nearMint', sprache: 'de', finish: 'nonfoil' },
  ])

  return execFileSync(`${wurzel}.go/bin/mse`, ['handel', 'eroeffnen', '', art, gegenueber], {
    cwd: wurzel,
    input: positionen,
    env: kindumgebung(ort),
  }).toString().split('\t')[0]
}

const raum = "[...document.querySelectorAll('section')].find((s) => s.querySelector('h3')?.innerText.includes('Ein Handel von der Eröffnung'))"
// **Alle Status des Raums** — seit TRADE-02 trägt auch die Lagezeile
// („Du bist dran: …") `role=status`, und sie steht vor dem Stand.
const status = `[...(${raum})?.querySelectorAll('[role=status]') ?? []]`
const standtext = `return ${status}.map((e) => e.innerText).join('\\n')`
const standknoten = `(${status}.find((e) => e.innerText.includes('Art:')) ?? ${status}.at(-1))`
const angebot = "document.querySelector('[data-handelsassistent] nav li:nth-child(2) button')"
const verbuchen = "document.querySelector('[data-handelsassistent] nav li:nth-child(4) button')"
const verlaufKnopf = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Verlauf')"
const verlauf = "return [...(document.querySelector('ol[aria-label=Verlauf]')?.querySelectorAll('li') ?? [])].map((l) => l.innerText.trim())"
const zeile = (partner) => `[...document.querySelectorAll('li button')].find((b) => b.innerText.startsWith('${partner}'))`

// waehle klickt die Zeile eines Gegenübers in „Was gerade läuft" und wartet,
// bis sie gewählt ist — **auf die Wirkung, nicht auf den Klick** (§3): ein
// Klick während des Nachladens der Liste geht sonst ins Leere.
async function waehle(p, partner, beschreibung) {
  await p.klickeAuf(zeile(partner), beschreibung)
  await p.warte(`${partner} ist gewählt`, () => p.seite(`return (${zeile(partner)})?.getAttribute('aria-current') === 'true'`), 20000)
}

export default async function trade03(p) {
  const ort = p.ort()

  p.notiz('1 — ein Tausch mit Einigung und eine Deckleihe')
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'peer-anlegen')
  const tausch = eroeffne(ort, 'tausch', 'beinigung')
  p.mse(ort, 'handel', 'anbieten', tausch)
  p.mse(ort, 'handel', 'zustimmen', tausch)
  p.mse(ort, 'handel', 'zustimmen', tausch, 'beinigung')
  const leihe = eroeffne(ort, 'deckleihe', 'bleihe')
  p.zeige(`Tausch ${tausch}, Deckleihe ${leihe}`)

  await p.starte({ ort })
  await p.taste('5', 'befehl')

  p.notiz('2 — der Tausch ist eine Vereinbarung')
  await waehle(p, 'beinigung', 'die Zeile des Tauschs')
  // **Seit TRADE-02 steht der Stand in „Was tauscht ihr"** (#1716).
  await p.klickeAuf(angebot, '»Was tauscht ihr«')
  await p.warte('der Stand steht', async () => (await p.seite(standtext)).includes('Angenommen'), 20000)
  const stand = await p.seite(standtext)
  p.zeige(stand.replace(/\s+/g, ' '))
  p.pruefe('Art Tausch', stand.includes('Art: Tausch'), stand)
  p.pruefe('eine Vereinbarung', stand.includes('Vereinbarung'), stand)
  p.pruefe('keine Stimmen ohne Beschluss', !stand.includes('Zugestimmt haben'), stand)
  p.pruefe('der Satz steht in einer Zeile', /eine Vereinbarung:/.test(stand.replace(/\s+/g, ' ')), stand)
  p.pruefe('abbrechen ist angeboten', await p.seite("return [...document.querySelectorAll('[role=group] button')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Abbrechen')"))
  await p.seite(`(${standknoten})?.scrollIntoView({ block: 'center' })`)
  await p.bild('vereinbarung')

  p.notiz('3 — die Deckleihe bewegt keinen Bestand')
  await waehle(p, 'bleihe', 'die Zeile der Deckleihe')
  await p.klickeAuf(angebot, '»Was tauscht ihr«')
  await p.warte('der Stand der Leihe steht', async () => (await p.seite(standtext)).includes('Deckleihe'), 20000)
  const leihstand = await p.seite(standtext)
  p.pruefe('sie sagt es', leihstand.includes('bewegt keinen Bestand'), leihstand)
  await p.seite(`(${standknoten})?.scrollIntoView({ block: 'center' })`)
  await p.bild('deckleihe')

  p.notiz('4 — der Verlauf des Tauschs')
  await waehle(p, 'beinigung', 'die Zeile des Tauschs')
  await p.klickeAuf(verbuchen, '»Verbuchen«')
  await p.klickeAuf(verlaufKnopf, 'den Kasten »Verlauf«')
  await p.warte('der Verlauf steht', async () => (await p.seite(verlauf)).length >= 3, 20000).catch(async (fehler) => {
    p.zeige(`Kasten: ${String(await p.seite("return document.querySelector('ol[aria-label=Verlauf]')?.parentElement?.innerText ?? [...document.querySelectorAll('[role=alert]')].map((e) => e.innerText).join(' | ')"))}`)
    await p.bild('verlauf-fehlt')
    throw fehler
  })
  const wechsel = await p.seite(verlauf)
  p.zeige(wechsel.join(' | '))
  p.pruefe('er beginnt mit dem Entwurf', /^Revision 1: Entwurf · /.test(wechsel[0] ?? ''), wechsel[0])
  p.pruefe('er nennt das Angebot', wechsel.some((w) => w.startsWith('Revision 1: Entwurf → Angeboten · ')), wechsel.join(' | '))
  p.pruefe('er endet bei der Einigung', /^Revision 1: Angeboten → Angenommen · /.test(wechsel.at(-1) ?? ''), wechsel.at(-1))
  p.pruefe('die Zeit ist lesbar', !wechsel.some((w) => /\d{4}-\d{2}-\d{2}T/.test(w)), wechsel.join(' | '))
  await p.seite("document.querySelector('ol[aria-label=Verlauf]')?.scrollIntoView({ block: 'center' })")
  await p.bild('verlauf')

  p.pruefe('⌘Q beendet', await p.beende())
}
