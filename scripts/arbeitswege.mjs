// **Die vier Kernwege, die bis Phase 5.1 niemand gezählt hat** — Werkstück U1
// (#1412).
//
// # Was hier steht und was nicht
//
// Hier steht, **welcher Weg** gemessen wird — Schritt für Schritt, so wie ein
// Mensch ihn geht. Womit, steht in `arbeitsweggriffe.mjs`; der Lauf darüber in
// `arbeitswegmass.mjs`.
//
// **Der Weg „Karte anlegen, füllen, weiter" fehlt hier mit Absicht**: ihn
// zählen `kernwegmass.mjs` und `erstlaufmass.mjs` seit Phase 4.3 und 4.10,
// und eine zweite Zählung daneben wäre eine zweite Wahrheit.
//
// # Jeder Weg legt an, was er braucht — und räumt es weg
//
// Ein Deck braucht eine Karte, ein Profilwechsel ein zweites Profil, eine
// Paketkorrektur ein Paket, das niemandem gehört. Alles davon entsteht über
// die Brücke (**keine** Handlung) und verschwindet im `aufraeumen`, auch nach
// einem Fehlschlag. Der Messbestand trägt die Prozesskennung im Namen, damit
// zwei Läufe zugleich nicht dasselbe messen (Befund 3 des Nachlaufs über
// Phase 4.3).
//
// # Die Saat ist der Ausgangsbestand
//
// Gemessen wird gegen ein verlegtes Zuhause aus `sammlungssaat.mjs` — 10.000
// Exemplare, ein Nutzer, ein Profil, die sieben Pakete der Saat und **kein**
// Set. Das Spiel heißt dort `probe`, der Stil `probe-schlicht`.

import { feldUnter, knopfBenannt, knopfMit, reiterMit } from "./arbeitsweggriffe.mjs";

const kennung = String(process.pid);

// hauptZeigt sagt, ob der Text im Hauptbereich auf ein Muster passt.
function hauptZeigt(muster) {
  return `${muster}.test(document.querySelector('main')?.textContent ?? '')`;
}

// ── Paket korrigieren ────────────────────────────────────────────────────
//
// ⌘5 → Paket wählen → Datei aufschlagen → in den Text → tippen → Speichern.
export const paket = {
  schluessel: "paket",
  name: "Paket korrigieren",
  erwartet: { handlungen: 5 },

  async vorbereiten({ rufe }, ws) {
    const antwort = await rufe(ws, "PaketAnlegen", "mse-style", `messlauf-${kennung}`, "Messlauf", "probe");

    if (antwort.fehler) {
      throw new Error(`Paket anlegen: ${antwort.fehler.meldung}`);
    }

    return { name: antwort.paket?.name ?? `probe-messlauf-${kennung}`, pfad: antwort.paket?.pfad ?? "" };
  },

  async fahren(g, ws, stand) {
    await g.kommando(ws, "5", "Digit5", 53);

    // **Die Datei steht im Baum** (U7): kein Auswahlfeld, keine Dateiliste —
    // der Zweig des Messpakets trägt seinen »style«-Knopf, und der ist die
    // erste Handlung nach dem Kürzel.
    const dateiImZweig = `(() => {
      const zweig = [...document.querySelectorAll('nav li')]
        .find((li) => li.querySelector('button')?.textContent.includes(${JSON.stringify(stand.name)}));
      return zweig && [...zweig.querySelectorAll('ul button')].find((b) => b.textContent.trim() === 'style');
    })()`;
    const ladezeit = await g.bis(ws, `!!${dateiImZweig}`, "der Baum mit der Datei »style« des Messpakets");

    // **Der Baum ist eine Tabstelle** (#1290, wie die Paketliste): gezählt wird
    // sein Halt — der Knopf mit `tabIndex 0` —, von dem aus die Pfeile zur
    // Datei führen. Die Datei selbst trägt −1, solange sie nicht der Halt ist.
    //
    // **`nav[data-baum]` und nicht `nav`** — angestrichen beim Messen am
    // 2026-09-10: seit die Seitenleiste sich einklappen lässt (#1490) trägt
    // **sie** ein `nav`, und ihr Halt kam im Baum zuerst. Die Zahl war −1,
    // weil eine eingeklappte Leiste `hidden` ist und niemand sie mit Tab
    // erreicht — eine Aussage über die Leiste, nicht über den Kernweg (§10).
    const tabstellen = await g.tabstelle(ws, `document.querySelector('nav[data-baum] [tabindex="0"]')`);

    await g.klick(ws, dateiImZweig, "die Datei »style«");
    await g.bis(ws, `document.querySelector('textarea[aria-label="Inhalt"]')?.value.length > 0`, "der Inhalt der Datei");

    await g.insFeld(ws, `document.querySelector('textarea[aria-label="Inhalt"]')`, "die Textfläche");
    await g.tippe(ws, `# Messlauf ${kennung}\n`);
    await g.klick(ws, knopfMit("Speichern"), "der Knopf »Speichern«");
    await g.bis(ws, hauptZeigt("/Gespeichert\\./"), "die Bestätigung »Gespeichert.«");

    return { tabstellen, ladezeit };
  },

  async aufraeumen({ rufe }, ws, stand) {
    await rufe(ws, "PaketEntfernen", "mse-style", stand.name);
  },
};

