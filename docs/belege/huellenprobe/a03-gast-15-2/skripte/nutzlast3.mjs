import { writeFileSync } from 'node:fs'
import { join } from 'node:path'
const ziel = process.argv[2]
const ruf = (name, js) => writeFileSync(join(ziel, `${name}.json`), JSON.stringify({ jsonrpc: '2.0', id: 1, method: 'tools/call', params: { name: 'js_eval', arguments: { js: `return JSON.stringify(await (async () => { ${js} })())` } } }))
ruf('hell-setzen', "const a = await window.go.app.Bindungen.EinstellungSetzen('erscheinungsbild', 'hell'); return a.fehler ?? 'gesetzt'")
ruf('neu-laden', "setTimeout(() => location.reload(), 50); return 'neu'")
