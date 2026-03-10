// ── Copy install command ──────────────────────────
function copyInstall() {
  navigator.clipboard.writeText('go install github.com/torryt/riff@latest');
  const tip = document.getElementById('copyTooltip');
  tip.classList.add('show');
  setTimeout(() => tip.classList.remove('show'), 1500);
}

// ── Terminal content definitions ──────────────────
const terminals = {
  'hero': [
    { type: 'cmd',    html: '<span class="prompt-char">$</span> <span class="cmd">riff</span> <span class="flag">new react</span>' },
    { type: 'output', html: '<span class="highlight">Created</span> project <span class="id">a7x9k2m</span> with template react' },
    { type: 'output', html: '<span class="output">Initialized git repository</span>' },
    { type: 'blank' },
    { type: 'cmd',    html: '<span class="prompt-char">$</span> <span class="cmd">riff</span> <span class="flag">list</span>' },
    { type: 'output', html: '<span class="id">a7x9k2m</span>  <span class="highlight">react</span>   <span class="desc">realtime chat app with websocket hooks</span>' },
    { type: 'output', html: '<span class="id">p3nq8f1</span>  <span class="highlight">rust</span>    <span class="desc">cli tool for parsing csv to json</span>' },
    { type: 'output', html: '<span class="id">w2kd5t9</span>  <span class="highlight">python</span>  <span class="desc">matplotlib dashboard for sensor data</span>' },
    { type: 'blank' },
    { type: 'cmd',    html: '<span class="prompt-char">$</span> <span class="cmd">riff</span> <span class="flag">open a7x9k2m</span>' },
    { type: 'output', html: '<span class="output">Jumped to</span> <span class="id">~/.riff/projects/a7x9k2m/</span>' },
    { type: 'cursor' },
  ],
  'step-new': [
    { type: 'cmd',    html: '<span class="prompt-char">$</span> <span class="cmd">riff</span> <span class="flag">new react</span>' },
    { type: 'output', html: '<span class="highlight">Created</span> project <span class="id">a7x9k2m</span> with template react' },
    { type: 'output', html: '<span class="output">Initialized git repository</span>' },
    { type: 'output', html: '<span class="output">Running:</span> bunx create-vite . --template react-ts' },
    { type: 'blank' },
    { type: 'output', html: '<span class="highlight">Done!</span> <span class="output">Project ready at</span> <span class="id">~/.riff/projects/a7x9k2m/</span>' },
    { type: 'cursor' },
  ],
  'step-list': [
    { type: 'cmd',    html: '<span class="prompt-char">$</span> <span class="cmd">riff</span> <span class="flag">list</span>' },
    { type: 'blank' },
    { type: 'output', html: '<span class="id">a7x9k2m</span>  <span class="highlight">react</span>   <span class="desc">realtime chat app with websocket hooks</span>' },
    { type: 'output', html: '<span class="id">p3nq8f1</span>  <span class="highlight">rust</span>    <span class="desc">cli tool for parsing csv to json</span>' },
    { type: 'output', html: '<span class="id">w2kd5t9</span>  <span class="highlight">python</span>  <span class="desc">matplotlib dashboard for sensor data</span>' },
    { type: 'output', html: '<span class="id">k9m2x4j</span>  <span class="highlight">go</span>      <span class="desc">http proxy with request logging</span>' },
    { type: 'blank' },
    { type: 'output', html: '<span class="output">4 projects</span>' },
    { type: 'cursor' },
  ],
  'step-export': [
    { type: 'cmd',    html: '<span class="prompt-char">$</span> <span class="cmd">riff</span> <span class="flag">export a7x9k2m</span> <span class="output">~/projects/chat-app</span>' },
    { type: 'blank' },
    { type: 'output', html: '<span class="highlight">Exported</span> <span class="id">a7x9k2m</span> <span class="output">to</span> ~/projects/chat-app' },
    { type: 'output', html: '<span class="output">Git history preserved. Go ship it.</span>' },
    { type: 'cursor' },
  ],
  'cta-install': [
    { type: 'cmd',    html: '<span class="prompt-char">$</span> <span class="cmd">go install</span> <span class="output">github.com/torryt/riff@latest</span>' },
    { type: 'blank' },
    { type: 'cmd',    html: '<span class="prompt-char">$</span> <span class="cmd">riff</span> <span class="flag">new python</span>' },
    { type: 'output', html: '<span class="highlight">Created</span> project <span class="id">f4kq7n2</span> with template python' },
    { type: 'output', html: '<span class="output">Initialized git repository</span>' },
    { type: 'output', html: '<span class="output">Running:</span> uv init' },
    { type: 'blank' },
    { type: 'output', html: '<span class="highlight">Done!</span> <span class="output">Happy hacking.</span>' },
    { type: 'cursor' },
  ],
};

// ── Render a terminal by data-terminal key ────────
function renderTerminal(el) {
  const key = el.getAttribute('data-terminal');
  const lines = terminals[key];
  if (!lines) return;

  const body = el.querySelector('.terminal-body');
  if (!body || body.children.length > 0) return; // already rendered

  lines.forEach((line, i) => {
    const div = document.createElement('div');
    div.classList.add('line');
    div.style.animationDelay = `${0.15 + i * 0.1}s`;

    if (line.type === 'blank') {
      div.innerHTML = '&nbsp;';
    } else if (line.type === 'cursor') {
      div.innerHTML = '<span class="prompt-char">$</span> <span class="cursor"></span>';
    } else {
      div.innerHTML = line.html;
    }

    body.appendChild(div);
  });
}

// ── Observe all terminals for visibility ──────────
const termObserver = new IntersectionObserver((entries) => {
  entries.forEach(e => {
    if (e.isIntersecting) {
      renderTerminal(e.target);
      termObserver.unobserve(e.target);
    }
  });
}, { threshold: 0.15 });

document.querySelectorAll('[data-terminal]').forEach(el => {
  termObserver.observe(el);
});

// ── Scroll reveal ─────────────────────────────────
const revealObserver = new IntersectionObserver((entries) => {
  entries.forEach(e => {
    if (e.isIntersecting) {
      e.target.classList.add('visible');
    }
  });
}, { threshold: 0.1, rootMargin: '0px 0px -40px 0px' });

document.querySelectorAll('.reveal').forEach(el => revealObserver.observe(el));
