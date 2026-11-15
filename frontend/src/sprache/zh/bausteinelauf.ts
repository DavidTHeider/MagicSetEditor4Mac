// **The state of a file run** — FMT-03.AK-02 (#1692).
export const laufstand = {
  vonGesamt: '{erledigt} / {gesamt} {einheit}',
  einheitenVon: {
    karte: { eine: '张卡牌', viele: '张卡牌' },
    set: { eine: '个系列', viele: '个系列' },
    paket: { eine: '个资源包', viele: '个资源包' },
    bild: { eine: '张图片', viele: '张图片' },
    schritt: { eine: '个步骤', viele: '个步骤' },
  },
  einheiten: {
    karte: { eine: '张卡牌', viele: '张卡牌' },
    set: { eine: '个系列', viele: '个系列' },
    paket: { eine: '个资源包', viele: '个资源包' },
    bild: { eine: '张图片', viele: '张图片' },
    schritt: { eine: '个步骤', viele: '个步骤' },
  },
  laeuft: '正在运行…',
  abbrechen: '取消',
  wirdAngehalten: '正在停止…',
  abgebrochen: '已取消。已经完成的部分会保留——不会留下写了一半的内容。',
  bisDahin: '截至目前：{stand}。',
} as const
