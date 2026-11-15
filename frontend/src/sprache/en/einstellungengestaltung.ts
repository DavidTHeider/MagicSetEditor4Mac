// **Appearance on the settings page** — UX-09 (#1681). Mirrors the German
// catalogue part of the same name.

export const gestaltungsnamen = {
  farbwelt: 'Colour scheme',
  schriftbereich: 'Display typeface applies to',
  hintergrund: 'Window background',
  gestaltung: 'Custom colours',
}

export const gestaltungswerte = {
  'farbwelt.pergament': 'Parchment — warm, with grain',
  'farbwelt.klar': 'Clear — neutral, no grain',
  'farbwelt.schiefer': 'Slate — cool blue-grey',
  'farbwelt.moos': 'Moss — muted green',
  'schriftbereich.ueberschriften': 'Headings only',
  'schriftbereich.ganz': 'The whole interface',
  'hintergrund.aus': 'Off',
  'hintergrund.fuellen': 'Fill',
  'hintergrund.einpassen': 'Fit',
  'hintergrund.kacheln': 'Tile',
  'schriftart.buch': 'Book — New York',
  'schriftart.rund': 'Rounded — SF Pro Rounded',
  'schriftart.serifenlos': 'Sans serif — the system typeface',
}

export const gestaltungshinweise = {
  farbwelt:
    'The colours of the interface, light and dark, each one calculated. With high contrast the ink applies, whatever is chosen here.',
  schriftbereich:
    'The whole interface only takes a text face; a decorative face stays with the headings. Cards keep the typefaces of their template.',
  hintergrund:
    'Puts your own picture beneath the background, under a veil that keeps text readable. Choose the picture below; **Off** removes it at once.',
  gestaltung:
    'Custom colours for the ten properties of the interface, one per line. Hex colours only — no addresses, no fonts. Anything that would be unreadable does not apply.',
}

export const gestaltungsbeiwerk = {
  fensterbild: 'Your window background',
  vorschau: 'Preview of the window background',
  keinBild:
    'No picture of your own yet. Choose one — it shows here and, as soon as the window background is not off, behind the interface.',
  aus: 'The picture is ready, but the window background is set to **Off**.',
  waehlen: 'Choose picture …',
  entfernen: 'Remove picture',
  merkmale: 'Custom colours',
  merkmaleHilfe:
    'One entry per line, such as “--mse-marke: #4e2f16;”. Allowed are {merkmale}.',
  uebernehmen: 'Apply',
  zeilenfehler:
    'Line {zeile}: expected an entry such as “--mse-grund: #1e1b16;” — one of the allowed properties and a hex colour.',
  nichtLesbar: 'Not applied — it would not be readable:',
  paar: '{vorn} on {hinten}: {verhaeltnis} : 1, needs {schwelle} : 1',
  paarnamen: {
    'status:warnung': 'Warnings',
    'status:erfolg': 'Success messages',
    'fensterbild:dunkel': '--mse-grund over a dark window image',
    'fensterbild:hell': '--mse-grund over a light window image',
  } as Record<string, string>,
  angewandt: 'Applied.',
  ruhtBeiKontrast:
    'Saved — custom colours do not apply while high contrast is on; they return once it is off.',
  schmuckNurOben:
    'The chosen display typeface is decorative — it stays with the headings.',
  zuruecksetzenGescheitert: 'The appearance could not be reset:',
  zurVorgabe: 'Reset your own colours to the default',
}
