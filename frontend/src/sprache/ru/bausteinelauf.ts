// **The state of a file run** — FMT-03.AK-02 (#1692).
export const laufstand = {
  vonGesamt: '{erledigt} из {gesamt} {einheit}',
  einheitenVon: {
    karte: { eine: 'карты', wenige: 'карт', viele: 'карт' },
    set: { eine: 'набора', wenige: 'наборов', viele: 'наборов' },
    paket: { eine: 'пакета', wenige: 'пакетов', viele: 'пакетов' },
    bild: { eine: 'изображения', wenige: 'изображений', viele: 'изображений' },
    schritt: { eine: 'шага', wenige: 'шагов', viele: 'шагов' },
  },
  einheiten: {
    karte: { eine: 'карта', wenige: 'карты', viele: 'карт' },
    set: { eine: 'набор', wenige: 'набора', viele: 'наборов' },
    paket: { eine: 'пакет', wenige: 'пакета', viele: 'пакетов' },
    bild: { eine: 'изображение', wenige: 'изображения', viele: 'изображений' },
    schritt: { eine: 'шаг', wenige: 'шага', viele: 'шагов' },
  },
  laeuft: 'Выполняется …',
  abbrechen: 'Отменить',
  wirdAngehalten: 'Остановка …',
  abgebrochen:
    'Отменено. Всё, что уже было готово, сохраняется — наполовину записанного не остаётся.',
  bisDahin: 'К этому моменту: {stand}.',
} as const
