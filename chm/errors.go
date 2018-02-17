package chm

import "errors"

// ErrUnindent is returned when trying to unindent the root node
var ErrUnindent = errors.New("TOC: unindent: root node")

// ErrIndent is returned when doing invalid indent
var ErrIndent = errors.New("TOC: indent: no parent")