// ── Sammlung suchen ──────────────────────────────────────────────────────
//
// ⌘6 → in die Suche → tippen. Die Ladezeit ist die der Seite mit 10.000
// Exemplaren, die Suchzeit steht daneben.
export const sammlung = {
  schluessel: "sammlung",
  name: "Sammlung suchen",
  erwartet: { handlungen: 3 },

  async vorbereiten() {
    return {};
  },

  async fahren(g, ws) {
    await g.kommando(ws, "6", "Digit6", 54);
    const ladezeit = await g.bis(ws, hauptZeigt("/\\d+ von \\d+/"), "die Sammlung mit ihrer Zählung");

    const suche = `document.querySelector('input[type="search"]')`;
    const tabstellen = await g.tabstelle(ws, suche);

    await g.insFeld(ws, suche, "das Suchfeld");
    const gesamt = await g.zaehleExemplare(ws);

    // **Die Zählzeile ist „N von 10000", gefolgt vom ersten Kartennamen** —
    // ohne Leerzeichen dazwischen. „0 von 10000 gewählt" (die Auswahl) und
    // „200 von 10000 gezeichnet" (die Kappung) tragen eines; die Zeile, die
    // die Suche verändert, ist die ohne.
    await g.tippe(ws, "Dragon 4");
    const suchzeit = await g.bis(
      ws,
      `(() => { const t = (document.querySelector('main')?.textContent ?? '').match(/(\\d+) von (\\d+)[^\\d\\s]/); return t !== null && Number(t[1]) < ${String(gesamt)}; })()`,
      "das Ergebnis der Suche",
    );

    return { tabstellen, ladezeit, zusatz: `Suche ${String(suchzeit)} ms` };
  },

  async aufraeumen() {},
};

