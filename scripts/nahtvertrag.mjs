// **Der Vertrag der Wails-Naht — eine Antwortform je Fall**
// (#1288).
//
// # Warum es diese Datei gibt
//
// `docs/FAEHIGKEITEN.md` sagte es bis zum 2026-08-29 selbst: „**Und das Fenster
// ist nicht dabei.** Die Läufe fahren Kommandozeile, MCP und Dienst; die
// Wails-Naht steht daneben." `internal/flaechen` nennt `Bindungen` genau
// **einmal** — in einem Kommentar.
//
// **Zwei grüne Seiten einer Grenze sind kein Nachweis für die Grenze.** Go-Tests
// sehen die Struct-Marken nicht, Frontend-Tests gehen über `test/api.ts` und
// nie durch die JSON-Serialisierung (§3). Was dazwischen liegt, hat bis zum
// 2026-08-29 niemand betreten.
//
// # Was hier steht
//
// Jeder Fall ist eine **Form**, die eine Antwort annehmen kann — keine
// Bindung. Welche Bindung den Fall auslöst, ist Mittel zum Zweck; was geprüft
// wird, ist die Gestalt, in der sie durch die Brücke kommt. **Wie viele es
// sind, steht nur hier**: die Zahl stand an vier Stellen in Prosa und war an
// dreien falsch (13, 17, 19, als es 20 waren — nachgezählt bei ENG-02).
//
// # Warum die Saat und kein bezogenes Paket
//
// Weil der Lauf sonst von dem abhinge, was auf **diesem** Rechner installiert
// ist (§3). `probe.mse-game` und `probe-schlicht.mse-style` liegen im Bündel;
// der Lauf legt sein Set selbst an und räumt sein Zuhause hinterher weg.
//
// # Was hier **nicht** steht, und warum
//
//   - **Die dreizehn Fähigkeiten am nativen Dateidialog.** Sie öffnen ein
//     modales Fenster und kämen nie zurück. Was sie tun, prüft
//     `internal/app` mit gestopfter Naht — und das ist die Go-Seite.
//
//     **Und „kämen nie zurück" ist wörtlich gemeint**: ein Fall über einen
//     Dialog hängt den Lauf ohne eine Zeile Ausgabe, weil der Dialog aufgeht
//     und auf einen Menschen wartet — am 2026-08-31 mit `SymbolNachzeichnen`
//     (#1340) **fünfundzwanzig Minuten** lang. Die Zusage — `teile` ist `[]`
//     und ein Abbruch kein Fehler — steht in
//     `internal/app/symbolnachzeichnenbindung_test.go`, wo die Naht gestopft
//     ist.
//
//     **Fähigkeiten, nicht Bindungen**: `docs/FAEHIGKEITEN.md` zählt
//     dreizehn Fähigkeiten mit Dialog, und ein Wächter hält die Zahl; die
//     Bindungen darunter sind **vierzehn** (gemessen am 2026-08-31, über die
//     Aufrufkette von `Bindungen.*` nach `dateiwahl`) — `MTGJSONOeffnen` und
//     `MTGJSONVerzeichnisOeffnen` sind zwei Wege in dieselbe Fähigkeit.
//   - **Alles, was ins Netz greift.** Der Lauf hat kein Netz und soll keines
//     brauchen.
//   - **Die 163 Bindungen als Ganzes.** Das zählt der Wächter in
//     `internal/werkzeuge`, und ihre Antworten prüft
//     `internal/app/leerelistenlaufzeit_test.go`. Hier geht es um die
//     **Brücke**, nicht um die Vollständigkeit.

