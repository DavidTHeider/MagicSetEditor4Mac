// **Die Vorbelegung des Fensters für Tests** — aus `api.ts` herausgelöst,
// weil jene mit der Datenpflege (O8, #1432) über die Zeilengrenze aus §2 ging.
//
// **Der Schnitt folgt der Frage** (§2): dort steht, **wie** eine Fassade im
// Test entsteht und was ein unbelegter Aufruf tut; hier, **was ohne Zutun
// belegt ist** — die Wege, die jede Seite beim Aufschlagen holt.
import type { Api, EinstellungDTO } from '../api'
import { standMit } from './sicherungsstand'

export function belegeFenster(sprache = ''): Partial<Api> {
  return {
    Einstellungen: belegeEinstellungen(sprache),
    // **Der Schnellfilter antwortet leer** (U5): eine Seite, die ihn ruft,
    // ohne dass ein Test ihn belegt, sieht keine Treffer statt eines Absturzes.
    KartenFiltern: () => Promise.resolve({ kartenIDs: [] }),
    // **Der Verlauf antwortet leer** (U6): der Anhang der Kartenseite zeigt
    // ihn immer, und ohne Belegung stünde neben jeder geprüften Meldung eine
    // zweite — `findByRole('alert')` wurde mehrdeutig, gefunden in der
    // Pipeline von #1417 (`karteanlegen`, `kartennotizenseite`).
    Verlauf: () => Promise.resolve({ schritte: [], fehler: null }),
    SpracheSetzen: () => Promise.resolve({ fehler: null }),
    // **Und der Stand der Darstellung** (UXR-F14, #2029): der Raum um die
    // Anwendung meldet beim Einhängen, ohne dass ein Test es erwähnt.
    DarstellungsstandMelden: () => Promise.resolve({ fehler: null }),
    // **Und das gewählte Objekt** (GESTE, #2358) — aus demselben Grund.
    ObjekthandlungenMelden: () => Promise.resolve({ fehler: null }),

    // **Der Verschlüsselungsstand gehört dazu, seit die Einstellungsseite ihn
    // liest** — Werkstück 14 von Phase 4.12.
    //
    // Ohne ihn zeigte jeder Lauf über diese Seite eine zweite Meldung („Die
    // Methode Verschluesselungsstand ist nicht belegt"), und zwei
    // `role="alert"` machen `findByRole('alert')` mehrdeutig. Genau so ist es
    // beim ersten Anlauf zugeschlagen: zwei Läufe der Bestandspflege wurden
    // rot, und keiner an dem, was er prüft.
    //
    // **Klartext als Vorgabe**, weil das der Stand eines frischen Bestands
    // ist; wer den anderen prüft, belegt ihn selbst.
    Verschluesselungsstand: () =>
      Promise.resolve({ verschluesselt: false, fehler: null }),

    // **Und das Alter der Sicherung** (OPS01, #2347) — aus demselben Grund:
    // die Fläche „Letzte Sicherung" liest es beim Aufschlagen, und ohne
    // Beleg stünde neben jeder geprüften Meldung eine zweite. Beim ersten
    // Anlauf hat es vier Läufe über Bestand, Datenpflege und Rückkehr rot
    // gemacht, und keinen an dem, was er prüft. **Unbekannt als Vorgabe**:
    // der Stand eines frischen Bestands.
    Sicherungsstand: () =>
      Promise.resolve({ stand: standMit(), planfehler: null, fehler: null }),

    // **Und die drei lesenden Wege der Rückkehr** (O7, #1431) — aus genau
    // demselben Grund wie der Verschlüsselungsstand darüber: das Gebiet
    // „Bestand" holt sie beim Aufschlagen, und ohne Beleg stünde neben jeder
    // geprüften Meldung eine zweite. Beim ersten Anlauf hat es zwei Läufe der
    // Bestandspflege rot gemacht, und keinen an dem, was er prüft.
    //
    // **Leer als Vorgabe**, weil das der Stand eines frischen Bestands ist;
    // wer eine Rückkehr prüft, belegt sie selbst.
    Ruecksprungarten: () => Promise.resolve({ arten: [], fehler: null }),

    // **Die Pflegevorschau wird nicht vorbelegt** — sie läuft **nicht** beim
    // Aufschlagen, sondern erst auf Knopfdruck. Wer sie prüft, belegt sie
    // selbst; wer es nicht tut, drückt den Knopf auch nicht.
    Rueckkehren: () => Promise.resolve({ rueckkehren: [], fehler: null }),
    Ruecksprunglage: () =>
      Promise.resolve({
        datenstand: '',
        zeilen: [],
        unklar: [],
        handpruefungNoetig: false,
        fehler: null,
      }),

    // **Wer arbeitet** — Werkstück 17 von Phase 4.12. Die Nutzerseite fragt
    // im selben Zug wie nach der Liste; ohne Beleg zeigte jeder Lauf über sie
    // eine zweite Meldung.
    WerArbeitet: () => Promise.resolve({ nutzerID: '1', fehler: null }),

    // **Die Zugriffstoken fragen beim Aufbau nach den Ablaufwahlen** —
    // Werkstück O2 von Phase 5.2 (#1426); ohne Beleg meldete jeder Lauf über
    // ein technisches Profil eine unbelegte Methode.
    Tokenablaufwahlen: () =>
      Promise.resolve({
        wahlen: [
          { name: 'tag', ablauf: '2026-09-09T12:00:00Z' },
          { name: 'monat', ablauf: '2026-10-08T12:00:00Z' },
          { name: 'jahr', ablauf: '2027-09-08T12:00:00Z' },
          { name: 'nie', ablauf: '' },
        ],
        fehler: null,
      }),
    // **Die Kontextleiste fragt beim Aufbau nach Nutzern und Bildern** —
    // Werkstück U4 von Phase 5.1 (#1415); ohne Beleg meldete jeder Lauf über
    // die Anwendung zwei unbelegte Methoden.
    Nutzer: () => Promise.resolve({ nutzer: [], fehler: null }),
    Nutzerbilder: () =>
      Promise.resolve({ bilder: {}, badges: {}, fehler: null }),

    // **Und der Arbeitskontext** — Werkstück O1 von Phase 5.2, an der Stelle
    // der Leseart aus Werkstück 16 von Phase 4.12. Der Streifen über der
    // Anwendung fragt ihn beim Aufbau; ohne Beleg meldete jeder Lauf über die
    // Anwendung eine unbelegte Methode.
    Arbeitskontext: () =>
      Promise.resolve({
        nutzerID: '1',
        profilID: '1',
        eignerID: '1',
        rolle: 'verwalter',
        nurLesend: false,
        modus: 'fenster',
        revision: 1,
        fehler: null,
      }),
  }
}

