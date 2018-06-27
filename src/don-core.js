'use strict'

// Dependencies

; const
    fs = require('fs')
  , util = require('util')
  , {Writable, Readable, Transform} = require('stream')

  , _ = require('lodash')
  //, bigInt = require('big-integer')
  , weak = require('weak')

  , {parseStream: parser, parseIter} = require('./don-parse.js')

// Utility
; const
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
  //, onDemandStream
  //  = stream =>
  //    Duplex
  //    ( { read(size) {stream.pipe(this)}
  //      , write(chunk, enc, cb)
  //        {stream.unpipe(this); this.push(chunk, enc); cb(null)}
  //      , decodeStrings: false})
  //, promiseSyncMap
  //  = (arrIn, promiseFn) =>
  //      _.reduce
  //      ( arrIn
  //      , (prm, nextIn, idx) =>
  //          prm.then
  //          ( arrOut =>
  //              promiseFn(nextIn).then
  //              (newVal => (arrOut[idx] = newVal, Promise.resolve(arrOut))))
  //      , Promise.resolve(Array(arrIn.length)))

// Stuff

exports = module.exports = {parseStream: parser, parseAsyncIterable: parseIter};

; function Continue(cont, arg)
  { const
      {type: contType, data: contData} = cont
    , {type: argType, data: argData} = arg
  ; return (
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

; function mk(label, data) {return {type: label, data: data}}

; function makeCall(fnExpr, argExpr) {return mk(callLabel, {fnExpr, argExpr})}

; function fnOfType(type, fn)
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

; const
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
                  : onErr)])

; const funResToThreads
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
              , arg: res.val}]

; function makeFun(fn)
  { return (
      arrToObj
      ( [ [ applySym
          , makeCont
            ( arg =>
              { let cc, onerr
              ; return (
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

; function makeList(vals)
  {return _.reduceRight(vals, _.ary(_.flip(makePair), 2), unit)}

; function just(val) {return mk(maybeLabel, {is: true, val})}

; function makeInt(Int) {return mk(intLabel, Int)}

; function makeChar(codepoint) {return mk(charLabel, codepoint)}

; function gensym(debugId) {return mk(symLabel, {sym: debugId})}

; function makeIdent(key)
  { const
      appFun
      = makeFun
        ( arg =>
          ({fn: arg, arg: key, okThen: {fn: makeFun(fn => ({fn, arg}))}}))
    , This
      = arrToObj
        ( [ [ applySym
            , makeCont(arg => ([mCast(appFun, applySym, arg)]))]
          , [identKeySym, makeCont(cont => ([{cont, arg: key}]))]])
  ; return This}

; function makeBool(val) {return mk(boolLabel, val)}

; function okResult(val) {return mk(resultLabel, {ok: true, val})}

; function errResult(val) {return mk(resultLabel, {ok: false, val})}

; function makeCont(fn) {return mk(contLabel, fn)}
; exports.makeCont = makeCont

; function makePair(first, last) {return mk(pairLabel, {first, last})}

//; function makeStrable(fn)
//  { return (
//      makeCont
//      ( msg =>
//          msg.type === pairLabel
//          && msg.data.first === toStrSym
//          && msg.data.last.type === pairLabel
//          && msg.data.last.data.first.type === contLabel
//          && msg.data.last.data.last.type === contLabel
//          ? fn(msg.data.last.data.first, msg.data.last.data.last)
//          : []))}

; function makeChannel()
  { let mode, queue = []
  ; return [
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
      , readableObjectMode: true})

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
                  yield await new Promise
                  ( (...a) =>
                    execDone
                    = Promise.all([execDone, topContinue(getNext(...a))])
                      .then(_.noop))})
            ()
        , execWait: () => execDone})}

function makeStream(rStream, cleanup)
{ return (
    makeFun
    ( () =>
      { let prmRes, prm, reader;
        const [write, read] = makeChannel()
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
                    , prm
                    )
                  )
              }
            )
        , eosThreads
          = cont =>
            [{cont: read, arg: makeCont(eosThreads)}, {cont, arg: nothing}]
        , writable
          = Writable
            ( { write(...[chr,, cb])
                { const res = prmRes;
                  topContinue
                  ([{cont: reader, arg: just(chr)}, handleThread(cb)])
                  .then(res)
                }
              , final(cb)
                { const res = prmRes;
                  cb(null);
                  topContinue(eosThreads(reader)).then(res)
                }
              , objectMode: true
              }
            );
        weak(write, cleanup);
        return (
          [ handleThread
            ( () =>
              ( rStream.pipe(writable)
              , weak(write, () => rStream.unpipe(writable))
              )
            )
          , {val: write}
          ]
        )
      }
    )
  )
}

