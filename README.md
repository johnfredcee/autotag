*Autotag

This exists to parallelize the problem of running a ctags.exe - like command on a large codebase (in this case, Unreal Engine 4)

The codebase can be split into different sections each with it's own section in the conf.json file and it's own specified command line parameters. Autotags will build a list of files to tag (based on wildcards) and produce a tag file, in parallel for each sections.

I hope the conf.json is self-explanatory

*TODO

Watch directories and kick off rescans
Store tags in some database like SQLite for faster indexing and retreival

