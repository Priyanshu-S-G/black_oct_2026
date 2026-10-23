// Cockpit shell wiring: tabs, modals, and an engine-reachability indicator.
// This is presentation only. It owns no game state and decides no outcomes.

import { getHealth } from "./api.js";

const HEALTH_POLL_MS = 5000;
const HEALTH_TIMEOUT_MS = 3000;

function initTabs() {
  const tabs = Array.from(document.querySelectorAll('[role="tab"]'));

  function select(tab) {
    for (const t of tabs) {
      const selected = t === tab;
      t.setAttribute("aria-selected", String(selected));
      t.tabIndex = selected ? 0 : -1;
      document.getElementById(t.getAttribute("aria-controls")).hidden = !selected;
    }
  }

  tabs.forEach((tab, i) => {
    tab.addEventListener("click", () => select(tab));
    tab.addEventListener("keydown", (ev) => {
      let next = null;
      if (ev.key === "ArrowRight") next = tabs[(i + 1) % tabs.length];
      if (ev.key === "ArrowLeft") next = tabs[(i - 1 + tabs.length) % tabs.length];
      if (next) {
        ev.preventDefault();
        select(next);
        next.focus();
      }
    });
  });
}

function initDialogs() {
  const pairs = [
    ["open-stack", "stack-dialog"],
    ["open-ledger", "ledger-dialog"],
  ];
  for (const [buttonId, dialogId] of pairs) {
    const dialog = document.getElementById(dialogId);
    document.getElementById(buttonId).addEventListener("click", () => dialog.showModal());
  }
}

function initHealth() {
  const el = document.getElementById("engine-status");

  async function check() {
    try {
      const body = await getHealth(AbortSignal.timeout(HEALTH_TIMEOUT_MS));
      if (body.status !== "ok") throw new Error("unexpected health payload");
      el.dataset.state = "ok";
      el.textContent = "Engine API: reachable";
    } catch {
      el.dataset.state = "down";
      el.textContent = "Engine API: unreachable";
    }
  }

  check();
  setInterval(check, HEALTH_POLL_MS);
}

initTabs();
initDialogs();
initHealth();
