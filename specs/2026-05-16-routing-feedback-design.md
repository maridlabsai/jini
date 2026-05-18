# Routing And Model Feedback Design

Date: 2026-05-16

## Goal

Make Jini's runtime choice more legible and more improvable without pushing
tooling jargon onto the normal user path.

## Product Shape

Normal path:

1. install Jini
2. run `jini`
3. paste the work you want finished
4. if setup is missing, type `Use Auto`

Advanced path:

- strict route users can still force Claude, Bedrock, or Azure OpenAI
- power users can inspect tool, provider, model, effort, and reason

## Model Transparency

Whenever Jini chooses a model, the shell should show:

- `Model`
- `Why this model`

After work exists, the user should be able to vote on the model choice:

- `Model upvote`
- `Model downvote`

That feedback is stored with the work so later heuristic revisions can use it.

## First Heuristic Upgrade

The old router was fixed candidate ordering with keyword buckets.

The first upgrade keeps the same supported runtime surface, but switches to a
scored chooser:

- classify work type
- classify depth
- classify modality
- score each route
- select the best ready route
- explain why

Current route families:

- Claude Code
- Bedrock Sonnet
- Azure-backed writing route
- Azure-backed code route
- Azure OpenAI
- Local preview

The public shell no longer presents Azure-backed routes as if they were direct
ChatGPT or Codex products.

## Local SLM Pool

The router now has explicit local profile slots in code:

- `local-fast`
- `local-workhorse`
- `local-deep`
- `local-multimodal`

These are policy slots only for now. They are not a public route until a real
runtime exists.

## Install Truthfulness

The installer should:

- try a prebuilt release binary first
- fall back to source build only when needed

That allows the normal install story to stop treating Go and Git as required
for everyone while keeping the source-build fallback explicit.

## Follow-on Work

1. add a real local SLM transport
2. feed model upvotes and downvotes back into route scoring
3. add confidence scores and fallback explanations
4. separate policy constraints from cost and depth heuristics
