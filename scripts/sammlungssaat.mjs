// **Fünftausend Exemplare in einen eigenen Bestand legen** — Werkstück 1 von
// Phase 4.5 (#1292).
//
// # Warum ein eigenes Zuhause und nicht der echte Bestand
//
// Weil eine Messung, die fünftausend Zeilen in die Sammlung der nutzenden
// Person schreibt, kein Test ist, sondern ein Eingriff. Am 2026-08-30 sind in
// dieser Phase schon einmal drei Probepakete im **echten** Paketverzeichnis
// gelandet, weil eine Vorrichtung das Zuhause nicht verlegt hat.
//
// Verlegt wird `HOME`, und der Bestand liegt darunter am Vorgabeort — dort
// sucht ihn auch ein Fenster, das mit demselben Zuhause startet, und mit ihm
// Pakete und Bilder (`bestandsort.go`, QA-10.NAT-01).
//
// ```bash
// node scripts/sammlungssaat.mjs                    # legt /tmp/mse45heim an
// HOME=/tmp/mse45heim make dev                      # in einem eigenen Fenster
// node scripts/sammlungsmass.mjs
// ```
//
// # Warum über `mse` und nicht über SQL
//
// Weil der Bestand **migriert** sein muss, und die Migrationen liegen in Go.
// Ein `CREATE TABLE` daneben wäre eine zweite Wahrheit über das Schema, und
// die liefe auseinander. Der erste Ruf legt die Datei an und migriert sie; die
// Zeilen kommen danach über `sqlite3`, weil fünftausend einzelne Aufrufe von
// `mse exemplar` fünftausend Prozesse wären.
//
// **Das ist genau die Lücke, die Werkstück 2 schließt** (Import aus einer
// Liste). Bis dahin ist dieses Skript der Behelf, und es sagt es selbst.

import { execFileSync } from "node:child_process";
import { mkdirSync, rmSync } from "node:fs";
import { join } from "node:path";

const heim = process.argv[2] ?? "/tmp/mse45heim";
const anzahl = Number(process.argv[3] ?? 5000);

const stand = join(heim, "Library", "Application Support", "MagicSetEditor4Mac");
const bestand = join(stand, "bestand.db");

rmSync(heim, { recursive: true, force: true });
mkdirSync(stand, { recursive: true });

// **Erst `mse`, damit migriert wird.** Ohne diesen Ruf gäbe es keine Tabelle.
//
// # Warum gebaut und nicht `go run`
//
// Weil `go run` mit verlegtem `HOME` seinen Modul-Cache dort sucht — und ihn
// dann anlegt, samt Go-Werkzeugkette. Beim ersten Anlauf lagen 300 MB unter
// `/tmp/mse45heim/go`, schreibgeschützt, und das Aufräumen scheiterte.
//
// **Gebaut wird in den Cache des Projekts** (`scripts/activate.sh`), gefahren
// wird mit dem verlegten Zuhause.
const werkzeug = join(heim, "mse");

execFileSync("go", ["build", "-o", werkzeug, "./cmd/mse"], { stdio: "inherit" });

execFileSync(werkzeug, ["sammlung", "--json"], {
  env: { ...process.env, HOME: heim, MSE_BESTAND: bestand },
  stdio: "ignore",
});

// **Jedes Exemplar gehört einem Profil** — seit Phase 4.12
// (#1351)
// trägt `sammlung.profil_id` ein `NOT NULL`, und der erste Ruf von `mse` legt
// den ersten Nutzer samt Profil an. Gefunden am 2026-09-08, als die Saat für
// Phase 5.1 (#1412)
// zehntausendmal `NOT NULL constraint failed` gab und **trotzdem** mit dem
// Satz „10000 Exemplare" hätte enden können — `sqlite3` gibt bei Zeilenfehlern
// den Rest weiter. Deshalb zählt der Lauf unten nach und weist ab.
const profilID = execFileSync("sqlite3", [bestand, "SELECT id FROM profile ORDER BY id LIMIT 1"])
  .toString()
  .trim();

