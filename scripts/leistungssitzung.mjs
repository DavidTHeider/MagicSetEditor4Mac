// **Zwei Sitzungen, eine Frage: wie lange dauert ein Befehl im laufenden
// Programm?** — QA-04 (#1735).
//
// # Warum Sitzungen und nicht ein Prozess je Befehl
//
// **Ein Prozess je Befehl misst den Start mit.** Wer im Fenster von Karte zu
// Karte geht, startet nichts — er bittet ein laufendes Programm um das
// nächste Bild. Kopflos gibt es dafür auf beiden Seiten genau einen Weg:
//
//  - **MSE2** `--cli <set> --quiet --raw` liest Ausdrücke von der
//    Standardeingabe und antwortet je Zeile mit **einem Satz**: Status (0 gut,
//    1 Warnung, 2 Fehler), Zeilenzahl, Zeilen (`src/main.cpp`, Zeilen 201–207
//    und 217–234; `src/cli/text_io_handler.cpp`, `flushRaw`;
//    `src/cli/cli_main.cpp`, `CLISetInterface::run`; Fassung f9d9356 vom
//    2021-07-06). Ein Satz kommt auch für `:quit`.
//  - **MSE4Mac** `mse-mcp` über stdio: JSON-RPC 2.0, eine Nachricht je Zeile,
//    `tools/call` mit dem Ergebnis als Text in `content` und `isError`
//    (`internal/mcp/server.go`, `internal/mcp/inhalt.go`).
//
// Gemessen wird **von der geschriebenen Zeile bis zur vollständigen
// Antwort** — mit `process.hrtime`, auf derselben Uhr für beide.
//
// # Was hier nicht steht
//
// **Kein `spawn`.** Eine Sitzung bekommt einen Kindprozess — oder in den
// Tests zwei Ströme — und startet nichts selbst. Gestartet wird im Läufer.

import { performance } from 'node:perf_hooks'

const statuszeile = /^[0-2]$/
const zahlzeile = /^\d+$/

/**
 * Rohleser setzt die Sätze des Rohmodus aus Stücken der Standardausgabe zusammen.
 *
 * **Vor dem ersten Satz steht Text**: steht das Set vor `--raw`, schreibt MSE2
 * die Meldungen beim Laden, bevor der Rohmodus gilt (`src/main.cpp`, Zeilen
 * 221–231 — die Argumente werden der Reihe nach gelesen). Bis zum ersten Satz
 * wird deshalb verworfen, was nicht wie sein Anfang aussieht: eine Statuszeile
 * und darunter eine Zahl. **Danach nicht mehr** — dort ist eine Zeile, die
 * keine Zahl ist, ein Bruch des Protokolls und kein Vorspann.
 */
export class Rohleser {
  #zeilen = []
  #rest = ''
  #begonnen = false

