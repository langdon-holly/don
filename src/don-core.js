'use strict';

// Dependencies

const
  fs = require('fs')
  , util = require('util')
  , {Writable, Readable, Transform} = require('stream')
  , path = require('path')

  , _ = require('lodash')
  , weak = require('weak')
  , {blue, red, bold} = require('chalk')

  , {parseStream: parser, parseIter, iterableIntoIterator}
    = require('./don-parse2.js');

// Utility
const
  debug = true
  , inspect = o => util.inspect(o, {depth: null, colors: true})
  , log
    = (...args) =>
      (debug && console.log(args.map(inspect).join("\n")), _.last(args))
  , bufStream2byteStream
    = () =>
      Transform
      ( { transform(...[buf,, cb])
          {for (const byte of buf) this.push(byte); cb(null);}
        , readableObjectMode: true});
//  , onDemandStream
//    = stream =>
//      Duplex
//      ( { read(size) {stream.pipe(this)}
//        , write(chunk, enc, cb)
//          {stream.unpipe(this); this.push(chunk, enc); cb(null)}
//        , decodeStrings: false})
//  , promiseSyncMap
//    = (arrIn, promiseFn) =>
//      _.reduce
//      ( arrIn
//      , (prm, nextIn, idx) =>
//        prm.then
//        ( arrOut =>
//          promiseFn(nextIn).then
//          (newVal => (arrOut[idx] = newVal, Promise.resolve(arrOut))))
//      , Promise.resolve(Array(arrIn.length)))

// Stuff

exports.parseStream = parser;

function Continue(cont, arg)
{ const
    {type: contType, data: contData} = cont
    , {type: argType, data: argData} = arg;
  return (
    contType === contLabel ? contData(arg)
    : contType === quoteLabel
      ? [{cont: makeFun(_.constant({val: contData})), arg}]
      : contType === callLabel
        ? [ { cont
              : makeFun
                ( arg =>
                  ( { fn: contData.fnExpr
                    , arg
                    , okThen
                      : { fn
                          : makeFun
                            ( fnVal =>
                              ( { fn: contData.argExpr
                                , arg
                                , okThen
                                  : { fn
                                      : makeFun
                                        ( argVal =>
                                          ( { fn: fnVal
                                            , arg: argVal}))}}))}}))
            , arg}]
        : contType === boolLabel
          ? [ { cont
                : makeFun
                  ( val =>
                    ( { val
                      : contData ? makeFun(_.constant({val})) : I}))
              , arg}]
          : Null
            ("Tried to continue a non-continuation:\n" + inspect(cont)))}

function mk(type, data) {return {type, data}}

function makeCall(fnExpr, argExpr) {return mk(callLabel, {fnExpr, argExpr})}

function fnOfType(type, fn)
{ return (
    makeFun
    ( (arg, ...ons) =>
      arg.type !== type
      ? { ok: false
        , val
          : strToInts
            ( "Typed function received garbage:\n"
              + inspect(type)
              + "\n"
              + inspect(arg))}
      : fn(arg.data, ...ons)))}

const
  makeThenOns
  = (then, onOk, onErr) =>
    makeCont
    ( arg => 
      [ mCall
        ( then.hasOwnProperty('fn') ? then.fn : arg
        , applySym
        , then.hasOwnProperty('arg') ? then.arg : arg
        , then.hasOwnProperty('onOk')
          ? then.onOk
          : then.hasOwnProperty('okThen')
            ? makeThenOns(then.okThen, onOk, onErr)
            : onOk
        , then.hasOwnProperty('onErr')
          ? then.onErr
          : then.hasOwnProperty('errThen')
            ? makeThenOns(then.errThen, onOk, onErr)
            : onErr)]);

const
  funResToThreads
  = (res, onOk, onErr) =>
    _.isArray(res)
    ? _.flatMap(res, res => funResToThreads(res, onOk, onErr))
    : res.hasOwnProperty('cont')
      ? [res]
      : res.hasOwnProperty('fn')
        ? [ mCall
            ( res.fn
            , applySym
            , res.hasOwnProperty('arg') ? res.arg : unit
            , res.hasOwnProperty('onOk')
              ? res.onOk
              : res.hasOwnProperty('okThen')
                ? makeThenOns(res.okThen, onOk, onErr)
                : onOk
            , res.hasOwnProperty('onErr')
              ? res.onErr
              : res.hasOwnProperty('errThen')
                ? makeThenOns(res.errThen, onOk, onErr)
                : onErr)]
        : [ { cont: !res.hasOwnProperty('ok') || res.ok ? onOk : onErr
            , arg: res.val}];

function makeFun(fn)
{ return (
    arrToObj
    ( [ [ applySym
        , makeCont
          ( arg =>
            { let cc, onerr;
              return (
                arg.type === pairLabel
                && (cc = arg.data.first).type === contLabel
                && arg.data.last.type === pairLabel
                && (onerr = arg.data.last.data.last).type === contLabel
                ? ( async () =>
                    funResToThreads
                    ( await
                        fn(arg.data.last.data.first, cc, onerr), cc, onerr))
                  ()
                : Null("Function requires enpaired continuations"))})]]))}
exports.makeFun = makeFun;

function quote(val) {return mk(quoteLabel, val)}

function makeList(vals)
{return _.reduceRight(vals, _.ary(_.flip(makePair), 2), unit)}

function makeInt(Int) {return mk(intLabel, Int)}

function gensym(debugId) {return mk(symLabel, {sym: debugId})}

function makeIdent(key)
{ const
    appFun
    = makeFun
      ( arg =>
        ( { fn: arg
          , arg: key
          , okThen
            : { fn
                : fnOfType
                  ( resultLabel
                  , mExpr =>
                    mExpr.ok ? {fn: mExpr.val, arg} : unknownKeyThread(key))}}))
    , This
      = arrToObj
        ( [ [ applySym
            , makeCont(arg => [mCast(appFun, applySym, arg)])]
          , [identKeySym, makeCont(cont => ([{cont, arg: key}]))]]);
  return This}

function makeBool(val) {return mk(boolLabel, val)}

function ok(val) {return mk(resultLabel, {ok: true, val})}
exports.ok = ok;

function err(val) {return mk(resultLabel, {ok: false, val})}

