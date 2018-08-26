'use strict';

// Dependencies

const
  fs = require('fs')
  , util = require('util')
  , {Writable, Readable, Transform} = require('stream')

  , _ = require('lodash')
  , weak = require('weak')
  , {blue, red, bold} = require('chalk')
  //, {iterableIntoIterator} = require('list-parsing')

  , {parseStream: parser, parseIter, iterableIntoIterator}
    = require('./don-parse2.js');

// Utility
const
  debug = true
  , inspect = o => util.inspect(o, {depth: null, colors: true})
  , log
    = (...args) =>
      (debug && console.log(args.map(inspect).join("\n")), _.last(args))
  , strStream2chrStream
    = () =>
      Transform
      ( { transform(str, enc, cb)
          {Array.from(str).forEach(chr => this.push(chr, enc)); cb(null)}
        , decodeStrings: false})
      .setEncoding("utf8")
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
;

// Stuff

exports = module.exports = {parseStream: parser, parseAsyncIterable: parseIter};

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
          : strToChars
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
        ( then.fn
        , applySym
        , arg
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

function just(val) {return mk(maybeLabel, {is: true, val})}

function makeInt(Int) {return mk(intLabel, Int)}

function makeChar(codepoint) {return mk(charLabel, codepoint)}

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
                  ( maybeLabel
                  , mExpr =>
                    mExpr.is ? {fn: mExpr.val, arg} : unknownKeyThread(key))}}))
    , This
      = arrToObj
        ( [ [ applySym
            , makeCont(arg => [mCast(appFun, applySym, arg)])]
          , [identKeySym, makeCont(cont => ([{cont, arg: key}]))]]);
  return This}

function makeBool(val) {return mk(boolLabel, val)}

function okResult(val) {return mk(resultLabel, {ok: true, val})}

function errResult(val) {return mk(resultLabel, {ok: false, val})}

function makeCont(fn) {return mk(contLabel, fn)}
exports.makeCont = makeCont

function makePair(first, last) {return mk(pairLabel, {first, last})}

//function makeStrable(fn)
//{ return (
//    makeCont
//    ( msg =>
//      msg.type === pairLabel
//      && msg.data.first === toStrSym
//      && msg.data.last.type === pairLabel
//      && msg.data.last.data.first.type === contLabel
//      && msg.data.last.data.last.type === contLabel
//      ? fn(msg.data.last.data.first, msg.data.last.data.last)
//      : []))}

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
  chrStream2charStream
  = () =>
    Transform
    ( { transform(chr, enc, cb)
        {this.push(makeChar(chr.codePointAt(0))); cb(null)}
      , decodeStrings: false
      , readableObjectMode: true});

const
  charStream2asyncIter
  = cont =>
    { let execDone = Promise.resolve();
      const
        getNext
        = (res, rej) =>
          [ { cont
            , arg
              : makeCont
                ( next =>
                  next.type === maybeLabel
                  ? next.data.is
                    ? next.data.val.type === charLabel
                      ? (res(charToStr(next.data.val)), [])
                      : Null("Streamt without character")
                    : (rej(), [])
                  : Null("Stream argued with definition"))}];
      return (
        { rIter
          : ( async function*()
              { while (true)
                  yield (
                    await
                      new Promise
                      ( (...a) =>
                        execDone
                        = Promise.all([execDone, topContinue(getNext(...a))])
                          .then(_.noop)))})
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
              ( { write(...[chr,, cb])
                  { const res = prmRes;
                    topContinue
                    ([{cont: reader, arg: just(chr)}, handleThread(cb)])
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

function objToNs(o)
{ return (
    makeFun
    ( identKey =>
      ( { val
          : isString(identKey)
            ? (keyStr => o.hasOwnProperty(keyStr) ? just(o[keyStr]) : nothing)
              (strVal(identKey))
            : nothing})))}

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
        (_.findIndex(arr, p => eq(p[0], msg.data.first)))
      : Null("Bad message:\n%s\n\n%s", inspect(msg), inspect(arr))))}

function isList(val)
{ while (val.type === pairLabel) val = val.data.last;
  return val === unit}

function* listIter(list)
{ while (list.type === pairLabel) yield list.data.first, list = list.data.last;
  return list === unit}

function reverseConcat(list0, list1)
{ while (list0 !== unit)
    list1 = makePair(list0.data.first, list1), list0 = list0.data.last;
  return list1}

function listReverse(list) {return reverseConcat(list, unit)}

function listConcat(list0, list1)
{return list1 === unit ? list0 : reverseConcat(listReverse(list0), list1)}

function listsConcat(lists)
{return lists.reduceRight((list1, list0) => listConcat(list0, list1), unit)}

function isString(val)
{ return (
    isList(val)
    && _.every([...listIter(val)], elem => elem.type === charLabel))}

function charToStr(Char)
{ if (Char.type !== charLabel)
    return Null("charToStr nonchar: " + strVal(toString(Char)));
  return String.fromCodePoint(Char.data)}

function strVal(list)
{ if (!isString(list)) return Null("Tried to strVal nonlist");
  let str = "";
  while (list.type === pairLabel)
    str += charToStr(list.data.first), list = list.data.last;
  return str}
exports.strVal = strVal

function stringIs(list, str)
{return strVal(list) === str}

function strToChar(chr) {return makeChar(chr.codePointAt(0))}

function strToChars(str) {return makeList(Array.from(str).map(strToChar))}

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
      ||
        val0.type === maybeLabel
        && val0.data.is === val1.data.is
        && (!val0.data.is || eq(val0.data.val, val1.data.val))
      || val0.type === resultLabel
         && val0.data.ok === val1.data.ok
         && eq(val0.data.val, val1.data.val)
      || val0.type === pairLabel
         && eq(val0.data.first, val1.data.first)
         && eq(val0.data.last, val1.data.last)))}

