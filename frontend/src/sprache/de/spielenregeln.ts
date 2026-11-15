// **Die Wörter des Regelbaums** — PLAY-02
// (#1710).
//
// # Warum die Bereiche hier Namen bekommen
//
// **Weil bis heute der Schlüssel dastand.** Die Regelschau zeigte
// `zustandsbasiert` und `automatisch` — die Kennungen aus dem Kern, ungefragt
// und unübersetzt. Derselbe Befund wie beim Stand einer Partie in der Liste
// (2026-09-17): zwei Schreibweisen für dieselbe Sache, und eine davon ist
// keine Sprache.
//
// **Ein Schlüssel, den das Fenster nicht kennt, bleibt stehen** — dafür gibt
// es `regelbereichwort`; eine Fläche, die einen unbekannten Bereich
// verschwiege, verschwiege eine Regel.
export const regelwoerter = {
  // **Die zwei Äste** (AK-02).
  regelgruppeAblauf: 'Ablauf einer Partie',
  regelgruppeMechanik: 'Mechanik einzelner Karten',

  // **Die dreizehn Bereiche.**
  regelbereichZonen: 'Zonen',
  regelbereichZugstruktur: 'Zugstruktur',
  regelbereichPrioritaet: 'Priorität',
  regelbereichStapel: 'Stapel',
  regelbereichCombat: 'Kampf',
  regelbereichZustandsbasiert: 'Was von selbst geschieht',
  regelbereichTypen: 'Typzeile',
  regelbereichMarken: 'Marken',
  regelbereichAnhaengsel: 'Auren und Ausrüstung',
  regelbereichSpielsteine: 'Spielsteine',
  regelbereichMana: 'Mana',
  regelbereichWechselwirkungen: 'Verrechnen und Kopieren',
  regelbereichCommander: 'Commander',

  // **Die vier Grade** — und jeder sagt, wer es tut.
  //
  // **„Automatisch“ mit seinen Grenzen** (PLAY-01.AK-02, #1709): jeder
  // automatische Bereich nennt Grenzen, und das Wort zeigt auf sie, statt
  // einen ganzen Bereich zuzusagen.
  regelgradAutomatisch:
    'Das Programm setzt es durch — bis auf die Grenzen darunter',
  regelgradBegleitet: 'Das Programm warnt',
  regelgradManuell: 'Ihr macht es selbst',
  regelgradUnbekannt: 'Niemand hat es geprüft',

  // **Die zwei Bedienmodi** (AK-01).
  modusTitel: 'Wie viel soll das Programm übernehmen?',
  modusAnfaenger: 'So viel wie möglich',
  modusFrei: 'Wir wählen selbst',
  modusErklaerung:
    '„So viel wie möglich“ ist ein Bedien- und Unterstützungsmodus und keine Zusage über die vollständigen Turnierregeln. Was dieses Programm wirklich durchsetzt, steht darunter — samt dem, was es nicht tut.',

  // **Die Abwahl** (AK-02).
  regelAbwaehlen: '{bereich} selbst übernehmen',
  regelWidersprueche: 'Das hängt zusammen',
  regelWiderspruch: '{haengt} braucht {bereich} — {quelle}',
  regelSelbst: 'Ihr macht es selbst',
}
