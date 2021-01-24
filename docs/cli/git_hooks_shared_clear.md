## git hooks shared clear

Clear shared repositories.

### Synopsis

Clears every item in the shared repositories list.
If `--local|--global` is given, then the `githooks.shared` local/global
Git configuration is modified, or if the `--shared` option (default) is set, the `.githooks/.shared.yaml`
file is modified in the local repository.
The `--all` option clears all three lists.

```
git hooks shared clear [flags]
```

### Options

```
      --shared   Modify the shared hooks list `.githooks/.shared.yaml` (default).
      --local    Modify the shared hooks list in the local Git config.
      --global   Modify the shared hooks list in the global Git config.
      --all      Modify all shared hooks lists (`--shared`, `--local`, `--global`).
  -h, --help     help for clear
```

### SEE ALSO

* [git hooks shared](git_hooks_shared.md)	 - Manages the shared hook repositories.

###### Auto generated by spf13/cobra 