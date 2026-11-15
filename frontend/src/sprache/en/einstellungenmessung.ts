// **The measurement settings** — OPS-05
// (#1722).
export const messnamen = {
  'messung.spur': 'Trace',
  'messung.abtastung': 'Share of calls recorded, in percent',
  'messung.spurgrenze': 'At most this many entries',
  'messung.aufbewahrung': 'Entries stay this many minutes',
  'messung.profile': 'Performance profiles kept',
  'zwischenspeicher.grenze': 'Limit per disk cache in MB',
} as const

export const messwerte = {
  'messung.spur.aus': 'Off',
  'messung.spur.an': 'On',
} as const

export const messhinweise = {
  'messung.spur':
    'One line per call to the service and the tool server — with operation id, route, duration and outcome, without contents. It lives in the memory of the process and applies from the next start.',
  'messung.abtastung':
    'At 10 the trace records every tenth call — counted exactly, not at random. Empty means 100.',
  'messung.spurgrenze':
    'When the trace is full, the oldest entry goes. 10 to 10000, empty means 1000.',
  'messung.aufbewahrung':
    'Older entries drop out. 1 to 1440 minutes, empty means 60.',
  'messung.profile':
    'The newest stay next to the store; older ones go with the next profile. 1 to 20, empty means 3.',
  'zwischenspeicher.grenze':
    'For Scryfall answers and images, each on its own. At start the oldest go until it fits. 1 to 8192, empty means 512.',
} as const
