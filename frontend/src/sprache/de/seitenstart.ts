// **Die Startseite** — herausgelöst, weil `seiten.ts` mit ihr über 400 Zeilen
// ginge (§2).
//
// **Ein Katalog darf sich auf mehrere Dateien verteilen**, solange er als
// **einer** ankommt — so wie `seitenkarten.ts`. Was gleich bleibt, ist die
// Zahl der Kataloge; an ihr hängt der Vertragstest gegen die Go-Seite (§9).
export const start = {
  ueberblick: {
    titel: 'Dein Überblick',
    // **Die Probleme auf den Kacheln** (AP-8, #2365, B-13) — eine Zahl nur,
    // wenn jedes Set aktuell geprüft ist; sonst wie viele es sind.
    mitProblemen: '{anzahl} mit Problemen',
    geprueftVon: '{geprueft} von {gesamt} geprüft',
    mitBefund: '{anzahl} mit Befund',
    nichtGeprueft: 'nicht vollständig geprüft',
    problemeGeprueft:
      'Probleme geprüft {wann} · {geprueft} von {gesamt} Karten',
    problemeOffen: 'Probleme: {offen} von {gesamt} Sets nicht geprüft',
    pruefen: 'Jetzt prüfen',
    setsMitProblemen: '{sets} mit Problemen',
    kartenMitBefund: '{karten} mit Befund in {sets}',
    aeltesterPreis: 'ältester Preis {wann}',
    ohnePreis: '{anzahl} ohne Preis',
    mitFehlenden: '{anzahl} mit fehlenden Karten',
    decksMitFehlenden: '{decks} mit fehlenden Karten',
    alter: {
      gerade: 'gerade eben',
      minuten: 'vor {minuten} min',
      stunden: 'vor {stunden} h',
      tage: 'vor {tage} Tagen',
      tag: 'vor einem Tag',
    },
  },
  // **Weiterarbeiten** — Werkstück U11 von Phase 5.1
  // (#1422).
  // Die Schlüssel der Arten und Vorgänge kommen aus
  // `internal/domain/arbeitsverlauf.go`.
  // **Aufmerksamkeit und Hintergrundaufgaben** (#1422).
  aufmerksamkeit: {
    titel: 'Was Aufmerksamkeit braucht',
    nichts: 'Nichts — {sets}, {decks} und {pakete} durchgesehen.',
    durchgesehen: '{sets}, {decks} und {pakete} durchgesehen.',
    arten: {
      setVorbehalte: '{anzahl} nicht gedeutet',
      deckFehlend: '{anzahl} fehlen in der Sammlung',
      paketBefund: '{anzahl} im Skript',
    },
    // **Alle Pakete in einer Zeile** (AP-8, #2365, B-14) — `pakete` ist
    // die gezählte Form von „eigenes Paket“.
    paketeMitBefund: '{pakete} mit Befund im Skript',
    gruendeZeigen: 'Gründe zeigen',
    gruendeVerbergen: 'Gründe verbergen',
    // **Ein Set mit Befunden** (AP-8, Teil 2b) — `karten` gezählt.
    setBefunde: '{karten} mit Befund',
    befunde: 'Nicht durchgesehen:',
    alle: 'Alle {anzahl} zeigen',
    weniger: 'Weniger zeigen',
    // **Gründe mit eigenem Satz** (#1963): `vorbehalt` ist Datei und Zeile
    // des ersten Vorbehalts eines Sets. Die Gründe eines Pakets kommen aus
    // der Paketprüfung (`fehler.skript`, `seiten.vorlagen`); `amOrt` setzt
    // die Datei davor.
    gruende: {
      vorbehalt: '{0}, Zeile {1}',
    },
    amOrt: '{ort}: {satz}',
    // **Das Betriebsband** (OPS02, #2348): die Arten aus
    // `internal/domain/betriebsband.go`; `frei` kommt als Gigabyte an.
    betrieb: {
      arten: {
        platzKnapp: 'Nur noch {frei} frei neben dem Bestand',
        sicherungAlt: 'Letzte Sicherung vor {tage} Tagen',
        dienstBeendet: 'Der Dienst beim Anmelden endete mit Code {ausgang}',
        webprozessNeu: 'Die Webansicht wurde {anzahl}-mal neu geladen',
      },
      rat: {
        platzKnapp:
          'Unter rund 2 GB oder 5 % frei wird jede Sicherung knapp — räume in „Daten und Sicherung“ auf.',
        sicherungAlt:
          'Älter als 14 Tage — sichere den Bestand in „Daten und Sicherung“.',
        dienstBeendet:
          'Nach einem Fehler startet launchd ihn neu — warum er endete, steht in der Diagnose.',
        webprozessNeu:
          'WebKit hatte ihren Prozess beendet; was in einem offenen Formular nicht gespeichert war, kann fehlen — prüfe es in der Diagnose.',
      },
    },
  },
  auftraege: {
    titel: 'Hintergrundaufgaben',
    staende: {
      laeuft: 'läuft',
      fertig: 'fertig',
      gescheitert: 'gescheitert',
      abgebrochen: 'abgebrochen',
    },
    arten: {
      preise: 'Preise für die Sammlung holen',
      sammeldaten: 'Sammeldaten von Scryfall laden',
      fassung: 'Eine neue Fassung holen',
      bogen: 'Druckbogen legen',
      set: 'Set einlesen',
      mtgjson: 'MTG JSON einlesen',
      vorlage: 'Über eine Exportvorlage ausgeben',
      installer: 'Installer einlesen',
      weitergabe: 'Als Installer weitergeben',
      probleme: 'Sets auf Probleme prüfen',
    },
    abbrechen: '{art} abbrechen',
    dauer: '{sekunden} s',
    laufend: { eine: 'Aufgabe läuft', viele: 'Aufgaben laufen' },
  },
  weiterarbeiten: {
    titel: 'Weiterarbeiten',
    weiterAn: 'Weiter an {art} „{name}“',
    zuletztWann: 'Zuletzt {vorgang} am {zeit}',
    // **Die Marke einer Zeile, die oben steht** — seit UXR-03 (#1911) kein
    // Listentitel mehr, sondern das Wort neben „Set · ausgegeben“.
    angeheftet: 'angeheftet',
    angeheftetLeer:
      'Nichts angeheftet — was oben bleiben soll, heftest du an einer Zeile an.',
    leer: 'Hier steht, woran du zuletzt gearbeitet hast. Noch ist es leer — der erste Schritt ist ein Set oder ein Vorlagenpaket.',
    handlungSets: 'Ein Set anlegen oder öffnen',
    handlungPakete: 'Ein Paket beziehen',
    anheften: 'Anheften',
    loesen: 'Lösen',
    anheftenVon: '„{name}“ anheften',
    loesenVon: '„{name}“ lösen',
    arten: { set: 'Set', deck: 'Deck', paket: 'Paket' },
    vorgaenge: {
      geoeffnet: 'geöffnet',
      eingefuehrt: 'eingeführt',
      ausgegeben: 'ausgegeben',
    },
  },
  titel: 'Start',
  ueberschrift: 'MagicSetEditor4Mac',
  was: 'Karten für Magic und andere Spiele entwerfen, sichten und drucken — mit den Vorlagen, die Magic Set Editor 2 benutzt.',
  ersteSchritte: 'Erste Schritte',

  // **Die Texte des Pop-ups** — #1325. Bis dahin waren die ersten Schritte
  // eine **Hälfte der Startseite**, die verschwand, sobald der Bestand ein
  // Set führte; jetzt sind sie eine Einführung, die man wiederholen kann.
  einfuehrung: {
    willkommen:
      'Drei Schritte, und die erste Karte steht. Du kannst dieses Fenster jederzeit über die Einstellungen wieder aufrufen.',
    wiederzeigen: 'Beim Start wieder zeigen',
    schliessen: 'Los geht’s',
    bezeichnung: 'Einführung',
  },
  schritte: {
    setOeffnen: {
      titel: 'Ein Set öffnen',
      erklaerung:
        'Eine vorhandene .mse-set-Datei aufschlagen — oder ein Set aus MTG JSON übernehmen.',
      handlung: 'Zu den Sets',
    },
    kartenBearbeiten: {
      titel: 'Karten bearbeiten',
      erklaerung:
        'Die Karten des offenen Sets ansehen und ändern; die Vorschau zeigt jede Änderung sofort.',
      handlung: 'Zu den Karten',

      // **Der Schritt bleibt stehen, auch wenn er noch nicht geht** (#673).
      // Ihn auszublenden nähme dem Anfänger die Landkarte — und weil es der
      // mittlere ist, sah es aus, als fehlte ein anderer.
      ersteinSet: 'Dafür muss zuerst ein Set offen sein.',
      stattdessen: 'Erst ein Set öffnen',
    },
    paketeHolen: {
      titel: 'Vorlagen beziehen',
      erklaerung:
        'Acht Vorlagenpakete sind schon da. Wer mehr Spiele oder Stile will, holt sie hier.',
      handlung: 'Zu den Paketen',
    },
  },
}
