#!/usr/bin/env node

'use strict'

// Dependencies

; const
    fs = require('fs')
  , util = require('util')
  , {Readable} = require('stream')

  , program = require('commander')
  , {red} = require('chalk')

  , don = require('./don-core.js')

// Utility

; const
    inspect = o => util.inspect(o, {depth: null, colors: true})
  , log = (...args) =>
    (console.log(args.map(inspect).join("\n")), args[args.length - 1])

// Stuff

; program.parse(process.argv)

; const
    hasFileArg = program.args.length > 0
  , [{file: source, cleanup: sourceCleanup}, input]
    = hasFileArg
      ? [don.readFile(program.args[0]), don.stdin()]
      : [ don.stdin()
        , {file: Readable({read() {this.push(null)}}), cleanup: ()=>0}]

; don.parse(source).then
  ( parsed =>
    {
//; fs.readFile
//  ( hasFileArg ? program.args[0] : 0
//  , 'utf8'
//  , (err, data) =>
//    { if (err) console.log("Couldn't read file"), process.exit(1)
//
//    ; const parsed = don.parse(data)

      if (parsed.success)
        don.topApply
        ( don.bindRest
          ( parsed.ast
          , {rest: {file: parsed.rest, cleanup: sourceCleanup}, input})
        , don.initEnv
        , don.nullCont
        , don.makeCont
          ( e =>
              ( console.log(red(don.strVal(don.toString(e))))
              , process.exit(1))))
        .then(() => process.exit())
    ; else
        console.log(red(parsed.error(hasFileArg ? program.args[0] : "STDIN")))
        , process.exit(2)})

