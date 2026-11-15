// **Der Nahtlauf — durch die laufende Wails-Brücke, nicht durch eine
// Vorrichtung** (#1288).
//
// # Wozu
//
// Es gibt vier Flächen, und für drei davon fährt `internal/flaechen` einen
// Lauf. Für das Fenster nicht: Go-Tests sehen die Struct-Marken nicht,
// Frontend-Tests gehen über `src/test/api.ts` und nie durch die
// JSON-Serialisierung (§3). **Zwei grüne Seiten einer Grenze sind kein
// Nachweis für die Grenze.**
//
// Dieser Lauf startet die Anwendung im Browsermodus (`scripts/entwickeln.sh
// browser`) mit **eigenem Zuhause**, schickt einen Chrome ohne Fenster nach
// `34115` und ruft `window.go.app.Bindungen.*` — also dieselben Aufrufe, die
// die Oberfläche macht, über dieselbe Brücke. Seit Wails 3 (ENG-02) ist das
// der Weg über HTTP, den `@wailsio/runtime` auch im Fenster nimmt; was das
// Fenster **anders** macht, steht in `docs/WAILS3.md`.
//
// ```bash
// node scripts/nahtlauf.mjs
// ```
//
// # Warum nicht in `make ci`
//
// Weil er **baut**: `scripts/entwickeln.sh` übersetzt die Anwendung und
// startet Vite, und das kostet mehr als die ganze Pipeline. Er gehört auf die Liste vor jeder
// Freigabe (`docs/RELEASE.md`), zu den anderen Läufen, die dort stehen, weil
// sie sonst niemand fährt.
//
// **Ein Fenster öffnet er seit Wails 3 nicht mehr** — der Browsermodus ist
// ein HTTP-Server. Einen entfernten Lauf gibt es trotzdem nicht
// ([0007](../docs/entscheidungen/0007-kein-entfernter-lauf.md)).
//
// # Was er **nicht** prüft
//
//   - **Nicht alle Bindungen.** Antwort**formen** aus `nahtvertrag.mjs`,
//     keine Bindungen. Wer eine Bindung hinzufügt, prüft sie hier nicht
//     mit — das tut `internal/app/leerelistenlaufzeit_test.go` auf der
//     Go-Seite.
//   - **Nichts am Aussehen.** Dafür gibt es die Sichtprüfung (§3) und die
//     Hüllenprobe `readmebilder`.
//   - **Keine Dauer als Zusage.** Er misst die Zeit je Fall und schreibt sie
//     hin; eine Schranke daraus zu machen hieße, die Geschwindigkeit dieses
//     Rechners zur Zusage zu erheben.

import { spawn } from "node:child_process";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { schicke, verbinden, warte, werte } from "./abzuegecdp.mjs";
import { faelle } from "./nahtvertrag.mjs";

const adresse = process.argv[2] ?? "http://localhost:34115";