// belegeEinstellungen gibt eine Einstellungsantwort für die Vorrichtung.
//
// **Seit #245
// braucht `App` sie**, um überhaupt zu zeichnen: die Sprache kommt aus dem
// Bestand, und gezeichnet wird erst, wenn sie feststeht.
//
// Die Werte sind die Vorgabewerte aus `internal/domain`; wer eine andere
// Sprache braucht, gibt sie mit.
export function belegeEinstellungen(sprache = ''): () => Promise<{
  einstellungen: EinstellungDTO[]
}> {
  return () =>
    Promise.resolve({
      einstellungen: [
        {
          schluessel: 'sprache',
          wert: sprache,
          wirksam: sprache,
          werte: ['de', 'en'],
          bereich: 'anwendung',
        },
        {
          schluessel: 'erscheinungsbild',
          wert: 'system',
          wirksam: 'system',
          werte: ['system', 'hell', 'dunkel'],
          bereich: 'anwendung',
        },
        {
          schluessel: 'schriftgroesse',
          wert: 'normal',
          wirksam: 'normal',
          werte: ['klein', 'normal', 'gross', 'sehrgross'],
          bereich: 'anwendung',
        },

        // **Und die der Bildausgabe** (#491) — sie steht in derselben Antwort
        // und auf einer anderen Seite.
        {
          schluessel: 'bildkonflikt',
          wert: 'nummerierenneu',
          wirksam: 'nummerierenneu',
          werte: [
            'altbehalten',
            'ueberschreiben',
            'nummerieren',
            'nummerierenneu',
          ],
          bereich: 'bildausfuhr',
        },
      ],
    })
}
