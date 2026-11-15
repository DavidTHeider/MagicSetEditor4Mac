// The texts for **database maintenance** — back up and restore
// (#940).
//
// Split out of `seiten.ts` the same way the card page was: one catalogue may
// live in several files as long as it arrives as **one** (§9).

export const bestandspflege = {
  titel: 'Database',
  hinweis:
    'Everything this application keeps: sets, cards, field values, the ' +
    'history and the images you have placed. A backup carries both in one ' +
    'file — the database and the images.',
  sichern: 'Back up database …',
  waehlen: 'Restore a backup …',
  rueckfrage: 'Restore this backup?',
  datei: 'File',
  angelegtAm: 'Database created',
  sets: 'Sets',
  karten: 'Cards',
  unbekannt: '—',
  warnung:
    'The current database will be replaced. It is not lost: it stays ' +
    'alongside as bestand-vor-rueckspielung.db — until the next restore.',
  bilderLage: 'Placed images',
  bilderErsetzt:
    'This backup carries images. They take the place of the current ones; ' +
    'those stay alongside as beiwerk-vorher.',
  bilderGeleert:
    'This backup carries no images — it comes from a state where none were ' +
    'placed. The current ones are removed; they stay alongside as ' +
    'beiwerk-vorher.',
  bilderUnberuehrt:
    'This backup is a single database from 0.1.0 or 0.2.0 and carries no ' +
    'images. The current ones stay untouched — on another machine they would ' +
    'be missing entirely.',
  bestaetigen: 'Restore',
  abbrechen: 'Cancel',

  // **The password of a backup** (OPS-06, #1723).
  kennwort: 'Password (optional)',
  kennwortWiederholen: 'Repeat password',
  kennwortHinweis:
    'With a password the backup is encrypted and can be opened on another computer too. A forgotten password cannot be recovered — the backup is then lost.',
  kennwortVerschieden: 'The two passwords do not match.',
  klartext: 'Back up without password — the file is then plain text',
  klartextHinweis:
    'The collection is encrypted. Without a password the backup would lie unprotected on the medium it goes to.',
  gesperrt: 'This backup is encrypted.',
  gesperrtKennwort: 'Password of the backup',
  entsperren: 'Open',
  verschluesselt: 'Encrypted',
  ja: 'yes',
  nein: 'no',
  paketeEnthalten: 'Own packages inside',
  paketeFehlen: 'To be fetched again',
  keine: 'none',
  zurueckgespielt: 'Restored. The previous database is now {beiseite}.',
  rueckkehrUnveraendert: 'Your previous database is unchanged.',
  rueckkehrUnvollstaendig:
    'Undoing it failed as well: parts of the previous database lie next to it as bestand-vor-rueckspielung.db and beiwerk-vorher. Restore a backup or move them back by hand.',
  wiederherstellungTitel: 'The database could not be opened',
  wiederherstellungHinweis:
    'The application will work again once a usable database file is in ' +
    'place. Restoring a backup is the way there — it replaces the file ' +
    'that is currently failing.',
  wiederherstellungGrund: 'What was reported',
  wiederherstellungHinweisHeil:
    'The library file is probably fine — something around it is the problem. Fix it as described below and try again; there is nothing to replace.',
  erneutVersuchen: 'Try again',
  weiterGescheitert: 'Tried again — the library still could not be opened.',

  wegZurueckTitel: 'The way back',
  wegZurueck:
    'To get back, use \u201c{datei}\u201d in your library\u2019s folder \u2014 ' +
    '{sets} set(s), {karten} cards, schema step {stand}. Pick it below.',

  chiffreTitel: 'Library encryption',
  chiffreAnHinweis:
    'The library is stored encrypted on disk. Without the key it cannot be ' +
    'opened \u2014 not even by you.',
  chiffreAusHinweis:
    'The library is stored in the clear. Anyone holding the file can read ' +
    'everything in it: names, dates of birth, addresses, photos and sign-in ' +
    'methods.',
  chiffreAnschalten: 'Turn encryption on \u2026',
  chiffreAusschalten: 'Turn encryption off \u2026',
  chiffreFrageAn: 'Encrypt the library now?',
  chiffreFrageAus: 'Turn encryption back off?',
  chiffreWarnungAn:
    'The key is created now and goes into this Mac\u2019s keychain. It cannot ' +
    'be recomputed and exists nowhere else: if it is gone, so is the data. ' +
    'Make a backup first.',
  chiffreWarnungAus:
    'Everything will be stored in the clear again, and the key is removed ' +
    'from the keychain.',
  chiffreGrenzen:
    'Encryption does **not** protect against tampering. Anyone comparing ' +
    'several copies of the same file can see which parts changed. ' +
    '**A backup gets its own password** — above, when backing up; without ' +
    'one it would be plain text, and that then needs an explicit “back up ' +
    'without password”.',
  chiffreLaeuft: 'Rewriting the library \u2026',

  platzUeberschrift: 'Disk space',
  platzErklaerung:
    'What this application stores and how much it weighs \u2014 the whole folder, since anything no item claims is counted under **Other**. **Backups before migrations** are made on every start that rebuilds the library; what stays is the **way back** \u2014 the most recent backup this version can open. Of the **log**, the running file stays.',
  platzDateien: '{anzahl} file(s)',
  platzFreigeben: 'Clear older ones',
  platzFreigebenFuer: '{posten}: clear the older ones',
  platzFreigebenFrage: '{posten}: really clear the older ones?',
  platzFreigebenErklaerung:
    'The older files are deleted and cannot be brought back. What is needed as the way back stays.',
  platzLoeschenJa: 'Clear',
  platzNamen: {
    bestand: 'Library',
    bilder: 'Embedded images',
    pakete: 'Downloaded packages',
    vorsicherungen: 'Backups before migrations',
    protokoll: 'Log',
    sonstiges: 'Other files in the application folder',
  },
  alter: {
    ueberschrift: 'Last backup',
    unbekannt: 'unknown',
    unbekanntSatz:
      'This version has never backed up the library. Unknown means unknown — this never says “today” unless a file was read back.',
    heute: 'today',
    gestern: 'a day ago',
    vorTagen: '{tage} days ago',
    alt: '{alter} · older than 14 days',
    unstimmig: 'time after the clock',
    satz: 'Last backed up on {zeitpunkt} via {flaeche} to {datei}, read back and checked.',
    satzPlan:
      'Last backed up on {zeitpunkt} by the plan to {datei}, read back and checked.',
    unstimmigSatz:
      'The last backup is dated {zeitpunkt} — a time after this Mac’s clock. Its age can’t be told; back up to be sure.',
    flaechen: {
      fenster: 'the window',
      cli: 'the command line',
      mcp: 'the tool server',
      rest: 'the service',
    },
    planSchalter: 'Back up weekly to a folder, keep seven',
    planHilfe:
      'Only after you choose a folder. Only its own backups are deleted, never other files. It backs up while the window is open.',
    planOrdner: 'Folder: {ordner}',
    planZuletzt: 'Last backed up by the plan on {zeitpunkt}.',
    planNie: 'Never backed up by the plan yet.',
    planFehler: 'The plan could not back up last time: {grund}',
  },
} as const
