// 工坊短视频计划：每段以已有工坊图片作为首帧，只描述动态，保持原图构图、配色与画风。
// 生成由 tools/generate-videos.mjs 在用户本机、用户自己的 Key 下显式运行；本文件只是计划，不代表已生成。
// 选图原则：一种视觉风格一段、动态自然、不含写实人脸；提示词为英文以便视频模型理解，motion 为中文展示说明。
const keep = 'Keep the exact composition, framing, color palette and art style of the source image. No new characters, objects, text, logos or watermarks. Smooth seamless motion, no flicker, no sudden cuts.';

export const VIDEO_DEFAULTS = Object.freeze({ model: 'grok-imagine-video', duration: 5, resolution: '480p' });

export const videoPlan = Object.freeze([
 { id: 'clip-pixel-night', source: 'pixel-night', motion: '像素港口：窗灯与灯笼轻轻闪烁，水面倒影泛起像素波纹，渡船随浪微微起伏，镜头固定。',
  prompt: 'Subtle pixel-art animation of this night harbor. Warm window lights and lanterns gently flicker, the water reflections ripple in crisp stepped pixel waves, the mint-green ferry bobs slightly on the swell, a few stars twinkle. Locked-off camera. ' + keep },
 { id: 'clip-ink-koi', source: 'ink-koi', motion: '水墨锦鲤：鲤鱼沿弧线缓慢游动，尾鳍摆动，墨雾在山间流淌、在宣纸上晕开。',
  prompt: 'Gentle ink-wash animation. The koi glides slowly along its curved path with a soft flick of its tail fin, misty ink drifts and diffuses across the mountains like wet ink spreading on xuan paper. Very slow, calm motion, static camera. ' + keep },
 { id: 'clip-watercolor-island', source: 'watercolor-island', motion: '水彩海湾：浪花在岸边轻轻翻涌，橙色船帆缓慢漂移，云朵掠过，花枝随风轻摆。',
  prompt: 'Soft watercolor animation of this bay. Small waves lap along the curving shore, the orange sail drifts slowly across the water, light clouds glide overhead, bougainvillea sways in a light breeze, with a subtle painted-paper texture shimmer. Static elevated camera. ' + keep },
 { id: 'clip-papercut-tiger', source: 'papercut-tiger', motion: '剪纸老虎：纸层间轻微视差，小老虎迈出一小步并轻摆尾巴，纸叶微微颤动。',
  prompt: 'Layered paper-cut stop-motion animation. The paper layers shift with a slight parallax, the tiger takes one small step forward and gently swishes its curved tail, paper leaves tremble as if moved by a soft breeze. Handcrafted stop-motion feel, nearly static camera. ' + keep },
 { id: 'clip-red-panda-moss', source: 'red-panda-moss', motion: '苔藓小熊猫：小熊猫沿树枝慢慢走一步，尾巴摆动，薄雾与光斑在林间移动。',
  prompt: 'Natural wildlife footage feel. The red panda takes a slow careful step along the mossy branch, its long ringed tail sways, soft mist drifts between the trees and dappled light moves gently over the moss. Shallow depth of field, very slight handheld drift. ' + keep },
 { id: 'clip-anime-rain-tram', source: 'anime-rain-tram', motion: '雨后电车：屋檐滴水，积水倒影荡开，电车灯光闪动，人物衣角随风轻摆。',
  prompt: 'Hand-painted 2D anime animation after rain. Drops fall from the eaves, puddle reflections ripple, the green tram\'s lights flicker softly, the photographer\'s raincoat hem moves slightly in the breeze while she stays in the same pose, a light mist drifts. Static camera. ' + keep },
 { id: 'clip-lantern-alley', source: 'lantern-alley', motion: '狸猫食堂：灯笼轻轻摇晃，拉面碗冒出热气，暖光在木窗上缓慢流动。',
  prompt: 'Cozy miniature stop-motion style animation. The orange lanterns sway gently, steam rises softly from the ramen bowl, warm light flickers over the wooden windows, the tanuki chef makes a small friendly nod. Soft focus, nearly static camera. ' + keep },
 { id: 'clip-jade-product', source: 'jade-product', motion: '玉色瓶身：一道柔和高光缓慢扫过磨砂瓶身，阴影随光轻移，瓶子保持静止。',
  prompt: 'Premium product commercial shot. A soft studio highlight slowly sweeps across the frosted jade-green glass bottle and the brushed metal cap, shadows on the limestone plinth shift gently with the light. The bottle itself stays still and undistorted. Locked-off camera. ' + keep },
]);