; function objToNs(o)
  { return (
      makeFun
      ( identKey =>
        isString(identKey)
        ? ( keyStr =>
            o.hasOwnProperty(keyStr)
            ? {val: o[keyStr]}
            : { ok: false
              , val
                : listConcat
                  (objToNsNotFoundStr, toString(makeIdent(identKey)))})
          (strVal(identKey))
        : { ok: false
          , val
            : listConcat(objToNsNotFoundStr, toString(makeIdent(identKey)))}))}

; function arrToObj(arr)
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

; function isList(val)
  { while (val.type === pairLabel) val = val.data.last
  ; return val === unit}

; function* listIter(list)
  { while (list.type === pairLabel) yield list.data.first, list = list.data.last
  ; return list === unit}

; function reverseConcat(list0, list1)
  { while (list0 !== unit)
      list1 = makePair(list0.data.first, list1), list0 = list0.data.last
  ; return list1}

; function listReverse(list) {return reverseConcat(list, unit)}

; function listConcat(list0, list1)
  {return list1 === unit ? list0 : reverseConcat(listReverse(list0), list1)}

; function listsConcat(lists)
  {return lists.reduceRight((list1, list0) => listConcat(list0, list1), unit)}

; function isString(val)
  { return (
      isList(val)
      && _.every([...listIter(val)], elem => elem.type === charLabel))}

; function charToStr(Char)
  { if (Char.type !== charLabel)
      return Null("charToStr nonchar: " + strVal(toString(Char)))
  ; return String.fromCodePoint(Char.data)}

; function strVal(list)
  { if (!isString(list)) return Null("Tried to strVal nonlist")
  ; let str = ""
  ; while (list.type === pairLabel)
      str += charToStr(list.data.first), list = list.data.last
  ; return str}
exports.strVal = strVal

; function stringIs(list, str)
  {return strVal(list) === str}

; function strToChar(chr) {return makeChar(chr.codePointAt(0))}

; function strToChars(str) {return makeList(Array.from(str).map(strToChar))}

; function mCast(cont, mSym, arg) {return {cont, arg: makePair(mSym, arg)}}

; function mCall(cont, mSym, arg, onOk, onErr)
  {return {cont, arg: makePair(mSym, makePair(onOk, makePair(arg, onErr)))}}

; function eq(val0, val1)
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

; function parseTreeToAST([label, data])
  { switch (label)
    { case 'char': return makeCall(charVar, quote(makeInt(data)))

    ; case 'delimited':
        return (
          makeCall
          ( makeCall
            ( makeCall(delimitedVar, quote(strToChar(data[0])))
            , quote(strToChar(data[2])))
          , makeFun
            ( env =>
              ( { fn: syncMap
                , arg: makeFun(expr => ({fn: expr, arg: env}))
                , okThen
                  : { fn
                      : makeFun
                        ( soFar =>
                          ( { fn: soFar
                            , arg
                              : makeList(data[1].map(parseTreeToAST))}))}}))))

      ; case 'quote': return quote(parseTreeToAST(data))

      ; case 'ident': return makeIdent(makeList(data.map(makeChar)))

      ; default: return Null("unknown parse-tree type '" + label)}}

