// Animated background: a subtle particle network (mint/sky/violet) with pointer
// repulsion, matching the brand treatment of our other projects.
(() => {
  "use strict";

  const canvas = document.getElementById("bg-particles");
  if (!canvas) return;
  const ctx = canvas.getContext("2d");
  if (!ctx) return;

  const reduce = window.matchMedia("(prefers-reduced-motion: reduce)");
  const LINK = 130;
  const POINTER = 150;

  let w = 0;
  let h = 0;
  let particles = [];
  let palette = [];
  let frame = 0;
  let running = false;
  let last = 0;
  const pointer = { x: -9999, y: -9999 };

  function toRgb(value) {
    let hex = (value || "").trim().replace("#", "");
    if (hex.length === 3) hex = hex.split("").map((c) => c + c).join("");
    if (hex.length !== 6) return null;
    const n = parseInt(hex, 16);
    if (Number.isNaN(n)) return null;
    return { r: (n >> 16) & 255, g: (n >> 8) & 255, b: n & 255 };
  }

  function readPalette() {
    const styles = getComputedStyle(document.documentElement);
    const colors = ["--mint", "--sky", "--violet"]
      .map((name) => toRgb(styles.getPropertyValue(name)))
      .filter(Boolean);
    palette = colors.length ? colors : [{ r: 0, g: 229, b: 168 }];
  }

  function count() {
    return Math.max(24, Math.min(80, Math.round((w * h) / 22000)));
  }

  function rnd(a, b) {
    return a + Math.random() * (b - a);
  }

  function make() {
    return {
      x: Math.random() * w,
      y: Math.random() * h,
      vx: rnd(-10, 10) || 0.5,
      vy: rnd(-10, 10) || 0.5,
      r: rnd(0.7, 1.8),
      c: palette[Math.floor(Math.random() * palette.length)],
    };
  }

  function seed() {
    readPalette();
    particles = Array.from({ length: count() }, make);
  }

  function resize() {
    const dpr = Math.min(window.devicePixelRatio || 1, 2);
    w = window.innerWidth;
    h = window.innerHeight;
    canvas.width = Math.round(w * dpr);
    canvas.height = Math.round(h * dpr);
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    seed();
    if (reduce.matches) draw();
  }

  function update(dt) {
    for (const p of particles) {
      p.x += p.vx * dt;
      p.y += p.vy * dt;

      const dx = p.x - pointer.x;
      const dy = p.y - pointer.y;
      const d = Math.hypot(dx, dy);
      if (d > 0 && d < POINTER) {
        const f = (1 - d / POINTER) * 26;
        p.x += (dx / d) * f * dt;
        p.y += (dy / d) * f * dt;
      }

      if (p.x < -20) p.x = w + 20;
      if (p.x > w + 20) p.x = -20;
      if (p.y < -20) p.y = h + 20;
      if (p.y > h + 20) p.y = -20;
    }
  }

  function draw() {
    ctx.clearRect(0, 0, w, h);

    for (let i = 0; i < particles.length; i += 1) {
      const a = particles[i];
      for (let j = i + 1; j < particles.length; j += 1) {
        const b = particles[j];
        const dx = a.x - b.x;
        const dy = a.y - b.y;
        const d = Math.hypot(dx, dy);
        if (d > LINK) continue;
        const alpha = (1 - d / LINK) * 0.16;
        ctx.strokeStyle = `rgba(${b.c.r}, ${b.c.g}, ${b.c.b}, ${alpha})`;
        ctx.lineWidth = 0.7;
        ctx.beginPath();
        ctx.moveTo(a.x, a.y);
        ctx.lineTo(b.x, b.y);
        ctx.stroke();
      }
    }

    for (const p of particles) {
      ctx.fillStyle = `rgba(${p.c.r}, ${p.c.g}, ${p.c.b}, 0.6)`;
      ctx.beginPath();
      ctx.arc(p.x, p.y, p.r, 0, Math.PI * 2);
      ctx.fill();
    }
  }

  function loop(now) {
    if (!running) return;
    const dt = Math.min((now - last) / 1000, 0.05) || 0;
    last = now;
    update(dt);
    draw();
    frame = requestAnimationFrame(loop);
  }

  function start() {
    if (running || reduce.matches || document.hidden) return;
    running = true;
    last = performance.now();
    frame = requestAnimationFrame(loop);
  }

  function stop() {
    running = false;
    if (frame) cancelAnimationFrame(frame);
    frame = 0;
  }

  function sync() {
    if (reduce.matches) {
      stop();
      draw();
      return;
    }
    if (document.hidden) {
      stop();
      return;
    }
    start();
  }

  function move(e) {
    if (e.pointerType === "touch") return;
    pointer.x = e.clientX;
    pointer.y = e.clientY;
  }

  function clear() {
    pointer.x = -9999;
    pointer.y = -9999;
  }

  window.addEventListener("resize", resize, { passive: true });
  window.addEventListener("pointermove", move, { passive: true });
  document.addEventListener("pointerleave", clear, { passive: true });
  document.addEventListener("visibilitychange", sync);
  if (typeof reduce.addEventListener === "function") reduce.addEventListener("change", sync);

  resize();
  sync();
})();
