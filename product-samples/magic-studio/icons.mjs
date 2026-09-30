const paths = {
 search: '<circle cx="10.8" cy="10.8" r="6.8"/><path d="m16 16 4.5 4.5"/>',
 image: '<rect x="3" y="3" width="18" height="18" rx="3"/><circle cx="8" cy="8" r="1.5"/><path d="m3 16 5-5 4 4 3-3 6 6"/>',
 video: '<rect x="2.5" y="5" width="14" height="14" rx="3"/><path d="m16.5 10 5-3v10l-5-3"/>',
 spark: '<path d="m12 3 2.5 6.5L21 12l-6.5 2.5L12 21l-2.5-6.5L3 12l6.5-2.5L12 3Z"/>',
 bookmark: '<path d="M6 3h12v18l-6-4-6 4V3Z"/>',
 history: '<path d="M3 11a9 9 0 1 1 2.7 7M3 4v7h7m2-4v5l3 2"/>',
 sun: '<circle cx="12" cy="12" r="4"/><path d="M12 2v2m0 16v2M2 12h2m16 0h2M5 5l1.4 1.4m11.2 11.2L19 19M5 19l1.4-1.4M17.6 6.4 19 5"/>',
 moon: '<path d="M20.7 13.3A9 9 0 0 1 10.7 3 9 9 0 1 0 20.7 13.3Z"/>',
 arrow: '<path d="M5 12h14m-6-6 6 6-6 6"/>',
 copy: '<rect x="8" y="8" width="13" height="13" rx="2"/><path d="M16 8V3H3v13h5"/>',
 check: '<path d="m4 12 5 5L20 6"/>',
 close: '<path d="m6 6 12 12M6 18 18 6"/>',
 external: '<path d="M14 3h7v7m0-7L10 14m0-11H3v18h18v-7"/>',
 download: '<path d="M12 3v12m-5-5 5 5 5-5M4 16v5h16v-5"/>',
 play: '<path d="m8 4 12 8-12 8V4Z"/>',
 pause: '<path d="M8 4v16M16 4v16"/>',
 sliders: '<path d="M4 7h7m5 0h4M4 17h2m5 0h9"/><circle cx="13" cy="7" r="2"/><circle cx="8" cy="17" r="2"/>',
 back: '<path d="M19 12H5m6-6-6 6 6 6"/>',
 book: '<path d="M3 4h7l2 2 2-2h7v16h-7l-2 2-2-2H3V4Zm9 2v16"/>',
 code: '<path d="m8 5-6 7 6 7m8-14 6 7-6 7M14 3l-4 18"/>'
};
export function icon(name) {
 const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
 svg.setAttribute('viewBox', '0 0 24 24');
 svg.setAttribute('fill', 'none'); svg.setAttribute('stroke', 'currentColor');
 svg.setAttribute('stroke-width', '1.7'); svg.setAttribute('stroke-linecap', 'round'); svg.setAttribute('stroke-linejoin', 'round');
 svg.setAttribute('aria-hidden', 'true'); svg.innerHTML = paths[name] || paths.spark;
 return svg;
}
