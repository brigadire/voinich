(() => {
  const toggle = document.querySelector('.nav-toggle');
  const links = document.querySelector('.nav-links');
  if (toggle && links) {
    toggle.addEventListener('click', () => {
      const open = links.classList.toggle('open');
      toggle.setAttribute('aria-expanded', String(open));
    });
  }

  const search = document.querySelector('[data-claim-search]');
  const status = document.querySelector('[data-claim-status]');
  const phase = document.querySelector('[data-claim-phase]');
  const rows = [...document.querySelectorAll('[data-claim-row]')];
  const empty = document.querySelector('[data-empty]');
  const filter = () => {
    const q = (search?.value || '').trim().toLowerCase();
    let visible = 0;
    rows.forEach(row => {
      const hit = (!q || row.textContent.toLowerCase().includes(q)) &&
        (!status?.value || row.dataset.status === status.value) &&
        (!phase?.value || row.dataset.phase === phase.value);
      row.hidden = !hit;
      if (hit) visible++;
    });
    if (empty) empty.style.display = visible ? 'none' : 'block';
  };
  [search, status, phase].forEach(el => el?.addEventListener('input', filter));

  document.querySelectorAll('[data-copy]').forEach(button => {
    button.addEventListener('click', async () => {
      const target = document.querySelector(button.dataset.copy);
      if (!target) return;
      await navigator.clipboard.writeText(target.textContent.trim());
      const old = button.textContent;
      button.textContent = 'Copied';
      setTimeout(() => button.textContent = old, 1400);
    });
  });
})();
