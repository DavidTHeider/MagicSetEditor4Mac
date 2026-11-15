# 0134 — Eigene Gestaltung sind Farben, kein Stylesheet

**Stand:** angenommen
**Datum:** 2026-09-25

## Lage

UX-09 (#1681)
verlangt „benutzerdefinierte Hintergründe, Schriften und Textfarben über die
ganze App" (AK-01 bis AK-03) und **eigene CSS-Erweiterungen**, die auf
„deklarative, eingeschränkte Gestaltung begrenzt" sind — ohne
Skriptausführung, ohne unbemerkte Netzbezüge, ohne verteilte Schriftdateien
(AK-04). Und ein sicherer Standard muss sich wiederherstellen lassen, auch
wenn die Konfiguration unlesbar ist (AK-03).

## Entscheidung

**Die eigene Gestaltung sind Farben für die zehn Merkmale der
Gestaltungssprache** (`--mse-grund` bis `--mse-warnung`), eine Zeile je
Merkmal, eine Hexzahl je Zeile (`domain.PruefeGestaltung`). Das Fenster setzt
sie einzeln über `style.setProperty` — nie als Text in ein `<style>` — und
**nimmt sie zurück, wenn ein Paar unter seiner Schwelle liegt** (Text
4,5 : 1, Ränder 3 : 1), gemessen am Wurzelelement nach Farbwelt und
Erscheinungsbild, bevor gezeichnet wird.

**Nachgetragen am 2026-09-26, nach dem Zweitblick über Phase 6.3 (S4)**: die
Sperre maß bis dahin nur die zehn Merkmale. **Drei Dinge lagen daneben** und
werden seitdem mitgemessen oder weggenommen: die Farben der Meldungen
(Tailwinds Rot, Bernstein und Smaragd folgen dem Erscheinungsbild und nicht
den Merkmalen — bei dunklen eigenen Farben im hellen Bild stand jede
Fehlermeldung bei 2,07 : 1), der Grund über einem Fensterbild (gemessen unter
dem Schleier über Schwarz und Weiß) und die Faser, die über eigenen Farben
jetzt nicht mehr liegt. Der Ring auf der betonten Fläche ist das siebzehnte
Paar.

Dazu drei feste Farbwelten neben Pergament (Klar, Schiefer, Moos), gerechnet
wie Pergament; ein eigenes **Fensterbild** unter einem Schleier von 90 % (bis zum Nachlauf über Phase 6.3 88 %: die Warnung stand auf Pergament hell dann bei 4,45 : 1),
der Text über jedem Bildpunkt lesbar hält; und die **Anzeigeschrift über die
ganze Oberfläche** — aber nur eine der vier Textschriften.

**Hoher Kontrast sticht alles davon.** Und **„Darstellung zurücksetzen"**
steht in der Menüleiste, die macOS zeichnet: die Layoutrücksetzung erreicht
auch, wer im Fenster nichts mehr lesen kann.

## Verworfen

- **Ein Stylesheet mit Selektoren.** Ein Selektor ist eine Adresse in einen
  Baum, der sich mit jeder Fassung ändert — eine Zusage über Fassungen hinweg
  (0033), die niemand halten kann. Und eine einzige Regel `display: none` am
  Knopf „Zurücksetzen" nähme den Weg zurück gleich mit. Prüfen ließe sich
  ein solches Blatt nur, indem man einen CSS-Zerteiler mitliefert und jede
  Eigenschaft einzeln erlaubt — am Ende stünde dieselbe Liste von Merkmalen,
  nur schwerer zu lesen.
- **Farben ohne Sperre, nur mit Warnung.** Eine Warnung ist Text, und Text
  ist genau das, was eine unlesbare Farbe unlesbar macht.
- **Eigene Schriftdateien.** Eine mitgebrachte Schrift wäre eine Datei im
  Bestand, die mit einem Austausch weiterreiste — die Verteilung, die AK-04
  ausschließt. Gewählt wird unter Schriften, die auf jedem Mac liegen.
- **Das Tischbild als Fensterbild.** Ein Tisch trägt Filz oder Holz, ein
  Fenster ein ruhiges Motiv; wer beides gleich will, wählt dieselbe Datei.

## Folgen

- Was außerhalb der zehn Merkmale liegt — Abstände, Radien, der Tisch —,
  lässt sich nicht selbst gestalten. Wer das braucht, braucht ein neues
  Merkmal, und das ist eine Änderung an `stil.css` mit gemessenem Kontrast.
- Eine eigene Farbe, die im hellen Bild trägt und im dunklen nicht, gilt dort
  nicht; die Einstellungsseite nennt das Paar.
- Der Schleier macht das Bild blass. Das ist der Preis der Lesbarkeit, und
  er steht in `docs/KONTRASTE.md` mit seiner Rechnung.