function subDelimited({t: label, d: data})
{ return (
    label === 'delimited'
    ? makePair(makeInt(2), delimitedTree(data))
    : makePair(makeInt(label === 'char' ? 1 : 0), makeChar(data)));}

function delimitedTree(data)
{ return (
    makePair
    ( makePair(strToChar(data.start.delim), strToChar(data.end.delim))
    , makeList(data.inner.map(subDelimited))));}

function parseTreeToAST({t: label, d: data})
{ switch (label)
  { case 'char': return makeCall(charVar, quote(makeInt(data)));

    case 'delimited': return makeCall(delimitedVar, quote(delimitedTree(data)))

    case 'quote': return quote(parseTreeToAST(data));

    case 'ident': return makeIdent(makeList(data.map(makeChar)));

    default: return Null("unknown parse-tree type '" + label)}}

const intLabel = {label: 'int'};
exports.intLabel = intLabel;

const charLabel = {label: 'char'};
exports.charLabel = charLabel;

const symLabel = {label: 'sym'};
exports.symLabel = symLabel;

const quoteLabel = {label: 'quote'};
exports.quoteLabel = quoteLabel;

const unit = mk({label: 'unit'});
exports.unit = unit;

const callLabel = {label: 'call'};
exports.callLabel = callLabel;

const maybeLabel = {label: 'maybe'};
exports.maybeLabel = maybeLabel;

const boolLabel = {label: 'bool'};
exports.boolLabel = boolLabel;

const resultLabel = {label: 'result'};
exports.resultLabel = resultLabel;

const contLabel = {label: 'cont'};
exports.contLabel = contLabel;

const pairLabel = {label: 'pair'};
exports.pairLabel = pairLabel;

const toStrSym = gensym(strToChars('to-str-sym'));
const applySym = gensym(strToChars('apply-sym'));
const identKeySym = gensym(strToChars('ident-key-sym'));

const delimitedVarSym = gensym(strToChars('delimited-var'));
const delimitedVar = makeIdent(delimitedVarSym);
exports.delimitedVar = delimitedVar;

const charVarSym = gensym(strToChars('char-var'));
const charVar = makeIdent(charVarSym);
exports.charVar = charVar;

const Null = (...args) => {throw new Error("Null: " + util.format(...args))};
exports.Null = Null;

const nullCont = makeCont(_.constant([]));
exports.nullCont = nullCont;

const nothing = mk(maybeLabel, {is: false});

const I = makeFun(val => ({val}));

const objToNsNotFoundStr = strToChars("Var not found in ns: ");

const makeMapCall
  = makeFun
    ( arg =>
      { if (!isList(arg))
          return {ok: false, val: strToChars("Insequential cartographic call")};
        const args = [...listIter(arg)];
        if (args.length % 2 != 1)
          return {ok: false, val: strToChars("Tried to brace evenness")};
        const pairs = _.chunk(_.tail(args), 2);
        return (
          { fn: args[0]
          , arg
            : makeFun
              ( arg =>
                { let toReturn = nothing;
                  _.forEach
                  ( pairs
                  , pair =>
                    eq(arg, pair[0])
                    ? (toReturn = just(pair[1]), false)
                    : true);
                  return {val: toReturn}})})});

