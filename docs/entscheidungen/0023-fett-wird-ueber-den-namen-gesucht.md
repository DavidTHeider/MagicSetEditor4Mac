# 0023 — Fett wird über den Namen gesucht, nicht über ein Gewicht

**Stand:** 2026-08-17
**Betrifft:** #326

## Der Zwang

MSE behandelt kursiv und fett **nicht** gleich. Kursiv ist ein *Name*: der Stil
nennt die Schrift unter `italic name:`. Fett ist ein *Gewicht*:

```cpp
static const String BOLD_STRING = _(" Bold");
…
String familyName = name();
if (familyName.EndsWith(BOLD_STRING)) {
  familyName = familyName.Left(familyName.length() - BOLD_STRING.length());
  weight_i = wxFONTWEIGHT_BOLD;
}
font = wxFont(size_i, wxFONTFAMILY_DEFAULT, style_i, weight_i, underline(), familyName);
```

`src/data/font.cpp`, Zeilen 82–108, Fassung f9d9356 vom 2021-07-06.

MSE reicht **Familienname und Gewicht** an wxWidgets weiter, und das
Betriebssystem sucht die passende Schnittdatei. `render.Schriftquelle` nimmt
dagegen nur einen **Namen** entgegen, und `internal/schriften` findet eine
Datei über drei Wege: Dateiname, Paketbaum, und den Namen, den die Schrift
über sich selbst führt (`NameIDFull`).

## Die Entscheidung

**Der Wunsch trägt ein Gewicht, die Suche macht daraus einen Namen.**

`Schriftquelle.Suche(name string, fett bool)`. Aufgelöst wird in dieser
Reihenfolge:

1. Ein abschließendes ` Bold` am angefragten Namen wird **abgeschnitten** und
   als „fett" gelesen — genau MSEs eigene Regel. `Beleren Bold` ist damit die
   Familie `Beleren` im fetten Schnitt, und `<b>` an einem solchen Feld ändert
   nichts. Das ist die Antwort auf die offene Frage am Issue.
2. Ist fett verlangt, wird `<Familie> Bold` gesucht.
3. Findet sich die nicht, wird die Familie **aufrecht** genommen.

Schritt 2 ist Namensarithmetik, und sie ist gedeckt: der Index von
`internal/schriften` steht auf `NameIDFull` — dem vollen Namen, den die Datei
über sich selbst führt —, und ein voller Name lautet gemeinhin
„Familie Schnitt". `Beleren Bold` liegt in `M15-Magic-Pack` als
`beleren-bold_P1.01.ttf` und wird über genau diesen Weg schon heute gefunden
(#240).

## Was verworfen wurde

**Ein Gewicht bis in die Schriftdatei durchreichen** — also einen zweiten Index
über Familie (`NameIDFamily`) und Schnitt (`NameIDSubfamily`) und darin nach
`Bold` suchen. Das wäre die Nachbildung dessen, was das Betriebssystem tut, und
es ist die robustere Antwort für Schriften, deren voller Name der Konvention
nicht folgt.

Es ist **nicht** gebaut, und der Grund ist nicht „zu aufwendig": der Index wird
ohnehin über jede Datei gebaut, die Ergänzung wären wenige Zeilen. Der Grund
ist, dass er eine zweite Auflösungsregel neben die erste stellte, ohne dass ein
Fall vorliegt, den die erste verfehlt. Sobald einer auftaucht — eine Schrift,
deren fetter Schnitt nicht `<Familie> Bold` heißt —, ist das der Anlass, und
dann steht hier ein Nachtrag statt einer Regel auf Vorrat.

## Der Preis

**Ein fetter Schnitt, der anders heißt, wird nicht gefunden**, und der Text
steht dann aufrecht statt fett. Das ist dieselbe Bestenfalls-Arbeit wie bei der
Kursiven (§8): ein aufrechter Text ist eine Karte, ein fehlender ist keine.
Gemeldet wird es nicht — eine Meldung je Textebene ohne fetten Schnitt stünde
auf jeder Karte, deren Stil eine Systemschrift ohne Bold nennt.

**~~`<bullet>` setzt fett mit.~~ Zurückgenommen am 2026-08-18**
(#429).

Hier stand: „MSE prüft `is_tag(text, pos, "<b")` als **Präfix** und trifft
damit auch `<bullet>`, bevor der eigene `</bullet`-Zweig überhaupt erreicht
wird. Das ist nachgebildet, nicht behoben."

**Das ist falsch, und der Fehler ist ein halb gelesenes `is_tag`:**

```cpp
bool is_tag(const String& str, size_t pos, const String& tag) {
  return is_substr(str, pos, tag)
      && pos+tag.size() < str.size()
      && is_tag_end_char(str[pos+tag.size()]);
}
```

`src/util/tagged_string.cpp`, Zeilen 270–272, Fassung f9d9356 vom 2021-07-06.
Es ist **kein** reiner Präfixvergleich: hinter dem Präfix muss ein
Endzeichen stehen — `>`, `-`, `:` oder ein Leerzeichen (ebd. Zeilen 28–30). In
`<bullet>` steht dort ein `u`, und `<b` trifft nicht.

**Die Gegenprobe stand die ganze Zeit im Altcode selbst**: `element.cpp:171`
führt einen eigenen `</bullet`-Zweig. Den erreichte niemand, wenn `<b` ihn
vorher schluckte — der Satz „bevor der eigene `</bullet`-Zweig überhaupt
erreicht wird" widerlegt sich an dem Zweig, den er nennt.

**Was daran teuer war:** die Abweichung war als *gewollte Parität* aufgeschrieben
und mit Datei und Zeile belegt. Ein Leser, der die Fundstelle nicht selbst
aufschlägt, hat keinen Anlass zu zweifeln — genau die Lage, vor der §10 warnt
(„die Erinnerung eines Werkzeugs ist kein Prüfergebnis"). Der Beleg war echt,
die Aussage daraus nicht.

**Was jetzt gilt:** die Endzeichenregel steht als `service.trifft` an **einer**
Stelle und deckt `<i>`, `<b>`, `<sym>`, `<line>`, `<soft>` **und** die Frage,
welche Auszeichnung MSE überhaupt deutet. `<bullet>` wird gemeldet, nicht fett
gesetzt: `</bullet` steht in MSEs Kette, und der Rand des Aufzählungspunktes
ist Verhalten, das hier noch fehlt.
