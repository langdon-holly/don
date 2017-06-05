#!/usr/bin/env node

'use strict'

// Dependencies

; var
    don = require('./don-core.js')
  , program = require('commander')
  , fs = require('fs')
  , _ = require('lodash')

// Stuff

; program.parse(process.argv)

; function indexToLineColumn(index, string)
  { var arr = Array.from(string), line = 0, col = 0
  ; for (var i = 0; i < arr.length; i++)
    { if (i == index)
        return {line0: line, col0: col, line1: ++line, col1: ++col}
    ; if (arr[i] === '\n') line++, col = 0
    ; else col++}
    throw new RangeError
              ("indexToLineColumn: index=" + index + " is out of bounds")}

; var arg0 = program.args[0]

; fs.readFile
  ( program.args.length > 0 ? arg0 : 0
  , 'utf8'
  , function(err, data)
    { if (err) console.log("Couldn't read file"), process.exit(1)

      ; var parsed = don.parse(data)

      ; if (parsed.status === 'match')
          don.topEval(parsed.ast, parsed.rest)
          , process.exit(0)
      ; else if (parsed.status === 'eof')
          console.log
          ( "Syntax error: "
            + arg0
            + " should be at least "
            + (Array.from(data).length + 1)
            + " characters")
      ; else
        { var errAt = parsed.index
        ; if (errAt == 0) console.log("Error in the syntax")
        ; else
          { var lineCol = indexToLineColumn(errAt - 1, data)
          ; console.log("Syntax error at "
                        + arg0
                        + " "
                        + lineCol.line1
                        + ","
                        + lineCol.col1
                        + ":")
          ; console.log(data.split('\n')[lineCol.line0]
                        + "\n"
                        + " ".repeat(lineCol.col0)
                        + "^")}

        //; var trace = parsed.trace
        //; _.forEachRight(trace, function(frame) {console.log("in", frame[0])})
        ; console.log(parsed.parser)

        ; process.exit(2)}})