function makeCont(fn) {return mk(contLabel, fn)}
exports.makeCont = makeCont

function makePair(first, last) {return mk(pairLabel, {first, last})}

function makeChannel()
{ let mode, queue = [];
  return [
    makeCont
    ( arg =>
      mode || !queue.length
      ? (mode = true, queue.push(arg), [])
      : [{cont: queue.pop(), arg}])
  , makeCont
    ( cont =>
      !mode || !queue.length
      ? (mode = false, queue.push(cont), [])
      : [{cont, arg: queue.pop()}])]}

const
  byteStream2intStream
  = () =>
    Transform
    ( { transform(...[byte,, cb]) {this.push(makeInt(byte)); cb(null)}
      , objectMode: true});

const
  intStream2asyncIter
  = cont =>
    { let execDone = Promise.resolve();
      const
        getNext
        = res =>
          [ { cont
            , arg
              : makeCont
                ( next =>
                  ( next.type === resultLabel
                    ? next.data.ok
                      ? next.data.val.type === intLabel
                        ? res({value: next.data.val.data})
                        : Null("Streamt without integrity")
                      : next.data.val === unit
                        ? res({done: true})
                        : Null("Non-unitial stream err")
                    : Null("Stream argued without result")
                  , []))}];
      return (
        { rIter
          : ( async function*()
              { while (true)
                { const
                    next
                    = await
                        new Promise
                        ( (...a) =>
                          execDone
                          = Promise.all([execDone, topContinue(getNext(...a))])
                            .then(_.noop));
                  if (next.done) return;
                  yield next.value;}})
            ()
        , execWait: () => execDone})};

function makeStream(rStream, cleanup)
{ return (
    makeFun
    ( () =>
      { let prmRes, prm, reader;
        const
          [write, read] = makeChannel()
          , handleThread
            = inner =>
              ( { cont: read
                , arg
                  : makeCont
                    ( arg =>
                      ( reader = arg
                      , prm
                        = [promiseWaitThread(new Promise(res => prmRes = res))]
                      , inner(null)
                      , prm))})
          , eosThreads
            = cont =>
              [{cont: read, arg: makeCont(eosThreads)}, {cont, arg: nothing}]
          , writable
            = Writable
              ( { write(...[elem,, cb])
                  { const res = prmRes;
                    topContinue
                    ([{cont: reader, arg: ok(elem)}, handleThread(cb)])
                    .then(res)}
                , final(cb)
                  { const res = prmRes;
                    cb(null);
                    topContinue(eosThreads(reader)).then(res)}
                , objectMode: true});
        weak(write, cleanup);
        return (
          [ handleThread
            ( () =>
              ( rStream.pipe(writable)
              , weak(write, () => rStream.unpipe(writable))))
          , {val: write}])}))}

function arrToObj(arr)
{ return (
    makeCont
    ( msg =>
      msg.type === pairLabel
      && msg.data.first.type === symLabel
      ? ( index =>
          index >= 0
          ? [{cont: arr[index][1], arg: msg.data.last}]
          : Null("Wrong message type"))
        (arr.findIndex(p => eq(p[0], msg.data.first)))
      : Null("Bad message:\n%s\n\n%s", inspect(msg), inspect(arr))))}

function isList(val)
{ while (val.type === pairLabel) val = val.data.last;
  return val === unit}

function* listIter(list)
{ while (list.type === pairLabel) yield list.data.first, list = list.data.last;
  return list === unit}

function listToArr(list) {return [...listIter(list)];}

function reverseConcat(list0, list1)
{ while (list0 !== unit)
    list1 = makePair(list0.data.first, list1), list0 = list0.data.last;
  return list1}

function listReverse(list) {return reverseConcat(list, unit)}

function listConcat(list0, list1)
{return list1 === unit ? list0 : reverseConcat(listReverse(list0), list1)}

function listsConcat(lists)
{return lists.reduceRight((list1, list0) => listConcat(list0, list1), unit)}

function isBytes(val)
{ return (
    isList(val)
    &&
      listToArr(val).every
      (elem => elem.type === intLabel && elem.data >= 0 && elem.data < 256))}

function intToStr(int)
{ if (int.type !== intLabel || int.data < 0 || int.data >= 256)
    return Null("intToStr nonbyte: " + strVal(toString(int)));
  return String.fromCodePoint(int.data)}

function bufVal(list)
{ if (!isBytes(list)) return Null("Tried to bufVal nonbytes");
  return Buffer.from(listToArr(list).map(int => int.data));}

function strVal(list) {return bufVal(list).toString();}
exports.strVal = strVal

function bufToInts(buf) {return makeList([...buf].map(makeInt));}

function strToInts(str) {return bufToInts(Buffer.from(str));}
exports.strToInts = strToInts;

function mCast(cont, mSym, arg) {return {cont, arg: makePair(mSym, arg)}}

function mCall(cont, mSym, arg, onOk, onErr)
{return {cont, arg: makePair(mSym, makePair(onOk, makePair(arg, onErr)))}}

function eq(val0, val1)
{ return (
    val0.type === val1.type
    &&
      ( val0.data === val1.data
      || val0.type === quoteLabel && eq(val0.data, val1.data)
      ||
        val0.type === callLabel
        && eq(val0.data.fnExpr, val1.data.fnExpr)
        && eq(val0.data.argExr, val1.data.argExr)
      || val0.type === resultLabel
         && val0.data.ok === val1.data.ok
         && eq(val0.data.val, val1.data.val)
      || val0.type === pairLabel
         && eq(val0.data.first, val1.data.first)
         && eq(val0.data.last, val1.data.last)))}

function subDelimited({t, d})
{ return (
    t === 'delimited'
    ? makePair(makeInt(2), delimitedTree(d))
    : makePair(makeInt(t === 'esc' ? 1 : 0), makeInt(d)));}

function delimitedTree(data)
{ return (
    makePair
    ( makePair(makeInt(data.start.delim), makeInt(data.end.delim))
    , makeList(data.inner.map(subDelimited))));}

function parseTreeToAST({t, d})
{ if (t !== 'delimited') return Null("unknown parse-tree type '" + label);
  return makeCall(delimitedVar, quote(delimitedTree(d)));}

