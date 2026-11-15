// Prüfungen von scripts/lastlaufdienste.mjs — Z4 des Zweitblicks über Phase 6.11.
//
// **Ein abgebrochener Lastlauf räumt auf, und ein Ruf wartet nicht ewig.**
// Gestartete Dienste blieben liegen, wenn der zweite nicht anlief, und ein
// MCP-Ruf über stdio, dessen Prozess gestorben war, wartete auf eine Antwort,
// die nie kam. Die Dienste hier sind kleine Node-Programme, die sich
// verhalten, wie der Test es will.

import { spawn as erzeuge } from 'node:child_process'
import { after, test as pruefung } from 'node:test'
import assert from 'node:assert/strict'
import { setTimeout as schlafen } from 'node:timers/promises'

import { Kinderregister, beende, mcpUeberStdio, starteDienst, starteNacheinander } from './lastlaufdienste.mjs'

/** test gibt jedem Fall eine Frist: ein Ruf, der nie endet, soll rot werden und nicht hängen. */
const test = (name, fn) => pruefung(name, { timeout: 8_000 }, fn)

/**
 * Jedes Register dieser Datei wird am Ende beendet. **Ein Fall, der scheitert,
 * lässt seine Programme sonst laufen** — und der Läufer der Prüfungen wartete
 * auf sie, statt rot zu enden.
 */
const register = []
const neuesRegister = () => {
  const neu = new Kinderregister(erzeuge)
  register.push(neu)
  return neu
}
after(async () => {
  for (const r of register) await r.beendeAlle(300)
})

const lebt = (pid) => {
  try {
    process.kill(pid, 0)
    return true
  } catch {
    return false
  }
}

/** tot wartet kurz, bis der Prozess weg ist — das Betriebssystem braucht einen Augenblick. */
async function tot(pid) {
  for (let i = 0; i < 50 && lebt(pid); i++) await schlafen(20)
  return !lebt(pid)
}

const knoten = (code) => [process.execPath, ['-e', code]]

/** Ein Dienst, der Adresse und Marke ansagt und dann läuft. */
const ansagender = `console.log('Dienst auf http://127.0.0.1:4711/mcp'); console.log('Marke abc123'); setInterval(() => {}, 1000); process.on('SIGTERM', () => process.exit(0))`
/** Ein Dienst, der gleich endet, ohne etwas anzusagen. */
const sterbender = 'console.error("kein Platz"); process.exit(3)'
/** Ein Dienst, der schweigt. */
const schweigender = 'setInterval(() => {}, 1000)'
/** Ein Dienst, der SIGTERM überhört. */
const taub = `console.log('Dienst auf http://127.0.0.1:1'); console.log('Marke m'); process.on('SIGTERM', () => {}); setInterval(() => {}, 1000)`

test('ein Dienst sagt Adresse, Pfad und Marke an und läuft bis zum Ende', async () => {
  const register = neuesRegister()
  const [programm, argumente] = knoten(ansagender)
  const dienst = await starteDienst(register, programm, argumente, process.env, 16)
  assert.deepEqual([dienst.adresse, dienst.pfad, dienst.marke], ['http://127.0.0.1:4711', '/mcp', 'abc123'])
  assert.equal(register.zahl, 1)
  const pid = dienst.kind.pid
  await beende(dienst)
  assert.equal(await tot(pid), true)
  assert.equal(register.zahl, 0, 'ein beendeter Prozess steht nicht mehr im Register')
})

test('ein Dienst, der ohne Ansage endet, scheitert mit seinem Ende und seiner Mitschrift', async () => {
  const register = neuesRegister()
  const [programm, argumente] = knoten(sterbender)
  await assert.rejects(starteDienst(register, programm, argumente, process.env, 16), /endete mit 3.*kein Platz/s)
  assert.equal(register.zahl, 0)
})

test('ein Dienst, der nichts ansagt, wird nach der Frist abgewiesen — und beendet', async () => {
  const register = neuesRegister()
  const [programm, argumente] = knoten(schweigender)
  const kinder = []
  const alt = register.starte.bind(register)
  register.starte = (...a) => { const kind = alt(...a); kinder.push(kind); return kind }
  await assert.rejects(starteDienst(register, programm, argumente, process.env, 16, { frist: 150 }), /hat nichts angesagt/)
  assert.equal(await tot(kinder[0].pid), true, 'der Prozess läuft nicht als Waise weiter')
})

test('scheitert der zweite Dienst beim Start, wird der erste beendet', async () => {
  const register = neuesRegister()
  const erster = knoten(ansagender)
  const zweiter = knoten(sterbender)
  const kinder = []
  const alt = register.starte.bind(register)
  register.starte = (...a) => { const kind = alt(...a); kinder.push(kind); return kind }
  await assert.rejects(starteNacheinander(register, [
    [erster[0], erster[1], process.env, 16], [zweiter[0], zweiter[1], process.env, 64],
  ]), /endete mit 3/)
  assert.equal(kinder.length, 2)
  assert.equal(await tot(kinder[0].pid), true, 'der erste Dienst ist weg')
  assert.equal(register.zahl, 0)
})