// pruefungen sind die Zusagen, die ein Fall stellen kann.
//
// **Jede ist ein Satz über eine Stelle der Antwort.** Ein Vertrag, der aus
// Code bestünde, wäre ein zweites Programm — und zwei Programme, die dasselbe
// meinen, laufen auseinander.
//
// **Jede nimmt genau eine Angabe, und die ist flach** — eine Liste von Pfaden,
// eine Tabelle Pfad→Wert oder ein Paar. Verschachtelte Angaben lasen sich von
// außen nicht mehr von den Prüfungsarten unterscheiden, und der Wächter, der
// beide Dateien zusammenhält (`internal/werkzeuge/nahtlauf_test.go`), hielt
// `pfad` und `zahl` für Prüfungsarten.
const pruefungen = {
  liste: (wert, pfade) =>
    pfade.map((pfad) =>
      Array.isArray(an(wert, pfad))
        ? null
        : `${pfad} ist keine Liste, sondern ${JSON.stringify(an(wert, pfad))}`,
    ),
  // **Leer heißt: eine Liste, und die hat keine Glieder.** Ohne die erste
  // Hälfte ginge eine leere Zeichenkette durch, und `""` ist keine leere
  // Liste (Nachlauf über Phase 4.1, Befund 3).
  leer: (wert, pfade) =>
    pfade.map((pfad) => {
      const gefunden = an(wert, pfad);

      return Array.isArray(gefunden) && gefunden.length === 0
        ? null
        : `${pfad} ist ${JSON.stringify(gefunden)} und keine leere Liste`;
    }),
  // **Eine Tabelle und ausdrücklich keine Liste.** `typeof [] === "object"`
  // ist in JavaScript wahr, und damit ging ein `werte: []` als Tabelle durch —
  // dieselbe kaputte Antwort, die diese Form fangen soll. Angestrichen vom
  // Zweitblick über Phase 4.1, Befund 3.
  objekt: (wert, pfade) =>
    pfade.map((pfad) => {
      const gefunden = an(wert, pfad);
      const tabelle =
        gefunden !== null &&
        typeof gefunden === "object" &&
        !Array.isArray(gefunden);

      return tabelle
        ? null
        : `${pfad} ist ${JSON.stringify(gefunden)} und keine Tabelle`;
    }),
  istNull: (wert, pfade) =>
    pfade.map((pfad) =>
      an(wert, pfad) === null ? null : `${pfad} ist nicht null`,
    ),
  // **Vorhanden heißt: da **und** nicht `null`.** Der erste Anlauf wies nur
  // `undefined` ab — ein Feld, das als `null` ankommt, galt als vorhanden.
  // Für eine Liste ist das genau der Bruch, den dieser Vertrag fangen soll
  // (Nachlauf über Phase 4.1, Befund 3).
  vorhanden: (wert, pfade) =>
    pfade.map((pfad) => {
      const gefunden = an(wert, pfad);

      if (gefunden === undefined) {
        return `${pfad} fehlt`;
      }

      return gefunden === null ? `${pfad} ist null` : null;
    }),
  fehlt: (wert, pfade) =>
    pfade.map((pfad) =>
      an(wert, pfad) === undefined
        ? null
        : `${pfad} steht da: ${JSON.stringify(an(wert, pfad))}`,
    ),
  gleich: (wert, erwartet) =>
    Object.entries(erwartet).map(([pfad, soll]) =>
      an(wert, pfad) === soll
        ? null
        : `${pfad} ist ${JSON.stringify(an(wert, pfad))}, erwartet ${JSON.stringify(soll)}`,
    ),
  // **Genau einmal das eine und genau einmal das andere.**
  //
  // Der erste Anlauf verlangte nur „genau ein `abgebrochen`" — und
  // `["abgebrochen", "unbekannt"]` wäre grün gewesen, obwohl **kein** Lauf
  // durchkam. Eine Nebenläufigkeitszusage, die den Erfolg nicht mitprüft,
  // prüft die Hälfte (Zweitblick über Phase 4.1, Befund 3).
  genauJeEinmal: (wert, [pfad, eines, anderes]) => {
    const alle = an(wert, pfad) ?? [];
    const zaehle = (was) => alle.filter((x) => x === was).length;

    return [
      zaehle(eines) === 1 && zaehle(anderes) === 1 && alle.length === 2
        ? null
        : `${pfad} ist ${JSON.stringify(alle)} — erwartet genau ein ${JSON.stringify(eines)} und genau ein ${JSON.stringify(anderes)}`,
    ];
  },
  mindestens: (wert, [pfad, zahl]) => {
    const gefunden = an(wert, pfad);

    return [
      Array.isArray(gefunden) && gefunden.length >= zahl
        ? null
        : `${pfad} ist ${JSON.stringify(gefunden)} — erwartet war eine Liste mit mindestens ${String(zahl)} Einträgen`,
    ];
  },
  // **Beide müssen Listen sein.** Der erste Anlauf setzte eine fehlende Liste
  // auf Länge 0 — und eine **fehlende** Ergebnisliste galt damit als
  // gelungene Verkleinerung (Nachlauf über Phase 4.1, Befund 3).
  kleinerAls: (wert, [pfad, alsPfad]) => {
    const kurz = an(wert, pfad);
    const lang = an(wert, alsPfad);

    if (!Array.isArray(kurz) || !Array.isArray(lang)) {
      return [
        `${pfad} ist ${JSON.stringify(kurz)}, ${alsPfad} ist ${JSON.stringify(lang)} — beide müssen Listen sein`,
      ];
    }

    return [
      kurz.length < lang.length
        ? null
        : `${pfad} (${String(kurz.length)}) ist nicht kürzer als ${alsPfad} (${String(lang.length)})`,
    ];
  },
};

