# Neovim integration

[`lua/lint/linters/dyshellint.lua`](lua/lint/linters/dyshellint.lua) is a linter
definition for [nvim-lint](https://github.com/mfussenegger/nvim-lint).

It pipes the buffer to `dyshellint` on standard input and passes the real path in
`--stdin-filename`, so a script is checked while it is being written, before it
is ever saved, and `dyshellint` still finds the `.shellcheckrc` of the project and
resolves `# shellcheck source=` directives.

Every finding becomes a diagnostic with:

| Field                    | Value                                            |
| ------------------------ | ------------------------------------------------ |
| `severity`               | `ERROR` for ✔️ SHOULD and ❌ AVOID, `WARN` for ⚠️ CONSIDER |
| `code`                   | `BSG###`, `SC####` or `FMT001`                    |
| `source`                 | `dyshellint`, `shellcheck` or `shfmt`              |
| `user_data.section`      | the heading of the style guide the rule comes from |

## Install the binary

```sh
go install gitlab.com/dynamo-tools/dyshellint/cmd/dyshellint@latest
```

Any `dyshellint` on `$PATH` will do; set `cmd` on the linter to point somewhere
else.

## Register the linter

nvim-lint loads a linter by name from `lua/lint/linters/` anywhere on the
runtimepath, so adding this directory is enough:

```lua
{
  'mfussenegger/nvim-lint',
  opts = function()
    vim.opt.runtimepath:append('/path/to/dyshellint/editors/nvim')
    require('lint').linters_by_ft.sh = { 'dyshellint' }
  end,
}
```

Copying the single file into your own configuration works the same way:

```sh
cp editors/nvim/lua/lint/linters/dyshellint.lua ~/.config/nvim/lua/lint/linters/
```

Without lazy.nvim, or to keep everything in one file, declare it inline:

```lua
local lint = require('lint')
lint.linters.dyshellint = dofile('/path/to/dyshellint.lua')
lint.linters_by_ft.sh = { 'dyshellint' }
```

## Run it

nvim-lint does not lint on its own. The usual trigger, with the buffer checked
on every save and every time insert mode is left:

```lua
vim.api.nvim_create_autocmd({ 'BufWritePost', 'InsertLeave', 'TextChanged' }, {
  pattern = { '*.sh', '*.bash', '*.bats' },
  callback = function() require('lint').try_lint() end,
})
```

## Options

The linter takes no options of its own: it reads `.shellcheckrc` from the
project, like the command line does. To change what is reported, edit `args`:

```lua
require('lint').linters.dyshellint.args = {
  '--format', 'json',
  '--no-warnings',              -- only the ✔️ SHOULD and ❌ AVOID tier
  '--exclude-rules', 'BSG070',  -- and not the line length either
  '--stdin-filename', function() return vim.api.nvim_buf_get_name(0) end,
  '-',
}
```