; const intLabel = {label: 'int'}
; exports.intLabel = intLabel

; const charLabel = {label: 'char'}
; exports.charLabel = charLabel

; const symLabel = {label: 'sym'}
; exports.symLabel = symLabel

; const quoteLabel = {label: 'quote'}
; exports.quoteLabel = quoteLabel

; const unit = mk({label: 'unit'})
; exports.unit = unit

; const callLabel = {label: 'call'}
; exports.callLabel = callLabel

; const maybeLabel = {label: 'maybe'}
; exports.maybeLabel = maybeLabel

; const boolLabel = {label: 'bool'}
; exports.boolLabel = boolLabel

; const resultLabel = {label: 'result'}
; exports.resultLabel = resultLabel

; const contLabel = {label: 'cont'}
; exports.contLabel = contLabel

; const pairLabel = {label: 'pair'}
; exports.pairLabel = pairLabel

; const toStrSym = gensym(strToChars('to-str-sym'))
; const applySym = gensym(strToChars('apply-sym'))
; const identKeySym = gensym(strToChars('ident-key-sym'))

; const delimitedVarSym = gensym(strToChars('delimited-var'))
; const delimitedVar = makeIdent(delimitedVarSym)
; exports.delimitedVar = delimitedVar

; const charVarSym = gensym(strToChars('char-var'))
; const charVar = makeIdent(charVarSym)
; exports.charVar = charVar

; const Null
  = (...args) => {throw new Error("Null: " + util.format(...args))}
; exports.Null = Null

; const nullCont = makeCont(_.constant([]))
; exports.nullCont = nullCont

; const nothing = mk(maybeLabel, {is: false})

; const I = makeFun(val => ({val}))

; const objToNsNotFoundStr = strToChars("Var not found in ns: ")

; const makeMap
    = makeFun
      ( arg =>
        { if (!isList(arg))
            return {ok: false, val: strToChars("Insequential cartography")}
        ; const args = [...listIter(arg)]
        ; if (args.length % 2 != 0)
            return {ok: false, val: strToChars("Tried to brace oddity")}
        ; const pairs = _.chunk(args, 2)
        ; return (
            { val
              : makeFun
                ( arg =>
                  { let toReturn = nothing
                  ; _.forEach
                    ( pairs
                    , pair =>
                        eq(arg, pair[0])
                        ? (toReturn = just(pair[1]), false)
                        : true)
                  ; return {val: toReturn}})})})

; const
    syncMap
    = makeFun
      ( fn =>
        { const listFn
            = makeFun
              ( list =>
                { if (!isList(list))
                    return (
                      { ok: false
                      , val
                        : strToChars("Insequential synchronous cartography")})
                ; return (
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
                                                      , newTail)}))}}))}})})
        ; return {val: listFn}})

; const
    readFile
    = filename =>
      { const 
          pipeFrom = fs.createReadStream(filename, {encoding: 'utf8'})
        , pipeTo = strStream2chrStream()
      ; return (
          {file: pipeFrom.pipe(pipeTo), cleanup() {pipeFrom.unpipe(pipeTo)}})}
exports.readFile = readFile

; const indexToLineColumn
  = (index, string) =>
      { const arr = Array.from(string)
      ; let line = 0, col = 0
      ; for (let i = 0; i < arr.length; i++)
        { if (i == index)
            return {line0: line, col0: col, line1: ++line, col1: ++col}
        ; if (arr[i] === '\n') line++, col = 0
        ; else col++}
        throw new RangeError
                  ("indexToLineColumn: index=" + index + " is out of bounds")}

