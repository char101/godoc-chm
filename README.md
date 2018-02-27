# godoc-chm

This is a tool to generate CHM project files from go documentation hosted on a godoc HTTP server.

## Features

* All packages, variables, constants, functions, and types shown hierarchically in
  the table of contents.
* All packages, variables, constants, functions, and types searchable in the index.

## Download

An example CHM can be downloaded from the [releases](https://github.com/char101/godoc-chm/releases) tab.

## Install

```
go get github.com/char101/godoc-chm
```

## Usage

```
godoc-chm [-cache] [-output directory] [-chm path-to-compiled-chm] [-open] [-compile] godoc-url
```

## Notes

If you are using Windows, you need IE9 (because the godoc
javascript uses `getElementsByClassName` which is only supported by IE9 upwards).
Then change the browser emulation compatibility for `hh.exe` in the registry. The value below 
sets the compatibility value for IE11.

```
[HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Internet Explorer\Main\FeatureControl\FEATURE_BROWSER_EMULATION]
"hh.exe"=dword:00002af8
```

For the list of values see https://docs.microsoft.com/en-us/previous-versions/windows/internet-explorer/ie-developer/general-info/ee330730(v=vs.85)
