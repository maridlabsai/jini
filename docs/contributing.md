---
title: Contributing
description: Repo-local Claude Code workflows for contributors changing Jini.
eyebrow: Contributor workflow
context_line: "Keep the customer shell small while giving contributors sharp review, gate, and route-audit shortcuts."
highlights:
  - Customer shell stays `jini`
  - Claude Code assets are repo-local
  - Public and commercial boundaries stay separate
quick_links:
  - label: Command Catalog
    href: /cli.html
  - label: Proof
    href: /proof.html
  - label: Install
    href: /install.html
---

<p class="page-lead">Jini should stay simple for users and strict for contributors. The public shell starts with <code>jini</code>. Repo-local Claude Code assets exist to keep implementation work aligned before a commit.</p>

<div class="section-card">
  <span class="section-kicker">Public repo</span>
  <h2>Claude Code helpers for Jini core</h2>
  <p>The public repo includes <code>CLAUDE.md</code> and <code>.claude/</code>. These assets are development helpers, not public Jini shell commands.</p>
  <div class="checklist-grid">
    <div class="checklist-card">
      <h3><code>/gates</code></h3>
      <p>Run the required public commit gate and report release-relevant evidence.</p>
    </div>
    <div class="checklist-card">
      <h3><code>/review</code></h3>
      <p>Review the current diff for product, implementation, test, release, and security risk.</p>
    </div>
    <div class="checklist-card">
      <h3><code>/route-audit</code></h3>
      <p>Inspect routing, offline fallback, provider selection, and installed-CLI handoff behavior.</p>
    </div>
    <div class="checklist-card">
      <h3><code>/prd-drift</code></h3>
      <p>Check changed work against the canonical product direction and remove stale requirements instead of adding drift.</p>
    </div>
    <div class="checklist-card">
      <h3><code>/alpha-smoke</code></h3>
      <p>Verify alpha-tester prompts still produce compact, competitor-grade CLI output.</p>
    </div>
  </div>
</div>

<div class="section-card section-card-soft">
  <span class="section-kicker">Commercial repo</span>
  <h2>Paid-surface helpers stay private</h2>
  <p>The private commercial repo has its own <code>CLAUDE.md</code> and <code>.claude/</code> assets. They cover paid packaging, entitlement, app readiness, renewal proof, commercial tests, and the commercial-only agent/skills OS.</p>
  <div class="pill-list">
    <span>/commercial-test</span>
    <span>/entitlement-audit</span>
    <span>/app-readiness</span>
    <span>/renewal-proof</span>
    <span>/agent-skills-os</span>
  </div>
  <p>Commercial work must extend the public Jini contract without redefining the free CLI, leaking paid-only features into the free tier, or claiming app release readiness without matching evidence.</p>
</div>

<div class="section-card">
  <span class="section-kicker">Boundary</span>
  <h2>Do not teach users contributor commands</h2>
  <p>Customer docs should continue to show the normal flow: install once, run <code>jini</code>, describe the task, and inspect setup only when Jini asks for it. Contributor slash commands belong in repo workflows and review notes, not in the first-run product experience.</p>
</div>