const intLabel = {label: 'int'};
const symLabel = {label: 'sym'};
const quoteLabel = {label: 'quote'};
const callLabel = {label: 'call'};
const boolLabel = {label: 'bool'};
const resultLabel = {label: 'result'};
const contLabel = {label: 'cont'};
const pairLabel = {label: 'pair'};

const unit = mk({label: 'unit'});
exports.unit = unit;


const toStrSym = gensym(strToInts('to-str-sym'));
const applySym = gensym(strToInts('apply-sym'));
const identKeySym = gensym(strToInts('ident-key-sym'));
const
  filetype
  = { reg: gensym(strToInts('file-type-reg'))
    , dir: gensym(strToInts('file-type-dir'))
    , other: gensym(strToInts('file-type-?'))};

const delimitedVarSym = gensym(strToInts('delimited-var'));
const delimitedVar = makeIdent(delimitedVarSym);

const srcPathVarSym = gensym(strToInts('src-path-var'));
const srcPathVar = makeIdent(srcPathVarSym);

const Null = (...args) => {throw new Error("Null: " + util.format(...args))};

const nullCont = makeCont(_.constant([]));
exports.nullCont = nullCont;

const nothing = exports.nothing = err(unit);

const I = makeFun(val => ({val}));

const makeMapCall
  = makeFun
    ( arg =>
      { if (!isList(arg))
          return {ok: false, val: strToInts("Insequential cartographic call")};
        const args = listToArr(arg);
        if (args.length % 2 != 1)
          return {ok: false, val: strToInts("Tried to brace evenness")};
        const pairs = _.chunk(_.tail(args), 2);
        return (
          { fn: args[0]
          , arg
            : makeFun
              ( arg =>
                { let toReturn = nothing;
                  pairs.forEach
                  ( pair =>
                    eq(arg, pair[0]) ? (toReturn = ok(pair[1]), false) : true);
                  return {val: toReturn}})})});

const
  syncMap
  = makeFun
    ( fn =>
      { const
          listFn
          = makeFun
            ( list =>
              { if (!isList(list))
                  return (
                    { ok: false
                    , val
                      : strToInts("Insequential synchronous cartography")});
                return (
                  list === unit
                  ? {val: unit}
                  : { fn
                    , arg: list.data.first
                    , okThen
                      : { fn
                          : makeFun
                            ( newHead =>
                              ( { fn: listFn
                                , arg: list.data.last
                                , okThen
                                  : { fn
                                      : makeFun
                                        ( newTail =>
                                          ( { val
                                              : makePair
                                                ( newHead
                                                , newTail)}))}}))}})});
        return {val: listFn}});

const
  readFile
  = filepath =>
    { const
        pipeFrom = fs.createReadStream(filepath)
        , pipeTo = bufStream2byteStream();
      return (
        {file: pipeFrom.pipe(pipeTo), cleanup() {pipeFrom.unpipe(pipeTo)}})};
exports.readFile = readFile;

const
  writeFile
  = (filepath, flags) =>
{ const
    file = fs.createWriteStream(filepath, {flags}), write = writeToStream(file);
  return (
    weak
    ( makeFun
      ( arg =>
        isBytes(arg)
        ? [promiseWaitThread(write(arg)), {val: unit}]
        : {ok: false, val: strToInts('Tried to write nonbytes')})
    , () => file.end()));}

const
  replacement = 65533 // Replacement character
  , decodeUtf8Indices
    = (arr, idxs) =>
      { let
          codepoints = []
          , map = []
          , idx = -1
          , cont = 0
          , val
          , byte
          , mapIdx = 0;
        const
          push
          = (cp, idx) =>
            { while (mapIdx < idxs.length && idx > idxs[mapIdx])
                map.push(codepoints.length), ++mapIdx;
              codepoints.push(cp);}

        while (++idx < arr.length)
        { byte = arr[idx];
          if (byte < 0b10000000)
          { if (cont) push(replacement, idx), cont = 0; push(byte, idx + 1);}
          else if (byte < 0b11000000)
          { if (cont)
            {val += byte - 0b10000000 << --cont * 6; if (!cont) push(val, idx + 1);}
            else push(replacement, idx);}
          else if (byte < 0b11100000)
          { if (cont) push(replacement, idx), cont = 0;
            cont = 1, val = byte - 0b11000000 << 6;}
          else if (byte < 0b11110000)
          { if (cont) push(replacement, idx), cont = 0;
            cont = 2, val = byte - 0b11100000 << 12;}
          else if (byte < 0b11111000)
          { if (cont) push(replacement, idx), cont = 0;
            cont = 3, val = byte - 0b11110000 << 18;}
          else push(replacement, idx), cont = 0, val = 0;}

        if (cont) push(replacement, idx);
        while (mapIdx++ < idxs.length) map.push(codepoints.length);
        return {codepoints, map};}

const
  parseFile
  = async (stream, parseFn) =>
    { const parsed = await parseFn(stream);

    switch (parsed.status)
    { case 'match'
      : return {success: true, ast: parseTreeToAST(parsed.result.tree)};
      case 'eof'
      : case 'doomed'
      : const errAt = parsed.index;
        if (errAt == 0)
          return (
            { success: false
            , error: _.constant(strToInts("Error in the syntax"))});
        else
        { const {nest, lines, line, col, expect} = parsed.result;
          let onNote = 0, color = blue;
          return (
            { success: false
            , error
              : filepath =>
                listsConcat
                ( [ strToInts
                    ( (nest.length ? "Nested (outer first):" : "Not nested:")
                      + "\n"
                      + _.reduce
                        ( [...nest, {line, col}]
                        , (out, next) =>
                          ( out.length && next.line === _.last(out).line
                            ? _.last(out).cols.push(next.col)
                            : out.push({line: next.line, cols: [next.col]})
                          , out)
                        , [])
                        .map
                        ( ({line, cols}) =>
                          { const
                              {codepoints: theLine, map: colCps}
                              = decodeUtf8Indices(lines[line], cols);
                            let print = ["", ""], idx = 0, str;
                            for (let col of colCps)
                              onNote++ === nest.length ? color = red : 0
                              , print[0]
                                +=
                                  String.fromCodePoint
                                  (...theLine.slice(idx, col))
                                  + ( col < theLine.length
                                    ? bold
                                      ( color
                                        (String.fromCodePoint(theLine[col])))
                                    : "")
                              , print[1]
                                += " ".repeat(col - idx) + bold(color("^"))
                              , idx = col + 1;
                            print[0]
                            += String.fromCodePoint(...theLine.slice(idx));
                            return (
                              " ┌Line "
                              + (line + 1)
                              + "; Column"
                              + (colCps.length - 1 ? "s" : "")
                              + " "
                              + colCps.map(col => col + 1).join()
                              + "\n │"
                              + print[0]
                              + "\n └"
                              + print[1]
                              + "\n")})
                        .join(""))
                  , ...
                      parsed.status === 'eof'
                      ? [ strToInts("Syntax error: unfinished program in ")
                        , filepath]
                      : [ strToInts("Syntax error at ")
                        , filepath
                        , strToInts(" " + (line + 1) + "," + (col + 1))]
                  , strToInts("\n  Expected " + expect)])})}}};