; const parseFile
  = async (stream, parseFn) =>
      { const parsed = await parseFn(stream)

      ; if (parsed.status === 'match')
          return {success: true, ast: parseTreeToAST(parsed.result)}
      ; else if (parsed.status === 'eof')
          return (
            { success: false
            , error
              : filename =>
                  "Syntax error: "
                  + filename
                  + " should have at least "
                  + parsed.index
                  + " codepoints"})
      ; else
        { const errAt = parsed.index
        ; if (errAt == 0)
            return {success: false, error: _.constant("Error in the syntax")}
        ; else
          { //const lineCol = indexToLineColumn(errAt - 1, data)
          ; return (
              { success: false
              , error
                : filename =>
                    "Syntax error at "
                    + filename
                    + " at codepoint #"
                    + errAt
                    //+ lineCol.line1
                    //+ ","
                    //+ lineCol.col1
                    //+ ":\n"
                    //+ data.split('\n')[lineCol.line0]
                    //+ "\n"
                    //+ " ".repeat(lineCol.col0)
                    //+ "^"
                    /*+ inspect(parsed.parser.traceStack)*/})}

        //; const trace = parsed.trace
        //; _.forEachRight(trace, function(frame) {console.log("in", frame[0])})
        //; console.log(parsed.parser)
        }}
; exports.parse = parseFile

; const topApply
  = (fn, ...stuf) => topContinue([mCall(fn, applySym, ...stuf)])
; exports.topApply = topApply

; const topContinue
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
      .then(_.noop)
; exports.topContinue = topContinue;

const bindRest
= (expr, {rest, input}) =>
  makeFun
  ( () =>
    ( { fn: makeStream(input.file.pipe(chrStream2charStream()), input.cleanup)
      , okThen
        : { fn: quoteFn
          , okThen
            : { fn
                : makeFun
                  ( quotedStdin =>
                    ( { fn
                        : makeStream
                          (rest.file.pipe(chrStream2charStream()), rest.cleanup)
                      , okThen
                        : { fn: quoteFn
                          , okThen
                            : { fn
                                : makeFun
                                  ( quotedSourceData =>
                                    ( { val
                                        : makeFun
                                          ( env =>
                                            ( { fn: expr
                                              , arg
                                                : makeFun
                                                  ( varKey =>
                                                    eq
                                                    ( varKey
                                                    , strToChars('source-data'))
                                                    ? {val: quotedSourceData}
                                                    : eq
                                                      ( varKey
                                                      , strToChars('stdin'))
                                                      ? {val: quotedStdin}
                                                      : { fn: env
                                                        , arg
                                                          : varKey
                                                        })}))}))}}}))}}}));
exports.bindRest = bindRest

; const
    cp
    = { 35: "hash"
      , 59: "semicolon"
      , 92: "backslash"
      , 96: "backtick"
      , 124: "pipe"}
  , charName = chr => cp[chr.data] || "other"
  , makeBacktick = () => makeChar(96)
; function escInIdent(charArr)
  { let ticked = false, identStack = [[[]]]
  ; const
      stackLog = () => log(identStack.map(o => o.map(o => o.map(charToStr))))
  ; for (let chr of charArr)
      ( name = charName(chr)
      , ticked
        ? ( name === "other"
            ? _.last(_.last(identStack)).push(chr)
            : _.last(identStack).push([chr])
          , ticked = false)
        : name === "other"
          ? _.last(_.last(identStack)).push(chr)
          : name === "backslash"
            ? identStack.push([[chr]])
            : identStack.length === 1
              ? identStack[0].push([chr])
              : name === "pipe"
                ? _.last(identStack[identStack.length - 2]).push
                  (..._.flatten(identStack.pop()), chr)
                : name === "hash" || name === "semicolon"
                  ? identStack = [[..._.flatten(identStack), [chr]]]
                  : /* name === "backtick" */
                    (_.last(identStack).push([chr]), ticked = true)
      /*, stackLog()*/)
  ; return (
      _.reduce(_.flatten(identStack), (a, b) => [...a, makeBacktick(), ...b]))}