test('starten beide, bekommt der Aufrufer beide', async () => {
  const register = neuesRegister()
  const [programm, argumente] = knoten(ansagender)
  const dienste = await starteNacheinander(register, [[programm, argumente, process.env, 16], [programm, argumente, process.env, 64]])
  assert.equal(dienste.length, 2)
  assert.equal(register.zahl, 2)
  await register.beendeAlle()
  assert.equal(register.zahl, 0)
})

test('beendeAlle holt, was SIGTERM überhört, mit SIGKILL nach der Frist', async () => {
  const register = neuesRegister()
  const [programm, argumente] = knoten(taub)
  const dienst = await starteDienst(register, programm, argumente, process.env, 16)
  await register.beendeAlle(200)
  assert.equal(await tot(dienst.kind.pid), true)
})

test('beendeSofort beendet ohne zu warten — für den Handler beim Verlassen des Prozesses', async () => {
  const register = neuesRegister()
  const [programm, argumente] = knoten(taub)
  const dienst = await starteDienst(register, programm, argumente, process.env, 16)
  register.beendeSofort()
  assert.equal(await tot(dienst.kind.pid), true)
})

/** mcpProgramm gibt ein Programm, das über stdio antwortet: `verhalten` sagt, was es bei einem Werkzeugruf tut. */
const mcpProgramm = (verhalten) => knoten(`
const rl = require('node:readline').createInterface({ input: process.stdin })
rl.on('line', (zeile) => {
  const n = JSON.parse(zeile)
  if (n.method === 'initialize') return console.log(JSON.stringify({ jsonrpc: '2.0', id: n.id, result: {} }))
  if (n.method === 'tools/call') { ${verhalten} }
})
`)

const antwortet = "console.log(JSON.stringify({ jsonrpc: '2.0', id: n.id, result: { content: [{ type: 'text', text: '[1,2]' }] } }))"

test('ein Werkzeugruf über stdio bekommt seine Antwort — samt Dauer, Status und Inhalt', async () => {
  const register = neuesRegister()
  const [programm, argumente] = mcpProgramm(antwortet)
  const sitzung = await mcpUeberStdio(register, programm, argumente, process.env)
  const a = await sitzung.ruf('lesen.karten', { setID: 1 })
  assert.equal(a.status, 200)
  assert.equal(a.inhalt, '[1,2]')
  assert.equal(a.mcpFehler, false)
  assert.ok(a.ms >= 0)
  await sitzung.ende()
  assert.equal(register.zahl, 0)
})

test('stirbt der Prozess mitten im Ruf, wird der Ruf abgewiesen — und wartet nicht ewig', async () => {
  const register = neuesRegister()
  const [programm, argumente] = mcpProgramm('process.exit(7)')
  const sitzung = await mcpUeberStdio(register, programm, argumente, process.env)
  await assert.rejects(sitzung.ruf('lesen.karten', {}), /endete \(Code 7.*bevor es antwortete/)
})

test('alle offenen Rufe werden abgewiesen, nicht nur einer', async () => {
  const register = neuesRegister()
  const [programm, argumente] = mcpProgramm('setTimeout(() => process.exit(9), 100)')
  const sitzung = await mcpUeberStdio(register, programm, argumente, process.env)
  const ergebnisse = await Promise.allSettled([sitzung.ruf('a', {}), sitzung.ruf('b', {}), sitzung.ruf('c', {})])
  assert.deepEqual(ergebnisse.map((e) => e.status), ['rejected', 'rejected', 'rejected'])
})

test('ein Ruf nach dem Ende wird sofort abgewiesen', async () => {
  const register = neuesRegister()
  const [programm, argumente] = mcpProgramm('process.exit(7)')
  const sitzung = await mcpUeberStdio(register, programm, argumente, process.env)
  await assert.rejects(sitzung.ruf('a', {}))
  const vorher = Date.now()
  await assert.rejects(sitzung.ruf('b', {}), /endete/)
  assert.ok(Date.now() - vorher < 500, 'ohne zu warten')
})

test('eine Antwort, die ausbleibt, weist der Ruf nach der Frist ab', async () => {
  const register = neuesRegister()
  const [programm, argumente] = mcpProgramm('')
  const sitzung = await mcpUeberStdio(register, programm, argumente, process.env, { frist: 150 })
  await assert.rejects(sitzung.ruf('lesen.karten', {}), /keine Antwort auf tools\/call nach 150 ms/)
  await sitzung.ende(200)
})

test('ein Prozess, der nie anlief, scheitert mit dem Grund', async () => {
  const register = neuesRegister()
  await assert.rejects(mcpUeberStdio(register, '/gibt/es/nicht', [], process.env), /ENOENT|nicht gestartet/)
})