const
  makeFunCall
  = makeFun
    ( (elems, onOk, onErr) =>
      !isList(elems)
      ? { ok: false
        , value
          : strToChars("Insequential delimition")}
      : elems === unit
        ? {val: I}
        : { cont
            : _.reduceRight
              ( [...listIter(elems.data.last)]
              , (onOk, arg) =>
                makeCont
                ( fn =>
                  [ mCall
                    ( fn
                    , applySym
                    , arg
                    , onOk
                    , onErr)])
              , onOk)
          , arg: elems.data.first});

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
                      : strToChars("Insequential synchronous cartography")});
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
  = filename =>
    { const
        pipeFrom = fs.createReadStream(filename, {encoding: 'utf8'})
        , pipeTo = strStream2chrStream();
      return (
        {file: pipeFrom.pipe(pipeTo), cleanup() {pipeFrom.unpipe(pipeTo)}})};
exports.readFile = readFile;

//const
//  indexToLineColumn
//  = (index, string) =>
//    { const arr = Array.from(string);
//      let line = 0, col = 0, i = 0;
//      while (++i < arr.length)
//      { if (i === index)
//          return {line0: line, col0: col, line1: ++line, col1: ++col};
//        if (arr[i] === '\n') line++, col = 0;
//        else col++}
//      throw (
//        new RangeError
//        ("indexToLineColumn: index=" + index + " is out of bounds"))};

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
          return {success: false, error: _.constant("Error in the syntax")};
        else
        { const {nest, lines, line, col, expect} = parsed.result;
          let onNote = 0, color = blue;
          return (
            { success: false
            , error
              : filename =>
                (nest.length ? "Nested (outer first):" : "Not nested:")
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
                    { const theLine = lines[line];
                      let print = [[], []], idx = 0;
                      for (let col of cols)
                        onNote++ === nest.length ? color = red : 0
                        , print[0].push
                          ( ...theLine.slice(idx, col)
                          , ...col < theLine.length
                            ? [bold(color(theLine[col]))]
                            : [])
                        , print[1].push(" ".repeat(col - idx), bold(color("^")))
                        , idx = col + 1;
                      print[0].push(...theLine.slice(idx));
                      return (
                        " ┌Line "
                        + (line + 1)
                        + "; Column"
                        + (cols.length - 1 ? "s" : "")
                        + " "
                        + cols.map(col => col + 1).join()
                        + "\n │"
                        + print[0].join("")
                        + "\n └"
                        + print[1].join("")
                        + "\n")})
                  .join("")
                + ( parsed.status === 'eof'
                    ? "Syntax error: unfinished program in " + filename
                    : "Syntax error at "
                      + filename
                      + " "
                      + (line + 1)
                      + ","
                      + (col + 1))
                + "\n  Expected "
                + expect})}

    //const trace = parsed.trace;
    //_.forEachRight(trace, function(frame) {console.log("in", frame[0])});
    //console.log(parsed.parser);
    }};
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
exports.topContinue = topContinue;

const
  bindRest
  = (expr, {rest, input}) =>
    makeFun
    ( () =>
      ( { fn: makeStream(input.file.pipe(chrStream2charStream()), input.cleanup)
        , okThen
          : { fn
              : makeFun
                ( theStdin =>
                  ( justQuoteStdin =>
                    ( { fn
                        : makeStream
                          (rest.file.pipe(chrStream2charStream()), rest.cleanup)
                      , okThen
                        : { fn
                            : makeFun
                              ( theSourceData =>
                                ( justQuoteSourceData =>
                                  ( { val
                                      : makeFun
                                        ( fn =>
                                          ( { fn: expr
                                            , arg
                                              : makeFun
                                                ( arg =>
                                                  eq
                                                  ( arg
                                                  , strToChars('source-data'))
                                                  ? {val: justQuoteSourceData}
                                                  : eq
                                                    ( arg
                                                    , strToChars('stdin'))
                                                    ? {val: justQuoteStdin}
                                                    : {fn, arg}
                                                )}))}))
                                (just(quote(theSourceData))))}}))
            (just(quote(theStdin))))}}));
exports.bindRest = bindRest;

