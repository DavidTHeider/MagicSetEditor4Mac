// Baut die MCP-Rufe für die A03-Gegenprobe im Gast — dieselben Messungen wie das Drehbuch.
import { writeFileSync } from 'node:fs'
import { join } from 'node:path'
import { texteMessen, kantenMessen, eingabenMessen, symboleMessen } from './a03kopie.mjs'

const ziel = process.argv[2]
const nur = { seite: (js) => js }
const ruf = (name, js) => writeFileSync(join(ziel, `${name}.json`), JSON.stringify({
  jsonrpc: '2.0', id: 1, method: 'tools/call',
  params: { name: 'js_eval', arguments: { js: `return JSON.stringify(await (async () => { ${js} })())` } },
}))
ruf('lage', "return { erscheinungsbild: document.documentElement.dataset.erscheinungsbild ?? '', mantine: document.documentElement.dataset.mantineColorScheme ?? '', dunkel: matchMedia('(prefers-color-scheme: dark)').matches, einfuehrung: !!document.querySelector('[data-einfuehrung]')?.checkVisibility(), dialog: [...document.querySelectorAll('[role=dialog]')].some((d) => d.checkVisibility()), ua: navigator.userAgent }")
for (const [name, wurzel] of [['einfuehrung', '[data-einfuehrung]'], ['palette', '[role=dialog]']]) {
  ruf(`${name}-texte`, texteMessen(nur, wurzel))
  ruf(`${name}-kanten`, kantenMessen(nur, wurzel))
  ruf(`${name}-felder`, eingabenMessen(nur, wurzel))
  ruf(`${name}-symbole`, symboleMessen(nur, wurzel))
}
// **Die Gegenprobe**: die Regeln aus `mantinemerkmale.css` entfernen — das ist der Stand von 0.7.0.
ruf('ohne-korrektur', "let weg = 0; for (const blatt of document.styleSheets) { let regeln; try { regeln = blatt.cssRules } catch { continue } for (let i = regeln.length - 1; i >= 0; i--) { const r = regeln[i]; if (r.selectorText === ':root[data-mantine-color-scheme]' && r.style.getPropertyValue('--mantine-color-text').includes('--mse-text')) { blatt.deleteRule(i); weg++ } } } return weg")
