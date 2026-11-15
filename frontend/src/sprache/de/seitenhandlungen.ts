// Die Texte der **Handlungsleiste** —
// #1623.
//
// **Ein eigener Baustein**, wie `seitenkarten.ts` und aus demselben Grund:
// `seitenspielen.ts` steht an der Zeilengrenze aus §2.
//
// **Die Schlüssel sind die Befehlsarten des Kerns** — ein Name, den man am
// Tisch sagen würde, und kein übersetzter Bezeichner. Was ein Befehl
// bewirkt, steht in `docs/REGELABDECKUNG.md`; hier steht, wie er heißt.
export const handlungen = {
  handlungenKarte: 'Mit dieser Karte',
  handlungenZone: 'Mit einer Zone',
  handlungenSitz: 'Du selbst',
  handlungenKampf: 'Kampf',
  handlungenZug: 'Der Zug',

  handlungenLeer:
    'Gerade ist nichts zu stellen. Wer nicht an der Reihe ist, darf nur, was keine Priorität braucht.',
  handlungenKeineKarte: 'Wähle eine Karte, um zu sehen, was mit ihr geht.',

  // **Warum ein Knopf fehlt** — Artboard *Tisch*, unter den Handlungen.
  handlungenNurErlaubt:
    'Angeboten wird nur, was gerade geht — was fehlt, lässt der Tisch jetzt nicht zu.',

  // Die Namen der Befehle — je Art einer.
  handlungsname: {
    'objekt.tappen': 'Tappen',
    'objekt.enttappen': 'Enttappen',
    'objekt.umdrehen': 'Umdrehen',
    'objekt.schieben': 'Schieben nach',
    'objekt.ziehen': 'Legen nach',
    'objekt.zaehler': 'Marken',
    'objekt.anhaengen': 'Anhängen an',
    // **Der Haufen** — CR 400.5, eine Anordnung und keine Regelbewegung
    // (#1756).
    'objekt.haeufen': 'Auf einen Haufen legen',
    'objekt.enthaeufen': 'Aus dem Haufen nehmen',
    'objekt.loesen': 'Lösen',
    'objekt.schaden': 'Schaden markieren',
    'objekt.mana': 'Für Mana tappen',
    'objekt.land': 'Als Land spielen',
    'objekt.wirken': 'Wirken',
    'objekt.zerstoeren': 'Zerstören',
    'objekt.opfern': 'Opfern',
    'objekt.exilieren': 'Exilieren',
    'objekt.neutralisieren': 'Neutralisieren',
    'objekt.kaempfen': 'Kämpfen gegen',
    'objekt.beherrschen': 'Kontrolle geben an',
    'objekt.umwandeln': 'Umwandeln',
    'objekt.kopieren': 'Kopieren von',
    'spieler.legende': 'Als Legende behalten',
    'spieler.reihung': 'Fähigkeit als Nächste legen',
    'spieler.heimholen': 'In die Kommandozone',
    'zug.abwerfen': 'Abwerfen',
    'zone.blick': 'Hineinsehen',
    'zone.mischen': 'Mischen',
    'zone.mahlen': 'Mahlen',
    'zone.hellsehen': 'Hellsehen',
    'zone.ueberwachen': 'Überwachen',
    'spieler.leben': 'Leben',
    'spieler.vergiftung': 'Vergiftungsmarken',
    'spieler.mana': 'Mana nachtragen',
    'spieler.muenzwurf': 'Münze werfen',
    'spieler.wuerfeln': 'Würfeln',
    'spieler.monarch': 'Monarch werden',
    'spieler.initiative': 'Initiative nehmen',
    'spieler.strahlung': 'Radioaktivitätsmarken',
    'spieler.spielstein': 'Spielstein erzeugen',
    'spieler.emblem': 'Emblem erzeugen',
    'kampf.angreifer': 'Angreifer erklären',
    'kampf.blocker': 'Blocker erklären',
    'zug.passen': 'Passen',
    'zug.abkuerzen': 'Abkürzen',
    'zug.beenden': 'Zug beenden',
    'zug.aufgeben': 'Aufgeben',
    'spieler.zuruecknehmen': 'Zurücknehmen',
    // **Der Beginn** — CR 103.5 (#1839).
    'spieler.starthand': 'Starthand ziehen',
    'spieler.mulligan': 'Mulligan',
    'spieler.unterlegen': 'Unter die Bibliothek legen',
    'spieler.behalten': 'Hand behalten',
    'sitz.platznehmen': 'Platz nehmen',
  },

  // Die Namen der eingebauten ausgelösten Fähigkeiten — CR 725.2, CR 726.2
  // und CR 728.1 (M9, #1611).
  //
  // **Sie beschreiben, was geschieht, und nicht, wie die Regel heißt.** Wer
  // am Tisch ordnet, entscheidet zwischen zwei Wirkungen; „monarch.wechsel“
  // wäre die Kennung des Kerns und keine Auskunft.
  ausloesername: {
    'monarch.zug': 'Der Monarch zieht eine Karte',
    'monarch.wechsel': 'Der Angreifer wird Monarch',
    'initiative.wechsel': 'Der Angreifer nimmt die Initiative',
    strahlung: 'Radioaktivitätsmarken mahlen',
  },

  // Die Beschriftungen der Felder — je Nutzlastschlüssel eine.
  handlungsfeld: {
    an: 'An',
    art: 'Art',
    felder: 'Eigenschaften',
    gegen: 'Gegen',
    kennung: 'Kennung',
    menge: 'Anzahl',
    ruf: 'Ruf',
    seiten: 'Seiten',
    stelle: 'Stelle',
    sitz: 'Sitz',
    um: 'Um',
    wer: 'Wer',
    wie: 'Wie viele',
    wuerfel: 'Würfel',
    ziele: 'Ziele',
    vorlage: 'Vorlage',
    zone: 'Zone',
    angriffe: 'Wer greift wen an',
    bloecke: 'Wer blockt wen',
  },

  handlungStellen: 'Stellen',
  handlungKeineWahl: 'nichts',
  handlungRufKopf: 'Kopf',
  handlungRufZahl: 'Zahl',
  handlungRufOhne: 'ohne Ruf',

  // Der Spielstein bringt seine Eigenschaften mit — CR 701.7a.
  spielsteinName: 'Kartenname',
  spielsteinTyp: 'Typzeile',
  spielsteinStaerke: 'Stärke',
  spielsteinWiderstand: 'Widerstandskraft',
  spielsteinKennung: 'Kennung',

  // Der Kampf.
  kampfKeineKreatur:
    'Auf deiner Seite steht keine Kreatur, die angreifen könnte.',
  kampfKeinAngreifer: 'Es greift niemand an.',
  kampfGreiftAn: 'greift an',
  kampfBlockt: 'blockt',
}