// ── Deck bauen ───────────────────────────────────────────────────────────
//
// ⌘7 → »Deck anlegen« aufklappen → Name → tippen → Spiel wählen → Anlegen →
// Set → Karte → ⏎.
//
// **Neun statt elf** — der Zweitblick über Phase 5.1 hat gezeigt, dass der
// Weg trotz U10 noch bei elf stand (Befund 6). Drei Handlungen sind weg, eine
// ist dazugekommen:
//
//   - **weg**: das Spielfeld wird gewählt statt angeklickt, markiert und neu
//     getippt (zwei), und der Umweg über »Öffnen« entfällt, weil das neue
//     Deck von selbst aufgeht (eine);
//   - **dazu**: der Kasten »Deck anlegen« wird aufgeklappt (eine). Er steht
//     nur offen, solange es **kein** Deck gibt; das Messzuhause hat eines.
//
// **Das Ziel aus U1 ist sechs, und die drei übrigen Handlungen sind benannt**:
//
//   1. **Der Kasten.** Er kostet hier eine Handlung und spart 625 Punkte
//      Seitenhöhe — ohne ihn stand die Kartensuche unter der Falz. Wer ihn
//      wegnimmt, tauscht das eine gegen das andere.
//   2. **Die Wahl des Spiels.** Sie fiele weg, wenn das Messset das einzige
//      Spiel des Profils wäre — die Vorbelegung nimmt das erste Set.
//   3. **Die Wahl des Sets.** Sie fiele weg, wenn die Kartenwahl das Set des
//      Arbeitskontexts vorbelegte — den kennt das Fenster heute nicht
//      (`useArbeitskontext` trägt Nutzer, Profil und Rolle, keinen offenen
//      Bestand).
//
// Alles drei steht als Zahl da und nicht als Absicht.
export const deck = {
  schluessel: "deck",
  name: "Deck bauen",
  erwartet: { handlungen: 9 },

  async vorbereiten({ rufe }, ws) {
    const set = await rufe(ws, "SetAnlegen", `Messlauf Deck ${kennung}`, "probe", "probe-schlicht");
    const setID = String(set.setID ?? set.id ?? "");

    if (setID === "") {
      throw new Error(`kein Set angelegt: ${JSON.stringify(set)}`);
    }

    const karte = await rufe(ws, "KarteAnlegen", setID);
    const karteID = String(karte.karteID ?? karte.karte?.id ?? karte.id ?? "");

    if (karteID === "") {
      throw new Error(`keine Karte angelegt: ${JSON.stringify(karte)}`);
    }

    await rufe(ws, "FeldSetzen", setID, karteID, "name", "Messkarte");

    return { setID, karteID, setname: `Messlauf Deck ${kennung}`, deckname: `Messdeck ${kennung}` };
  },

  async fahren(g, ws, stand) {
    await g.kommando(ws, "7", "Digit7", 55);
    const ladezeit = await g.bis(ws, `${knopfMit("Deck anlegen")} !== undefined`, "der Kasten der Deckanlage");

    // **Der Kasten »Deck anlegen« ist zu**, solange es ein Deck gibt (U10).
    await g.klick(ws, knopfMit("Deck anlegen"), "der Kasten »Deck anlegen«");

    const name = feldUnter("Name");
    await g.bis(ws, `${name} !== null`, "das Namensfeld");
    const tabstellen = await g.tabstelle(ws, name);

    await g.insFeld(ws, name, "das Namensfeld");
    await g.tippe(ws, stand.deckname);
    // **Das Spiel wird gewählt, nicht getippt** (Befund 6): eine Handlung
    // statt dreier. Die Liste kommt aus den Sets des Profils; das Messset
    // gehört zum Spiel `probe`, und eine Karte daraus geht nur in ein Deck
    // desselben Spiels.
    // **Erst die Wahlmöglichkeit, dann die Wahl** — dieselbe Falle wie bei
    // der Kartenwahl: die Spiele kommen aus `Sets()`, und der Kasten fragt
    // erst, wenn er aufgeklappt ist.
    await g.bis(
      ws,
      `[...document.querySelectorAll('select option')].some((o) => o.value === 'probe')`,
      "das Spiel »probe« in der Wahl",
    );
    await g.waehle(ws, feldUnter("Spiel"), "probe", "das Spiel »probe«");
    await g.klick(ws, knopfMit("Anlegen"), "der Knopf »Anlegen«");

    // **Das neue Deck geht von selbst auf** (U10, Befund 3 des Zweitblicks):
    // `Deckanlage` schlägt es nach dem Anlegen auf, und der Umweg über
    // »Öffnen« in der Liste entfällt.
    await g.bis(ws, `${feldUnter("Set")} !== null`, "die Kartenwahl des offenen Decks");

    // **Erst die Wahlmöglichkeit, dann die Wahl**: das Feld steht vor seinen
    // Sets da, und mit sichtbarem Navigator (eingeklappte Seitenleiste, #1490)
    // war die Reihenfolge knapp genug, dass der Lauf ins Leere wählte.
    await g.bis(ws, `[...document.querySelectorAll('select option')].some((o) => o.value === ${JSON.stringify(stand.setID)})`, "das Messset in der Wahl");
    await g.waehle(ws, feldUnter("Set"), stand.setID, "das Messset");
    await g.bis(ws, `[...document.querySelectorAll('select option')].some((o) => o.value === ${JSON.stringify(stand.karteID)})`, "die Messkarte in der Wahl");
    await g.waehle(ws, feldUnter("Karte"), stand.karteID, "die Messkarte");
    // **Die Menge steht auf 4** — die Vorgabe der Kartenwahl, ein Spielsatz.
    // Wer sie nicht anfasst, legt vier; der Weg zählt das nicht als Handlung.
    //
    // **⏎ legt** (U10, Befund 3): der Weg über den Knopf »Ins Deck legen«
    // steht weiter da, kostet aber eine Handlung mehr.
    await g.taste(ws, feldUnter("Karte"), "Enter", "⏎ in der Kartenwahl");
    // **Zwischen der Menge und dem Namen steht jetzt Text** — die zwei
    // Zeichenknöpfe tragen seit dem Zweitblick einen Tooltip, und der steht
    // auch verborgen im Baum (`hidden` nimmt ihn nicht aus `textContent`).
    await g.bis(ws, hauptZeigt("/4\\+Eine mehr\\s*Messkarte\\s*Hauptdeck\\s*Herausnehmen/"), "der Platz »4 × Messkarte« im offenen Deck");

    return { tabstellen, ladezeit };
  },

  async aufraeumen({ rufe }, ws, stand) {
    const decks = await rufe(ws, "Decks", "probe");
    for (const eines of decks.decks ?? []) {
      if (eines.name === stand.deckname) {
        await rufe(ws, "DeckLoeschen", String(eines.id));
      }
    }

    await rufe(ws, "SetLoeschen", stand.setID, stand.setname);
  },
};

