// 程序动效配方：每段指定来源原图、输出规格与各类局部动态。区域为 0–1 归一化坐标，按原图人工标定。
// 改动效果或区域时提高 version；入库工具会核对录制时的版本，避免把旧配方录的片段当成新版本。
const rect = (x1, y1, x2, y2) => [[x1, y1], [x2, y1], [x2, y2], [x1, y2]];

export const motionRecipes = {
 'clip-pixel-night': {
  version: 1, source: 'pixel-night', art: 'pixel-night', width: 768, height: 512, fps: 30, seconds: 5, bitrate: 2200000, seed: 20261021,
  motion: '像素港口：窗灯、灯笼与灯塔明暗起伏，水面倒影随波纹晃动，星星闪烁，月光轻微呼吸；人物、渡船与码头不动。',
  // 一整块水面：码头前沿以下（避开左侧石阶与底部木板）、渡船船头以右、右侧栈桥以上；小船、鹈鹕与前景大木桩人工排除，其余木桩按颜色排除。
  water: {
   regions: [[[0.137, 0.668], [0.781, 0.645], [0.781, 0.6], [0.823, 0.56], [0.823, 0.395], [1, 0.395], [1, 0.773], [0.79, 0.773], [0.775, 0.82], [0.775, 0.98], [0.391, 0.98], [0.137, 0.844]]],
   edge: 6,
   exclude: [rect(0.135, 0.755, 0.345, 0.87), rect(0.866, 0.55, 0.932, 0.8), rect(0.425, 0.795, 0.49, 0.985)],
   amplitude: 2, wavelength: 9, cycles: 4, span: 90, drift: 1, shimmer: 0.14, shimmerCycles: 5,
  },
  lights: { regions: [rect(0, 0, 1, 0.645)], amplitude: [0.08, 0.18], spill: 4, spillStrength: 0.5 },
  beacons: [{ x: 0.94, y: 0.242, radius: 0.02, cycles: 2, amplitude: 0.45 }],
  stars: { regions: [rect(0, 0, 1, 0.235)], minLum: 150, maxSize: 14, amplitude: [0.3, 0.55] },
  moon: { x: 0.794, y: 0.088, radius: 0.035, amplitude: 0.06 },
 },
};
