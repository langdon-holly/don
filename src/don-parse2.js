'use strict';

const util = require('util');

const _ = require('lodash');

const
  inspect = o => util.inspect(o, {depth: null, colors: true})
  , log
    = (...args) =>
      (console.log(args.map(inspect).join("\n")), _.last(args));

const
  iterableIntoIterator
  = (to, from) =>
  { from = from[Symbol.iterator]();
    let res = to.next();
    while (!res.done) res = to.next(from.next());
    return res.value;}
, asyncIterableIntoIterator
  = async (to, from) =>
    { from = from[Symbol.asyncIterator]();
      let res = to.next();
      while (!res.done) res = to.next(await from.next());
      return res.value;}
, streamIntoIterator
  = (to, from) =>
    new Promise
    ( resolve =>
      { let res = to.next();
        const wStream
          = require('stream').Writable
            ( { write(value, encoding, cb)
                { if ((res = to.next({value})).done)
                    resolve(res.value), from.unpipe(this);
                  cb(null);}
              , final(cb)
                { resolve(to.next({done: true}).value), from.unpipe(this);
                  cb(null);}
              , objectMode: true});
        from.pipe(wStream);})
, ws = Buffer.from(' \t\n\r')
, delimL = Buffer.from('([{\\')
, delimR = Buffer.from(')]}|')
, msg
  = { preExpr: "whitespace, comment, or delimitation"
    , esc: "target of escape"}
, ascii
  = _.mapValues
    ( {lf: '\n', hash: '#', backtick: '`', semicolon: ';', bang: '!'}
    , s => s.codePointAt(0));

Object.assign
( exports
, { iterableIntoIterator
  , parseStream: str => streamIntoIterator(it()[Symbol.iterator](), str)
  , parseIter
    : async asyncIterable =>
      asyncIterableIntoIterator(it()[Symbol.iterator](), asyncIterable)});

function *it()
{ const
    n/*ext*/
    = (yielded, expected) =>
      ( expect = expected
      , value === ascii.lf
        ? (++line, col = 0, currLine = [], lines.push(currLine))
        : (++col, currLine.push(value))
      , ++index
      , ({value} = yielded).done)
  , doomed
    = () =>
      ( currLine.push(value)
      , { status: 'doomed'
        , index
        , result
          : { line
            , col
            , lines
            , expect
            , nest: nest()}})
  , e/*of*/
    = () =>
      ( { status: 'eof'
        , index
        , result: {line, col, lines, expect, nest: nest()}})
  , delimited
    = delim => ({t: 'delimited', d: {...stack.pop(), end: {line, col, delim}}})
  , begin = () => stack.push({start: {line, col, delim: value}, inner: []})
  , pushPos = () => stack.push({start: {line, col}})
  , commentBegin = () => (++nestLevel, pushPos())
  , nest = () => stack.map(({start: {line, col}}) => ({line, col}));

  let value = ascii.lf
  , index = 0
  , stack = []
  , commentLevel = 0
  , nestLevel = 0
  , line = -1
  , col = 0
  , lines = []
  , currLine
  , expect;

  if (n(yield, "shebang, whitespace, comment, or delimitation")) return e();

  // Shebang
  if (value === ascii.hash)
  { pushPos();
    if (n(yield, "`!")) return e();
    if (value !== ascii.bang) return doomed();
    do if (n(yield, "rest of shebang")) return e(); while (value !== ascii.lf);
    stack.pop();
    if (n(yield, msg.preExpr)) return e();}

  while (true)
  { if (delimL.includes(value))
      if (commentLevel--)
      { commentBegin();
        do
        { if (n(yield, "rest of comment")) return e();
          if (delimL.includes(value)) commentBegin();
          else if (delimR.includes(value)) nestLevel--, stack.pop();
          else if (value === ascii.backtick)
          {pushPos(); if (n(yield, msg.esc)) return e(); stack.pop();}}
        while (nestLevel);}
      else break;
    else if (value === ascii.semicolon) ++commentLevel;
    else if (!ws.includes(value))
      return doomed();
    if (n(yield, msg.preExpr)) return e();}

  begin();

  while (true)
  { if (n(yield, "rest of delimitation")) return e();

    if (delimL.includes(value)) begin();
    else if (delimR.includes(value))
    { if (stack.length == 1)
        return (
          {status: 'match', index, result: {tree: delimited(value), lines}});
      stack[stack.length - 2].inner.push(delimited(value))}
    else if (value === ascii.backtick)
    { pushPos();
      if (n(yield, msg.esc)) return e();
      stack.pop();
      _.last(stack).inner.push({t: 'esc', d: value});}
    else
      _.last(stack).inner.push({t: 'elem', d: value})}}