  /** fuettere nimmt ein Stück und gibt jeden Satz, der damit vollständig ist. */
  fuettere(stueck) {
    this.#rest += stueck
    let ende
    while ((ende = this.#rest.indexOf('\n')) >= 0) {
      this.#zeilen.push(this.#rest.slice(0, ende))
      this.#rest = this.#rest.slice(ende + 1)
    }
    const saetze = []
    while (this.#zeilen.length >= 2) {
      const [status, anzahl] = this.#zeilen
      if (!statuszeile.test(status) || !zahlzeile.test(anzahl)) {
        if (this.#begonnen) throw new Error(`kein Satz des Rohmodus: ${JSON.stringify(status)}`)
        this.#zeilen.shift()
        continue
      }
      if (this.#zeilen.length < 2 + Number(anzahl)) break
      const satz = this.#zeilen.splice(0, 2 + Number(anzahl))
      this.#begonnen = true
      saetze.push({ status: Number(satz[0]), zeilen: satz.slice(2) })
    }
    return saetze
  }
}

/** Mse2Sitzung schickt Ausdrücke an `magicseteditor --cli … --raw`. */
export class Mse2Sitzung {
  #wartend = []
  #beendet = false

  constructor(kind) {
    this.kind = kind
    const leser = new Rohleser()
    kind.stdout.setEncoding('utf8')
    kind.stdout.on('data', (stueck) => {
      let saetze
      try {
        saetze = leser.fuettere(stueck)
      } catch (fehler) {
        for (const w of this.#wartend.splice(0)) w.scheitern(fehler)
        return
      }
      for (const satz of saetze) this.#wartend.shift()?.erfuellen(satz)
    })
    kind.stdout.on('end', () => {
      this.#beendet = true
      for (const w of this.#wartend.splice(0)) w.scheitern(new Error('MSE2 hat sich beendet, bevor es antwortete'))
    })
  }

  /** befehl schickt eine Zeile und gibt den Satz samt Dauer in Millisekunden. */
  befehl(zeile) {
    if (this.#beendet) return Promise.reject(new Error('MSE2 hat sich beendet, bevor es antwortete'))
    return new Promise((erfuellen, scheitern) => {
      const anfang = performance.now()
      this.#wartend.push({
        erfuellen: (satz) => erfuellen({ ...satz, dauerMs: performance.now() - anfang }),
        scheitern,
      })
      this.kind.stdin.write(zeile + '\n')
    })
  }
}

/** McpSitzung spricht JSON-RPC 2.0 über die Ströme von `mse-mcp`. */
export class McpSitzung {
  #naechste = 1
  #offen = new Map()
  #rest = ''

  constructor(kind) {
    this.kind = kind
    kind.stdout.setEncoding('utf8')
    kind.stdout.on('data', (stueck) => this.#lies(stueck))
    kind.stdout.on('end', () => {
      for (const [, w] of this.#offen) w.scheitern(new Error('mse-mcp hat sich beendet, bevor es antwortete'))
      this.#offen.clear()
    })
  }

  #lies(stueck) {
    this.#rest += stueck
    let ende
    while ((ende = this.#rest.indexOf('\n')) >= 0) {
      const zeile = this.#rest.slice(0, ende).trim()
      this.#rest = this.#rest.slice(ende + 1)
      if (zeile === '') continue
      const nachricht = JSON.parse(zeile)
      const wartend = this.#offen.get(nachricht.id)
      if (!wartend) continue
      this.#offen.delete(nachricht.id)
      if (nachricht.error) wartend.scheitern(new Error(nachricht.error.message))
      else wartend.erfuellen(nachricht.result)
    }
  }

  /** rufe schickt eine Anfrage und gibt ihr Ergebnis. */
  rufe(methode, parameter) {
    const id = this.#naechste++
    return new Promise((erfuellen, scheitern) => {
      this.#offen.set(id, { erfuellen, scheitern })
      this.kind.stdin.write(JSON.stringify({ jsonrpc: '2.0', id, method: methode, params: parameter }) + '\n')
    })
  }

  /** beginne stellt sich vor und meldet, dass es losgehen kann. */
  async beginne() {
    const antwort = await this.rufe('initialize', {
      protocolVersion: '2025-06-18', capabilities: {},
      clientInfo: { name: 'leistungsvergleich', version: '1' },
    })
    this.kind.stdin.write(JSON.stringify({ jsonrpc: '2.0', method: 'notifications/initialized' }) + '\n')
    return antwort
  }

  /** werkzeug ruft ein Werkzeug und gibt seinen Wert — JSON, wo es welches ist. */
  async werkzeug(name, argumente) {
    const anfang = performance.now()
    const ergebnis = await this.rufe('tools/call', { name, arguments: argumente })
    const dauerMs = performance.now() - anfang
    const text = (ergebnis.content ?? []).map((teil) => teil.text).join('\n')
    if (ergebnis.isError) throw new Error(`${name}: ${text}`)
    let wert = text
    try {
      wert = JSON.parse(text)
    } catch {
      // Ein Text, der kein JSON ist, bleibt Text.
    }
    return { wert, dauerMs }
  }
}