if (profilID === "") {
  throw new Error(`kein Profil in ${bestand} — hat \`mse sammlung\` den Bestand angelegt?`);
}

const namen = [
  "Lightning Bolt",
  "Counterspell",
  "Dark Ritual",
  "Giant Growth",
  "Healing Salve",
  "Shivan Dragon",
  "Serra Angel",
  "Llanowar Elves",
  "Wrath of God",
  "Birds of Paradise",
];

// **Eine Zeile je Druck**, damit die Eindeutigkeit über Druck, Zustand,
// Sprache und Ausführung nicht greift.
const zeilen = [];
for (let i = 0; i < anzahl; i += 1) {
  const kennung = `00000000-0000-4000-8000-${String(i).padStart(12, "0")}`;
  const name = `${namen[i % namen.length]} ${String(i)}`;
  zeilen.push(
    `INSERT INTO sammlung (printing, menge, zustand, sprache, foil, ` +
      `erwerbsdatum, erwerbspreis, erwerbswaehrung, notiz, fassung, name, ` +
      `setcode, nummer, profil_id) VALUES ('${kennung}', ${String(1 + (i % 4))}, ` +
      `'nearMint', 'en', ${String(i % 7 === 0 ? 1 : 0)}, '', 0, '', '', 1, ` +
      `'${name}', 'p${String(i % 50).padStart(2, "0")}', '${String(i % 250)}', ${profilID});`,
  );
}

execFileSync("sqlite3", [bestand], {
  input: `BEGIN;\n${zeilen.join("\n")}\nCOMMIT;\n`,
});

// **Und ein Set mit fünf Karten** — seit Phase 5.1
// (#1412).
//
// Die Kernwege „Deck bauen" und die Laschen des Serienmaßes brauchen ein
// offenes Set; die Saat der Anwendung bringt in einem Entwicklungsbau keines
// mit (`build/bin/…/Resources/saat/` trägt sieben Pakete und kein Set). Über
// `mse`, aus demselben Grund wie oben: der Weg ist derselbe wie im Fenster.
//
// **Die Namen sind erfunden**, damit keine echte Karte darin liegt (§11).
const kommandozeile = (...args) =>
  execFileSync(werkzeug, [...args, "--json"], {
    env: { ...process.env, HOME: heim, MSE_BESTAND: bestand },
  }).toString();

const angelegt = JSON.parse(kommandozeile("set-neu", "Messset 5.1", "probe", "probe-schlicht"));
const setID = String(angelegt.setID ?? angelegt.id ?? "");

if (setID === "") {
  throw new Error(`kein Messset angelegt: ${JSON.stringify(angelegt)}`);
}

for (const name of ["Silberdistel", "Nachtfalter", "Steinwacht", "Flussgeist", "Glutkern"]) {
  const karte = JSON.parse(kommandozeile("karte", setID, "anlegen"));
  const karteID = String(karte.karteID ?? karte.id ?? "");

  if (karteID === "") {
    throw new Error(`keine Karte angelegt: ${JSON.stringify(karte)}`);
  }

  kommandozeile("feld", setID, karteID, "name", name);
  kommandozeile("feld", setID, karteID, "text", `Ein Probetext für ${name}, der über zwei Zeilen geht.`);
}

const gezaehlt = execFileSync("sqlite3", [bestand, "SELECT count(*) FROM sammlung"])
  .toString()
  .trim();

// **Der Lauf weist ab, was er nicht gelegt hat.** Ein Messbestand, der
// weniger Zeilen trägt als bestellt, misst etwas anderes als behauptet (§10).
if (Number(gezaehlt) !== anzahl) {
  throw new Error(`${gezaehlt} Exemplare statt ${String(anzahl)} — die Saat ist nicht aufgegangen`);
}

process.stdout.write(
  `${gezaehlt} Exemplare und das Messset ${setID} mit fünf Karten in ${bestand}\n` +
    `Jetzt:  HOME=${heim} make dev\n` +
    `Dann:   node scripts/sammlungsmass.mjs\n`,
);