const
  delimL = Array.from('([{\\').map(o => o.codePointAt(0))
  , delimR = Array.from(')]}|').map(o => o.codePointAt(0))
  , charName
    = chr =>
      delimL.includes(chr)
      ? 'left'
      : delimR.includes(chr)
        ? 'right'
        : chr === 59 ? 'semicolon' : chr === 96 ? 'backtick' : 'other'
  , makeBacktick = () => makeChar(96);
function escInIdent(charArr)
{ let ticked = false, identStack = [[[]]], name;
  const stackLog = () => log(identStack.map(o => o.map(o => o.map(charToStr))));
  for (let chr of charArr)
    ( name = charName(chr.data)
    , ticked
      ? ( name === 'other'
          ? _.last(_.last(identStack)).push(chr)
          : _.last(identStack).push([chr])
        , ticked = false)
      : name === 'other'
        ? _.last(_.last(identStack)).push(chr)
        : name === 'left'
          ? identStack.push([[chr]])
          : identStack.length === 1
            ? identStack[0].push([chr])
            : name === 'right'
              ? _.last(identStack[identStack.length - 2]).push
                (..._.flatten(identStack.pop()), chr)
              : /* name === 'backtick' || name === 'semicolon' */
                (_.last(identStack).push([chr]), ticked = name === 'backtick')
    /*, stackLog()*/);
  return (
    _.reduce(_.flatten(identStack), (a, b) => [...a, makeBacktick(), ...b]))}

function toString(arg)
{ const {type: argLabel, data: argData} = arg;
  return (
    argLabel === charLabel ? makeList([strToChar("`"), arg])

    : argLabel === intLabel ? strToChars(argData.toString() + ' ')

    : isString(arg) && arg !== unit
      ? listsConcat
        ( _.findIndex
          ( [...listIter(arg)]
          , chr =>
            [32, 10, 9, 13, 40, 41, 91, 93, 123, 125, 96, 92, 124, 59, 35, 34]
            .includes(chr.data))
          >= 0
          ? [ strToChars("\\'")
            , makeList(escInIdent([...listIter(arg)]))
            , strToChars('|')]
          : [ strToChars("'"), arg, strToChars(' ')])

    : isList(arg)
      ? listsConcat
        ( [ strToChars('[')
          , ...[...listIter(arg)].map(o => toString(o))
          , strToChars(']')])

    : argLabel === quoteLabel
      ? listsConcat([strToChars("(q "), toString(argData), strToChars(")")])

    : argLabel === callLabel
      ? listsConcat
        ( [ strToChars("(make-call ")
          , toString(argData.fnExpr)
          , toString(argData.argExpr)
          , strToChars(")")])

    : argLabel === symLabel
      ? listsConcat
        ([strToChars("(gensym "), toString(argData.sym), strToChars(")")])

    : argLabel === maybeLabel
      ? argData.is
        ? listsConcat
          ([strToChars("(just "), toString(argData.val), strToChars(")")])
        : strToChars("nothing ")

    : argLabel === boolLabel
      ? argData ? strToChars("true ") : strToChars("false ")

    : argLabel === resultLabel
      ? listsConcat
        ( [ strToChars(argData.ok ? "(ok " : "(err ")
          , toString(argData.val)
          , strToChars(")")])

    : argLabel === contLabel ? strToChars("(cont ... )")

    : argLabel === pairLabel
      ? listsConcat
        ( [ strToChars("(cons ")
          , toString(argData.first)
          , toString(argData.last)
          , strToChars(')')])

    : Null("->str unknown type:", inspect(arg)))}
exports.toString = toString;

const
  [lParen, rParen, lBracket, rBracket, lBrace, rBrace, backslash, pipe]
  = Array.from('()[]{}\\|').map(strToChar);

const
  unknownKeyThread
  = key =>
    ( { ok: false
      , val
        : listConcat
          ( strToChars("Unknown variable key: ")
          , toString(key))});

const unicode2Char = fnOfType(intLabel, _.flow(makeChar, val => ({val})));

const
  stdin
  = () =>
    ( toChars =>
      ( { file: process.stdin.setEncoding('utf8').pipe(toChars)
        , cleanup() {process.stdin.unpipe(toChars)}}))
    (strStream2chrStream());
exports.stdin = stdin;

