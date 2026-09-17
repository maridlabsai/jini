# Community provider catalog

`providers.json` is the community catalog of OpenAI-compatible BYO providers. It
ships embedded in the binary and overlays the built-in providers; a user's local
`~/.jini/providers.json` overlays this in turn.

## Add a provider

Open a PR that adds one entry to `providers` in `providers.json`:

```json
{
  "id": "together",
  "label": "Together AI",
  "key_env": "TOGETHER_API_KEY",
  "default_base": "https://api.together.xyz",
  "chat_path": "/v1/chat/completions",
  "model_env": "TOGETHER_MODEL",
  "default_model": "meta-llama/Llama-3.3-70B"
}
```

| field | required | notes |
| --- | --- | --- |
| `id` | yes | unique route id; also the `jini route set <id>` name |
| `label` | yes | display name |
| `key_env` | yes | env var / config key holding the API key (Bearer) |
| `default_base` | routable | API root, e.g. `https://api.together.xyz` |
| `chat_path` | routable | e.g. `/v1/chat/completions` |
| `model_env` | routable | env var to pick a model |
| `default_model` | routable | model id used when `model_env` is unset |
| `models_path` | no | defaults to `/models` |
| `privacy_note` | no | shown on validate; also gates the provider out of the automatic fallback ladder |

An entry matching a built-in id **overrides only the fields it sets** — e.g.
`{"id": "groq", "default_model": "…"}` fixes a stale default without redefining
the provider.

## How it's gated (automated, no manual judgment)

The `Catalog Auto-Merge` workflow:

1. **Schema gate** — the Go suite validates every entry (always, no secrets).
2. **Live quality gate** — for each newly-added provider, if a `<ID>_API_KEY`
   secret is configured, it runs `jini provider validate` + `jini check model`
   (the automated capability floor). Both must pass.
3. **Auto-merge** — only when the schema passes **and** every new provider was
   live-verified. A provider with no configured key is never auto-merged; it's
   labelled `needs-maintainer-key` for a maintainer to add the secret.

Only OpenAI-compatible (Bearer) providers fit this schema. A provider with a
bespoke API needs a built-in shape (a code change), not a catalog entry.
