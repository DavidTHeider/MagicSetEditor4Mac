# Beleg: was MSE2 mit Zusatzdaten macht

**Gemessen** am 2026-09-29 20:07 UTC mit `scripts/mse2-zusatzdaten.sh`, Mac17,3, macOS 27.0, Commit `1e9da894` — FMT-05.AK-02 (#1694).

```
MSE2: f9d9356d vom 2021-07-06

── Öffnen (Meldungen von MSE2)
         On line 6: 	Unexpected key: 'mse4mac_preisstand'
         On line 10: 	Unexpected key: 'mse4mac_exemplar'

── Nach dem Speichern durch MSE2
  verloren: mse4mac preisstand
  verloren: mse4mac exemplar
  bleibt:   notes: eine Notiz
  die Datei daneben liegt unverändert da, und MSE2 nahm sie nicht mit: card eins set 
```

**Gelesen**: eingebettete Felder meldet MSE2 beim Öffnen als unerwarteten Schlüssel (`Reader::unknownKey`, `src/util/io/reader.cpp` Zeilen 231–246, Fassung f9d9356 vom 2021-07-06) und schreibt sie beim Speichern nicht zurück; das MSE2-eigene `notes` bleibt. Eine Datei daneben fasst MSE2 weder an noch nimmt es sie mit.