const
  stdout
  = ( (toWrite, writable, prmRes, nextPrmRes, prm, nextPrm) =>
      { const
          getNextPrm = () => nextPrm = new Promise(res => nextPrmRes = res)
          , writeIt
            = str =>
              { writable = false;
                prmRes = nextPrmRes, getNextPrm();
                process.stdout.write
                ( str
                , 'utf8'
                , () =>
                  { writable = true;
                    prmRes([]);
                    toWrite && writeIt((o => o)(toWrite, toWrite = ''))})};
        getNextPrm();
        return (
          str =>
          str
          ? (prm = nextPrm, writable ? writeIt(str) : toWrite += str, prm)
          : Promise.resolve())})
    ('', true);

const
  promiseWaitThread
  = prm => ({cont: makeCont(() => prm.then(_.constant([]))), arg: unit});

//const quoteFn = makeFun(_.flow(quote, val => ({val})));

const carFn = fnOfType(pairLabel, ({first}) => ({val: first}));
const cdrFn = fnOfType(pairLabel, ({last}) => ({val: last}));

const wsNums = Array.from(' \t\n\r').map(o => o.codePointAt(0))
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
          || value.data.first.data !== 0
          || value.data.last.type !== charLabel
          || [34, 59].includes(value.data.last.data)
      , doomed
        = () =>
          ( { ok: false
            , val: strToChars("Delimition enlistment doomed at index " + index)}
          )
      , eof
        = () =>
          ( { ok: false
            , val: strToChars("EOF reached while enlisting delimition")})
      , push
        = expr =>
          { if (commentLevel > 0) commentLevel--;
            else
            { while (quoteLevel-- > 0) expr = quote(expr);
              ++quoteLevel, exprs.push(expr);}};

      while (!next(yield))
      { if (doom()) return doomed();

        switch (value.data.first.data)
        { case 0 // elem
          : if (value.data.last.type !== charLabel) return doomed();
            if (wsNums.includes(value.data.last.data));
            else if (value.data.last.data == 59) ++commentLevel; // ;
            else if (value.data.last.data == 34) // "
            {if (commentLevel == 0) ++quoteLevel;}
            else
            { name = [];
              do
              { name.push(value.data.last.data);
                if (next(yield)) return eof();
                if (nameDoom()) return doomed();}
              while (!wsNums.includes(value.data.last.data));
              push(makeIdent(makeList(name.map(makeChar))));}
            break;
          case 1 // char
          : if (value.data.last.type !== charLabel) return doomed();
            push(makeCall(charVar, quote(makeInt(value.data.last.data))));
            break;
          case 2 // delimited
          : push(makeCall(delimitedVar, quote(value.data.last)));}}
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
      ? {ok: false, value: strToChars("Insequential delimition")}
      : elems === unit
        ? {val: I}
        : { cont
            : _
              .reduceRight
              ( [...listIter(elems.data.last)]
              , (onOk, arg) =>
                makeCont(fn => [mCall(fn, applySym, arg, onOk, onErr)])
              , onOk)
          , arg: elems.data.first})

const
  flattenDelimited
  = delimition =>
    { if
      ( delimition.type !== pairLabel
      || delimition.data.first.type !== pairLabel)
        return {};
      const chars = [delimition.data.first.data.first];
      for (const elem of listIter(delimition.data.last))
      { if (elem.type !== pairLabel || elem.data.first.type !== intLabel)
          return {};
        switch (elem.data.first.data)
        { case 0
          : chars.push(elem.data.last);
            break;
          case 1
          : chars.push(strToChar('`'), elem.data.last);
            break;
          case 2
          : const sub = flattenDelimited(elem.data.last)
            if (!sub.is) return {};
            chars.push(...sub.val);}}
      return {is: true, val: [...chars, delimition.data.first.data.last]};}
, delimitedIdentGen
  = function *(env)
    { let value, index = 0, commentLevel = 0;

      const chars = []
      , next
        = yielded => (++index, ({value} = yielded).done)
      , doom
        = () => value.type !== pairLabel || value.data.first.type !== intLabel
      , doomed
        = () =>
          ( { ok: false
            , val: strToChars("Identical delimition doomed at index " + index)})
      , eof
        = () =>
          ( { ok: false
            , val: strToChars("EOF reached in identical delimition")});

      while (!next(yield))
      { if (doom()) return doomed();

        switch (value.data.first.data)
        { case 0 // elem
          : if (value.data.last.type !== charLabel) return doomed();
            else if (value.data.last.data == 59) // ;
              { ++commentLevel;
                do
                { if (next(yield)) return eof();
                  if (doom() || value.data.first.data == 1) return doomed();
                  if (value.data.first.data == 2) commentLevel--;
                  else
                  { if (value.data.last.type !== charLabel) return doomed();
                    if (value.data.last.data == 59) ++commentLevel; //;
                    else if
                    (!wsNums.includes(value.data.last.data)) return doomed();}}
                while (commentLevel);
                continue;}
          case 1 // char
          : chars.push(value.data.last);
            break;
          case 2 // delimited
          : const sub = flattenDelimited(value.data.last);
            if (!sub.is) return doomed();
            chars.push(...sub.val)}}
      return {fn: makeIdent(makeList(chars)), arg: env};}