exports.parse = parseFile;

const topApply = (fn, ...stuf) => topContinue([mCall(fn, applySym, ...stuf)]);
exports.topApply = topApply;

const
  topContinue
  = threads =>
    Promise.all
    ( threads.map
      ( t =>
        new Promise
        ( res =>
          setImmediate
          ( () =>
            Promise.resolve(Continue(t.cont, t.arg)).then(topContinue)
            .then(res)))))
    .then(_.noop);

const
  bindRest
  = (expr, {rest, input, srcPath}) =>
    ( sourcePath =>
      makeFun
      ( () =>
        ( { fn
            : makeStream(input.file.pipe(byteStream2intStream()), input.cleanup)
          , okThen
            : { fn
                : makeFun
                  ( theStdin =>
                    ( okQuoteStdin =>
                      ( { fn
                          : makeStream
                            ( rest.file.pipe(byteStream2intStream())
                            , rest.cleanup)
                        , okThen
                          : { fn
                              : makeFun
                                ( theSourceData =>
                                  ( okQuoteSourceData =>
                                    ( { val
                                        : makeFun
                                          ( fn =>
                                            ( { fn: expr
                                              , arg
                                                : makeFun
                                                  ( arg =>
                                                    eq
                                                    ( arg
                                                    , strToInts('source-data'))
                                                    ? {val: okQuoteSourceData}
                                                    : eq
                                                      ( arg
                                                      , strToInts('stdin'))
                                                      ? {val: okQuoteStdin}
                                                      : eq(arg, srcPathVarSym)
                                                        ? {val: sourcePath}
                                                        : {fn, arg})}))}))
                                  (ok(quote(theSourceData))))}}))
              (ok(quote(theStdin))))}})))
    (ok(quote(srcPath)));
exports.bindRest = bindRest;

//const
//  delimL = Buffer.from('([{\\')
//  , delimR = Buffer.from(')]}|')
//  , byteName
//    = byte =>
//      delimL.includes(byte)
//      ? 'left'
//      : delimR.includes(byte)
//        ? 'right'
//        : byte === 59 ? 'semicolon' : byte === 96 ? 'backtick' : 'other'
//  , makeBacktick = () => makeInt(96);
//function escInIdent(intArr)
//{ let ticked = false, identStack = [[[]]], name;
//  const stackLog = () => log(identStack.map(o => o.map(o => o.map(intToStr))));
//  for (let int of intArr)
//    ( name = byteName(int.data)
//    , ticked
//      ? ( name === 'other'
//          ? _.last(_.last(identStack)).push(int)
//          : _.last(identStack).push([int])
//        , ticked = false)
//      : name === 'other'
//        ? _.last(_.last(identStack)).push(int)
//        : name === 'left'
//          ? identStack.push([[int]])
//          : identStack.length === 1
//            ? identStack[0].push([int])
//            : name === 'right'
//              ? _.last(identStack[identStack.length - 2]).push
//                (..._.flatten(identStack.pop()), int)
//              : /* name === 'backtick' || name === 'semicolon' */
//                (_.last(identStack).push([int]), ticked = name === 'backtick')
//    /*, stackLog()*/);
//  return (
//    _.reduce(_.flatten(identStack), (a, b) => [...a, makeBacktick(), ...b]))}

function toString(arg)
{ const {type: argLabel, data: argData} = arg;
  let list;
  return (
    /*argLabel === charLabel ? makeList([...strToInts("`"), arg])

    : */argLabel === intLabel ? strToInts(argData.toString() + ' ')

    : isList(arg)
      ? ( list = listToArr(arg)
        //, arg !== unit && isBytes(arg)
        //  ? listsConcat
        //    ( list.some
        //      ( int =>
        //        [32, 10, 9, 13, 40, 41, 91, 93, 123, 125, 96, 92, 124, 59, 34]
        //        .includes(int.data))
        //      ? [strToInts("\\'"), makeList(escInIdent(list)), strToInts('|')]
        //      : [strToInts("'"), arg, strToInts(' ')])

        , listsConcat
          ([strToInts('['), ...list.map(o => toString(o)), strToInts(']')]))

    : argLabel === quoteLabel
      ? listsConcat([strToInts("(q "), toString(argData), strToInts(")")])

    : argLabel === callLabel
      ? listsConcat
        ( [ strToInts("(make-call ")
          , toString(argData.fnExpr)
          , toString(argData.argExpr)
          , strToInts(")")])

    : argLabel === symLabel
      ? listsConcat
        ([strToInts("(gensym "), toString(argData.sym), strToInts(")")])

    : argLabel === boolLabel
      ? argData ? strToInts("true ") : strToInts("false ")

    : argLabel === resultLabel
      ? listsConcat
        ( [ strToInts(argData.ok ? "(ok " : "(err ")
          , toString(argData.val)
          , strToInts(")")])

    : argLabel === contLabel ? strToInts("(cont ... )")

    : argLabel === pairLabel
      ? listsConcat
        ( [ strToInts("(cons ")
          , toString(argData.first)
          , toString(argData.last)
          , strToInts(')')])

    : Null("->str unknown type:", inspect(arg)))}

const
  [lParen, rParen, lBracket, rBracket, lBrace, rBrace, backslash, pipe]
  = [...Buffer.from('()[]{}\\|')].map(makeInt);

const
  unknownKeyThread
  = key =>
    ( { ok: false
      , val
        : listConcat
          ( strToInts("Unknown variable key: ")
          , toString(key))});