// an liest eine Stelle wie `sets.sets.0.name` aus einem Wert.
function an(wert, pfad) {
  let hier = wert;

  for (const teil of pfad.split(".")) {
    if (hier === null || hier === undefined) {
      return undefined;
    }

    hier = hier[teil];
  }

  return hier;
}

// heimStarten baut ein leeres Zuhause und startet die Anwendung darin.
//
// **Ein eigenes Zuhause, und das ist keine Bequemlichkeit** (§3): der Bestand
// liegt unter `$HOME/Library/Application Support`, und ein Lauf ohne eigenes
// schriebe in die Sets dessen, der ihn fährt.
async function heimStarten() {
  const heim = await mkdtemp(join(tmpdir(), "nahtlauf-heim-"));

  // **Die Go-Verzeichnisse bleiben im Projekt** — angestrichen am 2026-09-07.
  //
  // `HOME` umzulegen legt auch Gos Modulcache um: ohne diese drei Zeilen
  // sucht er ihn unter `$HOME/go/pkg/mod`, findet ein leeres Zuhause und
  // **lädt jede Abhängigkeit neu**. Der Lauf dauerte damit Minuten statt
  // Sekunden — und hinterließ einen Baum aus **schreibgeschützten** Dateien,
  // an dem das Aufräumen unten mit `EACCES` abbrach.
  //
  // `scripts/activate.sh` setzt dieselben drei; wer den Lauf ohne sie fährt,
  // bekam den langen Weg. **Ein Messlauf, dessen Dauer davon abhängt, wie man
  // seine Shell vorbereitet hat, ist zwei Läufe** (§4, sinngemäß).
  const wurzel = join(import.meta.dirname, "..");

  const dev = spawn("./scripts/entwickeln.sh", ["browser"], {
    cwd: wurzel,
    env: {
      ...process.env,
      HOME: heim,
      GOPATH: join(wurzel, ".go"),
      GOMODCACHE: join(wurzel, ".go", "mod"),
      GOCACHE: join(wurzel, ".go", "build"),
    },
    stdio: ["ignore", "ignore", "pipe"],
  });

  let letzteMeldung = "";
  dev.stderr.on("data", (teil) => {
    letzteMeldung = String(teil);
  });

  return { heim, dev, letzte: () => letzteMeldung };
}

// aufWarten wartet, bis die Anwendung antwortet — mit einer Frist.
//
// **Ohne Frist wird aus einem gescheiterten Bau ein hängendes Skript**, und
// das ist der Fehler, den man nicht sieht (derselbe Satz steht in
// `abzuegecdp.mjs`).
async function aufWarten(letzte) {
  for (let versuch = 0; versuch < 240; versuch += 1) {
    try {
      const antwort = await fetch(adresse, { signal: AbortSignal.timeout(2000) });
      if (antwort.ok) {
        return;
      }
    } catch {
      // noch nicht da — die Frist zählt weiter
    }

    await warte(1000);
  }

  throw new Error(`${adresse} kam nicht hoch — zuletzt: ${letzte()}`);
}

// brueckeAbwarten wartet, bis Wails seine Bindungen gesetzt hat.
async function brueckeAbwarten(ws) {
  for (let versuch = 0; versuch < 100; versuch += 1) {
    const da = await werte(
      ws,
      `typeof window.go?.app?.Bindungen?.Sets === "function"`,
    );

    if (da === true) {
      return;
    }

    await warte(200);
  }

  throw new Error("window.go.app.Bindungen kam nicht — steht die Brücke?");
}

