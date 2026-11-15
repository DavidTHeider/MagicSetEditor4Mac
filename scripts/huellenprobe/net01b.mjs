// **Eine Einladung genügt, am nativen Fenster** — NET-01b
// (#1954):
//
// - Bert horcht in einem eigenen `mse horchen` und lädt an der
//   Kommandozeile **einlösbar** ein (`--adresse`),
// - Anna gibt den Code im Fenster ein; gefragt wird, wie sie sich nennt —
//   **leer**, nicht vorausgefüllt —, und nach „Später vergleichen" steht da,
//   dass Bert sie jetzt kennt,
// - bei Bert steht Anna danach unbestätigt und unter dem Namen, den sie
//   sich gab — ohne dass Anna einen Code zurückgeschickt hat.
//
// **Der Code kommt über JavaScript ins Feld** — derselbe Grund wie in
// `net01.mjs`: fünfhundert Zeichen, und die Zwischenablage bleibt unberührt.
import { spawn } from 'node:child_process'
import { join } from 'node:path'
import { kindumgebung } from '../huellenorte.mjs'

const eingabefeld = '[data-einladungstausch] textarea:not([readonly])'
const setze = (text) => `(() => { const feld = document.querySelector('${eingabefeld}'); const setzer = Object.getOwnPropertyDescriptor(HTMLTextAreaElement.prototype, 'value').set; setzer.call(feld, ${JSON.stringify(text)}); feld.dispatchEvent(new Event('input', { bubbles: true })); return feld.value.length })()`
const meldefeld = "[...document.querySelectorAll('[data-einladungsvergleich] label')].find((l) => l.innerText.trim().startsWith('Wie soll dein Gegenüber dich sehen?'))?.querySelector('input')"
const zeilenwert = (ausgabe, name) => ausgabe.split('\n').find((z) => z.startsWith(`${name}\t`))?.split('\t')[1] ?? ''

// horche startet `mse horchen` für einen Ort und gibt die Adresse aus seiner
// ersten Zeile — der Prozess bleibt, bis die Probe aufräumt.
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
    prozess.on('exit', (code) => verworfen(new Error(`mse horchen endete mit ${code}`)))
  })
}

export default async function net01b(p) {
  const anna = p.ort()
  const bert = p.ort()

  p.notiz('1 — Bert horcht in einem eigenen Prozess und lädt einlösbar ein')
  for (const ort of [anna, bert]) {
    p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
    p.mse(ort, 'peer-anlegen')
  }
  const adresse = await horche(p, bert)
  p.zeige(`Bert horcht unter ${adresse}`)
  const bertsCode = zeilenwert(p.mse(bert, 'einladung-erstellen', 'Bert', 'handel', '--adresse', adresse), 'code')
  p.pruefe('die Vorschau an der Kommandozeile nennt die Adresse', p.mse(anna, 'einladung-pruefen', bertsCode).includes(adresse))

  await p.starte({ ort: anna })
  await p.taste('5', 'befehl')
  // **Seit UXR-15 in der Schublade hinter „Neuer Tausch …"** (#1923).
  await p.klickeText('Neuer Tausch …')
  await p.klickeText('Gegenüber aufnehmen oder bestätigen')
  await p.wartetext('Code eingeben')

  p.notiz('2 — Anna gibt Berts Code ein; gefragt wird, wie sie sich nennt')
  await p.klickeText('Code eingeben')
  await p.warte('das Eingabefeld steht', () => p.seite(`return document.querySelector('${eingabefeld}') !== null`))
  p.pruefe('der Code steht im Feld', (await p.seite(`return ${setze(bertsCode)}`)) === bertsCode.length)
  await p.klickeText('Prüfen')
  await p.wartetext('Bert bestätigen')
  p.pruefe('das Namensfeld steht da und ist leer', (await p.seite(`return ${meldefeld}?.value`)) === '')
  await p.klickeAuf(meldefeld, 'das Namensfeld für Bert')
  await p.tippeInsFeld('Anna')
  await p.bild('vorher')

  p.notiz('3 — später vergleichen: Anna ist aufgenommen, und Bert kennt sie')
  await p.klickeText('Später vergleichen')
  await p.wartetext('Bert kennt dich jetzt auch')
  await p.bild('gemeldet')

  const kontakte = p.mse(bert, 'kontakte')
  p.pruefe('bei Bert steht Anna, unbestätigt und unter ihrem Namen', /^unbestaetigt\t\S+\t\S+\tAnna$/m.test(kontakte), kontakte)
}