const
  stdin
  = () =>
    ( toBytes =>
      ( { file: process.stdin.pipe(toBytes)
        , cleanup() {process.stdin.unpipe(toBytes)}}))
    (bufStream2byteStream());
exports.stdin = stdin;

const
  writeToStream
  = wStream =>
    ( (toWrite, writable, prmRes, nextPrmRes, prm, nextPrm, buf) =>
        { const
            getNextPrm = () => nextPrm = new Promise(res => nextPrmRes = res)
            , writeIt
              = buf =>
                { writable = false;
                  prmRes = nextPrmRes, getNextPrm();
                  wStream.write
                  ( buf
                  , () =>
                    { writable = true;
                      prmRes([]);
                      toWrite.length
                      && writeIt((o => o)(Buffer.from(toWrite), toWrite = []));}
                  )
                };
          getNextPrm();
          return (
            ints =>
            ( buf = bufVal(ints)
            , buf.length
              ? ( prm = nextPrm
                , writable ? writeIt(buf) : toWrite.push(...buf)
                , prm)
              : Promise.resolve()));})
      ([], true)
  , stdout = writeToStream(process.stdout);

const
  promiseWaitThread
  = prm => ({cont: makeCont(() => prm.then(_.constant([]))), arg: unit});

const wsNums = Buffer.from(' \t\n\r')
, delimitedListGen
  = function *(env, listFn)
    { let value, index = 0, exprs = [], commentLevel = 0, name, quoteLevel = 0;

      const
        next
        = yielded => (++index, ({value} = yielded).done)
      , doom
        = () => value.type !== pairLabel || value.data.first.type !== intLabel
      , nameDoom
        = () =>
          doom()
          || value.data.first.data === 2
          || value.data.last.type !== intLabel
          ||
            value.data.first.data === 0
            && [34, 59].includes(value.data.last.data)
      , doomed
        = () =>
          ( { ok: false
            , val: strToInts("Delimitation enlistment doomed at index " + index)}
          )
      , eof
        = () =>
          ( { ok: false
            , val: strToInts("EOF reached while enlisting delimitation")})
      , push
        = expr =>
          { if (commentLevel > 0) commentLevel--;
            else
            { while (quoteLevel-- > 0) expr = quote(expr);
              ++quoteLevel, exprs.push(expr);}};

      while (!next(yield))
      { if (doom()) return doomed();
        const type = value.data.first.data;

        if (type === 2) push(makeCall(delimitedVar, quote(value.data.last)));
        else
        { if (value.data.last.type !== intLabel) return doomed();
          if (type === 0)
          { const byte = value.data.last.data;
            if (wsNums.includes(byte)) continue;
            if (byte === 59) {++commentLevel; continue;} // ;
            if (byte === 34) {if (!commentLevel) ++quoteLevel; continue;}} // "

          name = [];
          do
          { name.push(value.data.last.data);
            if (next(yield)) return eof();
            if (nameDoom()) return doomed();}
          while
          ( value.data.first.data === 1
            || !wsNums.includes(value.data.last.data));
          push(makeIdent(makeList(name.map(makeInt))));}}
      return (
        commentLevel > 0 || quoteLevel > 0
        ? eof()
        : { fn: syncMap
          , arg: makeFun(fn => ({fn, arg: env}))
          , okThen
            : { fn
                : makeFun
                  (fn => ({fn, arg: makeList(exprs), okThen: {fn: listFn}}))}});
    }
, delimitedList
  = (list, env, listFn) =>
    iterableIntoIterator(delimitedListGen(env, listFn), listIter(list));

const
  parenListFn
  = makeFun
    ( (elems, onOk, onErr) =>
      !isList(elems)
      ? {ok: false, value: strToInts("Insequential delimitation")}
      : elems === unit
        ? {val: I}
        : { cont
            : _
              .reduceRight
              ( listToArr(elems.data.last)
              , (onOk, arg) =>
                makeCont(fn => [mCall(fn, applySym, arg, onOk, onErr)])
              , onOk)
          , arg: elems.data.first})

const
  flattenDelimited
  = delimitation =>
    { if
      ( delimitation.type !== pairLabel
      || delimitation.data.first.type !== pairLabel)
        return {};
      const ints = [delimitation.data.first.data.first];
      for (const elem of listIter(delimitation.data.last))
      { if (elem.type !== pairLabel || elem.data.first.type !== intLabel)
          return {};
        switch (elem.data.first.data)
        { case 0
          : ints.push(elem.data.last);
            break;
          case 1
          : ints.push(...listIter(strToInts('`')), elem.data.last);
            break;
          case 2
          : const sub = flattenDelimited(elem.data.last)
            if (!sub.is) return {};
            ints.push(...sub.val);}}
      return {is: true, val: [...ints, delimitation.data.first.data.last]};}
, delimitedIdentGen
  = function *(env)
    { let value, index = 0, commentLevel = 0;

      const ints = []
      , next
        = yielded => (++index, ({value} = yielded).done)
      , doom
        = () => value.type !== pairLabel || value.data.first.type !== intLabel
      , doomed
        = () =>
          ( { ok: false
            , val: strToInts("Identical delimitation doomed at index " + index)})
      , eof
        = () =>
          ( { ok: false
            , val: strToInts("EOF reached in identical delimitation")});

      while (!next(yield))
      { if (doom()) return doomed();

        switch (value.data.first.data)
        { case 0 // elem
          : if (value.data.last.type !== intLabel) return doomed();
            else if (value.data.last.data == 59) // ;
              { ++commentLevel;
                do
                { if (next(yield)) return eof();
                  if (doom() || value.data.first.data == 1) return doomed();
                  if (value.data.first.data == 2) commentLevel--;
                  else
                  { if (value.data.last.type !== intLabel) return doomed();
                    if (value.data.last.data == 59) ++commentLevel; // ;
                    else if
                    (!wsNums.includes(value.data.last.data)) return doomed();}}
                while (commentLevel);
                continue;}
          case 1 // esc
          : ints.push(value.data.last);
            break;
          case 2 // delimited
          : const sub = flattenDelimited(value.data.last);
            if (!sub.is) return doomed();
            ints.push(...sub.val)}}
      return {fn: makeIdent(makeList(ints)), arg: env};}
