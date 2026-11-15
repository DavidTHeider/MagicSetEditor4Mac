# Hüllenprobe: uxrf1

**bestanden** — Commit `1c0b74221`,
2026-10-09 22:18 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-09T19-47-35Z-46912`, Bauabdruck `1b6e1cccf72a`.


**Zu: die Rolle, der Name, der Wert — und keine Liste im Baum**

- zu: {"rolle":"combobox","name":"Zeitzone ▾","wert":"Wie das System","platzhalter":null,"offen":"false","steuert":null,"haspopup":"listbox","autocomplete":"list","gesperrt":false,"listeRolle":null,"listeName":null,"listeId":null,"zeilen":0,"texte":[],"alleTexte":[],"stelle":-1,"aktivRolle":null,"aktivText":null,"aktivGewaehlt":null,"fokusImFeld":false,"ansage":"","listeImFenster":null,"listeBreit":null,"feldBreit":320,"obenAufliegend":null,"markiert":[14,14]}
- ✓ das Feld ist ein Kombinationsfeld
- ✓ es trägt den Namen der Einstellung
- ✓ es steht zu, ohne Liste im Baum
- ✓ leer heißt „wie das System", mit Namen im Feld
- Bild: `/Users/david/MagicSetEditor4Mac/.huellenprobe/laeufe/2026-10-09T19-47-35Z-46912/bilder/uxrf1/zeitzone-zu.png`

**Ein Klick öffnet die Liste — der Fokus bleibt im Feld, der Text ist gewählt**

- offen: {"rolle":"combobox","name":"Zeitzone ▾","wert":"","platzhalter":"Wie das System","offen":"true","steuert":"_r_l_","haspopup":"listbox","autocomplete":"list","gesperrt":false,"listeRolle":"listbox","listeName":"Zeitzone","listeId":"_r_l_","zeilen":200,"texte":["Wie das System","Africa/Abidjan","Africa/Accra"],"alleTexte":[],"stelle":0,"aktivRolle":"option","aktivText":"Wie das System","aktivGewaehlt":"true","fokusImFeld":true,"ansage":"200 von 420 Treffern gezeigt — Suche eingrenzen …","listeImFenster":true,"listeBreit":320,"feldBreit":320,"obenAufliegend":true,"markiert":[0,0]}
- ✓ das Feld meldet aufgeklappt und steuert die Liste
- ✓ die Liste ist eine benannte listbox
- ✓ die ersten 200 von 420 Zeilen stehen da, die erste ist „wie das System"
- ✓ die hervorgehobene Zeile ist eine option und die gewählte
- ✓ der Fokus bleibt im Feld
- ✓ das Feld ist leer, die Wahl steht als Platzhalter da — {"wert":"","platzhalter":"Wie das System"}
- ✓ die Live-Region sagt, wie viele es sind und wie viele da stehen
- ✓ die Liste steht im Fenster und über allem
- ✓ sie ist so breit wie das Feld
- Bild: `/Users/david/MagicSetEditor4Mac/.huellenprobe/laeufe/2026-10-09T19-47-35Z-46912/bilder/uxrf1/zeitzone-offen.png`

**Tippen filtert — auch mit Leerzeichen für den Unterstrich**

- gefiltert: {"rolle":"combobox","name":"Zeitzone ▾","wert":"new york","platzhalter":null,"offen":"true","steuert":"_r_l_","haspopup":"listbox","autocomplete":"list","gesperrt":false,"listeRolle":"listbox","listeName":"Zeitzone","listeId":"_r_l_","zeilen":1,"texte":["America/New_York"],"alleTexte":["America/New_York"],"stelle":0,"aktivRolle":"option","aktivText":"America/New_York","aktivGewaehlt":"true","fokusImFeld":true,"ansage":"Treffer: 1","listeImFenster":true,"listeBreit":320,"feldBreit":320,"obenAufliegend":true,"markiert":[8,8]}
- ✓ „new york" findet genau America/New_York
- ✓ der erste Treffer ist hervorgehoben
- ✓ das Feld zeigt, was getippt ist
- ✓ nichts ist geschrieben, solange nicht gewählt ist
- Bild: `/Users/david/MagicSetEditor4Mac/.huellenprobe/laeufe/2026-10-09T19-47-35Z-46912/bilder/uxrf1/zeitzone-gefiltert.png`

**Enter wählt, schließt und schreibt**

- gewählt: {"rolle":"combobox","name":"Zeitzone ▾","wert":"America/New_York","platzhalter":null,"offen":"false","steuert":null,"haspopup":"listbox","autocomplete":"list","gesperrt":false,"listeRolle":null,"listeName":null,"listeId":null,"zeilen":0,"texte":[],"alleTexte":[],"stelle":-1,"aktivRolle":null,"aktivText":null,"aktivGewaehlt":null,"fokusImFeld":true,"ansage":"","listeImFenster":null,"listeBreit":null,"feldBreit":320,"obenAufliegend":null,"markiert":[16,16]}
- ✓ das Feld zeigt die Zone
- ✓ der Fokus ist im Feld geblieben

**Nur mit der Tastatur: Pfeil nach unten öffnet bei der Wahl, die Pfeile wandern**

- ✓ geöffnet steht die Hervorhebung auf der gewählten Zone
- zwei weiter: America/Noronha
- ✓ die Pfeile wandern, und der Fokus bleibt im Feld

**Escape schließt, verwirft die Eingabe und stellt den Wert her**

- ✓ der Wert steht wieder da
- ✓ nichts ist geschrieben
- Bild: `/Users/david/MagicSetEditor4Mac/.huellenprobe/laeufe/2026-10-09T19-47-35Z-46912/bilder/uxrf1/zeitzone-escape.png`

**Die letzte Zone steht hinter der Grenze — und lässt sich trotzdem finden**

- ✓ UTC steht hinter der Grenze

**„Wie das System" schreibt den leeren Wert — der Weg zurück**

- ✓ das Feld nennt „wie das System"

**Nichts passt: ein Satz, keine Liste, kein Schreiben**

- leer: Nichts passt zu „zzz“.
- ✓ die Live-Region sagt, dass nichts passt
- ✓ das Feld steht nicht als aufgeklappt da
- Bild: `/Users/david/MagicSetEditor4Mac/.huellenprobe/laeufe/2026-10-09T19-47-35Z-46912/bilder/uxrf1/zeitzone-leer.png`
- ✓ ⌘Q beendet die Anwendung

**Ein Name, den die Liste nicht kennt, bleibt sichtbar gewählt**

- ✓ US/Pacific steht nicht in der Liste
- ✓ das Feld zeigt ihn, wie er gespeichert ist
- ✓ geöffnet ist keine Zeile hervorgehoben, und die Liste steht trotzdem
- ✓ ⌘Q beendet die Anwendung (nach dem Neustart)
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht
