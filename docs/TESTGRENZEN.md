# Was die Tests hier nicht sehen werden

**Verlegt aus `AGENTS.md` am 2026-09-22** — ENG-01.AK-04
(#1726):
„Detailkonzepte wandern in passende Dokumente statt beide Regeldateien
weiter mit ganzen Projektplänen zu überladen." **Die Regel steht weiter
dort** — hier steht, warum sie so lautet, samt jeder Messung.

---

Diese Grenzen gelten von Anfang an, weil sie aus einem Schwesterprojekt teuer
bezahlt sind:

- **Zwischen den Prüfebenen.** Frontend-Tests gehen über `test/api.ts` und
  laufen nie durch die JSON-Serialisierung; Go-Tests sehen die Struct-Marken
  nicht. Ein `json:",omitempty"` auf einem Slice lässt eine **leere** Liste aus
  der Antwort verschwinden, und die Seite bleibt leer — bei grüner Pipeline.
  Deshalb: vor jeder Freigabe die **gebaute** Anwendung starten.
- **Eine Vorrichtung, die großzügiger ist als der Betrieb.** Wo eine Vorrichtung
  eine Antwort **zusammensetzt**, prüft ein Test sie und nicht den Weg; die
  Zusage gehört auf die Go-Seite, an die Naht.

  **Und wo sie schneller ist als der Betrieb, prüft er die Uhr.** Bis zum
  2026-09-02 antwortete `frontend/src/test/api.ts` **synchron**, während im
  Betrieb jeder Aufruf über die Wails-IPC geht — zwischen Ruf und Antwort
  liegt dort immer mindestens eine Aufgabe. Seit sie eine Makroaufgabe später
  antwortet, fielen **20 von 2.186** Tests um: jeder ein Griff in einen Baum,
  den es noch nicht gab. Sie waren nicht falsch geschrieben, sondern gegen
  eine Welt geschrieben, die es nicht gibt.
- **Farbkontraste.** Unter jsdom meldet axe `color-contrast` als *unentschieden*.
  Die Sichtprüfung im Browser findet, was die Tests durchlassen.
- **Eine Eingabe, die der Test schon aufgeräumt hat.** Wo etwas umgerechnet
  werden muss, gehört die **rohe** Form in den Test — die Zeichenkette, nicht
  der geparste Wert; die Wanduhrzeit mit Zone, nicht der UTC-Kalendertag.
- **Eine Zusicherung, die schon vor der Antwort zutrifft.** `waitFor` gibt
  zurück, sobald sie **einmal** hält — traf sie vorher schon zu, ist der Test
  grün, ohne die Wirkung geprüft zu haben. Gewartet wird deshalb auf ein
  eindeutiges **Abschlusssignal** — einen Spion, der erst nach der Antwort
  läuft —, und der Zustand wird **danach** zugesichert. Gefunden am
  2026-08-20 in #621:
  der Test sicherte zu, dass die aufgeschlagene Karte stehen bleibt, und sie
  stand zum Zeitpunkt der ersten Prüfung ohnehin noch da.

  **Ihre häufigste Gestalt ist ein Warten auf den Ruf statt auf die
  Antwort**: `await waitFor(() => expect(spion).toHaveBeenCalled())`, und
  darunter ein Griff an den DOM. Der Ruf ist geschehen, die Antwort nicht —
  und was sie bewirkt, steht noch nicht da. Gewartet wird auf die **Wirkung**
  (`findBy…`, oder `waitFor` über die Zusicherung selbst); wo es keine
  sichtbare Wirkung gibt, weil die Aussage eine Verneinung ist, wird die
  Antwort ausdrücklich abgewartet. Ein Wächter zählt die Gestalt nach —
  `frontend/src/wartemuster.test.ts`, sieben Fundstellen beim ersten Lauf.
- **Und ihre Schwester: eine Verneinung nach einem Ereignis, das niemand
  gehört hat.** `expect(x).not.toHaveBeenCalled()` trifft auch dann zu, wenn
  der Horcher noch gar nicht hing — ein `useEffect` läuft nach dem Zeichnen,
  und `fireEvent` davor tut nichts. Der Test ist grün, und der Zweig dahinter
  bleibt **unbetreten**; sichtbar wird das nur an der Abdeckung, und die
  flackert dann unter Last.

  Vor eine Verneinung gehört deshalb ein **Beleg, dass jemand zuhört**: ein
  zweites Ereignis durch denselben Weg, dessen Wirkung man sieht. Gefunden am
  2026-08-20 in #605
  — rund jeder vierte volle Lauf war rot, und nachgestellt hat es sich, indem
  die Taste absichtlich vor dem Zeichnen gefeuert wurde.
- **Und der Beleg selbst geht ins Leere, wenn er zu früh kommt.** `await
  findByRole(…)` sagt zu, dass ein Knoten im DOM steht — **nicht**, dass ein
  `useEffect` gelaufen ist. Testing Library schaltet für jede asynchrone
  Abfrage die act-Umgebung ausdrücklich ab und gibt danach eine **einzige**
  Makroaufgabe als Frist (`@testing-library/react@16.3.2`, `dist/pure.js`,
  Zeilen 83–99); React legt seine passiven Effekte über den Scheduler in eine
  **andere** Aufgabe. Unter Last reicht die Frist nicht.

  Dann steht der Knoten da, während der Horcher daneben noch nicht hängt — und
  **`waitFor` heilt das nicht**: es wiederholt die Zusicherung, nicht das
  Ereignis. Ein verlorener Tastendruck kommt nicht wieder, gleich wie lange
  gewartet wird. Wer auf einen asynchron aufgebauten Baum ein Fensterereignis
  feuert oder in einen Rückruf greift, den ein Effekt gesetzt hat, ruft vorher
  `effekteNachholen()` (`frontend/src/test/effekte.ts`).

  Gefunden am 2026-08-21 in #761:
  vierzig volle Läufe, **vier rot**, und zwar in **zwei** verschiedenen Tests
  derselben Datei — einmal ein verlorener Tastendruck, einmal ein Horcher, den
  eine `Map` noch nicht kannte und den `?.()` still fallen ließ. Gezählt hat es
  eine Messung an `window.addEventListener`: einmal von zwölf Läufen unter Last
  hingen nach der Abfrage **null** Horcher — und nach `act` einer. Dieselbe
  Zahl belegt den Grund und die Abhilfe.
- **Und ein Test, der nach draußen greift, ohne dass jemand es wollte.** Eine
  Naht, die einen fremden Dienst oder das Zuhause des Rechners anfasst, tut es
  im Test auch — wenn nichts sie daran hindert. Der Lauf ist dann **grün**,
  weil er wirklich funktioniert hat.

  **Zwei Dinge gehören an jeden solchen Test**: `t.Setenv("HOME", t.TempDir())`,
  damit er nicht den Stand dieses Rechners sieht, **und** eine umgelenkte
  Adresse. Beides einzeln reicht nicht.

  Gefunden am 2026-08-24 in #888:
  ein Bindungstest hat **zweimal 25 MB von Scryfall geladen**. Beim ersten Mal
  fehlte die Adresse. Beim zweiten war sie gesetzt — und der Zwischenspeicher im
  **echten** Zuhause hielt noch das Verzeichnis vom vorigen Lauf, dessen
  Dateiadressen auf `data.scryfall.io` zeigen, das die Umlenkung gar nicht
  erfasst. Beide Male fiel es nur daran auf, dass der Lauf grün war, wo ein
  Fehler erwartet wurde.

- **Und der Schlüsselbund dieses Rechners ist der dritte, und er fragt
  zurück.** Netz und Zuhause fallen still aus; der Anmeldebund von macOS
  öffnet einen **Dialog** — mitten in einem Lauf, und ein Mensch muss ihn
  wegklicken. Ein Lauf, der jemanden um eine Entscheidung bittet, hat die
  Grenze schon überschritten, gleich wie sie ausgeht.

  **Die Regel dafür stand seit jeher da, und sie hat nicht getragen.**
  Gemessen am 2026-09-14 mit einer Instrumentierung an
  `internal/schluesselbund`: **84** Zugriffe auf den Anmeldebund aus **acht**
  Paketen, darunter **zwei Schreibvorgänge** von `peeridentitaet`. Der Dialog
  kam dreimal.

  **Deshalb ist die Vorgabe jetzt sicher, statt die Regel zu wiederholen**
  (§10): `internal/schluesselbund` weist den Anmeldebund ab, sobald das
  laufende Programm ein Testbinärprogramm ist (`os.Args[0]` endet auf
  `.test`). Wer ihn braucht, bekommt ihn nicht — er bekommt einen eigenen.
  Der Wächter dazu steht in
  [`internal/schluesselbund/keinanmeldebund_test.go`](../internal/schluesselbund/keinanmeldebund_test.go).

  **Was das nicht fängt, ist ein Kindprozess**: `mse`, `mse-mcp` und
  `mse-rest` heißen nicht `.test`. Für sie setzt `internal/flaechen` die
  Umgebung an **einer** Stelle — `eigenerBestand`, an der jede Kindumgebung
  vorbeikommt. Nach beidem war die Messung **null**.

  **Und es hat etwas freigelegt, das sonst niemand gesehen hätte**: zwei
  Bindungen gaben auf ihrem **Fehlerweg** `null` statt einer leeren Liste.
  Sie trugen den Satz „eine leere Liste und kein `nil`" seit ihrer Entstehung
  im Kopf — er galt für den Erfolgsweg. Vorher kam dort nie ein Fehler an,
  weil der Schlüsselbund dieses Rechners immer antwortete.