, delimitedIdent
  = (list, env) => iterableIntoIterator(delimitedIdentGen(env), listIter(list));

const
  pathSepBuf = Buffer.from(path.sep)
  , derefSymlink
    = async fpath =>
      { fpath = bufVal(fpath);
        while ((await util.promisify(fs.lstat)(fpath)).isSymbolicLink())
        { const lnBuf = await util.promisify(fs.readlink)(fpath, 'buffer');
          fpath
          = path.isAbsolute(lnBuf.toString())
            ? lnBuf
            : Buffer.concat
              ( [Buffer.from(path.dirname(fpath.toString())), pathSepBuf, lnBuf]
              );}
        return fpath;}
  , derefSymlinkStr = async fpath => (await derefSymlink(fpath)).toString()
  , dirBaseOf
    = dirBase =>
      makeFun
      ( async filepath =>
        isBytes(filepath)
        ? { val
            : strToInts(path[dirBase + "name"](await derefSymlinkStr(filepath)))
          }
        : {ok: false, val: strToInts(`Tried to ${dirBase}-of nonbytes`)})
  , pathSepIter = listIter(strToInts(path.sep))
  , pathFromDir
    = makeFun
      ( dir =>
        { if (!isBytes(dir))
            return (
              { ok: false
              , val: strToInts('Tried to path-from-dir of nonbytes dir')});
          return (
            { val
              : makeFun
                ( val =>
                  { if (!isBytes(val))
                      return (
                        { ok: false
                        , val
                          : strToInts('Tried to path-from-dir of nonbytes path')
                        });
                    if (path.isAbsolute(strVal(val))) return {val};
                    return (
                      { val
                        : makeList
                          ( [ ...listIter(dir)
                            , ...pathSepIter
                            , ...listIter(val)])});})})});

const
  initEnv
  = makeFun
    ( varKey =>
      varKey.type === symLabel

      ? varKey === delimitedVarSym
        ? { val
            : ok
              ( makeFun
                ( env =>
                  ( { val
                      : fnOfType
                        ( pairLabel
                        , arg =>
                          { if (arg.first.type !== pairLabel)
                              return (
                                { ok: false
                                , val: strToInts("delimited with non-pair car")}
                              );
                            const
                              {first, last} = arg.first.data
                              , children = arg.last;

                            do
                            { if (eq(first, lParen))
                              { if (!eq(last, rParen)) break;
                                return (
                                  delimitedList(children, env, parenListFn));}
                              if (eq(first, lBracket))
                              { if (!eq(last, rBracket)) break;
                                return delimitedList(children, env, I);}
                              if (eq(first, lBrace))
                              { if (!eq(last, rBrace)) break;
                                return (
                                  delimitedList(children, env, makeMapCall));}
                              if (eq(first, backslash))
                              { if (!eq(last, pipe)) break;
                                return delimitedIdent(children, env);}}
                            while (0);
                            return (
                              { ok: false
                              , val
                                : listsConcat
                                  ( [ strToInts
                                      ("Unspecified delimited action for ")
                                    , makeList([first, last])])});})})))}

        : {val: nothing}

      : isBytes(varKey)

        ? ( keyStr =>
            initEnvObj.hasOwnProperty(keyStr)
            ? initEnvObj[keyStr]
            : varKey !== unit && varKey.data.first.data === "'".codePointAt(0)
              ? {val: ok(quote(varKey.data.last))}

            //        var varParts = maybeStr[1].split(':');
            //        if (varParts.length >= 2) {
            //          return _.reduce(varParts.slice(1, varParts.length),
            //                          function(fn, argument) {
            //                            return apply(fn,
            //                                         valObj(strLabel,
            //                                                argument))},
            //                          apply(env,
            //                                valObj(strLabel, varParts[0])))}

              : /^(\-|\+)?[0-9]+$/.test(keyStr)
                ? {val: ok(quote(makeInt(parseInt(keyStr, 10))))}

                : {val: nothing})
          (strVal(varKey))

        : {val: nothing});
exports.initEnv = initEnv;

