# Commercial Plan

This is the commercialization plan for Jini.

It is intentionally narrow:

- what stays free
- what gets sold first
- who buys it
- when to introduce paid product surfaces
- what metrics decide the next move

## Revenue Thesis

Jini should make money in this order:

1. paid adoption help
2. paid design-partner work
3. premium product surfaces
4. enterprise support and governance

It should **not** try to monetize the core framework early.

## Boundary Rule

The boundary between the public repo and the paid surface should be explicit.

Public and free:

- framework code
- CLI
- protocol docs
- schemas
- install and proof path
- public example packs
- tests

Paid:

- adoption help
- onboarding
- custom workflow mapping
- custom integrations
- premium surfaces only after repeated demand proves they should exist
- enterprise support and governance

Not acceptable:

- paywalling the core CLI
- hiding the proof path
- moving essential protocol docs out of the public repo
- making the free version intentionally incomplete just to force services revenue

The rule is simple: charge for acceleration, customization, and trust. Do not
charge for basic access to the framework.

## What Stays Free

These remain free:

- framework core
- public docs
- public example packs
- proof assets
- benchmark harness
- starter install path

These should remain in the public repo itself, not just in marketing copy.

Reason:

- free evaluation drives adoption
- adoption creates trust
- trust creates paid demand

## What Gets Sold First

The first revenue should come from services, not SaaS.

Reason:

- the repo can already generate interest
- services validate real demand quickly
- services reveal which premium surfaces are worth building
- hosted product too early would be speculation

This means the first paid offer is not “pay to unlock Jini.” It is “pay to get
Jini working correctly and faster in a real environment.”

## Target Buyers

Primary buyers:

- Head of Engineering
- CTO
- VP Engineering
- Platform / Developer Experience lead
- AI product lead
- PM or operations lead in high-governance environments

Primary users:

- engineering teams
- product teams
- platform teams
- operations teams
- regulated or audit-sensitive teams

Buyer problem:

- work is spread across prompts, docs, tickets, chat, and memory
- teams can generate output but cannot keep the work coherent
- verification, approvals, and handoffs are brittle

## Initial Offers

These are the only offers Jini should launch with.

| Offer | Buyer | Deliverable | Delivery model | Price band |
|---|---|---|---|---:|
| Setup Sprint | team lead / CTO | one repo running Jini with proof path and starter workflow | fixed scope | $3k-$7.5k |
| Workflow Review | engineering or product leadership | current-state audit + target-state design memo | fixed scope | $5k-$12k |
| Team Workshop | team manager / PM / eng lead | guided onboarding session with repo-specific walkthrough | fixed scope | $2k-$5k |
| Design Partner Program | serious early adopter | custom pack/control work + feedback loop + priority access | quarterly | $15k-$40k / quarter |

## Offer Definitions

### Setup Sprint

Goal:

- get one team to first real usage fast

Scope:

- install and packaging setup
- one repo bootstrap
- one workflow mapped into Jini
- one example pack selected or adapted
- proof command path working in the client repo

Success condition:

- client can run the proof path and use Jini in one real workflow without live help

### Workflow Review

Goal:

- diagnose where the client's current AI workflow breaks

Scope:

- workflow interviews
- artifact and handoff review
- gap analysis
- Jini fit assessment
- written recommendation memo

Success condition:

- buyer gets a concrete adoption decision and rollout sequence

### Team Workshop

Goal:

- shorten time-to-understanding for a real team

Scope:

- 60-120 minute session
- install path
- proof command path
- work-unit and artifact overview
- pack/adaptor overview relevant to the team

Success condition:

- team understands where Jini fits and how to try it without more explanation

### Design Partner Program

Goal:

- fund product evolution with real users

Scope:

- regular working sessions
- prioritized support
- custom pack or control-surface work
- feedback incorporated into roadmap

Success condition:

- partner uses Jini in recurring work and produces concrete product feedback

## What Not To Sell Yet

Do not sell these as primary offers yet:

- seat-based SaaS plans
- per-token charges
- complex SKU matrix
- heavy hosted platform
- paywalled core packs
- premium-only basic install path
- premium-only proof path

Reason:

- none of these improve adoption right now
- all of them add packaging and support burden

## Repo Boundary

Keep in the public repo:

- code
- schemas
- stable specs
- public-safe packs and examples
- tests
- community docs
- commercial overview

Keep out of the public repo:

- customer-specific packs
- customer data or benchmarks
- internal launch planning
- internal traces and generated learning logs
- private sales material
- paid-only delivery artifacts

