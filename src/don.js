#!/usr/bin/env node

'use strict'

// Dependencies

; const
    fs = require('fs')
  , util = require('util')

  , program = require('commander')
  , red = require('chalk').red

  , don = require('./don-core.js')

// Stuff

; program.parse(process.argv)

; const hasFileArg = program.args.length > 0

; fs.readFile
  (hasFileArg ? program.args[0] : 0
  , 'utf8'
  , (err, data) =>
    { if (err) console.log("Couldn't read file"), process.exit(1)

    ; const parsed = don.parse(data)

    ; if (parsed.success)
        don.topApply
        ( don.bindRest(parsed.ast, parsed.rest)
        , don.initEnv
        , don.nullCont
        , don.makeCont
          ( e =>
              ( console.error(red(don.strVal(don.toString(e))))
              , process.exit(1))))
        , process.exit(0)
    ; else
        console.error(red(parsed.error(hasFileArg ? program.args[0] : "STDIN")))
        , process.exit(2)})

