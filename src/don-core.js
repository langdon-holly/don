'use strict'

// Dependencies

; const
    fs = require('fs')
  , util = require('util')

  , _ = require('lodash')
  //, bigInt = require('big-integer')

  , ps = require('list-parsing')
  , parser = require('./don-parse.js')

// Utility
; const
    debug = true
  , log = (...args) => (debug && console.log(...args), _.last(args))
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

; exports = module.exports

; function apply(fn, ...ons)
  { return (
      fn.type === fnLabel
      ? fn.data(...ons)
      : Continue(fn, makeList(ons)))}
; exports.apply = apply

; function Continue(cont, arg)
  { const
      {type: contType, data: contData} = cont
    , {type: argType, data: argData} = arg
  ; return (
      contType === contLabel ? contData(arg)
      : contType === fnLabel
        ? _.isArray(argData)
          && argData[1].type === contLabel
          && argData[2].type === contLabel
          ? apply(cont, ...argData)
          : Null("Fun requires arrayed continuations")
        : contType === listLabel
          ? [ { cont
                : fnOfType
                  ( intLabel
                  , idx =>
                      idx < 0 || idx >= contData.length
                      ? {ok: false, val: strToChars("Array index out of bounds")}
                      : {val: contData[idx]})
              , arg}]
          : contType === quoteLabel
            ? [{cont: makeFun(_.constant({val: contData})), arg}]
            : contType === callLabel
              ? [ { cont
                    : makeFun
                      ( (arg, ...ons) =>
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
              : contType === identLabel
                ? [ { cont
                      : makeFun
                        ( arg =>
                            ( { fn: arg
                              , arg: cont
                              , okThen
                                : {fn: makeFun(expr => ({fn: expr, arg}))}}))
                    , arg}]
                : contType === boolLabel
                  ? [ { cont
                        : makeFun
                          ( val =>
                              ( { val
                                : contData ? makeFun(_.constant({val})) : I}))
                      , arg}]
                  : Null("Tried to continue a non-continuation"))}

; function mk(label, data) {return {type: label, data: data}}

; function makeCall(fnExpr, argExpr) {return mk(callLabel, {fnExpr, argExpr})}

; function fnOfType(type, fn)
  { return (
      makeFun
      ( (arg, ...ons) =>
          arg.type !== type
          ? {ok: false, val: strToChars("typed function received garbage")}
          : fn(arg.data, ...ons)))}

; const
    makeThenOns
    = (then, onOk, onErr) =>
        makeCont
        ( arg => 
            [ { cont: then.fn
              , arg
                : makeList
                  ( [ arg
                    , then.hasOwnProperty('onOk')
                      ? then.onOk
                      : then.hasOwnProperty('okThen')
                        ? makeThenOns(then.okThen, onOk, onErr)
                        : onOk
                    , then.hasOwnProperty('onErr')
                      ? then.onErr
                      : then.hasOwnProperty('errThen')
                        ? makeThenOns(then.errThen, onOk, onErr)
                        : onErr])}])

; const funResToThreads
  = (res, onOk, onErr) =>
      _.isArray(res)
      ? _.flatMap(res, res => funResToThreads(res, onOk, onErr))
      : res.hasOwnProperty('cont')
        ? [res]
        : res.hasOwnProperty('fn')
          ? [ { cont: res.fn
              , arg
                : makeList
                  ( [ res.hasOwnProperty('arg') ? res.arg : unit
                    , res.hasOwnProperty('onOk')
                      ? res.onOk
                      : res.hasOwnProperty('okThen')
                        ? makeThenOns(res.okThen, onOk, onErr)
                        : onOk
                    , res.hasOwnProperty('onErr')
                      ? res.onErr
                      : res.hasOwnProperty('errThen')
                        ? makeThenOns(res.errThen, onOk, onErr)
                        : onErr])}]
          : [ { cont: !res.hasOwnProperty('ok') || res.ok ? onOk : onErr
              , arg: res.val}]

; function makeFun(fn)
  { return (
      makeFn
      ( (arg, onOk, onErr) =>
          funResToThreads(fn(arg, onOk, onErr), onOk, onErr)))}

; function makeFn(fn) {return mk(fnLabel, fn)}

//; function constFn(val) {return makeFn(_.constant(val))}

; function quote(val) {return mk(quoteLabel, val)}

; function makeList(vals) {return mk(listLabel, vals)}

; function just(val) {return mk(maybeLabel, {is: true, val})}

; function makeInt(Int) {return mk(intLabel, Int)}

; function makeChar(codepoint) {return mk(charLabel, codepoint)}

; function gensym(debugId) {return mk(symLabel, {sym: debugId})}

; function makeIdent(val) {return mk(identLabel, val)}

; function makeBool(val) {return mk(boolLabel, val)}

; function okResult(val) {return mk(resultLabel, {ok: true, val})}

; function errResult(val) {return mk(resultLabel, {ok: false, val})}

; function makeCell(val) {return mk(cellLabel, {val})}

; function makeCont(fn) {return mk(contLabel, fn)}
; exports.makeCont = makeCont

; function makePair(first, last) {return mk(pairLabel, {first, last})}

; const objToNsNotFoundStr = "Var not found in ns"

; function objToNs(o)
  { return (
      fnOfType
      ( identLabel
      , identKey =>
        { if (!isString(identKey))
            return {ok: false, val: strToChars(objToNsNotFoundStr)}
        ; const keyStr = strVal(identKey)
        ; return (
            o.hasOwnProperty(keyStr)
            ? {val: o[keyStr]}
            : { ok: false
              , val
                : strToChars
                  ( objToNsNotFoundStr
                    + ": "
                    + strVal(toString(makeIdent(identKey))))})}))}

; function isString(val)
  { return (
      val.type === listLabel
      && _.every(val.data, elem => elem.type === charLabel))}

; function strVal(list)
  { if (list.type !== listLabel) return Null("Tried to strVal nonlist")
  ; return list.data.reduce((soFar, chr) => soFar + charToStr(chr), '')}
exports.strVal = strVal

; function stringIs(list, str)
  {return strVal(list) === str}

; function charToStr(Char)
  { if (Char.type !== charLabel)
      return Null("charToStr nonchar: " + strVal(toString(Char)))
  ; return String.fromCodePoint(Char.data)}

; function strToChar(chr) {return makeChar(chr.codePointAt(0))}

; function strToChars(str)
  {return makeList(Array.from(str).map(strToChar))}

; function eq(val0, val1)
  { return (
      val0.type === val1.type
      &&
        ( val0.data === val1.data
        ||
          val0.type === listLabel
          && val0.data.length == val1.data.length
          && _.every(val0.data, (elem, index) => eq(elem, val1.data[index]))
        || val0.type === quoteLabel && eq(val0.data, val1.data)
        || val0.type === identLabel && eq(val0.data, val1.data)
        ||
          val0.type === callLabel
          && eq(val0.data.fnExpr, val1.data.fnExpr)
          && eq(val0.data.argExr, val1.data.argExr)
        ||
          val0.type === maybeLabel
          && val0.data.is === val1.data.is
          && (!val0.data.is || eq(val0.data.val, val1.data.val)))
        || val0.type === resultLabel
           && val0.data.ok === val1.data.ok
           && eq(val0.data.val, val1.data.val)
        || val0.type === pairLabel
           && eq(val0.data.first, val1.data.first)
           && eq(val0.data.last, val1.data.last))}

; function parseTreeToAST(pt)
  { const label = pt[0]
  ; const data = pt[1]

  ; if (label === 'char') return quote(makeChar(data))
  //; if (label === 'call')
  //    return (
  //      data.length === 0
  //      ? quote(I)
  //      : _.reduce(_.map(data, parseTreeToAST), makeCall))
  ; if (label === 'delimited')
      return (
        makeCall
        ( makeCall
          ( makeCall(delimitedVar, quote(strToChar(data[0])))
          , quote(strToChar(data[2])))
        , makeFun
          ( env =>
              ( { fn: syncMap
                , arg: makeList(data[1].map(parseTreeToAST))
                , okThen
                  : { fn
                      : makeFun
                        ( soFar =>
                            ( { fn: soFar
                              , arg
                                : makeFun
                                  (expr => ({fn: expr, arg: env}))}))}}))))
  //; if (label === 'heredoc')
  //    return quote(makeList(data.map(parseTreeToAST)))

  ; if (label === 'quote') return quote(parseTreeToAST(data))

  ; if (label === 'ident') return makeIdent(makeList(data.map(makeChar)))

  ; return Null('unknown parse-tree type "' + label)}

; function parseStr(str)
  { const parsed = parser(str)
  ; return (
      parsed.status === 'match'
      ? _.assign({ast: parseTreeToAST(parsed.result)}, parsed)
      : parsed)}

//; function ttyLog()
//  { if (process.stdout.isTTY) console.log(...arguments)}

//; const labels
//  = _.fromPairs
//    ( [ 'fn'
//      , 'list'
//      , 'int'
//      , 'char'
//      , 'sym'
//      , 'quote'
//      , 'unit'
//      , 'call'
//      , 'ident'
//      , 'maybe'
//      , 'bool'
//      , 'result'
//      , 'cell'
//      , 'cont']
//      .map(name => [name + 'Label', Symbol(name)]))
//console.log(labels)
//
//; const
//    { fnLabel
//    , listLabel
//    , intLabel
//    , charLabel
//    , symLabel
//    , quoteLabel
//    , unitLabel
//    , callLabel
//    , identLabel
//    , maybeLabel
//    , boolLabel
//    , resultLabel
//    , cellLabel
//    , contLabel}
//    = labels

; const fnLabel = {label: 'fn'}
; exports.fnLabel = fnLabel

; const listLabel = {label: 'list'}
; exports.listLabel = listLabel

; const intLabel = {label: 'int'}
; exports.intLabel = intLabel

; const charLabel = {label: 'char'}
; exports.charLabel = charLabel

; const symLabel = {label: 'sym'}
; exports.symLabel = symLabel

; const quoteLabel = {label: 'quote'}
; exports.quoteLabel = quoteLabel

; const unitLabel = {label: 'unit'}
; exports.unitLabel = unitLabel
; const unit = mk(unitLabel)
; exports.unit = unit

; const callLabel = {label: 'call'}
; exports.callLabel = callLabel

; const identLabel = {label: 'ident'}
; exports.identLabel = identLabel

; const maybeLabel = {label: 'maybe'}
; exports.maybeLabel = maybeLabel

; const boolLabel = {label: 'bool'}
; exports.boolLabel = boolLabel

; const resultLabel = {label: 'result'}
; exports.resultLabel = resultLabel

; const cellLabel = {label: 'cell'}
; exports.cellLabel = cellLabel

; const contLabel = {label: 'cont'}
; exports.contLabel = contLabel

; const pairLabel = {label: 'pair'}
; exports.pairLabel = pairLabel

; const delimitedVarSym = gensym(strToChars('delimited-var'))
; const delimitedVar = makeIdent(delimitedVarSym)
; exports.delimitedVar = delimitedVar

; const Null
  = (...args) => {throw new Error("Null: " + util.format(...args))}
; exports.Null = Null

; const nullCont = makeCont(_.constant([]))
; exports.nullCont = nullCont

; const nothing = mk(maybeLabel, {is: false})

; const I = makeFun(val => ({val}))

; const makeMap
    = fnOfType
      ( listLabel
      , args =>
        { if (args.length % 2 != 0)
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
    = fnOfType
      ( listLabel
      , arrIn =>
          ( { val
              : makeFun
                ( (fn, ...ons) =>
                    ( arrOut =>
                        _.reduceRight
                        ( arrIn
                        , (doNext, nextIn, idx) =>
                            ( { fn
                              , arg: nextIn
                              , okThen
                                : { fn
                                    : makeFun
                                      ( newVal =>
                                          (arrOut[idx] = newVal, doNext))}})
                        , {val: makeList(arrOut)}))
                    (Array(arrIn.length)))}))

; const readFile = filename => fs.readFileSync(filename, 'utf8')

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
  = data =>
      { const parsed = parseStr(data)

      ; if (parsed.status === 'match')
          return {success: true, ast: parsed.ast, rest: parsed.rest}
      ; else if (parsed.status === 'eof')
          return (
            { success: false
            , error
              : filename =>
                  "Syntax error: "
                  + filename
                  + " should have at least "
                  + (Array.from(data).length + 1)
                  + " codepoints"})
      ; else
        { const errAt = parsed.index
        ; if (errAt == 0)
            return {success: false, error: _.constant("Error in the syntax")}
        ; else
          { const lineCol = indexToLineColumn(errAt - 1, data)
          ; return (
              { success: false
              , error
                : filename =>
                    "Syntax error at "
                    + filename
                    + " "
                    + lineCol.line1
                    + ","
                    + lineCol.col1
                    + ":\n"
                    + data.split('\n')[lineCol.line0]
                    + "\n"
                    + " ".repeat(lineCol.col0)
                    + "^"
                    /*+ util.inspect(parsed.parser.traceStack, {depth: null})*/})}

        //; const trace = parsed.trace
        //; _.forEachRight(trace, function(frame) {console.log("in", frame[0])})
        //; console.log(parsed.parser)
        }}
; exports.parse = parseFile

; const srcDataIdent = makeIdent(strToChars('source-data'))

; const topApply
  = (fn, ...stuf) => topContinue([{cont: fn, arg: makeList(stuf)}])
; exports.topApply = topApply

; const threadToList = t => [t.cont, t.arg]

; const topContinue
  = threads =>
    {while (threads.length > 0) threads.push(...Continue(...threadToList(threads.shift())))}
; exports.topContinue = topContinue

; const evalProgram
  = (expr, rest) => ({fn: bindRest(expr, rest), arg: initEnv})

; const bindRest
  = (expr, rest) =>
      ( quotedSourceDataVal =>
          makeFun
          ( env =>
              ( { fn: expr
                , arg
                  : makeFun
                    ( Var =>
                        eq(Var, srcDataIdent)
                        ? quotedSourceDataVal
                        : {fn: env, arg: Var})})))
      ({val: quote(strToChars(rest))})
; exports.bindRest = bindRest

; function escInIdent(charArr)
  { return (
      _.flatMap
      ( charArr
      , chr =>
          chr.data == 92 || chr.data == 124
          ? [makeChar(92), chr]
          : [chr]))}

; function toString(arg)
  { const argLabel = arg.type, argData = arg.data
  ; if (argLabel === charLabel) return makeList([strToChar("`"), arg])

  ; if (argLabel === intLabel) return strToChars(argData.toString() + ' ')

  ; if (argLabel === listLabel)
      return (
        makeList
        ( argData.length > 0 && isString(arg)
          ? strToChars('|"').data.concat(escInIdent(argData), [strToChar('|')])
          : [strToChar('[')].concat
            ( _.reduce
              ( argData.map(o => toString(o).data)
              , (soFar, elem) => soFar.concat(elem)
              , [])
            , [strToChar(']')])))

  ; if (argLabel === quoteLabel)
      return makeList([makeChar(34)].concat(toString(argData).data))

  ; if (argLabel === unitLabel) return strToChars('unit ')

  ; if (argLabel === callLabel)
      return (
        makeList
        ( strToChars("(make-call ").data.concat
          ( toString(argData.fnExpr).data
          , toString(argData.argExpr).data
          , [strToChar(")")])))

  ; if (argLabel === identLabel)
      return (
        makeList
        ( isString(argData)
          ? strToChars('"|').data.concat
            (escInIdent(argData.data), [strToChar("|")])
          : strToChars("(make-ident ").data.concat
            (toString(argData).data, [strToChar(")")])))

  ; if (argLabel === fnLabel) return strToChars("(fn ... )")

  ; if (argLabel === symLabel)
      return (
        makeList
        ( strToChars("(gensym ").data.concat
          (toString(argData.sym).data , [strToChar(")")])))

  ; if (argLabel === maybeLabel)
      return (
        argData.is
        ? makeList
          ( strToChars("(just ").data.concat
            (toString(argData.val).data, [strToChar(")")]))
        : strToChars("nothing "))

  ; if (argLabel === boolLabel)
      return argData ? strToChars("true ") : strToChars("false ")

  ; if (argLabel === resultLabel)
      return (
        makeList
        ( strToChars(argData.ok ? "(ok " : "(err ").data.concat
          (toString(argData.val).data, [strToChar(")")])))

  ; if (argLabel === cellLabel)
      return (
        makeList
        ( strToChars("(make-cell ").data.concat
          (toString(argData.val).data, [strToChar(")")])))

  ; if (argLabel === contLabel) return strToChars("(cont ... )")

  ; if (argLabel === pairLabel)
      return (
        makeList
        ( strToChars("(cons ").data.concat
          ( toString(argData.first).data
          , toString(argData.last).data
          , [strToChar(')')])))

  ; return Null("->str unknown type:", arg)}
; exports.toString = toString

; const makeContOnOk
  = cont =>
      makeCont
      ( res =>
          res.type !== listLabel
          ? [{cont, arg: strToChars("cont body must return list")}]
          : res.data.length !== 2
            ? [{cont, arg: strToChars("cont body must return doubleton list")}]
            : [{cont: res.data[0], arg: res.data[1]}])

; const makeSyncContOnOk
  = (state, onErr) =>
      makeCont
      ( res =>
          ( state.fn = res
          , state.queue.length > 0
            ? [ { cont: state.fn
                , arg
                  : makeList
                    ( [ state.queue.shift()
                      , makeSyncContOnOk(state, onErr)
                      , onErr])}]
            : (state.busy = false, [])))

; const
    [lParen, rParen, lBracket, rBracket, lBrace, rBrace]
    = Array.from('()[]{}').map(strToChar)

; const initEnv
  = fnOfType
    ( identLabel
    , varKey =>
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
                                    : fnOfType
                                      ( listLabel
                                        , (elems, onOk, onErr) =>
                                            elems.length === 0
                                            ? {val: I}
                                            : { cont
                                                : _.reduceRight
                                                  ( _.tail(elems)
                                                  , (onOk, arg) =>
                                                      makeCont
                                                      ( cont =>
                                                          [ { cont
                                                            , arg
                                                              : makeList
                                                                ( [ arg
                                                                  , onOk
                                                                  , onErr])}])
                                                  , onOk)
                                              , arg: elems[0]})}
                                : eq(begin, lBracket) && eq(end, rBracket)
                                  ? {val: I}
                                : eq(begin, lBrace) && eq(end, rBrace)
                                  ? {val: makeMap}
                                : { ok: false
                                  , val
                                    : strToChars
                                      ("Unspecified delimited action")})})))}

          : { ok: false
            , val: strToChars("Symbol variable not found in environment")}

        : isString(varKey)

          ? stringIs(varKey, 'fn')
            ? { val
                : makeFun
                  ( env =>
                      ( { val
                          : makeFun
                            ( param =>
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
                                                              eq(varKey, param)
                                                              ? {val: quote(arg)}
                                                              : { fn: env
                                                                , arg
                                                                  : varKey})}))}))}))}))}

            : stringIs(varKey, 'continue')
              ? { val
                  : quote
                    (makeFun(cont => ({val: makeFun(arg => ({cont, arg}))})))}

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
                                              [ { cont: fn
                                                , arg
                                                  : makeList
                                                    ( [ arg
                                                      , makeContOnOk(onErr)
                                                      , onErr])}])}))})))}

            : stringIs(varKey, 'make-sync-cont')
              ? { val
                  : quote
                    ( makeFun
                      ( onErr =>
                          ( { val
                              : makeFun
                                ( fn =>
                                    ( state =>
                                        ( { val
                                            : makeCont
                                              ( arg =>
                                                  state.busy
                                                  ? (state.queue.push(arg), [])
                                                  : ( state.busy = true
                                                    , [ { cont: state.fn
                                                        , arg
                                                          : makeList
                                                            ( [ arg
                                                              , makeSyncContOnOk
                                                                (state, onErr)
                                                              , onErr])}]))}))
                                    ({fn, busy: false, queue: []}))})))}

            : stringIs(varKey, "async")
              ? { val
                  : makeFun
                    ( arg =>
                        ( { val
                            : makeFun
                              ( fn =>
                                  [{fn, arg, onOk: nullCont}, {val: unit}])}))}

            : stringIs(varKey, '+')
              ? { val
                  : quote
                    ( fnOfType
                      ( listLabel
                      , args =>
                          _.every(args, arg => arg.type === intLabel)
                          ? { val
                              : _.reduce
                                ( args
                                , (arg0, arg1) => makeInt(arg0.data + arg1.data)
                                , makeInt(0))}
                          : { ok: false
                            , val
                              : strToChars
                                ("Additional argument wasn't integral")}))}

            : stringIs(varKey, '-')
              ? { val
                  : quote
                    ( fnOfType
                      ( listLabel
                      , args =>
                        { if (args.length === 0) return {val: makeInt(-1)}

                        ; if (args[0].type !== intLabel) return {ok: false}
                        ; if (args.length === 1)
                            return {val: makeInt(-args[0].data)}

                        ; return (
                            _.every(args, arg => arg.type === intLabel)
                            ? { val
                                : _.reduce
                                  ( args
                                  , (arg0, arg1) =>
                                      makeInt(arg0.data - arg1.data))}
                            : { ok: false
                              , val
                                : strToChars
                                  ( "Subtractional argument wasn't integral"
                                    + "")})}))}

            : stringIs(varKey, '<')
              ? { val
                  : quote
                    ( fnOfType
                      ( intLabel
                      , arg0 =>
                          ( { val
                              : fnOfType
                                ( intLabel
                                , arg1 => ({val: makeBool(arg0 < arg1)}))})))}

            : stringIs(varKey, '=')
              ? { val
                  : quote
                    ( makeFun
                      ( arg0 =>
                          ( { val
                              : makeFun
                                ( arg1 =>
                                    ({val: makeBool(eq(arg0, arg1))}))})))}

            : stringIs(varKey, "env") ? {val: I}

            : stringIs(varKey, "init-env") ? {val: quote(initEnv)}

            : stringIs(varKey, "print")
              ? { val
                  : quote
                    ( makeFun
                      ( arg =>
                          isString(arg)
                          ? ( process.stdout.write(strVal(arg))
                            , {val: unit})
                          : { ok: false
                            , val: strToChars('Tried to print nonstring')}))}

            : stringIs(varKey, "say")
              ? { val
                  : quote
                    ( makeFun
                      ( _.flow
                        ( toString
                        , strVal
                        , process.stdout.write.bind(process.stdout)
                        , _.constant({val: unit}))))}

            : stringIs(varKey, "->str")
              ? {val: quote(makeFun(_.flow(toString, val => ({val}))))}

            : stringIs(varKey, "char->unicode")
              ? { val
                  : quote
                    ( fnOfType
                      (charLabel, _.flow(makeInt, val => ({val}))))}

            : stringIs(varKey, "unicode->char")
              ? { val
                  : quote
                    ( fnOfType
                      (intLabel, _.flow(makeChar, val => ({val}))))}

            : stringIs(varKey, "length")
              ? { val
                  : quote
                    ( fnOfType
                      (listLabel, arg => ({val: makeInt(arg.length)})))}

            : stringIs(varKey, "->list")
              ? { val
                  : quote
                    ( makeFun
                      ( arg =>
                          ( { val
                              : fnOfType
                                ( intLabel
                                , length =>
                                    length < 0
                                    ? { ok: false
                                      , val
                                        : strToChars
                                          ( "Lists must be nonnegative in "
                                            + "length")}
                                    : { fn: syncMap
                                      , arg
                                        : makeList(_.range(length).map(makeInt))
                                      , okThen
                                        : { fn
                                            : makeFun
                                              (fn => ({fn, arg}))}})})))}

            : stringIs(varKey, "true") ? {val: quote(makeBool(true))}

            : stringIs(varKey, "false") ? {val: quote(makeBool(false))}

            : stringIs(varKey, "unit") ? {val: quote(unit)}

            : stringIs(varKey, "read-file")
              ? { val
                  : quote
                    ( makeFun
                      ( arg =>
                          isString(arg)
                          ? {val: strToChars(readFile(strVal(arg)))}
                          : { ok: false
                            , val
                              : strToChars
                                ('Tried to read-file of nonstring')}))}

            : stringIs(varKey, "parse-prog")
              ? { val
                  : quote
                    ( makeFun
                      ( arg =>
                          isString(arg)
                          ? { val
                              : ( parsed =>
                                    parsed.success
                                    ? okResult
                                      ( objToNs
                                        ( { expr: quote(parsed.ast)
                                          , rest
                                            : quote(strToChars(parsed.rest))}))
                                    : errResult
                                      ( makeFun
                                        ( _.flow
                                          ( strVal
                                          , parsed.error
                                          , strToChars
                                          , val => ({val})))))
                                (parseFile(strVal(arg)))}
                          : { ok: false
                            , val: strToChars('Tried to parse nonstring')}))}

            : stringIs(varKey, "eval-file")
              ? { val
                  : quote
                    ( makeFun
                      ( (arg, ...ons) =>
                          isString(arg)
                          ? ( parsed =>
                                parsed.success
                                ? evalProgram(parsed.ast, parsed.rest)
                                : { ok: false
                                  , val: strToChars(parsed.error(strVal(arg)))})
                            (parseFile(readFile(strVal(arg))))
                          : { ok: false
                            , val
                              : strToChars
                                ('Tried to eval-file of nonstring')}))}

            : stringIs(varKey, "gensym")
              ? {val: quote(makeFun(_.flow(gensym, val => ({val}))))}

            : stringIs(varKey, "symbol-debug-info")
              ? {val: quote(fnOfType(symLabel, data => ({val: data.sym})))}

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
                                    ({val: makeCall(fnExpr, argExpr)}))})))}

            : stringIs(varKey, "call-fn-expr")
              ? {val: quote(fnOfType(callLabel, ({fnExpr}) => ({val: fnExpr})))}

            : stringIs(varKey, "call-arg-expr")
              ? { val
                  : quote(fnOfType(callLabel, ({argExpr}) => ({val: argExpr})))}

            : stringIs(varKey, "make-ident")
              ? {val: quote(makeFun(_.flow(makeIdent, val => ({val}))))}

            : stringIs(varKey, "ident-key")
              ? {val: quote(fnOfType(identLabel, val => ({val})))}

            : stringIs(varKey, "delimited-var") ? {val: quote(delimitedVar)}

            : stringIs(varKey, "just")
              ? {val: quote(makeFun(_.flow(just, val => ({val}))))}

            : stringIs(varKey, "nothing") ? {val: quote(nothing)}

            : stringIs(varKey, "justp")
              ? { val
                  : quote
                    ( fnOfType(maybeLabel, arg => ({val: makeBool(arg.is)})))}

            : stringIs(varKey, "unjust")
              ? { val
                  : quote
                    ( fnOfType
                      ( maybeLabel
                      , arg =>
                          arg.is
                          ? {val: arg.val}
                          : { ok: false
                            , val: strToChars("Nothing was unjustified")}))}

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
                              : strToChars
                                ( isString(arg.val)
                                  ? "Err: " + strVal(arg.val)
                                  : "Result was not ok")}))}

            : stringIs(varKey, "unerr")
              ? { val
                  : quote
                    ( fnOfType
                      ( resultLabel
                      , arg =>
                          arg.ok
                          ? { ok: false
                            , val
                              : strToChars
                                ( isString(arg.val)
                                  ? "Ok: " + strVal(arg.val)
                                  : "Result was ok")}
                          : {val: arg.val}))}

            : stringIs(varKey, "make-cell")
              ? {val: quote(makeFun(_.flow(makeCell, val => ({val}))))}

            : stringIs(varKey, "cell-val")
              ? {val: quote(fnOfType(cellLabel, ({val}) => ({val})))}

            : stringIs(varKey, "set")
              ? { val
                  : quote
                    ( fnOfType
                      ( cellLabel
                      , cell =>
                          ( { val
                              : makeFun
                                (val => (cell.val = val, {val: unit}))})))}

            : stringIs(varKey, "cas")
              ? { val
                  : quote
                    ( fnOfType
                      ( cellLabel
                      , cell =>
                          ( { val
                              : makeFun
                                ( oldVal =>
                                    ( { val
                                        : makeFun
                                          ( newVal =>
                                              ( { val
                                                  : eq(cell.val, oldVal)
                                                    ? ( cell.val = newVal
                                                      , oldVal)
                                                    : newVal}))}))})))}

            : stringIs(varKey, "cons")
              ? { val
                : quote
                  ( makeFun
                    ( first =>
                        ( { val
                            : makeFun
                              (last => ({val: makePair(first, last)}))})))}

            : stringIs(varKey, "car")
              ? {val: quote(fnOfType(pairLabel, ({first}) => ({val: first})))}

            : stringIs(varKey, "cdr")
              ? {val: quote(fnOfType(pairLabel, ({last}) => ({val: last})))}

            : varKey.data[0].data == '"'.codePointAt(0)
              ? {val: quote(makeList(varKey.data.slice(1)))}

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
                : { ok: false
                  , val
                    : strToChars
                      ( 'string variable not found in environment: "'
                        + strVal(varKey))})
              (strVal(varKey))
        : {ok: false, val: strToChars("unknown variable: " + strVal(toString(varKey)))})
; exports.initEnv = initEnv

; Error.stackTraceLimit = Infinity

