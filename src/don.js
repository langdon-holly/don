#!/usr/bin/env node

'use strict';

// Dependencies

const
  util = require('util')
  , {Readable} = require('stream')

  , program = require('commander')

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
  , [{file, cleanup}, input, srcPath, srcDispPath]
    = hasFileArg
      ? [ don.readFile(Buffer.from(program.args[0]))
        , don.stdin()
        , don.ok(don.strToInts(program.args[0]))
        , program.args[0]]
      : [ don.stdin()
        , {file: Readable({read() {this.push(null)}}), cleanup: () => 0}
        , don.nothing
        , "standard input"]
  , error = e => console.error(don.strVal(e));

don.parse(file, don.parseStream).then
( parsed =>
  { if (parsed.success)
      don.topApply
      ( don.makeFun
        ( () =>
          ( { fn
              : don.bindRest
                (parsed.ast, {rest: {file, cleanup}, input, srcPath})
            , okThen: {fn: don.makeFun(fn => ({fn, arg: don.initEnv}))}}))
      , don.unit
      , don.nullCont
      , don.makeCont(e => (error(e), process.exit(1))))
      .then(() => process.exit());
    else error(parsed.error(don.strToInts(srcDispPath))), process.exit(2)});