If a stranger needs it to understand, install, verify, trust, or extend Jini,
it belongs in the public repo. If it is customer-specific, internal, or part of
a paid engagement, it does not.

## Product Monetization Later

Introduce premium product surfaces only after repeated demand appears in paid work.

Strong candidates:

| Surface | Why someone pays |
|---|---|
| premium domain packs | avoids internal pack design for specialized work |
| premium control packs | stronger governance and operational controls |
| managed install / activation | lower setup burden for teams |
| hosted evidence / approval dashboard | better visibility and auditability |
| team memory layer | better shared retrieval and continuity |
| analytics / drift monitoring | helps operators improve workflow quality |

## Enterprise Layer Later

Enterprise should be sold only after the open-source and services motion works.

Enterprise value comes from:

- SSO
- admin controls
- private deployment help
- compliance support
- custom adapters
- SLAs
- policy hardening

Indicative starting point:

- `$25k+` annual

This should be contract revenue, not lightweight self-serve pricing.

## Decision Gates

Jini should not add a premium product layer until all of these are true:

- at least `3` paid customers completed services work
- at least `2` design partners are active
- at least `2` premium asks repeat across different customers
- at least `1` surface clearly causes repeated delivery cost in services

Jini should not add enterprise packaging until all of these are true:

- at least `2` customers ask for security, governance, or deployment features
- at least `1` customer is willing to pay annual contract pricing
- the implementation burden is understood well enough to scope reliably

## KPI Set

Track these metrics monthly:

| KPI | Why it matters |
|---|---|
| repo clones | top-of-funnel interest |
| proof-path completions | real evaluation, not vanity traffic |
| install support requests | packaging friction |
| qualified inbound leads | commercial demand |
| paid conversion rate | whether the offer is real |
| service gross margin | whether services are worth scaling |
| repeated premium requests | signal for premium product buildout |
| average time to first value | adoption quality |

## 90-Day Execution Plan

### Days 0-30

- publish public repo
- add community files
- add issue forms and discussions
- add `Commercial` section to README
- add one contact path
- publish Setup Sprint and Team Workshop offers

Exit criteria:

- public repo is live
- commercial path is visible
- at least one buyer can contact you without ambiguity

### Days 31-60

- run outreach from article, proof asset, and social post
- close first workshop or setup sprint
- record objections and repeated asks
- refine offer scope and pricing based on real conversations

Exit criteria:

- at least one paid engagement
- top objections documented
- offer language updated from reality, not theory

### Days 61-90

- close design partner if demand exists
- publish FAQ from repeated objections
- decide whether GitHub Sponsors is worth enabling
- decide whether any premium surface has enough repeated demand

Exit criteria:

- at least two paid engagements or a clear diagnosis of why not
- one decision on whether to stay services-first longer or start premium planning

## Pricing Rules

Use these rules:

- fixed fee for scoped early offers
- do not start with hourly pricing in public
- do not publish enterprise pricing early
- increase price only after repeated delivery, not after one good call
- keep the first paid offer simple enough to say yes to quickly

## Contact And Community Separation

Community path:

- GitHub Issues
- GitHub Discussions
- public docs

Commercial path:

- contact email or form
- workshop inquiry
- setup sprint inquiry
- design-partner inquiry

Do not mix public support with private sales discussions.

## Immediate TODOs

- [x] Add a `Commercial` section to [README.md](./README.md)
- [x] Add a contact path:
  `maridlabsai@gmail.com`
- [ ] Publish a one-page Setup Sprint offer
- [ ] Publish a one-page Team Workshop offer
- [ ] Create a design-partner intake checklist
- [ ] Decide whether to enable GitHub Sponsors
- [ ] Review pricing after first two paid engagements

## Lessons From Current Winners

Current open-core and AI-adjacent winners consistently monetize:

- managed control surfaces
- hosted layers
- usage at scale
- governance and enterprise trust
- paid acceleration around an otherwise accessible core

That is the correct pattern for Jini to copy.

The references I used for this pattern are:

- [LangSmith pricing](https://www.langchain.com/pricing)
- [LlamaIndex pricing](https://www.llamaindex.ai/pricing)
- [n8n pricing](https://n8n.io/pricing/)
- [Supabase pricing](https://supabase.com/pricing)
- [PostHog pricing](https://posthog.com/pricing)
- [GitHub Sponsors](https://docs.github.com/en/sponsors/getting-started-with-github-sponsors/about-github-sponsors)

## Default Answer

If someone asks how Jini makes money, the answer should be:

Jini core is free. Revenue comes first from adoption help, then from design
partners, then from premium and enterprise layers once repeated demand proves
they are worth building.
