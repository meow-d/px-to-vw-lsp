# px to vw conversion lsp
before you tell me i could have saved 90% of the time by not implementating an lsp, have you considered that maybe out of the 8 billion people on earth, maybe there will be someone with this very specific use case who also happens to use helix? have they implemented plugins yet? why they even settle for their own scheme implementation after spending years arguing how every other option creates maintaince burden? would learning lisp unlock my third eye and chakhra and fix all of my life's problems? whqat the hell? all questions for you to ponder and distract you from the true problem that is my insecurities.

## usage
### 1. install
1. [download](https://github.com/meow-d/px-to-vw-lsp/releases) or build the binary
2. put it somewhere in your `$PATH`, like `~/.local/bin`

### 2. configure your editor
for example, in neovim:

```lua
-- css px-to-vw
vim.lsp.config("px_to_vw_lsp", {
  -- alternatively cmd = { "px-to-vw-lsp", "--log-level=debug" },
  cmd = { "px-to-vw-lsp"},
  filetypes = { "css", "scss", "less" },
  root_dir = vim.fn.getcwd(),
  workspace_required = false,
  name = "px_to_vw_lsp",
})

vim.lsp.enable("px_to_vw_lsp")
```

### 3. configure window height
- global config: `~/.local/share/px-to-vw-lsp/config.json`
- per-project config: `.cssrem` file in project root

it uses the same json as the [cssrem vscode extension](https://marketplace.visualstudio.com/items?itemName=cipchk.cssrem), though all options other than the two above are ignored here.

```json
{
    "$schema": "https://raw.githubusercontent.com/cipchk/vscode-cssrem/master/schema.json",
    "fixedDigits": 3,
    "vwDesign": 1920
}
```
## development
### clone
```sh
git clone https://github.com/meow-d/px-to-vw-lsp
cd px-to-vw-lsp
```

### build
```sh
go get ./cmd/px-to-vw-lsp
go build ./cmd/px-to-vw-lsp
# or: go build -o ~/.local/bin/px-to-vw-lsp ./cmd/px-to-vw-lsp
```

### test
```sh
go test ./... -v
```

### debug
```sh
# to view logs
tail -f /tmp/px-to-vw-lsp.log
# for editor specific issues
tail -f ~/.local/state/nvim/lsp.log
```

## todo
- [x] workspace folders support
- [x] stop dumping everything into logs so that they are actually readable
- [x] more useful logs because now the logs have zero info
- [x] refactor config loading code
- [x] global configuration support with automatic file monitoring
- [x] testing

- [ ] clean up ai generated code
- [ ] monitor .cssrem for changes (rather than just reading once on startup)
- [ ] conversion in code lens, like what cssrem does?