// fallFahren ruft einen Fall über die Brücke und gibt seine Befunde.
async function fallFahren(ws, fall) {
  const begonnen = Date.now();

  const roh = await werte(
    ws,
    `(async () => {
      const B = window.go.app.Bindungen;
      return JSON.stringify(await (${fall.ruf}));
    })()`,
  );

  const gedauert = Date.now() - begonnen;

  if (typeof roh !== "string") {
    return { gedauert, befunde: [`kein JSON zurück: ${JSON.stringify(roh)}`] };
  }

  const antwort = JSON.parse(roh);

  const befunde = Object.entries(fall.pruefe)
    .flatMap(([art, angabe]) => pruefungen[art](antwort, angabe))
    .filter((eines) => eines !== null);

  return { gedauert, befunde, antwort };
}

const { heim, dev, letzte } = await heimStarten();

let rot = 0;

try {
  process.stdout.write(`Nahtlauf — Zuhause ${heim}\n`);

  await aufWarten(letzte);

  const profil = await mkdtemp(join(tmpdir(), "nahtlauf-chrome-"));
  const chrome = spawn(
    "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
    [
      "--headless=new",
      "--remote-debugging-port=0",
      `--user-data-dir=${profil}`,
      "about:blank",
    ],
    { stdio: ["ignore", "ignore", "ignore"] },
  );

  try {
    const ws = await verbinden(profil);

    await schicke(ws, "Page.navigate", { url: adresse });
    await brueckeAbwarten(ws);

    for (const fall of faelle) {
      const { gedauert, befunde } = await fallFahren(ws, fall);

      if (befunde.length === 0) {
        process.stdout.write(`  ✓ ${fall.name} (${String(gedauert)} ms)\n`);

        continue;
      }

      rot += 1;
      process.stdout.write(`  ✗ ${fall.name} (${String(gedauert)} ms)\n`);

      // **Die Fälle danach sind nicht mehr unabhängig** — Befund 6 des
      // Nachlaufs über Phase 4.3. Sie tragen ausdrücklich gemeinsamen Zustand
      // (ein Set, das ein früherer Fall angelegt hat); nach einem gebrochenen
      // zustandsbildenden Fall belegt ein grüner Folgefall nichts mehr.
      //
      // **Gefahren werden sie trotzdem**: ein Abbruch beim ersten Bruch nähme
      // die Übersicht, welche Formen sonst noch brechen. Was er nicht mehr
      // hergibt, steht hier — und die Zeile darunter sagt es auch im Ergebnis.
      if (rot === 1) {
        process.stdout.write(
          `      (ab hier tragen die Fälle den Zustand eines gebrochenen mit)\n`,
        );
      }
      for (const befund of befunde) {
        process.stdout.write(`      ${befund}\n`);
      }
    }
  } finally {
    chrome.kill();
    await warte(200);
    await rm(profil, { recursive: true, force: true });
  }
} finally {
  dev.kill();
  await warte(500);

  // **Ein Aufräumen, das scheitert, hält den Lauf nicht auf** — derselbe Satz
  // stand im früheren `abzuege.mjs` und steht in `kuerzellauf.mjs`, und hier
  // fehlte er.
  //
  // **Er verdeckte am 2026-09-07 die eigentliche Meldung**: der Lauf brach in
  // `aufWarten` ab, sprang hierher, und das `EACCES` beim Löschen eines
  // schreibgeschützten Modulcaches war das Einzige, was ankam. Ein Fehler beim
  // Aufräumen ist nie die Nachricht, auf die es ankommt.
  await rm(heim, { recursive: true, force: true }).catch(() => {});
}

if (rot > 0) {
  process.stdout.write(`\n${String(rot)} von ${String(faelle.length)} Formen brechen den Vertrag\n`);
  process.exit(1);
}

process.stdout.write(`\nalle ${String(faelle.length)} Antwortformen halten den Vertrag\n`);