// ── Profil wechseln ──────────────────────────────────────────────────────
//
// ⌘' → Reiter »Profile« → »Zu … wechseln«. Die Ladezeit ist die des Neuaufbaus
// der ganzen Anwendung, den ein Profilwechsel auslöst (§7, Regel 1).
export const profil = {
  schluessel: "profil",
  name: "Profil wechseln",
  erwartet: { handlungen: 3 },

  async vorbereiten({ rufe }, ws) {
    const alle = await rufe(ws, "Nutzer");
    const erster = (alle.nutzer ?? [])[0];

    if (!erster) {
      throw new Error("kein Nutzer im Bestand");
    }

    const vorher = await rufe(ws, "GezeigtesProfil");
    const neu = await rufe(ws, "ProfilAnlegen", String(erster.id), `Messlauf ${kennung}`);
    const profilID = String(neu.profil?.id ?? "");

    if (profilID === "") {
      throw new Error(`kein Profil angelegt: ${JSON.stringify(neu)}`);
    }

    return {
      nutzername: erster.name,
      profilname: `Messlauf ${kennung}`,
      profilID,
      vorherID: String(vorher.profil?.id ?? ""),
    };
  },

  async fahren(g, ws, stand) {
    await g.kommando(ws, "'", "Quote", 222);
    const ladezeit = await g.bis(ws, `${reiterMit("Profile")} !== undefined`, "die Seite »Nutzer und Profile«");

    await g.klick(ws, reiterMit("Profile"), "der Reiter »Profile«");

    const knopf = knopfBenannt(`Zu ${stand.nutzername} · ${stand.profilname} wechseln`);
    await g.bis(ws, `${knopf} !== undefined`, "der Wechselknopf");
    const tabstellen = await g.tabstelle(ws, knopf);

    await g.klick(ws, knopf, "der Wechselknopf");
    const wechselzeit = await g.bis(
      ws,
      `(async () => { const p = await window.go.app.Bindungen.GezeigtesProfil(); return String(p.profil?.id) === ${JSON.stringify(stand.profilID)}; })()`,
      "das neue Profil als gezeigtes",
    );
    await g.bisRuhe(ws);

    return { tabstellen, ladezeit, zusatz: `Wechsel ${String(wechselzeit)} ms` };
  },

  async aufraeumen({ rufe }, ws, stand) {
    if (stand.vorherID !== "") {
      await rufe(ws, "ProfilWechseln", stand.vorherID, "");
    }

    await rufe(ws, "ProfilLoeschen", stand.profilID, stand.profilname);
  },
};

export const arbeitswege = [paket, sammlung, deck, profil];
