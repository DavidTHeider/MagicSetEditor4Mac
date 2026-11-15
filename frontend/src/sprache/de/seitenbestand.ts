// Die Texte der **Bestandspflege** — sichern und zurückspielen
// (#940).
//
// Aus `seiten.ts` herausgelöst, wie die Kartenseite vor ihr: die Datei stand
// bei 396 Zeilen, und die Grenze aus §2 liegt bei 400. Ein Katalog darf sich
// auf mehrere Dateien verteilen, solange er als **einer** ankommt — die Zahl
// der Kataloge bleibt fünf, und daran hängt der Vertragstest gegen Go (§9).

export const bestandspflege = {
  titel: 'Bestand',
  hinweis:
    'Alles, was diese Anwendung führt: Sets, Karten, Feldwerte, der Verlauf ' +
    'und die eingesetzten Bilder. Eine Sicherung trägt beides in einer ' +
    'Datei — den Bestand und die Bilder.',
  sichern: 'Bestand sichern …',
  waehlen: 'Sicherung zurückspielen …',
  rueckfrage: 'Diese Sicherung zurückspielen?',
  datei: 'Datei',
  angelegtAm: 'Bestand angelegt am',
  sets: 'Sets',
  karten: 'Karten',
  unbekannt: '—',
  warnung:
    'Der gegenwärtige Bestand wird ersetzt. Er geht nicht verloren: er ' +
    'bleibt als bestand-vor-rueckspielung.db daneben liegen — bis zum ' +
    'nächsten Zurückspielen.',
  bilderLage: 'Eingesetzte Bilder',
  bilderErsetzt:
    'Diese Sicherung trägt Bilder. Sie treten an die Stelle der vorhandenen; ' +
    'die bisherigen bleiben als beiwerk-vorher daneben liegen.',
  bilderGeleert:
    'Diese Sicherung trägt keine Bilder — sie stammt aus einem Stand, in dem ' +
    'keine eingesetzt waren. Die vorhandenen werden entfernt; sie bleiben als ' +
    'beiwerk-vorher daneben liegen.',
  bilderUnberuehrt:
    'Diese Sicherung ist eine einzelne Datenbank aus 0.1.0 oder 0.2.0 und ' +
    'trägt keine Bilder. Die vorhandenen bleiben unverändert stehen — auf ' +
    'einem anderen Rechner fehlen sie dann ganz.',
  bestaetigen: 'Zurückspielen',
  abbrechen: 'Abbrechen',

  // **Das Kennwort einer Sicherung** (OPS-06, #1723).
  kennwort: 'Kennwort (freiwillig)',
  kennwortWiederholen: 'Kennwort wiederholen',
  kennwortHinweis:
    'Mit Kennwort ist die Sicherung verschlüsselt und lässt sich auch auf einem anderen Rechner öffnen. Ein vergessenes Kennwort lässt sich nicht wiederherstellen — die Sicherung ist dann verloren.',
  kennwortVerschieden: 'Die beiden Kennwörter stimmen nicht überein.',
  klartext: 'Ohne Kennwort sichern — die Datei liegt dann im Klartext',
  klartextHinweis:
    'Der Bestand ist verschlüsselt. Ohne Kennwort läge die Sicherung ungeschützt auf dem Medium, auf das sie geht.',
  gesperrt: 'Diese Sicherung ist verschlüsselt.',
  gesperrtKennwort: 'Kennwort der Sicherung',
  entsperren: 'Öffnen',
  verschluesselt: 'Verschlüsselt',
  ja: 'ja',
  nein: 'nein',
  paketeEnthalten: 'Eigene Pakete darin',
  paketeFehlen: 'Neu zu beziehen',
  keine: 'keine',
  zurueckgespielt:
    'Zurückgespielt. Der bisherige Bestand liegt als {beiseite}.',
  rueckkehrUnveraendert: 'Dein bisheriger Bestand ist unverändert.',
  rueckkehrUnvollstaendig:
    'Auch der Weg zurück scheiterte: Teile des bisherigen Bestands liegen als bestand-vor-rueckspielung.db und beiwerk-vorher daneben. Spiele eine Sicherung zurück oder lege sie von Hand zurück.',
  wiederherstellungTitel: 'Der Bestand ließ sich nicht öffnen',
  wiederherstellungHinweis:
    'Die Anwendung arbeitet erst wieder, wenn eine brauchbare ' +
    'Bestandsdatei dasteht. Eine Sicherung zurückzuspielen ist der Weg ' +
    'dorthin — er ersetzt die Datei, an der es gerade scheitert.',
  wiederherstellungGrund: 'Was gemeldet wurde',
  // **Wenn die Datei vermutlich heil ist** — F05 (#2344): belegt, voll,
  // gesperrt, nur lesbar oder ein Fehler des Datenträgers.
  wiederherstellungHinweisHeil:
    'Die Bestandsdatei ist vermutlich in Ordnung — es liegt an etwas um sie herum. Behebe es, wie unten beschrieben, und versuche es noch einmal; ersetzen musst du nichts.',
  erneutVersuchen: 'Erneut versuchen',
  weiterGescheitert:
    'Erneut versucht — der Bestand ließ sich weiterhin nicht öffnen.',

  // **Der Weg zurück, benannt** — Werkstück 5 von Phase 4.13 (#1352).
  //
  // „Spiele eine Sicherung zurück“ ist kein Rat, solange dreißig Dateien mit
  // Fassungsnummern im Namen nebeneinanderliegen und nur eine davon aufgeht.
  // Welche das ist, hat Go nachgesehen — hier steht sie mit dem, woran man
  // sie wiedererkennt.
  wegZurueckTitel: 'Der Weg zurück',
  wegZurueck:
    'Zurück kommst du über »{datei}« im Ordner deines Bestands — {sets} ' +
    'Set(s), {karten} Karten, Schemastand {stand}. Wähle sie unten aus.',

  // **Die Verschlüsselung** — Werkstück 14 von Phase 4.12, Vermerk 0072.
  //
  // **Was sie nicht kann, steht neben dem, was sie kann.** Der Vermerk sagt
  // es ausdrücklich: die Chiffre ist deterministisch, und wer mehrere
  // Sicherungen nebeneinanderlegt, sieht, welche Blöcke sich geändert haben.
  // Ein Versprechen ohne seine Grenze ist eine Falle.
  chiffreTitel: 'Verschlüsselung des Bestands',
  chiffreAnHinweis:
    'Der Bestand liegt verschlüsselt auf der Platte. Ohne den Schlüssel ' +
    'lässt er sich nicht öffnen — auch nicht von dir.',
  chiffreAusHinweis:
    'Der Bestand liegt im Klartext auf der Platte. Wer die Datei hat, kann ' +
    'alles darin lesen: Namen, Geburtsdatum, Anschrift, Fotos und die ' +
    'Anmeldewege.',
  chiffreAnschalten: 'Verschlüsselung einschalten …',
  chiffreAusschalten: 'Verschlüsselung ausschalten …',
  chiffreFrageAn: 'Den Bestand jetzt verschlüsseln?',
  chiffreFrageAus: 'Die Verschlüsselung wieder ausschalten?',
  chiffreWarnungAn:
    'Der Schlüssel entsteht dabei und geht in den Schlüsselbund dieses ' +
    'Rechners. Er lässt sich nicht neu berechnen und steht nirgendwo sonst: ' +
    'ist er fort, sind die Daten fort. Sorge vorher für eine Sicherung.',
  chiffreWarnungAus:
    'Danach liegt wieder alles im Klartext, und der Schlüssel wird aus dem ' +
    'Schlüsselbund entfernt.',
  chiffreGrenzen:
    'Die Verschlüsselung schützt **nicht** gegen Fälschung. Wer mehrere ' +
    'Abzüge derselben Datei vergleicht, sieht, welche Stellen sich ' +
    'geändert haben. **Eine Sicherung bekommt ihr eigenes Kennwort** — oben ' +
    'beim Sichern; ohne Kennwort läge sie im Klartext, und das verlangt dann ' +
    'ein ausdrückliches „ohne Kennwort sichern“.',
  chiffreLaeuft: 'Der Bestand wird umgeschrieben …',

  // **Was Platz belegt** — Werkstück 4 von Phase 4.13
  // (#1352).
  //
  // **Am 2026-09-07 gemessen**: neben einem Bestand von 127 MB lagen dreißig
  // Sicherungen vor Migrationen, zusammen über zwei Gigabyte — und nichts
  // sagte es.
  platzUeberschrift: 'Platz auf der Platte',
  // **Nicht „die jüngste ist der Rückweg“** (Werkstück 5 und 6 von Phase
  // 4.13, #1352): nach zwei Aktualisierungen ist das die falsche.
  platzErklaerung:
    'Was diese Anwendung ablegt und wie viel es wiegt — der ganze Ordner, denn was keinem Posten gehört, zählt unter **Sonstiges**. **Sicherungen vor Migrationen** entstehen bei jedem Start, der den Bestand umbaut; stehen bleibt der **Weg zurück** — die jüngste Sicherung, die diese Fassung öffnen kann. Vom **Protokoll** bleibt die laufende Datei.',
  platzDateien: '{anzahl} Datei(en)',
  platzFreigeben: 'Ältere wegräumen',
  platzFreigebenFuer: '{posten}: die älteren wegräumen',
  platzFreigebenFrage: '{posten}: die älteren wirklich wegräumen?',
  platzFreigebenErklaerung:
    'Die älteren Dateien werden gelöscht und lassen sich nicht zurückholen. Was als Weg zurück gebraucht wird, bleibt stehen.',
  platzLoeschenJa: 'Wegräumen',
  platzNamen: {
    bestand: 'Bestand',
    bilder: 'Eingesetzte Bilder',
    pakete: 'Bezogene Pakete',
    vorsicherungen: 'Sicherungen vor Migrationen',
    protokoll: 'Protokoll',

    // **Der Rest, und er hat einen Namen** — Werkstück 6 von Phase 4.13
    // (#1352). Ohne ihn zeigte diese Seite 1,5 von 13 Gigabyte an und sähe
    // dabei vollständig aus.
    sonstiges: 'Sonstiges im Anwendungsordner',
  },

  // **Wann zuletzt gesichert wurde, und der freiwillige Plan** — OPS01
  // (#2347).
  alter: {
    ueberschrift: 'Letzte Sicherung',
    unbekannt: 'unbekannt',
    unbekanntSatz:
      'Über diese Fassung wurde der Bestand noch nie gesichert. Unbekannt heißt unbekannt — hier steht nie „heute“, ohne dass eine Datei zurückgelesen wurde.',
    heute: 'heute',
    gestern: 'vor einem Tag',
    vorTagen: 'vor {tage} Tagen',
    alt: '{alter} · älter als 14 Tage',
    unstimmig: 'Zeit nach der Uhr',
    satz: 'Zuletzt gesichert am {zeitpunkt} über {flaeche} in {datei}, zurückgelesen und geprüft.',
    satzPlan:
      'Zuletzt gesichert am {zeitpunkt} nach Plan in {datei}, zurückgelesen und geprüft.',
    unstimmigSatz:
      'Die letzte Sicherung trägt den {zeitpunkt} — eine Zeit nach der Uhr dieses Rechners. Ein Alter lässt sich nicht sagen; sichere, um sicherzugehen.',
    flaechen: {
      fenster: 'das Fenster',
      cli: 'die Kommandozeile',
      mcp: 'den Werkzeugserver',
      rest: 'den Dienst',
    },
    planSchalter: 'Wöchentlich in einen Ordner sichern, sieben behalten',
    planHilfe:
      'Nur nach Ordnerwahl. Gelöscht werden nur eigene Sicherungen, nie fremde Dateien. Gesichert wird, solange das Fenster offen ist.',
    planOrdner: 'Ordner: {ordner}',
    planZuletzt: 'Zuletzt nach Plan am {zeitpunkt}.',
    planNie: 'Nach Plan noch nie gesichert.',
    planFehler: 'Der Plan konnte zuletzt nicht sichern: {grund}',
  },
} as const
