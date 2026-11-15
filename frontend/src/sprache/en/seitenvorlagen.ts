// Die Texte der **Paketwerkstatt** (#1164) — eine Vorlage öffnen, ändern,
// speichern.
//
// **Eine eigene Datei**, weil `seiten.ts` an der Zeilengrenze aus §2 steht.
// Ein Katalog darf sich auf mehrere Dateien verteilen, solange er als **einer**
// ankommt (§9).
export const vorlagen = {
  titel: 'Templates',
  erklaerung:
    'What a card can do and how it looks lives in text files inside the template package — fields, scripts, measurements. They can be changed here. The previous state stays as a backup next to the package.',
  paket: 'Package',
  keinPaket: 'No package chosen yet.',
  waehleEines:
    'Pick a package above to see and change its text files — fields, scripts, measurements.',
  keineDateien: 'This package has no file to edit.',
  dateien: 'Files',
  datei: 'File',
  inhalt: 'Content',
  speichern: 'Save',
  gespeichert: 'Saved.',
  ungespeichert: 'Not saved.',
  warnung:
    'This file applies to every set that uses this package — not just the open one.',
  konflikt:
    'The file changed since it was opened. Load their version — your text is lost — or keep yours: the changes then show what you overwrite in theirs.',
  neuLaden: 'Load their version',
  behalten: 'Keep mine',
  vorschau: 'Preview',
  vorschauErklaerung:
    'The first card of the first set that uses this template — drawn with the text shown here. Nothing is saved in the process.',
  vorschauLaeuft: 'Drawing …',
  vorschauKeineKarte:
    'No open set uses this template. Open one under “Sets” that names it — then you can see here what a change does.',

  arbeitsplatz: 'Template workshop',
  baumBezeichnung: 'Packages and files',
  inspectorBezeichnung: 'Preview and diagnosis',
  baumSuche: 'Search packages and files',
  baumSuchePlatzhalter: 'Package or file …',
  baumLeer: 'No package and no file matches the search.',
  baumOhnePakete:
    'No package to edit yet. Create one under “Packages” or fetch one — then it shows up here.',
  nurLesbar: 'Bundled — read-only. Your own copy can be edited.',
  baumNurMitgeliefert:
    'Every package here is bundled. Create your own copy of one, or a new one under “Packages” — both can be edited and show up at the top.',
  dateienLaden: 'Reading files …',
  inspectorLeer:
    'The preview, diagnostics and origin of the open file appear here.',
  waehleImBaum:
    'Pick a file in the tree on the left — fields, scripts and measurements live there as text and can be changed here.',
  verwerfenFrage:
    'The open file has unsaved changes. Discard them and open {datei}?',
  verwerfen: 'Discard and open',
  hierBleiben: 'Stay here',
  testkarte: 'Test card',
  testkarteVorgabe: 'First card of the first set',
  testkarteOhneNamen: 'Card {nummer}',
  testkarteErklaerung:
    'A template rarely breaks on the first card. The example set “Werkstatt-Testkarten” brings long text, a special layout, a second face, a missing image and special characters — open it under “Sets” and switch its style to this package.',
  veraltet:
    'Stale — the file has a finding in line {zeile}; showing the last valid state.',
  wirdNeuGezeichnet: 'Redrawing …',
  diagnose: 'Diagnosis',
  diagnoseFrei: 'No finding in this file.',
  feldartFehlt:
    'This field has no “type:”. MSE2 stops reading the package here; this app treats it as a text field.',
  feldartUnbekannt:
    'MSE2 does not know the field type “{0}” and stops reading the package here. This app treats the field as text. Known types: {1}.',
  diagnoseStelle: 'Line {zeile}, column {spalte}',
  zurStelle: 'Go to position',
  herkunft: 'Where a name comes from',
  herkunftKeine:
    'Put the caret on a name — then this shows where it comes from and who uses it.',
  herkunftUnbekannt:
    '“{name}” is not defined by this package — the name comes from the library or from an included package.',
  herkunftStelle: '{art} in {datei}, line {zeile}',
  zurDefinition: 'Go to definition',
  verwendungen: 'References',
  verwendungenKeine: 'No other place names this.',
  verwendungLaden: 'Looking for references …',
  symbolarten: {
    'card field': 'Card field',
    'set field': 'Set field',
    'styling field': 'Styling option',
    'extra card field': 'Extra field',
    script: 'Script name',
  } as Record<string, string>,
  symbolartSonst: 'Defined here',
  aenderungen: 'Changes before saving',
  aenderungenZeigen: 'Show changes',
  aenderungenVerbergen: 'Hide changes',
  aenderungenKeine: 'No change compared to the state that was read.',
  aenderungenZusammenfassung: '{neu} line(s) added, {weg} removed.',
  aenderungenAbschnitt: 'Change {nummer}',
  zeileWeg: 'removed',
  zeileNeu: 'added',
  kopieHinweis: 'Origin of the package',
  kopieren: 'Create own copy',
  kopieMitgeliefert:
    'This package ships with the app and cannot be changed. Create your own copy — it then appears among your packages, and the original stays as it is.',
  kopieFremd:
    'This package comes from elsewhere. A change here counts as a local change and competes with the new version on the next download. Create your own copy if it should last.',
  kopieOriginal:
    'The original stays. You are working on a copy — {original} itself is unchanged.',
  dateiaktionen: 'Package files',
  dateiNeu: 'New file',
  dateiName: 'Name of the new file',
  dateiAnlegen: 'Create',
  dateiAbbrechen: 'Cancel',
  dateiHinzufuegen: 'Add image or font …',
  dateiUmbenennen: 'Rename',
  dateiNeuerName: 'New name',
  dateiUmbenennenJa: 'Rename',
  dateiLoeschen: 'Delete …',
  dateiLoeschenFrage:
    'Delete? The file then lies in the backup next to the package.',
  dateiLoeschenJa: 'Yes, delete',
  dateiErstSichern: 'Save or discard first.',
  beiwerk: 'Asset',
  beiwerkHinweis:
    'Not text — the package uses this file as it is. Rename and delete it here.',
  beiwerkZeile: 'Assets ({anzahl})',
  weitereZeigen: 'Show more ({gezeigt} of {gesamt})',
} as const
