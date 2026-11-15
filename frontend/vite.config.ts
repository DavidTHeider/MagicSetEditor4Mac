import { readdirSync, readFileSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

import { defineConfig } from 'vitest/config'
import type { Plugin } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// `main.go` bettet `frontend/dist` über `go:embed` ein, und `go:embed` scheitert
// beim Übersetzen, wenn das Verzeichnis fehlt. Nach einem frischen Klon gibt es
// es nur, weil `.gitkeep` verfolgt wird — und `emptyOutDir` räumt genau die
// wieder weg. Ohne diesen Haken stünde nach jedem Bau eine gelöschte, verfolgte
// Datei im Arbeitsbaum, und irgendwann committet sie jemand mit.
function platzhalterBewahren(): Plugin {
  return {
    name: 'platzhalter-bewahren',
    apply: 'build',
    closeBundle() {
      writeFileSync(join(import.meta.dirname, 'dist', '.gitkeep'), '')
    },
  }
}

// **Der Bildweg gehört Go, nicht Vite** —
// #1270.
//
// # Warum das nötig ist
//
// Eine Kartenvorschau kommt seit #1270 als **Adresse** unter `/vorschau/` und
// nicht mehr als Data-URI; die Bytes liefert der Handler aus
// `internal/app/bildweg.go` über den Assetserver, den Wails ohnehin betreibt.
//
// Im gebauten Programm trägt das: Wails fragt das eingebettete `dist`, findet
// nichts und reicht an den Handler weiter.
//
// **Unter `wails dev` nicht.** Dort steht Vite davor, und Vite beantwortet
// jeden unbekannten Weg mit `index.html` — die Rückfallebene einer
// Einseitenanwendung. Wails greift auf seinen Handler aber nur zurück, wenn der
// Entwicklungsserver **404** oder **405** meldet
// (`pkg/assetserver/assethandler_external.go`, `ModifyResponse`); eine 200 mit
// HTML sieht für ihn aus wie eine gefundene Datei. Nachgemessen am 2026-08-28:
// `GET /vorschau/1.png` gab `200 text/html, 562 Byte` — auch mit einem
// `Accept`, das nur Bilder nennt.
//
// # Warum ein Vierhundertvier und kein Weiterleiten
//
// Ein `server.proxy` auf `34115` liefe im Kreis: Wails reicht GET **zuerst** an
// Vite. Die 404 ist die Antwort, auf die Wails ohnehin wartet.
//
// # Und es sind zwei Vorsätze, nicht einer
//
// **`/spielbild/` kam mit G2 dazu** (#1459)
// und stand hier zuerst **nicht**. Gemessen am 2026-09-15 unter `wails dev`:
// `GET /spielbild/<hash>.png` gab `200 text/html` — Vites Rückfallebene —,
// während `GET /vorschau/…` sauber `404 text/plain` gab. Am Spieltisch hingen
// achtzehn Bilder mit `naturalWidth` **null**: die Adresse war richtig, das
// Teilen griff, und zu sehen war nichts.
//
// **Im gebauten Programm wäre es nie aufgefallen** — dort fragt Wails das
// eingebettete `dist`, findet nichts und reicht weiter. Nur die Sichtprüfung
// findet das (§3), und auch die nur, wenn jemand hinsieht.
//
// # Und die Liste wird gelesen, nicht gepflegt
//
// **Ein zweiter Ort für dieselbe Wahrheit geht wieder schief.** Der erste
// Anlauf schrieb die Vorsätze hier noch einmal hin und ließ einen Wächter
// nachzählen. Das fängt den Fehler — aber erst in `make ci`, zwanzig Minuten
// später, und dazwischen sitzt jemand vor leeren Kästchen und sucht in der
// falschen Schicht.
//
// **Also steht die Wahrheit an einer Stelle**: in `internal/vorschauweg`.
// Diese Datei liest die Konstanten von dort und weiß damit von jedem neuen
// Vorsatz, ohne dass jemand daran denkt.
//
// **Und sie bricht laut, wenn sie nichts findet.** Ein Ausdruck, der ins
// Leere greift, gäbe eine leere Liste — und die verhielte sich genau wie der
// Fehler, den es hier zu verhindern gilt: still.
const vorsatzregel = /^const \w*[Pp]fad = '(\/[a-z0-9-]+\/)'$/gm

function bildvorsaetzeAusGo(): string[] {
  const ort = join(import.meta.dirname, '..', 'internal', 'vorschauweg')
  const gefunden = new Set<string>()

  for (const datei of readdirSync(ort)) {
    if (!datei.endsWith('.go') || datei.endsWith('_test.go')) {
      continue
    }

    const inhalt = readFileSync(join(ort, datei), 'utf8').replaceAll('"', "'")

    for (const treffer of inhalt.matchAll(vorsatzregel)) {
      const vorsatz = treffer[1]

      if (vorsatz !== undefined) {
        gefunden.add(vorsatz)
      }
    }
  }

  if (gefunden.size === 0) {
    throw new Error(
      'in internal/vorschauweg steht kein Bildvorsatz — entweder ist das ' +
        'Paket weg, oder die Konstanten heißen anders und dieser Ausdruck ' +
        'greift ins Leere. Ohne sie beantwortet Vite jeden Bildweg mit ' +
        'index.html, und jedes Kartenbild bleibt leer.',
    )
  }

  return [...gefunden]
}

const bildvorsaetze = bildvorsaetzeAusGo()

function bildwegAnGo(): Plugin {
  return {
    name: 'bildweg-an-go',
    apply: 'serve',
    configureServer(server) {
      server.middlewares.use((anfrage, antwort, weiter) => {
        const weg = anfrage.url ?? ''

        if (bildvorsaetze.some((vorsatz) => weg.startsWith(vorsatz))) {
          antwort.statusCode = 404
          antwort.end()

          return
        }

        weiter()
      })
    },
  }
}

// Die Abdeckungsschwelle ist hart und liegt bei 100 % (AGENTS.md §3).
//
// Nicht als Zielmarke, sondern als Tor: eine weiche Schwelle wird bei jedem
// eiligen Merge um einen Punkt gesenkt, und niemand stimmt darüber ab.
//
// Ausgenommen ist nur, was sich ohne echtes Fenster nicht ausführen lässt. Die
// Begründung je Ausnahme steht in `.covignore` im Projektstamm — hier stünde
// sie an zweiter Stelle und liefe mit der ersten auseinander.
export default defineConfig({
  plugins: [react(), tailwindcss(), platzhalterBewahren(), bildwegAnGo()],

  // **Was ein Werkzeug hier ablegt, ist keine Änderung am Programm.**
  //
  // Der Riegel (`make vorpruefung`) fährt `vitest --coverage`, und der
  // schreibt rund 900 HTML-Dateien nach `frontend/coverage/lcov-report/` —
  // eine je Quelltextdatei. Vite beobachtet sein Wurzelverzeichnis und hielt
  // jede davon für eine Änderung: am 2026-09-17 lud das Fenster in Folge
  // hunderte Male neu, und in der Zeit ließ sich nichts ansehen. Die
  // Sichtprüfung nach einem Riegellauf war damit unbrauchbar.
  //
  // **Weder der Riegel noch Vite tun etwas Falsches** — falsch ist nur, dass
  // der eine im Hof des anderen ablegt. `dist/` steht aus demselben Grund
  // dabei; es entsteht seltener, aber es entsteht genauso.
  //
  // Ein Wächter hält es: `internal/werkzeuge/beobachterausnahmen_test.go`.
  server: {
    watch: {
      ignored: ['**/coverage/**', '**/dist/**'],
    },
  },

  build: {
    // Wails bettet dieses Verzeichnis ein (siehe wails.json, Issue #3).
    outDir: 'dist',
    emptyOutDir: true,
  },
  test: {
    environment: 'jsdom',

    // **Kein `--no-experimental-webstorage`, und das gehört begründet.**
    //
    // Der dritte Durchgang über Phase 6 hat es vorgeschlagen: Node 24 lege ein
    // eigenes `localStorage` vor jsdom, und die Sprachtests erreichten keine
    // einzige Zusicherung. **Nachgemessen reproduziert das hier nicht** — unter
    // dem gepinnten Node 24.15.0 laufen 38 Dateien und 305 Fälle mit und ohne
    // die Angabe gleich durch, einzeln wie im vollen Lauf.
    //
    // Node 24 legt `localStorage` nur an, wenn `--experimental-webstorage`
    // **und** `--localstorage-file` gesetzt sind. Wer die Angabe hier aufnähme,
    // schriebe eine Begründung hin, die für dieses Projekt nicht stimmt — und
    // eine falsche Begründung ist schlimmer als keine Änderung (§10).
    globals: true,
    setupFiles: ['./src/test/einrichtung.ts'],

    // **Die Grenze eines Falls liegt über der von `waitFor`.**
    //
    // `einrichtung.ts` setzt `asyncUtilTimeout` auf 5 s, damit ein Lauf auf
    // einem beschäftigten Rechner nicht an der Uhr scheitert. Stünde die
    // Grenze des Falls ebenso bei 5 s, käme ein **echter** Fehlschlag als
    // „Test timed out" heraus statt als „Unable to find …" — die Meldung, die
    // sagt, was fehlt, ginge verloren. Genau das ist am 2026-08-21 passiert.
    //
    // **Sechzig seit dem 2026-09-09**, mit `asyncUtilTimeout` auf 15 s: drei
    // Wartezeiten passen in einen Fall, und ein echter Fehlschlag heißt
    // weiterhin „Unable to find …".
    testTimeout: 60_000,
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov'],
      include: ['src/**/*.{ts,tsx}'],
      exclude: [
        'src/**/*.test.{ts,tsx}',
        'src/test/**',
        'src/main.tsx',

        // **Der Ruf an PixiJS** — Werkstück G3 von Phase 5.5, Issue 1460.
        // Der Grund steht in `.covignore`, und ein Wächter hält beide Listen
        // zusammen: eine Ausnahme, die nur an einer Stelle steht, wirkt nicht.
        //
        // **Ohne eckige Klammern hier drin**, und das ist keine Marotte: der
        // Wächter liest dieses Feld mit einem nicht gierigen Klammerausdruck,
        // und eine Verweisklammer im Kommentar beendete ihm die Liste zu früh.
        'src/tisch/pixinaht.ts',
        'src/tisch/pixibewegung.ts',
        'src/tisch/pixihintergrund.ts',
        'src/tisch/pixibeiwerk.ts',
        'src/tisch/pixitextur.ts',
      ],
      thresholds: {
        statements: 100,
        branches: 100,
        functions: 100,
        lines: 100,
      },
    },
  },
})