// faelle sind die Vertragsfälle, in der Reihenfolge, in der sie laufen.
//
// **Die Reihenfolge trägt Zustand.** Fall 1 sieht einen leeren Bestand, Fall 2
// legt an. Wer sie umstellt, stellt den Vertrag um — deshalb steht der Zustand
// im Namen und nicht in einem Kommentar.
export const faelle = [
  {
    name: "leere Liste kommt als [] und nicht als null",
    ruf: `B.Sets()`,
    pruefe: { liste: ["sets"], leer: ["sets"], istNull: ["fehler"] },
  },
  {
    name: "Erfolg mit Inhalt — und Unicode überlebt die Brücke",
    ruf: `(async () => {
      const angelegt = await B.SetAnlegen(${JSON.stringify(namemitUnicode())}, "probe", "probe-schlicht");
      return { angelegt, sets: await B.Sets() };
    })()`,
    pruefe: {
      istNull: ["angelegt.fehler"],
      gleich: { "sets.sets.0.name": namemitUnicode(), "sets.sets.0.spiel": "probe" },
      liste: ["sets.sets"],
    },
  },
  {
    name: "Feldfehler nennt sein Feld — und werte ist {} und nicht null",
    ruf: `B.SetAnlegen("ohne Stil", "probe", "")`,
    pruefe: {
      gleich: { "fehler.code": "validierung", "fehler.feld": "stil" },
      objekt: ["fehler.werte"],
    },
  },
  {
    name: "eine unlesbare Kennung gibt leere Liste **und** Fehler",
    ruf: `B.Karten("keinezahl")`,
    pruefe: {
      liste: ["karten"],
      leer: ["karten"],
      gleich: { "fehler.feld": "id" },
      objekt: ["fehler.werte"],
    },
  },
  // **Ein Set ohne Karten und nicht ein Set, das es nicht gibt**: seit #1818
  // antwortet eine Kennung, die es nicht gibt, mit „nicht gefunden" wie eine
  // fremde (Vermerk 0131). Die Zusage, um die es hier geht — ein Erfolg trägt
  // `fehler: null` —, prüft deshalb das Set aus Fall 2, das noch keine Karte
  // hat.
  {
    name: "ein optionales Feld ist null, und das ist die Zusage",
    ruf: `B.Karten("1")`,
    pruefe: { liste: ["karten"], leer: ["karten"], istNull: ["fehler"] },
  },
  {
    name: "eine unbekannte Kennung ist nicht gefunden — und die Liste trotzdem []",
    ruf: `B.Karten("9999")`,
    pruefe: {
      liste: ["karten"],
      leer: ["karten"],
      gleich: { "fehler.code": "nichtGefunden" },
      objekt: ["fehler.werte"],
    },
  },
  {
    name: "eine Karte anlegen und wiederfinden — Antwort nach Zustandswechsel",
    ruf: `(async () => {
      const angelegt = await B.KarteAnlegen("1");
      await B.FeldSetzen("1", angelegt.karteID, "name", ${JSON.stringify(namemitUnicode())});
      return { angelegt, karte: await B.Karte(angelegt.karteID), liste: await B.Karten("1") };
    })()`,
    pruefe: {
      istNull: ["angelegt.fehler"],
      gleich: {
        "karte.karte.werte.0.feld": "name",
        "karte.karte.werte.0.wert": namemitUnicode(),
      },
      liste: ["liste.karten", "karte.karte.werte"],
    },
  },
  {
    name: "ein Konflikt trägt keinen Feldbezug",
    ruf: `B.ErsetzungAusfuehren("1", {
      text: "Ω", ersatz: "A", feld: "",
      grossKleinBeachten: false, ganzesWort: false, stand: "veralteter Stand",
    })`,
    pruefe: {
      gleich: { "fehler.code": "konflikt" },
      objekt: ["fehler.werte"],
      fehlt: ["fehler.feld"],
      liste: ["stellen"],
    },
  },
  {
    name: "zwei Vorschauen zugleich — genau eine wird abgebrochen",
    ruf: `(async () => {
      const beide = await Promise.all([B.Vorschau("1", 1), B.Vorschau("1", 2)]);
      return { codes: beide.map((eine) => (eine.fehler ? eine.fehler.code : "ok")) };
    })()`,
    pruefe: { genauJeEinmal: ["codes", "abgebrochen", "ok"] },
  },
  {
    name: "eine große Antwort kommt ganz an",
    ruf: `(async () => {
      for (let i = 0; i < 200; i += 1) { await B.KarteAnlegen("1"); }
      return await B.Karten("1");
    })()`,
    pruefe: { liste: ["karten"], mindestens: ["karten", 201] },
  },
  {
    name: "eine gelöschte Karte ist weg — auch in der Antwort danach",
    ruf: `(async () => {
      const vorher = await B.Karten("1");
      const geloescht = await B.KarteLoeschen("1", vorher.karten[0].id, false);
      return { geloescht, nachher: await B.Karten("1"), vorher };
    })()`,
    pruefe: {
      istNull: ["geloescht.fehler"],
      kleinerAls: ["nachher.karten", "vorher.karten"],
    },
  },
  {
    // **Dieselbe Regel wie bei der Vorschau, an einer zweiten Bindung.**
    // `Packarten` rechnet über jede Karte; wer das Set wechselt, während sie
    // laufen, will die Antwort des vorigen nicht mehr (§7, Regel 4).
    name: "zwei Packartenläufe zugleich — genau einer wird abgebrochen",
    ruf: `(async () => {
      const beide = await Promise.all([B.Packarten("1"), B.Packarten("1")]);
      return { codes: beide.map((einer) => (einer.fehler ? einer.fehler.code : "ok")) };
    })()`,
    pruefe: { genauJeEinmal: ["codes", "abgebrochen", "ok"] },
  },
  {
    // **Der zweite Aufruf gibt dasselbe.** Das Angebot wird gemerkt; was der
    // Speicher zusagt, ist Gleichheit — und nicht eine Dauer, denn die wäre
    // eine Aussage über diesen Rechner.
    name: "ein gemerktes Angebot ist dasselbe Angebot",
    ruf: `(async () => {
      const erst = await B.Packarten("1");
      const dann = await B.Packarten("1");
      return { gleich: JSON.stringify(erst) === JSON.stringify(dann), erst };
    })()`,
    pruefe: { gleich: { gleich: true }, liste: ["erst.arten"] },
  },
  {
    name: "eine Bindung ohne Argumente antwortet auch",
    ruf: `B.Datenstand()`,
    pruefe: { istNull: ["fehler"], vorhanden: ["stand"] },
  },
  {
    // **Eine flache Antwort mit zwei Listen** (#1337). Sie war einmal um ein
    // `*KartenskriptlaufDTO` herum gebaut, und das kam bei jedem Fehler als
    // `null` an — ein Zweig, den die Seite nie erreicht, weil `auspacken`
    // vorher wirft. Was hier geprüft wird, ist die **Form nach dem Umbau**:
    // beide Listen sind da, auf beiden Wegen.
    //
    // Der Lauf legt zwei Karten an; das ist der letzte Fall, damit die
    // Zählungen davor stehen bleiben.
    name: "ein Kartenskript legt an — beide Listen kommen als [] und nicht null",
    ruf: `(async () => {
      const angebot = await B.Kartenskripte("1");
      const leer = await B.KartenskriptAusfuehren("1", "Leerer Zug");
      const zwei = await B.KartenskriptAusfuehren("1", "Zwei Farben");
      const nichts = await B.KartenskriptAusfuehren("1", "gibt es nicht");
      return { angebot, leer, zwei, nichts };
    })()`,
    pruefe: {
      liste: [
        "angebot.skripte",
        "leer.karten",
        "leer.befunde",
        "zwei.karten",
        "nichts.karten",
        "nichts.befunde",
      ],
      leer: ["leer.karten", "nichts.karten"],
      mindestens: ["angebot.skripte", 2],
      gleich: { "zwei.karten.length": 2, "nichts.fehler.code": "nichtGefunden" },
      istNull: ["angebot.fehler", "zwei.fehler"],
    },
  },
  {
    // **Eine gerechnete Antwort mit einer erklärten Reihenfolge** (#1338).
    // Sie ist die einzige Form im Vertrag, deren Inhalt aus einem **Skript
    // je Karte** kommt — und die einzige, bei der eine Gruppe mit **null**
    // Karten dastehen muss: das ist die Zusage der erklärten Reihenfolge.
    //
    // Der Fall läuft nach dem Kartenskript, damit es Karten mit einer Farbe
    // gibt; die Zahlen darunter zählen deshalb dessen zwei.
    name: "eine Auszählung gibt die erklärte Reihenfolge — samt leerer Gruppe",
    ruf: `(async () => {
      const angebot = await B.Auszaehldimensionen("1");
      const gezaehlt = await B.Auszaehlen("1", "farbe");
      const nichts = await B.Auszaehlen("1", "gibt es nicht");
      return { angebot, gezaehlt, nichts };
    })()`,
    pruefe: {
      liste: [
        "angebot.dimensionen",
        "gezaehlt.gruppen",
        "gezaehlt.befunde",
        "nichts.gruppen",
        "nichts.befunde",
      ],
      leer: ["nichts.gruppen"],
      gleich: {
        "angebot.dimensionen.0.name": "farbe",
        "gezaehlt.gruppen.0.name": "blau",
        "gezaehlt.gruppen.1.name": "rot",
        "nichts.fehler.code": "nichtGefunden",
      },
      istNull: ["angebot.fehler", "gezaehlt.fehler"],
    },
  },
  {
    // **Zwei Antworten aus einem Zustand, den dieser Fall selbst herstellt**
    // (#1342). Er legt ein Schlüsselwort an, setzt sein Muster und fragt
    // dann nach den Hilfen — die Verweise hängen daran, was im Muster steht,
    // und ohne den Schreibschritt gäbe es keinen einzigen.
    //
    // **Und der zweite Ruf zeigt, dass sie mitwandern**: nach einem zweiten
    // Parameter im Muster sind es zwei Verweise. Ein Vertrag, der nur den
    // ersten Stand prüfte, ließe eine gemerkte Antwort durchgehen.
    name: "die Parameterhilfen folgen dem Muster",
    ruf: `(async () => {
      const angelegt = await B.SchluesselwortAnlegen("1");
      const eins = await B.SchluesselwortSetzen("1", angelegt.wortID,
        "muster", "Kicker <atom-param>farbwort</atom-param>");
      const erst = await B.Parameterhilfen("1", angelegt.wortID);
      await B.SchluesselwortSetzen("1", angelegt.wortID,
        "muster", "Kicker <atom-param>farbwort</atom-param> <atom-param>zahl</atom-param>");
      const dann = await B.Parameterhilfen("1", angelegt.wortID);
      return { angelegt, eins, erst, dann };
    })()`,
    pruefe: {
      liste: ["erst.arten", "erst.verweise", "erst.befunde", "dann.verweise"],
      istNull: ["erst.fehler", "dann.fehler"],
      gleich: {
        "erst.arten.length": 2,
        "erst.verweise.length": 2,
        "dann.verweise.length": 3,
        "erst.verweise.0.einfuegen": "{param1}",
        "erst.verweise.1.einfuegen": "({param1})",
        "dann.verweise.2.einfuegen": "{param2}",
      },
    },
  },
  {
    // **Die eine Form, die nur die Verdrahtung prüfen kann** (#1341).
    //
    // Eine eigene Autoersetzungsregel liegt im Bestand; ob sie beim Tippen
    // greift, entscheidet eine **Zeile im Composition Root** —
    // `a.MitAutoersetzung(ersetzung)` in `werkzeugnaehte.go`. Ohne sie fragte
    // `AutoErsetzen` weiter den Stilwähler und sähe nur die Regeln des
    // Pakets.
    //
    // **Und keine Schicht für sich merkt das.** Die Gegenprobe hat es
    // gezeigt: die Zeile entfernt, und `internal/flaechen` blieb grün — dort
    // wird nicht getippt. Hier bricht es.
    // **Decks über die laufende Brücke** (#1293). Zwei grüne Seiten einer
    // Grenze sind kein Nachweis für die Grenze — und diese Antwort hat drei
    // Fallen: eine leere Liste, eine Kennung als **Text** und ein `formatID`,
    // das leer ist statt "0".
    name: "ein Deck kommt mit Textkennungen und leerer Platzliste",
    ruf: `(async () => {
      const angelegt = await B.DeckAnlegen("Aggro", "magic", "");
      const gelesen = await B.Deck(angelegt.deck.id);
      const liste = await B.Decks("magic");
      return { angelegt, gelesen, liste };
    })()`,
    pruefe: {
      istNull: ["angelegt.fehler", "gelesen.fehler", "liste.fehler"],
      liste: ["gelesen.deck.plaetze", "liste.decks"],
      gleich: {
        "angelegt.deck.spiel": "magic",
        "gelesen.deck.formatID": "",
        "gelesen.deck.id": "1",
      },
    },
  },
  {
    // **Und ein Format mit seinen vier Zahlen.** Eine Null bedeutet hier drei
    // verschiedene Dinge, und ein `omitempty` auf einem `int` ließe sie
    // verschwinden — genau der Fall aus §3.
    name: "ein Format trägt seine Nullen über die Brücke",
    ruf: `(async () => {
      const bauarten = await B.Bauarten();
      const angelegt = await B.FormatAusBauart("commander", "C", "magic", "commander");
      return { bauarten, angelegt };
    })()`,
    pruefe: {
      istNull: ["angelegt.fehler"],
      liste: ["bauarten.bauarten"],
      gleich: {
        "angelegt.format.mindestgroesse": 100,
        "angelegt.format.hoechstgroesse": 100,
        "angelegt.format.sideboardgroesse": 0,
        "angelegt.format.hoechstzahlGleich": 1,
      },
    },
  },
  {
    name: "eine eigene Autoersetzung greift beim Tippen",
    ruf: `(async () => {
      const gesetzt = await B.AutoersetzungSetzen("1", "(W)", "weiss", true, true);
      const getippt = await B.AutoErsetzen("1", "Text (W)", 8);
      const aus = await B.AutoersetzungSchalten("1", false);
      const danach = await B.AutoErsetzen("1", "Text (W)", 8);
      await B.AutoersetzungenZuruecksetzen("1");
      return { gesetzt, getippt, aus, danach };
    })()`,
    pruefe: {
      istNull: ["gesetzt.fehler", "getippt.fehler", "danach.fehler"],
      gleich: {
        "getippt.text": "Text weiss",
        "danach.text": "Text (W)",
      },
    },
  },
];

// namemitUnicode ist der Name, an dem die Kodierung geprüft wird.
//
// **Drei Klassen in einem Wort**: ein griechischer Buchstabe (zwei Byte in
// UTF-8), ein Zeichen jenseits der Grundebene (vier Byte, in JavaScript ein
// Ersatzpaar) und ein Umlaut. Wer nur ASCII prüfte, prüfte die Brücke nicht.
export function namemitUnicode() {
  return "Naht-Ω-Prüfung 🜂";
}