; function toString(arg)
  { const {type: argLabel, data: argData} = arg
  ; if (argLabel === charLabel) return makeList([strToChar("`"), arg])

  ; if (argLabel === intLabel) return strToChars(argData.toString() + ' ')

  ; if (isString(arg) && arg !== unit)
      return (
        listsConcat
        ( _.findIndex
          ( [...listIter(arg)]
          , chr =>
            [32, 10, 9, 13, 40, 41, 91, 93, 123, 125, 96, 92, 124, 59, 35, 34]
            .includes(chr.data))
          >= 0
          ? [ strToChars("\\'")
            , makeList(escInIdent([...listIter(arg)]))
            , strToChars('|')]
          : [ strToChars("'"), arg, strToChars(' ')]))

  ; if (isList(arg))
      return (
        listsConcat
        ( [ strToChars('[')
          , ...[...listIter(arg)].map(o => toString(o))
          , strToChars(']')]))

  ; if (argLabel === quoteLabel)
      return (
        listsConcat([strToChars("(q "), toString(argData), strToChars(")")]))

  ; if (argLabel === callLabel)
      return (
        listsConcat
        ( [ strToChars("(make-call ")
          , toString(argData.fnExpr)
          , toString(argData.argExpr)
          , strToChars(")")]))

  ; if (argLabel === symLabel)
      return (
        listsConcat
        ([strToChars("(gensym "), toString(argData.sym), strToChars(")")]))

  ; if (argLabel === maybeLabel)
      return (
        argData.is
        ? listsConcat
          ([strToChars("(just "), toString(argData.val), strToChars(")")])
        : strToChars("nothing "))

  ; if (argLabel === boolLabel)
      return argData ? strToChars("true ") : strToChars("false ")

  ; if (argLabel === resultLabel)
      return (
        listsConcat
        ( [ strToChars(argData.ok ? "(ok " : "(err ")
          , toString(argData.val)
          , strToChars(")")]))

  ; if (argLabel === contLabel) return strToChars("(cont ... )")

  ; if (argLabel === pairLabel)
      return (
        listsConcat
        ( [ strToChars("(cons ")
          , toString(argData.first)
          , toString(argData.last)
          , strToChars(')')]))

  ; return Null("->str unknown type:", inspect(arg))}
; exports.toString = toString

; const
    [lParen, rParen, lBracket, rBracket, lBrace, rBrace]
    = Array.from('()[]{}').map(strToChar)

; const unknownKeyThread
  = key =>
    ( { ok: false
      , val
        : listConcat
          ( strToChars("Unknown variable key: ")
          , toString(key))})

; const unicode2Char = fnOfType(intLabel, _.flow(makeChar, val => ({val})))

; const stdin
  = () =>
    ( toChars =>
      ( { file: process.stdin.setEncoding('utf8').pipe(toChars)
        , cleanup() {process.stdin.unpipe(toChars)}}))
    (strStream2chrStream())
; exports.stdin = stdin

; const stdout
  = ( (toWrite, writable, prmRes, nextPrmRes, prm, nextPrm) =>
      { const
          getNextPrm = () => nextPrm = new Promise(res => nextPrmRes = res)
          , writeIt
            = str =>
              { writable = false
              ; prmRes = nextPrmRes, getNextPrm()
              ; process.stdout.write
                ( str
                , 'utf8'
                , () =>
                  { writable = true
                  ; prmRes([])
                  ; toWrite && writeIt((o => o)(toWrite, toWrite = ''))})}
      ; getNextPrm()
      ; return (
          str =>
          str
          ? (prm = nextPrm, writable ? writeIt(str) : toWrite += str, prm)
          : Promise.resolve())})
    ('', true)

; const promiseWaitThread
  = prm => ({cont: makeCont(() => prm.then(_.constant([]))), arg: unit});

