# BVR Skills

BVR discovers Agent Skills from default directories including:

- `~/.config/bvr/skills`
- `~/.config/agents/skills`
- `~/.agents/skills`
- `~/.claude/skills`
- `~/Documents/AI-SKILLS`

Project-level skill directories such as `.agents/skills` and `.bvr/skills`
are also discovered automatically. Each skill must be a directory containing a
valid `SKILL.md` with frontmatter.

The local master catalog at `~/Documents/AI-SKILLS` is preloaded by default, so
skills there are available immediately after install. If you want to refresh
explicitly, run:

```sh
bvr-cli skills scan
```

Or use the legacy installer/symlink approach:

```sh
./scripts/install-master-skills.sh
```

Override the default discovery paths when needed:

```sh
BVR_SKILLS_DIR=/path/to/skills bvr-cli
```

## Builtin skills

BVR also embeds a set of builtin skills directly into the binary via
`internal/skills/builtin/`. These are always available — no installation or
symlinks needed — and are surfaced to the coding agent as `<available_skills>`
in the coder system prompt. Builtin skills use virtual
`bvr://skills/<name>/SKILL.md` locations, which the View tool resolves from the
embedded filesystem.

| Skill            | Description                                                  |
| ---------------- | ------------------------------------------------------------ |
| `bvr-config`    | BVR configuration help                                      |
| `bvr-hooks`     | Authoring, configuring, and debugging hooks                  |
| `jq`             | jq JSON processor usage guide                                |
| `cline`          | Autonomous terminal coding agent (delegates a coding job)    |
| `hermes`         | Self-improving autonomous agent runtime and model router     |

`cline` and `hermes` are coding-assistant skills: when the local coder model is a
small model that struggles with multi-step coding, the coder system prompt
(`internal/agent/templates/coder.md.tpl`) directs the agent to hand off to Cline
or Hermes. Both follow the
[agentskills.io](https://agentskills.io) open standard — the same format BVR
skills use — so project rules and skills can be shared. See
[`BVR_CODER.md`](BVR_CODER.md) for the coder-entry design.

User skills with the same name as a builtin override the builtin (last
occurrence wins in `skills.Deduplicate()`).