, delimitedIdent
  = (list, env) => iterableIntoIterator(delimitedIdentGen(env), listIter(list));

const
  initEnv
  = makeFun
    ( varKey =>
      varKey.type === symLabel

      ? varKey === delimitedVarSym
        ? { val
            : just
              ( makeFun
                ( env =>
                  ( { val
                      : makeFun
                        ( arg =>
                          ( { fn: carFn
                            , arg
                            , okThen
                              : { fn
                                  : makeFun
                                    ( delims =>
                                      ( { fn: carFn
                                        , arg: delims
                                        , okThen
                                          : { fn
                                              : makeFun
                                                ( begin =>
                                                  ( { fn: cdrFn
                                                    , arg: delims
                                                    , okThen
                                                      : { fn
                                                          : makeFun
                                                            ( end =>
                                                              ( { fn: cdrFn
                                                                , arg
                                                                , okThen
                                                                  : { fn
                                                                      : makeFun
                                                                        ( children =>
                                                                          eq
                                                                          ( begin
                                                                          , lParen)
                                                                          &&
                                                                            eq
                                                                            ( end
                                                                            , rParen)
                                                                          ? delimitedList
                                                                            ( children
                                                                            , env
                                                                            , parenListFn
                                                                            )
                                                                          : eq
                                                                            ( begin
                                                                            , lBracket
                                                                            )
                                                                            &&
                                                                              eq
                                                                              ( end
                                                                              , rBracket
                                                                              )
                                                                            ? delimitedList
                                                                              ( children
                                                                              , env
                                                                              , I
                                                                              )
                                                                          : eq
                                                                            ( begin
                                                                            , lBrace
                                                                            )
                                                                            &&
                                                                              eq
                                                                              ( end
                                                                              , rBrace
                                                                              )
                                                                            ? delimitedList
                                                                              ( children
                                                                              , env
                                                                              , makeMapCall
                                                                              )
                                                                          : eq
                                                                            ( begin
                                                                            , backslash
                                                                            )
                                                                            &&
                                                                              eq
                                                                              ( end
                                                                              , pipe
                                                                              )
                                                                            ? delimitedIdent
                                                                              ( children
                                                                              , env
                                                                              )
                                                                          : { ok
                                                                              : false
                                                                            , val
                                                                              : strToChars
                                                                                ( "Unspecified delimited action for "
                                                                                  + charToStr
                                                                                    (begin)
                                                                                  + charToStr
                                                                                    (end)
                                                                                )
                                                                            })
                                                                    }}))}}))}}
                                      ))}})
                          /*eq(begin, lParen) && eq(end, rParen)
                          ? {val: makeFunCall}
                          : eq(begin, lBracket) && eq(end, rBracket)
                            ? {val: I}
                          : eq(begin, lBrace) && eq(end, rBrace)
                            ? {val: makeMapCall}
                          : eq(begin, backslash) && eq(end, pipe)
                            ? {val: makeFun(arg => ({val: unit}))}
                          : { ok: false
                            , val
                              : strToChars("Unspecified delimited action")}*/)})))}

        : varKey === charVarSym
          ? {val: just(quote(unicode2Char))}
          : {val: nothing}

      : isString(varKey)

        ? ( keyStr =>
            initEnvObj.hasOwnProperty(keyStr)
            ? initEnvObj[keyStr]
            : varKey !== unit && varKey.data.first.data === "'".codePointAt(0)
              ? {val: just(quote(varKey.data.last))}

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
                ? {val: just(quote(makeInt(parseInt(keyStr, 10))))}

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
                                            ? {val: just(quote(arg))}
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
              isString(arg)
              ? [promiseWaitThread(stdout(strVal(arg))), {val: unit}]
              : {ok: false, val: strToChars('Tried to print nonstring')}))

      , "say"
        : quote
          ( makeFun
            ( _.flow
              ( toString
              , strVal
              , stdout
              , prm => [promiseWaitThread(prm), {val: unit}])))

      , "->str": quote(makeFun(_.flow(toString, val => ({val}))))

      , "char->unicode"
        : quote(fnOfType(charLabel, _.flow(makeInt, val => ({val}))))

      , "unicode->char": quote(unicode2Char)

      //, "length"
      //  : quote(fnOfType(listLabel, arg => ({val: makeInt(arg.length)})))
        
      //, "->list"
      //  : quote
      //    ( makeFun
      //      ( arg =>
      //        ( { val
      //            : fnOfType
      //              ( intLabel
      //              , length =>
      //                length < 0
      //                ? { ok: false
      //                  , val: strToChars("Lists must be nonnegative in length")
      //                  }
      //                : { fn: syncMap
      //                  , arg: makeList(_.range(length).map(makeInt))
      //                  , okThen: {fn: makeFun(fn => ({fn, arg}))}})})))

      , "true": quote(makeBool(true))

      , "false": quote(makeBool(false))

      , "unit": quote(unit)

      , "read-file"
        : quote
          ( makeFun
            ( arg =>
              isString(arg)
              ? { fn
                  : ( ({file, cleanup}) =>
                      makeStream(file.pipe(chrStream2charStream()), cleanup))
                    (readFile(strVal(arg)))}
              : {ok: false, val: strToChars('Tried to read-file of nonstring')}))

      , "parse-prog"
        : quote
          ( makeFun
            ( (...[arg,, onErr]) =>
              { const {rIter, execWait} = charStream2asyncIter(arg);
                return (
                  parseFile(rIter, parseIter).then
                  ( parsed =>
                    [ promiseWaitThread(execWait())
                    , { val
                        : parsed.success
                          ? okResult(parsed.ast)
                          : errResult
                            ( makeFun
                              ( _.flow
                                ( strVal
                                , parsed.error
                                , strToChars
                                , val => ({val}))))}]))}))

      , "eval-file"
        : quote
          ( makeFun
            ( (arg, ...ons) =>
              isString(arg)
              ? ( ({file, cleanup}) =>
                  parseFile(file, parser).then
                  ( parsed =>
                    parsed.success
                    ? { fn
                        : bindRest
                          ( parsed.ast
                          , { rest: {file, cleanup}
                            , input
                              : { file: Readable({read() {this.push(null)}})
                                , cleanup: () => 0}})
                      , okThen: {fn: makeFun(fn => ({fn, arg: initEnv}))}}
                    : {ok: false, val: strToChars(parsed.error(strVal(arg)))}))
                (readFile(strVal(arg)))
              : {ok: false, val: strToChars('Tried to eval-file of nonstring')}))

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

      , "char-var-sym": quote(charVarSym)

      , "just": quote(makeFun(_.flow(just, val => ({val}))))

      , "nothing": quote(nothing)

      , "justp": quote(fnOfType(maybeLabel, arg => ({val: makeBool(arg.is)})))

      , "unjust"
        : quote
          ( fnOfType
            ( maybeLabel
            , arg =>
              arg.is
              ? {val: arg.val}
              : {ok: false, val: strToChars("Nothing was unjustified")}))

      , "ok": quote(makeFun(_.flow(okResult, val => ({val}))))

      , "err": quote(makeFun(_.flow(errResult, val => ({val}))))

      , "okp": quote(fnOfType(resultLabel, arg => ({val: makeBool(arg.ok)})))

      , "unok"
        : quote
          ( fnOfType
            ( resultLabel
            , arg =>
              arg.ok
              ? {val: arg.val}
              : { ok: false
                , val: listConcat(strToChars("Err: "), toString(arg.val))}))

      , "unerr"
        : quote
          ( fnOfType
            ( resultLabel
            , arg =>
              arg.ok
              ? { ok: false
                , val: listConcat(strToChars("Ok: "), toString(arg.val))}
              : {val: arg.val}))

      , "cons"
        : quote
          ( makeFun
            (first => ({val: makeFun(last => ({val: makePair(first, last)}))})))

      , "car": quote(carFn)

      , "cdr": quote(cdrFn)

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
    , expr => ({val: just(expr)}));

Error.stackTraceLimit = Infinity;
