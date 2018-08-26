#!/usr/bin/env node

'use strict';

// Dependencies

const
  fs = require('fs')
  , util = require('util')
  , {Readable} = require('stream')

  , program = require('commander')
  , {red} = require('chalk')

  , don = require('./don-core.js');

// Utility

const
  inspect = o => util.inspect(o, {depth: null, colors: true})
  , log
    = (...args) =>
      (console.log(args.map(inspect).join("\n")), args[args.length - 1]);

// Stuff

program.parse(process.argv);

const
  hasFileArg = program.args.length > 0
  , [{file, cleanup: sourceCleanup}, input]
    = hasFileArg
      ? [don.readFile(program.args[0]), don.stdin()]
      : [ don.stdin()
        , {file: Readable({read() {this.push(null)}}), cleanup: () => 0}];

don.parse(file, don.parseStream).then
( parsed =>
  { if (parsed.success)
      don.topApply
      ( don.makeFun
        ( () =>
          ( { fn
              : don.bindRest
                (parsed.ast, {rest: {file, cleanup: sourceCleanup}, input})
            , okThen: {fn: don.makeFun(fn => ({fn, arg: don.initEnv}))}}))
      , don.unit
      , don.nullCont
      , don.makeCont
        ( e =>
          (console.log(red(don.strVal(don.toString(e)))), process.exit(1))))
      .then(() => process.exit());
    else
      console.log(parsed.error(hasFileArg ? program.args[0] : "standard input"))
      , process.exit(2)});
