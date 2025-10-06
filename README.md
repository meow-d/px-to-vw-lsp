# px to vw conversion lsp
before you tell me i could have saved 90% of the time by not implementating an lsp, have you considered that maybe out of the 8 billion people on earth, maybe there will be someone with this very specific use case who also happens to use helix? have they implemented plugins yet? why they even settle for their own scheme implementation after spending years arguing how every other option creates maintaince burden? would learning lisp unlock my third eye and chakhra and fix all of my life's problems? whqat the hell? all questions for you to ponder and distract you from the true problem that is my insecurities.

## usage
1. download or build the binary
2. put it in `~/.local/bin` or somewhere in your `$PATH`
3. configure your editor. for example, in neovim:

```lua
-- css px-to-vw
local configs = require("lspconfig.configs")
local util = require("lspconfig.util")

if not configs["px_to_vw_lsp"] then
  configs["px_to_vw_lsp"] = {
    default_config = {
      cmd = { "px-to-vw-lsp" },
      filetypes = { "css", "scss", "less" },
      root_dir = util.root_pattern(".git") or vim.fn.getcwd(),
      single_file_support = true,
      name = "px-to-vw",
    },
  }
end

lspconfig["px_to_vw_lsp"].setup {
  on_attach = nvlsp.on_attach,
  on_init = nvlsp.on_init,
  capabilities = nvlsp.capabilities,
}
```

4. in your project root, create a `.cssrem` file with the following:

```json
{
  "$schema": "https://raw.githubusercontent.com/cipchk/vscode-cssrem/master/schema.json",
  "fixedDigits": 3,
  "vwDesign": 1920
}
```

this config file is used by the [cssrem vscode extension](https://marketplace.visualstudio.com/items?itemName=cipchk.cssrem). though all options other than the two above are ignored here.

## build
```sh
git clone https://github.com/meow-d/px-to-vw-lsp
cd px-to-vw-lsp
go get ./cmd/px-to-vw-lsp
go build ./cmd/px-to-vw-lsp
# or: go build -o ~/.local/bin/px-to-vw-lsp ./cmd/px-to-vw-lsp
```

## todo
- [x] workspace folders support

- [x] stop dumping everything into logs so that they are actually readable
- [ ] more useful logs because now the logs have zero info

- [ ] monitor .cssrem for changes (rather than just reading once on startup)
- [ ] conversion in code lens, like what cssrem does?
- [ ] refactor config loading code
