// Die Hilfeinhalte.
//
// **Bis #679
// las diesen Katalog niemand.** Er hatte neun Zeilen und keinen Leser — der
// einzige der fünf, den keine Seite benutzte.
//
// # Was hier steht
//
// **Die Begriffe, die aus MSE kommen und die niemand mitbringt.** Stilbogen,
// Symbolschrift, Exportvorlage, Vorbehalt: wer sie nicht kennt, findet in der
// Anwendung nichts, was sie erklärt.
//
// **Nicht die Dokumentation von MSE.** Was aus einem Paket kommt — Feldnamen,
// Stilnamen — bleibt unübersetzt (§9); hier stehen die Begriffe der
// **Anwendung**, je in einem Satz.
import { betrieb } from './hilfebetrieb'
import { suche } from './hilfesuche'

export const hilfe = {
  betrieb,
  suche,
  // **Eine Hilfe je Seite** — #806.
  //
  // Die Schlüssel sind die aus `navigation.ts`, und ein Wächter hält beide
  // Richtungen gegeneinander (`seitenhilfe.test.ts`). Eine zu kurze Aufzählung
  // wäre sonst grün.
  //
  // **Je ein Satz darüber, wofür die Seite da ist, und einer über den nächsten
  // Schritt.** Nicht, was die Bedienelemente heißen — das steht daneben und
  // veraltet mit jedem Umbau.
  seiten: {
    start: {
      ueberschrift: 'Start',
      text: 'Der Einstieg. Oben die Zahlen des Bestands — Sets, Karten, Pakete, Exemplare —, darunter, woran du zuletzt gearbeitet hast: geöffnet, eingeführt, ausgegeben, und was du oben festhältst. Dann, was Aufmerksamkeit braucht — Sets mit Vorbehalten, Decks mit Fehlkarten, Pakete mit Skriptbefunden — und was gerade im Hintergrund geholt wird. Die drei ersten Schritte stehen beim ersten Start als Fenster davor; die Führungen in der Hilfe zeigen sie wieder.',
    },
    sets: {
      ueberschrift: 'Sets',
      text: 'Alle Sets, die diese Anwendung kennt. Von hier aus wird eines geöffnet, ein neues angelegt oder eine `.mse-set`-Datei eingelesen. Das offene Set entscheidet, was seine Laschen und die Schubladen „Eigenschaften“ und „Werkzeuge“ zeigen.',
    },
    pakete: {
      ueberschrift: 'Pakete',
      text: 'Spiele, Stilbögen, Symbolschriften und Exportvorlagen. Sie werden hier bezogen — an einem festen Stand und mit Prüfsumme — und eingebunden. Ohne sie hat eine Karte keine Felder und kein Aussehen.',
    },
    vorlagen: {
      ueberschrift: 'Vorlagen',
      text: 'Die Textdateien in einem Vorlagenpaket — Felder, Skripte, Maße. Hier lassen sie sich ändern, ohne einen Texteditor daneben. Eine Änderung gilt für jedes Set, das dieses Paket benutzt; der vorige Stand bleibt als Sicherung neben dem Paket liegen.',
    },
    sammlung: {
      ueberschrift: 'Sammlung',
      text: 'Was du wirklich besitzt — je Druck, Zustand, Sprache und Ausführung eine Zeile. Sie hat mit den Sets nichts zu tun: eine entworfene Karte ist ein Entwurf, ein Exemplar ein Stück Papier. Preise werden auf Anweisung geholt und sind Schätzwerte, für die niemand einsteht.',
    },
    decks: {
      ueberschrift: 'Decks',
      text: 'Zusammenstellungen aus eigenen Karten und echten Drucken. Ein Deck verweist auf Karten, statt sie zu kopieren, und gehört zu einem Spiel — ohne das gäbe es kein Format. Ein Format ist hier ein Datensatz und keine eingebaute Liste: es wird angelegt, nicht mitgeliefert. Die Prüfung fragt für echte Drucke bei Scryfall; eine selbst entworfene Karte bleibt ungeprüft, und das heißt nicht unzulässig.',
    },
    turniere: {
      ueberschrift: 'Turniere',
      text: 'Ein Turnier ist ein Verlauf über die Zeit: Runden, Paarungen, Ergebnisse. Die Rangliste wird daraus gerechnet und nirgends abgelegt — wer ein Ergebnis berichtigt, sieht sofort den neuen Stand. Gepaart wird nach dem Schweizer System; die vier Tiebreaker und ihre Zahlen stehen in den Magic Tournament Rules. Wer ein Deck anmeldet, reicht eine Liste ein: sie wird abgezogen, und eine spätere Änderung am Deck erreicht sie nicht mehr.',
    },
    spielen: {
      ueberschrift: 'Spielen',
      text: 'Eine Partie ist ein Tisch mit eingefrorenen Decks: was beim Anfangen aus einem Deck wird, ändert sich nicht mehr, wenn jemand sein Deck später bearbeitet. Was am Tisch geschieht, steht als Strom von Ereignissen fest — jedes hängt am vorigen, und keines wird überschrieben. Gezeigt wird immer die Sicht **eines** Sitzes: eine Bibliothek ist auch für ihren Eigner verdeckt, und wer sie durchsieht, schuldet danach ein Mischen. Der Wirt hält den Vollzustand; das steht auf der Seite, weil es vor dem Mitspielen zu wissen ist.',
    },
    handeln: {
      ueberschrift: 'Handeln',
      text: 'Gehandelt wird mit einem Gegenüber, dessen Fingerabdruck du verglichen hast — ohne Konto und ohne Vermittler. Der Abgleich rechnet auf **diesem** Rechner: hinaus geht dabei nichts, und was ein Gegenüber von dir sieht, entscheidest du unter „Einstellungen › Verbindungen“. Verbucht wird erst, wenn die Karten wirklich den Besitzer gewechselt haben; dein Bestand ändert sich dann, und das Journal hält fest, was wohin ging. **Bei dir verbucht ist nicht beidseitig verbucht** — zwei Bestände auf zwei Rechnern sind keine gemeinsame Buchung, und der Stand sagt es. Was deine Sammlung wert ist, steht daneben als Beobachtung und nicht als Zusage.',
    },
    hilfe: {
      ueberschrift: 'Hilfe',
      text: 'Die Begriffe, die aus Magic Set Editor stammen und die niemand mitbringt, je in einem Satz. Dazu die Führungen, die Tastenkürzel, warum manche Aufschriften englisch bleiben und was diese Anwendung nach draußen gibt.',
    },
    nutzer: {
      ueberschrift: 'Nutzer und Profile',
      text: 'Wem gehört, was hier liegt. Ein Nutzer ist die Person; ein Profil eine Arbeitsumgebung dieser Person — eigene Sets, eigene Sammlung, eigene Einstellungen. Wer für ein Turnier getrennt sammeln will, braucht kein zweites Konto, sondern ein zweites Profil.',
    },
    einstellungen: {
      ueberschrift: 'Einstellungen',
      text: 'Sprache, Erscheinungsbild und die Orte, an denen Pakete gesucht werden. Was hier steht, gilt für das ganze Programm und nicht für ein einzelnes Set.',
    },
    karten: {
      ueberschrift: 'Karten',
      text: 'Die Karten des offenen Sets: links die Liste, in der Mitte die Vorschau, rechts die Felder der aufgeschlagenen Karte. Eine Änderung an einem Feld ist sofort zu sehen; bestätigt werden muss nichts.',
    },
    setangaben: {
      ueberschrift: 'Setangaben',
      text: 'Was für das ganze Set gilt: Name, Beschreibung, Setsymbol, Urheber. Diese Angaben stehen jeder Karte zur Verfügung — ein Stilbogen kann das Symbol auf jede Karte zeichnen.',
    },
    stil: {
      ueberschrift: 'Stil',
      text: 'Der Stilbogen des Sets und seine Optionen. Der Bogen sagt, wie eine Karte aussieht; die Optionen stellen ihn ein — etwa auf einen anderen Rahmen oder eine andere Auflage.',
    },
    suche: {
      ueberschrift: 'Suche',
      text: 'Karten über ihre Feldwerte finden, quer durch alle Sets oder in einem. Gesucht wird im Volltext, und die Treffer führen unmittelbar zur Karte.',
    },
    ersetzen: {
      ueberschrift: 'Ersetzen',
      text: 'Ein Wort in den Textfeldern eines Sets austauschen. Erst wird gezeigt, was sich ändert; ein ⌘Z nimmt den ganzen Zug zurück.',
    },
    statistik: {
      ueberschrift: 'Statistik',
      text: 'Was im offenen Set steckt: wie viele Karten, wie oft ein Feld gefüllt ist, wie sich ein Feldwert verteilt. Zum Nachsehen, ob eine Reihe vollständig ist.',
    },
    bilder: {
      ueberschrift: 'Bilder',
      text: 'Schreibt die Karten des Sets als Bilddateien heraus — alle auf einmal, mit einem Namensmuster und einer Regel für gleiche Namen. **Zugewiesen** wird ein Kartenbild nicht hier, sondern am Bildfeld der Karte.',
    },
    druckbogen: {
      ueberschrift: 'Druckbogen',
      text: 'Die Karten nebeneinander auf Papier, als PDF mit Schnittmarken. Für den Drucker gedacht und nicht zum Weitergeben — dafür ist die Ausgabe da.',
    },
    schluesselwoerter: {
      ueberschrift: 'Schlüsselwörter',
      text: 'Ein Schlüsselwort erkennt seinen Text auf jeder Karte und setzt den Erinnerungstext dahinter. Die des Spiels sind fest und stehen hier zum Nachsehen; eigene gelten nur in diesem Set und lassen sich anlegen, ändern und löschen.',
    },
    pack: {
      ueberschrift: 'Zufallspack',
      text: 'Ein Booster aus den Karten dieses Sets, nach den Regeln des Spielpakets — welche Packarten es gibt, steht dort. Derselbe Same gibt dieselbe Ziehung; wer ein Ergebnis zeigen will, gibt die Zahl weiter.',
    },
    symbol: {
      ueberschrift: 'Symbol',
      text: 'Das Symbol dieses Sets — gezeichnet, nicht geladen. Eine Grundform ist der Anfang; danach lässt sich jeder Punkt verschieben. Tabulator wechselt zwischen den Punkten, die Pfeiltasten bewegen sie. Was hier entsteht, steht sofort in der Vorschau und kommt beim Sichern des Sets in die Datei.',
    },
    konsole: {
      ueberschrift: 'Konsole',
      text: 'Ein MSE-Script-Ausdruck im Kontext der aufgeschlagenen Karte — dieselbe Umgebung wie beim Zeichnen. Sie ändert nichts; sie rechnet und gibt zurück. Das Ergebnis kommt als Quelltext, damit eine Zeichenkette von einem Namen zu unterscheiden ist.',
    },
    ausgeben: {
      ueberschrift: 'Ausgeben',
      text: 'Karten als Bild oder über eine Exportvorlage als Text. Die Vorlage kommt aus einem Paket und bestimmt, was herauskommt: eine Liste für ein Forum, eine Tabelle, eine Datei für ein anderes Programm.',
    },
  },

  // **Über** (#1168) — die Rechtelage, die AGENTS.md §11 hier verlangt.
  ueber: {
    ueberschrift: 'Über MagicSetEditor4Mac',
    fassung: 'Fassung',
    lizenzName: 'Lizenz',
    lizenzSatz:
      'Der Quelltext dieser Anwendung steht unter der GNU General Public License, Fassung 2 oder später.',
    rechteName: 'Inhalte Dritter',
    rechteSatz:
      'Die Lizenz gilt für den Quelltext, **nicht** für Kartenrahmen, Mana-Symbole, Schriftarten, Regeltexte, Namen und Marken. Sie gehören ihren Rechteinhabern — unter anderem Wizards of the Coast — und nicht diesem Projekt.',
    quelleName: 'Quelltext',
    quelleSatz: 'github.com/DavidTHeider/MagicSetEditor4Mac',
  },

  weiterlesen: {
    ueberschrift: 'Weiterlesen',
    einleitung:
      'Zwei Schriften gehören zum Quelltext dieser Anwendung — im Projektverzeichnis unter `docs/`, nicht im Abbild, das du installiert hast. Diese Anwendung kann sie nicht selbst öffnen: sie führt keinen Weg nach draußen, und ein Verweis, der ins Leere zeigt, wäre schlechter als keiner.',
    handbuchName: 'Handbuch',
    handbuchSatz:
      'Was hier in Stichworten steht, in ganzen Sätzen: der erste Start, eine ' +
      'Karte bauen, ausgeben, die Kürzel — und was zu tun ist, wenn eine Karte ' +
      'leer bleibt. `docs/HANDBUCH.md`',
    administrationName: 'Administration',
    administrationSatz:
      'Wo die Daten liegen, wie man sichert und zurückspielt, was das Protokoll ' +
      'enthält und was bei einem Fassungswechsel geschieht — samt dem Satz, dass ' +
      'eine Sicherung **ohne Kennwort** Klartext ist. `docs/ADMINISTRATION.md`',
  },

  erststart: {
    ueberschrift: 'Beim ersten Start',
    text:
      'Diese Anwendung ist nicht notarisiert. macOS weist sie einmalig ab; ' +
      'die Freigabe steht in den Systemeinstellungen unter „Datenschutz & Sicherheit“.',
  },

  // **Führungen, und die Bereiche, die keine Seite sind** — Werkstück U11 von
  // Phase 5.1 (#1422).
  bereiche: {
    ueberschrift: 'Und ein Bereich, der keine Seite in der Leiste hat',
    einleitung:
      'Nicht alles, was diese Anwendung kann, steht in der Seitenleiste. Was von außen hereinkommt, ist kein Ort — und hat deshalb keine Seite.',
    kiName: 'KI',
    kiSatz:
      'Ein Sprachmodell liest und ändert Karten über den **MCP-Server** — ein eigener Prozess auf demselben Bestand, den du in deinem KI-Werkzeug einträgst. Was dabei hinausgeht und was nicht, steht unter **Schnittstellen**.',
    kiKnopf: 'Zu den Schnittstellen',
  },
  fuehrungen: {
    ueberschrift: 'Führungen',
    einleitung:
      'Eine Führung zeigt einen Weg durch die Anwendung, Schritt für Schritt, und lässt sich jederzeit wiederholen. Daneben steht in je einem Satz, wie die Anwendung aufgebaut ist.',
    startName: 'Die ersten Schritte',
    startSatz:
      'Drei Schritte bis zur ersten Karte: ein Set öffnen, Karten bearbeiten, Vorlagen beziehen. Sie stehen beim ersten Start von selbst; ob sie beim nächsten Start wiederkommen, ist eine Einstellung.',
    startZeigen: 'Die ersten Schritte jetzt zeigen',
    arbeitsplatzName: 'Der Arbeitsplatz',
    arbeitsplatzSatz:
      'Karten, Vorlagen, Sammlung und Decks teilen eine Ordnung: Navigator links, Inspector rechts, das Zentrum dazwischen. **⌥⌘[** klappt den Navigator, **⌥⌘]** den Inspector — und was du einklappst, bleibt je Profil eingeklappt. **⌥⌘S** klappt die Seitenleiste ein und aus, damit ihre Breite den Arbeitsplätzen gehört — der Zustand gehört diesem Gerät, gespeichert wird er je Profil.',
    paletteName: 'Die Befehlspalette',
    paletteSatz:
      '**⌘K** öffnet sie über jeder Seite: Seiten, Sets und Befehle, mit ein paar Buchstaben gefunden.',
    leisteName: 'Die Kontextleiste',
    leisteSatz:
      'Unter dem Anwendungsnamen steht, wer wo als was arbeitet — Mensch, Profil, Rolle. Gewechselt wird auf der Seite **Nutzer und Profile**; die Leiste führt hin.',
  },

  begriffe: {
    set: {
      name: 'Set',
      satz: 'Eine Sammlung von Karten samt ihren Angaben — Name, Beschreibung, Symbol. Sie liegt in einer Datei mit der Endung `.mse-set`.',
    },
    karte: {
      name: 'Karte',
      satz: 'Der Gegenstand, um den es geht. Eine Karte trägt Feldwerte — Name, Kosten, Regeltext — und wird nach einem Stilbogen gezeichnet.',
    },
    spiel: {
      name: 'Spiel',
      satz: 'Sagt, **welche Felder** eine Karte hat. Ein `.mse-game` erklärt sie: `magic` kennt Kosten und Seltenheit, ein anderes Spiel andere.',
    },
    stil: {
      name: 'Stilbogen',
      satz: 'Sagt, **wie** eine Karte aussieht: welche Ebene wo liegt, welches Bild den Rahmen gibt. Ein `.mse-style` gehört immer zu einem Spiel.',
    },
    stiloption: {
      name: 'Stiloption',
      satz: 'Eine Einstellung des Stilbogens, die für das ganze Set oder für eine Karte gilt — etwa der Rahmen einer Draft-Karte oder eine andere Auflage.',
    },
    symbolschrift: {
      name: 'Symbolschrift',
      satz: 'Zeichnet Manasymbole und Ähnliches. Sie ist keine Schriftart, sondern eine Sammlung kleiner Bilder mit Regeln, welcher Text welches zeigt.',
    },
    paket: {
      name: 'Paket',
      satz: 'Der Sammelbegriff für Spiel, Stilbogen, Symbolschrift und Exportvorlage. Pakete werden bezogen und eingebunden; ohne sie bleibt eine Karte leer.',
    },
    exportvorlage: {
      name: 'Exportvorlage',
      satz: 'Ein kleines Programm, das aus den Karten eines Sets **Text** macht — eine Liste für ein Forum, eine Tabelle, eine Datei für ein anderes Programm.',
    },
    druckbogen: {
      name: 'Druckbogen',
      satz: 'Ein PDF mit den Karten nebeneinander auf Papier, samt Schnittmarken. Zum Drucken und Ausschneiden — nicht zum Weitergeben als Datei.',
    },
    vorbehalt: {
      name: 'Vorbehalt',
      satz: 'Eine Zeile aus einer geöffneten Datei, die dieses Programm nicht gedeutet hat. Sie geht nicht verloren: sie reist unverändert mit und steht beim Sichern wieder da.',
    },
    befund: {
      name: 'Befund',
      satz: 'Ein Hinweis auf etwas, das beim Zeichnen nicht aufging — eine fehlende Ebene, ein Zeichen ohne Symbol. Die Karte wird trotzdem gezeichnet.',
    },
  },

  sprache: {
    ueberschrift: 'Warum manche Texte englisch sind',
    einleitung:
      'Die Aufschrift eines Kartenfeldes und der Satz darunter kommen aus dem **Spielpaket**, nicht aus dieser Anwendung. Sie stehen dort so, wie der Verfasser des Pakets sie geschrieben hat — meist englisch.',
    schluesselName: 'Der Name ist der Schlüssel',
    schluesselSatz:
      'Unter ihm steht der Wert in deiner `.mse-set`-Datei. Eine erfundene Übersetzung sähe hübscher aus und machte die Datei unlesbar für jeden, der darin sucht — deshalb erfindet diese Anwendung keine.',
    uebersetztName: 'Was übersetzt ist, wird übersetzt',
    uebersetztSatz:
      'Bringt ein Paket eine Übersetzung mit, steht sie da. Im bezogenen `M15-Magic-Pack` waren das am 2026-08-21 25 Kartenfelder — daher **Manakosten** und **Seltenheit**. Für die übrigen 224 brachte es keine mit, und dann steht dort dasselbe Wort wie in Magic Set Editor 2.',
    fensterName: 'Das Fenster selbst spricht acht Sprachen',
    fensterSatz:
      'Alles, was diese Anwendung selbst sagt — Knöpfe, Meldungen, diese Seite —, folgt der Sprache aus den Einstellungen. Deutsch ist die Ausgangssprache; die anderen sieben sind maschinell übersetzt und noch nicht redaktionell geprüft — das steht auch unter der Sprachwahl.',
    leisteName: 'Auch die Menüleiste',
    leisteSatz:
      'Das Anwendungsmenü und die Menüs zum Bearbeiten und für das Fenster tragen die Wörter, die macOS in dieser Sprache dafür benutzt. **Ausschneiden**, **Kopieren** und **Einsetzen** behalten dabei ihre Anbindung an das Betriebssystem.',
  },

  // **Was diese Anwendung über dich weiß** — Werkstück 13 von Phase 4.12
  // (#1351).
  //
  // # Warum es diesen Abschnitt gibt
  //
  // **Weil hier bis zum 2026-09-06 „Sonst nichts. Keine Telemetrie, kein
  // Konto, keine Anmeldung." stand.** Der Satz war wahr, als er geschrieben
  // wurde; seit Phase 4.12 gibt es Konten, eine Anmeldung über vier fremde
  // Anbieter, einen zweiten Faktor und ein Formular mit Anrede, Namensteilen
  // und Biografien.
  //
  // **Ein Wächter hält den Abschnitt jetzt gegen den Quelltext**
  // (`internal/werkzeuge/datenschutz_test.go`): jedes personenbezogene Feld
  // und jede angesprochene Adresse muss hier vorkommen.
  datenschutz: {
    ueberschrift: 'Was diese Anwendung über dich weiß',
    einleitung:
      'Seit es Nutzer und Profile gibt, erfasst diese Anwendung **personenbezogene Daten**. Hier steht, welche das sind, wo sie liegen, was davon diesen Rechner verlässt — und was ausdrücklich nicht.',

    wasName: 'Was erfasst wird',
    wasSatz:
      'Alles davon ist **freiwillig** außer dem Anzeigenamen. Zur Person: **Anrede**, **Namenszusatz**, **Vorname**, **weitere Vornamen**, **Nachname**, **Slogan** und zwei **Biografien** — eine kurze und eine lange. Zum Konto: **Adresse**, **Zweitadresse**, **Rufnummer** und, wenn du eines setzt, ein **Kennwort**. Dazu ein **Profilbild** und je Profil ein **Abzeichen**.',

    woName: 'Wo es liegt',
    woSatz:
      'In deinem Bestand — einer SQLite-Datei in deinem Benutzerordner — und in Bildern daneben. **Drei Dinge liegen woanders**, nämlich im **Schlüsselbund** von macOS: das Geheimnis deines zweiten Faktors, das Geheimnis einer Anmeldung beim Anbieter, und Zugangsdaten fremder Kartendienste. Den Bestand gibt man weiter, den Schlüsselbund nicht — deshalb die Trennung.',

    formName: 'In welcher Form',
    formSatz:
      'Ein **Kennwort** wird nie im Klartext abgelegt: es wird mit **Argon2id** gehütet, und die gehütete Form trägt ihre eigenen Zahlen mit. Ein **Zugriffstoken** liegt als Prüfsumme da; den Klartext gibt es genau einmal, gleich nach dem Anlegen. Alles andere steht so da, wie du es getippt hast — es ist deine Datei. **Und die Datei selbst lässt sich verschlüsseln** — unter **Einstellungen › Daten und Sicherung › Bestand**. Dann liegt nichts davon mehr lesbar auf der Platte; der Schlüssel geht in den Schlüsselbund dieses Rechners und ist **nicht wiederherstellbar**. Eine **Sicherung** bekommt ihr eigenes **Kennwort** — dann ist sie verschlüsselt und geht trotzdem auf jedem Rechner auf; ohne Kennwort gibt ein verschlüsselter Bestand sie nur auf ausdrücklichen Wunsch her.',

    ausName: 'Was diesen Rechner verlässt',
    ausSatz:
      'Bei einer **Anmeldung** geht deine Wahl an den Anbieter, den du gewählt hast: `github.com` und `api.github.com`, `accounts.google.com` und `oauth2.googleapis.com`, `login.microsoftonline.com`, oder `twitter.com` und `api.twitter.com`. Gefragt wird dabei nur nach deinem **Namen und deiner Kennung** dort — **nicht** nach deiner E-Mail-Adresse, nicht nach deinen Kontakten, nicht nach deinen Beiträgen. Zurück kommen und gespeichert werden **drei** Dinge: der Anbieter, deine Kennung bei ihm, und dein Anzeigename dort.',

    nichtName: 'Was nicht hinausgeht',
    nichtSatz:
      'Deine **Namensteile, Biografien, Rufnummer und Adresse** verlassen diesen Rechner nie. Sie stehen in deinem Bestand und werden an keine Stelle geschickt. Es gibt **keine Telemetrie**, keine Absturzberichte, keine Nutzungszahlen und kein Konto bei uns — es gibt kein „uns“.',

    dienstName: 'Und die Kartendienste',
    dienstSatz:
      'Wer nach Karten sucht, fragt `api.scryfall.com` und — wenn eingeschaltet — `api.magicthegathering.io`, `graphql.mtgjson.com` oder `api.scrydex.com`. Dabei geht die **Suchanfrage** hinaus und sonst nichts: kein Nutzername, keine Kennung, nichts aus deinem Profil. **Die Vorschaubilder der Treffer** holt das Fenster selbst — von `cards.scryfall.io` und, bei magicthegathering.io, von `gatherer.wizards.com` und `gatherer-static.wizards.com`; dorthin geht die Adresse des Bildes und sonst nichts. Preise für die ganze Sammlung holt — wenn eingeschaltet — `mtgjson.com`, Marktangebote `api.cardtrader.com`; dorthin geht die Kennung der Karte und, bei CardTrader, deine Zugangsmarke, und sonst nichts. Vorlagenpakete kommen über `codeload.github.com`, Fassungshinweise über `api.github.com`. Was jeder Dienst hier kann und unter welchen Bedingungen, steht unter **Einstellungen › Dienste › Fremde Kartendienste**; die Verweise dort führen zu `scryfall.com`, `docs.magicthegathering.io`, `mtgjson.com`, `scrydex.com` und `www.cardtrader.com` — die Anwendung selbst ruft diese Seiten nicht, sie öffnen sich erst auf deinen Klick im Browser. **Und die Marktverweise an einer Karte** (T8): sie führen zu `www.cardmarket.com`, `www.cardtrader.com`, `www.tcgplayer.com`, `www.cardkingdom.com`, `gatherer.wizards.com` und `www.mtggoldfish.com` — **gebaut, nicht aufgerufen**: die Adresse entsteht auf diesem Rechner aus dem Kartennamen oder einer Anbieterkennung, und hinaus geht sie erst auf deinen Klick, im Browser deiner Wahl. Ein Schlüssel ist dafür nicht nötig und steckt auch in keiner dieser Adressen.',

    // **Und die KI** — Werkstück A8 von Phase 5.3
    // (#1447).
    // Sie steht **nach** den Kartendiensten, weil sie die unangenehmere
    // Auskunft ist: ein Kartendienst bekommt eine Suchanfrage, ein
    // Cloudmodell bekommt den Inhalt der Karte.
    kiName: 'Und die KI',
    kiSatz:
      'Ein **lokales** Modell läuft auf diesem Rechner und verlässt ihn nicht: die Anwendung spricht `127.0.0.1` an und sonst nichts. Ein **Cloudmodell** ist etwas anderes — wer Anthropic, OpenAI oder xAI einschaltet, schickt an `api.anthropic.com`, `api.openai.com` beziehungsweise `api.x.ai` den **Auftrag und die Felder der Karte**, um die es geht, dazu deinen Schlüssel aus dem Schlüsselbund. **Bei OpenAI bleibt nichts dort liegen**: die Schnittstelle legt eine Antwort von sich aus ab, und diese Anwendung schaltet das ausdrücklich aus. **Bei xAI ebenso**: der Weg, der Antworten 30 Tage hält, wird gar nicht erst benutzt. Nichts aus deinem Profil, keine zweite Karte, kein Set. Beides ist **aus**, bis du es einschaltest, und unter **Einstellungen › Dienste › Nur lokal arbeiten** bleibt jedes Cloudmodell aus, auch wenn sein Schalter anders steht. Was zurückkommt, wird **gezeigt und nicht übernommen** — was in ein Feld soll, bestätigt ein Mensch. Was Anthropic hier kann und unter welchen Bedingungen, steht unter **Einstellungen › Dienste › Fremde Kartendienste**; die Verweise dort führen zu `www.anthropic.com`, `platform.claude.com`, `openai.com`, `developers.openai.com`, `x.ai` und `docs.x.ai` — die Anwendung selbst ruft diese Seiten nicht, sie öffnen sich erst auf deinen Klick im Browser.',

    auskunftName: 'Auskunft und Löschen',
    auskunftSatz:
      'Unter **Nutzer und Profile › Mein Zugang** gibst du dir deine Daten als Datei aus — alles, was hier steht, in einem Zug. Und unter **Nutzer und Profile** löschst du einen Nutzer — mit seinen Profilen, seinem Bild und dem Geheimnis seines zweiten Faktors. **Was in seinen Profilen liegt, bleibt**: Sets, Decks, Sammlung und Turniere sind dann keinem Profil mehr zugeordnet und über die Oberfläche nicht mehr erreichbar. Beides braucht kein Netz und niemanden, der zustimmt.',
  },

  schnittstellen: {
    ueberschrift: 'Was hinausgeht und hereinkommt',
    einleitung:
      'Diese Anwendung arbeitet offline. **Nach außen ins Netz** geht sie nur, wenn jemand es auslöst: das Beziehen einer Vorlage, eine Kartensuche und eine Anmeldung brauchen einen Knopfdruck; nach neuen Fassungen sieht sie nur, wenn eine Einstellung es erlaubt, und die ist von Haus aus **aus**. Welche Adressen dabei angesprochen werden und was dabei hinausgeht, steht unter **Was diese Anwendung über dich weiß**. Daneben gibt es drei Wege auf diesen Rechner, die ebenfalls niemand von selbst nimmt — und einen, der von selbst läuft: scheitert der Start, zeigt sie den Grund in einem Fenster.',
    bezugName: 'Vorlagen beziehen',
    bezugSatz:
      'Auf Knopfdruck holt die Anwendung ein Vorlagenpaket über HTTP, an einem festen Stand und mit Prüfsumme. Ohne Knopfdruck geschieht nichts.',
    scryfallName: 'Kartendaten von Scryfall',
    scryfallSatz:
      'Auf Knopfdruck fragt die Anwendung `api.scryfall.com` nach einer Karte, einem Satz oder einem Kartenbild. Was von dort kommt, wird **gezeigt und nicht übernommen** — was in ein Feld soll, wählt ein Mensch. Ohne Netz bleibt der letzte gute Stand stehen, und er sagt, wie alt er ist.',
    fassungName: 'Nach neuen Fassungen sehen',
    fassungSatz:
      'Die einzige Stelle, die ohne Knopfdruck fragen kann — und **nur**, wenn sie unter **Einstellungen** eingeschaltet ist. Die Vorgabe ist **aus**, und dann geht keine einzige Anfrage hinaus. Eingeschaltet fragt sie `api.github.com` nach der neuesten Freigabe. Geladen wird erst auf Knopfdruck, und dann werden **Signatur und Prüfsumme geprüft, bevor eine Datei liegt**; installiert wird nichts von selbst.',
    mcpName: 'Der MCP-Server',
    mcpSatz:
      'Ein eigenes Programm namens `mse-mcp`, mit dem ein Sprachmodell an einem Set arbeiten kann: Karten und Felder lesen, Felder setzen, Karten anlegen, kopieren, löschen und umordnen, ein Set einfahren, ein Set sichern und eine Karte zeichnen. Der Maßstab ist nicht die Menge, sondern die Umkehrbarkeit: kein Werkzeug überschreibt etwas Vorhandenes — wo schon eine Datei liegt, weist es ab —, und jede Änderung sagt, wie sie zurückzunehmen ist. Startet es ein Klient wie Claude Desktop selbst, spricht es über Standardein- und -ausgabe: kein Port, keine Adresse. **Über HTTP** lauscht es erst, wenn du es unter **Einstellungen › Verbindungen** startest, und nur dort, wo **Wo die Dienste lauschen** es sagt.',
    mcpNichtAn:
      'Es läuft nicht mit, aber es **liegt bei**: im Bündel unter `MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp`, zusammen mit `mse`, der Kommandozeile. **Von selbst legt diese Anwendung nichts außerhalb ihres Bündels ab.** Unter **Einstellungen › Verbindungen › Werkzeuge ohne Fenster** richtest du auf Knopfdruck ein, was du willst — den Verweis auf `mse` in `~/.local/bin` samt `PATH`-Eintrag in `~/.zprofile`, den Eintrag für Codex, Claude Desktop oder Claude Code —, und dort nimmst du jeden Eintrag auch wieder zurück; was außerhalb ihres eigenen Blocks steht, fasst sie nicht an. Was die Werkzeuge ändern, merkt dieses Fenster von selbst: es fragt alle zwei Sekunden nach dem Stand des Bestands und lädt die betroffene Seite nach.',
    dienstName: 'Der Dienst für Skripte',
    dienstSatz:
      'Ein eigenes Programm namens `mse-rest`, über das ein Skript dieselben Fähigkeiten erreicht wie `mse` — Sets und Karten lesen, ein Feld setzen, ein Set einfahren oder sichern, eine Karte zeichnen. Jede Anfrage braucht eine Marke. **Wo er lauscht, entscheidest du** unter **Wo die Dienste lauschen**: vorgegeben ist nur dieser Mac (`127.0.0.1`); wählst du eine Adresse, die diesen Mac verlässt, verlangt er TLS und einen Zugang.',
    dienstNichtAn:
      'Er liegt im Bündel unter `MagicSetEditor4Mac.app/Contents/MacOS/mse-rest` und startet nicht von selbst: wer keinen Dienst will, hat keinen. Gestartet wird er unter **Einstellungen › Verbindungen** — oder als LaunchAgent unter **Werkzeuge ohne Fenster**; dann startet macOS ihn bei jeder Anmeldung, bis du ihn dort zurücknimmst. Wer ihn startet, bekommt Adresse und Marke angesagt. Anfragen aus einem Browser werden abgewiesen, bevor die Marke überhaupt geprüft wird.',
    startfensterName: 'Die Meldung, wenn der Start scheitert',
    startfensterSatz:
      'Der einzige Weg, den die Anwendung **von selbst** nimmt. Lässt sich der Bestand nicht öffnen, ruft sie `osascript` von macOS auf, um den Grund in einem Fenster zu zeigen — sonst stünde er nur auf einem Kanal, den bei einem Doppelklick niemand sieht. Der Text geht als Wert hinein, nicht als Skript.',
    dateienName: 'Dateien auf diesem Rechner',
    dateienSatz:
      'Sets, Vorlagenpakete, geschriebene Bilder und PDF. Sie liegen dort, wo du sie hinlegst; der Bestand liegt in deinem Benutzerordner.',
    sonstSatz:
      'Sonst nichts. **Keine Telemetrie**, keine Absturzberichte, keine Nutzungszahlen. Ein Konto gibt es — aber bei dem Anbieter, den du wählst, und nicht bei uns; was dabei hinausgeht, steht unter **Was diese Anwendung über dich weiß**.',
    grenzenName: 'Grenzen',
    grenzenSatz:
      'Jeder Maschinenzugang hat ein **Guthaben**, und jeder Ruf kostet nach seiner Art; was davon leer ist, sagt die Antwort samt Wartezeit. Schwere Arbeiten — Rastern, Einfuhr, Ausfuhr, Massenvorgänge — haben je Klasse eine feste Zahl an Plätzen, die Fenster, Kommandozeile, MCP und Dienst teilen. Wer rät, sperrt nur seinen eigenen Nachweis.',
    grenzenZahlen:
      'Heute: {burst} Einheiten am Stück, {nachfuellen} je Minute; Lesen kostet {lesen}, Schreiben {schreiben}, eine Einfuhr {einfuhr}; höchstens {rendern} Rasterungen zugleich; {reserve} Einheiten bleiben dem Fenster.',
    grenzenOhneKontingente:
      '**Die Kontingente sind abgeschaltet** (`grenzen.kontingente`) — für eine kontrollierte Umgebung. Die Plätze je Arbeit, die Grenzen für Rumpf und Antwort, die Fristen, die Sperre gegen das Raten, Zugang und TLS gelten weiter.',
  },
} as const
