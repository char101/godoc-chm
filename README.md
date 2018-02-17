# godoc-chm

Generates CHM documentation from godoc http server.

## Features

* All packages, variables, constants, functions, and types shown hierarchically in
  the table of contents
* All packages, variables, constants, functions, and types searchable in the index

## Download

An example CHM for go 1.10 can be downloaded from the releases tab.

## Install

```
go get github.com/char101/godoc-chm
```

## Usage

```
godoc-chm [-cache] [-output directory] godoc-url
```

## Notes

If you are using Windows, you need to have at least IE9 (because the godoc
javascript uses `getElementsByClassName` which is supported only by IE9 upwards
and enable the browser emulation compatibility value. The value below is for
IE11.

```
[HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Internet Explorer\Main\FeatureControl\FEATURE_BROWSER_EMULATION]
"hh.exe"=dword:00002af8
```

For the list of values see https://docs.microsoft.com/en-us/previous-versions/windows/internet-explorer/ie-developer/general-info/ee330730(v=vs.85)