const
  initEnvObj
  = _.mapValues
    ( { 'fn'
        : makeFun
          ( env =>
            ( { val
                : makeFun
                  ( paramKey =>
                    ( { val
                        : makeFun
                          ( body =>
                            ( { val
                                : makeFun
                                  ( arg =>
                                    ( { fn: body
                                      , arg
                                        : makeFun
                                          ( varKey =>
                                            eq(paramKey, varKey)
                                            ? {val: ok(quote(arg))}
                                            : {fn: env, arg: varKey}
                                          )}))}))}))}))

      , "apply-m": quote(applySym)

      , 'continue': quote(makeFun(cont => ({val: makeFun(arg => ({cont, arg}))})))

      , 'call/cc': quote(makeFun((fn, arg) => ({fn, arg})))

      , 'call/onerr': quote(makeFun((...[fn,, arg]) => ({fn, arg})))

      , 'make-cont'
        : quote
          ( makeFun
            ( onErr =>
              ( { val
                  : makeFun
                    ( fn =>
                      ( { val
                          : makeCont
                            ( arg =>
                              [ mCall
                                (fn, applySym, arg, nullCont, onErr)])}))})))

      , 'make-channel': makeFun(() => ({val: makePair(...makeChannel())}))

      , "async"
        : makeFun
          ( arg =>
            ({val: makeFun(fn => [{fn, arg, onOk: nullCont}, {val: unit}])}))

      , '+'
        : quote
          ( fnOfType
            ( intLabel
            , int0 =>
              ({val: fnOfType(intLabel, int1 => ({val: makeInt(int0 + int1)}))})))

      , '-'
        : quote
          ( fnOfType
            ( intLabel
            , int0 =>
              ({val: fnOfType(intLabel, int1 => ({val: makeInt(int0 - int1)}))})))

      , '<'
        : quote
          ( fnOfType
            ( intLabel
            , arg0 =>
              ({val: fnOfType(intLabel, arg1 => ({val: makeBool(arg0 < arg1)}))})
            ))

      , '='
        : quote
          ( makeFun
            (arg0 => ({val: makeFun(arg1 => ({val: makeBool(eq(arg0, arg1))}))})))

      , "env": I

      , "init-env": quote(initEnv)

      , "print"
        : quote
          ( makeFun
            ( arg =>
              isBytes(arg)
              ? [promiseWaitThread(stdout(arg)), {val: unit}]
              : {ok: false, val: strToInts('Tried to print nonbytes')}))

      , "say"
        : quote
          ( makeFun
            ( _.flow
              ( toString
              , stdout
              , prm => [promiseWaitThread(prm), {val: unit}])))

      , "->str": quote(makeFun(_.flow(toString, val => ({val}))))

      , "utf8-decode"
        : quote
          ( makeFun
            ( bytes =>
              { if (!isBytes(bytes))
                  return (
                    {ok: false, val: strToInts("Tried to utf8-decode nonbytes")}
                  );
                let idx = 0, cont, val;
                for (const Byte of listIter(bytes))
                { const byte = Byte.data;
                  if (cont)
                  { if (byte < 0b10000000 || byte >= 0b11000000)
                      return {val: makePair(nothing, makeInt(idx))};
                    val += byte - 0b10000000 << --cont * 6;
                    if (!cont)
                      return (
                        {val: makePair(ok(makeInt(val)), makeInt(idx + 1))});}
                  else
                  { if (byte < 0b10000000)
                      return {val: makePair(ok(makeInt(byte)), makeInt(1))};
                    else if (byte < 0b11000000)
                      return {val: makePair(nothing, makeInt(1))};
                    else if (byte < 0b11100000)
                      cont = 1, val = byte - 0b11000000 << 6;
                    else if (byte < 0b11110000)
                      cont = 2, val = byte - 0b11100000 << 12;
                    else if (byte < 0b11111000)
                      cont = 3, val = byte - 0b11110000 << 18;
                    else return {val: makePair(nothing, makeInt(1))};}
                  ++idx;}
                return {val: makePair(nothing, makeInt(0))};}))

      , "utf8-encode"
        : quote
          ( fnOfType
            ( intLabel
            , cp =>
              cp < 0 || cp >= 0x110000
              ? { ok: false
                , val: strToInts("Codepoint out of bounds: " + cp + " ")}
              : {val: strToInts(String.fromCodePoint(cp))}))

      , "flatten"
        : quote
          ( makeFun
            ( list =>
              { if (!isList(list))
                  return {ok: false, val: strToInts("flatten nonlist")};
                const lists = listToArr(list);
                if (!lists.every(isList))
                  return {ok: false, val: strToInts("flatten nonlistlist")};
                return {val: listsConcat(lists)};}))

      , "reverse-concat"
        : quote
          ( makeFun
            ( first =>
              ( { val
                  : makeFun
                    ( last =>
                      isList(first)
                      ? isList(last)
                        ? {val: reverseConcat(first, last)}
                        : { ok: false
                          , val: strToInts("nonlist nonreversed")}
                      : { ok: false
                        , val: strToInts("nonlist reversed")})})))

      , "true": quote(makeBool(true))

      , "false": quote(makeBool(false))

      , "unit": quote(unit)

      , "write-file"
        : quote
          ( makeFun
            ( arg =>
              { if (!isBytes(arg))
                  return (
                    { ok: false
                    , val: strToInts('Tried to write-file of nonbytes')});
                const buf = bufVal(arg);
                if (buf.includes(0))
                  return (
                    { ok: false
                    , val: strToInts('Tried to write-file with 0 byte')});
                return {val: writeFile(buf, 'w')};}))

      , "append-file"
        : quote
          ( makeFun
            ( arg =>
              { if (!isBytes(arg))
                  return (
                    { ok: false
                    , val: strToInts('Tried to append-file of nonbytes')});
                const buf = bufVal(arg);
                if (buf.includes(0))
                  return (
                    { ok: false
                    , val: strToInts('Tried to append-file with 0 byte')});
                return {val: writeFile(buf, 'a')};}))

      , "read-file"
        : quote
          ( makeFun
            ( arg =>
              { if (!isBytes(arg))
                  return (
                    { ok: false
                    , val: strToInts('Tried to read-file of nonbytes')});
                const buf = bufVal(arg);
                if (buf.includes(0))
                  return (
                    { ok: false
                    , val: strToInts('Tried to read-file with 0 byte')});
                const {file, cleanup} = readFile(buf);
                return (
                  {fn: makeStream(file.pipe(byteStream2intStream()), cleanup)});
              }))

      , "rm-file"
        : quote
          ( makeFun
            ( arg =>
              { if (!isBytes(arg))
                  return (
                    { ok: false
                    , val: strToInts('Tried to rm-file of nonbytes')});
                const buf = bufVal(arg);
                if (buf.includes(0))
                  return (
                    { ok: false
                    , val: strToInts('Tried to rm-file with 0 byte')});
                return (
                  util.promisify(fs.unlink)(buf).then
                  (() => ({val: unit})));}))

      , "mkdir"
        : quote
          ( makeFun
            ( arg =>
              { if (!isBytes(arg))
                  return (
                    {ok: false, val: strToInts('Tried to mkdir of nonbytes')});
                const buf = bufVal(arg);
                if (buf.includes(0))
                  return (
                    {ok: false, val: strToInts('Tried to mkdir with 0 byte')});
                return (
                  util.promisify(fs.mkdir)(buf).then
                  (() => ({val: unit})));}))

      , "read-dir"
        : quote
          ( makeFun
            ( arg =>
              { if (!isBytes(arg))
                  return (
                    { ok: false
                    , val: strToInts('Tried to read-dir of nonbytes')});
                const buf = bufVal(arg);
                if (buf.includes(0))
                  return (
                    { ok: false
                    , val: strToInts('Tried to read-dir with 0 byte')});
                return (
                  util.promisify(fs.readdir)(buf, 'buffer').then
                  (names => ({val: makeList(names.map(bufToInts))})));}))

      , "rmdir"
        : quote
          ( makeFun
            ( arg =>
              { if (!isBytes(arg))
                  return (
                    {ok: false, val: strToInts('Tried to rmdir of nonbytes')});
                const buf = bufVal(arg);
                if (buf.includes(0))
                  return (
                    {ok: false, val: strToInts('Tried to rmdir with 0 byte')});
                return (
                  util.promisify(fs.rmdir)(buf).then
                  (() => ({val: unit})));}))

      , "file-type"
        : quote
          ( makeFun
            ( arg =>
              { if (!isBytes(arg))
                  return (
                    { ok: false
                    , val: strToInts('Tried to file-type of nonbytes')});
                const buf = bufVal(arg);
                if (buf.includes(0))
                  return (
                    { ok: false
                    , val: strToInts('Tried to file-type with 0 byte')});
                return (
                  util.promisify(fs.stat)(buf).then
                  ( stats =>
                    ( { val
                        : stats.isFile()
                          ? filetype.reg
                          : stats.isDirectory() ? filetype.dir : filetype.other}
                    )));}))

      , "file-type-reg": quote(filetype.reg)
      , "file-type-dir": quote(filetype.dir)
      , "file-type-?": quote(filetype.other)

      , "parse-prog"
        : quote
          ( makeFun
            ( (...[arg,, onErr]) =>
              { const {rIter, execWait} = intStream2asyncIter(arg);
                return (
                  parseFile(rIter, parseIter).then
                  ( parsed =>
                    [ promiseWaitThread(execWait())
                    , { val
                        : parsed.success
                          ? ok(parsed.ast)
                          : err(makeFun(name => ({val: parsed.error(name)})))}])
                )})
          )

      , "path-from-dir": quote(pathFromDir)

      , "deref-symlink"
        : quote
          ( makeFun
            ( async fpath =>
              { if (!isBytes(fpath))
                  return (
                    { ok: false
                    , val: strToInts('Tried to deref-symlink nonbytes')});
                  return {val: bufToInts(await derefSymlink(fpath))};}))

      , "dir-of": quote(dirBaseOf('dir'))

      , "base-of": quote(dirBaseOf('base'))

      , "src-path-var-sym": quote(srcPathVarSym)

      , "src-path": makeFun(arg => ({fn: srcPathVar, arg}))

      , "use"
        : makeFun
          ( arg =>
            ( { fn: srcPathVar
              , arg
              , okThen
                : { fn
                    : fnOfType
                      ( resultLabel
                      , srcPath =>
                        { if (
                            srcPath.ok
                            ? !isBytes(srcPath.val)
                            : srcPath.val !== unit)
                            return (
                              { ok: false
                              , val: strToInts('Non-maybe-bytes src-path')});
                          const
                            withPath
                            = makeFun
                              ( srcPath =>
                                { const buf = bufVal(srcPath);
                                  if (buf.includes(0))
                                    return (
                                      { ok: false
                                      , val
                                        : strToInts('Tried to use with 0 byte')}
                                    );
                                  const {file, cleanup} = readFile(buf);
                                  return (
                                    parseFile(file, parser).then
                                    ( parsed =>
                                      parsed.success
                                      ? { fn
                                          : bindRest
                                            ( parsed.ast
                                            , { rest: {file, cleanup}
                                              , input
                                                : { file
                                                    : Readable
                                                      ( { read()
                                                          {this.push(null);}})
                                                  , cleanup: () => 0}
                                              , srcPath: ok(srcPath)})
                                        , okThen
                                          : { fn
                                              : makeFun
                                                (fn => ({fn, arg: initEnv}))}}
                                      : {ok: false, val: parsed.error(srcPath)})
                                  );});
                          return (
                            { val
                              : srcPath.ok
                                ? makeFun
                                  ( async arg =>
                                    ( { fn: pathFromDir
                                      , arg
                                        : strToInts
                                          ( path.dirname
                                            ( ( await
                                                  derefSymlinkStr(srcPath.val)))
                                          )
                                      , okThen: {arg, okThen: {fn: withPath}}}))
                                : withPath})})}}))

      , "gensym": quote(makeFun(_.flow(gensym, val => ({val}))))

      , "symbol-debug-info": quote(fnOfType(symLabel, data => ({val: data.sym})))

      , "q": quote(makeFun(_.flow(quote, val => ({val}))))

      , "make-call"
        : quote
          ( makeFun
            ( fnExpr =>
              ({val: makeFun(argExpr => ({val: makeCall(fnExpr, argExpr)}))})))

      , "call-fn-expr": quote(fnOfType(callLabel, ({fnExpr}) => ({val: fnExpr})))

      , "call-arg-expr"
        : quote(fnOfType(callLabel, ({argExpr}) => ({val: argExpr})))

      , "make-ident": quote(makeFun(_.flow(makeIdent, val => ({val}))))

      , "ident-key"
        : quote(makeFun((ident, onOk) => mCast(ident, identKeySym, onOk)))

      , "ident-key-m": quote(identKeySym)

      , "delimited-var-sym": quote(delimitedVarSym)

      , "delimited": makeFun(arg => ({fn: delimitedVar, arg}))

      , "ok": quote(makeFun(_.flow(ok, val => ({val}))))

      , "err": quote(makeFun(_.flow(err, val => ({val}))))

      , "okp": quote(fnOfType(resultLabel, arg => ({val: makeBool(arg.ok)})))

      , "unok"
        : quote
          ( fnOfType
            ( resultLabel
            , arg =>
              arg.ok
              ? {val: arg.val}
              : { ok: false
                , val: listConcat(strToInts("Err: "), toString(arg.val))}))

      , "unerr"
        : quote
          ( fnOfType
            ( resultLabel
            , arg =>
              arg.ok
              ? { ok: false
                , val: listConcat(strToInts("Ok: "), toString(arg.val))}
              : {val: arg.val}))

      , "cons"
        : quote
          ( makeFun
            (first => ({val: makeFun(last => ({val: makePair(first, last)}))})))

      , "car": quote(fnOfType(pairLabel, ({first}) => ({val: first})))

      , "cdr": quote(fnOfType(pairLabel, ({last}) => ({val: last})))

      , "to-str-m": quote(toStrSym)

      , "strable"
        : quote
          ( makeFun
            ( fn =>
              ( { val
                  : arrToObj
                    ( [ [ toStrSym
                        , makeCont
                          ( msg =>
                            msg.type === pairLabel
                            ? [ mCall
                                ( fn
                                , applySym
                                , unit
                                , msg.data.first
                                , msg.data.last)]
                            : [])]])})))

      , "null": makeFun(() => [])}
    , expr => ({val: ok(expr)}));

Error.stackTraceLimit = Infinity;