const quoteFn = makeFun(_.flow(quote, val => ({val})));

  const initEnv
  = makeFun
    ( varKey =>
      varKey.type === symLabel

      ? varKey === delimitedVarSym
        ? { val
            : quote
              ( makeFun
                ( begin =>
                    ( { val
                        : makeFun
                          ( end =>
                              eq(begin, lParen) && eq(end, rParen)
                              ? { val
                                  : makeFun
                                    ( (elems, onOk, onErr) =>
                                        !isList(elems)
                                        ? { ok: false
                                          , value
                                            : strToChars
                                              ("Insequential delimition")}
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
                                            , arg: elems.data.first})}
                              : eq(begin, lBracket) && eq(end, rBracket)
                                ? {val: I}
                              : eq(begin, lBrace) && eq(end, rBrace)
                                ? {val: makeMap}
                              : { ok: false
                                , val
                                  : strToChars
                                    ( "Unspecified delimited "
                                      + "action")})})))}

        : varKey === charVarSym ? {val: quote(unicode2Char)}

          : unknownKeyThread(varKey)

      : isString(varKey)

        ? stringIs(varKey, 'fn')
          ? { val
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
                                                  ? {val: quote(arg)}
                                                  : { fn: env
                                                    , arg: varKey})}))}))}))}))}

          : stringIs(varKey, "apply-m") ? {val: quote(applySym)}

          : stringIs(varKey, 'continue')
            ? { val
                : quote
                  ( makeFun
                    (cont => ({val: makeFun(arg => ({cont, arg}))})))}

          : stringIs(varKey, 'call/cc')
            ? {val: quote(makeFun((fn, arg) => ({fn, arg})))}

          : stringIs(varKey, 'call/onerr')
            ? {val: quote(makeFun((...[fn,, arg]) => ({fn, arg})))}

          : stringIs(varKey, 'make-cont')
            ? { val
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
                                              ( fn
                                              , applySym
                                              , arg
                                              , nullCont
                                              , onErr)])}))})))}

          : stringIs(varKey, 'make-channel')
            ? {val: makeFun(() => ({val: makePair(...makeChannel())}))}

          : stringIs(varKey, "async")
            ? { val
                : makeFun
                  ( arg =>
                      ( { val
                          : makeFun
                            ( fn =>
                                [ {fn, arg, onOk: nullCont}
                                , {val: unit}])}))}

          : stringIs(varKey, '+')
            ? { val
                : quote
                  ( fnOfType
                    ( intLabel
                    , int0 =>
                        ( { val
                            : fnOfType
                              ( intLabel
                              , int1 =>
                                  ({val: makeInt(int0 + int1)}))})))}

          : stringIs(varKey, '-')
            ? { val
                : quote
                  ( fnOfType
                    ( intLabel
                    , int0 =>
                        ( { val
                            : fnOfType
                              ( intLabel
                              , int1 =>
                                  ({val: makeInt(int0 - int1)}))})))}

          : stringIs(varKey, '<')
            ? { val
                : quote
                  ( fnOfType
                    ( intLabel
                    , arg0 =>
                        ( { val
                            : fnOfType
                              ( intLabel
                              , arg1 =>
                                ({val: makeBool(arg0 < arg1)}))})))}

          : stringIs(varKey, '=')
            ? { val
                : quote
                  ( makeFun
                    ( arg0 =>
                        ( { val
                            : makeFun
                              ( arg1 =>
                                  ( { val
                                      : makeBool(eq(arg0, arg1))}))})))}

          : stringIs(varKey, "env") ? {val: I}

          : stringIs(varKey, "init-env") ? {val: quote(initEnv)}

          : stringIs(varKey, "print")
            ? { val
                : quote
                  ( makeFun
                    ( arg =>
                        isString(arg)
                        ? [ promiseWaitThread(stdout(strVal(arg)))
                          , {val: unit}]
                        : { ok: false
                          , val
                            : strToChars('Tried to print nonstring')}))}

          : stringIs(varKey, "say")
            ? { val
                : quote
                  ( makeFun
                    ( _.flow
                      ( toString
                      , strVal
                      , stdout
                      , prm => [promiseWaitThread(prm), {val: unit}])))}

          : stringIs(varKey, "->str")
            ? {val: quote(makeFun(_.flow(toString, val => ({val}))))}

          : stringIs(varKey, "char->unicode")
            ? { val
                : quote
                  ( fnOfType
                    (charLabel, _.flow(makeInt, val => ({val}))))}

          : stringIs(varKey, "unicode->char") ? {val: quote(unicode2Char)}

          //: stringIs(varKey, "length")
          //  ? { val
          //      : quote
          //        ( fnOfType
          //          (listLabel, arg => ({val: makeInt(arg.length)})))}

          //: stringIs(varKey, "->list")
          //  ? { val
          //      : quote
          //        ( makeFun
          //          ( arg =>
          //              ( { val
          //                  : fnOfType
          //                    ( intLabel
          //                    , length =>
          //                        length < 0
          //                        ? { ok: false
          //                          , val
          //                            : strToChars
          //                              ( "Lists must be nonnegative in"
          //                                + " length")}
          //                        : { fn: syncMap
          //                          , arg
          //                            : makeList
          //                              (_.range(length).map(makeInt))
          //                          , okThen
          //                            : { fn
          //                                : makeFun
          //                                  (fn => ({fn, arg}))}})})))}

          : stringIs(varKey, "true") ? {val: quote(makeBool(true))}

          : stringIs(varKey, "false") ? {val: quote(makeBool(false))}

          : stringIs(varKey, "unit") ? {val: quote(unit)}

          : stringIs(varKey, "read-file")
            ? { val
                : quote
                  ( makeFun
                    ( arg =>
                        isString(arg)
                        ? { fn
                            : ( ({file, cleanup}) =>
                                makeStream
                                (file.pipe(chrStream2charStream()), cleanup))
                              (readFile(strVal(arg)))}
                        : { ok: false
                          , val
                            : strToChars
                              ('Tried to read-file of nonstring')}))}

          : stringIs(varKey, "parse-prog")
            ? { val
                : quote
                  ( makeFun
                    ( (...[arg,, onErr]) =>
                      { const {rIter, execWait} = charStream2asyncIter(arg)
                      ; return (
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
                                        , val => ({val}))))}]))}))}

          : stringIs(varKey, "eval-file")
            ? { val
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
                                            : { file
                                                : Readable
                                                  ({read() {this.push(null)}})
                                              , cleanup: ()=>0}})
                                    , okThen
                                      : { fn
                                          : makeFun
                                            (fn => ({fn, arg: initEnv}))}}
                                    
                                  : { ok: false
                                    , val
                                      : strToChars
                                        (parsed.error(strVal(arg)))}))
                          (readFile(strVal(arg)))
                        : { ok: false
                          , val
                            : strToChars
                              ('Tried to eval-file of nonstring')}))}

          : stringIs(varKey, "gensym")
            ? {val: quote(makeFun(_.flow(gensym, val => ({val}))))}

          : stringIs(varKey, "symbol-debug-info")
            ? { val
                : quote(fnOfType(symLabel, data => ({val: data.sym})))}

          : stringIs(varKey, "q")
            ? {val: quote(makeFun(_.flow(quote, val => ({val}))))}

          : stringIs(varKey, "make-call")
            ? { val
                : quote
                  ( makeFun
                    ( fnExpr =>
                        ( { val
                            : makeFun
                              ( argExpr =>
                                  ( { val
                                      : makeCall
                                        (fnExpr, argExpr)}))})))}

          : stringIs(varKey, "call-fn-expr")
            ? { val
                : quote
                  (fnOfType(callLabel, ({fnExpr}) => ({val: fnExpr})))}

          : stringIs(varKey, "call-arg-expr")
            ? { val
                : quote
                  ( fnOfType
                    (callLabel, ({argExpr}) => ({val: argExpr})))}

          : stringIs(varKey, "make-ident")
            ? {val: quote(makeFun(_.flow(makeIdent, val => ({val}))))}

          : stringIs(varKey, "ident-key")
            ? { val
                : quote
                  ( makeFun
                    ((ident, onOk) => mCast(ident, identKeySym, onOk)))}

          : stringIs(varKey, "ident-key-m") ? {val: quote(identKeySym)}

          : stringIs(varKey, "delimited-var-sym")
            ? {val: quote(delimitedVarSym)}

          : stringIs(varKey, "char-var-sym") ? {val: quote(charVarSym)}

          : stringIs(varKey, "just")
            ? {val: quote(makeFun(_.flow(just, val => ({val}))))}

          : stringIs(varKey, "nothing") ? {val: quote(nothing)}

          : stringIs(varKey, "justp")
            ? { val
                : quote
                  ( fnOfType
                    (maybeLabel, arg => ({val: makeBool(arg.is)})))}

          : stringIs(varKey, "unjust")
            ? { val
                : quote
                  ( fnOfType
                    ( maybeLabel
                    , arg =>
                        arg.is
                        ? {val: arg.val}
                        : { ok: false
                          , val
                            : strToChars("Nothing was unjustified")}))}

          : stringIs(varKey, "ok")
            ? {val: quote(makeFun(_.flow(okResult, val => ({val}))))}

          : stringIs(varKey, "err")
            ? {val: quote(makeFun(_.flow(errResult, val => ({val}))))}

          : stringIs(varKey, "okp")
            ? { val
                : quote
                  ( fnOfType
                    (resultLabel, arg => ({val: makeBool(arg.ok)})))}

          : stringIs(varKey, "unok")
            ? { val
                : quote
                  ( fnOfType
                    ( resultLabel
                    , arg =>
                        arg.ok
                        ? {val: arg.val}
                        : { ok: false
                          , val
                            : listConcat
                              (strToChars("Err: "), toString(arg.val))}))}

          : stringIs(varKey, "unerr")
            ? { val
                : quote
                  ( fnOfType
                    ( resultLabel
                    , arg =>
                        arg.ok
                        ? { ok: false
                          , val
                            : listConcat(strToChars("Ok: "), toString(arg.val))}
                        : {val: arg.val}))}

          : stringIs(varKey, "cons")
            ? { val
              : quote
                ( makeFun
                  ( first =>
                      ( { val
                          : makeFun
                            ( last =>
                              ({val: makePair(first, last)}))})))}

          : stringIs(varKey, "car")
            ? { val
                : quote
                  (fnOfType(pairLabel, ({first}) => ({val: first})))}

          : stringIs(varKey, "cdr")
            ? { val
                : quote(fnOfType(pairLabel, ({last}) => ({val: last})))}

          : stringIs(varKey, "to-str-m") ? {val: quote(toStrSym)}

          : stringIs(varKey, "strable")
            ? { val
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
                                        : [])]])})))}

          : stringIs(varKey, "null") ? {val: makeFun(() => [])}

          : varKey !== unit && varKey.data.first.data === "'".codePointAt(0)
            ? {val: quote(varKey.data.last)}

//          var varParts = maybeStr[1].split(':');
//          if (varParts.length >= 2) {
//            return _.reduce(varParts.slice(1, varParts.length),
//                            function(fn, argument) {
//                              return apply(fn,
//                                           valObj(strLabel,
//                                                  argument))},
//                            apply(env,
//                                  valObj(strLabel, varParts[0])))}

          : (varStr =>
              /^(\-|\+)?[0-9]+$/.test(varStr)
              ? {val: quote(makeInt(parseInt(varStr, 10)))}
              : unknownKeyThread(varKey))
            (strVal(varKey))
      : unknownKeyThread(varKey))
; exports.initEnv = initEnv

//; Error.stackTraceLimit = Infinity

